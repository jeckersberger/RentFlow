package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jeckersberger/rentflow/pkg/common/database"
	"github.com/jeckersberger/rentflow/services/inventory-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/inventory-service/internal/ports"
)

type FlightcasePostgres struct {
	db *database.PostgresPool
}

func NewFlightcasePostgres(db *database.PostgresPool) *FlightcasePostgres {
	return &FlightcasePostgres{db: db}
}

func (r *FlightcasePostgres) Create(ctx context.Context, fc *domain.Flightcase) error {
	query := `
		INSERT INTO inventory.flightcases (
			id, tenant_id, name, description, barcode, weight, location_id,
			created_at, updated_at, created_by_user_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := r.db.Exec(ctx, query,
		fc.ID, fc.TenantID, fc.Name, fc.Description, fc.Barcode, fc.Weight,
		fc.LocationID, fc.CreatedAt, fc.UpdatedAt, fc.CreatedByUserID,
	)

	if err != nil {
		return fmt.Errorf("failed to create flightcase: %w", err)
	}

	// Insert items
	for _, item := range fc.Contents {
		itemQuery := `
			INSERT INTO inventory.flightcase_items (
				flightcase_id, equipment_id, quantity, added_at
			) VALUES ($1, $2, $3, $4)
		`
		_, err := r.db.Exec(ctx, itemQuery, fc.ID, item.EquipmentID, item.Quantity, item.AddedAt)
		if err != nil {
			return fmt.Errorf("failed to insert flightcase item: %w", err)
		}
	}

	return nil
}

func (r *FlightcasePostgres) Update(ctx context.Context, fc *domain.Flightcase) error {
	query := `
		UPDATE inventory.flightcases SET
			name = $3, description = $4, weight = $5, location_id = $6, updated_at = $7
		WHERE id = $1 AND tenant_id = $2
	`

	result, err := r.db.Exec(ctx, query,
		fc.ID, fc.TenantID, fc.Name, fc.Description, fc.Weight,
		fc.LocationID, fc.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update flightcase: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("flightcase not found: %s", fc.ID)
	}

	// Delete existing items and re-insert
	deleteQuery := "DELETE FROM inventory.flightcase_items WHERE flightcase_id = $1"
	_, err = r.db.Exec(ctx, deleteQuery, fc.ID)
	if err != nil {
		return fmt.Errorf("failed to delete flightcase items: %w", err)
	}

	// Insert updated items
	for _, item := range fc.Contents {
		itemQuery := `
			INSERT INTO inventory.flightcase_items (
				flightcase_id, equipment_id, quantity, added_at
			) VALUES ($1, $2, $3, $4)
		`
		_, err := r.db.Exec(ctx, itemQuery, fc.ID, item.EquipmentID, item.Quantity, item.AddedAt)
		if err != nil {
			return fmt.Errorf("failed to insert flightcase item: %w", err)
		}
	}

	return nil
}

func (r *FlightcasePostgres) GetByID(ctx context.Context, tenantID, flightcaseID string) (*domain.Flightcase, error) {
	query := `
		SELECT id, tenant_id, name, description, barcode, weight, location_id,
		       created_at, updated_at, created_by_user_id
		FROM inventory.flightcases
		WHERE id = $1 AND tenant_id = $2
	`

	row := r.db.QueryRow(ctx, query, flightcaseID, tenantID)
	fc, err := r.scanFlightcase(row)
	if err != nil {
		return nil, err
	}

	// Load items
	items, err := r.getFlightcaseItems(ctx, flightcaseID)
	if err != nil {
		return nil, err
	}
	fc.Contents = items

	return fc, nil
}

func (r *FlightcasePostgres) GetByBarcode(ctx context.Context, tenantID, barcode string) (*domain.Flightcase, error) {
	query := `
		SELECT id, tenant_id, name, description, barcode, weight, location_id,
		       created_at, updated_at, created_by_user_id
		FROM inventory.flightcases
		WHERE barcode = $1 AND tenant_id = $2
	`

	row := r.db.QueryRow(ctx, query, barcode, tenantID)
	fc, err := r.scanFlightcase(row)
	if err != nil {
		return nil, err
	}

	// Load items
	items, err := r.getFlightcaseItems(ctx, fc.ID)
	if err != nil {
		return nil, err
	}
	fc.Contents = items

	return fc, nil
}

func (r *FlightcasePostgres) List(ctx context.Context, tenantID string, limit, offset int) (*ports.FlightcaseListResult, error) {
	// Get total count
	countQuery := "SELECT COUNT(*) FROM inventory.flightcases WHERE tenant_id = $1"
	row := r.db.QueryRow(ctx, countQuery, tenantID)
	var total int64
	if err := row.Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count flightcases: %w", err)
	}

	// Get paginated results
	listQuery := `
		SELECT id, tenant_id, name, description, barcode, weight, location_id,
		       created_at, updated_at, created_by_user_id
		FROM inventory.flightcases
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, listQuery, tenantID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query flightcases: %w", err)
	}
	defer rows.Close()

	var items []*domain.Flightcase
	for rows.Next() {
		fc, err := r.scanFlightcaseRow(rows)
		if err != nil {
			return nil, err
		}

		// Load items for each flightcase
		fcItems, err := r.getFlightcaseItems(ctx, fc.ID)
		if err != nil {
			return nil, err
		}
		fc.Contents = fcItems

		items = append(items, fc)
	}

	return &ports.FlightcaseListResult{
		Items:  items,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}, nil
}

func (r *FlightcasePostgres) Delete(ctx context.Context, tenantID, flightcaseID string) error {
	// Delete items first
	itemQuery := "DELETE FROM inventory.flightcase_items WHERE flightcase_id = $1"
	_, err := r.db.Exec(ctx, itemQuery, flightcaseID)
	if err != nil {
		return fmt.Errorf("failed to delete flightcase items: %w", err)
	}

	// Delete flightcase
	query := "DELETE FROM inventory.flightcases WHERE id = $1 AND tenant_id = $2"
	_, err = r.db.Exec(ctx, query, flightcaseID, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete flightcase: %w", err)
	}

	return nil
}

// Helper methods

func (r *FlightcasePostgres) scanFlightcase(row *sql.Row) (*domain.Flightcase, error) {
	fc := &domain.Flightcase{}

	err := row.Scan(
		&fc.ID, &fc.TenantID, &fc.Name, &fc.Description, &fc.Barcode, &fc.Weight,
		&fc.LocationID, &fc.CreatedAt, &fc.UpdatedAt, &fc.CreatedByUserID,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrFlightcaseNotFound
		}
		return nil, fmt.Errorf("failed to scan flightcase: %w", err)
	}

	fc.Contents = []domain.FlightcaseItem{}
	return fc, nil
}

func (r *FlightcasePostgres) scanFlightcaseRow(rows interface {
	Scan(...interface{}) error
}) (*domain.Flightcase, error) {
	fc := &domain.Flightcase{}

	err := rows.Scan(
		&fc.ID, &fc.TenantID, &fc.Name, &fc.Description, &fc.Barcode, &fc.Weight,
		&fc.LocationID, &fc.CreatedAt, &fc.UpdatedAt, &fc.CreatedByUserID,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to scan flightcase: %w", err)
	}

	fc.Contents = []domain.FlightcaseItem{}
	return fc, nil
}

func (r *FlightcasePostgres) getFlightcaseItems(ctx context.Context, flightcaseID string) ([]domain.FlightcaseItem, error) {
	query := `
		SELECT equipment_id, quantity, added_at
		FROM inventory.flightcase_items
		WHERE flightcase_id = $1
	`

	rows, err := r.db.Query(ctx, query, flightcaseID)
	if err != nil {
		return nil, fmt.Errorf("failed to query flightcase items: %w", err)
	}
	defer rows.Close()

	var items []domain.FlightcaseItem
	for rows.Next() {
		var item domain.FlightcaseItem
		if err := rows.Scan(&item.EquipmentID, &item.Quantity, &item.AddedAt); err != nil {
			return nil, fmt.Errorf("failed to scan flightcase item: %w", err)
		}
		items = append(items, item)
	}

	return items, nil
}

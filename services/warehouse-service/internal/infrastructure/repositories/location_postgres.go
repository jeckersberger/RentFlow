package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jeckersberger/rentflow/pkg/common/database"
	"github.com/jeckersberger/rentflow/services/warehouse-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/warehouse-service/internal/ports"
)

type LocationPostgres struct {
	db *database.PostgresPool
}

func NewLocationPostgres(db *database.PostgresPool) ports.LocationRepository {
	return &LocationPostgres{db: db}
}

func (r *LocationPostgres) Create(ctx context.Context, location *domain.Location) error {
	query := `
		INSERT INTO locations
		(id, tenant_id, name, type, parent_id, path, capacity, current_count, sort_order, barcode, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err := r.db.Exec(ctx, query,
		location.ID,
		location.TenantID,
		location.Name,
		string(location.Type),
		location.ParentID,
		location.Path,
		location.Capacity,
		location.CurrentCount,
		location.SortOrder,
		location.Barcode,
		location.CreatedAt,
		location.UpdatedAt,
	)

	return err
}

func (r *LocationPostgres) GetByID(ctx context.Context, tenantID, id string) (*domain.Location, error) {
	query := `
		SELECT id, tenant_id, name, type, parent_id, path, capacity, current_count, sort_order, barcode, created_at, updated_at
		FROM locations
		WHERE tenant_id = $1 AND id = $2
	`

	row := r.db.QueryRow(ctx, query, tenantID, id)
	return locationRowToLocation(row)
}

func (r *LocationPostgres) List(ctx context.Context, tenantID string, limit, offset int) ([]*domain.Location, int, error) {
	query := `
		SELECT id, tenant_id, name, type, parent_id, path, capacity, current_count, sort_order, barcode, created_at, updated_at
		FROM locations
		WHERE tenant_id = $1
		ORDER BY sort_order ASC, created_at ASC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	locations := make([]*domain.Location, 0)
	for rows.Next() {
		location, err := locationRowsToLocation(rows)
		if err != nil {
			return nil, 0, err
		}
		locations = append(locations, location)
	}

	// Get total count
	countQuery := "SELECT COUNT(*) FROM locations WHERE tenant_id = $1"
	var total int
	if err := r.db.QueryRow(ctx, countQuery, tenantID).Scan(&total); err != nil {
		return nil, 0, err
	}

	return locations, total, rows.Err()
}

func (r *LocationPostgres) ListByParent(ctx context.Context, tenantID string, parentID *string) ([]*domain.Location, error) {
	var query string
	var args []interface{}

	if parentID == nil {
		query = `
			SELECT id, tenant_id, name, type, parent_id, path, capacity, current_count, sort_order, barcode, created_at, updated_at
			FROM locations
			WHERE tenant_id = $1 AND parent_id IS NULL
			ORDER BY sort_order ASC
		`
		args = []interface{}{tenantID}
	} else {
		query = `
			SELECT id, tenant_id, name, type, parent_id, path, capacity, current_count, sort_order, barcode, created_at, updated_at
			FROM locations
			WHERE tenant_id = $1 AND parent_id = $2
			ORDER BY sort_order ASC
		`
		args = []interface{}{tenantID, *parentID}
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	locations := make([]*domain.Location, 0)
	for rows.Next() {
		location, err := locationRowsToLocation(rows)
		if err != nil {
			return nil, err
		}
		locations = append(locations, location)
	}

	return locations, rows.Err()
}

func (r *LocationPostgres) Update(ctx context.Context, location *domain.Location) error {
	query := `
		UPDATE locations
		SET name = $1, capacity = $2, current_count = $3, barcode = $4, sort_order = $5, updated_at = $6
		WHERE id = $7 AND tenant_id = $8
	`

	_, err := r.db.Exec(ctx, query,
		location.Name,
		location.Capacity,
		location.CurrentCount,
		location.Barcode,
		location.SortOrder,
		location.UpdatedAt,
		location.ID,
		location.TenantID,
	)

	return err
}

func (r *LocationPostgres) Delete(ctx context.Context, tenantID, id string) error {
	query := `DELETE FROM locations WHERE tenant_id = $1 AND id = $2`
	_, err := r.db.Exec(ctx, query, tenantID, id)
	return err
}

func locationRowToLocation(row *sql.Row) (*domain.Location, error) {
	location := &domain.Location{}
	err := row.Scan(
		&location.ID,
		&location.TenantID,
		&location.Name,
		(*string)(&location.Type),
		&location.ParentID,
		&location.Path,
		&location.Capacity,
		&location.CurrentCount,
		&location.SortOrder,
		&location.Barcode,
		&location.CreatedAt,
		&location.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("location not found")
		}
		return nil, err
	}
	return location, nil
}

func locationRowsToLocation(rows *sql.Rows) (*domain.Location, error) {
	location := &domain.Location{}
	err := rows.Scan(
		&location.ID,
		&location.TenantID,
		&location.Name,
		(*string)(&location.Type),
		&location.ParentID,
		&location.Path,
		&location.Capacity,
		&location.CurrentCount,
		&location.SortOrder,
		&location.Barcode,
		&location.CreatedAt,
		&location.UpdatedAt,
	)
	return location, err
}

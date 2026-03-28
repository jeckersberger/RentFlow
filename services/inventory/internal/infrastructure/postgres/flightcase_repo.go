package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jeckersberger/EquipFlow/services/inventory/internal/domain"

	apperrors "github.com/jeckersberger/EquipFlow/pkg/common/errors"
)

// flightcaseColumns lists all columns of the flightcases table.
const flightcaseColumns = `id, tenant_id, name, barcode, qr_code, description, weight_grams, is_active, created_at, updated_at`

// FlightcaseRepo implements domain.FlightcaseRepository using PostgreSQL.
type FlightcaseRepo struct {
	pool *pgxpool.Pool
}

// NewFlightcaseRepo creates a new FlightcaseRepo.
func NewFlightcaseRepo(pool *pgxpool.Pool) *FlightcaseRepo {
	return &FlightcaseRepo{pool: pool}
}

// scanFlightcase scans a single flightcase row into a domain.Flightcase.
func scanFlightcase(row pgx.Row) (*domain.Flightcase, error) {
	f := &domain.Flightcase{}
	var barcode, qrCode, description *string

	err := row.Scan(
		&f.ID, &f.TenantID, &f.Name, &barcode, &qrCode,
		&description, &f.WeightGrams,
		&f.IsActive, &f.CreatedAt, &f.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if barcode != nil {
		f.Barcode = *barcode
	}
	if qrCode != nil {
		f.QRCode = *qrCode
	}
	if description != nil {
		f.Description = *description
	}
	return f, nil
}

// Create inserts a new flightcase and scans back the generated fields.
func (r *FlightcaseRepo) Create(ctx context.Context, flightcase *domain.Flightcase) error {
	query := `
		INSERT INTO flightcases (
			id, tenant_id, name, barcode, qr_code, description, weight_grams, is_active
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8
		) RETURNING id, created_at, updated_at`

	err := r.pool.QueryRow(ctx, query,
		flightcase.ID, flightcase.TenantID, flightcase.Name,
		nilIfEmpty(flightcase.Barcode), nilIfEmpty(flightcase.QRCode),
		nilIfEmpty(flightcase.Description), flightcase.WeightGrams,
		flightcase.IsActive,
	).Scan(&flightcase.ID, &flightcase.CreatedAt, &flightcase.UpdatedAt)
	if err != nil {
		return fmt.Errorf("flightcase_repo: create: %w", err)
	}
	return nil
}

// GetByID retrieves a flightcase by primary key scoped to a tenant.
func (r *FlightcaseRepo) GetByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*domain.Flightcase, error) {
	query := fmt.Sprintf(`SELECT %s FROM flightcases WHERE id = $1 AND tenant_id = $2`, flightcaseColumns)
	f, err := scanFlightcase(r.pool.QueryRow(ctx, query, id, tenantID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("flightcase_repo: get_by_id: %w", err)
	}
	return f, nil
}

// List returns a paginated list of flightcases for a tenant plus total count.
func (r *FlightcaseRepo) List(ctx context.Context, tenantID uuid.UUID, page int, perPage int) ([]*domain.Flightcase, int64, error) {
	offset := (page - 1) * perPage

	var total int64
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM flightcases WHERE tenant_id = $1`, tenantID,
	).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("flightcase_repo: list count: %w", err)
	}

	query := fmt.Sprintf(
		`SELECT %s FROM flightcases WHERE tenant_id = $1 ORDER BY name ASC LIMIT $2 OFFSET $3`,
		flightcaseColumns,
	)

	rows, err := r.pool.Query(ctx, query, tenantID, perPage, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("flightcase_repo: list query: %w", err)
	}
	defer rows.Close()

	var flightcases []*domain.Flightcase
	for rows.Next() {
		f, scanErr := scanFlightcase(rows)
		if scanErr != nil {
			return nil, 0, fmt.Errorf("flightcase_repo: list scan: %w", scanErr)
		}
		flightcases = append(flightcases, f)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("flightcase_repo: list rows: %w", err)
	}
	return flightcases, total, nil
}

// Update modifies an existing flightcase.
func (r *FlightcaseRepo) Update(ctx context.Context, flightcase *domain.Flightcase) error {
	query := `
		UPDATE flightcases SET
			name = $3, barcode = $4, qr_code = $5, description = $6,
			weight_grams = $7, is_active = $8,
			updated_at = NOW()
		WHERE id = $1 AND tenant_id = $2
		RETURNING updated_at`

	err := r.pool.QueryRow(ctx, query,
		flightcase.ID, flightcase.TenantID,
		flightcase.Name, nilIfEmpty(flightcase.Barcode), nilIfEmpty(flightcase.QRCode),
		nilIfEmpty(flightcase.Description), flightcase.WeightGrams,
		flightcase.IsActive,
	).Scan(&flightcase.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apperrors.ErrNotFound
		}
		return fmt.Errorf("flightcase_repo: update: %w", err)
	}
	return nil
}

// Delete removes a flightcase by primary key scoped to a tenant.
func (r *FlightcaseRepo) Delete(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error {
	tag, err := r.pool.Exec(ctx,
		`DELETE FROM flightcases WHERE id = $1 AND tenant_id = $2`,
		id, tenantID,
	)
	if err != nil {
		return fmt.Errorf("flightcase_repo: delete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

// AddItem inserts a new flightcase item linking equipment to a flightcase.
func (r *FlightcaseRepo) AddItem(ctx context.Context, item *domain.FlightcaseItem) error {
	query := `
		INSERT INTO flightcase_items (
			id, flightcase_id, equipment_id, quantity, sort_order
		) VALUES (
			$1, $2, $3, $4, $5
		) RETURNING created_at`

	err := r.pool.QueryRow(ctx, query,
		item.ID, item.FlightcaseID, item.EquipmentID,
		item.Quantity, item.SortOrder,
	).Scan(&item.CreatedAt)
	if err != nil {
		return fmt.Errorf("flightcase_repo: add_item: %w", err)
	}
	return nil
}

// RemoveItem deletes a flightcase item by its ID.
func (r *FlightcaseRepo) RemoveItem(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx,
		`DELETE FROM flightcase_items WHERE id = $1`,
		id,
	)
	if err != nil {
		return fmt.Errorf("flightcase_repo: remove_item: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

// GetItems returns all items in a flightcase with equipment data joined.
func (r *FlightcaseRepo) GetItems(ctx context.Context, flightcaseID uuid.UUID) ([]*domain.FlightcaseItem, error) {
	query := `
		SELECT fi.id, fi.flightcase_id, fi.equipment_id, fi.quantity, fi.sort_order, fi.created_at
		FROM flightcase_items fi
		WHERE fi.flightcase_id = $1
		ORDER BY fi.sort_order ASC`

	rows, err := r.pool.Query(ctx, query, flightcaseID)
	if err != nil {
		return nil, fmt.Errorf("flightcase_repo: get_items query: %w", err)
	}
	defer rows.Close()

	var items []*domain.FlightcaseItem
	for rows.Next() {
		item := &domain.FlightcaseItem{}
		scanErr := rows.Scan(
			&item.ID, &item.FlightcaseID, &item.EquipmentID,
			&item.Quantity, &item.SortOrder, &item.CreatedAt,
		)
		if scanErr != nil {
			return nil, fmt.Errorf("flightcase_repo: get_items scan: %w", scanErr)
		}
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("flightcase_repo: get_items rows: %w", err)
	}
	return items, nil
}

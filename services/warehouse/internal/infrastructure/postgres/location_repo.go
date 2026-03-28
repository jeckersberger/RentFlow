package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jeckersberger/EquipFlow/services/warehouse/internal/domain"

	apperrors "github.com/jeckersberger/EquipFlow/pkg/common/errors"
)

// locationColumns lists all columns of the stock_locations table.
const locationColumns = `id, rack_id, zone_id, tenant_id, code, barcode, level, bay, max_weight_kg, is_active, created_at`

// LocationRepo implements domain.StockLocationRepository using PostgreSQL.
type LocationRepo struct {
	pool *pgxpool.Pool
}

// NewLocationRepo creates a new LocationRepo.
func NewLocationRepo(pool *pgxpool.Pool) *LocationRepo {
	return &LocationRepo{pool: pool}
}

// scanLocation scans a single stock_location row into a domain.StockLocation.
func scanLocation(row pgx.Row) (*domain.StockLocation, error) {
	l := &domain.StockLocation{}
	var barcode *string

	err := row.Scan(
		&l.ID, &l.RackID, &l.ZoneID, &l.TenantID,
		&l.Code, &barcode, &l.Level, &l.Bay,
		&l.MaxWeightKg, &l.IsActive, &l.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	if barcode != nil {
		l.Barcode = *barcode
	}
	return l, nil
}

// Create inserts a new stock location and scans back the generated fields.
func (r *LocationRepo) Create(ctx context.Context, location *domain.StockLocation) error {
	query := `
		INSERT INTO stock_locations (
			id, rack_id, zone_id, tenant_id, code, barcode, level, bay, max_weight_kg, is_active
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		) RETURNING id, created_at`

	err := r.pool.QueryRow(ctx, query,
		location.ID, location.RackID, location.ZoneID, location.TenantID,
		location.Code, nilIfEmpty(location.Barcode),
		location.Level, location.Bay, location.MaxWeightKg, location.IsActive,
	).Scan(&location.ID, &location.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			if strings.Contains(pgErr.ConstraintName, "code") {
				return domain.ErrDuplicateCode
			}
		}
		return fmt.Errorf("location_repo: create: %w", err)
	}
	return nil
}

// GetByID retrieves a stock location by primary key scoped to a tenant.
func (r *LocationRepo) GetByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*domain.StockLocation, error) {
	query := fmt.Sprintf(`SELECT %s FROM stock_locations WHERE id = $1 AND tenant_id = $2`, locationColumns)
	l, err := scanLocation(r.pool.QueryRow(ctx, query, id, tenantID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("location_repo: get_by_id: %w", err)
	}
	return l, nil
}

// List returns a paginated list of stock locations for a tenant plus total count.
func (r *LocationRepo) List(ctx context.Context, tenantID uuid.UUID, page int, perPage int) ([]*domain.StockLocation, int64, error) {
	offset := (page - 1) * perPage

	var total int64
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM stock_locations WHERE tenant_id = $1`, tenantID,
	).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("location_repo: list count: %w", err)
	}

	query := fmt.Sprintf(
		`SELECT %s FROM stock_locations WHERE tenant_id = $1 ORDER BY code ASC LIMIT $2 OFFSET $3`,
		locationColumns,
	)

	rows, err := r.pool.Query(ctx, query, tenantID, perPage, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("location_repo: list query: %w", err)
	}
	defer rows.Close()

	var locations []*domain.StockLocation
	for rows.Next() {
		l, scanErr := scanLocation(rows)
		if scanErr != nil {
			return nil, 0, fmt.Errorf("location_repo: list scan: %w", scanErr)
		}
		locations = append(locations, l)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("location_repo: list rows: %w", err)
	}
	return locations, total, nil
}

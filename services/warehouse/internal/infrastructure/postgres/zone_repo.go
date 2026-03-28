package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jeckersberger/EquipFlow/services/warehouse/internal/domain"

	apperrors "github.com/jeckersberger/EquipFlow/pkg/common/errors"
)

// zoneColumns lists all columns of the zones table.
const zoneColumns = `id, warehouse_id, tenant_id, name, code, climate_controlled, max_weight_kg, notes, created_at`

// ZoneRepo implements domain.ZoneRepository using PostgreSQL.
type ZoneRepo struct {
	pool *pgxpool.Pool
}

// NewZoneRepo creates a new ZoneRepo.
func NewZoneRepo(pool *pgxpool.Pool) *ZoneRepo {
	return &ZoneRepo{pool: pool}
}

// scanZone scans a single zone row into a domain.Zone.
func scanZone(row pgx.Row) (*domain.Zone, error) {
	z := &domain.Zone{}
	var code, notes *string

	err := row.Scan(
		&z.ID, &z.WarehouseID, &z.TenantID, &z.Name,
		&code, &z.ClimateControlled, &z.MaxWeightKg,
		&notes, &z.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	if code != nil {
		z.Code = *code
	}
	if notes != nil {
		z.Notes = *notes
	}
	return z, nil
}

// Create inserts a new zone and scans back the generated fields.
func (r *ZoneRepo) Create(ctx context.Context, zone *domain.Zone) error {
	query := `
		INSERT INTO zones (
			id, warehouse_id, tenant_id, name, code, climate_controlled, max_weight_kg, notes
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8
		) RETURNING id, created_at`

	err := r.pool.QueryRow(ctx, query,
		zone.ID, zone.WarehouseID, zone.TenantID, zone.Name,
		nilIfEmpty(zone.Code), zone.ClimateControlled, zone.MaxWeightKg,
		nilIfEmpty(zone.Notes),
	).Scan(&zone.ID, &zone.CreatedAt)
	if err != nil {
		return fmt.Errorf("zone_repo: create: %w", err)
	}
	return nil
}

// GetByID retrieves a zone by primary key scoped to a tenant.
func (r *ZoneRepo) GetByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*domain.Zone, error) {
	query := fmt.Sprintf(`SELECT %s FROM zones WHERE id = $1 AND tenant_id = $2`, zoneColumns)
	z, err := scanZone(r.pool.QueryRow(ctx, query, id, tenantID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("zone_repo: get_by_id: %w", err)
	}
	return z, nil
}

// List returns all zones for a tenant ordered by name.
func (r *ZoneRepo) List(ctx context.Context, tenantID uuid.UUID) ([]*domain.Zone, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM zones WHERE tenant_id = $1 ORDER BY name ASC`,
		zoneColumns,
	)

	rows, err := r.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("zone_repo: list query: %w", err)
	}
	defer rows.Close()

	var zones []*domain.Zone
	for rows.Next() {
		z, scanErr := scanZone(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("zone_repo: list scan: %w", scanErr)
		}
		zones = append(zones, z)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("zone_repo: list rows: %w", err)
	}
	return zones, nil
}

// Update modifies an existing zone.
func (r *ZoneRepo) Update(ctx context.Context, zone *domain.Zone) error {
	query := `
		UPDATE zones SET
			name = $3, code = $4, climate_controlled = $5, max_weight_kg = $6, notes = $7
		WHERE id = $1 AND tenant_id = $2
		RETURNING created_at`

	err := r.pool.QueryRow(ctx, query,
		zone.ID, zone.TenantID,
		zone.Name, nilIfEmpty(zone.Code), zone.ClimateControlled,
		zone.MaxWeightKg, nilIfEmpty(zone.Notes),
	).Scan(&zone.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apperrors.ErrNotFound
		}
		return fmt.Errorf("zone_repo: update: %w", err)
	}
	return nil
}

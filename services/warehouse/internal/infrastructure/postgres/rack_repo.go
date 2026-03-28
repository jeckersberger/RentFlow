package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jeckersberger/EquipFlow/services/warehouse/internal/domain"
)

// rackColumns lists all columns of the racks table.
const rackColumns = `id, zone_id, tenant_id, name, code, levels, bays_per_level, max_weight_kg, created_at`

// RackRepo implements domain.RackRepository using PostgreSQL.
type RackRepo struct {
	pool *pgxpool.Pool
}

// NewRackRepo creates a new RackRepo.
func NewRackRepo(pool *pgxpool.Pool) *RackRepo {
	return &RackRepo{pool: pool}
}

// scanRack scans a single rack row into a domain.Rack.
func scanRack(row pgx.Row) (*domain.Rack, error) {
	rk := &domain.Rack{}
	var code *string

	err := row.Scan(
		&rk.ID, &rk.ZoneID, &rk.TenantID, &rk.Name,
		&code, &rk.Levels, &rk.BaysPerLevel, &rk.MaxWeightKg,
		&rk.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	if code != nil {
		rk.Code = *code
	}
	return rk, nil
}

// Create inserts a new rack and scans back the generated fields.
func (r *RackRepo) Create(ctx context.Context, rack *domain.Rack) error {
	query := `
		INSERT INTO racks (
			id, zone_id, tenant_id, name, code, levels, bays_per_level, max_weight_kg
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8
		) RETURNING id, created_at`

	err := r.pool.QueryRow(ctx, query,
		rack.ID, rack.ZoneID, rack.TenantID, rack.Name,
		nilIfEmpty(rack.Code), rack.Levels, rack.BaysPerLevel, rack.MaxWeightKg,
	).Scan(&rack.ID, &rack.CreatedAt)
	if err != nil {
		return fmt.Errorf("rack_repo: create: %w", err)
	}
	return nil
}

// List returns all racks for a tenant ordered by name.
func (r *RackRepo) List(ctx context.Context, tenantID uuid.UUID) ([]*domain.Rack, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM racks WHERE tenant_id = $1 ORDER BY name ASC`,
		rackColumns,
	)

	rows, err := r.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("rack_repo: list query: %w", err)
	}
	defer rows.Close()

	var racks []*domain.Rack
	for rows.Next() {
		rk, scanErr := scanRack(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("rack_repo: list scan: %w", scanErr)
		}
		racks = append(racks, rk)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rack_repo: list rows: %w", err)
	}
	return racks, nil
}

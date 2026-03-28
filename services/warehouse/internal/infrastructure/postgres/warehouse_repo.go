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

// warehouseColumns lists all columns of the warehouses table.
const warehouseColumns = `id, tenant_id, name, code, address, capacity_description, is_active, created_at`

// WarehouseRepo implements domain.WarehouseRepository using PostgreSQL.
type WarehouseRepo struct {
	pool *pgxpool.Pool
}

// NewWarehouseRepo creates a new WarehouseRepo.
func NewWarehouseRepo(pool *pgxpool.Pool) *WarehouseRepo {
	return &WarehouseRepo{pool: pool}
}

// scanWarehouse scans a single warehouse row into a domain.Warehouse.
func scanWarehouse(row pgx.Row) (*domain.Warehouse, error) {
	w := &domain.Warehouse{}
	var code, address, capacityDesc *string

	err := row.Scan(
		&w.ID, &w.TenantID, &w.Name, &code,
		&address, &capacityDesc, &w.IsActive, &w.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	if code != nil {
		w.Code = *code
	}
	if address != nil {
		w.Address = *address
	}
	if capacityDesc != nil {
		w.CapacityDescription = *capacityDesc
	}
	return w, nil
}

// Create inserts a new warehouse and scans back the generated fields.
func (r *WarehouseRepo) Create(ctx context.Context, warehouse *domain.Warehouse) error {
	query := `
		INSERT INTO warehouses (
			id, tenant_id, name, code, address, capacity_description, is_active
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		) RETURNING id, created_at`

	err := r.pool.QueryRow(ctx, query,
		warehouse.ID, warehouse.TenantID, warehouse.Name,
		nilIfEmpty(warehouse.Code), nilIfEmpty(warehouse.Address),
		nilIfEmpty(warehouse.CapacityDescription), warehouse.IsActive,
	).Scan(&warehouse.ID, &warehouse.CreatedAt)
	if err != nil {
		return fmt.Errorf("warehouse_repo: create: %w", err)
	}
	return nil
}

// GetByID retrieves a warehouse by primary key scoped to a tenant.
func (r *WarehouseRepo) GetByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*domain.Warehouse, error) {
	query := fmt.Sprintf(`SELECT %s FROM warehouses WHERE id = $1 AND tenant_id = $2`, warehouseColumns)
	w, err := scanWarehouse(r.pool.QueryRow(ctx, query, id, tenantID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("warehouse_repo: get_by_id: %w", err)
	}
	return w, nil
}

// List returns all warehouses for a tenant ordered by name.
func (r *WarehouseRepo) List(ctx context.Context, tenantID uuid.UUID) ([]*domain.Warehouse, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM warehouses WHERE tenant_id = $1 ORDER BY name ASC`,
		warehouseColumns,
	)

	rows, err := r.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("warehouse_repo: list query: %w", err)
	}
	defer rows.Close()

	var warehouses []*domain.Warehouse
	for rows.Next() {
		w, scanErr := scanWarehouse(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("warehouse_repo: list scan: %w", scanErr)
		}
		warehouses = append(warehouses, w)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("warehouse_repo: list rows: %w", err)
	}
	return warehouses, nil
}

package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jeckersberger/EquipFlow/services/inventory/internal/domain"

	apperrors "github.com/jeckersberger/EquipFlow/pkg/common/errors"
)

// equipmentTypeColumns lists all columns of the equipment_types table.
const equipmentTypeColumns = `
	id, tenant_id, category_id, name, description,
	default_rental_price_day, default_rental_price_week, default_replacement_value,
	specifications, is_active, created_at, updated_at`

// EquipmentTypeRepo implements domain.EquipmentTypeRepository using PostgreSQL.
type EquipmentTypeRepo struct {
	pool *pgxpool.Pool
}

// NewEquipmentTypeRepo creates a new EquipmentTypeRepo.
func NewEquipmentTypeRepo(pool *pgxpool.Pool) *EquipmentTypeRepo {
	return &EquipmentTypeRepo{pool: pool}
}

// scanEquipmentType scans a single equipment type row into a domain.EquipmentType.
func scanEquipmentType(row pgx.Row) (*domain.EquipmentType, error) {
	et := &domain.EquipmentType{}
	var description *string
	var specifications []byte

	err := row.Scan(
		&et.ID, &et.TenantID, &et.CategoryID,
		&et.Name, &description,
		&et.DefaultRentalPriceDay, &et.DefaultRentalPriceWeek, &et.DefaultReplacementValue,
		&specifications, &et.IsActive,
		&et.CreatedAt, &et.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if description != nil {
		et.Description = *description
	}
	if specifications != nil {
		et.Specifications = json.RawMessage(specifications)
	}
	return et, nil
}

// Create inserts a new equipment type and scans back the generated fields.
func (r *EquipmentTypeRepo) Create(ctx context.Context, equipmentType *domain.EquipmentType) error {
	query := `
		INSERT INTO equipment_types (
			id, tenant_id, category_id, name, description,
			default_rental_price_day, default_rental_price_week, default_replacement_value,
			specifications, is_active
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8,
			$9, $10
		) RETURNING id, created_at, updated_at`

	err := r.pool.QueryRow(ctx, query,
		equipmentType.ID, equipmentType.TenantID, equipmentType.CategoryID,
		equipmentType.Name, nilIfEmpty(equipmentType.Description),
		equipmentType.DefaultRentalPriceDay, equipmentType.DefaultRentalPriceWeek,
		equipmentType.DefaultReplacementValue,
		nilIfEmptyJSON(equipmentType.Specifications), equipmentType.IsActive,
	).Scan(&equipmentType.ID, &equipmentType.CreatedAt, &equipmentType.UpdatedAt)
	if err != nil {
		return fmt.Errorf("equipment_type_repo: create: %w", err)
	}
	return nil
}

// GetByID retrieves an equipment type by primary key scoped to a tenant.
func (r *EquipmentTypeRepo) GetByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*domain.EquipmentType, error) {
	query := fmt.Sprintf(`SELECT %s FROM equipment_types WHERE id = $1 AND tenant_id = $2`, equipmentTypeColumns)
	et, err := scanEquipmentType(r.pool.QueryRow(ctx, query, id, tenantID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("equipment_type_repo: get_by_id: %w", err)
	}
	return et, nil
}

// List returns all equipment types for a tenant ordered by name.
func (r *EquipmentTypeRepo) List(ctx context.Context, tenantID uuid.UUID) ([]*domain.EquipmentType, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM equipment_types WHERE tenant_id = $1 ORDER BY name ASC`,
		equipmentTypeColumns,
	)

	rows, err := r.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("equipment_type_repo: list query: %w", err)
	}
	defer rows.Close()

	var types []*domain.EquipmentType
	for rows.Next() {
		et, scanErr := scanEquipmentType(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("equipment_type_repo: list scan: %w", scanErr)
		}
		types = append(types, et)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("equipment_type_repo: list rows: %w", err)
	}
	return types, nil
}

// Update modifies an existing equipment type.
func (r *EquipmentTypeRepo) Update(ctx context.Context, equipmentType *domain.EquipmentType) error {
	query := `
		UPDATE equipment_types SET
			category_id = $3, name = $4, description = $5,
			default_rental_price_day = $6, default_rental_price_week = $7,
			default_replacement_value = $8,
			specifications = $9, is_active = $10,
			updated_at = NOW()
		WHERE id = $1 AND tenant_id = $2
		RETURNING updated_at`

	err := r.pool.QueryRow(ctx, query,
		equipmentType.ID, equipmentType.TenantID,
		equipmentType.CategoryID, equipmentType.Name, nilIfEmpty(equipmentType.Description),
		equipmentType.DefaultRentalPriceDay, equipmentType.DefaultRentalPriceWeek,
		equipmentType.DefaultReplacementValue,
		nilIfEmptyJSON(equipmentType.Specifications), equipmentType.IsActive,
	).Scan(&equipmentType.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apperrors.ErrNotFound
		}
		return fmt.Errorf("equipment_type_repo: update: %w", err)
	}
	return nil
}

// Delete removes an equipment type by primary key scoped to a tenant.
func (r *EquipmentTypeRepo) Delete(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error {
	tag, err := r.pool.Exec(ctx,
		`DELETE FROM equipment_types WHERE id = $1 AND tenant_id = $2`,
		id, tenantID,
	)
	if err != nil {
		return fmt.Errorf("equipment_type_repo: delete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jeckersberger/EquipFlow/services/transport/internal/domain"

	apperrors "github.com/jeckersberger/EquipFlow/pkg/common/errors"
)

// vehicleColumns lists all columns of the vehicles table for consistent scanning.
const vehicleColumns = `
	id, tenant_id, name, license_plate, type, capacity_kg,
	capacity_description, payload_kg, volume_m3, fuel_type, fuel_consumption,
	is_active, notes, created_at`

// VehicleRepo implements domain.VehicleRepository using PostgreSQL.
type VehicleRepo struct {
	pool *pgxpool.Pool
}

// NewVehicleRepo creates a new VehicleRepo.
func NewVehicleRepo(pool *pgxpool.Pool) *VehicleRepo {
	return &VehicleRepo{pool: pool}
}

// scanVehicle scans a single vehicles row into a domain.Vehicle.
func scanVehicle(row pgx.Row) (*domain.Vehicle, error) {
	v := &domain.Vehicle{}
	var (
		licensePlate        *string
		vType               *string
		capacityDescription *string
		fuelType            *string
		notes               *string
	)

	err := row.Scan(
		&v.ID, &v.TenantID, &v.Name, &licensePlate, &vType, &v.CapacityKg,
		&capacityDescription, &v.PayloadKg, &v.VolumeM3, &fuelType, &v.FuelConsumption,
		&v.IsActive, &notes, &v.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	if licensePlate != nil {
		v.LicensePlate = *licensePlate
	}
	if vType != nil {
		v.Type = *vType
	}
	if capacityDescription != nil {
		v.CapacityDescription = *capacityDescription
	}
	if fuelType != nil {
		v.FuelType = *fuelType
	}
	if notes != nil {
		v.Notes = *notes
	}

	return v, nil
}

// Create inserts a new vehicle record.
func (r *VehicleRepo) Create(ctx context.Context, vehicle *domain.Vehicle) error {
	query := `
		INSERT INTO vehicles (
			id, tenant_id, name, license_plate, type, capacity_kg,
			capacity_description, payload_kg, volume_m3, fuel_type, fuel_consumption,
			is_active, notes
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING created_at`

	err := r.pool.QueryRow(ctx, query,
		vehicle.ID, vehicle.TenantID, vehicle.Name,
		nilIfEmpty(vehicle.LicensePlate), nilIfEmpty(vehicle.Type),
		vehicle.CapacityKg, nilIfEmpty(vehicle.CapacityDescription),
		vehicle.PayloadKg, vehicle.VolumeM3,
		nilIfEmpty(vehicle.FuelType), vehicle.FuelConsumption,
		vehicle.IsActive, nilIfEmpty(vehicle.Notes),
	).Scan(&vehicle.CreatedAt)
	if err != nil {
		return fmt.Errorf("vehicle_repo: create: %w", err)
	}
	return nil
}

// List returns all vehicles for a tenant.
func (r *VehicleRepo) List(ctx context.Context, tenantID uuid.UUID) ([]*domain.Vehicle, error) {
	query := fmt.Sprintf(`SELECT %s FROM vehicles WHERE tenant_id = $1 ORDER BY name ASC`, vehicleColumns)

	rows, err := r.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("vehicle_repo: list query: %w", err)
	}
	defer rows.Close()

	var items []*domain.Vehicle
	for rows.Next() {
		v, scanErr := scanVehicle(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("vehicle_repo: list scan: %w", scanErr)
		}
		items = append(items, v)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("vehicle_repo: list rows: %w", err)
	}
	return items, nil
}

// GetByID retrieves a vehicle by its primary key within a tenant scope.
func (r *VehicleRepo) GetByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*domain.Vehicle, error) {
	query := fmt.Sprintf(`SELECT %s FROM vehicles WHERE id = $1 AND tenant_id = $2`, vehicleColumns)
	v, err := scanVehicle(r.pool.QueryRow(ctx, query, id, tenantID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("vehicle_repo: get_by_id: %w", err)
	}
	return v, nil
}

// Update updates an existing vehicle record.
func (r *VehicleRepo) Update(ctx context.Context, vehicle *domain.Vehicle) error {
	query := `
		UPDATE vehicles SET
			name = $3,
			license_plate = $4,
			type = $5,
			capacity_kg = $6,
			capacity_description = $7,
			payload_kg = $8,
			volume_m3 = $9,
			fuel_type = $10,
			fuel_consumption = $11,
			is_active = $12,
			notes = $13
		WHERE id = $1 AND tenant_id = $2`

	tag, err := r.pool.Exec(ctx, query,
		vehicle.ID, vehicle.TenantID, vehicle.Name,
		nilIfEmpty(vehicle.LicensePlate), nilIfEmpty(vehicle.Type),
		vehicle.CapacityKg, nilIfEmpty(vehicle.CapacityDescription),
		vehicle.PayloadKg, vehicle.VolumeM3,
		nilIfEmpty(vehicle.FuelType), vehicle.FuelConsumption,
		vehicle.IsActive, nilIfEmpty(vehicle.Notes),
	)
	if err != nil {
		return fmt.Errorf("vehicle_repo: update: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

// nilIfEmpty returns nil if the string is empty, otherwise a pointer to it.
func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// Delete removes a vehicle by its ID within a tenant scope.
func (r *VehicleRepo) Delete(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error {
	query := `DELETE FROM vehicles WHERE id = $1 AND tenant_id = $2`
	tag, err := r.pool.Exec(ctx, query, id, tenantID)
	if err != nil {
		return fmt.Errorf("vehicle_repo: delete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

// Ensure interface compliance at compile time.
var _ domain.VehicleRepository = (*VehicleRepo)(nil)

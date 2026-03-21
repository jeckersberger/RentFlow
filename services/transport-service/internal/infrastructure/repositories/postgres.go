package repositories

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/database"
	"github.com/jeckersberger/rentflow/services/transport-service/internal/domain"
)

type VehiclePostgres struct {
	db *database.PostgresPool
}

func NewVehiclePostgres(db *database.PostgresPool) *VehiclePostgres {
	return &VehiclePostgres{db: db}
}

func (r *VehiclePostgres) Create(ctx context.Context, vehicle *domain.Vehicle) error {
	query := `
		INSERT INTO vehicles (id, tenant_id, name, license_plate, type, capacity, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.db.Exec(ctx, query, vehicle.ID, vehicle.TenantID, vehicle.Name, vehicle.LicensePlate, vehicle.Type, vehicle.Capacity, vehicle.Status, vehicle.CreatedAt, vehicle.UpdatedAt)
	return err
}

func (r *VehiclePostgres) GetByID(ctx context.Context, tenantID, id string) (*domain.Vehicle, error) {
	query := `SELECT id, tenant_id, name, license_plate, type, capacity, status, created_at, updated_at FROM vehicles WHERE id = $1 AND tenant_id = $2`
	row := r.db.QueryRow(ctx, query, id, tenantID)
	vehicle := &domain.Vehicle{}
	err := row.Scan(&vehicle.ID, &vehicle.TenantID, &vehicle.Name, &vehicle.LicensePlate, &vehicle.Type, &vehicle.Capacity, &vehicle.Status, &vehicle.CreatedAt, &vehicle.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return vehicle, nil
}

func (r *VehiclePostgres) ListByTenant(ctx context.Context, tenantID string) ([]*domain.Vehicle, error) {
	query := `SELECT id, tenant_id, name, license_plate, type, capacity, status, created_at, updated_at FROM vehicles WHERE tenant_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vehicles []*domain.Vehicle
	for rows.Next() {
		v := &domain.Vehicle{}
		if err := rows.Scan(&v.ID, &v.TenantID, &v.Name, &v.LicensePlate, &v.Type, &v.Capacity, &v.Status, &v.CreatedAt, &v.UpdatedAt); err != nil {
			return nil, err
		}
		vehicles = append(vehicles, v)
	}
	return vehicles, rows.Err()
}

func (r *VehiclePostgres) Update(ctx context.Context, vehicle *domain.Vehicle) error {
	query := `UPDATE vehicles SET name = $1, status = $2, updated_at = $3 WHERE id = $4 AND tenant_id = $5`
	_, err := r.db.Exec(ctx, query, vehicle.Name, vehicle.Status, vehicle.UpdatedAt, vehicle.ID, vehicle.TenantID)
	return err
}

func (r *VehiclePostgres) Delete(ctx context.Context, tenantID, id string) error {
	query := `DELETE FROM vehicles WHERE id = $1 AND tenant_id = $2`
	_, err := r.db.Exec(ctx, query, id, tenantID)
	return err
}

type TourPostgres struct {
	db *database.PostgresPool
}

func NewTourPostgres(db *database.PostgresPool) *TourPostgres {
	return &TourPostgres{db: db}
}

func (r *TourPostgres) Create(ctx context.Context, tour *domain.Tour) error {
	stopsJSON, _ := json.Marshal(tour.Stops)
	query := `
		INSERT INTO tours (id, tenant_id, project_id, vehicle_id, driver_id, date, stops, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.db.Exec(ctx, query, tour.ID, tour.TenantID, tour.ProjectID, tour.VehicleID, tour.DriverID, tour.Date, stopsJSON, tour.Status, tour.CreatedAt, tour.UpdatedAt)
	return err
}

func (r *TourPostgres) GetByID(ctx context.Context, tenantID, id string) (*domain.Tour, error) {
	query := `SELECT id, tenant_id, project_id, vehicle_id, driver_id, date, stops, status, created_at, updated_at FROM tours WHERE id = $1 AND tenant_id = $2`
	row := r.db.QueryRow(ctx, query, id, tenantID)
	tour := &domain.Tour{}
	var stopsJSON []byte
	err := row.Scan(&tour.ID, &tour.TenantID, &tour.ProjectID, &tour.VehicleID, &tour.DriverID, &tour.Date, &stopsJSON, &tour.Status, &tour.CreatedAt, &tour.UpdatedAt)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(stopsJSON, &tour.Stops)
	return tour, nil
}

func (r *TourPostgres) ListByTenant(ctx context.Context, tenantID string) ([]*domain.Tour, error) {
	query := `SELECT id, tenant_id, project_id, vehicle_id, driver_id, date, stops, status, created_at, updated_at FROM tours WHERE tenant_id = $1 ORDER BY date DESC`
	rows, err := r.db.Query(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tours []*domain.Tour
	for rows.Next() {
		t := &domain.Tour{}
		var stopsJSON []byte
		if err := rows.Scan(&t.ID, &t.TenantID, &t.ProjectID, &t.VehicleID, &t.DriverID, &t.Date, &stopsJSON, &t.Status, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		json.Unmarshal(stopsJSON, &t.Stops)
		tours = append(tours, t)
	}
	return tours, rows.Err()
}

func (r *TourPostgres) ListByDate(ctx context.Context, tenantID string, date time.Time) ([]*domain.Tour, error) {
	query := `SELECT id, tenant_id, project_id, vehicle_id, driver_id, date, stops, status, created_at, updated_at FROM tours WHERE tenant_id = $1 AND DATE(date) = DATE($2) ORDER BY date ASC`
	rows, err := r.db.Query(ctx, query, tenantID, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tours []*domain.Tour
	for rows.Next() {
		t := &domain.Tour{}
		var stopsJSON []byte
		if err := rows.Scan(&t.ID, &t.TenantID, &t.ProjectID, &t.VehicleID, &t.DriverID, &t.Date, &stopsJSON, &t.Status, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		json.Unmarshal(stopsJSON, &t.Stops)
		tours = append(tours, t)
	}
	return tours, rows.Err()
}

func (r *TourPostgres) Update(ctx context.Context, tour *domain.Tour) error {
	query := `UPDATE tours SET status = $1, updated_at = $2 WHERE id = $3 AND tenant_id = $4`
	_, err := r.db.Exec(ctx, query, tour.Status, tour.UpdatedAt, tour.ID, tour.TenantID)
	return err
}

func (r *TourPostgres) Delete(ctx context.Context, tenantID, id string) error {
	query := `DELETE FROM tours WHERE id = $1 AND tenant_id = $2`
	_, err := r.db.Exec(ctx, query, id, tenantID)
	return err
}

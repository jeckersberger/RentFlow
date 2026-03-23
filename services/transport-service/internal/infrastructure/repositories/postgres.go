package repositories

import (
	"context"

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
		INSERT INTO vehicles (id, tenant_id, name, license_plate, capacity_kg, capacity_m3, vehicle_type, status, dguv_last_check, dguv_next_check, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	_, err := r.db.Exec(ctx, query, vehicle.ID, vehicle.TenantID, vehicle.Name, vehicle.LicensePlate, vehicle.CapacityKg, vehicle.CapacityM3, vehicle.VehicleType, vehicle.Status, vehicle.DGUVLastCheck, vehicle.DGUVNextCheck, vehicle.CreatedAt, vehicle.UpdatedAt)
	return err
}

func (r *VehiclePostgres) GetByID(ctx context.Context, tenantID, id string) (*domain.Vehicle, error) {
	query := `SELECT id, tenant_id, name, license_plate, capacity_kg, capacity_m3, vehicle_type, status, dguv_last_check, dguv_next_check, created_at, updated_at FROM vehicles WHERE id = $1 AND tenant_id = $2`
	row := r.db.QueryRow(ctx, query, id, tenantID)
	vehicle := &domain.Vehicle{}
	err := row.Scan(&vehicle.ID, &vehicle.TenantID, &vehicle.Name, &vehicle.LicensePlate, &vehicle.CapacityKg, &vehicle.CapacityM3, &vehicle.VehicleType, &vehicle.Status, &vehicle.DGUVLastCheck, &vehicle.DGUVNextCheck, &vehicle.CreatedAt, &vehicle.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return vehicle, nil
}

func (r *VehiclePostgres) ListByTenant(ctx context.Context, tenantID string) ([]*domain.Vehicle, error) {
	query := `SELECT id, tenant_id, name, license_plate, capacity_kg, capacity_m3, vehicle_type, status, dguv_last_check, dguv_next_check, created_at, updated_at FROM vehicles WHERE tenant_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vehicles []*domain.Vehicle
	for rows.Next() {
		v := &domain.Vehicle{}
		if err := rows.Scan(&v.ID, &v.TenantID, &v.Name, &v.LicensePlate, &v.CapacityKg, &v.CapacityM3, &v.VehicleType, &v.Status, &v.DGUVLastCheck, &v.DGUVNextCheck, &v.CreatedAt, &v.UpdatedAt); err != nil {
			return nil, err
		}
		vehicles = append(vehicles, v)
	}
	return vehicles, rows.Err()
}

func (r *VehiclePostgres) Update(ctx context.Context, vehicle *domain.Vehicle) error {
	query := `UPDATE vehicles SET name = $1, status = $2, dguv_last_check = $3, dguv_next_check = $4, updated_at = $5 WHERE id = $6 AND tenant_id = $7`
	_, err := r.db.Exec(ctx, query, vehicle.Name, vehicle.Status, vehicle.DGUVLastCheck, vehicle.DGUVNextCheck, vehicle.UpdatedAt, vehicle.ID, vehicle.TenantID)
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
	query := `
		INSERT INTO tours (id, tenant_id, project_id, vehicle_id, driver_id, status, departure_at, arrival_at, km_start, km_end, total_cost, fuel_cost, delivery_note_number, delivery_note_generated_at, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
	`
	_, err := r.db.Exec(ctx, query, tour.ID, tour.TenantID, tour.ProjectID, tour.VehicleID, tour.DriverID, tour.Status, tour.DepartureAt, tour.ArrivalAt, tour.KmStart, tour.KmEnd, tour.TotalCost, tour.FuelCost, tour.DeliveryNoteNumber, tour.DeliveryNoteGeneratedAt, tour.Notes, tour.CreatedAt, tour.UpdatedAt)
	return err
}

func (r *TourPostgres) GetByID(ctx context.Context, tenantID, id string) (*domain.Tour, error) {
	query := `SELECT id, tenant_id, project_id, vehicle_id, driver_id, status, departure_at, arrival_at, km_start, km_end, total_cost, fuel_cost, delivery_note_number, delivery_note_generated_at, notes, created_at, updated_at FROM tours WHERE id = $1 AND tenant_id = $2`
	row := r.db.QueryRow(ctx, query, id, tenantID)
	tour := &domain.Tour{}
	err := row.Scan(&tour.ID, &tour.TenantID, &tour.ProjectID, &tour.VehicleID, &tour.DriverID, &tour.Status, &tour.DepartureAt, &tour.ArrivalAt, &tour.KmStart, &tour.KmEnd, &tour.TotalCost, &tour.FuelCost, &tour.DeliveryNoteNumber, &tour.DeliveryNoteGeneratedAt, &tour.Notes, &tour.CreatedAt, &tour.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return tour, nil
}

func (r *TourPostgres) ListByTenant(ctx context.Context, tenantID string) ([]*domain.Tour, error) {
	query := `SELECT id, tenant_id, project_id, vehicle_id, driver_id, status, departure_at, arrival_at, km_start, km_end, total_cost, fuel_cost, delivery_note_number, delivery_note_generated_at, notes, created_at, updated_at FROM tours WHERE tenant_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tours []*domain.Tour
	for rows.Next() {
		t := &domain.Tour{}
		if err := rows.Scan(&t.ID, &t.TenantID, &t.ProjectID, &t.VehicleID, &t.DriverID, &t.Status, &t.DepartureAt, &t.ArrivalAt, &t.KmStart, &t.KmEnd, &t.TotalCost, &t.FuelCost, &t.DeliveryNoteNumber, &t.DeliveryNoteGeneratedAt, &t.Notes, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		tours = append(tours, t)
	}
	return tours, rows.Err()
}

func (r *TourPostgres) Update(ctx context.Context, tour *domain.Tour) error {
	query := `UPDATE tours SET status = $1, departure_at = $2, arrival_at = $3, km_start = $4, km_end = $5, total_cost = $6, fuel_cost = $7, delivery_note_number = $8, delivery_note_generated_at = $9, notes = $10, updated_at = $11 WHERE id = $12 AND tenant_id = $13`
	_, err := r.db.Exec(ctx, query, tour.Status, tour.DepartureAt, tour.ArrivalAt, tour.KmStart, tour.KmEnd, tour.TotalCost, tour.FuelCost, tour.DeliveryNoteNumber, tour.DeliveryNoteGeneratedAt, tour.Notes, tour.UpdatedAt, tour.ID, tour.TenantID)
	return err
}

func (r *TourPostgres) Delete(ctx context.Context, tenantID, id string) error {
	query := `DELETE FROM tours WHERE id = $1 AND tenant_id = $2`
	_, err := r.db.Exec(ctx, query, id, tenantID)
	return err
}

type TourEquipmentPostgres struct {
	db *database.PostgresPool
}

func NewTourEquipmentPostgres(db *database.PostgresPool) *TourEquipmentPostgres {
	return &TourEquipmentPostgres{db: db}
}

func (r *TourEquipmentPostgres) Create(ctx context.Context, equipment *domain.TourEquipment) error {
	query := `
		INSERT INTO tour_equipment (id, tour_id, equipment_id, weight_kg, volume_m3, loaded_at, unloaded_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.db.Exec(ctx, query, equipment.ID, equipment.TourID, equipment.EquipmentID, equipment.WeightKg, equipment.VolumeM3, equipment.LoadedAt, equipment.UnloadedAt, equipment.CreatedAt, equipment.UpdatedAt)
	return err
}

func (r *TourEquipmentPostgres) GetByID(ctx context.Context, id string) (*domain.TourEquipment, error) {
	query := `SELECT id, tour_id, equipment_id, weight_kg, volume_m3, loaded_at, unloaded_at, created_at, updated_at FROM tour_equipment WHERE id = $1`
	row := r.db.QueryRow(ctx, query, id)
	equipment := &domain.TourEquipment{}
	err := row.Scan(&equipment.ID, &equipment.TourID, &equipment.EquipmentID, &equipment.WeightKg, &equipment.VolumeM3, &equipment.LoadedAt, &equipment.UnloadedAt, &equipment.CreatedAt, &equipment.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return equipment, nil
}

func (r *TourEquipmentPostgres) ListByTour(ctx context.Context, tourID string) ([]*domain.TourEquipment, error) {
	query := `SELECT id, tour_id, equipment_id, weight_kg, volume_m3, loaded_at, unloaded_at, created_at, updated_at FROM tour_equipment WHERE tour_id = $1 ORDER BY created_at ASC`
	rows, err := r.db.Query(ctx, query, tourID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var equipment []*domain.TourEquipment
	for rows.Next() {
		e := &domain.TourEquipment{}
		if err := rows.Scan(&e.ID, &e.TourID, &e.EquipmentID, &e.WeightKg, &e.VolumeM3, &e.LoadedAt, &e.UnloadedAt, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, err
		}
		equipment = append(equipment, e)
	}
	return equipment, rows.Err()
}

func (r *TourEquipmentPostgres) DeleteByTourAndEquipment(ctx context.Context, tourID, equipmentID string) error {
	query := `DELETE FROM tour_equipment WHERE tour_id = $1 AND equipment_id = $2`
	_, err := r.db.Exec(ctx, query, tourID, equipmentID)
	return err
}

func (r *TourEquipmentPostgres) GetCapacityByTour(ctx context.Context, tourID string) (float64, float64, error) {
	query := `SELECT COALESCE(SUM(weight_kg), 0), COALESCE(SUM(volume_m3), 0) FROM tour_equipment WHERE tour_id = $1`
	row := r.db.QueryRow(ctx, query, tourID)
	var totalWeight, totalVolume float64
	err := row.Scan(&totalWeight, &totalVolume)
	return totalWeight, totalVolume, err
}

type DriverLogPostgres struct {
	db *database.PostgresPool
}

func NewDriverLogPostgres(db *database.PostgresPool) *DriverLogPostgres {
	return &DriverLogPostgres{db: db}
}

func (r *DriverLogPostgres) Create(ctx context.Context, log *domain.DriverLog) error {
	query := `
		INSERT INTO driver_logs (id, tour_id, driver_id, start_time, end_time, break_minutes, km_driven, activity_type, rest_minutes, location_start, location_end, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`
	_, err := r.db.Exec(ctx, query, log.ID, log.TourID, log.DriverID, log.StartTime, log.EndTime, log.BreakMinutes, log.KmDriven, log.ActivityType, log.RestMinutes, log.LocationStart, log.LocationEnd, log.Notes, log.CreatedAt, log.UpdatedAt)
	return err
}

func (r *DriverLogPostgres) GetByID(ctx context.Context, id string) (*domain.DriverLog, error) {
	query := `SELECT id, tour_id, driver_id, start_time, end_time, break_minutes, km_driven, activity_type, rest_minutes, location_start, location_end, notes, created_at, updated_at FROM driver_logs WHERE id = $1`
	row := r.db.QueryRow(ctx, query, id)
	log := &domain.DriverLog{}
	err := row.Scan(&log.ID, &log.TourID, &log.DriverID, &log.StartTime, &log.EndTime, &log.BreakMinutes, &log.KmDriven, &log.ActivityType, &log.RestMinutes, &log.LocationStart, &log.LocationEnd, &log.Notes, &log.CreatedAt, &log.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return log, nil
}

func (r *DriverLogPostgres) ListByTour(ctx context.Context, tourID string) ([]*domain.DriverLog, error) {
	query := `SELECT id, tour_id, driver_id, start_time, end_time, break_minutes, km_driven, activity_type, rest_minutes, location_start, location_end, notes, created_at, updated_at FROM driver_logs WHERE tour_id = $1 ORDER BY start_time ASC`
	rows, err := r.db.Query(ctx, query, tourID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*domain.DriverLog
	for rows.Next() {
		l := &domain.DriverLog{}
		if err := rows.Scan(&l.ID, &l.TourID, &l.DriverID, &l.StartTime, &l.EndTime, &l.BreakMinutes, &l.KmDriven, &l.ActivityType, &l.RestMinutes, &l.LocationStart, &l.LocationEnd, &l.Notes, &l.CreatedAt, &l.UpdatedAt); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, rows.Err()
}

func (r *DriverLogPostgres) Update(ctx context.Context, log *domain.DriverLog) error {
	query := `UPDATE driver_logs SET end_time = $1, break_minutes = $2, km_driven = $3, activity_type = $4, rest_minutes = $5, location_start = $6, location_end = $7, notes = $8, updated_at = $9 WHERE id = $10`
	_, err := r.db.Exec(ctx, query, log.EndTime, log.BreakMinutes, log.KmDriven, log.ActivityType, log.RestMinutes, log.LocationStart, log.LocationEnd, log.Notes, log.UpdatedAt, log.ID)
	return err
}

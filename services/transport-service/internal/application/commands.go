package application

import "time"

type CreateVehicleCommand struct {
	TenantID      string     `json:"tenant_id"`
	Name          string     `json:"name"`
	LicensePlate  string     `json:"license_plate"`
	CapacityKg    float64    `json:"capacity_kg"`
	CapacityM3    float64    `json:"capacity_m3"`
	VehicleType   string     `json:"vehicle_type"`
	DGUVLastCheck *time.Time `json:"dguv_last_check"`
	DGUVNextCheck *time.Time `json:"dguv_next_check"`
}

type UpdateVehicleCommand struct {
	TenantID      string     `json:"tenant_id"`
	VehicleID     string     `json:"vehicle_id"`
	Status        string     `json:"status"`
	DGUVLastCheck *time.Time `json:"dguv_last_check"`
	DGUVNextCheck *time.Time `json:"dguv_next_check"`
}

type CreateTourCommand struct {
	TenantID    string     `json:"tenant_id"`
	ProjectID   string     `json:"project_id"`
	VehicleID   string     `json:"vehicle_id"`
	DriverID    string     `json:"driver_id"`
	DepartureAt *time.Time `json:"departure_at"`
	Notes       *string    `json:"notes"`
}

type UpdateTourStatusCommand struct {
	TenantID string `json:"tenant_id"`
	TourID   string `json:"tour_id"`
	Status   string `json:"status"`
}

type StartTourCommand struct {
	TenantID string   `json:"tenant_id"`
	TourID   string   `json:"tour_id"`
	KmStart  *float64 `json:"km_start"`
}

type CompleteTourCommand struct {
	TenantID  string   `json:"tenant_id"`
	TourID    string   `json:"tour_id"`
	KmEnd     *float64 `json:"km_end"`
	TotalCost *float64 `json:"total_cost"`
}

type AddEquipmentToTourCommand struct {
	TenantID    string  `json:"tenant_id"`
	TourID      string  `json:"tour_id"`
	EquipmentID string  `json:"equipment_id"`
	WeightKg    float64 `json:"weight_kg"`
	VolumeM3    float64 `json:"volume_m3"`
}

type RemoveEquipmentFromTourCommand struct {
	TenantID    string `json:"tenant_id"`
	TourID      string `json:"tour_id"`
	EquipmentID string `json:"equipment_id"`
}

type LogDriverActivityCommand struct {
	TenantID     string     `json:"tenant_id"`
	TourID       string     `json:"tour_id"`
	DriverID     string     `json:"driver_id"`
	StartTime    time.Time  `json:"start_time"`
	EndTime      *time.Time `json:"end_time"`
	BreakMinutes int        `json:"break_minutes"`
	KmDriven     *float64   `json:"km_driven"`
	Notes        *string    `json:"notes"`
}

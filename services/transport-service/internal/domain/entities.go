package domain

import "time"

type VehicleType string

const (
	VehicleTypeVan     VehicleType = "van"
	VehicleTypeTruck   VehicleType = "truck"
	VehicleTypeTrailer VehicleType = "trailer"
)

type VehicleStatus string

const (
	VehicleStatusAvailable   VehicleStatus = "available"
	VehicleStatusInUse       VehicleStatus = "in_use"
	VehicleStatusMaintenance VehicleStatus = "maintenance"
)

type TourStatus string

const (
	TourStatusPlanned   TourStatus = "planned"
	TourStatusLoading   TourStatus = "loading"
	TourStatusInTransit TourStatus = "in_transit"
	TourStatusDelivered TourStatus = "delivered"
	TourStatusCompleted TourStatus = "completed"
)

type Vehicle struct {
	ID               string        `json:"id"`
	TenantID         string        `json:"tenant_id"`
	Name             string        `json:"name"`
	LicensePlate     string        `json:"license_plate"`
	CapacityKg       float64       `json:"capacity_kg"`
	CapacityM3       float64       `json:"capacity_m3"`
	VehicleType      VehicleType   `json:"vehicle_type"`
	Status           VehicleStatus `json:"status"`
	DGUVLastCheck    *time.Time    `json:"dguv_last_check"`
	DGUVNextCheck    *time.Time    `json:"dguv_next_check"`
	CreatedAt        time.Time     `json:"created_at"`
	UpdatedAt        time.Time     `json:"updated_at"`
}

type Tour struct {
	ID          string     `json:"id"`
	TenantID    string     `json:"tenant_id"`
	ProjectID   string     `json:"project_id"`
	VehicleID   string     `json:"vehicle_id"`
	DriverID    string     `json:"driver_id"`
	Status      TourStatus `json:"status"`
	DepartureAt *time.Time `json:"departure_at"`
	ArrivalAt   *time.Time `json:"arrival_at"`
	KmStart     *float64   `json:"km_start"`
	KmEnd       *float64   `json:"km_end"`
	TotalCost   *float64   `json:"total_cost"`
	Notes       *string    `json:"notes"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type TourEquipment struct {
	ID         string    `json:"id"`
	TourID     string    `json:"tour_id"`
	EquipmentID string    `json:"equipment_id"`
	WeightKg   float64   `json:"weight_kg"`
	VolumeM3   float64   `json:"volume_m3"`
	LoadedAt   *time.Time `json:"loaded_at"`
	UnloadedAt *time.Time `json:"unloaded_at"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type DriverLog struct {
	ID            string    `json:"id"`
	TourID        string    `json:"tour_id"`
	DriverID      string    `json:"driver_id"`
	StartTime     time.Time `json:"start_time"`
	EndTime       *time.Time `json:"end_time"`
	BreakMinutes  int       `json:"break_minutes"`
	KmDriven      *float64   `json:"km_driven"`
	Notes         *string    `json:"notes"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

package domain

import "time"

type VehicleType string

const (
	VehicleTypeVan    VehicleType = "van"
	VehicleTypeTruck  VehicleType = "truck"
	VehicleTypeTrailer VehicleType = "trailer"
)

type VehicleStatus string

const (
	VehicleStatusAvailable VehicleStatus = "available"
	VehicleStatusInUse     VehicleStatus = "in_use"
	VehicleStatusMaintenance VehicleStatus = "maintenance"
)

type TourStatus string

const (
	TourStatusPlanned    TourStatus = "planned"
	TourStatusInProgress TourStatus = "in_progress"
	TourStatusCompleted  TourStatus = "completed"
)

type Vehicle struct {
	ID           string        `json:"id"`
	TenantID     string        `json:"tenant_id"`
	Name         string        `json:"name"`
	LicensePlate string        `json:"license_plate"`
	Type         VehicleType   `json:"type"`
	Capacity     float64       `json:"capacity"`
	Status       VehicleStatus `json:"status"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
}

type TourStop struct {
	Address       string    `json:"address"`
	ArrivalTime   time.Time `json:"arrival_time"`
	DepartureTime time.Time `json:"departure_time"`
	Type          string    `json:"type"` // pickup, delivery
}

type Tour struct {
	ID        string      `json:"id"`
	TenantID  string      `json:"tenant_id"`
	ProjectID string      `json:"project_id"`
	VehicleID string      `json:"vehicle_id"`
	DriverID  string      `json:"driver_id"`
	Date      time.Time   `json:"date"`
	Stops     []TourStop  `json:"stops"`
	Status    TourStatus  `json:"status"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

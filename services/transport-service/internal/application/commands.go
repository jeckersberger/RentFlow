package application

import "time"

type CreateVehicleCommand struct {
	TenantID     string  `json:"tenant_id"`
	Name         string  `json:"name"`
	LicensePlate string  `json:"license_plate"`
	Type         string  `json:"type"`
	Capacity     float64 `json:"capacity"`
}

type UpdateVehicleCommand struct {
	TenantID string `json:"tenant_id"`
	VehicleID string `json:"vehicle_id"`
	Status   string `json:"status"`
}

type CreateTourCommand struct {
	TenantID  string      `json:"tenant_id"`
	ProjectID string      `json:"project_id"`
	VehicleID string      `json:"vehicle_id"`
	DriverID  string      `json:"driver_id"`
	Date      time.Time   `json:"date"`
	Stops     []map[string]interface{} `json:"stops"`
}

type UpdateTourStatusCommand struct {
	TenantID string `json:"tenant_id"`
	TourID   string `json:"tour_id"`
	Status   string `json:"status"`
}

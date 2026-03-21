package application

import (
	"time"
	"github.com/jeckersberger/rentflow/services/transport-service/internal/domain"
)

type VehicleDTO struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	LicensePlate string    `json:"license_plate"`
	Type         string    `json:"type"`
	Capacity     float64   `json:"capacity"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

type TourDTO struct {
	ID        string                 `json:"id"`
	ProjectID string                 `json:"project_id"`
	VehicleID string                 `json:"vehicle_id"`
	DriverID  string                 `json:"driver_id"`
	Date      time.Time              `json:"date"`
	Stops     []map[string]interface{} `json:"stops"`
	Status    string                 `json:"status"`
	CreatedAt time.Time              `json:"created_at"`
}

func VehicleToDTO(v *domain.Vehicle) *VehicleDTO {
	return &VehicleDTO{
		ID:           v.ID,
		Name:         v.Name,
		LicensePlate: v.LicensePlate,
		Type:         string(v.Type),
		Capacity:     v.Capacity,
		Status:       string(v.Status),
		CreatedAt:    v.CreatedAt,
	}
}

func TourToDTO(t *domain.Tour) *TourDTO {
	stops := make([]map[string]interface{}, len(t.Stops))
	for i, s := range t.Stops {
		stops[i] = map[string]interface{}{
			"address":        s.Address,
			"arrival_time":   s.ArrivalTime,
			"departure_time": s.DepartureTime,
			"type":           s.Type,
		}
	}
	return &TourDTO{
		ID:        t.ID,
		ProjectID: t.ProjectID,
		VehicleID: t.VehicleID,
		DriverID:  t.DriverID,
		Date:      t.Date,
		Stops:     stops,
		Status:    string(t.Status),
		CreatedAt: t.CreatedAt,
	}
}

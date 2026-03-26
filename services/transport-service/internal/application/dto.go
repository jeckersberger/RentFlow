package application

import (
	"time"

	"github.com/jeckersberger/rentflow/services/transport-service/internal/domain"
)

type VehicleDTO struct {
	ID            string     `json:"id"`
	Name          string     `json:"name"`
	LicensePlate  string     `json:"license_plate"`
	CapacityKg    float64    `json:"capacity_kg"`
	CapacityM3    float64    `json:"capacity_m3"`
	VehicleType   string     `json:"vehicle_type"`
	Status        string     `json:"status"`
	DGUVLastCheck *time.Time `json:"dguv_last_check"`
	DGUVNextCheck *time.Time `json:"dguv_next_check"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type TourDTO struct {
	ID                      string     `json:"id"`
	ProjectID               string     `json:"project_id"`
	VehicleID               string     `json:"vehicle_id"`
	DriverID                string     `json:"driver_id"`
	Status                  string     `json:"status"`
	DepartureAt             *time.Time `json:"departure_at"`
	ArrivalAt               *time.Time `json:"arrival_at"`
	KmStart                 *float64   `json:"km_start"`
	KmEnd                   *float64   `json:"km_end"`
	TotalCost               *float64   `json:"total_cost"`
	FuelCost                *float64   `json:"fuel_cost"`
	DeliveryNoteNumber      *string    `json:"delivery_note_number"`
	DeliveryNoteGeneratedAt *time.Time `json:"delivery_note_generated_at"`
	Notes                   *string    `json:"notes"`
	CreatedAt               time.Time  `json:"created_at"`
	UpdatedAt               time.Time  `json:"updated_at"`
}

type TourEquipmentDTO struct {
	ID          string     `json:"id"`
	TourID      string     `json:"tour_id"`
	EquipmentID string     `json:"equipment_id"`
	WeightKg    float64    `json:"weight_kg"`
	VolumeM3    float64    `json:"volume_m3"`
	LoadedAt    *time.Time `json:"loaded_at"`
	UnloadedAt  *time.Time `json:"unloaded_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type DriverLogDTO struct {
	ID               string     `json:"id"`
	TourID           string     `json:"tour_id"`
	DriverID         string     `json:"driver_id"`
	StartTime        time.Time  `json:"start_time"`
	EndTime          *time.Time `json:"end_time"`
	BreakMinutes     int        `json:"break_minutes"`
	KmDriven         *float64   `json:"km_driven"`
	ActivityType     string     `json:"activity_type"`
	RestMinutes      int        `json:"rest_minutes"`
	LocationStart    *string    `json:"location_start"`
	LocationEnd      *string    `json:"location_end"`
	Notes            *string    `json:"notes"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type TourCapacityDTO struct {
	TourID           string  `json:"tour_id"`
	TotalCapacityKg  float64 `json:"total_capacity_kg"`
	TotalCapacityM3  float64 `json:"total_capacity_m3"`
	UsedCapacityKg   float64 `json:"used_capacity_kg"`
	UsedCapacityM3   float64 `json:"used_capacity_m3"`
	RemainingKg      float64 `json:"remaining_kg"`
	RemainingM3      float64 `json:"remaining_m3"`
	CapacityExceeded bool    `json:"capacity_exceeded"`
}

func VehicleToDTO(v *domain.Vehicle) *VehicleDTO {
	return &VehicleDTO{
		ID:            v.ID,
		Name:          v.Name,
		LicensePlate:  v.LicensePlate,
		CapacityKg:    v.CapacityKg,
		CapacityM3:    v.CapacityM3,
		VehicleType:   string(v.VehicleType),
		Status:        string(v.Status),
		DGUVLastCheck: v.DGUVLastCheck,
		DGUVNextCheck: v.DGUVNextCheck,
		CreatedAt:     v.CreatedAt,
		UpdatedAt:     v.UpdatedAt,
	}
}

func TourToDTO(t *domain.Tour) *TourDTO {
	return &TourDTO{
		ID:                      t.ID,
		ProjectID:               t.ProjectID,
		VehicleID:               t.VehicleID,
		DriverID:                t.DriverID,
		Status:                  string(t.Status),
		DepartureAt:             t.DepartureAt,
		ArrivalAt:               t.ArrivalAt,
		KmStart:                 t.KmStart,
		KmEnd:                   t.KmEnd,
		TotalCost:               t.TotalCost,
		FuelCost:                t.FuelCost,
		DeliveryNoteNumber:      t.DeliveryNoteNumber,
		DeliveryNoteGeneratedAt: t.DeliveryNoteGeneratedAt,
		Notes:                   t.Notes,
		CreatedAt:               t.CreatedAt,
		UpdatedAt:               t.UpdatedAt,
	}
}

func TourEquipmentToDTO(te *domain.TourEquipment) *TourEquipmentDTO {
	return &TourEquipmentDTO{
		ID:          te.ID,
		TourID:      te.TourID,
		EquipmentID: te.EquipmentID,
		WeightKg:    te.WeightKg,
		VolumeM3:    te.VolumeM3,
		LoadedAt:    te.LoadedAt,
		UnloadedAt:  te.UnloadedAt,
		CreatedAt:   te.CreatedAt,
		UpdatedAt:   te.UpdatedAt,
	}
}

func DriverLogToDTO(dl *domain.DriverLog) *DriverLogDTO {
	return &DriverLogDTO{
		ID:               dl.ID,
		TourID:           dl.TourID,
		DriverID:         dl.DriverID,
		StartTime:        dl.StartTime,
		EndTime:          dl.EndTime,
		BreakMinutes:     dl.BreakMinutes,
		KmDriven:         dl.KmDriven,
		ActivityType:     dl.ActivityType,
		RestMinutes:      dl.RestMinutes,
		LocationStart:    dl.LocationStart,
		LocationEnd:      dl.LocationEnd,
		Notes:            dl.Notes,
		CreatedAt:        dl.CreatedAt,
		UpdatedAt:        dl.UpdatedAt,
	}
}

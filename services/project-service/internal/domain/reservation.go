package domain

import (
	"fmt"
	"time"
)

type Reservation struct {
	ID          string
	TenantID    string
	ProjectID   string
	EquipmentID string
	StartDate   time.Time
	EndDate     time.Time
	Status      ReservationStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ReservationStatus string

const (
	ReservationPending   ReservationStatus = "pending"
	ReservationConfirmed ReservationStatus = "confirmed"
	ReservationActive    ReservationStatus = "active"
	ReservationCompleted ReservationStatus = "completed"
	ReservationCancelled ReservationStatus = "cancelled"
)

func NewReservation(id, tenantID, projectID, equipmentID string, startDate, endDate time.Time) *Reservation {
	now := time.Now()
	return &Reservation{
		ID:          id,
		TenantID:    tenantID,
		ProjectID:   projectID,
		EquipmentID: equipmentID,
		StartDate:   startDate,
		EndDate:     endDate,
		Status:      ReservationPending,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func (r *Reservation) Confirm() error {
	if r.Status != ReservationPending {
		return fmt.Errorf("can only confirm pending reservations")
	}
	r.Status = ReservationConfirmed
	r.UpdatedAt = time.Now()
	return nil
}

func (r *Reservation) Activate() error {
	if r.Status != ReservationConfirmed {
		return fmt.Errorf("can only activate confirmed reservations")
	}
	r.Status = ReservationActive
	r.UpdatedAt = time.Now()
	return nil
}

func (r *Reservation) Complete() error {
	if r.Status != ReservationActive {
		return fmt.Errorf("can only complete active reservations")
	}
	r.Status = ReservationCompleted
	r.UpdatedAt = time.Now()
	return nil
}

func (r *Reservation) Cancel() error {
	if r.Status == ReservationCancelled || r.Status == ReservationCompleted {
		return fmt.Errorf("cannot cancel %s reservation", r.Status)
	}
	r.Status = ReservationCancelled
	r.UpdatedAt = time.Now()
	return nil
}

func (r *Reservation) Validate() error {
	if r.ID == "" {
		return fmt.Errorf("reservation ID cannot be empty")
	}
	if r.TenantID == "" {
		return fmt.Errorf("tenant ID cannot be empty")
	}
	if r.ProjectID == "" {
		return fmt.Errorf("project ID cannot be empty")
	}
	if r.EquipmentID == "" {
		return fmt.Errorf("equipment ID cannot be empty")
	}
	if r.EndDate.Before(r.StartDate) {
		return fmt.Errorf("end date cannot be before start date")
	}
	return nil
}

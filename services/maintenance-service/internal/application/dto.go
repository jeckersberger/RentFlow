package application

import (
	"github.com/jeckersberger/rentflow/services/maintenance-service/internal/domain"
	"time"
)

type MaintenanceRecordDTO struct {
	ID             string     `json:"id"`
	EquipmentID    string     `json:"equipment_id"`
	Type           string     `json:"type"`
	Status         string     `json:"status"`
	ScheduledDate  time.Time  `json:"scheduled_date"`
	CompletedDate  *time.Time `json:"completed_date,omitempty"`
	Technician     string     `json:"technician"`
	Cost           float64    `json:"cost"`
	Notes          string     `json:"notes"`
	CertificateRef string     `json:"certificate_ref"`
	CreatedAt      time.Time  `json:"created_at"`
}

type MaintenanceScheduleDTO struct {
	ID            string     `json:"id"`
	EquipmentID   string     `json:"equipment_id"`
	IntervalDays  int        `json:"interval_days"`
	LastPerformed *time.Time `json:"last_performed,omitempty"`
	NextDue       time.Time  `json:"next_due"`
}

func RecordToDTO(r *domain.MaintenanceRecord) *MaintenanceRecordDTO {
	return &MaintenanceRecordDTO{
		ID:             r.ID,
		EquipmentID:    r.EquipmentID,
		Type:           string(r.Type),
		Status:         string(r.Status),
		ScheduledDate:  r.ScheduledDate,
		CompletedDate:  r.CompletedDate,
		Technician:     r.Technician,
		Cost:           r.Cost,
		Notes:          r.Notes,
		CertificateRef: r.CertificateRef,
		CreatedAt:      r.CreatedAt,
	}
}

func ScheduleToDTO(s *domain.MaintenanceSchedule) *MaintenanceScheduleDTO {
	return &MaintenanceScheduleDTO{
		ID:            s.ID,
		EquipmentID:   s.EquipmentID,
		IntervalDays:  s.IntervalDays,
		LastPerformed: s.LastPerformed,
		NextDue:       s.NextDue,
	}
}

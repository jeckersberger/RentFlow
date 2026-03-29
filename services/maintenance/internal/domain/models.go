package domain

import (
	"time"

	"github.com/google/uuid"
)

type MaintenanceSchedule struct {
	ID              uuid.UUID  `json:"id"`
	TenantID        uuid.UUID  `json:"tenant_id"`
	EquipmentID     uuid.UUID  `json:"equipment_id"`
	Name            string     `json:"name"`
	IntervalDays    int        `json:"interval_days"`
	LastPerformedAt *time.Time `json:"last_performed_at,omitempty"`
	NextDueAt       *time.Time `json:"next_due_at,omitempty"`
	IsActive        bool       `json:"is_active"`
	Notes           string     `json:"notes,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

type MaintenanceTask struct {
	ID          uuid.UUID  `json:"id"`
	TenantID    uuid.UUID  `json:"tenant_id"`
	ScheduleID  *uuid.UUID `json:"schedule_id,omitempty"`
	EquipmentID uuid.UUID  `json:"equipment_id"`
	Title       string     `json:"title"`
	Description string     `json:"description,omitempty"`
	Status      string     `json:"status"`
	Priority    string     `json:"priority"`
	AssignedTo  *uuid.UUID `json:"assigned_to,omitempty"`
	DueDate     string     `json:"due_date,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	Cost        int64      `json:"cost"`
	Notes       string     `json:"notes,omitempty"`
	CreatedBy   *uuid.UUID `json:"created_by,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type MaintenanceLog struct {
	ID          uuid.UUID  `json:"id"`
	TaskID      uuid.UUID  `json:"task_id"`
	TenantID    uuid.UUID  `json:"tenant_id"`
	Action      string     `json:"action"`
	PerformedBy *uuid.UUID `json:"performed_by,omitempty"`
	Notes       string     `json:"notes,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

// ECheckRecord represents an electrical safety check (E-Check / DGUV V3).
type ECheckRecord struct {
	ID                              uuid.UUID  `json:"id"`
	TenantID                        uuid.UUID  `json:"tenant_id"`
	EquipmentID                     uuid.UUID  `json:"equipment_id"`
	CheckDate                       string     `json:"check_date"`
	NextCheckDate                   *string    `json:"next_check_date,omitempty"`
	Result                          string     `json:"result"`
	PerformedBy                     string     `json:"performed_by,omitempty"`
	MeasuringDevice                 string     `json:"measuring_device,omitempty"`
	ProtectionConductorResistance   *float64   `json:"protection_conductor_resistance,omitempty"`
	InsulationResistance            *float64   `json:"insulation_resistance,omitempty"`
	LeakageCurrent                  *float64   `json:"leakage_current,omitempty"`
	Notes                           string     `json:"notes,omitempty"`
	CreatedAt                       time.Time  `json:"created_at"`
}

// ECheckFilter holds optional filter criteria for listing E-Check records.
type ECheckFilter struct {
	Page        int
	PerPage     int
	EquipmentID *uuid.UUID
}

type ScheduleFilter struct {
	Page        int
	PerPage     int
	EquipmentID *uuid.UUID
}

type TaskFilter struct {
	Page        int
	PerPage     int
	EquipmentID *uuid.UUID
	Status      string
}

type LogFilter struct {
	Page    int
	PerPage int
}

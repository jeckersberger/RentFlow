package domain

import "time"

type MaintenanceType string

const (
	MaintenanceTypeScheduled   MaintenanceType = "scheduled"
	MaintenanceTypeUnscheduled MaintenanceType = "unscheduled"
	MaintenanceTypeDGUVV3      MaintenanceType = "dguv_v3"
	MaintenanceTypeRepair      MaintenanceType = "repair"
)

type MaintenanceStatus string

const (
	MaintenanceStatusScheduled  MaintenanceStatus = "scheduled"
	MaintenanceStatusInProgress MaintenanceStatus = "in_progress"
	MaintenanceStatusCompleted  MaintenanceStatus = "completed"
)

type MaintenanceRecord struct {
	ID             string            `json:"id"`
	TenantID       string            `json:"tenant_id"`
	EquipmentID    string            `json:"equipment_id"`
	Type           MaintenanceType   `json:"type"`
	Status         MaintenanceStatus `json:"status"`
	ScheduledDate  time.Time         `json:"scheduled_date"`
	CompletedDate  *time.Time        `json:"completed_date,omitempty"`
	Technician     string            `json:"technician"`
	Cost           float64           `json:"cost"`
	Notes          string            `json:"notes"`
	CertificateRef string            `json:"certificate_ref"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
}

type MaintenanceSchedule struct {
	ID            string     `json:"id"`
	TenantID      string     `json:"tenant_id"`
	EquipmentID   string     `json:"equipment_id"`
	IntervalDays  int        `json:"interval_days"`
	LastPerformed *time.Time `json:"last_performed,omitempty"`
	NextDue       time.Time  `json:"next_due"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

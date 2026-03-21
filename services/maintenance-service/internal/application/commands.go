package application

import "time"

type CreateMaintenanceRecordCommand struct {
	TenantID      string    `json:"tenant_id"`
	EquipmentID   string    `json:"equipment_id"`
	Type          string    `json:"type"`
	ScheduledDate time.Time `json:"scheduled_date"`
	Technician    string    `json:"technician"`
	Cost          float64   `json:"cost"`
	Notes         string    `json:"notes"`
}

type CompleteMaintenanceCommand struct {
	TenantID       string    `json:"tenant_id"`
	RecordID       string    `json:"record_id"`
	CompletedDate  time.Time `json:"completed_date"`
	Notes          string    `json:"notes"`
	CertificateRef string    `json:"certificate_ref"`
}

type CreateScheduleCommand struct {
	TenantID     string `json:"tenant_id"`
	EquipmentID  string `json:"equipment_id"`
	IntervalDays int    `json:"interval_days"`
}

type UpdateScheduleCommand struct {
	TenantID     string `json:"tenant_id"`
	ScheduleID   string `json:"schedule_id"`
	IntervalDays int    `json:"interval_days"`
}

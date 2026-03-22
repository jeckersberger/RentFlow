package application

import (
	"time"
)

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

type CreateMaintenancePlanCommand struct {
	TenantID           string                       `json:"tenant_id"`
	EquipmentID        string                       `json:"equipment_id"`
	PlanType           string                       `json:"plan_type"`
	IntervalDays       *int                         `json:"interval_days,omitempty"`
	IntervalHours      *int                         `json:"interval_hours,omitempty"`
	Name               string                       `json:"name"`
	Description        string                       `json:"description"`
	ChecklistTemplateID *string                     `json:"checklist_template_id,omitempty"`
}

type CreateMaintenanceTaskCommand struct {
	TenantID   string    `json:"tenant_id"`
	PlanID     string    `json:"plan_id"`
	EquipmentID string   `json:"equipment_id"`
	Priority   string    `json:"priority"`
	AssignedTo *string   `json:"assigned_to,omitempty"`
	ScheduledAt time.Time `json:"scheduled_at"`
}

type StartMaintenanceTaskCommand struct {
	TenantID string `json:"tenant_id"`
	TaskID   string `json:"task_id"`
}

type CompleteMaintenanceTaskCommand struct {
	TenantID      string                           `json:"tenant_id"`
	TaskID        string                           `json:"task_id"`
	ChecklistData map[string]interface{}           `json:"checklist_data,omitempty"`
	Notes         string                           `json:"notes"`
}

type CreateChecklistCommand struct {
	TenantID    string                           `json:"tenant_id"`
	Name        string                           `json:"name"`
	Description string                           `json:"description"`
	Items       []map[string]interface{}         `json:"items"`
}

type RecordElectricalTestCommand struct {
	TenantID                    string     `json:"tenant_id"`
	TaskID                      *string    `json:"task_id,omitempty"`
	EquipmentID                 string     `json:"equipment_id"`
	TesterID                    string     `json:"tester_id"`
	TestType                    string     `json:"test_type"`
	TestDate                    time.Time  `json:"test_date"`
	Result                      string     `json:"result"`
	InsulationResistanceMohm    *float64   `json:"insulation_resistance_mohm,omitempty"`
	ProtectiveConductorResistanceOhm *float64 `json:"protective_conductor_resistance_ohm,omitempty"`
	LeakageCurrentMA            *float64   `json:"leakage_current_ma,omitempty"`
	VisualInspectionOK          bool       `json:"visual_inspection_ok"`
	FunctionalTestOK            bool       `json:"functional_test_ok"`
	TestDeviceID                string     `json:"test_device_id"`
	TestDeviceName              string     `json:"test_device_name"`
	CertificateNumber           string     `json:"certificate_number"`
	Notes                       string     `json:"notes"`
}

type IzytronImportCommand struct {
	TenantID string `json:"tenant_id"`
	XMLData  string `json:"xml_data"`
}

package domain

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

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

type PlanType string

const (
	PlanTypeInterval   PlanType = "interval-based"
	PlanTypeAfterUse   PlanType = "after-use"
	PlanTypeHoursBased PlanType = "hours-based"
)

type MaintenancePlan struct {
	ID                 string     `json:"id"`
	TenantID           string     `json:"tenant_id"`
	EquipmentID        string     `json:"equipment_id"`
	PlanType           PlanType   `json:"plan_type"`
	IntervalDays       *int       `json:"interval_days,omitempty"`
	IntervalHours      *int       `json:"interval_hours,omitempty"`
	Name               string     `json:"name"`
	Description        string     `json:"description"`
	ChecklistTemplateID *string   `json:"checklist_template_id,omitempty"`
	LastExecutedAt     *time.Time `json:"last_executed_at,omitempty"`
	NextDueAt          *time.Time `json:"next_due_at,omitempty"`
	IsActive           bool       `json:"is_active"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

type TaskStatus string

const (
	TaskStatusPlanned    TaskStatus = "planned"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusOverdue    TaskStatus = "overdue"
	TaskStatusCancelled  TaskStatus = "cancelled"
)

type TaskPriority string

const (
	PriorityLow      TaskPriority = "low"
	PriorityMedium   TaskPriority = "medium"
	PriorityHigh     TaskPriority = "high"
	PriorityCritical TaskPriority = "critical"
)

type MaintenanceTask struct {
	ID             string           `json:"id"`
	TenantID       string           `json:"tenant_id"`
	PlanID         string           `json:"plan_id"`
	EquipmentID    string           `json:"equipment_id"`
	AssignedTo     *string          `json:"assigned_to,omitempty"`
	Status         TaskStatus       `json:"status"`
	Priority       TaskPriority     `json:"priority"`
	ScheduledAt    time.Time        `json:"scheduled_at"`
	StartedAt      *time.Time       `json:"started_at,omitempty"`
	CompletedAt    *time.Time       `json:"completed_at,omitempty"`
	Notes          string           `json:"notes"`
	ChecklistData  *ChecklistData   `json:"checklist_data,omitempty"`
	CreatedAt      time.Time        `json:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at"`
}

type ChecklistData struct {
	Items []ChecklistItemResult `json:"items"`
}

func (cd ChecklistData) Value() (driver.Value, error) {
	return json.Marshal(cd)
}

type ChecklistItemResult struct {
	ItemID string      `json:"item_id"`
	Name   string      `json:"name"`
	Result interface{} `json:"result"`
	Notes  string      `json:"notes,omitempty"`
}

type Checklist struct {
	ID          string           `json:"id"`
	TenantID    string           `json:"tenant_id"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Items       []ChecklistItem  `json:"items"`
	Version     int              `json:"version"`
	CreatedAt   time.Time        `json:"created_at"`
}

type ChecklistItem struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Description string `json:"description"`
	CheckType string  `json:"check_type"`
	Required  bool    `json:"required"`
	Unit      *string `json:"unit,omitempty"`
	MinValue  *float64 `json:"min_value,omitempty"`
	MaxValue  *float64 `json:"max_value,omitempty"`
}

type TestType string

const (
	TestTypeVDE0701 TestType = "vde_0701"
	TestTypeVDE0702 TestType = "vde_0702"
)

type TestResult string

const (
	TestResultPassed      TestResult = "passed"
	TestResultFailed      TestResult = "failed"
	TestResultConditional TestResult = "conditional"
)

type ElectricalTest struct {
	ID                          string     `json:"id"`
	TenantID                    string     `json:"tenant_id"`
	TaskID                      *string    `json:"task_id,omitempty"`
	EquipmentID                 string     `json:"equipment_id"`
	TesterID                    string     `json:"tester_id"`
	TestType                    TestType   `json:"test_type"`
	TestDate                    time.Time  `json:"test_date"`
	NextTestDate                time.Time  `json:"next_test_date"`
	Result                      TestResult `json:"result"`
	InsulationResistanceMohm    *float64   `json:"insulation_resistance_mohm,omitempty"`
	ProtectiveConductorResistanceOhm *float64 `json:"protective_conductor_resistance_ohm,omitempty"`
	LeakageCurrentMA            *float64   `json:"leakage_current_ma,omitempty"`
	VisualInspectionOK          bool       `json:"visual_inspection_ok"`
	FunctionalTestOK            bool       `json:"functional_test_ok"`
	TestDeviceID                string     `json:"test_device_id"`
	TestDeviceName              string     `json:"test_device_name"`
	CertificateNumber           string     `json:"certificate_number"`
	Notes                       string     `json:"notes"`
	IzytronImportID             *string    `json:"izytron_import_id,omitempty"`
	RawXML                      *string    `json:"raw_xml,omitempty"`
	CreatedAt                   time.Time  `json:"created_at"`
}

package application

import (
	"time"

	"github.com/jeckersberger/rentflow/services/maintenance-service/internal/domain"
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

type MaintenancePlanDTO struct {
	ID                 string     `json:"id"`
	EquipmentID        string     `json:"equipment_id"`
	PlanType           string     `json:"plan_type"`
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

type MaintenanceTaskDTO struct {
	ID            string                 `json:"id"`
	PlanID        string                 `json:"plan_id"`
	EquipmentID   string                 `json:"equipment_id"`
	AssignedTo    *string                `json:"assigned_to,omitempty"`
	Status        string                 `json:"status"`
	Priority      string                 `json:"priority"`
	ScheduledAt   time.Time              `json:"scheduled_at"`
	StartedAt     *time.Time             `json:"started_at,omitempty"`
	CompletedAt   *time.Time             `json:"completed_at,omitempty"`
	Notes         string                 `json:"notes"`
	ChecklistData map[string]interface{} `json:"checklist_data,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

type ChecklistDTO struct {
	ID          string                   `json:"id"`
	Name        string                   `json:"name"`
	Description string                   `json:"description"`
	Items       []map[string]interface{} `json:"items"`
	Version     int                      `json:"version"`
	CreatedAt   time.Time                `json:"created_at"`
}

type ElectricalTestDTO struct {
	ID                          string     `json:"id"`
	TaskID                      *string    `json:"task_id,omitempty"`
	EquipmentID                 string     `json:"equipment_id"`
	TesterID                    string     `json:"tester_id"`
	TestType                    string     `json:"test_type"`
	TestDate                    time.Time  `json:"test_date"`
	NextTestDate                time.Time  `json:"next_test_date"`
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
	IzytronImportID             *string    `json:"izytron_import_id,omitempty"`
	CreatedAt                   time.Time  `json:"created_at"`
}

type MaintenanceDashboardDTO struct {
	DueTasks     []*MaintenanceTaskDTO  `json:"due_tasks"`
	OverdueTasks []*MaintenanceTaskDTO  `json:"overdue_tasks"`
	DuePlans     []*MaintenancePlanDTO  `json:"due_plans"`
	RecentTests  []*ElectricalTestDTO   `json:"recent_tests"`
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

func PlanToDTO(p *domain.MaintenancePlan) *MaintenancePlanDTO {
	return &MaintenancePlanDTO{
		ID:                 p.ID,
		EquipmentID:        p.EquipmentID,
		PlanType:           string(p.PlanType),
		IntervalDays:       p.IntervalDays,
		IntervalHours:      p.IntervalHours,
		Name:               p.Name,
		Description:        p.Description,
		ChecklistTemplateID: p.ChecklistTemplateID,
		LastExecutedAt:     p.LastExecutedAt,
		NextDueAt:          p.NextDueAt,
		IsActive:           p.IsActive,
		CreatedAt:          p.CreatedAt,
		UpdatedAt:          p.UpdatedAt,
	}
}

func TaskToDTO(t *domain.MaintenanceTask) *MaintenanceTaskDTO {
	var checklistMap map[string]interface{}
	if t.ChecklistData != nil {
		checklistMap = map[string]interface{}{"items": t.ChecklistData.Items}
	}
	return &MaintenanceTaskDTO{
		ID:            t.ID,
		PlanID:        t.PlanID,
		EquipmentID:   t.EquipmentID,
		AssignedTo:    t.AssignedTo,
		Status:        string(t.Status),
		Priority:      string(t.Priority),
		ScheduledAt:   t.ScheduledAt,
		StartedAt:     t.StartedAt,
		CompletedAt:   t.CompletedAt,
		Notes:         t.Notes,
		ChecklistData: checklistMap,
		CreatedAt:     t.CreatedAt,
		UpdatedAt:     t.UpdatedAt,
	}
}

func ChecklistToDTO(c *domain.Checklist) *ChecklistDTO {
	items := make([]map[string]interface{}, len(c.Items))
	for i, item := range c.Items {
		items[i] = map[string]interface{}{
			"id":           item.ID,
			"name":         item.Name,
			"description":  item.Description,
			"check_type":   item.CheckType,
			"required":     item.Required,
			"unit":         item.Unit,
			"min_value":    item.MinValue,
			"max_value":    item.MaxValue,
		}
	}
	return &ChecklistDTO{
		ID:          c.ID,
		Name:        c.Name,
		Description: c.Description,
		Items:       items,
		Version:     c.Version,
		CreatedAt:   c.CreatedAt,
	}
}

func TestToDTO(t *domain.ElectricalTest) *ElectricalTestDTO {
	return &ElectricalTestDTO{
		ID:                          t.ID,
		TaskID:                      t.TaskID,
		EquipmentID:                 t.EquipmentID,
		TesterID:                    t.TesterID,
		TestType:                    string(t.TestType),
		TestDate:                    t.TestDate,
		NextTestDate:                t.NextTestDate,
		Result:                      string(t.Result),
		InsulationResistanceMohm:    t.InsulationResistanceMohm,
		ProtectiveConductorResistanceOhm: t.ProtectiveConductorResistanceOhm,
		LeakageCurrentMA:            t.LeakageCurrentMA,
		VisualInspectionOK:          t.VisualInspectionOK,
		FunctionalTestOK:            t.FunctionalTestOK,
		TestDeviceID:                t.TestDeviceID,
		TestDeviceName:              t.TestDeviceName,
		CertificateNumber:           t.CertificateNumber,
		Notes:                       t.Notes,
		IzytronImportID:             t.IzytronImportID,
		CreatedAt:                   t.CreatedAt,
	}
}

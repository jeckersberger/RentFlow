package domain

import "time"

type ReportType string

const (
	ReportTypeRevenue    ReportType = "revenue"
	ReportTypeUtilization ReportType = "utilization"
	ReportTypeInventory  ReportType = "inventory"
	ReportTypeProject    ReportType = "project"
)

type ReportStatus string

const (
	ReportStatusPending   ReportStatus = "pending"
	ReportStatusGenerating ReportStatus = "generating"
	ReportStatusCompleted ReportStatus = "completed"
	ReportStatusFailed    ReportStatus = "failed"
)

type Report struct {
	ID          string       `json:"id"`
	TenantID    string       `json:"tenant_id"`
	Type        ReportType   `json:"type"`
	Parameters  map[string]interface{} `json:"parameters"`
	Status      ReportStatus `json:"status"`
	GeneratedAt *time.Time   `json:"generated_at,omitempty"`
	FileRef     string       `json:"file_ref,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

type KPI struct {
	Name         string  `json:"name"`
	Value        float64 `json:"value"`
	PreviousValue float64 `json:"previous_value"`
	ChangePercent float64 `json:"change_percent"`
	Period       string  `json:"period"`
}

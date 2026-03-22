package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type ReportType string

const (
	ReportTypeRevenue    ReportType = "revenue"
	ReportTypeUtilization ReportType = "utilization"
	ReportTypeDunning    ReportType = "dunning"
	ReportTypeInventory  ReportType = "inventory"
	ReportTypeMaintenance ReportType = "maintenance"
	ReportTypeCrewHours  ReportType = "crew_hours"
)

type ReportFormat string

const (
	ReportFormatPDF  ReportFormat = "pdf"
	ReportFormatCSV  ReportFormat = "csv"
	ReportFormatBoth ReportFormat = "both"
)

type ReportDefinition struct {
	ID              uuid.UUID           `db:"id"`
	TenantID        uuid.UUID           `db:"tenant_id"`
	Name            string              `db:"name"`
	ReportType      ReportType          `db:"report_type"`
	Description     string              `db:"description"`
	Parameters      json.RawMessage     `db:"parameters"`
	ScheduleCron    *string             `db:"schedule_cron"`
	EmailRecipients json.RawMessage     `db:"email_recipients"`
	Format          ReportFormat        `db:"format"`
	IsActive        bool                `db:"is_active"`
	CreatedBy       uuid.UUID           `db:"created_by"`
	CreatedAt       time.Time           `db:"created_at"`
	UpdatedAt       time.Time           `db:"updated_at"`
}

type ReportRunStatus string

const (
	ReportRunStatusQueued    ReportRunStatus = "queued"
	ReportRunStatusRunning   ReportRunStatus = "running"
	ReportRunStatusCompleted ReportRunStatus = "completed"
	ReportRunStatusFailed    ReportRunStatus = "failed"
)

type ReportRun struct {
	ID                uuid.UUID           `db:"id"`
	TenantID          uuid.UUID           `db:"tenant_id"`
	ReportDefinitionID uuid.UUID           `db:"report_definition_id"`
	Status            ReportRunStatus     `db:"status"`
	ParametersUsed    json.RawMessage     `db:"parameters_used"`
	PeriodStart       *time.Time          `db:"period_start"`
	PeriodEnd         *time.Time          `db:"period_end"`
	FilePath          *string             `db:"file_path"`
	FileSize          *int64              `db:"file_size"`
	ErrorMessage      *string             `db:"error_message"`
	StartedAt         *time.Time          `db:"started_at"`
	CompletedAt       *time.Time          `db:"completed_at"`
	CreatedAt         time.Time           `db:"created_at"`
}

type Period string

const (
	PeriodWeek    Period = "week"
	PeriodMonth   Period = "month"
	PeriodQuarter Period = "quarter"
	PeriodYear    Period = "year"
	PeriodCustom  Period = "custom"
)

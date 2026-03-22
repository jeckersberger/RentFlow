package application

import (
	"encoding/json"

	"github.com/google/uuid"
)

type CreateReportDefinitionCommand struct {
	TenantID        uuid.UUID
	Name            string
	ReportType      string
	Description     string
	Parameters      json.RawMessage
	ScheduleCron    *string
	EmailRecipients json.RawMessage
	Format          string
	CreatedBy       uuid.UUID
}

type UpdateReportDefinitionCommand struct {
	ReportDefinitionID uuid.UUID
	TenantID           uuid.UUID
	Name               string
	Description        string
	Parameters         json.RawMessage
	ScheduleCron       *string
	EmailRecipients    json.RawMessage
	Format             string
	IsActive           bool
}

type GenerateReportCommand struct {
	ReportDefinitionID uuid.UUID
	TenantID           uuid.UUID
	Parameters         json.RawMessage
	PeriodStart        *string
	PeriodEnd          *string
}

type SnapshotKPIsCommand struct {
	TenantID uuid.UUID
}

type GetKPIDashboardCommand struct {
	TenantID uuid.UUID
	Period   string
}

type GetKPITrendsCommand struct {
	TenantID uuid.UUID
	KPIType  string
	Periods  int
}

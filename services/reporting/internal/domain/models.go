package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type ReportDefinition struct {
	ID          uuid.UUID       `json:"id"`
	TenantID    uuid.UUID       `json:"tenant_id"`
	Name        string          `json:"name"`
	Type        string          `json:"type"`
	QueryConfig json.RawMessage `json:"query_config"`
	Schedule    string          `json:"schedule"`
	Format      string          `json:"format"`
	IsActive    bool            `json:"is_active"`
	CreatedBy   *uuid.UUID      `json:"created_by,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type ReportSnapshot struct {
	ID           uuid.UUID       `json:"id"`
	DefinitionID uuid.UUID       `json:"definition_id"`
	TenantID     uuid.UUID       `json:"tenant_id"`
	Title        string          `json:"title"`
	Data         json.RawMessage `json:"data"`
	FilePath     string          `json:"file_path"`
	FileSize     int             `json:"file_size"`
	Format       string          `json:"format"`
	PeriodStart  string          `json:"period_start"`
	PeriodEnd    string          `json:"period_end"`
	GeneratedBy  *uuid.UUID      `json:"generated_by,omitempty"`
	GeneratedAt  time.Time       `json:"generated_at"`
}

type DashboardWidget struct {
	ID        uuid.UUID       `json:"id"`
	TenantID  uuid.UUID       `json:"tenant_id"`
	Name      string          `json:"name"`
	Type      string          `json:"type"`
	Config    json.RawMessage `json:"config"`
	Position  int             `json:"position"`
	IsActive  bool            `json:"is_active"`
	CreatedBy *uuid.UUID      `json:"created_by,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

type DefinitionFilter struct {
	Page    int
	PerPage int
	Type    string
}

type SnapshotFilter struct {
	Page         int
	PerPage      int
	DefinitionID *uuid.UUID
}

type WidgetFilter struct {
	Page    int
	PerPage int
}

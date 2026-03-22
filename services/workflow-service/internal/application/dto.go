package application

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type CreateWorkflowDefinitionRequest struct {
	Name             string          `json:"name"`
	Description      string          `json:"description"`
	TriggerType      string          `json:"trigger_type"`
	TriggerConfig    json.RawMessage `json:"trigger_config"`
	IsTemplate       bool            `json:"is_template"`
	TemplateCategory string          `json:"template_category"`
}

type WorkflowDefinitionResponse struct {
	ID               uuid.UUID       `json:"id"`
	TenantID         uuid.UUID       `json:"tenant_id"`
	Name             string          `json:"name"`
	Description      string          `json:"description"`
	TriggerType      string          `json:"trigger_type"`
	TriggerConfig    json.RawMessage `json:"trigger_config"`
	IsActive         bool            `json:"is_active"`
	IsTemplate       bool            `json:"is_template"`
	TemplateCategory string          `json:"template_category"`
	Version          int             `json:"version"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

type InstantiateWorkflowRequest struct {
	TriggerData json.RawMessage `json:"trigger_data"`
}

type WorkflowInstanceResponse struct {
	ID           uuid.UUID       `json:"id"`
	TenantID     uuid.UUID       `json:"tenant_id"`
	DefinitionID uuid.UUID       `json:"definition_id"`
	Status       string          `json:"status"`
	TriggerData  json.RawMessage `json:"trigger_data"`
	ContextData  json.RawMessage `json:"context_data"`
	StepIndex    int             `json:"current_step_index"`
	StartedAt    time.Time       `json:"started_at"`
	CompletedAt  *time.Time      `json:"completed_at"`
	ErrorMessage string          `json:"error_message"`
}

type TriggerWorkflowRequest struct {
	EventName string          `json:"event_name"`
	Data      json.RawMessage `json:"data"`
}

type DashboardResponse struct {
	ActiveCount      int `json:"active_count"`
	CompletedToday   int `json:"completed_today"`
	FailedCount      int `json:"failed_count"`
}

type CreateWorkflowStepRequest struct {
	StepIndex     int             `json:"step_index"`
	Name          string          `json:"name"`
	ActionType    string          `json:"action_type"`
	ActionConfig  json.RawMessage `json:"action_config"`
	OnSuccessStep *int            `json:"on_success_step"`
	OnFailureStep *int            `json:"on_failure_step"`
	TimeoutSecs   int             `json:"timeout_seconds"`
	RetryCount    int             `json:"retry_count"`
}

type WorkflowStepResponse struct {
	ID            uuid.UUID       `json:"id"`
	DefinitionID  uuid.UUID       `json:"definition_id"`
	StepIndex     int             `json:"step_index"`
	Name          string          `json:"name"`
	ActionType    string          `json:"action_type"`
	ActionConfig  json.RawMessage `json:"action_config"`
	OnSuccessStep *int            `json:"on_success_step"`
	OnFailureStep *int            `json:"on_failure_step"`
	TimeoutSecs   int             `json:"timeout_seconds"`
	RetryCount    int             `json:"retry_count"`
	CreatedAt     time.Time       `json:"created_at"`
}

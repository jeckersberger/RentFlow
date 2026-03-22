package domain

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// TriggerType defines how a workflow is triggered
type TriggerType string

const (
	TriggerTypeEvent  TriggerType = "event"
	TriggerTypeCron   TriggerType = "cron"
	TriggerTypeManual TriggerType = "manual"
)

// Status represents workflow instance status
type Status string

const (
	StatusRunning   Status = "running"
	StatusCompleted Status = "completed"
	StatusFailed    Status = "failed"
	StatusCancelled Status = "cancelled"
)

// ActionType defines the type of action a workflow step performs
type ActionType string

const (
	ActionTypeEmail        ActionType = "email"
	ActionTypeWebhook      ActionType = "webhook"
	ActionTypeServiceCall  ActionType = "service_call"
	ActionTypeStatusChange ActionType = "status_change"
	ActionTypeDelay        ActionType = "delay"
	ActionTypeCondition    ActionType = "condition"
)

// WorkflowDefinition represents a workflow template
type WorkflowDefinition struct {
	ID               uuid.UUID       `json:"id"`
	TenantID         uuid.UUID       `json:"tenant_id"`
	Name             string          `json:"name"`
	Description      string          `json:"description"`
	TriggerType      TriggerType     `json:"trigger_type"`
	TriggerConfig    json.RawMessage `json:"trigger_config"`
	IsActive         bool            `json:"is_active"`
	IsTemplate       bool            `json:"is_template"`
	TemplateCategory string          `json:"template_category"`
	Version          int             `json:"version"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

// WorkflowInstance represents a running workflow
type WorkflowInstance struct {
	ID              uuid.UUID       `json:"id"`
	TenantID        uuid.UUID       `json:"tenant_id"`
	DefinitionID    uuid.UUID       `json:"definition_id"`
	Status          Status          `json:"status"`
	TriggerData     json.RawMessage `json:"trigger_data"`
	ContextData     json.RawMessage `json:"context_data"`
	CurrentStepIdx  int             `json:"current_step_index"`
	StartedAt       time.Time       `json:"started_at"`
	CompletedAt     *time.Time      `json:"completed_at"`
	ErrorMessage    string          `json:"error_message"`
}

// WorkflowStep represents a single step in a workflow
type WorkflowStep struct {
	ID            uuid.UUID       `json:"id"`
	DefinitionID  uuid.UUID       `json:"definition_id"`
	StepIndex     int             `json:"step_index"`
	Name          string          `json:"name"`
	ActionType    ActionType      `json:"action_type"`
	ActionConfig  json.RawMessage `json:"action_config"`
	OnSuccessStep *int            `json:"on_success_step"`
	OnFailureStep *int            `json:"on_failure_step"`
	TimeoutSecs   int             `json:"timeout_seconds"`
	RetryCount    int             `json:"retry_count"`
	CreatedAt     time.Time       `json:"created_at"`
}

// Implement driver.Valuer and sql.Scanner for custom types
func (t TriggerType) Value() (driver.Value, error) {
	return string(t), nil
}

func (s Status) Value() (driver.Value, error) {
	return string(s), nil
}

func (a ActionType) Value() (driver.Value, error) {
	return string(a), nil
}

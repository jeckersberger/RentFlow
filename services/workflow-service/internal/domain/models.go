package domain

import (
	"encoding/json"
	"errors"
	"time"
)

var (
	ErrWorkflowNotFound  = errors.New("workflow not found")
	ErrWorkflowRunNotFound = errors.New("workflow run not found")
	ErrInvalidInput      = errors.New("invalid input")
	ErrTenantIDRequired  = errors.New("tenant ID required")
)

type StepType string

const (
	StepTypeCondition    StepType = "condition"
	StepTypeAction       StepType = "action"
	StepTypeDelay        StepType = "delay"
	StepTypeNotification StepType = "notification"
)

type WorkflowRunStatus string

const (
	RunStatusRunning   WorkflowRunStatus = "running"
	RunStatusCompleted WorkflowRunStatus = "completed"
	RunStatusFailed    WorkflowRunStatus = "failed"
	RunStatusCancelled WorkflowRunStatus = "cancelled"
)

type Workflow struct {
	ID             string              `json:"id"`
	TenantID       string              `json:"tenant_id"`
	Name           string              `json:"name"`
	Description    string              `json:"description"`
	TriggerEvent   string              `json:"trigger_event"`
	Steps          []*WorkflowStep     `json:"steps"`
	IsActive       bool                `json:"is_active"`
	Version        int                 `json:"version"`
	CreatedAt      time.Time           `json:"created_at"`
	UpdatedAt      time.Time           `json:"updated_at"`
}

type WorkflowStep struct {
	ID               string          `json:"id"`
	Type             StepType        `json:"type"`
	Config           json.RawMessage `json:"config"`
	NextStepOnSuccess string         `json:"next_step_on_success"`
	NextStepOnFailure string         `json:"next_step_on_failure"`
}

type WorkflowRun struct {
	ID             string          `json:"id"`
	WorkflowID     string          `json:"workflow_id"`
	TriggerEventID string          `json:"trigger_event_id"`
	Status         WorkflowRunStatus `json:"status"`
	CurrentStep    string          `json:"current_step"`
	Context        json.RawMessage `json:"context"`
	StartedAt      time.Time       `json:"started_at"`
	CompletedAt    *time.Time      `json:"completed_at"`
	Error          string          `json:"error"`
}

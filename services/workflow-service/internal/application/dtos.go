package application

import (
	"encoding/json"
	"time"

	"github.com/jeckersberger/rentflow/services/workflow-service/internal/domain"
)

type CreateWorkflowRequest struct {
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	TriggerEvent string                 `json:"trigger_event"`
	Steps        []WorkflowStepRequest  `json:"steps"`
}

type WorkflowStepRequest struct {
	Type              string          `json:"type"`
	Config            json.RawMessage `json:"config"`
	NextStepOnSuccess string          `json:"next_step_on_success"`
	NextStepOnFailure string          `json:"next_step_on_failure"`
}

type WorkflowResponse struct {
	ID            string                   `json:"id"`
	Name          string                   `json:"name"`
	Description   string                   `json:"description"`
	TriggerEvent  string                   `json:"trigger_event"`
	Steps         []WorkflowStepResponse   `json:"steps"`
	IsActive      bool                     `json:"is_active"`
	Version       int                      `json:"version"`
	CreatedAt     time.Time                `json:"created_at"`
	UpdatedAt     time.Time                `json:"updated_at"`
}

type WorkflowStepResponse struct {
	ID                string          `json:"id"`
	Type              string          `json:"type"`
	Config            json.RawMessage `json:"config"`
	NextStepOnSuccess string          `json:"next_step_on_success"`
	NextStepOnFailure string          `json:"next_step_on_failure"`
}

type TriggerWorkflowRequest struct {
	EventType string                 `json:"event_type"`
	EventData map[string]interface{} `json:"event_data"`
}

type TriggerWorkflowResponse struct {
	RunID     string    `json:"run_id"`
	Status    string    `json:"status"`
	StartedAt time.Time `json:"started_at"`
}

type WorkflowRunResponse struct {
	ID            string     `json:"id"`
	WorkflowID    string     `json:"workflow_id"`
	TriggerEventID string    `json:"trigger_event_id"`
	Status        string     `json:"status"`
	CurrentStep   string     `json:"current_step"`
	StartedAt     time.Time  `json:"started_at"`
	CompletedAt   *time.Time `json:"completed_at"`
	Error         string     `json:"error"`
}

type ActivateWorkflowRequest struct {
	Activate bool `json:"activate"`
}

type CancelRunRequest struct {
	Reason string `json:"reason"`
}

func WorkflowToDTO(w *domain.Workflow) *WorkflowResponse {
	steps := make([]WorkflowStepResponse, len(w.Steps))
	for i, s := range w.Steps {
		steps[i] = WorkflowStepResponse{
			ID:                s.ID,
			Type:              string(s.Type),
			Config:            s.Config,
			NextStepOnSuccess: s.NextStepOnSuccess,
			NextStepOnFailure: s.NextStepOnFailure,
		}
	}

	return &WorkflowResponse{
		ID:           w.ID,
		Name:         w.Name,
		Description:  w.Description,
		TriggerEvent: w.TriggerEvent,
		Steps:        steps,
		IsActive:     w.IsActive,
		Version:      w.Version,
		CreatedAt:    w.CreatedAt,
		UpdatedAt:    w.UpdatedAt,
	}
}

func WorkflowRunToDTO(r *domain.WorkflowRun) *WorkflowRunResponse {
	return &WorkflowRunResponse{
		ID:             r.ID,
		WorkflowID:     r.WorkflowID,
		TriggerEventID: r.TriggerEventID,
		Status:         string(r.Status),
		CurrentStep:    r.CurrentStep,
		StartedAt:      r.StartedAt,
		CompletedAt:    r.CompletedAt,
		Error:          r.Error,
	}
}

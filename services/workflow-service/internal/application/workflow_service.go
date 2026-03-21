package application

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/workflow-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/workflow-service/internal/ports"
)

type WorkflowService struct {
	workflowRepo ports.WorkflowRepository
	runRepo      ports.WorkflowRunRepository
	logger       logger.Logger
}

func NewWorkflowService(
	workflowRepo ports.WorkflowRepository,
	runRepo ports.WorkflowRunRepository,
	log logger.Logger,
) *WorkflowService {
	return &WorkflowService{
		workflowRepo: workflowRepo,
		runRepo:      runRepo,
		logger:       log,
	}
}

func (s *WorkflowService) CreateWorkflow(ctx context.Context, tenantID string, req CreateWorkflowRequest) (*WorkflowResponse, error) {
	if tenantID == "" {
		return nil, domain.ErrTenantIDRequired
	}
	if req.Name == "" {
		return nil, domain.ErrInvalidInput
	}

	steps := make([]*domain.WorkflowStep, len(req.Steps))
	for i, stepReq := range req.Steps {
		steps[i] = &domain.WorkflowStep{
			ID:                fmt.Sprintf("step_%d_%d", time.Now().UnixNano(), i),
			Type:              domain.StepType(stepReq.Type),
			Config:            stepReq.Config,
			NextStepOnSuccess: stepReq.NextStepOnSuccess,
			NextStepOnFailure: stepReq.NextStepOnFailure,
		}
	}

	workflow := &domain.Workflow{
		ID:           fmt.Sprintf("wf_%d", time.Now().UnixNano()),
		TenantID:     tenantID,
		Name:         req.Name,
		Description:  req.Description,
		TriggerEvent: req.TriggerEvent,
		Steps:        steps,
		IsActive:     true,
		Version:      1,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.workflowRepo.Create(ctx, workflow); err != nil {
		s.logger.Error("Failed to create workflow", err)
		return nil, err
	}

	return WorkflowToDTO(workflow), nil
}

func (s *WorkflowService) GetWorkflow(ctx context.Context, tenantID, id string) (*WorkflowResponse, error) {
	if tenantID == "" || id == "" {
		return nil, domain.ErrInvalidInput
	}

	workflow, err := s.workflowRepo.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	if workflow == nil {
		return nil, domain.ErrWorkflowNotFound
	}

	return WorkflowToDTO(workflow), nil
}

func (s *WorkflowService) ListWorkflows(ctx context.Context, tenantID string) ([]*WorkflowResponse, error) {
	if tenantID == "" {
		return nil, domain.ErrTenantIDRequired
	}

	workflows, err := s.workflowRepo.ListByTenant(ctx, tenantID)
	if err != nil {
		s.logger.Error("Failed to list workflows", err)
		return nil, err
	}

	dtos := make([]*WorkflowResponse, len(workflows))
	for i, wf := range workflows {
		dtos[i] = WorkflowToDTO(wf)
	}

	return dtos, nil
}

func (s *WorkflowService) UpdateWorkflow(ctx context.Context, tenantID, id string, req CreateWorkflowRequest) (*WorkflowResponse, error) {
	if tenantID == "" || id == "" {
		return nil, domain.ErrInvalidInput
	}

	workflow, err := s.workflowRepo.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	if workflow == nil {
		return nil, domain.ErrWorkflowNotFound
	}

	workflow.Name = req.Name
	workflow.Description = req.Description
	workflow.TriggerEvent = req.TriggerEvent
	workflow.Version++
	workflow.UpdatedAt = time.Now()

	steps := make([]*domain.WorkflowStep, len(req.Steps))
	for i, stepReq := range req.Steps {
		steps[i] = &domain.WorkflowStep{
			ID:                fmt.Sprintf("step_%d_%d", time.Now().UnixNano(), i),
			Type:              domain.StepType(stepReq.Type),
			Config:            stepReq.Config,
			NextStepOnSuccess: stepReq.NextStepOnSuccess,
			NextStepOnFailure: stepReq.NextStepOnFailure,
		}
	}
	workflow.Steps = steps

	if err := s.workflowRepo.Update(ctx, workflow); err != nil {
		s.logger.Error("Failed to update workflow", err)
		return nil, err
	}

	return WorkflowToDTO(workflow), nil
}

func (s *WorkflowService) DeleteWorkflow(ctx context.Context, tenantID, id string) error {
	if tenantID == "" || id == "" {
		return domain.ErrInvalidInput
	}

	return s.workflowRepo.Delete(ctx, tenantID, id)
}

func (s *WorkflowService) ActivateWorkflow(ctx context.Context, tenantID, id string, activate bool) (*WorkflowResponse, error) {
	if tenantID == "" || id == "" {
		return nil, domain.ErrInvalidInput
	}

	workflow, err := s.workflowRepo.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	if workflow == nil {
		return nil, domain.ErrWorkflowNotFound
	}

	workflow.IsActive = activate
	workflow.UpdatedAt = time.Now()

	if err := s.workflowRepo.Update(ctx, workflow); err != nil {
		s.logger.Error("Failed to update workflow activation", err)
		return nil, err
	}

	return WorkflowToDTO(workflow), nil
}

func (s *WorkflowService) TriggerWorkflow(ctx context.Context, tenantID, workflowID string, req TriggerWorkflowRequest) (*TriggerWorkflowResponse, error) {
	if tenantID == "" || workflowID == "" {
		return nil, domain.ErrInvalidInput
	}

	workflow, err := s.workflowRepo.GetByID(ctx, tenantID, workflowID)
	if err != nil {
		return nil, err
	}
	if workflow == nil {
		return nil, domain.ErrWorkflowNotFound
	}

	if !workflow.IsActive {
		return nil, fmt.Errorf("workflow is not active")
	}

	eventData, _ := json.Marshal(req.EventData)

	run := &domain.WorkflowRun{
		ID:             fmt.Sprintf("run_%d", time.Now().UnixNano()),
		WorkflowID:     workflowID,
		TriggerEventID: fmt.Sprintf("evt_%d", time.Now().UnixNano()),
		Status:         domain.RunStatusRunning,
		CurrentStep:    workflow.Steps[0].ID,
		Context:        eventData,
		StartedAt:      time.Now(),
	}

	if err := s.runRepo.Create(ctx, run); err != nil {
		s.logger.Error("Failed to create workflow run", err)
		return nil, err
	}

	// In production, would asynchronously execute workflow
	return &TriggerWorkflowResponse{
		RunID:     run.ID,
		Status:    string(run.Status),
		StartedAt: run.StartedAt,
	}, nil
}

func (s *WorkflowService) GetWorkflowRun(ctx context.Context, runID string) (*WorkflowRunResponse, error) {
	if runID == "" {
		return nil, domain.ErrInvalidInput
	}

	run, err := s.runRepo.GetByID(ctx, runID)
	if err != nil {
		return nil, err
	}
	if run == nil {
		return nil, domain.ErrWorkflowRunNotFound
	}

	return WorkflowRunToDTO(run), nil
}

func (s *WorkflowService) ListWorkflowRuns(ctx context.Context, tenantID string) ([]*WorkflowRunResponse, error) {
	if tenantID == "" {
		return nil, domain.ErrTenantIDRequired
	}

	runs, err := s.runRepo.ListByTenant(ctx, tenantID)
	if err != nil {
		s.logger.Error("Failed to list workflow runs", err)
		return nil, err
	}

	dtos := make([]*WorkflowRunResponse, len(runs))
	for i, run := range runs {
		dtos[i] = WorkflowRunToDTO(run)
	}

	return dtos, nil
}

func (s *WorkflowService) CancelWorkflowRun(ctx context.Context, runID, reason string) (*WorkflowRunResponse, error) {
	if runID == "" {
		return nil, domain.ErrInvalidInput
	}

	run, err := s.runRepo.GetByID(ctx, runID)
	if err != nil {
		return nil, err
	}
	if run == nil {
		return nil, domain.ErrWorkflowRunNotFound
	}

	run.Status = domain.RunStatusCancelled
	run.Error = reason
	now := time.Now()
	run.CompletedAt = &now

	if err := s.runRepo.Update(ctx, run); err != nil {
		s.logger.Error("Failed to cancel workflow run", err)
		return nil, err
	}

	return WorkflowRunToDTO(run), nil
}

type WorkflowEngine struct {
	workflowRepo ports.WorkflowRepository
	runRepo      ports.WorkflowRunRepository
	logger       logger.Logger
}

func NewWorkflowEngine(
	workflowRepo ports.WorkflowRepository,
	runRepo ports.WorkflowRunRepository,
	log logger.Logger,
) *WorkflowEngine {
	return &WorkflowEngine{
		workflowRepo: workflowRepo,
		runRepo:      runRepo,
		logger:       log,
	}
}

func (e *WorkflowEngine) ExecuteStep(ctx context.Context, run *domain.WorkflowRun, step *domain.WorkflowStep) (bool, string, error) {
	switch step.Type {
	case domain.StepTypeCondition:
		return e.evaluateCondition(ctx, step)
	case domain.StepTypeAction:
		return e.executeAction(ctx, step)
	case domain.StepTypeDelay:
		return e.executeDelay(ctx, step)
	case domain.StepTypeNotification:
		return e.sendNotification(ctx, step)
	default:
		return false, "", fmt.Errorf("unknown step type: %s", step.Type)
	}
}

func (e *WorkflowEngine) evaluateCondition(ctx context.Context, step *domain.WorkflowStep) (bool, string, error) {
	// Simulate condition evaluation
	return true, step.NextStepOnSuccess, nil
}

func (e *WorkflowEngine) executeAction(ctx context.Context, step *domain.WorkflowStep) (bool, string, error) {
	// Simulate action execution
	return true, step.NextStepOnSuccess, nil
}

func (e *WorkflowEngine) executeDelay(ctx context.Context, step *domain.WorkflowStep) (bool, string, error) {
	// Simulate delay
	time.Sleep(100 * time.Millisecond)
	return true, step.NextStepOnSuccess, nil
}

func (e *WorkflowEngine) sendNotification(ctx context.Context, step *domain.WorkflowStep) (bool, string, error) {
	// Simulate notification sending
	return true, step.NextStepOnSuccess, nil
}

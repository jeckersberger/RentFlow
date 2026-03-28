package application

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/services/workflow/internal/domain"
)

// ---------------------------------------------------------------------------
// Request / Response DTOs
// ---------------------------------------------------------------------------

// CreateInstanceRequest holds the data for creating a workflow instance.
type CreateInstanceRequest struct {
	DefinitionID  uuid.UUID        `json:"definition_id"`
	ReferenceID   *uuid.UUID       `json:"reference_id,omitempty"`
	ReferenceType string           `json:"reference_type,omitempty"`
	Data          json.RawMessage  `json:"data,omitempty"`
	Notes         string           `json:"notes,omitempty"`
}

// PerformActionRequest holds the data for performing an action on a workflow instance step.
type PerformActionRequest struct {
	Action  string `json:"action"`
	Comment string `json:"comment,omitempty"`
}

// ---------------------------------------------------------------------------
// Service
// ---------------------------------------------------------------------------

// InstanceService implements the application-level use cases for workflow instances.
type InstanceService struct {
	instRepo   domain.WorkflowInstanceRepository
	defRepo    domain.WorkflowDefinitionRepository
	actionRepo domain.WorkflowActionRepository
	logger     zerolog.Logger
}

// NewInstanceService constructs a new InstanceService.
func NewInstanceService(
	instRepo domain.WorkflowInstanceRepository,
	defRepo domain.WorkflowDefinitionRepository,
	actionRepo domain.WorkflowActionRepository,
	logger zerolog.Logger,
) *InstanceService {
	return &InstanceService{
		instRepo:   instRepo,
		defRepo:    defRepo,
		actionRepo: actionRepo,
		logger:     logger.With().Str("service", "instance").Logger(),
	}
}

// Create creates a new workflow instance from a definition.
func (s *InstanceService) Create(
	ctx context.Context,
	tenantID uuid.UUID,
	userID uuid.UUID,
	req CreateInstanceRequest,
) (*domain.WorkflowInstance, error) {
	if req.DefinitionID == uuid.Nil {
		return nil, domain.ErrMissingDefinition
	}

	// Verify that the referenced definition exists.
	_, err := s.defRepo.GetByID(ctx, req.DefinitionID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("create instance: %w", err)
	}

	data := req.Data
	if data == nil {
		data = json.RawMessage("{}")
	}

	inst := &domain.WorkflowInstance{
		ID:            uuid.New(),
		DefinitionID:  req.DefinitionID,
		TenantID:      tenantID,
		ReferenceID:   req.ReferenceID,
		ReferenceType: req.ReferenceType,
		CurrentStep:   0,
		Status:        domain.StatusPending,
		Data:          data,
		StartedBy:     &userID,
		Notes:         req.Notes,
	}

	if err := s.instRepo.Create(ctx, inst); err != nil {
		s.logger.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Str("definition_id", req.DefinitionID.String()).
			Msg("failed to create workflow instance")
		return nil, fmt.Errorf("create instance: %w", err)
	}

	s.logger.Info().
		Str("instance_id", inst.ID.String()).
		Str("tenant_id", tenantID.String()).
		Msg("workflow instance created")

	return inst, nil
}

// GetByID returns a single workflow instance by its ID.
func (s *InstanceService) GetByID(
	ctx context.Context,
	id uuid.UUID,
	tenantID uuid.UUID,
) (*domain.WorkflowInstance, error) {
	inst, err := s.instRepo.GetByID(ctx, id, tenantID)
	if err != nil {
		s.logger.Error().Err(err).
			Str("instance_id", id.String()).
			Str("tenant_id", tenantID.String()).
			Msg("failed to get workflow instance")
		return nil, fmt.Errorf("get instance: %w", err)
	}
	return inst, nil
}

// List returns a filtered, paginated list of workflow instances for a tenant.
func (s *InstanceService) List(
	ctx context.Context,
	tenantID uuid.UUID,
	filter domain.InstanceFilter,
) ([]*domain.WorkflowInstance, int64, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PerPage < 1 {
		filter.PerPage = 20
	}

	items, total, err := s.instRepo.List(ctx, tenantID, filter)
	if err != nil {
		s.logger.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Msg("failed to list workflow instances")
		return nil, 0, fmt.Errorf("list instances: %w", err)
	}
	return items, total, nil
}

// PerformAction performs an action (approve/reject/escalate) on a workflow instance.
func (s *InstanceService) PerformAction(
	ctx context.Context,
	instanceID uuid.UUID,
	tenantID uuid.UUID,
	userID uuid.UUID,
	req PerformActionRequest,
) (*domain.WorkflowInstance, error) {
	if !domain.ValidAction(req.Action) {
		return nil, domain.ErrInvalidAction
	}

	inst, err := s.instRepo.GetByID(ctx, instanceID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("perform action: %w", err)
	}

	// Check that the instance is not already in a terminal state.
	if inst.Status == domain.StatusApproved || inst.Status == domain.StatusRejected || inst.Status == domain.StatusCancelled {
		return nil, domain.ErrInstanceCompleted
	}

	// Record the action.
	wfAction := &domain.WorkflowAction{
		ID:          uuid.New(),
		InstanceID:  instanceID,
		Step:        inst.CurrentStep,
		Action:      req.Action,
		PerformedBy: &userID,
		Comment:     req.Comment,
	}

	if err := s.actionRepo.Create(ctx, wfAction); err != nil {
		s.logger.Error().Err(err).
			Str("instance_id", instanceID.String()).
			Msg("failed to record workflow action")
		return nil, fmt.Errorf("perform action: %w", err)
	}

	// Determine the new state based on the action.
	now := time.Now()
	switch req.Action {
	case domain.ActionApprove:
		// Advance to next step. If no more steps, mark as approved.
		inst.CurrentStep++
		steps := countSteps(inst)
		if inst.CurrentStep >= steps {
			inst.Status = domain.StatusApproved
			inst.CompletedAt = &now
		} else {
			inst.Status = domain.StatusActive
		}
	case domain.ActionReject:
		inst.Status = domain.StatusRejected
		inst.CompletedAt = &now
	case domain.ActionEscalate:
		inst.Status = domain.StatusEscalated
	}

	if err := s.instRepo.Update(ctx, inst); err != nil {
		s.logger.Error().Err(err).
			Str("instance_id", instanceID.String()).
			Msg("failed to update workflow instance after action")
		return nil, fmt.Errorf("perform action update: %w", err)
	}

	s.logger.Info().
		Str("instance_id", instanceID.String()).
		Str("action", req.Action).
		Str("new_status", inst.Status).
		Str("tenant_id", tenantID.String()).
		Msg("workflow action performed")

	return inst, nil
}

// countSteps parses the definition steps JSON to determine the total number of steps.
// Falls back to 1 if the steps array cannot be parsed or is empty.
func countSteps(inst *domain.WorkflowInstance) int {
	var steps []json.RawMessage
	if err := json.Unmarshal(inst.Data, &steps); err == nil && len(steps) > 0 {
		return len(steps)
	}
	// Default: single-step workflow — approve completes it.
	return 1
}

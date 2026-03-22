package application

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/workflow-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/workflow-service/internal/ports"
)

type WorkflowService struct {
	definitionRepo ports.WorkflowDefinitionRepository
	instanceRepo   ports.WorkflowInstanceRepository
	stepRepo       ports.WorkflowStepRepository
	log            logger.Logger
}

func NewWorkflowService(
	dr ports.WorkflowDefinitionRepository,
	ir ports.WorkflowInstanceRepository,
	sr ports.WorkflowStepRepository,
	log logger.Logger,
) *WorkflowService {
	return &WorkflowService{
		definitionRepo: dr,
		instanceRepo:   ir,
		stepRepo:       sr,
		log:            log,
	}
}

func (s *WorkflowService) CreateDefinition(ctx context.Context, tenantID uuid.UUID, req *CreateWorkflowDefinitionRequest) (*WorkflowDefinitionResponse, error) {
	wd := &domain.WorkflowDefinition{
		ID:               uuid.New(),
		TenantID:         tenantID,
		Name:             req.Name,
		Description:      req.Description,
		TriggerType:      domain.TriggerType(req.TriggerType),
		TriggerConfig:    req.TriggerConfig,
		IsActive:         true,
		IsTemplate:       req.IsTemplate,
		TemplateCategory: req.TemplateCategory,
		Version:          1,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := s.definitionRepo.Create(ctx, wd); err != nil {
		s.log.Error("Failed to create workflow definition", err)
		return nil, err
	}

	return s.workflowDefinitionToResponse(wd), nil
}

func (s *WorkflowService) GetDefinition(ctx context.Context, id uuid.UUID) (*WorkflowDefinitionResponse, error) {
	wd, err := s.definitionRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.workflowDefinitionToResponse(wd), nil
}

func (s *WorkflowService) ListDefinitions(ctx context.Context, tenantID uuid.UUID) ([]*WorkflowDefinitionResponse, error) {
	wds, err := s.definitionRepo.ListByTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	var responses []*WorkflowDefinitionResponse
	for _, wd := range wds {
		responses = append(responses, s.workflowDefinitionToResponse(wd))
	}
	return responses, nil
}

func (s *WorkflowService) ListTemplates(ctx context.Context) ([]*WorkflowDefinitionResponse, error) {
	wds, err := s.definitionRepo.ListTemplates(ctx)
	if err != nil {
		return nil, err
	}

	var responses []*WorkflowDefinitionResponse
	for _, wd := range wds {
		responses = append(responses, s.workflowDefinitionToResponse(wd))
	}
	return responses, nil
}

func (s *WorkflowService) UpdateDefinition(ctx context.Context, id uuid.UUID, req *CreateWorkflowDefinitionRequest) (*WorkflowDefinitionResponse, error) {
	wd, err := s.definitionRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	wd.Name = req.Name
	wd.Description = req.Description
	wd.TriggerType = domain.TriggerType(req.TriggerType)
	wd.TriggerConfig = req.TriggerConfig
	wd.Version++
	wd.UpdatedAt = time.Now()

	if err := s.definitionRepo.Update(ctx, wd); err != nil {
		s.log.Error("Failed to update workflow definition", err)
		return nil, err
	}

	return s.workflowDefinitionToResponse(wd), nil
}

func (s *WorkflowService) DeleteDefinition(ctx context.Context, id uuid.UUID) error {
	return s.definitionRepo.Delete(ctx, id)
}

func (s *WorkflowService) InstantiateWorkflow(ctx context.Context, tenantID uuid.UUID, definitionID uuid.UUID, req *InstantiateWorkflowRequest) (*WorkflowInstanceResponse, error) {
	_, err := s.definitionRepo.GetByID(ctx, definitionID)
	if err != nil {
		return nil, err
	}

	wi := &domain.WorkflowInstance{
		ID:            uuid.New(),
		TenantID:      tenantID,
		DefinitionID:  definitionID,
		Status:        domain.StatusRunning,
		TriggerData:   req.TriggerData,
		ContextData:   json.RawMessage("{}"),
		CurrentStepIdx: 0,
		StartedAt:     time.Now(),
	}

	if err := s.instanceRepo.Create(ctx, wi); err != nil {
		s.log.Error("Failed to create workflow instance", err)
		return nil, err
	}

	s.log.Info("Instantiated workflow", "definition_id", definitionID, "instance_id", wi.ID)
	return s.workflowInstanceToResponse(wi), nil
}

func (s *WorkflowService) GetInstance(ctx context.Context, id uuid.UUID) (*WorkflowInstanceResponse, error) {
	wi, err := s.instanceRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.workflowInstanceToResponse(wi), nil
}

func (s *WorkflowService) ListInstances(ctx context.Context, tenantID uuid.UUID) ([]*WorkflowInstanceResponse, error) {
	wis, err := s.instanceRepo.ListByTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	var responses []*WorkflowInstanceResponse
	for _, wi := range wis {
		responses = append(responses, s.workflowInstanceToResponse(wi))
	}
	return responses, nil
}

func (s *WorkflowService) CancelInstance(ctx context.Context, id uuid.UUID) error {
	wi, err := s.instanceRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	wi.Status = domain.StatusCancelled
	return s.instanceRepo.Update(ctx, wi)
}

func (s *WorkflowService) GetDashboard(ctx context.Context, tenantID uuid.UUID) (*DashboardResponse, error) {
	running, _ := s.instanceRepo.CountByStatus(ctx, tenantID, domain.StatusRunning)
	failed, _ := s.instanceRepo.CountByStatus(ctx, tenantID, domain.StatusFailed)
	completed, _ := s.instanceRepo.CountCompletedSince(ctx, tenantID, time.Now().AddDate(0, 0, -1))

	return &DashboardResponse{
		ActiveCount:    running,
		CompletedToday: completed,
		FailedCount:    failed,
	}, nil
}

func (s *WorkflowService) workflowDefinitionToResponse(wd *domain.WorkflowDefinition) *WorkflowDefinitionResponse {
	return &WorkflowDefinitionResponse{
		ID:               wd.ID,
		TenantID:         wd.TenantID,
		Name:             wd.Name,
		Description:      wd.Description,
		TriggerType:      string(wd.TriggerType),
		TriggerConfig:    wd.TriggerConfig,
		IsActive:         wd.IsActive,
		IsTemplate:       wd.IsTemplate,
		TemplateCategory: wd.TemplateCategory,
		Version:          wd.Version,
		CreatedAt:        wd.CreatedAt,
		UpdatedAt:        wd.UpdatedAt,
	}
}

func (s *WorkflowService) workflowInstanceToResponse(wi *domain.WorkflowInstance) *WorkflowInstanceResponse {
	return &WorkflowInstanceResponse{
		ID:           wi.ID,
		TenantID:     wi.TenantID,
		DefinitionID: wi.DefinitionID,
		Status:       string(wi.Status),
		TriggerData:  wi.TriggerData,
		ContextData:  wi.ContextData,
		StepIndex:    wi.CurrentStepIdx,
		StartedAt:    wi.StartedAt,
		CompletedAt:  wi.CompletedAt,
		ErrorMessage: wi.ErrorMessage,
	}
}

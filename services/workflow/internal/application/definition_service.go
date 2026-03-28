package application

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/services/workflow/internal/domain"
)

// ---------------------------------------------------------------------------
// Request / Response DTOs
// ---------------------------------------------------------------------------

// CreateDefinitionRequest holds the data for creating a workflow definition.
type CreateDefinitionRequest struct {
	Name  string          `json:"name"`
	Type  string          `json:"type"`
	Steps json.RawMessage `json:"steps"`
}

// UpdateDefinitionRequest holds the data for updating a workflow definition.
type UpdateDefinitionRequest struct {
	Name     *string          `json:"name,omitempty"`
	Type     *string          `json:"type,omitempty"`
	Steps    *json.RawMessage `json:"steps,omitempty"`
	IsActive *bool            `json:"is_active,omitempty"`
}

// ---------------------------------------------------------------------------
// Service
// ---------------------------------------------------------------------------

// DefinitionService implements the application-level use cases for workflow definitions.
type DefinitionService struct {
	defRepo domain.WorkflowDefinitionRepository
	logger  zerolog.Logger
}

// NewDefinitionService constructs a new DefinitionService.
func NewDefinitionService(
	defRepo domain.WorkflowDefinitionRepository,
	logger zerolog.Logger,
) *DefinitionService {
	return &DefinitionService{
		defRepo: defRepo,
		logger:  logger.With().Str("service", "definition").Logger(),
	}
}

// Create creates a new workflow definition.
func (s *DefinitionService) Create(
	ctx context.Context,
	tenantID uuid.UUID,
	userID uuid.UUID,
	req CreateDefinitionRequest,
) (*domain.WorkflowDefinition, error) {
	if req.Name == "" {
		return nil, domain.ErrMissingName
	}

	wfType := req.Type
	if wfType == "" {
		wfType = domain.TypeApproval
	}
	if !domain.ValidType(wfType) {
		return nil, domain.ErrInvalidType
	}

	steps := req.Steps
	if steps == nil {
		steps = json.RawMessage("[]")
	}

	def := &domain.WorkflowDefinition{
		ID:        uuid.New(),
		TenantID:  tenantID,
		Name:      req.Name,
		Type:      wfType,
		Steps:     steps,
		IsActive:  true,
		CreatedBy: &userID,
	}

	if err := s.defRepo.Create(ctx, def); err != nil {
		s.logger.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Msg("failed to create workflow definition")
		return nil, fmt.Errorf("create definition: %w", err)
	}

	s.logger.Info().
		Str("definition_id", def.ID.String()).
		Str("tenant_id", tenantID.String()).
		Msg("workflow definition created")

	return def, nil
}

// GetByID returns a single workflow definition by its ID.
func (s *DefinitionService) GetByID(
	ctx context.Context,
	id uuid.UUID,
	tenantID uuid.UUID,
) (*domain.WorkflowDefinition, error) {
	def, err := s.defRepo.GetByID(ctx, id, tenantID)
	if err != nil {
		s.logger.Error().Err(err).
			Str("definition_id", id.String()).
			Str("tenant_id", tenantID.String()).
			Msg("failed to get workflow definition")
		return nil, fmt.Errorf("get definition: %w", err)
	}
	return def, nil
}

// List returns a paginated list of workflow definitions for a tenant.
func (s *DefinitionService) List(
	ctx context.Context,
	tenantID uuid.UUID,
	filter domain.DefinitionFilter,
) ([]*domain.WorkflowDefinition, int64, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PerPage < 1 {
		filter.PerPage = 20
	}

	items, total, err := s.defRepo.List(ctx, tenantID, filter)
	if err != nil {
		s.logger.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Msg("failed to list workflow definitions")
		return nil, 0, fmt.Errorf("list definitions: %w", err)
	}
	return items, total, nil
}

// Update updates an existing workflow definition.
func (s *DefinitionService) Update(
	ctx context.Context,
	id uuid.UUID,
	tenantID uuid.UUID,
	req UpdateDefinitionRequest,
) (*domain.WorkflowDefinition, error) {
	def, err := s.defRepo.GetByID(ctx, id, tenantID)
	if err != nil {
		return nil, fmt.Errorf("update definition: %w", err)
	}

	if req.Name != nil {
		def.Name = *req.Name
	}
	if req.Type != nil {
		if !domain.ValidType(*req.Type) {
			return nil, domain.ErrInvalidType
		}
		def.Type = *req.Type
	}
	if req.Steps != nil {
		def.Steps = *req.Steps
	}
	if req.IsActive != nil {
		def.IsActive = *req.IsActive
	}

	if err := s.defRepo.Update(ctx, def); err != nil {
		s.logger.Error().Err(err).
			Str("definition_id", id.String()).
			Str("tenant_id", tenantID.String()).
			Msg("failed to update workflow definition")
		return nil, fmt.Errorf("update definition: %w", err)
	}

	s.logger.Info().
		Str("definition_id", id.String()).
		Str("tenant_id", tenantID.String()).
		Msg("workflow definition updated")

	return def, nil
}

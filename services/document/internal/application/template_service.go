package application

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/services/document/internal/domain"
)

// ---------------------------------------------------------------------------
// Request DTOs
// ---------------------------------------------------------------------------

// CreateTemplateRequest holds the data needed to create a document template.
type CreateTemplateRequest struct {
	Name      string          `json:"name"`
	Type      string          `json:"type"`
	Content   string          `json:"content"`
	Variables json.RawMessage `json:"variables,omitempty"`
	IsDefault bool            `json:"is_default"`
}

// UpdateTemplateRequest holds the data for updating a document template.
type UpdateTemplateRequest struct {
	Name      *string          `json:"name,omitempty"`
	Type      *string          `json:"type,omitempty"`
	Content   *string          `json:"content,omitempty"`
	Variables *json.RawMessage `json:"variables,omitempty"`
	IsDefault *bool            `json:"is_default,omitempty"`
	IsActive  *bool            `json:"is_active,omitempty"`
}

// ---------------------------------------------------------------------------
// Service
// ---------------------------------------------------------------------------

// TemplateService implements the application-level use cases for document templates.
type TemplateService struct {
	templateRepo domain.DocumentTemplateRepository
	logger       zerolog.Logger
}

// NewTemplateService constructs a new TemplateService.
func NewTemplateService(
	templateRepo domain.DocumentTemplateRepository,
	logger zerolog.Logger,
) *TemplateService {
	return &TemplateService{
		templateRepo: templateRepo,
		logger:       logger.With().Str("service", "template").Logger(),
	}
}

// Create creates a new document template.
func (s *TemplateService) Create(
	ctx context.Context,
	tenantID uuid.UUID,
	req CreateTemplateRequest,
) (*domain.DocumentTemplate, error) {
	if req.Name == "" {
		return nil, domain.ErrNameRequired
	}
	if req.Type == "" {
		return nil, domain.ErrTypeRequired
	}

	variables := req.Variables
	if variables == nil {
		variables = json.RawMessage(`{}`)
	}

	tpl := &domain.DocumentTemplate{
		ID:        uuid.New(),
		TenantID:  tenantID,
		Name:      req.Name,
		Type:      req.Type,
		Content:   req.Content,
		Variables: variables,
		IsDefault: req.IsDefault,
		IsActive:  true,
	}

	if err := s.templateRepo.Create(ctx, tpl); err != nil {
		s.logger.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Msg("failed to create template")
		return nil, fmt.Errorf("create template: %w", err)
	}

	s.logger.Info().
		Str("template_id", tpl.ID.String()).
		Str("tenant_id", tenantID.String()).
		Msg("template created")

	return tpl, nil
}

// GetByID retrieves a document template by ID.
func (s *TemplateService) GetByID(
	ctx context.Context,
	id uuid.UUID,
	tenantID uuid.UUID,
) (*domain.DocumentTemplate, error) {
	tpl, err := s.templateRepo.GetByID(ctx, id, tenantID)
	if err != nil {
		s.logger.Error().Err(err).
			Str("template_id", id.String()).
			Str("tenant_id", tenantID.String()).
			Msg("failed to get template")
		return nil, fmt.Errorf("get template: %w", err)
	}
	return tpl, nil
}

// List returns all document templates for a tenant.
func (s *TemplateService) List(
	ctx context.Context,
	tenantID uuid.UUID,
) ([]*domain.DocumentTemplate, error) {
	templates, err := s.templateRepo.List(ctx, tenantID)
	if err != nil {
		s.logger.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Msg("failed to list templates")
		return nil, fmt.Errorf("list templates: %w", err)
	}
	return templates, nil
}

// Update updates an existing document template.
func (s *TemplateService) Update(
	ctx context.Context,
	id uuid.UUID,
	tenantID uuid.UUID,
	req UpdateTemplateRequest,
) (*domain.DocumentTemplate, error) {
	tpl, err := s.templateRepo.GetByID(ctx, id, tenantID)
	if err != nil {
		return nil, fmt.Errorf("update template: %w", err)
	}

	if req.Name != nil {
		tpl.Name = *req.Name
	}
	if req.Type != nil {
		tpl.Type = *req.Type
	}
	if req.Content != nil {
		tpl.Content = *req.Content
	}
	if req.Variables != nil {
		tpl.Variables = *req.Variables
	}
	if req.IsDefault != nil {
		tpl.IsDefault = *req.IsDefault
	}
	if req.IsActive != nil {
		tpl.IsActive = *req.IsActive
	}

	if err := s.templateRepo.Update(ctx, tpl); err != nil {
		s.logger.Error().Err(err).
			Str("template_id", id.String()).
			Str("tenant_id", tenantID.String()).
			Msg("failed to update template")
		return nil, fmt.Errorf("update template: %w", err)
	}

	s.logger.Info().
		Str("template_id", id.String()).
		Str("tenant_id", tenantID.String()).
		Msg("template updated")

	return tpl, nil
}

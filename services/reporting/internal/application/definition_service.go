package application

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/services/reporting/internal/domain"
)

// ---------------------------------------------------------------------------
// Request DTOs
// ---------------------------------------------------------------------------

type CreateDefinitionRequest struct {
	Name        string          `json:"name"`
	Type        string          `json:"type"`
	QueryConfig json.RawMessage `json:"query_config"`
	Schedule    string          `json:"schedule"`
	Format      string          `json:"format"`
	IsActive    *bool           `json:"is_active"`
}

type UpdateDefinitionRequest struct {
	Name        *string          `json:"name"`
	Type        *string          `json:"type"`
	QueryConfig *json.RawMessage `json:"query_config"`
	Schedule    *string          `json:"schedule"`
	Format      *string          `json:"format"`
	IsActive    *bool            `json:"is_active"`
}

// ---------------------------------------------------------------------------
// Service
// ---------------------------------------------------------------------------

type DefinitionService struct {
	repo   domain.ReportDefinitionRepository
	logger zerolog.Logger
}

func NewDefinitionService(repo domain.ReportDefinitionRepository, logger zerolog.Logger) *DefinitionService {
	return &DefinitionService{
		repo:   repo,
		logger: logger.With().Str("service", "definition").Logger(),
	}
}

func (s *DefinitionService) Create(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID, req CreateDefinitionRequest) (*domain.ReportDefinition, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}

	defType := "custom"
	if req.Type != "" {
		defType = req.Type
	}
	format := "pdf"
	if req.Format != "" {
		format = req.Format
	}
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	queryConfig := json.RawMessage("{}")
	if req.QueryConfig != nil {
		queryConfig = req.QueryConfig
	}

	def := &domain.ReportDefinition{
		ID:          uuid.New(),
		TenantID:    tenantID,
		Name:        req.Name,
		Type:        defType,
		QueryConfig: queryConfig,
		Schedule:    req.Schedule,
		Format:      format,
		IsActive:    isActive,
		CreatedBy:   &userID,
	}

	if err := s.repo.Create(ctx, def); err != nil {
		return nil, fmt.Errorf("create definition: %w", err)
	}

	s.logger.Info().Str("definition_id", def.ID.String()).Msg("report definition created")
	return def, nil
}

func (s *DefinitionService) GetByID(ctx context.Context, id, tenantID uuid.UUID) (*domain.ReportDefinition, error) {
	return s.repo.GetByID(ctx, id, tenantID)
}

func (s *DefinitionService) List(ctx context.Context, tenantID uuid.UUID, filter domain.DefinitionFilter) ([]*domain.ReportDefinition, int64, error) {
	return s.repo.List(ctx, tenantID, filter)
}

func (s *DefinitionService) Update(ctx context.Context, id, tenantID uuid.UUID, req UpdateDefinitionRequest) (*domain.ReportDefinition, error) {
	existing, err := s.repo.GetByID(ctx, id, tenantID)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		existing.Name = *req.Name
	}
	if req.Type != nil {
		existing.Type = *req.Type
	}
	if req.QueryConfig != nil {
		existing.QueryConfig = *req.QueryConfig
	}
	if req.Schedule != nil {
		existing.Schedule = *req.Schedule
	}
	if req.Format != nil {
		existing.Format = *req.Format
	}
	if req.IsActive != nil {
		existing.IsActive = *req.IsActive
	}

	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, err
	}

	return existing, nil
}

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

type CreateWidgetRequest struct {
	Name     string          `json:"name"`
	Type     string          `json:"type"`
	Config   json.RawMessage `json:"config"`
	Position *int            `json:"position"`
	IsActive *bool           `json:"is_active"`
}

type UpdateWidgetRequest struct {
	Name     *string          `json:"name"`
	Type     *string          `json:"type"`
	Config   *json.RawMessage `json:"config"`
	Position *int             `json:"position"`
	IsActive *bool            `json:"is_active"`
}

// ---------------------------------------------------------------------------
// Service
// ---------------------------------------------------------------------------

type WidgetService struct {
	repo   domain.DashboardWidgetRepository
	logger zerolog.Logger
}

func NewWidgetService(repo domain.DashboardWidgetRepository, logger zerolog.Logger) *WidgetService {
	return &WidgetService{
		repo:   repo,
		logger: logger.With().Str("service", "widget").Logger(),
	}
}

func (s *WidgetService) Create(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID, req CreateWidgetRequest) (*domain.DashboardWidget, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}

	widgetType := "chart"
	if req.Type != "" {
		widgetType = req.Type
	}
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	position := 0
	if req.Position != nil {
		position = *req.Position
	}
	config := json.RawMessage("{}")
	if req.Config != nil {
		config = req.Config
	}

	widget := &domain.DashboardWidget{
		ID:        uuid.New(),
		TenantID:  tenantID,
		Name:      req.Name,
		Type:      widgetType,
		Config:    config,
		Position:  position,
		IsActive:  isActive,
		CreatedBy: &userID,
	}

	if err := s.repo.Create(ctx, widget); err != nil {
		return nil, fmt.Errorf("create widget: %w", err)
	}

	s.logger.Info().Str("widget_id", widget.ID.String()).Msg("dashboard widget created")
	return widget, nil
}

func (s *WidgetService) GetByID(ctx context.Context, id, tenantID uuid.UUID) (*domain.DashboardWidget, error) {
	return s.repo.GetByID(ctx, id, tenantID)
}

func (s *WidgetService) List(ctx context.Context, tenantID uuid.UUID, filter domain.WidgetFilter) ([]*domain.DashboardWidget, int64, error) {
	return s.repo.List(ctx, tenantID, filter)
}

func (s *WidgetService) Update(ctx context.Context, id, tenantID uuid.UUID, req UpdateWidgetRequest) (*domain.DashboardWidget, error) {
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
	if req.Config != nil {
		existing.Config = *req.Config
	}
	if req.Position != nil {
		existing.Position = *req.Position
	}
	if req.IsActive != nil {
		existing.IsActive = *req.IsActive
	}

	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, err
	}

	return existing, nil
}

func (s *WidgetService) Delete(ctx context.Context, id, tenantID uuid.UUID) error {
	return s.repo.Delete(ctx, id, tenantID)
}

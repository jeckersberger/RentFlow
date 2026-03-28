package application

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/services/ai/internal/domain"
)

// ---------------------------------------------------------------------------
// Request DTOs
// ---------------------------------------------------------------------------

type CreateSuggestionRequest struct {
	Type        string          `json:"type"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Data        json.RawMessage `json:"data"`
}

// ---------------------------------------------------------------------------
// Service
// ---------------------------------------------------------------------------

type SuggestionService struct {
	repo   domain.AISuggestionRepository
	logger zerolog.Logger
}

func NewSuggestionService(repo domain.AISuggestionRepository, logger zerolog.Logger) *SuggestionService {
	return &SuggestionService{
		repo:   repo,
		logger: logger.With().Str("service", "suggestion").Logger(),
	}
}

func (s *SuggestionService) Create(ctx context.Context, tenantID uuid.UUID, req CreateSuggestionRequest) (*domain.AISuggestion, error) {
	if req.Type == "" {
		return nil, fmt.Errorf("type is required")
	}
	if req.Title == "" {
		return nil, fmt.Errorf("title is required")
	}

	data := json.RawMessage("{}")
	if req.Data != nil {
		data = req.Data
	}

	sug := &domain.AISuggestion{
		ID:          uuid.New(),
		TenantID:    tenantID,
		Type:        req.Type,
		Title:       req.Title,
		Description: req.Description,
		Data:        data,
		Status:      "pending",
	}

	if err := s.repo.Create(ctx, sug); err != nil {
		return nil, fmt.Errorf("create suggestion: %w", err)
	}

	s.logger.Info().Str("suggestion_id", sug.ID.String()).Msg("ai suggestion created")
	return sug, nil
}

func (s *SuggestionService) GetByID(ctx context.Context, id, tenantID uuid.UUID) (*domain.AISuggestion, error) {
	return s.repo.GetByID(ctx, id, tenantID)
}

func (s *SuggestionService) List(ctx context.Context, tenantID uuid.UUID, filter domain.SuggestionFilter) ([]*domain.AISuggestion, int64, error) {
	return s.repo.List(ctx, tenantID, filter)
}

func (s *SuggestionService) Accept(ctx context.Context, id, tenantID, userID uuid.UUID) (*domain.AISuggestion, error) {
	existing, err := s.repo.GetByID(ctx, id, tenantID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	existing.Status = "accepted"
	existing.AcceptedBy = &userID
	existing.AcceptedAt = &now

	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, err
	}

	s.logger.Info().Str("suggestion_id", id.String()).Msg("ai suggestion accepted")
	return existing, nil
}

func (s *SuggestionService) Dismiss(ctx context.Context, id, tenantID uuid.UUID) (*domain.AISuggestion, error) {
	existing, err := s.repo.GetByID(ctx, id, tenantID)
	if err != nil {
		return nil, err
	}

	existing.Status = "dismissed"

	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, err
	}

	s.logger.Info().Str("suggestion_id", id.String()).Msg("ai suggestion dismissed")
	return existing, nil
}

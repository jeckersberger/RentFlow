package application

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/services/ai/internal/domain"
)

// ---------------------------------------------------------------------------
// Request DTOs
// ---------------------------------------------------------------------------

type CreatePredictionRequest struct {
	Type          string          `json:"type"`
	ReferenceID   *uuid.UUID      `json:"reference_id"`
	ReferenceType string          `json:"reference_type"`
	Prediction    json.RawMessage `json:"prediction"`
	Confidence    *float64        `json:"confidence"`
	ModelVersion  string          `json:"model_version"`
}

// ---------------------------------------------------------------------------
// Service
// ---------------------------------------------------------------------------

type PredictionService struct {
	repo   domain.AIPredictionRepository
	logger zerolog.Logger
}

func NewPredictionService(repo domain.AIPredictionRepository, logger zerolog.Logger) *PredictionService {
	return &PredictionService{
		repo:   repo,
		logger: logger.With().Str("service", "prediction").Logger(),
	}
}

func (s *PredictionService) Create(ctx context.Context, tenantID uuid.UUID, req CreatePredictionRequest) (*domain.AIPrediction, error) {
	if req.Type == "" {
		return nil, fmt.Errorf("type is required")
	}

	prediction := json.RawMessage("{}")
	if req.Prediction != nil {
		prediction = req.Prediction
	}
	confidence := 0.0
	if req.Confidence != nil {
		confidence = *req.Confidence
	}

	pred := &domain.AIPrediction{
		ID:            uuid.New(),
		TenantID:      tenantID,
		Type:          req.Type,
		ReferenceID:   req.ReferenceID,
		ReferenceType: req.ReferenceType,
		Prediction:    prediction,
		Confidence:    confidence,
		ModelVersion:  req.ModelVersion,
	}

	if err := s.repo.Create(ctx, pred); err != nil {
		return nil, fmt.Errorf("create prediction: %w", err)
	}

	s.logger.Info().Str("prediction_id", pred.ID.String()).Msg("ai prediction created")
	return pred, nil
}

func (s *PredictionService) GetByID(ctx context.Context, id, tenantID uuid.UUID) (*domain.AIPrediction, error) {
	return s.repo.GetByID(ctx, id, tenantID)
}

func (s *PredictionService) List(ctx context.Context, tenantID uuid.UUID, filter domain.PredictionFilter) ([]*domain.AIPrediction, int64, error) {
	return s.repo.List(ctx, tenantID, filter)
}

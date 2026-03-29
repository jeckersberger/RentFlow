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

type CreateTrainingDataRequest struct {
	Type       string          `json:"type"`
	InputData  json.RawMessage `json:"input_data"`
	OutputData json.RawMessage `json:"output_data"`
	Feedback   string          `json:"feedback"`
}

// ---------------------------------------------------------------------------
// Service
// ---------------------------------------------------------------------------

type TrainingDataService struct {
	repo   domain.AITrainingDataRepository
	logger zerolog.Logger
}

func NewTrainingDataService(repo domain.AITrainingDataRepository, logger zerolog.Logger) *TrainingDataService {
	return &TrainingDataService{
		repo:   repo,
		logger: logger.With().Str("service", "training_data").Logger(),
	}
}

func (s *TrainingDataService) List(ctx context.Context, tenantID uuid.UUID, filter domain.TrainingDataFilter) ([]*domain.AITrainingData, int64, error) {
	items, total, err := s.repo.List(ctx, tenantID, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("list training data: %w", err)
	}
	return items, total, nil
}

func (s *TrainingDataService) Create(ctx context.Context, tenantID uuid.UUID, req CreateTrainingDataRequest) (*domain.AITrainingData, error) {
	if req.Type == "" {
		return nil, fmt.Errorf("type is required")
	}

	inputData := json.RawMessage("{}")
	if req.InputData != nil {
		inputData = req.InputData
	}
	outputData := json.RawMessage("{}")
	if req.OutputData != nil {
		outputData = req.OutputData
	}

	td := &domain.AITrainingData{
		ID:         uuid.New(),
		TenantID:   tenantID,
		Type:       req.Type,
		InputData:  inputData,
		OutputData: outputData,
		Feedback:   req.Feedback,
	}

	if err := s.repo.Create(ctx, td); err != nil {
		return nil, fmt.Errorf("create training data: %w", err)
	}

	s.logger.Info().Str("training_data_id", td.ID.String()).Msg("ai training data submitted")
	return td, nil
}

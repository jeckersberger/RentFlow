package domain

import (
	"context"

	"github.com/google/uuid"
)

type AIPredictionRepository interface {
	Create(ctx context.Context, pred *AIPrediction) error
	GetByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*AIPrediction, error)
	List(ctx context.Context, tenantID uuid.UUID, filter PredictionFilter) ([]*AIPrediction, int64, error)
}

type AISuggestionRepository interface {
	Create(ctx context.Context, sug *AISuggestion) error
	GetByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*AISuggestion, error)
	List(ctx context.Context, tenantID uuid.UUID, filter SuggestionFilter) ([]*AISuggestion, int64, error)
	Update(ctx context.Context, sug *AISuggestion) error
}

type AITrainingDataRepository interface {
	Create(ctx context.Context, td *AITrainingData) error
}

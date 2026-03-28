package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jeckersberger/EquipFlow/services/ai/internal/domain"
)

type TrainingDataRepo struct {
	pool *pgxpool.Pool
}

func NewTrainingDataRepo(pool *pgxpool.Pool) *TrainingDataRepo {
	return &TrainingDataRepo{pool: pool}
}

func (r *TrainingDataRepo) Create(ctx context.Context, td *domain.AITrainingData) error {
	query := `
		INSERT INTO ai_training_data (
			id, tenant_id, type, input_data, output_data, feedback
		) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at`

	err := r.pool.QueryRow(ctx, query,
		td.ID, td.TenantID, td.Type, td.InputData,
		td.OutputData, nilIfEmpty(td.Feedback),
	).Scan(&td.CreatedAt)
	if err != nil {
		return fmt.Errorf("training_data_repo: create: %w", err)
	}
	return nil
}

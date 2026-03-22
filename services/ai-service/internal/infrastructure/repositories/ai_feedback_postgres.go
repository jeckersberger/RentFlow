package repositories

import (
	"context"
	"database/sql"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/ai-service/internal/domain"
)

// PostgresAIFeedbackRepository implements AIFeedbackRepository
type PostgresAIFeedbackRepository struct {
	db     *sql.DB
	logger logger.Logger
}

// NewPostgresAIFeedbackRepository creates a new feedback repository
func NewPostgresAIFeedbackRepository(db *sql.DB, log logger.Logger) *PostgresAIFeedbackRepository {
	return &PostgresAIFeedbackRepository{
		db:     db,
		logger: log,
	}
}

// FindByID retrieves feedback by ID
func (r *PostgresAIFeedbackRepository) FindByID(ctx context.Context, id string) (*domain.AIFeedback, error) {
	query := `
		SELECT id, tenant_id, request_id, rating, comment, is_correct, created_at
		FROM ai_feedback
		WHERE id = $1
	`

	var feedback domain.AIFeedback
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&feedback.ID, &feedback.TenantID, &feedback.RequestID, &feedback.Rating,
		&feedback.Comment, &feedback.IsCorrect, &feedback.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		r.logger.Error("failed to query feedback", err)
		return nil, err
	}

	return &feedback, nil
}

// ListByRequest retrieves feedback for a request
func (r *PostgresAIFeedbackRepository) ListByRequest(ctx context.Context, requestID string) ([]*domain.AIFeedback, error) {
	query := `
		SELECT id, tenant_id, request_id, rating, comment, is_correct, created_at
		FROM ai_feedback
		WHERE request_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, requestID)
	if err != nil {
		r.logger.Error("failed to query feedback by request", err)
		return nil, err
	}
	defer rows.Close()

	var feedbacks []*domain.AIFeedback
	for rows.Next() {
		var f domain.AIFeedback
		if err := rows.Scan(
			&f.ID, &f.TenantID, &f.RequestID, &f.Rating,
			&f.Comment, &f.IsCorrect, &f.CreatedAt,
		); err != nil {
			r.logger.Error("failed to scan feedback", err)
			return nil, err
		}
		feedbacks = append(feedbacks, &f)
	}

	return feedbacks, rows.Err()
}

// Save persists feedback
func (r *PostgresAIFeedbackRepository) Save(ctx context.Context, feedback *domain.AIFeedback) error {
	query := `
		INSERT INTO ai_feedback (id, tenant_id, request_id, rating, comment, is_correct, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE SET
		rating = $4, comment = $5, is_correct = $6
	`

	_, err := r.db.ExecContext(ctx, query,
		feedback.ID, feedback.TenantID, feedback.RequestID, feedback.Rating,
		feedback.Comment, feedback.IsCorrect, feedback.CreatedAt,
	)

	if err != nil {
		r.logger.Error("failed to save feedback", err)
		return err
	}

	return nil
}

// Delete deletes feedback
func (r *PostgresAIFeedbackRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM ai_feedback WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Error("failed to delete feedback", err)
		return err
	}
	return nil
}

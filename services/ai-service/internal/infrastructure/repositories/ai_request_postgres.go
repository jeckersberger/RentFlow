package repositories

import (
	"context"
	"database/sql"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/ai-service/internal/domain"
)

// PostgresAIRequestRepository implements AIRequestRepository
type PostgresAIRequestRepository struct {
	db     *sql.DB
	logger logger.Logger
}

// NewPostgresAIRequestRepository creates a new AI request repository
func NewPostgresAIRequestRepository(db *sql.DB, log logger.Logger) *PostgresAIRequestRepository {
	return &PostgresAIRequestRepository{
		db:     db,
		logger: log,
	}
}

// FindByID retrieves an AI request by ID
func (r *PostgresAIRequestRepository) FindByID(ctx context.Context, id string) (*domain.AIRequest, error) {
	query := `
		SELECT id, tenant_id, provider_id, request_type, input_text, anonymized_input,
		       output_text, model_used, tokens_used, latency_ms, status, error_message,
		       created_at, completed_at
		FROM ai_requests
		WHERE id = $1
	`

	var req domain.AIRequest
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&req.ID, &req.TenantID, &req.ProviderID, &req.RequestType, &req.InputText,
		&req.AnonymizedInput, &req.OutputText, &req.ModelUsed, &req.TokensUsed,
		&req.LatencyMs, &req.Status, &req.ErrorMessage, &req.CreatedAt, &req.CompletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		r.logger.Error("failed to query AI request", err)
		return nil, err
	}

	return &req, nil
}

// List retrieves AI requests with pagination
func (r *PostgresAIRequestRepository) List(ctx context.Context, tenantID string, page, perPage int) ([]*domain.AIRequest, int, error) {
	offset := (page - 1) * perPage

	// Get total count
	countQuery := `SELECT COUNT(*) FROM ai_requests WHERE tenant_id = $1`
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, tenantID).Scan(&total); err != nil {
		r.logger.Error("failed to count AI requests", err)
		return nil, 0, err
	}

	// Get paginated results
	query := `
		SELECT id, tenant_id, provider_id, request_type, input_text, anonymized_input,
		       output_text, model_used, tokens_used, latency_ms, status, error_message,
		       created_at, completed_at
		FROM ai_requests
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID, perPage, offset)
	if err != nil {
		r.logger.Error("failed to query AI requests", err)
		return nil, 0, err
	}
	defer rows.Close()

	var requests []*domain.AIRequest
	for rows.Next() {
		var req domain.AIRequest
		if err := rows.Scan(
			&req.ID, &req.TenantID, &req.ProviderID, &req.RequestType, &req.InputText,
			&req.AnonymizedInput, &req.OutputText, &req.ModelUsed, &req.TokensUsed,
			&req.LatencyMs, &req.Status, &req.ErrorMessage, &req.CreatedAt, &req.CompletedAt,
		); err != nil {
			r.logger.Error("failed to scan AI request", err)
			return nil, 0, err
		}
		requests = append(requests, &req)
	}

	if err = rows.Err(); err != nil {
		r.logger.Error("error iterating AI requests", err)
		return nil, 0, err
	}

	return requests, total, nil
}

// ListByProvider retrieves requests for a specific provider
func (r *PostgresAIRequestRepository) ListByProvider(ctx context.Context, providerID string, page, perPage int) ([]*domain.AIRequest, int, error) {
	offset := (page - 1) * perPage

	countQuery := `SELECT COUNT(*) FROM ai_requests WHERE provider_id = $1`
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, providerID).Scan(&total); err != nil {
		r.logger.Error("failed to count requests by provider", err)
		return nil, 0, err
	}

	query := `
		SELECT id, tenant_id, provider_id, request_type, input_text, anonymized_input,
		       output_text, model_used, tokens_used, latency_ms, status, error_message,
		       created_at, completed_at
		FROM ai_requests
		WHERE provider_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, providerID, perPage, offset)
	if err != nil {
		r.logger.Error("failed to query requests by provider", err)
		return nil, 0, err
	}
	defer rows.Close()

	var requests []*domain.AIRequest
	for rows.Next() {
		var req domain.AIRequest
		if err := rows.Scan(
			&req.ID, &req.TenantID, &req.ProviderID, &req.RequestType, &req.InputText,
			&req.AnonymizedInput, &req.OutputText, &req.ModelUsed, &req.TokensUsed,
			&req.LatencyMs, &req.Status, &req.ErrorMessage, &req.CreatedAt, &req.CompletedAt,
		); err != nil {
			r.logger.Error("failed to scan request", err)
			return nil, 0, err
		}
		requests = append(requests, &req)
	}

	return requests, total, rows.Err()
}

// Save persists an AI request
func (r *PostgresAIRequestRepository) Save(ctx context.Context, request *domain.AIRequest) error {
	query := `
		INSERT INTO ai_requests 
		(id, tenant_id, provider_id, request_type, input_text, anonymized_input,
		 output_text, model_used, tokens_used, latency_ms, status, error_message,
		 created_at, completed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		ON CONFLICT (id) DO UPDATE SET
		output_text = $7, model_used = $8, tokens_used = $9, latency_ms = $10,
		status = $11, error_message = $12, completed_at = $14
	`

	_, err := r.db.ExecContext(ctx, query,
		request.ID, request.TenantID, request.ProviderID, request.RequestType,
		request.InputText, request.AnonymizedInput, request.OutputText,
		request.ModelUsed, request.TokensUsed, request.LatencyMs,
		request.Status, request.ErrorMessage, request.CreatedAt, request.CompletedAt,
	)

	if err != nil {
		r.logger.Error("failed to save AI request", err)
		return err
	}

	return nil
}

// Delete deletes an AI request
func (r *PostgresAIRequestRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM ai_requests WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Error("failed to delete AI request", err)
		return err
	}
	return nil
}

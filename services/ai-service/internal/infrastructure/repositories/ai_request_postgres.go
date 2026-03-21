package repositories

import (
	"context"
	"database/sql"
	"time"

	"github.com/jeckersberger/rentflow/services/ai-service/internal/domain"
)

type AIRequestPostgres struct {
	db *sql.DB
}

func NewAIRequestPostgres(db *sql.DB) *AIRequestPostgres {
	return &AIRequestPostgres{db: db}
}

func (r *AIRequestPostgres) Create(ctx context.Context, req *domain.AIRequest) error {
	query := `
		INSERT INTO ai_requests
		(id, tenant_id, provider, model, prompt, anonymized_prompt, response,
		 tokens_used, cost_estimate, latency_ms, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`
	_, err := r.db.ExecContext(ctx, query,
		req.ID, req.TenantID, req.Provider, req.Model, req.Prompt,
		req.AnonymizedPrompt, req.Response, req.TokensUsed, req.CostEstimate,
		req.LatencyMs, req.Status, req.CreatedAt, req.UpdatedAt,
	)
	return err
}

func (r *AIRequestPostgres) GetByID(ctx context.Context, tenantID, id string) (*domain.AIRequest, error) {
	query := `
		SELECT id, tenant_id, provider, model, prompt, anonymized_prompt, response,
		       tokens_used, cost_estimate, latency_ms, status, created_at, updated_at
		FROM ai_requests
		WHERE id = $1 AND tenant_id = $2
	`
	var req domain.AIRequest
	err := r.db.QueryRowContext(ctx, query, id, tenantID).Scan(
		&req.ID, &req.TenantID, &req.Provider, &req.Model, &req.Prompt,
		&req.AnonymizedPrompt, &req.Response, &req.TokensUsed, &req.CostEstimate,
		&req.LatencyMs, &req.Status, &req.CreatedAt, &req.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &req, nil
}

func (r *AIRequestPostgres) ListByTenant(ctx context.Context, tenantID string) ([]*domain.AIRequest, error) {
	query := `
		SELECT id, tenant_id, provider, model, prompt, anonymized_prompt, response,
		       tokens_used, cost_estimate, latency_ms, status, created_at, updated_at
		FROM ai_requests
		WHERE tenant_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []*domain.AIRequest
	for rows.Next() {
		var req domain.AIRequest
		err := rows.Scan(
			&req.ID, &req.TenantID, &req.Provider, &req.Model, &req.Prompt,
			&req.AnonymizedPrompt, &req.Response, &req.TokensUsed, &req.CostEstimate,
			&req.LatencyMs, &req.Status, &req.CreatedAt, &req.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		requests = append(requests, &req)
	}
	return requests, rows.Err()
}

func (r *AIRequestPostgres) Update(ctx context.Context, req *domain.AIRequest) error {
	query := `
		UPDATE ai_requests
		SET response = $1, tokens_used = $2, cost_estimate = $3, latency_ms = $4,
		    status = $5, updated_at = $6
		WHERE id = $7 AND tenant_id = $8
	`
	_, err := r.db.ExecContext(ctx, query,
		req.Response, req.TokensUsed, req.CostEstimate, req.LatencyMs,
		req.Status, time.Now(), req.ID, req.TenantID,
	)
	return err
}

package repositories

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/ai-service/internal/domain"
)

// PostgresAIProviderRepository implements AIProviderRepository
type PostgresAIProviderRepository struct {
	db     *sql.DB
	logger logger.Logger
}

// NewPostgresAIProviderRepository creates a new provider repository
func NewPostgresAIProviderRepository(db *sql.DB, log logger.Logger) *PostgresAIProviderRepository {
	return &PostgresAIProviderRepository{
		db:     db,
		logger: log,
	}
}

// FindByID retrieves a provider by ID
func (r *PostgresAIProviderRepository) FindByID(ctx context.Context, id string) (*domain.AIProvider, error) {
	query := `
		SELECT id, tenant_id, name, api_endpoint, model_name, is_active,
		       priority, config, created_at
		FROM ai_providers
		WHERE id = $1
	`

	var provider domain.AIProvider
	var configJSON []byte
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&provider.ID, &provider.TenantID, &provider.Name, &provider.APIEndpoint,
		&provider.ModelName, &provider.IsActive, &provider.Priority,
		&configJSON, &provider.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		r.logger.Error("failed to query provider", err)
		return nil, err
	}

	provider.Config = make(map[string]interface{})
	if configJSON != nil {
		if err := json.Unmarshal(configJSON, &provider.Config); err != nil {
			r.logger.Error("failed to unmarshal config", err)
		}
	}

	return &provider, nil
}

// ListByTenant retrieves providers for a tenant
func (r *PostgresAIProviderRepository) ListByTenant(ctx context.Context, tenantID string) ([]*domain.AIProvider, error) {
	query := `
		SELECT id, tenant_id, name, api_endpoint, model_name, is_active,
		       priority, config, created_at
		FROM ai_providers
		WHERE tenant_id = $1
		ORDER BY priority, created_at
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		r.logger.Error("failed to query providers by tenant", err)
		return nil, err
	}
	defer rows.Close()

	var providers []*domain.AIProvider
	for rows.Next() {
		var provider domain.AIProvider
		var configJSON []byte
		if err := rows.Scan(
			&provider.ID, &provider.TenantID, &provider.Name, &provider.APIEndpoint,
			&provider.ModelName, &provider.IsActive, &provider.Priority,
			&configJSON, &provider.CreatedAt,
		); err != nil {
			r.logger.Error("failed to scan provider", err)
			return nil, err
		}

		provider.Config = make(map[string]interface{})
		if configJSON != nil {
			if err := json.Unmarshal(configJSON, &provider.Config); err != nil {
				r.logger.Error("failed to unmarshal config", err)
			}
		}

		providers = append(providers, &provider)
	}

	return providers, rows.Err()
}

// ListActive retrieves active providers for a tenant
func (r *PostgresAIProviderRepository) ListActive(ctx context.Context, tenantID string) ([]*domain.AIProvider, error) {
	query := `
		SELECT id, tenant_id, name, api_endpoint, model_name, is_active,
		       priority, config, created_at
		FROM ai_providers
		WHERE tenant_id = $1 AND is_active = true
		ORDER BY priority, created_at
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		r.logger.Error("failed to query active providers", err)
		return nil, err
	}
	defer rows.Close()

	var providers []*domain.AIProvider
	for rows.Next() {
		var provider domain.AIProvider
		var configJSON []byte
		if err := rows.Scan(
			&provider.ID, &provider.TenantID, &provider.Name, &provider.APIEndpoint,
			&provider.ModelName, &provider.IsActive, &provider.Priority,
			&configJSON, &provider.CreatedAt,
		); err != nil {
			r.logger.Error("failed to scan provider", err)
			return nil, err
		}

		provider.Config = make(map[string]interface{})
		if configJSON != nil {
			if err := json.Unmarshal(configJSON, &provider.Config); err != nil {
				r.logger.Error("failed to unmarshal config", err)
			}
		}

		providers = append(providers, &provider)
	}

	return providers, rows.Err()
}

// Save persists a provider
func (r *PostgresAIProviderRepository) Save(ctx context.Context, provider *domain.AIProvider) error {
	configJSON, err := json.Marshal(provider.Config)
	if err != nil {
		r.logger.Error("failed to marshal config", err)
		return err
	}

	query := `
		INSERT INTO ai_providers 
		(id, tenant_id, name, api_endpoint, model_name, is_active, priority, config, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (id) DO UPDATE SET
		api_endpoint = $4, model_name = $5, is_active = $6, priority = $7, config = $8
	`

	_, err = r.db.ExecContext(ctx, query,
		provider.ID, provider.TenantID, provider.Name, provider.APIEndpoint,
		provider.ModelName, provider.IsActive, provider.Priority, configJSON, provider.CreatedAt,
	)

	if err != nil {
		r.logger.Error("failed to save provider", err)
		return err
	}

	return nil
}

// Delete deletes a provider
func (r *PostgresAIProviderRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM ai_providers WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Error("failed to delete provider", err)
		return err
	}
	return nil
}

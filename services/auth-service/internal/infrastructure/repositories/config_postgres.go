package repositories

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
)

// PostgresConfigRepository implements the ConfigRepository interface using PostgreSQL
type PostgresConfigRepository struct {
	db     *sql.DB
	logger logger.Logger
}

// NewPostgresConfigRepository creates a new PostgreSQL config repository
func NewPostgresConfigRepository(db *sql.DB, log logger.Logger) *PostgresConfigRepository {
	return &PostgresConfigRepository{
		db:     db,
		logger: log,
	}
}

// GetConfig retrieves a specific config value for a tenant
func (r *PostgresConfigRepository) GetConfig(ctx context.Context, tenantID, key string) (json.RawMessage, error) {
	query := `
		SELECT config_value
		FROM auth.tenant_config
		WHERE tenant_id = $1 AND config_key = $2
	`

	var value json.RawMessage
	err := r.db.QueryRowContext(ctx, query, tenantID, key).Scan(&value)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		r.logger.Error("failed to get config", err, "tenantID", tenantID, "key", key)
		return nil, err
	}

	return value, nil
}

// SetConfig sets a config value for a tenant
func (r *PostgresConfigRepository) SetConfig(ctx context.Context, tenantID, key string, value json.RawMessage) error {
	query := `
		INSERT INTO auth.tenant_config (tenant_id, config_key, config_value, updated_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (tenant_id, config_key)
		DO UPDATE SET config_value = $3, updated_at = NOW()
	`

	_, err := r.db.ExecContext(ctx, query, tenantID, key, value)
	if err != nil {
		r.logger.Error("failed to set config", err, "tenantID", tenantID, "key", key)
		return err
	}

	return nil
}

// GetAllConfigs retrieves all config values for a tenant
func (r *PostgresConfigRepository) GetAllConfigs(ctx context.Context, tenantID string) (map[string]json.RawMessage, error) {
	query := `
		SELECT config_key, config_value
		FROM auth.tenant_config
		WHERE tenant_id = $1
		ORDER BY config_key
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		r.logger.Error("failed to get all configs", err, "tenantID", tenantID)
		return nil, err
	}
	defer rows.Close()

	configs := make(map[string]json.RawMessage)
	for rows.Next() {
		var key string
		var value json.RawMessage
		if err := rows.Scan(&key, &value); err != nil {
			r.logger.Error("failed to scan config row", err)
			continue
		}
		configs[key] = value
	}

	if err = rows.Err(); err != nil {
		r.logger.Error("error iterating configs", err)
		return nil, err
	}

	return configs, nil
}

// DeleteConfig deletes a config value for a tenant
func (r *PostgresConfigRepository) DeleteConfig(ctx context.Context, tenantID, key string) error {
	query := `DELETE FROM auth.tenant_config WHERE tenant_id = $1 AND config_key = $2`
	_, err := r.db.ExecContext(ctx, query, tenantID, key)
	if err != nil {
		r.logger.Error("failed to delete config", err, "tenantID", tenantID, "key", key)
		return err
	}
	return nil
}

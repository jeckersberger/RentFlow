package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jeckersberger/EquipFlow/services/auth/internal/domain"

	apperrors "github.com/jeckersberger/EquipFlow/pkg/common/errors"
)

// ConfigRepo implements domain.TenantConfigRepository using PostgreSQL.
type ConfigRepo struct {
	pool *pgxpool.Pool
}

// NewConfigRepo creates a new ConfigRepo.
func NewConfigRepo(pool *pgxpool.Pool) *ConfigRepo {
	return &ConfigRepo{pool: pool}
}

// Get retrieves a single tenant config entry by tenant and key.
func (r *ConfigRepo) Get(ctx context.Context, tenantID uuid.UUID, key string) (*domain.TenantConfig, error) {
	query := `
		SELECT id, tenant_id, key, value, created_at, updated_at
		FROM tenant_config
		WHERE tenant_id = $1 AND key = $2`

	c := &domain.TenantConfig{}
	err := r.pool.QueryRow(ctx, query, tenantID, key).Scan(
		&c.ID, &c.TenantID, &c.Key, &c.Value, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("config_repo: get: %w", err)
	}
	return c, nil
}

// GetAll retrieves all config entries for a tenant, ordered by key.
func (r *ConfigRepo) GetAll(ctx context.Context, tenantID uuid.UUID) ([]*domain.TenantConfig, error) {
	query := `
		SELECT id, tenant_id, key, value, created_at, updated_at
		FROM tenant_config
		WHERE tenant_id = $1
		ORDER BY key`

	rows, err := r.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("config_repo: get_all query: %w", err)
	}
	defer rows.Close()

	var configs []*domain.TenantConfig
	for rows.Next() {
		c := &domain.TenantConfig{}
		if err := rows.Scan(
			&c.ID, &c.TenantID, &c.Key, &c.Value, &c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("config_repo: get_all scan: %w", err)
		}
		configs = append(configs, c)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("config_repo: get_all rows: %w", err)
	}
	return configs, nil
}

// Set upserts a tenant config entry. If a row with the same (tenant_id, key)
// exists, its value and updated_at are overwritten.
func (r *ConfigRepo) Set(ctx context.Context, tenantID uuid.UUID, key string, value json.RawMessage) error {
	query := `
		INSERT INTO tenant_config (id, tenant_id, key, value)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (tenant_id, key) DO UPDATE
			SET value = EXCLUDED.value, updated_at = NOW()`

	id := uuid.New()
	_, err := r.pool.Exec(ctx, query, id, tenantID, key, value)
	if err != nil {
		return fmt.Errorf("config_repo: set: %w", err)
	}
	return nil
}

// Delete removes a tenant config entry by tenant and key.
func (r *ConfigRepo) Delete(ctx context.Context, tenantID uuid.UUID, key string) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM tenant_config WHERE tenant_id = $1 AND key = $2`,
		tenantID, key,
	)
	if err != nil {
		return fmt.Errorf("config_repo: delete: %w", err)
	}
	return nil
}

package repositories

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/notification-service/internal/domain"
	"github.com/lib/pq"
)

type PostgresPreferenceRepository struct {
	db  *sql.DB
	log logger.Logger
}

func NewPostgresPreferenceRepository(db *sql.DB, log logger.Logger) *PostgresPreferenceRepository {
	return &PostgresPreferenceRepository{db: db, log: log}
}

func (r *PostgresPreferenceRepository) Create(ctx context.Context, p *domain.UserPreference) (*domain.UserPreference, error) {
	query := `INSERT INTO user_preferences (id, tenant_id, user_id, event_type, channels, is_enabled, quiet_hours_start, quiet_hours_end, digest_mode, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
	RETURNING id, created_at, updated_at`

	err := r.db.QueryRowContext(ctx, query,
		p.ID, p.TenantID, p.UserID, p.EventType, pq.Array(p.Channels), p.IsEnabled, p.QuietHoursStart, p.QuietHoursEnd, p.DigestMode).
		Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)

	if err != nil {
		r.log.Error("Failed to create preference", err)
		return nil, domain.ErrDatabaseError
	}

	return p, nil
}

func (r *PostgresPreferenceRepository) GetByUserAndEvent(ctx context.Context, tenantID, userID uuid.UUID, eventType string) (*domain.UserPreference, error) {
	query := `SELECT id, tenant_id, user_id, event_type, channels, is_enabled, quiet_hours_start, quiet_hours_end, digest_mode, created_at, updated_at
	FROM user_preferences WHERE tenant_id = $1 AND user_id = $2 AND event_type = $3`

	p := &domain.UserPreference{}
	var channels pq.StringArray

	err := r.db.QueryRowContext(ctx, query, tenantID, userID, eventType).Scan(
		&p.ID, &p.TenantID, &p.UserID, &p.EventType, &channels, &p.IsEnabled, &p.QuietHoursStart, &p.QuietHoursEnd, &p.DigestMode, &p.CreatedAt, &p.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, domain.ErrPreferenceNotFound
	}
	if err != nil {
		return nil, domain.ErrDatabaseError
	}

	p.Channels = []string(channels)
	return p, nil
}

func (r *PostgresPreferenceRepository) ListByUser(ctx context.Context, tenantID, userID uuid.UUID) ([]*domain.UserPreference, error) {
	query := `SELECT id, tenant_id, user_id, event_type, channels, is_enabled, quiet_hours_start, quiet_hours_end, digest_mode, created_at, updated_at
	FROM user_preferences WHERE tenant_id = $1 AND user_id = $2`

	rows, err := r.db.QueryContext(ctx, query, tenantID, userID)
	if err != nil {
		return nil, domain.ErrDatabaseError
	}
	defer rows.Close()

	var preferences []*domain.UserPreference

	for rows.Next() {
		p := &domain.UserPreference{}
		var channels pq.StringArray

		err := rows.Scan(&p.ID, &p.TenantID, &p.UserID, &p.EventType, &channels, &p.IsEnabled, &p.QuietHoursStart, &p.QuietHoursEnd, &p.DigestMode, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			continue
		}

		p.Channels = []string(channels)
		preferences = append(preferences, p)
	}

	return preferences, nil
}

func (r *PostgresPreferenceRepository) Update(ctx context.Context, p *domain.UserPreference) (*domain.UserPreference, error) {
	p.UpdatedAt = time.Now()
	query := `UPDATE user_preferences SET channels = $1, is_enabled = $2, quiet_hours_start = $3, quiet_hours_end = $4, digest_mode = $5, updated_at = NOW()
	WHERE tenant_id = $6 AND user_id = $7 AND event_type = $8
	RETURNING id, created_at, updated_at`

	err := r.db.QueryRowContext(ctx, query,
		pq.Array(p.Channels), p.IsEnabled, p.QuietHoursStart, p.QuietHoursEnd, p.DigestMode, p.TenantID, p.UserID, p.EventType).
		Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)

	if err != nil {
		r.log.Error("Failed to update preference", err)
		return nil, domain.ErrDatabaseError
	}

	return p, nil
}

func (r *PostgresPreferenceRepository) Delete(ctx context.Context, tenantID, userID uuid.UUID, eventType string) error {
	query := `DELETE FROM user_preferences WHERE tenant_id = $1 AND user_id = $2 AND event_type = $3`
	_, err := r.db.ExecContext(ctx, query, tenantID, userID, eventType)
	if err != nil {
		r.log.Error("Failed to delete preference", err)
		return domain.ErrDatabaseError
	}
	return nil
}

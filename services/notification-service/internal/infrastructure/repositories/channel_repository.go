package repositories

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/notification-service/internal/domain"
)

type PostgresChannelRepository struct {
	db  *sql.DB
	log logger.Logger
}

func NewPostgresChannelRepository(db *sql.DB, log logger.Logger) *PostgresChannelRepository {
	return &PostgresChannelRepository{db: db, log: log}
}

func (r *PostgresChannelRepository) Create(ctx context.Context, c *domain.NotificationChannel) (*domain.NotificationChannel, error) {
	configBytes, _ := json.Marshal(c.Config)

	query := `INSERT INTO notification_channels (id, tenant_id, user_id, channel_type, config, is_active, created_at)
	VALUES ($1, $2, $3, $4, $5, $6, NOW())
	RETURNING id, created_at`

	err := r.db.QueryRowContext(ctx, query, c.ID, c.TenantID, c.UserID, c.Type, configBytes, c.IsActive).
		Scan(&c.ID, &c.CreatedAt)

	if err != nil {
		r.log.Error("Failed to create channel", err)
		return nil, domain.ErrDatabaseError
	}

	return c, nil
}

func (r *PostgresChannelRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.NotificationChannel, error) {
	query := `SELECT id, tenant_id, user_id, channel_type, config, is_active, created_at FROM notification_channels WHERE id = $1`

	c := &domain.NotificationChannel{}
	var configBytes []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&c.ID, &c.TenantID, &c.UserID, &c.Type, &configBytes, &c.IsActive, &c.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, domain.ErrChannelNotFound
	}
	if err != nil {
		return nil, domain.ErrDatabaseError
	}

	json.Unmarshal(configBytes, &c.Config)
	return c, nil
}

func (r *PostgresChannelRepository) ListByUser(ctx context.Context, tenantID, userID uuid.UUID) ([]*domain.NotificationChannel, error) {
	query := `SELECT id, tenant_id, user_id, channel_type, config, is_active, created_at
	FROM notification_channels WHERE tenant_id = $1 AND user_id = $2 ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, tenantID, userID)
	if err != nil {
		return nil, domain.ErrDatabaseError
	}
	defer rows.Close()

	var channels []*domain.NotificationChannel

	for rows.Next() {
		c := &domain.NotificationChannel{}
		var configBytes []byte

		err := rows.Scan(&c.ID, &c.TenantID, &c.UserID, &c.Type, &configBytes, &c.IsActive, &c.CreatedAt)
		if err != nil {
			continue
		}

		json.Unmarshal(configBytes, &c.Config)
		channels = append(channels, c)
	}

	return channels, nil
}

func (r *PostgresChannelRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM notification_channels WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.log.Error("Failed to delete channel", err)
		return domain.ErrDatabaseError
	}
	return nil
}

func (r *PostgresChannelRepository) ListActive(ctx context.Context, tenantID, userID uuid.UUID, channelType string) ([]*domain.NotificationChannel, error) {
	query := `SELECT id, tenant_id, user_id, channel_type, config, is_active, created_at
	FROM notification_channels WHERE tenant_id = $1 AND user_id = $2 AND channel_type = $3 AND is_active = true`

	rows, err := r.db.QueryContext(ctx, query, tenantID, userID, channelType)
	if err != nil {
		return nil, domain.ErrDatabaseError
	}
	defer rows.Close()

	var channels []*domain.NotificationChannel

	for rows.Next() {
		c := &domain.NotificationChannel{}
		var configBytes []byte

		err := rows.Scan(&c.ID, &c.TenantID, &c.UserID, &c.Type, &configBytes, &c.IsActive, &c.CreatedAt)
		if err != nil {
			continue
		}

		json.Unmarshal(configBytes, &c.Config)
		channels = append(channels, c)
	}

	return channels, nil
}

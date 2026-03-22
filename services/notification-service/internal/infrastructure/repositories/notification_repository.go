package repositories

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/notification-service/internal/domain"
)

type PostgresNotificationRepository struct {
	db  *sql.DB
	log logger.Logger
}

func NewPostgresNotificationRepository(db *sql.DB, log logger.Logger) *PostgresNotificationRepository {
	return &PostgresNotificationRepository{db: db, log: log}
}

func (r *PostgresNotificationRepository) Create(ctx context.Context, n *domain.Notification) (*domain.Notification, error) {
	dataBytes, _ := json.Marshal(n.Data)

	query := `INSERT INTO notifications (id, tenant_id, user_id, event_type, title, body, data, status, created_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
	RETURNING id, created_at`

	err := r.db.QueryRowContext(ctx, query,
		n.ID, n.TenantID, n.UserID, n.EventType, n.Title, n.Body, dataBytes, n.Status).
		Scan(&n.ID, &n.CreatedAt)

	if err != nil {
		r.log.Error("Failed to create notification", err)
		return nil, domain.ErrDatabaseError
	}

	return n, nil
}

func (r *PostgresNotificationRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Notification, error) {
	query := `SELECT id, tenant_id, user_id, event_type, title, body, data, status, is_read, created_at
	FROM notifications WHERE id = $1`

	n := &domain.Notification{}
	var dataBytes []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&n.ID, &n.TenantID, &n.UserID, &n.EventType, &n.Title, &n.Body, &dataBytes, &n.Status, &n.IsRead, &n.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, domain.ErrNotificationNotFound
	}
	if err != nil {
		return nil, domain.ErrDatabaseError
	}

	json.Unmarshal(dataBytes, &n.Data)
	return n, nil
}

func (r *PostgresNotificationRepository) ListByUser(ctx context.Context, tenantID, userID uuid.UUID) ([]*domain.Notification, error) {
	query := `SELECT id, tenant_id, user_id, event_type, title, body, data, status, is_read, created_at
	FROM notifications WHERE tenant_id = $1 AND user_id = $2 ORDER BY created_at DESC LIMIT 100`

	rows, err := r.db.QueryContext(ctx, query, tenantID, userID)
	if err != nil {
		return nil, domain.ErrDatabaseError
	}
	defer rows.Close()

	var notifications []*domain.Notification

	for rows.Next() {
		n := &domain.Notification{}
		var dataBytes []byte

		err := rows.Scan(&n.ID, &n.TenantID, &n.UserID, &n.EventType, &n.Title, &n.Body, &dataBytes, &n.Status, &n.IsRead, &n.CreatedAt)
		if err != nil {
			r.log.Error("Failed to scan notification", err)
			continue
		}

		json.Unmarshal(dataBytes, &n.Data)
		notifications = append(notifications, n)
	}

	return notifications, nil
}

func (r *PostgresNotificationRepository) ListUnread(ctx context.Context, tenantID, userID uuid.UUID) ([]*domain.Notification, error) {
	query := `SELECT id, tenant_id, user_id, event_type, title, body, data, status, is_read, created_at
	FROM notifications WHERE tenant_id = $1 AND user_id = $2 AND is_read = false ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, tenantID, userID)
	if err != nil {
		return nil, domain.ErrDatabaseError
	}
	defer rows.Close()

	var notifications []*domain.Notification

	for rows.Next() {
		n := &domain.Notification{}
		var dataBytes []byte

		err := rows.Scan(&n.ID, &n.TenantID, &n.UserID, &n.EventType, &n.Title, &n.Body, &dataBytes, &n.Status, &n.IsRead, &n.CreatedAt)
		if err != nil {
			continue
		}

		json.Unmarshal(dataBytes, &n.Data)
		notifications = append(notifications, n)
	}

	return notifications, nil
}

func (r *PostgresNotificationRepository) MarkAsRead(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE notifications SET is_read = true, read_at = NOW() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.log.Error("Failed to mark notification as read", err)
		return domain.ErrDatabaseError
	}
	return nil
}

func (r *PostgresNotificationRepository) MarkAllAsRead(ctx context.Context, tenantID, userID uuid.UUID) error {
	query := `UPDATE notifications SET is_read = true, read_at = NOW() WHERE tenant_id = $1 AND user_id = $2 AND is_read = false`
	_, err := r.db.ExecContext(ctx, query, tenantID, userID)
	if err != nil {
		r.log.Error("Failed to mark all notifications as read", err)
		return domain.ErrDatabaseError
	}
	return nil
}

func (r *PostgresNotificationRepository) GetUnreadCount(ctx context.Context, tenantID, userID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM notifications WHERE tenant_id = $1 AND user_id = $2 AND is_read = false`
	var count int
	err := r.db.QueryRowContext(ctx, query, tenantID, userID).Scan(&count)
	if err != nil {
		return 0, domain.ErrDatabaseError
	}
	return count, nil
}

func (r *PostgresNotificationRepository) GetTodayCount(ctx context.Context, tenantID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM notifications WHERE tenant_id = $1 AND created_at::date = CURRENT_DATE`
	var count int
	err := r.db.QueryRowContext(ctx, query, tenantID).Scan(&count)
	if err != nil {
		return 0, domain.ErrDatabaseError
	}
	return count, nil
}

package repositories

import (
	"context"
	"github.com/jeckersberger/rentflow/pkg/common/database"
	"github.com/jeckersberger/rentflow/services/notification-service/internal/domain"
)

type NotificationPostgres struct {
	db *database.PostgresPool
}

func NewNotificationPostgres(db *database.PostgresPool) *NotificationPostgres {
	return &NotificationPostgres{db: db}
}

func (r *NotificationPostgres) Create(ctx context.Context, notif *domain.Notification) error {
	query := `INSERT INTO notifications (id, tenant_id, user_id, type, channel, title, message, link, read, sent_at, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`
	_, err := r.db.Exec(ctx, query, notif.ID, notif.TenantID, notif.UserID, notif.Type, notif.Channel, notif.Title, notif.Message, notif.Link, notif.Read, notif.SentAt, notif.CreatedAt)
	return err
}

func (r *NotificationPostgres) GetByID(ctx context.Context, tenantID, id string) (*domain.Notification, error) {
	query := `SELECT id, tenant_id, user_id, type, channel, title, message, link, read, sent_at, created_at FROM notifications WHERE id = $1 AND tenant_id = $2`
	row := r.db.QueryRow(ctx, query, id, tenantID)
	notif := &domain.Notification{}
	err := row.Scan(&notif.ID, &notif.TenantID, &notif.UserID, &notif.Type, &notif.Channel, &notif.Title, &notif.Message, &notif.Link, &notif.Read, &notif.SentAt, &notif.CreatedAt)
	if err != nil {
		return nil, err
	}
	return notif, nil
}

func (r *NotificationPostgres) ListByUser(ctx context.Context, tenantID, userID string) ([]*domain.Notification, error) {
	query := `SELECT id, tenant_id, user_id, type, channel, title, message, link, read, sent_at, created_at FROM notifications WHERE tenant_id = $1 AND user_id = $2 ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query, tenantID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifs []*domain.Notification
	for rows.Next() {
		notif := &domain.Notification{}
		if err := rows.Scan(&notif.ID, &notif.TenantID, &notif.UserID, &notif.Type, &notif.Channel, &notif.Title, &notif.Message, &notif.Link, &notif.Read, &notif.SentAt, &notif.CreatedAt); err != nil {
			return nil, err
		}
		notifs = append(notifs, notif)
	}
	return notifs, rows.Err()
}

func (r *NotificationPostgres) ListUnread(ctx context.Context, tenantID, userID string) ([]*domain.Notification, error) {
	query := `SELECT id, tenant_id, user_id, type, channel, title, message, link, read, sent_at, created_at FROM notifications WHERE tenant_id = $1 AND user_id = $2 AND read = false ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query, tenantID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifs []*domain.Notification
	for rows.Next() {
		notif := &domain.Notification{}
		if err := rows.Scan(&notif.ID, &notif.TenantID, &notif.UserID, &notif.Type, &notif.Channel, &notif.Title, &notif.Message, &notif.Link, &notif.Read, &notif.SentAt, &notif.CreatedAt); err != nil {
			return nil, err
		}
		notifs = append(notifs, notif)
	}
	return notifs, rows.Err()
}

func (r *NotificationPostgres) MarkAsRead(ctx context.Context, tenantID, id string) error {
	query := `UPDATE notifications SET read = true WHERE id = $1 AND tenant_id = $2`
	_, err := r.db.Exec(ctx, query, id, tenantID)
	return err
}

func (r *NotificationPostgres) MarkAllAsRead(ctx context.Context, tenantID, userID string) error {
	query := `UPDATE notifications SET read = true WHERE tenant_id = $1 AND user_id = $2`
	_, err := r.db.Exec(ctx, query, tenantID, userID)
	return err
}

func (r *NotificationPostgres) Delete(ctx context.Context, tenantID, id string) error {
	query := `DELETE FROM notifications WHERE id = $1 AND tenant_id = $2`
	_, err := r.db.Exec(ctx, query, id, tenantID)
	return err
}

type PreferencePostgres struct {
	db *database.PostgresPool
}

func NewPreferencePostgres(db *database.PostgresPool) *PreferencePostgres {
	return &PreferencePostgres{db: db}
}

func (r *PreferencePostgres) Create(ctx context.Context, pref *domain.NotificationPreference) error {
	query := `INSERT INTO notification_preferences (id, tenant_id, user_id, channel, event_type, enabled, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := r.db.Exec(ctx, query, pref.ID, pref.TenantID, pref.UserID, pref.Channel, pref.EventType, pref.Enabled, pref.CreatedAt, pref.UpdatedAt)
	return err
}

func (r *PreferencePostgres) GetByID(ctx context.Context, tenantID, id string) (*domain.NotificationPreference, error) {
	query := `SELECT id, tenant_id, user_id, channel, event_type, enabled, created_at, updated_at FROM notification_preferences WHERE id = $1 AND tenant_id = $2`
	row := r.db.QueryRow(ctx, query, id, tenantID)
	pref := &domain.NotificationPreference{}
	err := row.Scan(&pref.ID, &pref.TenantID, &pref.UserID, &pref.Channel, &pref.EventType, &pref.Enabled, &pref.CreatedAt, &pref.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return pref, nil
}

func (r *PreferencePostgres) ListByUser(ctx context.Context, tenantID, userID string) ([]*domain.NotificationPreference, error) {
	query := `SELECT id, tenant_id, user_id, channel, event_type, enabled, created_at, updated_at FROM notification_preferences WHERE tenant_id = $1 AND user_id = $2`
	rows, err := r.db.Query(ctx, query, tenantID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prefs []*domain.NotificationPreference
	for rows.Next() {
		pref := &domain.NotificationPreference{}
		if err := rows.Scan(&pref.ID, &pref.TenantID, &pref.UserID, &pref.Channel, &pref.EventType, &pref.Enabled, &pref.CreatedAt, &pref.UpdatedAt); err != nil {
			return nil, err
		}
		prefs = append(prefs, pref)
	}
	return prefs, rows.Err()
}

func (r *PreferencePostgres) Update(ctx context.Context, pref *domain.NotificationPreference) error {
	query := `UPDATE notification_preferences SET enabled = $1, updated_at = $2 WHERE id = $3 AND tenant_id = $4`
	_, err := r.db.Exec(ctx, query, pref.Enabled, pref.UpdatedAt, pref.ID, pref.TenantID)
	return err
}

func (r *PreferencePostgres) Delete(ctx context.Context, tenantID, id string) error {
	query := `DELETE FROM notification_preferences WHERE id = $1 AND tenant_id = $2`
	_, err := r.db.Exec(ctx, query, id, tenantID)
	return err
}

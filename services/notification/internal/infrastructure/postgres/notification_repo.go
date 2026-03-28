package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	apperrors "github.com/jeckersberger/EquipFlow/pkg/common/errors"
	"github.com/jeckersberger/EquipFlow/services/notification/internal/domain"
)

const notificationColumns = `
	id, tenant_id, user_id, type, title, message,
	reference_id, reference_type, is_read, read_at, created_at`

type NotificationRepo struct {
	pool *pgxpool.Pool
}

func NewNotificationRepo(pool *pgxpool.Pool) *NotificationRepo {
	return &NotificationRepo{pool: pool}
}

func scanNotification(row pgx.Row) (*domain.Notification, error) {
	n := &domain.Notification{}
	var (
		message       *string
		referenceID   *uuid.UUID
		referenceType *string
		readAt        *time.Time
	)

	err := row.Scan(
		&n.ID, &n.TenantID, &n.UserID, &n.Type, &n.Title, &message,
		&referenceID, &referenceType, &n.IsRead, &readAt, &n.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	if message != nil {
		n.Message = *message
	}
	n.ReferenceID = referenceID
	if referenceType != nil {
		n.ReferenceType = *referenceType
	}
	n.ReadAt = readAt

	return n, nil
}

func (r *NotificationRepo) Create(ctx context.Context, n *domain.Notification) error {
	query := `
		INSERT INTO notifications (
			id, tenant_id, user_id, type, title, message,
			reference_id, reference_type
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING created_at`

	err := r.pool.QueryRow(ctx, query,
		n.ID, n.TenantID, n.UserID, n.Type, n.Title,
		nilIfEmpty(n.Message), n.ReferenceID, nilIfEmpty(n.ReferenceType),
	).Scan(&n.CreatedAt)
	if err != nil {
		return fmt.Errorf("notification_repo: create: %w", err)
	}
	return nil
}

func (r *NotificationRepo) GetByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*domain.Notification, error) {
	query := fmt.Sprintf(`SELECT %s FROM notifications WHERE id = $1 AND tenant_id = $2`, notificationColumns)
	n, err := scanNotification(r.pool.QueryRow(ctx, query, id, tenantID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("notification_repo: get_by_id: %w", err)
	}
	return n, nil
}

func (r *NotificationRepo) List(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID, filter domain.NotificationFilter) ([]*domain.Notification, int64, error) {
	conditions := []string{"tenant_id = $1", "user_id = $2"}
	args := []interface{}{tenantID, userID}
	argIdx := 3

	if filter.IsRead != nil {
		conditions = append(conditions, fmt.Sprintf("is_read = $%d", argIdx))
		args = append(args, *filter.IsRead)
		argIdx++
	}

	where := strings.Join(conditions, " AND ")

	var total int64
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM notifications WHERE %s`, where)
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("notification_repo: list count: %w", err)
	}

	page := filter.Page
	if page < 1 {
		page = 1
	}
	perPage := filter.PerPage
	if perPage < 1 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	dataQuery := fmt.Sprintf(
		`SELECT %s FROM notifications WHERE %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		notificationColumns, where, argIdx, argIdx+1,
	)
	args = append(args, perPage, offset)

	rows, err := r.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("notification_repo: list query: %w", err)
	}
	defer rows.Close()

	var items []*domain.Notification
	for rows.Next() {
		n, scanErr := scanNotification(rows)
		if scanErr != nil {
			return nil, 0, fmt.Errorf("notification_repo: list scan: %w", scanErr)
		}
		items = append(items, n)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("notification_repo: list rows: %w", err)
	}
	return items, total, nil
}

func (r *NotificationRepo) UnreadCount(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM notifications WHERE tenant_id = $1 AND user_id = $2 AND is_read = FALSE`
	err := r.pool.QueryRow(ctx, query, tenantID, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("notification_repo: unread_count: %w", err)
	}
	return count, nil
}

func (r *NotificationRepo) MarkAsRead(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error {
	query := `UPDATE notifications SET is_read = TRUE, read_at = NOW() WHERE id = $1 AND tenant_id = $2 AND is_read = FALSE`
	tag, err := r.pool.Exec(ctx, query, id, tenantID)
	if err != nil {
		return fmt.Errorf("notification_repo: mark_as_read: %w", err)
	}
	if tag.RowsAffected() == 0 {
		// Check if notification exists at all.
		var exists bool
		r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM notifications WHERE id = $1 AND tenant_id = $2)`, id, tenantID).Scan(&exists)
		if !exists {
			return apperrors.ErrNotFound
		}
	}
	return nil
}

func (r *NotificationRepo) MarkAllRead(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID) (int64, error) {
	query := `UPDATE notifications SET is_read = TRUE, read_at = NOW() WHERE tenant_id = $1 AND user_id = $2 AND is_read = FALSE`
	tag, err := r.pool.Exec(ctx, query, tenantID, userID)
	if err != nil {
		return 0, fmt.Errorf("notification_repo: mark_all_read: %w", err)
	}
	return tag.RowsAffected(), nil
}

func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

package domain

import (
	"context"

	"github.com/google/uuid"
)

type NotificationRepository interface {
	Create(ctx context.Context, n *Notification) error
	GetByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*Notification, error)
	List(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID, filter NotificationFilter) ([]*Notification, int64, error)
	UnreadCount(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID) (int64, error)
	MarkAsRead(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error
	MarkAllRead(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID) (int64, error)
}

type NotificationPreferenceRepository interface {
	List(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID) ([]*NotificationPreference, error)
	Upsert(ctx context.Context, pref *NotificationPreference) error
}

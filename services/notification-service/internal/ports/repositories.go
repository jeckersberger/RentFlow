package ports

import (
	"context"
	"github.com/jeckersberger/rentflow/services/notification-service/internal/domain"
)

type NotificationRepository interface {
	Create(ctx context.Context, notif *domain.Notification) error
	GetByID(ctx context.Context, tenantID, id string) (*domain.Notification, error)
	ListByUser(ctx context.Context, tenantID, userID string) ([]*domain.Notification, error)
	ListUnread(ctx context.Context, tenantID, userID string) ([]*domain.Notification, error)
	MarkAsRead(ctx context.Context, tenantID, id string) error
	MarkAllAsRead(ctx context.Context, tenantID, userID string) error
	Delete(ctx context.Context, tenantID, id string) error
}

type PreferenceRepository interface {
	Create(ctx context.Context, pref *domain.NotificationPreference) error
	GetByID(ctx context.Context, tenantID, id string) (*domain.NotificationPreference, error)
	ListByUser(ctx context.Context, tenantID, userID string) ([]*domain.NotificationPreference, error)
	Update(ctx context.Context, pref *domain.NotificationPreference) error
	Delete(ctx context.Context, tenantID, id string) error
}

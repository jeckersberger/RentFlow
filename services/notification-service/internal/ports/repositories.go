package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/jeckersberger/rentflow/services/notification-service/internal/domain"
)

type NotificationRepository interface {
	Create(ctx context.Context, n *domain.Notification) (*domain.Notification, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Notification, error)
	ListByUser(ctx context.Context, tenantID, userID uuid.UUID) ([]*domain.Notification, error)
	ListUnread(ctx context.Context, tenantID, userID uuid.UUID) ([]*domain.Notification, error)
	ListQueued(ctx context.Context, tenantID uuid.UUID) ([]*domain.Notification, error)
	MarkAsRead(ctx context.Context, id uuid.UUID) error
	MarkAllAsRead(ctx context.Context, tenantID, userID uuid.UUID) error
	MarkAsSent(ctx context.Context, id uuid.UUID) error
	GetUnreadCount(ctx context.Context, tenantID, userID uuid.UUID) (int, error)
	GetTodayCount(ctx context.Context, tenantID uuid.UUID) (int, error)
}

type ChannelRepository interface {
	Create(ctx context.Context, c *domain.NotificationChannel) (*domain.NotificationChannel, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.NotificationChannel, error)
	ListByUser(ctx context.Context, tenantID, userID uuid.UUID) ([]*domain.NotificationChannel, error)
	Delete(ctx context.Context, id uuid.UUID) error
	ListActive(ctx context.Context, tenantID, userID uuid.UUID, channelType string) ([]*domain.NotificationChannel, error)
}

type PreferenceRepository interface {
	Create(ctx context.Context, p *domain.UserPreference) (*domain.UserPreference, error)
	GetByUserAndEvent(ctx context.Context, tenantID, userID uuid.UUID, eventType string) (*domain.UserPreference, error)
	ListByUser(ctx context.Context, tenantID, userID uuid.UUID) ([]*domain.UserPreference, error)
	Update(ctx context.Context, p *domain.UserPreference) (*domain.UserPreference, error)
	Delete(ctx context.Context, tenantID, userID uuid.UUID, eventType string) error
}

// MailRepository — Zugriff auf E-Mail-Daten
type MailRepository interface {
	// E-Mails
	CreateMail(ctx context.Context, m *domain.Mail) (*domain.Mail, error)
	GetMailByID(ctx context.Context, id uuid.UUID) (*domain.Mail, error)
	ListMails(ctx context.Context, tenantID uuid.UUID, filter domain.MailFilter) ([]*domain.Mail, error)
	MarkMailAsRead(ctx context.Context, id uuid.UUID) error
	AssignMailToProject(ctx context.Context, mailID, projectID uuid.UUID, confidence float64) error

	// Postfächer
	CreateMailbox(ctx context.Context, mb *domain.Mailbox) (*domain.Mailbox, error)
	GetMailboxByID(ctx context.Context, id uuid.UUID) (*domain.Mailbox, error)
	ListMailboxes(ctx context.Context, tenantID uuid.UUID) ([]*domain.Mailbox, error)
	UpdateMailbox(ctx context.Context, mb *domain.Mailbox) (*domain.Mailbox, error)
	DeleteMailbox(ctx context.Context, id uuid.UUID) error

	// Anhänge
	CreateAttachment(ctx context.Context, a *domain.MailAttachment) (*domain.MailAttachment, error)
	ListAttachments(ctx context.Context, mailID uuid.UUID) ([]domain.MailAttachment, error)
}

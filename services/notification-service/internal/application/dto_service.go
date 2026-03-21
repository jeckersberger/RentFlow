package application

import (
	"context"
	"fmt"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/notification-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/notification-service/internal/ports"
)

type NotificationDTO struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Type      string    `json:"type"`
	Channel   string    `json:"channel"`
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	Link      string    `json:"link,omitempty"`
	Read      bool      `json:"read"`
	SentAt    time.Time `json:"sent_at"`
	CreatedAt time.Time `json:"created_at"`
}

type PreferenceDTO struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	Channel   string `json:"channel"`
	EventType string `json:"event_type"`
	Enabled   bool   `json:"enabled"`
}

type SendNotificationCommand struct {
	TenantID  string `json:"tenant_id"`
	UserID    string `json:"user_id"`
	Type      string `json:"type"`
	Channel   string `json:"channel"`
	Title     string `json:"title"`
	Message   string `json:"message"`
	Link      string `json:"link,omitempty"`
}

type NotificationService struct {
	notifRepo ports.NotificationRepository
	prefRepo  ports.PreferenceRepository
	logger    logger.Logger
}

func NewNotificationService(
	notifRepo ports.NotificationRepository,
	prefRepo ports.PreferenceRepository,
	log logger.Logger,
) *NotificationService {
	return &NotificationService{
		notifRepo: notifRepo,
		prefRepo:  prefRepo,
		logger:    log,
	}
}

func (s *NotificationService) SendNotification(ctx context.Context, cmd SendNotificationCommand) (*NotificationDTO, error) {
	if cmd.TenantID == "" || cmd.UserID == "" {
		return nil, domain.ErrInvalidInput
	}

	notif := &domain.Notification{
		ID:        fmt.Sprintf("notif_%d", time.Now().UnixNano()),
		TenantID:  cmd.TenantID,
		UserID:    cmd.UserID,
		Type:      domain.NotificationType(cmd.Type),
		Channel:   domain.Channel(cmd.Channel),
		Title:     cmd.Title,
		Message:   cmd.Message,
		Link:      cmd.Link,
		Read:      false,
		SentAt:    time.Now(),
		CreatedAt: time.Now(),
	}

	if err := s.notifRepo.Create(ctx, notif); err != nil {
		s.logger.Error("Failed to send notification", err)
		return nil, err
	}

	return &NotificationDTO{
		ID:        notif.ID,
		UserID:    notif.UserID,
		Type:      string(notif.Type),
		Channel:   string(notif.Channel),
		Title:     notif.Title,
		Message:   notif.Message,
		Link:      notif.Link,
		Read:      notif.Read,
		SentAt:    notif.SentAt,
		CreatedAt: notif.CreatedAt,
	}, nil
}

func (s *NotificationService) GetNotification(ctx context.Context, tenantID, id string) (*NotificationDTO, error) {
	notif, err := s.notifRepo.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	if notif == nil {
		return nil, domain.ErrNotificationNotFound
	}

	return &NotificationDTO{
		ID:        notif.ID,
		UserID:    notif.UserID,
		Type:      string(notif.Type),
		Channel:   string(notif.Channel),
		Title:     notif.Title,
		Message:   notif.Message,
		Link:      notif.Link,
		Read:      notif.Read,
		SentAt:    notif.SentAt,
		CreatedAt: notif.CreatedAt,
	}, nil
}

func (s *NotificationService) ListNotifications(ctx context.Context, tenantID, userID string) ([]*NotificationDTO, error) {
	notifs, err := s.notifRepo.ListByUser(ctx, tenantID, userID)
	if err != nil {
		return nil, err
	}

	dtos := make([]*NotificationDTO, len(notifs))
	for i, n := range notifs {
		dtos[i] = &NotificationDTO{
			ID:        n.ID,
			UserID:    n.UserID,
			Type:      string(n.Type),
			Channel:   string(n.Channel),
			Title:     n.Title,
			Message:   n.Message,
			Link:      n.Link,
			Read:      n.Read,
			SentAt:    n.SentAt,
			CreatedAt: n.CreatedAt,
		}
	}
	return dtos, nil
}

func (s *NotificationService) ListUnread(ctx context.Context, tenantID, userID string) ([]*NotificationDTO, error) {
	notifs, err := s.notifRepo.ListUnread(ctx, tenantID, userID)
	if err != nil {
		return nil, err
	}

	dtos := make([]*NotificationDTO, len(notifs))
	for i, n := range notifs {
		dtos[i] = &NotificationDTO{
			ID:        n.ID,
			UserID:    n.UserID,
			Type:      string(n.Type),
			Channel:   string(n.Channel),
			Title:     n.Title,
			Message:   n.Message,
			Link:      n.Link,
			Read:      n.Read,
			SentAt:    n.SentAt,
			CreatedAt: n.CreatedAt,
		}
	}
	return dtos, nil
}

func (s *NotificationService) GetUnreadCount(ctx context.Context, tenantID, userID string) (int, error) {
	notifs, err := s.notifRepo.ListUnread(ctx, tenantID, userID)
	if err != nil {
		return 0, err
	}
	return len(notifs), nil
}

func (s *NotificationService) MarkAsRead(ctx context.Context, tenantID, id string) error {
	return s.notifRepo.MarkAsRead(ctx, tenantID, id)
}

func (s *NotificationService) MarkAllAsRead(ctx context.Context, tenantID, userID string) error {
	return s.notifRepo.MarkAllAsRead(ctx, tenantID, userID)
}

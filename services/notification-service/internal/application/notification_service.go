package application

import (
	"context"

	"github.com/google/uuid"
	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/notification-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/notification-service/internal/ports"
)

type NotificationService struct {
	notificationRepo ports.NotificationRepository
	channelRepo      ports.ChannelRepository
	preferenceRepo   ports.PreferenceRepository
	log              logger.Logger
}

func NewNotificationService(
	notificationRepo ports.NotificationRepository,
	channelRepo ports.ChannelRepository,
	preferenceRepo ports.PreferenceRepository,
	log logger.Logger,
) *NotificationService {
	return &NotificationService{
		notificationRepo: notificationRepo,
		channelRepo:      channelRepo,
		preferenceRepo:   preferenceRepo,
		log:              log,
	}
}

func (s *NotificationService) SendNotification(ctx context.Context, cmd domain.SendNotificationCmd) (*domain.Notification, error) {
	notification := &domain.Notification{
		ID:        uuid.New(),
		TenantID:  cmd.TenantID,
		UserID:    cmd.UserID,
		EventType: cmd.EventType,
		Title:     cmd.Title,
		Body:      cmd.Body,
		Data:      cmd.Data,
		Status:    "sent",
	}

	created, err := s.notificationRepo.Create(ctx, notification)
	if err != nil {
		s.log.Error("Failed to create notification", err)
		return nil, err
	}

	s.log.Info("Notification sent", "notificationID", created.ID, "userID", cmd.UserID, "eventType", cmd.EventType)
	return created, nil
}

func (s *NotificationService) BroadcastNotification(ctx context.Context, cmd domain.BroadcastNotificationCmd) (int, error) {
	count := 0
	for _, userID := range cmd.UserIDs {
		notification := &domain.Notification{
			ID:        uuid.New(),
			TenantID:  cmd.TenantID,
			UserID:    userID,
			EventType: cmd.EventType,
			Title:     cmd.Title,
			Body:      cmd.Body,
			Data:      cmd.Data,
			Status:    "sent",
		}

		_, err := s.notificationRepo.Create(ctx, notification)
		if err != nil {
			s.log.Error("Failed to broadcast to user", err, "userID", userID)
			continue
		}
		count++
	}

	s.log.Info("Broadcast completed", "totalSent", count, "requested", len(cmd.UserIDs))
	return count, nil
}

func (s *NotificationService) GetNotifications(ctx context.Context, tenantID, userID uuid.UUID) ([]*domain.Notification, error) {
	return s.notificationRepo.ListByUser(ctx, tenantID, userID)
}

func (s *NotificationService) GetUnreadCount(ctx context.Context, tenantID, userID uuid.UUID) (int, error) {
	return s.notificationRepo.GetUnreadCount(ctx, tenantID, userID)
}

func (s *NotificationService) MarkAsRead(ctx context.Context, notificationID uuid.UUID) error {
	return s.notificationRepo.MarkAsRead(ctx, notificationID)
}

func (s *NotificationService) MarkAllAsRead(ctx context.Context, tenantID, userID uuid.UUID) error {
	return s.notificationRepo.MarkAllAsRead(ctx, tenantID, userID)
}

func (s *NotificationService) GetDashboardStats(ctx context.Context, tenantID uuid.UUID) (*domain.GetDashboardStatsResponse, error) {
	todayCount, err := s.notificationRepo.GetTodayCount(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	return &domain.GetDashboardStatsResponse{
		TodayNotificationCount: todayCount,
		DeliveryRate:           0.95,
		TopEventTypes:          []string{"project_created", "invoice_overdue"},
		PendingNotifications:   0,
	}, nil
}

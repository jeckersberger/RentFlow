package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/notification-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/notification-service/internal/ports"
)

type DigestService struct {
	notificationRepo ports.NotificationRepository
	channelRepo      ports.ChannelRepository
	preferenceRepo   ports.PreferenceRepository
	drivers          map[string]ChannelDriver
	log              logger.Logger
}

func NewDigestService(
	notificationRepo ports.NotificationRepository,
	channelRepo ports.ChannelRepository,
	preferenceRepo ports.PreferenceRepository,
	log logger.Logger,
) *DigestService {
	return &DigestService{
		notificationRepo: notificationRepo,
		channelRepo:      channelRepo,
		preferenceRepo:   preferenceRepo,
		drivers:          make(map[string]ChannelDriver),
		log:              log,
	}
}

// RegisterDriver registers a channel driver
func (s *DigestService) RegisterDriver(driverType string, driver ChannelDriver) {
	s.drivers[driverType] = driver
	s.log.Info("Registered channel driver", "type", driverType)
}

func (s *DigestService) IsQuietHours(ctx context.Context, tenantID, userID uuid.UUID, now time.Time) (bool, error) {
	prefs, err := s.preferenceRepo.ListByUser(ctx, tenantID, userID)
	if err != nil {
		return false, err
	}

	for _, pref := range prefs {
		if pref.QuietHoursStart != nil && pref.QuietHoursEnd != nil {
			startHour := pref.QuietHoursStart.Hour()
			endHour := pref.QuietHoursEnd.Hour()
			currentHour := now.Hour()

			if startHour < endHour {
				if currentHour >= startHour && currentHour < endHour {
					return true, nil
				}
			} else if startHour > endHour {
				if currentHour >= startHour || currentHour < endHour {
					return true, nil
				}
			}
		}
	}

	return false, nil
}

func (s *DigestService) ProcessScheduledNotifications(ctx context.Context, tenantID uuid.UUID) error {
	s.log.Info("Processing scheduled notifications", "tenantID", tenantID)

	// Query all queued/scheduled notifications for the tenant
	notifications, err := s.notificationRepo.ListQueued(ctx, tenantID)
	if err != nil {
		s.log.Error("Failed to list queued notifications", err, "tenantID", tenantID)
		return fmt.Errorf("failed to list queued notifications: %w", err)
	}

	if len(notifications) == 0 {
		s.log.Info("No queued notifications to process", "tenantID", tenantID)
		return nil
	}

	// Group notifications by user and channel preference
	notificationsByUserChannel := make(map[string]map[string][]*domain.Notification)

	for _, notif := range notifications {
		userID := notif.UserID.String()

		// Get user preferences to determine which channels to use
		prefs, err := s.preferenceRepo.ListByUser(ctx, tenantID, notif.UserID)
		if err != nil {
			s.log.Error("Failed to get user preferences", err, "userID", userID, "tenantID", tenantID)
			continue
		}

		// Find matching preferences for this notification's event type
		var channels []string
		for _, pref := range prefs {
			if pref.EventType == notif.EventType && pref.IsEnabled {
				channels = pref.Channels
				break
			}
		}

		// If no preference found, skip
		if len(channels) == 0 {
			s.log.Warn("No channel preference found for notification", "userID", userID, "eventType", notif.EventType)
			continue
		}

		// Group by user and channel
		if notificationsByUserChannel[userID] == nil {
			notificationsByUserChannel[userID] = make(map[string][]*domain.Notification)
		}

		for _, channel := range channels {
			notificationsByUserChannel[userID][channel] = append(notificationsByUserChannel[userID][channel], notif)
		}
	}

	// Process each user's notifications per channel
	for userIDStr, channelNotifs := range notificationsByUserChannel {
		userID, _ := uuid.Parse(userIDStr)

		for channelType, notifs := range channelNotifs {
			// Get channel configuration from database
			channels, err := s.channelRepo.ListActive(ctx, tenantID, userID, channelType)
			if err != nil || len(channels) == 0 {
				s.log.Warn("No active channel found for sending", "channelType", channelType, "userID", userIDStr)
				continue
			}

			channelConfig := channels[0].Config

			// Get the driver for this channel type
			driver, ok := s.drivers[channelType]
			if !ok {
				s.log.Warn("No driver registered for channel type", "channelType", channelType)
				continue
			}

			// Send each notification using the channel driver
			for _, notif := range notifs {
				err := driver.Send(ctx, notif, channelConfig)
				if err != nil {
					s.log.Error("Failed to send notification", err, "notificationID", notif.ID, "channelType", channelType)
					continue
				}

				// Mark notification as sent
				err = s.notificationRepo.MarkAsSent(ctx, notif.ID)
				if err != nil {
					s.log.Error("Failed to mark notification as sent", err, "notificationID", notif.ID)
					continue
				}

				s.log.Info("Notification sent and marked as sent", "notificationID", notif.ID, "channelType", channelType)
			}
		}
	}

	return nil
}

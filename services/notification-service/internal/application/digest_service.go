package application

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/notification-service/internal/ports"
)

type DigestService struct {
	notificationRepo ports.NotificationRepository
	preferenceRepo   ports.PreferenceRepository
	log              logger.Logger
}

func NewDigestService(
	notificationRepo ports.NotificationRepository,
	preferenceRepo ports.PreferenceRepository,
	log logger.Logger,
) *DigestService {
	return &DigestService{
		notificationRepo: notificationRepo,
		preferenceRepo:   preferenceRepo,
		log:              log,
	}
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

	// In a production system, this would query all notifications with:
	// - status = "queued" or "scheduled"
	// - scheduled_for <= now
	// Then process them by:
	// 1. Calling the appropriate channel drivers
	// 2. Updating status to "sent"
	// 3. Recording sent timestamp
	//
	// This would be integrated with a background job processor or cron service.
	//
	// For now, this is a placeholder for the scheduled processing logic.
	// The actual implementation would depend on:
	// - A way to retrieve queued/scheduled notifications from the repository
	// - Integration with channel drivers to send them
	// - Update logic to mark notifications as sent

	return nil
}

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
	return nil
}

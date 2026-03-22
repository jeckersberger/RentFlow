package application

import (
	"context"

	"github.com/google/uuid"
	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/notification-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/notification-service/internal/ports"
)

type PreferenceService struct {
	preferenceRepo ports.PreferenceRepository
	log            logger.Logger
}

func NewPreferenceService(preferenceRepo ports.PreferenceRepository, log logger.Logger) *PreferenceService {
	return &PreferenceService{
		preferenceRepo: preferenceRepo,
		log:            log,
	}
}

func (s *PreferenceService) UpdatePreference(ctx context.Context, cmd domain.UpdatePreferenceCmd) (*domain.UserPreference, error) {
	existing, _ := s.preferenceRepo.GetByUserAndEvent(ctx, cmd.TenantID, cmd.UserID, cmd.EventType)

	preference := &domain.UserPreference{
		TenantID:        cmd.TenantID,
		UserID:          cmd.UserID,
		EventType:       cmd.EventType,
		Channels:        cmd.Channels,
		IsEnabled:       cmd.IsEnabled,
		QuietHoursStart: cmd.QuietHoursStart,
		QuietHoursEnd:   cmd.QuietHoursEnd,
		DigestMode:      cmd.DigestMode,
	}

	if existing != nil {
		preference.ID = existing.ID
		return s.preferenceRepo.Update(ctx, preference)
	}

	preference.ID = uuid.New()
	return s.preferenceRepo.Create(ctx, preference)
}

func (s *PreferenceService) GetPreferences(ctx context.Context, tenantID, userID uuid.UUID) ([]*domain.UserPreference, error) {
	return s.preferenceRepo.ListByUser(ctx, tenantID, userID)
}

func (s *PreferenceService) GetPreference(ctx context.Context, tenantID, userID uuid.UUID, eventType string) (*domain.UserPreference, error) {
	return s.preferenceRepo.GetByUserAndEvent(ctx, tenantID, userID, eventType)
}

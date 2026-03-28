package application

import (
	"context"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/services/notification/internal/domain"
)

// ---------------------------------------------------------------------------
// Request DTOs
// ---------------------------------------------------------------------------

type UpdatePreferenceRequest struct {
	Channel string `json:"channel"`
	Type    string `json:"type"`
	Enabled bool   `json:"enabled"`
}

// ---------------------------------------------------------------------------
// Service
// ---------------------------------------------------------------------------

type PreferenceService struct {
	repo   domain.NotificationPreferenceRepository
	logger zerolog.Logger
}

func NewPreferenceService(repo domain.NotificationPreferenceRepository, logger zerolog.Logger) *PreferenceService {
	return &PreferenceService{
		repo:   repo,
		logger: logger.With().Str("service", "preference").Logger(),
	}
}

func (s *PreferenceService) List(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID) ([]*domain.NotificationPreference, error) {
	return s.repo.List(ctx, tenantID, userID)
}

func (s *PreferenceService) Update(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID, req UpdatePreferenceRequest) (*domain.NotificationPreference, error) {
	channel := req.Channel
	if channel == "" {
		channel = "in_app"
	}

	pref := &domain.NotificationPreference{
		ID:       uuid.New(),
		TenantID: tenantID,
		UserID:   userID,
		Channel:  channel,
		Type:     req.Type,
		Enabled:  req.Enabled,
	}

	if err := s.repo.Upsert(ctx, pref); err != nil {
		return nil, err
	}

	s.logger.Info().
		Str("channel", pref.Channel).
		Str("type", pref.Type).
		Bool("enabled", pref.Enabled).
		Msg("preference updated")
	return pref, nil
}

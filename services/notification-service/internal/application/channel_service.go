package application

import (
	"context"

	"github.com/google/uuid"
	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/notification-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/notification-service/internal/ports"
)

type ChannelService struct {
	channelRepo ports.ChannelRepository
	log         logger.Logger
}

func NewChannelService(channelRepo ports.ChannelRepository, log logger.Logger) *ChannelService {
	return &ChannelService{
		channelRepo: channelRepo,
		log:         log,
	}
}

func (s *ChannelService) RegisterChannel(ctx context.Context, cmd domain.RegisterChannelCmd) (*domain.NotificationChannel, error) {
	channel := &domain.NotificationChannel{
		ID:       uuid.New(),
		TenantID: cmd.TenantID,
		UserID:   cmd.UserID,
		Type:     cmd.Type,
		Config:   cmd.Config,
		IsActive: true,
	}

	created, err := s.channelRepo.Create(ctx, channel)
	if err != nil {
		s.log.Error("Failed to register channel", err)
		return nil, err
	}

	s.log.Info("Channel registered", "channelID", created.ID, "type", cmd.Type)
	return created, nil
}

func (s *ChannelService) GetChannels(ctx context.Context, tenantID, userID uuid.UUID) ([]*domain.NotificationChannel, error) {
	return s.channelRepo.ListByUser(ctx, tenantID, userID)
}

func (s *ChannelService) DeleteChannel(ctx context.Context, channelID uuid.UUID) error {
	return s.channelRepo.Delete(ctx, channelID)
}

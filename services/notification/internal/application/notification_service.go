package application

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/services/notification/internal/domain"
)

// ---------------------------------------------------------------------------
// Request DTOs
// ---------------------------------------------------------------------------

type CreateNotificationRequest struct {
	UserID        uuid.UUID  `json:"user_id"`
	Type          string     `json:"type"`
	Title         string     `json:"title"`
	Message       string     `json:"message"`
	ReferenceID   *uuid.UUID `json:"reference_id"`
	ReferenceType string     `json:"reference_type"`
}

// ---------------------------------------------------------------------------
// Service
// ---------------------------------------------------------------------------

type NotificationService struct {
	repo   domain.NotificationRepository
	logger zerolog.Logger
}

func NewNotificationService(repo domain.NotificationRepository, logger zerolog.Logger) *NotificationService {
	return &NotificationService{
		repo:   repo,
		logger: logger.With().Str("service", "notification").Logger(),
	}
}

func (s *NotificationService) Create(ctx context.Context, tenantID uuid.UUID, req CreateNotificationRequest) (*domain.Notification, error) {
	if req.Title == "" {
		return nil, fmt.Errorf("title is required")
	}
	if req.Type == "" {
		return nil, fmt.Errorf("type is required")
	}
	if req.UserID == uuid.Nil {
		return nil, fmt.Errorf("user_id is required")
	}

	n := &domain.Notification{
		ID:            uuid.New(),
		TenantID:      tenantID,
		UserID:        req.UserID,
		Type:          req.Type,
		Title:         req.Title,
		Message:       req.Message,
		ReferenceID:   req.ReferenceID,
		ReferenceType: req.ReferenceType,
	}

	if err := s.repo.Create(ctx, n); err != nil {
		return nil, fmt.Errorf("create notification: %w", err)
	}

	s.logger.Info().Str("notification_id", n.ID.String()).Msg("notification created")
	return n, nil
}

func (s *NotificationService) List(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID, filter domain.NotificationFilter) ([]*domain.Notification, int64, error) {
	return s.repo.List(ctx, tenantID, userID, filter)
}

func (s *NotificationService) UnreadCount(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID) (int64, error) {
	return s.repo.UnreadCount(ctx, tenantID, userID)
}

func (s *NotificationService) MarkAsRead(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error {
	return s.repo.MarkAsRead(ctx, id, tenantID)
}

func (s *NotificationService) MarkAllRead(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID) (int64, error) {
	return s.repo.MarkAllRead(ctx, tenantID, userID)
}

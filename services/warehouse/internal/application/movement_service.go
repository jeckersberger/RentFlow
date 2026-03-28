package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/services/warehouse/internal/domain"
)

// ---------------------------------------------------------------------------
// Request DTOs
// ---------------------------------------------------------------------------

// CreateMovementRequest holds the data needed to book a stock movement.
type CreateMovementRequest struct {
	EquipmentID    uuid.UUID  `json:"equipment_id"`
	FromLocationID *uuid.UUID `json:"from_location_id,omitempty"`
	ToLocationID   *uuid.UUID `json:"to_location_id,omitempty"`
	Quantity       int        `json:"quantity"`
	Reason         string     `json:"reason,omitempty"`
	Notes          string     `json:"notes,omitempty"`
}

// ---------------------------------------------------------------------------
// Service
// ---------------------------------------------------------------------------

// MovementService implements the application-level use cases for movements.
type MovementService struct {
	movementRepo domain.MovementRepository
	logger       zerolog.Logger
}

// NewMovementService constructs a new MovementService.
func NewMovementService(
	movementRepo domain.MovementRepository,
	logger zerolog.Logger,
) *MovementService {
	return &MovementService{
		movementRepo: movementRepo,
		logger:       logger.With().Str("service", "movement").Logger(),
	}
}

// Create validates the request and persists a new movement.
func (s *MovementService) Create(
	ctx context.Context,
	tenantID uuid.UUID,
	userID uuid.UUID,
	req CreateMovementRequest,
) (*domain.Movement, error) {
	if req.EquipmentID == uuid.Nil {
		return nil, fmt.Errorf("equipment_id is required")
	}

	quantity := req.Quantity
	if quantity < 1 {
		quantity = 1
	}

	movement := &domain.Movement{
		ID:             uuid.New(),
		TenantID:       tenantID,
		EquipmentID:    req.EquipmentID,
		FromLocationID: req.FromLocationID,
		ToLocationID:   req.ToLocationID,
		Quantity:       quantity,
		Reason:         req.Reason,
		UserID:         &userID,
		Notes:          req.Notes,
		CreatedAt:      time.Now(),
	}

	if err := s.movementRepo.Create(ctx, movement); err != nil {
		s.logger.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Msg("failed to create movement")
		return nil, fmt.Errorf("create movement: %w", err)
	}

	s.logger.Info().
		Str("movement_id", movement.ID.String()).
		Str("tenant_id", tenantID.String()).
		Str("equipment_id", req.EquipmentID.String()).
		Msg("movement created")

	return movement, nil
}

// List returns a paginated list of movements for a tenant.
func (s *MovementService) List(
	ctx context.Context,
	tenantID uuid.UUID,
	page int,
	perPage int,
) ([]*domain.Movement, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}

	items, total, err := s.movementRepo.List(ctx, tenantID, page, perPage)
	if err != nil {
		s.logger.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Msg("failed to list movements")
		return nil, 0, fmt.Errorf("list movements: %w", err)
	}
	return items, total, nil
}

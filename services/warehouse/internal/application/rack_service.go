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

// CreateRackRequest holds the data needed to create a new rack.
type CreateRackRequest struct {
	ZoneID       uuid.UUID `json:"zone_id"`
	Name         string    `json:"name"`
	Code         string    `json:"code,omitempty"`
	Levels       int       `json:"levels"`
	BaysPerLevel int       `json:"bays_per_level"`
	MaxWeightKg  *int      `json:"max_weight_kg,omitempty"`
}

// ---------------------------------------------------------------------------
// Service
// ---------------------------------------------------------------------------

// RackService implements the application-level use cases for racks.
type RackService struct {
	rackRepo domain.RackRepository
	logger   zerolog.Logger
}

// NewRackService constructs a new RackService.
func NewRackService(
	rackRepo domain.RackRepository,
	logger zerolog.Logger,
) *RackService {
	return &RackService{
		rackRepo: rackRepo,
		logger:   logger.With().Str("service", "rack").Logger(),
	}
}

// Create validates the request and persists a new rack.
func (s *RackService) Create(
	ctx context.Context,
	tenantID uuid.UUID,
	req CreateRackRequest,
) (*domain.Rack, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("rack name is required")
	}
	if req.ZoneID == uuid.Nil {
		return nil, fmt.Errorf("zone_id is required")
	}

	levels := req.Levels
	if levels < 1 {
		levels = 4
	}
	baysPerLevel := req.BaysPerLevel
	if baysPerLevel < 1 {
		baysPerLevel = 6
	}

	rack := &domain.Rack{
		ID:           uuid.New(),
		ZoneID:       req.ZoneID,
		TenantID:     tenantID,
		Name:         req.Name,
		Code:         req.Code,
		Levels:       levels,
		BaysPerLevel: baysPerLevel,
		MaxWeightKg:  req.MaxWeightKg,
		CreatedAt:    time.Now(),
	}

	if err := s.rackRepo.Create(ctx, rack); err != nil {
		s.logger.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Msg("failed to create rack")
		return nil, fmt.Errorf("create rack: %w", err)
	}

	s.logger.Info().
		Str("rack_id", rack.ID.String()).
		Str("tenant_id", tenantID.String()).
		Msg("rack created")

	return rack, nil
}

// List returns all racks for a tenant.
func (s *RackService) List(
	ctx context.Context,
	tenantID uuid.UUID,
) ([]*domain.Rack, error) {
	racks, err := s.rackRepo.List(ctx, tenantID)
	if err != nil {
		s.logger.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Msg("failed to list racks")
		return nil, fmt.Errorf("list racks: %w", err)
	}
	return racks, nil
}

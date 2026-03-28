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

// CreateZoneRequest holds the data needed to create a new zone.
type CreateZoneRequest struct {
	WarehouseID       uuid.UUID `json:"warehouse_id"`
	Name              string    `json:"name"`
	Code              string    `json:"code,omitempty"`
	ClimateControlled bool      `json:"climate_controlled"`
	MaxWeightKg       *int      `json:"max_weight_kg,omitempty"`
	Notes             string    `json:"notes,omitempty"`
}

// UpdateZoneRequest holds optional fields for patching a zone.
type UpdateZoneRequest struct {
	Name              *string `json:"name,omitempty"`
	Code              *string `json:"code,omitempty"`
	ClimateControlled *bool   `json:"climate_controlled,omitempty"`
	MaxWeightKg       *int    `json:"max_weight_kg,omitempty"`
	Notes             *string `json:"notes,omitempty"`
}

// ---------------------------------------------------------------------------
// Service
// ---------------------------------------------------------------------------

// ZoneService implements the application-level use cases for zones.
type ZoneService struct {
	zoneRepo domain.ZoneRepository
	logger   zerolog.Logger
}

// NewZoneService constructs a new ZoneService.
func NewZoneService(
	zoneRepo domain.ZoneRepository,
	logger zerolog.Logger,
) *ZoneService {
	return &ZoneService{
		zoneRepo: zoneRepo,
		logger:   logger.With().Str("service", "zone").Logger(),
	}
}

// Create validates the request and persists a new zone.
func (s *ZoneService) Create(
	ctx context.Context,
	tenantID uuid.UUID,
	req CreateZoneRequest,
) (*domain.Zone, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("zone name is required")
	}
	if req.WarehouseID == uuid.Nil {
		return nil, fmt.Errorf("warehouse_id is required")
	}

	zone := &domain.Zone{
		ID:                uuid.New(),
		WarehouseID:       req.WarehouseID,
		TenantID:          tenantID,
		Name:              req.Name,
		Code:              req.Code,
		ClimateControlled: req.ClimateControlled,
		MaxWeightKg:       req.MaxWeightKg,
		Notes:             req.Notes,
		CreatedAt:         time.Now(),
	}

	if err := s.zoneRepo.Create(ctx, zone); err != nil {
		s.logger.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Msg("failed to create zone")
		return nil, fmt.Errorf("create zone: %w", err)
	}

	s.logger.Info().
		Str("zone_id", zone.ID.String()).
		Str("tenant_id", tenantID.String()).
		Msg("zone created")

	return zone, nil
}

// List returns all zones for a tenant.
func (s *ZoneService) List(
	ctx context.Context,
	tenantID uuid.UUID,
) ([]*domain.Zone, error) {
	zones, err := s.zoneRepo.List(ctx, tenantID)
	if err != nil {
		s.logger.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Msg("failed to list zones")
		return nil, fmt.Errorf("list zones: %w", err)
	}
	return zones, nil
}

// Update patches an existing zone with non-nil fields from the request.
func (s *ZoneService) Update(
	ctx context.Context,
	id uuid.UUID,
	tenantID uuid.UUID,
	req UpdateZoneRequest,
) (*domain.Zone, error) {
	existing, err := s.zoneRepo.GetByID(ctx, id, tenantID)
	if err != nil {
		s.logger.Error().Err(err).
			Str("zone_id", id.String()).
			Str("tenant_id", tenantID.String()).
			Msg("failed to fetch zone for update")
		return nil, fmt.Errorf("update zone – fetch: %w", err)
	}

	if req.Name != nil {
		if *req.Name == "" {
			return nil, fmt.Errorf("zone name must not be empty")
		}
		existing.Name = *req.Name
	}
	if req.Code != nil {
		existing.Code = *req.Code
	}
	if req.ClimateControlled != nil {
		existing.ClimateControlled = *req.ClimateControlled
	}
	if req.MaxWeightKg != nil {
		existing.MaxWeightKg = req.MaxWeightKg
	}
	if req.Notes != nil {
		existing.Notes = *req.Notes
	}

	if err := s.zoneRepo.Update(ctx, existing); err != nil {
		s.logger.Error().Err(err).
			Str("zone_id", id.String()).
			Str("tenant_id", tenantID.String()).
			Msg("failed to update zone")
		return nil, fmt.Errorf("update zone: %w", err)
	}

	s.logger.Info().
		Str("zone_id", id.String()).
		Str("tenant_id", tenantID.String()).
		Msg("zone updated")

	return existing, nil
}

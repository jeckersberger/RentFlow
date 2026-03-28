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

// CreateLocationRequest holds the data needed to create a new stock location.
type CreateLocationRequest struct {
	RackID      *uuid.UUID `json:"rack_id,omitempty"`
	ZoneID      *uuid.UUID `json:"zone_id,omitempty"`
	Code        string     `json:"code"`
	Barcode     string     `json:"barcode,omitempty"`
	Level       *int       `json:"level,omitempty"`
	Bay         *int       `json:"bay,omitempty"`
	MaxWeightKg *int       `json:"max_weight_kg,omitempty"`
}

// ---------------------------------------------------------------------------
// Service
// ---------------------------------------------------------------------------

// LocationService implements the application-level use cases for stock locations.
type LocationService struct {
	locationRepo domain.StockLocationRepository
	logger       zerolog.Logger
}

// NewLocationService constructs a new LocationService.
func NewLocationService(
	locationRepo domain.StockLocationRepository,
	logger zerolog.Logger,
) *LocationService {
	return &LocationService{
		locationRepo: locationRepo,
		logger:       logger.With().Str("service", "location").Logger(),
	}
}

// Create validates the request and persists a new stock location.
func (s *LocationService) Create(
	ctx context.Context,
	tenantID uuid.UUID,
	req CreateLocationRequest,
) (*domain.StockLocation, error) {
	if req.Code == "" {
		return nil, fmt.Errorf("location code is required")
	}

	location := &domain.StockLocation{
		ID:          uuid.New(),
		RackID:      req.RackID,
		ZoneID:      req.ZoneID,
		TenantID:    tenantID,
		Code:        req.Code,
		Barcode:     req.Barcode,
		Level:       req.Level,
		Bay:         req.Bay,
		MaxWeightKg: req.MaxWeightKg,
		IsActive:    true,
		CreatedAt:   time.Now(),
	}

	if err := s.locationRepo.Create(ctx, location); err != nil {
		s.logger.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Msg("failed to create stock location")
		return nil, fmt.Errorf("create stock location: %w", err)
	}

	s.logger.Info().
		Str("location_id", location.ID.String()).
		Str("tenant_id", tenantID.String()).
		Msg("stock location created")

	return location, nil
}

// List returns a paginated list of stock locations for a tenant.
func (s *LocationService) List(
	ctx context.Context,
	tenantID uuid.UUID,
	page int,
	perPage int,
) ([]*domain.StockLocation, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}

	items, total, err := s.locationRepo.List(ctx, tenantID, page, perPage)
	if err != nil {
		s.logger.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Msg("failed to list stock locations")
		return nil, 0, fmt.Errorf("list stock locations: %w", err)
	}
	return items, total, nil
}

// GetByID retrieves a single stock location by ID within a tenant scope.
func (s *LocationService) GetByID(
	ctx context.Context,
	id uuid.UUID,
	tenantID uuid.UUID,
) (*domain.StockLocation, error) {
	location, err := s.locationRepo.GetByID(ctx, id, tenantID)
	if err != nil {
		s.logger.Error().Err(err).
			Str("location_id", id.String()).
			Str("tenant_id", tenantID.String()).
			Msg("failed to get stock location by id")
		return nil, fmt.Errorf("get stock location by id: %w", err)
	}
	return location, nil
}

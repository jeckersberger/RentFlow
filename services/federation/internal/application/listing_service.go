package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/services/federation/internal/domain"
)

// ---------------------------------------------------------------------------
// Request DTOs
// ---------------------------------------------------------------------------

// CreateListingRequest holds the data needed to create a shared listing.
type CreateListingRequest struct {
	EquipmentID    uuid.UUID  `json:"equipment_id"`
	DailyRate      int64      `json:"daily_rate"`
	WeeklyRate     int64      `json:"weekly_rate"`
	AvailableFrom  *time.Time `json:"available_from,omitempty"`
	AvailableUntil *time.Time `json:"available_until,omitempty"`
	IsActive       *bool      `json:"is_active,omitempty"`
	Notes          string     `json:"notes,omitempty"`
}

// UpdateListingRequest holds the data needed to update a shared listing.
type UpdateListingRequest struct {
	EquipmentID    uuid.UUID  `json:"equipment_id"`
	DailyRate      int64      `json:"daily_rate"`
	WeeklyRate     int64      `json:"weekly_rate"`
	AvailableFrom  *time.Time `json:"available_from,omitempty"`
	AvailableUntil *time.Time `json:"available_until,omitempty"`
	IsActive       *bool      `json:"is_active,omitempty"`
	Notes          string     `json:"notes,omitempty"`
}

// ---------------------------------------------------------------------------
// Service
// ---------------------------------------------------------------------------

// ListingService implements the application-level use cases for shared listings.
type ListingService struct {
	listingRepo domain.SharedListingRepository
	logger      zerolog.Logger
}

// NewListingService constructs a new ListingService.
func NewListingService(
	listingRepo domain.SharedListingRepository,
	logger zerolog.Logger,
) *ListingService {
	return &ListingService{
		listingRepo: listingRepo,
		logger:      logger.With().Str("service", "listing").Logger(),
	}
}

// Create creates a new shared listing.
func (s *ListingService) Create(
	ctx context.Context,
	tenantID uuid.UUID,
	req CreateListingRequest,
) (*domain.SharedListing, error) {
	if req.EquipmentID == uuid.Nil {
		return nil, domain.ErrMissingEquipmentID
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	listing := &domain.SharedListing{
		ID:             uuid.New(),
		TenantID:       tenantID,
		EquipmentID:    req.EquipmentID,
		DailyRate:      req.DailyRate,
		WeeklyRate:     req.WeeklyRate,
		AvailableFrom:  req.AvailableFrom,
		AvailableUntil: req.AvailableUntil,
		IsActive:       isActive,
		Notes:          req.Notes,
	}

	if err := s.listingRepo.Create(ctx, listing); err != nil {
		s.logger.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Msg("failed to create shared listing")
		return nil, fmt.Errorf("create listing: %w", err)
	}

	s.logger.Info().
		Str("listing_id", listing.ID.String()).
		Str("tenant_id", tenantID.String()).
		Msg("shared listing created")

	return listing, nil
}

// GetByID retrieves a shared listing by ID.
func (s *ListingService) GetByID(
	ctx context.Context,
	id uuid.UUID,
	tenantID uuid.UUID,
) (*domain.SharedListing, error) {
	listing, err := s.listingRepo.GetByID(ctx, id, tenantID)
	if err != nil {
		s.logger.Error().Err(err).
			Str("listing_id", id.String()).
			Str("tenant_id", tenantID.String()).
			Msg("failed to get shared listing")
		return nil, fmt.Errorf("get listing: %w", err)
	}
	return listing, nil
}

// List returns a filtered, paginated list of shared listings.
func (s *ListingService) List(
	ctx context.Context,
	tenantID uuid.UUID,
	filter domain.ListingFilter,
) ([]*domain.SharedListing, int64, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PerPage < 1 {
		filter.PerPage = 20
	}

	items, total, err := s.listingRepo.List(ctx, tenantID, filter)
	if err != nil {
		s.logger.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Msg("failed to list shared listings")
		return nil, 0, fmt.Errorf("list listings: %w", err)
	}
	return items, total, nil
}

// Update updates an existing shared listing.
func (s *ListingService) Update(
	ctx context.Context,
	id uuid.UUID,
	tenantID uuid.UUID,
	req UpdateListingRequest,
) (*domain.SharedListing, error) {
	existing, err := s.listingRepo.GetByID(ctx, id, tenantID)
	if err != nil {
		return nil, fmt.Errorf("update listing: %w", err)
	}

	if req.EquipmentID != uuid.Nil {
		existing.EquipmentID = req.EquipmentID
	}
	existing.DailyRate = req.DailyRate
	existing.WeeklyRate = req.WeeklyRate
	existing.AvailableFrom = req.AvailableFrom
	existing.AvailableUntil = req.AvailableUntil
	if req.IsActive != nil {
		existing.IsActive = *req.IsActive
	}
	existing.Notes = req.Notes

	if err := s.listingRepo.Update(ctx, existing); err != nil {
		s.logger.Error().Err(err).
			Str("listing_id", id.String()).
			Str("tenant_id", tenantID.String()).
			Msg("failed to update shared listing")
		return nil, fmt.Errorf("update listing: %w", err)
	}

	s.logger.Info().
		Str("listing_id", id.String()).
		Str("tenant_id", tenantID.String()).
		Msg("shared listing updated")

	return existing, nil
}

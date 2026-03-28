package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/services/inventory/internal/domain"
)

// ---------------------------------------------------------------------------
// Request DTOs
// ---------------------------------------------------------------------------

// CreateFlightcaseRequest holds the data needed to create a new flightcase.
type CreateFlightcaseRequest struct {
	Name        string `json:"name"`
	Barcode     string `json:"barcode,omitempty"`
	Description string `json:"description,omitempty"`
	WeightGrams *int   `json:"weight_grams,omitempty"`
}

// UpdateFlightcaseRequest holds optional fields for patching a flightcase.
type UpdateFlightcaseRequest struct {
	Name        *string `json:"name,omitempty"`
	Barcode     *string `json:"barcode,omitempty"`
	Description *string `json:"description,omitempty"`
	WeightGrams *int    `json:"weight_grams,omitempty"`
}

// AddFlightcaseItemRequest holds the data needed to add an equipment item to a
// flightcase.
type AddFlightcaseItemRequest struct {
	EquipmentID uuid.UUID `json:"equipment_id"`
	Quantity    int       `json:"quantity"`
	SortOrder   int       `json:"sort_order"`
}

// ---------------------------------------------------------------------------
// Service
// ---------------------------------------------------------------------------

// FlightcaseService implements the application-level use cases for flightcases.
type FlightcaseService struct {
	flightcaseRepo domain.FlightcaseRepository
	logger         zerolog.Logger
}

// NewFlightcaseService constructs a new FlightcaseService.
func NewFlightcaseService(
	flightcaseRepo domain.FlightcaseRepository,
	logger zerolog.Logger,
) *FlightcaseService {
	return &FlightcaseService{
		flightcaseRepo: flightcaseRepo,
		logger:         logger.With().Str("service", "flightcase").Logger(),
	}
}

// Create validates the request and persists a new flightcase.
func (s *FlightcaseService) Create(
	ctx context.Context,
	tenantID uuid.UUID,
	req CreateFlightcaseRequest,
) (*domain.Flightcase, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("flightcase name is required")
	}

	now := time.Now()
	flightcase := &domain.Flightcase{
		ID:          uuid.New(),
		TenantID:    tenantID,
		Name:        req.Name,
		Barcode:     req.Barcode,
		Description: req.Description,
		WeightGrams: req.WeightGrams,
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.flightcaseRepo.Create(ctx, flightcase); err != nil {
		s.logger.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Msg("failed to create flightcase")
		return nil, fmt.Errorf("create flightcase: %w", err)
	}

	s.logger.Info().
		Str("flightcase_id", flightcase.ID.String()).
		Str("tenant_id", tenantID.String()).
		Msg("flightcase created")

	return flightcase, nil
}

// GetByID retrieves a single flightcase by ID within a tenant scope.
func (s *FlightcaseService) GetByID(
	ctx context.Context,
	id uuid.UUID,
	tenantID uuid.UUID,
) (*domain.Flightcase, error) {
	fc, err := s.flightcaseRepo.GetByID(ctx, id, tenantID)
	if err != nil {
		s.logger.Error().Err(err).
			Str("flightcase_id", id.String()).
			Str("tenant_id", tenantID.String()).
			Msg("failed to get flightcase by id")
		return nil, fmt.Errorf("get flightcase by id: %w", err)
	}
	return fc, nil
}

// List returns a paginated list of flightcases for a tenant.
func (s *FlightcaseService) List(
	ctx context.Context,
	tenantID uuid.UUID,
	page int,
	perPage int,
) ([]*domain.Flightcase, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}

	items, total, err := s.flightcaseRepo.List(ctx, tenantID, page, perPage)
	if err != nil {
		s.logger.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Msg("failed to list flightcases")
		return nil, 0, fmt.Errorf("list flightcases: %w", err)
	}
	return items, total, nil
}

// Update patches an existing flightcase with non-nil fields from the request.
func (s *FlightcaseService) Update(
	ctx context.Context,
	id uuid.UUID,
	tenantID uuid.UUID,
	req UpdateFlightcaseRequest,
) (*domain.Flightcase, error) {
	existing, err := s.flightcaseRepo.GetByID(ctx, id, tenantID)
	if err != nil {
		s.logger.Error().Err(err).
			Str("flightcase_id", id.String()).
			Str("tenant_id", tenantID.String()).
			Msg("failed to fetch flightcase for update")
		return nil, fmt.Errorf("update flightcase – fetch: %w", err)
	}

	if req.Name != nil {
		if *req.Name == "" {
			return nil, fmt.Errorf("flightcase name must not be empty")
		}
		existing.Name = *req.Name
	}
	if req.Barcode != nil {
		existing.Barcode = *req.Barcode
	}
	if req.Description != nil {
		existing.Description = *req.Description
	}
	if req.WeightGrams != nil {
		existing.WeightGrams = req.WeightGrams
	}

	existing.UpdatedAt = time.Now()

	if err := s.flightcaseRepo.Update(ctx, existing); err != nil {
		s.logger.Error().Err(err).
			Str("flightcase_id", id.String()).
			Str("tenant_id", tenantID.String()).
			Msg("failed to update flightcase")
		return nil, fmt.Errorf("update flightcase: %w", err)
	}

	s.logger.Info().
		Str("flightcase_id", id.String()).
		Str("tenant_id", tenantID.String()).
		Msg("flightcase updated")

	return existing, nil
}

// Delete removes a flightcase by ID.
func (s *FlightcaseService) Delete(
	ctx context.Context,
	id uuid.UUID,
	tenantID uuid.UUID,
) error {
	if err := s.flightcaseRepo.Delete(ctx, id, tenantID); err != nil {
		s.logger.Error().Err(err).
			Str("flightcase_id", id.String()).
			Str("tenant_id", tenantID.String()).
			Msg("failed to delete flightcase")
		return fmt.Errorf("delete flightcase: %w", err)
	}

	s.logger.Info().
		Str("flightcase_id", id.String()).
		Str("tenant_id", tenantID.String()).
		Msg("flightcase deleted")

	return nil
}

// AddItem adds an equipment item to a flightcase.
func (s *FlightcaseService) AddItem(
	ctx context.Context,
	tenantID uuid.UUID,
	flightcaseID uuid.UUID,
	req AddFlightcaseItemRequest,
) (*domain.FlightcaseItem, error) {
	if req.Quantity < 1 {
		return nil, fmt.Errorf("quantity must be at least 1")
	}

	// Verify the flightcase exists and belongs to the tenant.
	if _, err := s.flightcaseRepo.GetByID(ctx, flightcaseID, tenantID); err != nil {
		s.logger.Error().Err(err).
			Str("flightcase_id", flightcaseID.String()).
			Str("tenant_id", tenantID.String()).
			Msg("failed to verify flightcase for add item")
		return nil, fmt.Errorf("add item – verify flightcase: %w", err)
	}

	item := &domain.FlightcaseItem{
		ID:           uuid.New(),
		FlightcaseID: flightcaseID,
		EquipmentID:  req.EquipmentID,
		Quantity:     req.Quantity,
		SortOrder:    req.SortOrder,
		CreatedAt:    time.Now(),
	}

	if err := s.flightcaseRepo.AddItem(ctx, item); err != nil {
		s.logger.Error().Err(err).
			Str("flightcase_id", flightcaseID.String()).
			Str("equipment_id", req.EquipmentID.String()).
			Msg("failed to add item to flightcase")
		return nil, fmt.Errorf("add item to flightcase: %w", err)
	}

	s.logger.Info().
		Str("flightcase_id", flightcaseID.String()).
		Str("equipment_id", req.EquipmentID.String()).
		Int("quantity", req.Quantity).
		Msg("item added to flightcase")

	return item, nil
}

// RemoveItem removes an equipment item from a flightcase by item ID.
func (s *FlightcaseService) RemoveItem(
	ctx context.Context,
	itemID uuid.UUID,
) error {
	if err := s.flightcaseRepo.RemoveItem(ctx, itemID); err != nil {
		s.logger.Error().Err(err).
			Str("item_id", itemID.String()).
			Msg("failed to remove item from flightcase")
		return fmt.Errorf("remove item from flightcase: %w", err)
	}

	s.logger.Info().
		Str("item_id", itemID.String()).
		Msg("item removed from flightcase")

	return nil
}

// GetItems returns all equipment items in a flightcase.
func (s *FlightcaseService) GetItems(
	ctx context.Context,
	flightcaseID uuid.UUID,
) ([]*domain.FlightcaseItem, error) {
	items, err := s.flightcaseRepo.GetItems(ctx, flightcaseID)
	if err != nil {
		s.logger.Error().Err(err).
			Str("flightcase_id", flightcaseID.String()).
			Msg("failed to get flightcase items")
		return nil, fmt.Errorf("get flightcase items: %w", err)
	}
	return items, nil
}

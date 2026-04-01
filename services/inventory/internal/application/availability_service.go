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
// Request / Response DTOs
// ---------------------------------------------------------------------------

// AvailabilityRequest describes a single availability check for an equipment item.
type AvailabilityRequest struct {
	EquipmentID uuid.UUID `json:"equipment_id"`
	StartDate   string    `json:"start_date"`
	EndDate     string    `json:"end_date"`
	Quantity    int       `json:"quantity"`
}

// BatchAvailabilityRequest wraps multiple availability checks.
type BatchAvailabilityRequest struct {
	Items []AvailabilityRequest `json:"items"`
}

// AvailabilityResult contains the availability information for one equipment item.
type AvailabilityResult struct {
	EquipmentID   uuid.UUID `json:"equipment_id"`
	EquipmentName string    `json:"equipment_name"`
	TotalQty      int       `json:"total_qty"`
	Available     int       `json:"available"`
	Reserved      int       `json:"reserved"`
	InMaintenance int       `json:"in_maintenance"`
	Requested     int       `json:"requested"`
	Feasible      bool      `json:"feasible"`
}

// ---------------------------------------------------------------------------
// Service
// ---------------------------------------------------------------------------

// AvailabilityService handles equipment availability checks.
type AvailabilityService struct {
	equipmentRepo    domain.EquipmentRepository
	availabilityRepo domain.AvailabilityRepository
	logger           zerolog.Logger
}

// NewAvailabilityService creates a new AvailabilityService.
func NewAvailabilityService(
	equipmentRepo domain.EquipmentRepository,
	availabilityRepo domain.AvailabilityRepository,
	logger zerolog.Logger,
) *AvailabilityService {
	return &AvailabilityService{
		equipmentRepo:    equipmentRepo,
		availabilityRepo: availabilityRepo,
		logger:           logger.With().Str("service", "availability").Logger(),
	}
}

// CheckSingle checks availability for a single equipment item.
// For now this uses the static quantity_available from the equipment table.
// Date-range aware checks will be added in a future iteration.
func (s *AvailabilityService) CheckSingle(
	ctx context.Context,
	tenantID uuid.UUID,
	req AvailabilityRequest,
) (*AvailabilityResult, error) {
	if req.Quantity < 1 {
		req.Quantity = 1
	}

	equipment, err := s.equipmentRepo.GetByID(ctx, req.EquipmentID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("availability check – fetch equipment: %w", err)
	}

	reserved := equipment.QuantityTotal - equipment.QuantityAvailable
	inMaintenance := 0
	if equipment.Status == domain.StatusInMaintenance {
		inMaintenance = equipment.QuantityTotal
	}

	result := &AvailabilityResult{
		EquipmentID:   equipment.ID,
		EquipmentName: equipment.Name,
		TotalQty:      equipment.QuantityTotal,
		Available:     equipment.QuantityAvailable,
		Reserved:      reserved,
		InMaintenance: inMaintenance,
		Requested:     req.Quantity,
		Feasible:      equipment.QuantityAvailable >= req.Quantity,
	}

	s.logger.Debug().
		Str("equipment_id", req.EquipmentID.String()).
		Int("requested", req.Quantity).
		Int("available", equipment.QuantityAvailable).
		Bool("feasible", result.Feasible).
		Msg("availability check completed")

	return result, nil
}

// CheckBatch checks availability for multiple equipment items.
func (s *AvailabilityService) CheckBatch(
	ctx context.Context,
	tenantID uuid.UUID,
	req BatchAvailabilityRequest,
) ([]AvailabilityResult, error) {
	if len(req.Items) == 0 {
		return []AvailabilityResult{}, nil
	}

	results := make([]AvailabilityResult, 0, len(req.Items))
	for _, item := range req.Items {
		result, err := s.CheckSingle(ctx, tenantID, item)
		if err != nil {
			return nil, fmt.Errorf("batch availability check for %s: %w", item.EquipmentID, err)
		}
		results = append(results, *result)
	}

	return results, nil
}

// ---------------------------------------------------------------------------
// Type-level availability overview
// ---------------------------------------------------------------------------

// GetTypeAvailability returns an availability summary per equipment type,
// optionally filtered by category and date range.
func (s *AvailabilityService) GetTypeAvailability(
	ctx context.Context,
	tenantID uuid.UUID,
	from time.Time,
	to time.Time,
	categoryID *uuid.UUID,
) ([]domain.TypeAvailabilitySummary, error) {
	filter := domain.AvailabilityFilter{
		From:       from,
		To:         to,
		CategoryID: categoryID,
	}

	results, err := s.availabilityRepo.GetTypeAvailability(ctx, tenantID, filter)
	if err != nil {
		s.logger.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Msg("failed to get type availability")
		return nil, fmt.Errorf("get type availability: %w", err)
	}

	s.logger.Debug().
		Str("tenant_id", tenantID.String()).
		Int("types_count", len(results)).
		Msg("type availability retrieved")

	return results, nil
}

// ---------------------------------------------------------------------------
// Single-item booking timeline
// ---------------------------------------------------------------------------

// GetEquipmentBookings returns the booking/reservation periods for a single
// equipment item within the given date range.
func (s *AvailabilityService) GetEquipmentBookings(
	ctx context.Context,
	tenantID uuid.UUID,
	equipmentID uuid.UUID,
	from time.Time,
	to time.Time,
) ([]domain.EquipmentBooking, error) {
	// Verify equipment exists and belongs to tenant.
	equipment, err := s.equipmentRepo.GetByID(ctx, equipmentID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("get equipment bookings – fetch equipment: %w", err)
	}

	bookings, err := s.availabilityRepo.GetEquipmentBookings(ctx, tenantID, equipmentID, from, to)
	if err != nil {
		s.logger.Error().Err(err).
			Str("equipment_id", equipmentID.String()).
			Str("tenant_id", tenantID.String()).
			Msg("failed to get equipment bookings")
		return nil, fmt.Errorf("get equipment bookings: %w", err)
	}

	s.logger.Debug().
		Str("equipment_id", equipmentID.String()).
		Str("equipment_name", equipment.Name).
		Int("bookings_count", len(bookings)).
		Msg("equipment bookings retrieved")

	return bookings, nil
}

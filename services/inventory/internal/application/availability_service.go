package application

import (
	"context"
	"fmt"

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
	equipmentRepo domain.EquipmentRepository
	logger        zerolog.Logger
}

// NewAvailabilityService creates a new AvailabilityService.
func NewAvailabilityService(
	equipmentRepo domain.EquipmentRepository,
	logger zerolog.Logger,
) *AvailabilityService {
	return &AvailabilityService{
		equipmentRepo: equipmentRepo,
		logger:        logger.With().Str("service", "availability").Logger(),
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

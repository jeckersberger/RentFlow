package application

import (
	"context"
	"fmt"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/inventory-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/inventory-service/internal/ports"
)

type AvailabilityResult struct {
	EquipmentID string `json:"equipment_id"`
	Name        string `json:"name"`
	TotalQty    int    `json:"total_qty"`
	Available   int    `json:"available"`
	Reserved    int    `json:"reserved"`
	Requested   int    `json:"requested"`
	Feasible    bool   `json:"feasible"`
	Message     string `json:"message,omitempty"`
}

type AvailabilityService struct {
	equipRepo ports.EquipmentRepository
	logger    logger.Logger
}

func NewAvailabilityService(
	equipRepo ports.EquipmentRepository,
	logger logger.Logger,
) *AvailabilityService {
	return &AvailabilityService{
		equipRepo: equipRepo,
		logger:    logger,
	}
}

// CheckEquipmentAvailability checks if equipment is available for a given date range.
// For Phase 1, we only check the equipment status (available/not retired/not in maintenance).
// Future phases will implement cross-service communication for actual booking checks.
func (s *AvailabilityService) CheckEquipmentAvailability(
	ctx context.Context,
	tenantID string,
	equipmentID string,
	requestedQty int,
) (*AvailabilityResult, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant ID is required")
	}
	if equipmentID == "" {
		return nil, fmt.Errorf("equipment ID is required")
	}
	if requestedQty <= 0 {
		return nil, fmt.Errorf("requested quantity must be greater than 0")
	}

	eq, err := s.equipRepo.GetByID(ctx, tenantID, equipmentID)
	if err != nil {
		return nil, fmt.Errorf("equipment not found: %w", err)
	}

	// Phase 1: Check status only
	// Assume 1 unit per equipment for now (no inventory quantity tracking yet)
	totalQty := 1
	available := 0
	reserved := 0
	feasible := false
	message := ""

	if eq.Status == domain.StatusAvailable {
		available = totalQty
		feasible = requestedQty <= available
		if !feasible {
			message = fmt.Sprintf("requested quantity (%d) exceeds available units (%d)", requestedQty, available)
		}
	} else if eq.Status == domain.StatusRetired {
		message = "equipment is retired and not available for rental"
	} else if eq.Status == domain.StatusInMaintenance {
		message = "equipment is currently in maintenance"
	} else if eq.Status == domain.StatusDamaged {
		message = "equipment is damaged and not available for rental"
	} else if eq.Status == domain.StatusReserved || eq.Status == domain.StatusCheckedOut {
		reserved = totalQty
		message = "equipment is already reserved or checked out"
	}

	return &AvailabilityResult{
		EquipmentID: equipmentID,
		Name:        eq.Name,
		TotalQty:    totalQty,
		Available:   available,
		Reserved:    reserved,
		Requested:   requestedQty,
		Feasible:    feasible,
		Message:     message,
	}, nil
}

// CheckBatchAvailability checks availability for multiple equipment items
type BatchAvailabilityRequest struct {
	EquipmentID string `json:"equipment_id"`
	Quantity    int    `json:"quantity"`
}

type BatchAvailabilityResult struct {
	Items       []*AvailabilityResult `json:"items"`
	AllFeasible bool                  `json:"all_feasible"`
}

func (s *AvailabilityService) CheckBatchAvailability(
	ctx context.Context,
	tenantID string,
	requests []BatchAvailabilityRequest,
) (*BatchAvailabilityResult, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant ID is required")
	}
	if len(requests) == 0 {
		return nil, fmt.Errorf("at least one request is required")
	}

	results := make([]*AvailabilityResult, 0, len(requests))
	allFeasible := true

	for _, req := range requests {
		result, err := s.CheckEquipmentAvailability(ctx, tenantID, req.EquipmentID, req.Quantity)
		if err != nil {
			// Log error but continue processing other items
			s.logger.Error("failed to check availability", err, "equipment_id", req.EquipmentID)
			results = append(results, &AvailabilityResult{
				EquipmentID: req.EquipmentID,
				Feasible:    false,
				Message:     err.Error(),
			})
			allFeasible = false
			continue
		}
		if !result.Feasible {
			allFeasible = false
		}
		results = append(results, result)
	}

	return &BatchAvailabilityResult{
		Items:       results,
		AllFeasible: allFeasible,
	}, nil
}

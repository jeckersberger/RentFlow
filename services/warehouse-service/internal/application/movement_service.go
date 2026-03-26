package application

import (
	"context"
	"fmt"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/warehouse-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/warehouse-service/internal/ports"
)

type MovementService struct {
	movementRepo ports.MovementRepository
	locationRepo ports.LocationRepository
	logger       logger.Logger
}

func NewMovementService(
	movementRepo ports.MovementRepository,
	locationRepo ports.LocationRepository,
	logger logger.Logger,
) *MovementService {
	return &MovementService{
		movementRepo: movementRepo,
		locationRepo: locationRepo,
		logger:       logger,
	}
}

func (s *MovementService) RecordMovement(ctx context.Context, cmd RecordMovementCommand) (*MovementDTO, error) {
	if cmd.TenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}
	if cmd.EquipmentID == "" {
		return nil, domain.NewDomainError("EQUIPMENT_REQUIRED", "equipment ID is required", nil)
	}

	// Verify to-location exists
	toLocation, err := s.locationRepo.GetByID(ctx, cmd.TenantID, cmd.ToLocationID)
	if err != nil {
		return nil, domain.NewDomainError("LOCATION_NOT_FOUND", "to-location not found", err)
	}

	// Check capacity
	if !toLocation.CanAddCapacity(cmd.Quantity) {
		return nil, domain.NewDomainError("INSUFFICIENT_CAPACITY", "location has insufficient capacity", nil)
	}

	movementID := fmt.Sprintf("mov_%d", hashString(cmd.TenantID+cmd.EquipmentID+time.Now().String()))
	movement := domain.NewMovement(
		movementID,
		cmd.TenantID,
		cmd.EquipmentID,
		cmd.ToLocationID,
		cmd.MovementType,
		cmd.Quantity,
		cmd.UserID,
	)

	movement.FromLocationID = cmd.FromLocationID
	movement.ProjectID = cmd.ProjectID
	movement.Reason = cmd.Reason

	// Validate
	if err := movement.Validate(); err != nil {
		return nil, domain.NewDomainError("VALIDATION_ERROR", err.Error(), nil)
	}

	// Update capacity
	if err := toLocation.AddToCapacity(cmd.Quantity); err != nil {
		return nil, domain.NewDomainError("CAPACITY_ERROR", err.Error(), nil)
	}

	// Update from location if specified
	if cmd.FromLocationID != nil {
		fromLocation, err := s.locationRepo.GetByID(ctx, cmd.TenantID, *cmd.FromLocationID)
		if err != nil {
			return nil, domain.NewDomainError("FROM_LOCATION_NOT_FOUND", "from-location not found", err)
		}
		if err := fromLocation.RemoveFromCapacity(cmd.Quantity); err != nil {
			return nil, domain.NewDomainError("CAPACITY_ERROR", err.Error(), nil)
		}
		if err := s.locationRepo.Update(ctx, fromLocation); err != nil {
			return nil, domain.NewDomainError("UPDATE_ERROR", "failed to update from-location", err)
		}
	}

	// Persist movement
	if err := s.movementRepo.Create(ctx, movement); err != nil {
		return nil, domain.NewDomainError("CREATE_ERROR", "failed to record movement", err)
	}

	// Update to-location
	if err := s.locationRepo.Update(ctx, toLocation); err != nil {
		return nil, domain.NewDomainError("UPDATE_ERROR", "failed to update to-location", err)
	}

	s.logger.Info("Movement recorded", "id", movement.ID, "equipment", cmd.EquipmentID)
	return MovementToDTO(movement), nil
}

func (s *MovementService) GetMovementHistory(ctx context.Context, tenantID string, query HistoryQuery) (*PaginatedResult, error) {
	if tenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	repoQuery := &ports.MovementListQuery{
		EquipmentID:  query.EquipmentID,
		FromLocation: query.FromLocation,
		ToLocation:   query.ToLocation,
		MovementType: query.MovementType,
		Limit:        query.Limit,
		Offset:       query.Offset,
	}

	movements, total, err := s.movementRepo.List(ctx, tenantID, repoQuery)
	if err != nil {
		return nil, domain.NewDomainError("QUERY_ERROR", "failed to get movement history", err)
	}

	dtos := make([]*MovementDTO, len(movements))
	for i, mov := range movements {
		dtos[i] = MovementToDTO(mov)
	}

	return &PaginatedResult{
		Data:   dtos,
		Total:  total,
		Limit:  query.Limit,
		Offset: query.Offset,
	}, nil
}

func (s *MovementService) GetEquipmentLocation(ctx context.Context, tenantID, equipmentID string) (*MovementDTO, error) {
	if tenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	movements, _, err := s.movementRepo.GetByEquipmentID(ctx, tenantID, equipmentID, 1, 0)
	if err != nil || len(movements) == 0 {
		return nil, domain.NewDomainError("NOT_FOUND", "no movements found for equipment", nil)
	}

	return MovementToDTO(movements[0]), nil
}

func (s *MovementService) GetEquipmentHistory(ctx context.Context, tenantID, equipmentID string, limit, offset int) (*PaginatedResult, error) {
	if tenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	movements, total, err := s.movementRepo.GetByEquipmentID(ctx, tenantID, equipmentID, limit, offset)
	if err != nil {
		return nil, domain.NewDomainError("QUERY_ERROR", "failed to get equipment history", err)
	}

	dtos := make([]*MovementDTO, len(movements))
	for i, mov := range movements {
		dtos[i] = MovementToDTO(mov)
	}

	return &PaginatedResult{
		Data:   dtos,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}, nil
}

func hashString(s string) int64 {
	h := int64(5381)
	for _, c := range s {
		h = ((h << 5) + h) + int64(c)
	}
	return h & 0x7FFFFFFFFFFFFFFF
}

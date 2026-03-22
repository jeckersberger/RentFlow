package application

import (
	"context"
	"fmt"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/warehouse-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/warehouse-service/internal/ports"
)

type InventoryCheckService struct {
	checkRepo ports.InventoryCheckRepository
	logger    logger.Logger
}

func NewInventoryCheckService(
	checkRepo ports.InventoryCheckRepository,
	logger logger.Logger,
) *InventoryCheckService {
	return &InventoryCheckService{
		checkRepo: checkRepo,
		logger:    logger,
	}
}

func (s *InventoryCheckService) StartCheck(ctx context.Context, cmd StartInventoryCheckCommand) (*InventoryCheckDTO, error) {
	if cmd.TenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}
	if cmd.Name == "" {
		return nil, domain.NewDomainError("NAME_REQUIRED", "name is required", nil)
	}

	checkID := fmt.Sprintf("inv_%d", hashCheckString(cmd.TenantID+cmd.Name+time.Now().String()))
	checkType := domain.InventoryCheckTypeFull
	if cmd.CheckType != "" {
		checkType = domain.InventoryCheckType(cmd.CheckType)
	}
	check := domain.NewInventoryCheck(checkID, cmd.TenantID, cmd.Name, checkType)
	check.WarehouseID = cmd.WarehouseID
	check.ZoneID = cmd.ZoneID
	check.Description = cmd.Description

	// Validate
	if err := check.Validate(); err != nil {
		return nil, domain.NewDomainError("VALIDATION_ERROR", err.Error(), nil)
	}

	// Start the check
	if err := check.Start(); err != nil {
		return nil, domain.NewDomainError("INVALID_STATE", err.Error(), nil)
	}

	// Persist
	if err := s.checkRepo.Create(ctx, check); err != nil {
		return nil, domain.NewDomainError("CREATE_ERROR", "failed to start inventory check", err)
	}

	s.logger.Info("Inventory check started", "id", check.ID, "name", cmd.Name)
	return InventoryCheckToDTO(check), nil
}

func (s *InventoryCheckService) GetCheck(ctx context.Context, tenantID, checkID string) (*InventoryCheckDTO, error) {
	if tenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	check, err := s.checkRepo.GetByID(ctx, tenantID, checkID)
	if err != nil {
		return nil, domain.NewDomainError("NOT_FOUND", "inventory check not found", err)
	}

	return InventoryCheckToDTO(check), nil
}

func (s *InventoryCheckService) ListChecks(ctx context.Context, tenantID string, limit, offset int) (*PaginatedResult, error) {
	if tenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	checks, total, err := s.checkRepo.List(ctx, tenantID, limit, offset)
	if err != nil {
		return nil, domain.NewDomainError("QUERY_ERROR", "failed to list inventory checks", err)
	}

	dtos := make([]*InventoryCheckDTO, len(checks))
	for i, check := range checks {
		dtos[i] = InventoryCheckToDTO(check)
	}

	return &PaginatedResult{
		Data:   dtos,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}, nil
}

func (s *InventoryCheckService) ScanItem(ctx context.Context, cmd ScanInventoryItemCommand) (*InventoryCheckDTO, error) {
	if cmd.TenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	check, err := s.checkRepo.GetByID(ctx, cmd.TenantID, cmd.CheckID)
	if err != nil {
		return nil, domain.NewDomainError("NOT_FOUND", "inventory check not found", err)
	}

	if check.Status != domain.InventoryCheckStatusInProgress {
		return nil, domain.NewDomainError("INVALID_STATE", "inventory check is not in progress", nil)
	}

	err = check.ScanItem(cmd.EquipmentID, cmd.LocationID)
	if err != nil {
		return nil, domain.NewDomainError("SCAN_ERROR", err.Error(), nil)
	}

	// Update notes if provided
	if cmd.Notes != "" {
		for i := range check.Items {
			if check.Items[i].EquipmentID == cmd.EquipmentID {
				check.Items[i].Notes = cmd.Notes
				break
			}
		}
	}

	// Persist updated check
	if err := s.checkRepo.Update(ctx, check); err != nil {
		return nil, domain.NewDomainError("UPDATE_ERROR", "failed to update inventory check", err)
	}

	s.logger.Info("Item scanned", "check_id", cmd.CheckID, "equipment_id", cmd.EquipmentID)
	return InventoryCheckToDTO(check), nil
}

func (s *InventoryCheckService) CompleteCheck(ctx context.Context, cmd CompleteInventoryCheckCommand) (*InventoryCheckDTO, error) {
	if cmd.TenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	check, err := s.checkRepo.GetByID(ctx, cmd.TenantID, cmd.CheckID)
	if err != nil {
		return nil, domain.NewDomainError("NOT_FOUND", "inventory check not found", err)
	}

	if check.Status != domain.InventoryCheckStatusInProgress {
		return nil, domain.NewDomainError("INVALID_STATE", "inventory check is not in progress", nil)
	}

	if err := check.Complete(cmd.CompletedBy); err != nil {
		return nil, domain.NewDomainError("COMPLETION_ERROR", err.Error(), nil)
	}

	// Persist
	if err := s.checkRepo.Update(ctx, check); err != nil {
		return nil, domain.NewDomainError("UPDATE_ERROR", "failed to complete inventory check", err)
	}

	s.logger.Info("Inventory check completed", "id", check.ID)
	return InventoryCheckToDTO(check), nil
}

func (s *InventoryCheckService) GetDiscrepancies(ctx context.Context, tenantID, checkID string) (interface{}, error) {
	if tenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	check, err := s.checkRepo.GetByID(ctx, tenantID, checkID)
	if err != nil {
		return nil, domain.NewDomainError("NOT_FOUND", "inventory check not found", err)
	}

	discrepancies := check.GetDiscrepancies()
	dtos := make([]InventoryCheckItemDTO, len(discrepancies))

	for i, item := range discrepancies {
		dtos[i] = InventoryCheckItemDTO{
			EquipmentID:   item.EquipmentID,
			ExpectedCount: item.ExpectedCount,
			ActualCount:   item.ActualCount,
			Status:        string(item.Status),
			Notes:         item.Notes,
		}
		if item.ScannedAt != nil {
			scannedAtStr := item.ScannedAt.Format("2006-01-02T15:04:05Z")
			dtos[i].ScannedAt = &scannedAtStr
		}
	}

	return map[string]interface{}{
		"check_id":          check.ID,
		"discrepancy_count": len(dtos),
		"discrepancies":     dtos,
	}, nil
}

func hashCheckString(s string) int64 {
	h := int64(5381)
	for _, c := range s {
		h = ((h << 5) + h) + int64(c)
	}
	return h & 0x7FFFFFFFFFFFFFFF
}

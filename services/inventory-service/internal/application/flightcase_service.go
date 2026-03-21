package application

import (
	"context"
	"fmt"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/inventory-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/inventory-service/internal/ports"
)

type FlightcaseService struct {
	fcRepo    ports.FlightcaseRepository
	equipRepo ports.EquipmentRepository
	logger    logger.Logger
}

func NewFlightcaseService(
	fcRepo ports.FlightcaseRepository,
	equipRepo ports.EquipmentRepository,
	logger logger.Logger,
) *FlightcaseService {
	return &FlightcaseService{
		fcRepo:    fcRepo,
		equipRepo: equipRepo,
		logger:    logger,
	}
}

func (s *FlightcaseService) CreateFlightcase(ctx context.Context, cmd CreateFlightcaseCommand) (*FlightcaseDTO, error) {
	if cmd.TenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}
	if cmd.Barcode == "" {
		return nil, domain.NewDomainError("BARCODE_REQUIRED", "barcode is required", nil)
	}

	// Check if barcode already exists
	existing, _ := s.fcRepo.GetByBarcode(ctx, cmd.TenantID, cmd.Barcode)
	if existing != nil {
		return nil, domain.NewDomainError("BARCODE_EXISTS", "barcode already exists", nil)
	}

	// Generate ID
	flightcaseID := fmt.Sprintf("fc_%d", hashString(cmd.TenantID+cmd.Barcode))

	// Create new flightcase
	fc := domain.NewFlightcase(flightcaseID, cmd.TenantID, cmd.Name, cmd.Barcode, cmd.CreatedByUserID)
	fc.Description = cmd.Description
	fc.Weight = cmd.Weight

	// Validate
	if err := fc.Validate(); err != nil {
		return nil, domain.NewDomainError("VALIDATION_ERROR", err.Error(), nil)
	}

	// Persist
	if err := s.fcRepo.Create(ctx, fc); err != nil {
		return nil, domain.NewDomainError("CREATE_ERROR", "failed to create flightcase", err)
	}

	s.logger.Info("Flightcase created", "id", fc.ID, "tenant_id", cmd.TenantID)
	return FlightcaseToDTO(fc), nil
}

func (s *FlightcaseService) UpdateFlightcase(ctx context.Context, cmd UpdateFlightcaseCommand) (*FlightcaseDTO, error) {
	if cmd.TenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	fc, err := s.fcRepo.GetByID(ctx, cmd.TenantID, cmd.ID)
	if err != nil {
		return nil, domain.NewDomainError("NOT_FOUND", "flightcase not found", err)
	}

	// Update fields
	fc.Name = cmd.Name
	fc.Description = cmd.Description
	fc.Weight = cmd.Weight

	// Persist
	if err := s.fcRepo.Update(ctx, fc); err != nil {
		return nil, domain.NewDomainError("UPDATE_ERROR", "failed to update flightcase", err)
	}

	s.logger.Info("Flightcase updated", "id", cmd.ID, "tenant_id", cmd.TenantID)
	return FlightcaseToDTO(fc), nil
}

func (s *FlightcaseService) GetFlightcase(ctx context.Context, tenantID, flightcaseID string) (*FlightcaseDTO, error) {
	if tenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	fc, err := s.fcRepo.GetByID(ctx, tenantID, flightcaseID)
	if err != nil {
		return nil, domain.NewDomainError("NOT_FOUND", "flightcase not found", err)
	}

	return FlightcaseToDTO(fc), nil
}

func (s *FlightcaseService) ListFlightcases(ctx context.Context, tenantID string, limit, offset int) (*PaginatedResult, error) {
	if tenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	result, err := s.fcRepo.List(ctx, tenantID, limit, offset)
	if err != nil {
		return nil, domain.NewDomainError("QUERY_ERROR", "failed to list flightcases", err)
	}

	dtos := make([]*FlightcaseDTO, len(result.Items))
	for i, fc := range result.Items {
		dtos[i] = FlightcaseToDTO(fc)
	}

	return &PaginatedResult{
		Data:   dtos,
		Total:  result.Total,
		Limit:  result.Limit,
		Offset: result.Offset,
	}, nil
}

func (s *FlightcaseService) GetByBarcode(ctx context.Context, tenantID, barcode string) (*FlightcaseDTO, error) {
	if tenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}
	if barcode == "" {
		return nil, domain.NewDomainError("BARCODE_REQUIRED", "barcode is required", nil)
	}

	fc, err := s.fcRepo.GetByBarcode(ctx, tenantID, barcode)
	if err != nil {
		return nil, domain.NewDomainError("NOT_FOUND", "flightcase not found", err)
	}

	return FlightcaseToDTO(fc), nil
}

func (s *FlightcaseService) AddItem(ctx context.Context, cmd AddFlightcaseItemCommand) (*FlightcaseDTO, error) {
	if cmd.TenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	// Verify flightcase exists
	fc, err := s.fcRepo.GetByID(ctx, cmd.TenantID, cmd.FlightcaseID)
	if err != nil {
		return nil, domain.NewDomainError("NOT_FOUND", "flightcase not found", err)
	}

	// Verify equipment exists
	_, err = s.equipRepo.GetByID(ctx, cmd.TenantID, cmd.EquipmentID)
	if err != nil {
		return nil, domain.NewDomainError("EQUIPMENT_NOT_FOUND", "equipment not found", err)
	}

	// Add item
	if err := fc.AddItem(cmd.EquipmentID, cmd.Quantity); err != nil {
		return nil, domain.NewDomainError("ADD_ITEM_ERROR", err.Error(), nil)
	}

	// Persist
	if err := s.fcRepo.Update(ctx, fc); err != nil {
		return nil, domain.NewDomainError("UPDATE_ERROR", "failed to add item", err)
	}

	s.logger.Info("Item added to flightcase", "flightcase_id", cmd.FlightcaseID, "equipment_id", cmd.EquipmentID)
	return FlightcaseToDTO(fc), nil
}

func (s *FlightcaseService) RemoveItem(ctx context.Context, cmd RemoveFlightcaseItemCommand) (*FlightcaseDTO, error) {
	if cmd.TenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	// Verify flightcase exists
	fc, err := s.fcRepo.GetByID(ctx, cmd.TenantID, cmd.FlightcaseID)
	if err != nil {
		return nil, domain.NewDomainError("NOT_FOUND", "flightcase not found", err)
	}

	// Remove item
	if err := fc.RemoveItem(cmd.EquipmentID, cmd.Quantity); err != nil {
		return nil, domain.NewDomainError("REMOVE_ITEM_ERROR", err.Error(), nil)
	}

	// Persist
	if err := s.fcRepo.Update(ctx, fc); err != nil {
		return nil, domain.NewDomainError("UPDATE_ERROR", "failed to remove item", err)
	}

	s.logger.Info("Item removed from flightcase", "flightcase_id", cmd.FlightcaseID, "equipment_id", cmd.EquipmentID)
	return FlightcaseToDTO(fc), nil
}

func (s *FlightcaseService) DeleteFlightcase(ctx context.Context, tenantID, flightcaseID string) error {
	if tenantID == "" {
		return domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	// Verify exists
	_, err := s.fcRepo.GetByID(ctx, tenantID, flightcaseID)
	if err != nil {
		return domain.NewDomainError("NOT_FOUND", "flightcase not found", err)
	}

	if err := s.fcRepo.Delete(ctx, tenantID, flightcaseID); err != nil {
		return domain.NewDomainError("DELETE_ERROR", "failed to delete flightcase", err)
	}

	s.logger.Info("Flightcase deleted", "id", flightcaseID, "tenant_id", tenantID)
	return nil
}

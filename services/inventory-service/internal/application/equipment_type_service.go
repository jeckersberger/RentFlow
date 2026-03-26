package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/inventory-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/inventory-service/internal/ports"
)

type EquipmentTypeService struct {
	typeRepo  ports.EquipmentTypeRepository
	equipRepo ports.EquipmentRepository
	catRepo   ports.CategoryRepository
	logger    logger.Logger
}

func NewEquipmentTypeService(
	typeRepo ports.EquipmentTypeRepository,
	equipRepo ports.EquipmentRepository,
	catRepo ports.CategoryRepository,
	logger logger.Logger,
) *EquipmentTypeService {
	return &EquipmentTypeService{
		typeRepo:  typeRepo,
		equipRepo: equipRepo,
		catRepo:   catRepo,
		logger:    logger,
	}
}

func (s *EquipmentTypeService) CreateType(ctx context.Context, cmd CreateEquipmentTypeCommand) (*EquipmentTypeDTO, error) {
	if cmd.TenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}
	if cmd.Name == "" {
		return nil, domain.NewDomainError("NAME_REQUIRED", "equipment type name is required", nil)
	}
	// Verify category exists if specified
	if cmd.CategoryID != "" {
		_, err := s.catRepo.GetByID(ctx, cmd.TenantID, cmd.CategoryID)
		if err != nil {
			return nil, domain.NewDomainError("INVALID_CATEGORY", "category not found", err)
		}
	}

	// Generate ID
	typeID := fmt.Sprintf("eqtype_%d", hashString(cmd.TenantID+cmd.Name))

	et := domain.NewEquipmentType(typeID, cmd.TenantID, cmd.Name, cmd.CategoryID)
	et.Description = cmd.Description
	et.Manufacturer = cmd.Manufacturer
	et.Model = cmd.Model
	et.SKUPrefix = cmd.SKUPrefix
	et.RentalPriceDay = cmd.RentalPriceDay
	et.RentalPriceWeek = cmd.RentalPriceWeek
	et.ReplacementValue = cmd.ReplacementValue
	et.Weight = cmd.Weight
	et.Dimensions = domain.Dimensions{
		Length: cmd.Dimensions.Length,
		Width:  cmd.Dimensions.Width,
		Height: cmd.Dimensions.Height,
		Unit:   cmd.Dimensions.Unit,
	}
	if et.Dimensions.Unit == "" {
		et.Dimensions.Unit = "cm"
	}
	et.ImageURL = cmd.ImageURL
	if cmd.Tags != nil {
		et.Tags = cmd.Tags
	}
	if cmd.CustomFields != nil {
		et.CustomFields = cmd.CustomFields
	}

	if err := et.Validate(); err != nil {
		return nil, domain.NewDomainError("VALIDATION_ERROR", err.Error(), nil)
	}

	if err := s.typeRepo.Create(ctx, et); err != nil {
		return nil, domain.NewDomainError("CREATE_ERROR", "failed to create equipment type", err)
	}

	s.logger.Info("Equipment type created", "id", et.ID, "tenant_id", cmd.TenantID, "name", cmd.Name)
	dto := EquipmentTypeToDTO(et)
	dto.ItemCount = 0
	return dto, nil
}

func (s *EquipmentTypeService) UpdateType(ctx context.Context, cmd UpdateEquipmentTypeCommand) (*EquipmentTypeDTO, error) {
	if cmd.TenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	et, err := s.typeRepo.GetByID(ctx, cmd.TenantID, cmd.ID)
	if err != nil {
		return nil, domain.NewDomainError("NOT_FOUND", "equipment type not found", err)
	}

	// Update fields
	et.Name = cmd.Name
	et.Description = cmd.Description
	et.CategoryID = cmd.CategoryID
	et.Manufacturer = cmd.Manufacturer
	et.Model = cmd.Model
	et.SKUPrefix = cmd.SKUPrefix
	et.RentalPriceDay = cmd.RentalPriceDay
	et.RentalPriceWeek = cmd.RentalPriceWeek
	et.ReplacementValue = cmd.ReplacementValue
	et.Weight = cmd.Weight
	et.Dimensions = domain.Dimensions{
		Length: cmd.Dimensions.Length,
		Width:  cmd.Dimensions.Width,
		Height: cmd.Dimensions.Height,
		Unit:   cmd.Dimensions.Unit,
	}
	if et.Dimensions.Unit == "" {
		et.Dimensions.Unit = "cm"
	}
	et.ImageURL = cmd.ImageURL
	if cmd.Tags != nil {
		et.Tags = cmd.Tags
	}
	if cmd.CustomFields != nil {
		et.CustomFields = cmd.CustomFields
	}
	et.UpdatedAt = time.Now()

	if err := s.typeRepo.Update(ctx, et); err != nil {
		return nil, domain.NewDomainError("UPDATE_ERROR", "failed to update equipment type", err)
	}

	s.logger.Info("Equipment type updated", "id", cmd.ID, "tenant_id", cmd.TenantID)

	dto := EquipmentTypeToDTO(et)
	count, _ := s.typeRepo.CountItems(ctx, cmd.TenantID, cmd.ID)
	dto.ItemCount = count
	return dto, nil
}

func (s *EquipmentTypeService) GetType(ctx context.Context, tenantID, id string) (*EquipmentTypeDTO, error) {
	if tenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	et, err := s.typeRepo.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, domain.NewDomainError("NOT_FOUND", "equipment type not found", err)
	}

	dto := EquipmentTypeToDTO(et)
	count, _ := s.typeRepo.CountItems(ctx, tenantID, id)
	dto.ItemCount = count
	return dto, nil
}

func (s *EquipmentTypeService) ListTypes(ctx context.Context, tenantID, categoryID string, limit, offset int) (*PaginatedResult, error) {
	if tenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	items, total, err := s.typeRepo.List(ctx, tenantID, categoryID, limit, offset)
	if err != nil {
		return nil, domain.NewDomainError("QUERY_ERROR", "failed to list equipment types", err)
	}

	dtos := make([]*EquipmentTypeDTO, len(items))
	for i, et := range items {
		dtos[i] = EquipmentTypeToDTO(et)
		count, _ := s.typeRepo.CountItems(ctx, tenantID, et.ID)
		dtos[i].ItemCount = count
	}

	return &PaginatedResult{
		Data:   dtos,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}, nil
}

func (s *EquipmentTypeService) DeleteType(ctx context.Context, tenantID, id string) error {
	if tenantID == "" {
		return domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	// Verify exists
	_, err := s.typeRepo.GetByID(ctx, tenantID, id)
	if err != nil {
		return domain.NewDomainError("NOT_FOUND", "equipment type not found", err)
	}

	// Check if items exist
	count, _ := s.typeRepo.CountItems(ctx, tenantID, id)
	if count > 0 {
		return domain.NewDomainError("VALIDATION_ERROR",
			fmt.Sprintf("cannot delete equipment type with %d existing items", count), nil)
	}

	if err := s.typeRepo.Delete(ctx, tenantID, id); err != nil {
		return domain.NewDomainError("DELETE_ERROR", "failed to delete equipment type", err)
	}

	s.logger.Info("Equipment type deleted", "id", id, "tenant_id", tenantID)
	return nil
}

func (s *EquipmentTypeService) CreateItemsFromType(ctx context.Context, cmd CreateItemsFromTypeCommand) ([]*EquipmentDTO, error) {
	if cmd.TenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}
	if cmd.Quantity <= 0 {
		return nil, domain.NewDomainError("VALIDATION_ERROR", "quantity must be greater than 0", nil)
	}
	if cmd.Quantity > 100 {
		return nil, domain.NewDomainError("VALIDATION_ERROR", "maximum 100 items per batch", nil)
	}

	// Load the type
	et, err := s.typeRepo.GetByID(ctx, cmd.TenantID, cmd.TypeID)
	if err != nil {
		return nil, domain.NewDomainError("NOT_FOUND", "equipment type not found", err)
	}

	// Get next item number
	nextNum, err := s.typeRepo.NextItemNumber(ctx, cmd.TenantID, cmd.TypeID)
	if err != nil {
		return nil, domain.NewDomainError("QUERY_ERROR", "failed to get next item number", err)
	}

	// Generate short name for barcode (first 3 chars uppercase, no spaces)
	nameShort := strings.ToUpper(strings.ReplaceAll(et.Name, " ", ""))
	if len(nameShort) > 6 {
		nameShort = nameShort[:6]
	}

	skuPrefix := et.SKUPrefix
	if skuPrefix == "" {
		skuPrefix = nameShort
	}

	var results []*EquipmentDTO

	for i := 0; i < cmd.Quantity; i++ {
		itemNum := nextNum + i
		serialNumber := fmt.Sprintf("%s-%04d", skuPrefix, itemNum)
		barcode := fmt.Sprintf("RF-%s-%04d", nameShort, itemNum)
		sku := fmt.Sprintf("%s-%04d", skuPrefix, itemNum)

		// Generate unique equipment ID
		equipmentID := fmt.Sprintf("equip_%d", hashString(cmd.TenantID+barcode+fmt.Sprintf("%d", time.Now().UnixNano()+int64(i))))

		eq := domain.NewEquipment(equipmentID, cmd.TenantID, et.Name, et.CategoryID, sku, barcode, cmd.CreatedByUserID)
		eq.Description = et.Description
		eq.SerialNumber = serialNumber
		eq.RentalPriceDay = et.RentalPriceDay
		eq.RentalPriceWeek = et.RentalPriceWeek
		eq.PurchasePrice = et.ReplacementValue
		eq.Weight = et.Weight
		eq.Dimensions = et.Dimensions
		eq.Tags = et.Tags
		eq.CustomFields = make(map[string]string)
		for k, v := range et.CustomFields {
			eq.CustomFields[k] = v
		}

		if err := eq.Validate(); err != nil {
			return nil, domain.NewDomainError("VALIDATION_ERROR", fmt.Sprintf("item %d: %s", i+1, err.Error()), nil)
		}

		if err := s.equipRepo.Create(ctx, eq); err != nil {
			return nil, domain.NewDomainError("CREATE_ERROR",
				fmt.Sprintf("failed to create item %d: %s", i+1, err.Error()), err)
		}

		// Set the equipment_type_id and item_number via direct update
		// since Equipment domain doesn't have these fields yet
		s.setTypeFields(ctx, cmd.TenantID, equipmentID, cmd.TypeID, itemNum)

		results = append(results, EquipmentToDTO(eq))
	}

	s.logger.Info("Items created from type", "type_id", cmd.TypeID, "quantity", cmd.Quantity, "tenant_id", cmd.TenantID)
	return results, nil
}

// setTypeFields sets equipment_type_id and item_number on equipment after creation
func (s *EquipmentTypeService) setTypeFields(ctx context.Context, tenantID, equipmentID, typeID string, itemNumber int) {
	// This is a workaround since the Equipment domain doesn't have these fields yet.
	// We update them directly via the type repository's db connection.
	// In a future refactor, add these fields to the Equipment domain.
	if repo, ok := s.typeRepo.(interface {
		SetEquipmentTypeFields(ctx context.Context, tenantID, equipmentID, typeID string, itemNumber int) error
	}); ok {
		if err := repo.SetEquipmentTypeFields(ctx, tenantID, equipmentID, typeID, itemNumber); err != nil {
			s.logger.Error("Failed to set type fields on equipment", err, "equipment_id", equipmentID)
		}
	}
}

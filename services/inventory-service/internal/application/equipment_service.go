package application

import (
	"context"
	"fmt"
	"io"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/pkg/common/storage"
	"github.com/jeckersberger/rentflow/services/inventory-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/inventory-service/internal/ports"
)

type EquipmentService struct {
	equipRepo ports.EquipmentRepository
	catRepo   ports.CategoryRepository
	storage   storage.StorageAdapter
	logger    logger.Logger
}

func NewEquipmentService(
	equipRepo ports.EquipmentRepository,
	catRepo ports.CategoryRepository,
	storageAdapter storage.StorageAdapter,
	logger logger.Logger,
) *EquipmentService {
	return &EquipmentService{
		equipRepo: equipRepo,
		catRepo:   catRepo,
		storage:   storageAdapter,
		logger:    logger,
	}
}

func (s *EquipmentService) CreateEquipment(ctx context.Context, cmd CreateEquipmentCommand) (*EquipmentDTO, error) {
	if cmd.TenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}
	if cmd.Barcode == "" {
		return nil, domain.NewDomainError("BARCODE_REQUIRED", "barcode is required", nil)
	}

	// Verify category exists
	_, err := s.catRepo.GetByID(ctx, cmd.TenantID, cmd.CategoryID)
	if err != nil {
		return nil, domain.NewDomainError("INVALID_CATEGORY", "category not found", err)
	}

	// Check if barcode already exists
	existing, _ := s.equipRepo.GetByBarcode(ctx, cmd.TenantID, cmd.Barcode)
	if existing != nil {
		return nil, domain.NewDomainError("BARCODE_EXISTS", "barcode already exists", nil)
	}

	// Generate equipment ID (in real scenario, use UUID)
	equipmentID := fmt.Sprintf("equip_%d", hashString(cmd.TenantID+cmd.Barcode))

	// Create new equipment
	eq := domain.NewEquipment(equipmentID, cmd.TenantID, cmd.Name, cmd.CategoryID, cmd.SKU, cmd.Barcode, cmd.CreatedByUserID)
	eq.Description = cmd.Description
	eq.SerialNumber = cmd.SerialNumber
	eq.PurchaseDate = cmd.PurchaseDate
	eq.PurchasePrice = cmd.PurchasePrice
	eq.RentalPriceDay = cmd.RentalPriceDay
	eq.RentalPriceWeek = cmd.RentalPriceWeek
	eq.Weight = cmd.Weight
	eq.Dimensions = domain.Dimensions{
		Length: cmd.Dimensions.Length,
		Width:  cmd.Dimensions.Width,
		Height: cmd.Dimensions.Height,
		Unit:   cmd.Dimensions.Unit,
	}
	eq.LocationID = cmd.LocationID
	eq.Tags = cmd.Tags
	eq.CustomFields = cmd.CustomFields

	// Validate
	if err := eq.Validate(); err != nil {
		return nil, domain.NewDomainError("VALIDATION_ERROR", err.Error(), nil)
	}

	// Persist
	if err := s.equipRepo.Create(ctx, eq); err != nil {
		return nil, domain.NewDomainError("CREATE_ERROR", "failed to create equipment", err)
	}

	s.logger.Info("Equipment created", "id", eq.ID, "tenant_id", cmd.TenantID)
	return EquipmentToDTO(eq), nil
}

func (s *EquipmentService) UpdateEquipment(ctx context.Context, cmd UpdateEquipmentCommand) (*EquipmentDTO, error) {
	if cmd.TenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	eq, err := s.equipRepo.GetByID(ctx, cmd.TenantID, cmd.ID)
	if err != nil {
		return nil, domain.NewDomainError("NOT_FOUND", "equipment not found", err)
	}

	// Update fields
	eq.Name = cmd.Name
	eq.Description = cmd.Description
	eq.CategoryID = cmd.CategoryID
	eq.PurchaseDate = cmd.PurchaseDate
	eq.PurchasePrice = cmd.PurchasePrice
	eq.RentalPriceDay = cmd.RentalPriceDay
	eq.RentalPriceWeek = cmd.RentalPriceWeek
	eq.Weight = cmd.Weight
	eq.Dimensions = domain.Dimensions{
		Length: cmd.Dimensions.Length,
		Width:  cmd.Dimensions.Width,
		Height: cmd.Dimensions.Height,
		Unit:   cmd.Dimensions.Unit,
	}
	eq.Tags = cmd.Tags
	eq.CustomFields = cmd.CustomFields

	// Persist
	if err := s.equipRepo.Update(ctx, eq); err != nil {
		return nil, domain.NewDomainError("UPDATE_ERROR", "failed to update equipment", err)
	}

	s.logger.Info("Equipment updated", "id", cmd.ID, "tenant_id", cmd.TenantID)
	return EquipmentToDTO(eq), nil
}

func (s *EquipmentService) GetEquipment(ctx context.Context, tenantID, equipmentID string) (*EquipmentDTO, error) {
	if tenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	eq, err := s.equipRepo.GetByID(ctx, tenantID, equipmentID)
	if err != nil {
		return nil, domain.NewDomainError("NOT_FOUND", "equipment not found", err)
	}

	return EquipmentToDTO(eq), nil
}

func (s *EquipmentService) ListEquipment(ctx context.Context, query ListEquipmentQuery) (*PaginatedResult, error) {
	if query.TenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	repoQuery := &ports.EquipmentListQuery{
		TenantID:   query.TenantID,
		Status:     query.Status,
		CategoryID: query.CategoryID,
		LocationID: query.LocationID,
		Limit:      query.Limit,
		Offset:     query.Offset,
	}

	result, err := s.equipRepo.List(ctx, repoQuery)
	if err != nil {
		return nil, domain.NewDomainError("QUERY_ERROR", "failed to list equipment", err)
	}

	dtos := make([]*EquipmentDTO, len(result.Items))
	for i, eq := range result.Items {
		dtos[i] = EquipmentToDTO(eq)
	}

	return &PaginatedResult{
		Data:   dtos,
		Total:  result.Total,
		Limit:  result.Limit,
		Offset: result.Offset,
	}, nil
}

func (s *EquipmentService) SearchEquipment(ctx context.Context, query SearchEquipmentQuery) (*PaginatedResult, error) {
	if query.TenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	result, err := s.equipRepo.Search(ctx, query.TenantID, query.SearchTerm, query.Limit, query.Offset)
	if err != nil {
		return nil, domain.NewDomainError("SEARCH_ERROR", "failed to search equipment", err)
	}

	dtos := make([]*EquipmentDTO, len(result.Items))
	for i, eq := range result.Items {
		dtos[i] = EquipmentToDTO(eq)
	}

	return &PaginatedResult{
		Data:   dtos,
		Total:  result.Total,
		Limit:  result.Limit,
		Offset: result.Offset,
	}, nil
}

func (s *EquipmentService) ChangeStatus(ctx context.Context, cmd ChangeStatusCommand) error {
	if cmd.TenantID == "" {
		return domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	eq, err := s.equipRepo.GetByID(ctx, cmd.TenantID, cmd.ID)
	if err != nil {
		return domain.NewDomainError("NOT_FOUND", "equipment not found", err)
	}

	if err := eq.ChangeStatus(cmd.Status); err != nil {
		return domain.NewDomainError("INVALID_STATUS", err.Error(), nil)
	}

	if err := s.equipRepo.Update(ctx, eq); err != nil {
		return domain.NewDomainError("UPDATE_ERROR", "failed to update status", err)
	}

	s.logger.Info("Equipment status changed", "id", cmd.ID, "status", cmd.Status, "reason", cmd.Reason)
	return nil
}

func (s *EquipmentService) UpdateCondition(ctx context.Context, cmd UpdateConditionCommand) error {
	if cmd.TenantID == "" {
		return domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	eq, err := s.equipRepo.GetByID(ctx, cmd.TenantID, cmd.ID)
	if err != nil {
		return domain.NewDomainError("NOT_FOUND", "equipment not found", err)
	}

	if err := eq.UpdateCondition(cmd.Condition); err != nil {
		return domain.NewDomainError("INVALID_CONDITION", err.Error(), nil)
	}

	if err := s.equipRepo.Update(ctx, eq); err != nil {
		return domain.NewDomainError("UPDATE_ERROR", "failed to update condition", err)
	}

	s.logger.Info("Equipment condition updated", "id", cmd.ID, "condition", cmd.Condition)
	return nil
}

func (s *EquipmentService) SetLocation(ctx context.Context, cmd SetLocationCommand) error {
	if cmd.TenantID == "" {
		return domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	eq, err := s.equipRepo.GetByID(ctx, cmd.TenantID, cmd.ID)
	if err != nil {
		return domain.NewDomainError("NOT_FOUND", "equipment not found", err)
	}

	if err := eq.SetLocation(cmd.LocationID); err != nil {
		return domain.NewDomainError("INVALID_LOCATION", err.Error(), nil)
	}

	if err := s.equipRepo.Update(ctx, eq); err != nil {
		return domain.NewDomainError("UPDATE_ERROR", "failed to set location", err)
	}

	s.logger.Info("Equipment location changed", "id", cmd.ID, "location_id", cmd.LocationID)
	return nil
}

func (s *EquipmentService) AddImage(ctx context.Context, tenantID, equipmentID string, file io.Reader, filename string) (*EquipmentDTO, error) {
	if tenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	eq, err := s.equipRepo.GetByID(ctx, tenantID, equipmentID)
	if err != nil {
		return nil, domain.NewDomainError("NOT_FOUND", "equipment not found", err)
	}

	// Store image (if storage adapter available)
	var imageRef string
	if s.storage != nil {
		meta := storage.FileMetadata{
			TenantID:   tenantID,
			EntityType: "equipment",
			EntityID:   equipmentID,
			Filename:   filename,
		}
		ref, err := s.storage.Store(ctx, file, meta)
		if err != nil {
			return nil, domain.NewDomainError("STORAGE_ERROR", "failed to upload image", err)
		}
		imageRef = ref
	} else {
		imageRef = filename
	}

	if err := eq.AddImage(imageRef); err != nil {
		return nil, domain.NewDomainError("IMAGE_ERROR", err.Error(), nil)
	}

	if err := s.equipRepo.Update(ctx, eq); err != nil {
		return nil, domain.NewDomainError("UPDATE_ERROR", "failed to add image", err)
	}

	s.logger.Info("Image added to equipment", "id", equipmentID, "filename", filename)
	return EquipmentToDTO(eq), nil
}

func (s *EquipmentService) GetByBarcode(ctx context.Context, tenantID, barcode string) (*EquipmentDTO, error) {
	if tenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}
	if barcode == "" {
		return nil, domain.NewDomainError("BARCODE_REQUIRED", "barcode is required", nil)
	}

	eq, err := s.equipRepo.GetByBarcode(ctx, tenantID, barcode)
	if err != nil {
		return nil, domain.NewDomainError("NOT_FOUND", "equipment not found", err)
	}

	return EquipmentToDTO(eq), nil
}

func (s *EquipmentService) DeleteEquipment(ctx context.Context, tenantID, equipmentID string) error {
	if tenantID == "" {
		return domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	// Verify exists
	_, err := s.equipRepo.GetByID(ctx, tenantID, equipmentID)
	if err != nil {
		return domain.NewDomainError("NOT_FOUND", "equipment not found", err)
	}

	if err := s.equipRepo.Delete(ctx, tenantID, equipmentID); err != nil {
		return domain.NewDomainError("DELETE_ERROR", "failed to delete equipment", err)
	}

	s.logger.Info("Equipment deleted", "id", equipmentID, "tenant_id", tenantID)
	return nil
}

// Helper function for generating IDs
func hashString(s string) int64 {
	h := int64(5381)
	for _, c := range s {
		h = ((h << 5) + h) + int64(c)
	}
	return h & 0x7FFFFFFFFFFFFFFF
}

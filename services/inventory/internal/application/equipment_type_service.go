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

// CreateEquipmentTypeRequest holds the data needed to create a new equipment type.
type CreateEquipmentTypeRequest struct {
	CategoryID              *uuid.UUID `json:"category_id,omitempty"`
	Name                    string     `json:"name"`
	Description             string     `json:"description,omitempty"`
	DefaultRentalPriceDay   int64      `json:"default_rental_price_day"`
	DefaultRentalPriceWeek  int64      `json:"default_rental_price_week"`
	DefaultReplacementValue int64      `json:"default_replacement_value"`
}

// UpdateEquipmentTypeRequest holds optional fields for patching an equipment type.
type UpdateEquipmentTypeRequest struct {
	Name                    *string    `json:"name,omitempty"`
	Description             *string    `json:"description,omitempty"`
	CategoryID              *uuid.UUID `json:"category_id,omitempty"`
	DefaultRentalPriceDay   *int64     `json:"default_rental_price_day,omitempty"`
	DefaultRentalPriceWeek  *int64     `json:"default_rental_price_week,omitempty"`
	DefaultReplacementValue *int64     `json:"default_replacement_value,omitempty"`
}

// BulkCreateRequest holds the parameters for creating multiple equipment items
// from an equipment type template.
type BulkCreateRequest struct {
	Count       int    `json:"count"`
	NamePrefix  string `json:"name_prefix"`
	StartNumber int    `json:"start_number"`
}

// ---------------------------------------------------------------------------
// Service
// ---------------------------------------------------------------------------

// EquipmentTypeService implements the application-level use cases for equipment types.
type EquipmentTypeService struct {
	typeRepo      domain.EquipmentTypeRepository
	equipmentRepo domain.EquipmentRepository
	logger        zerolog.Logger
}

// NewEquipmentTypeService constructs a new EquipmentTypeService.
func NewEquipmentTypeService(
	typeRepo domain.EquipmentTypeRepository,
	equipmentRepo domain.EquipmentRepository,
	logger zerolog.Logger,
) *EquipmentTypeService {
	return &EquipmentTypeService{
		typeRepo:      typeRepo,
		equipmentRepo: equipmentRepo,
		logger:        logger.With().Str("service", "equipment_type").Logger(),
	}
}

// Create validates the request and persists a new equipment type.
func (s *EquipmentTypeService) Create(
	ctx context.Context,
	tenantID uuid.UUID,
	req CreateEquipmentTypeRequest,
) (*domain.EquipmentType, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("equipment type name is required")
	}

	now := time.Now()
	equipmentType := &domain.EquipmentType{
		ID:                      uuid.New(),
		TenantID:                tenantID,
		CategoryID:              req.CategoryID,
		Name:                    req.Name,
		Description:             req.Description,
		DefaultRentalPriceDay:   req.DefaultRentalPriceDay,
		DefaultRentalPriceWeek:  req.DefaultRentalPriceWeek,
		DefaultReplacementValue: req.DefaultReplacementValue,
		IsActive:                true,
		CreatedAt:               now,
		UpdatedAt:               now,
	}

	if err := s.typeRepo.Create(ctx, equipmentType); err != nil {
		s.logger.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Msg("failed to create equipment type")
		return nil, fmt.Errorf("create equipment type: %w", err)
	}

	s.logger.Info().
		Str("equipment_type_id", equipmentType.ID.String()).
		Str("tenant_id", tenantID.String()).
		Msg("equipment type created")

	return equipmentType, nil
}

// GetByID retrieves a single equipment type by ID within a tenant scope.
func (s *EquipmentTypeService) GetByID(
	ctx context.Context,
	id uuid.UUID,
	tenantID uuid.UUID,
) (*domain.EquipmentType, error) {
	et, err := s.typeRepo.GetByID(ctx, id, tenantID)
	if err != nil {
		s.logger.Error().Err(err).
			Str("equipment_type_id", id.String()).
			Str("tenant_id", tenantID.String()).
			Msg("failed to get equipment type by id")
		return nil, fmt.Errorf("get equipment type by id: %w", err)
	}
	return et, nil
}

// List returns all equipment types for a tenant.
func (s *EquipmentTypeService) List(
	ctx context.Context,
	tenantID uuid.UUID,
) ([]*domain.EquipmentType, error) {
	types, err := s.typeRepo.List(ctx, tenantID)
	if err != nil {
		s.logger.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Msg("failed to list equipment types")
		return nil, fmt.Errorf("list equipment types: %w", err)
	}
	return types, nil
}

// Update patches an existing equipment type with non-nil fields from the request.
func (s *EquipmentTypeService) Update(
	ctx context.Context,
	id uuid.UUID,
	tenantID uuid.UUID,
	req UpdateEquipmentTypeRequest,
) (*domain.EquipmentType, error) {
	existing, err := s.typeRepo.GetByID(ctx, id, tenantID)
	if err != nil {
		s.logger.Error().Err(err).
			Str("equipment_type_id", id.String()).
			Str("tenant_id", tenantID.String()).
			Msg("failed to fetch equipment type for update")
		return nil, fmt.Errorf("update equipment type – fetch: %w", err)
	}

	if req.Name != nil {
		if *req.Name == "" {
			return nil, fmt.Errorf("equipment type name must not be empty")
		}
		existing.Name = *req.Name
	}
	if req.Description != nil {
		existing.Description = *req.Description
	}
	if req.CategoryID != nil {
		existing.CategoryID = req.CategoryID
	}
	if req.DefaultRentalPriceDay != nil {
		existing.DefaultRentalPriceDay = *req.DefaultRentalPriceDay
	}
	if req.DefaultRentalPriceWeek != nil {
		existing.DefaultRentalPriceWeek = *req.DefaultRentalPriceWeek
	}
	if req.DefaultReplacementValue != nil {
		existing.DefaultReplacementValue = *req.DefaultReplacementValue
	}

	existing.UpdatedAt = time.Now()

	if err := s.typeRepo.Update(ctx, existing); err != nil {
		s.logger.Error().Err(err).
			Str("equipment_type_id", id.String()).
			Str("tenant_id", tenantID.String()).
			Msg("failed to update equipment type")
		return nil, fmt.Errorf("update equipment type: %w", err)
	}

	s.logger.Info().
		Str("equipment_type_id", id.String()).
		Str("tenant_id", tenantID.String()).
		Msg("equipment type updated")

	return existing, nil
}

// Delete removes an equipment type by ID.
func (s *EquipmentTypeService) Delete(
	ctx context.Context,
	id uuid.UUID,
	tenantID uuid.UUID,
) error {
	if err := s.typeRepo.Delete(ctx, id, tenantID); err != nil {
		s.logger.Error().Err(err).
			Str("equipment_type_id", id.String()).
			Str("tenant_id", tenantID.String()).
			Msg("failed to delete equipment type")
		return fmt.Errorf("delete equipment type: %w", err)
	}

	s.logger.Info().
		Str("equipment_type_id", id.String()).
		Str("tenant_id", tenantID.String()).
		Msg("equipment type deleted")

	return nil
}

// BulkCreate creates N equipment items from an equipment type template.
// Each item gets a sequentially numbered name based on the provided prefix
// and starting number, and inherits pricing from the type defaults.
func (s *EquipmentTypeService) BulkCreate(
	ctx context.Context,
	typeID uuid.UUID,
	tenantID uuid.UUID,
	req BulkCreateRequest,
) ([]*domain.Equipment, error) {
	if req.Count < 1 {
		return nil, fmt.Errorf("count must be at least 1")
	}
	if req.NamePrefix == "" {
		return nil, fmt.Errorf("name_prefix is required")
	}

	et, err := s.typeRepo.GetByID(ctx, typeID, tenantID)
	if err != nil {
		s.logger.Error().Err(err).
			Str("equipment_type_id", typeID.String()).
			Str("tenant_id", tenantID.String()).
			Msg("failed to fetch equipment type for bulk create")
		return nil, fmt.Errorf("bulk create – fetch type: %w", err)
	}

	now := time.Now()
	created := make([]*domain.Equipment, 0, req.Count)

	for i := 0; i < req.Count; i++ {
		num := req.StartNumber + i
		equipment := &domain.Equipment{
			ID:                uuid.New(),
			TenantID:          tenantID,
			CategoryID:        et.CategoryID,
			EquipmentTypeID:   &et.ID,
			Name:              fmt.Sprintf("%s %d", req.NamePrefix, num),
			Description:       et.Description,
			Status:            domain.StatusAvailable,
			Condition:         domain.ConditionOperational,
			QuantityTotal:     1,
			QuantityAvailable: 1,
			RentalPriceDay:    et.DefaultRentalPriceDay,
			RentalPriceWeek:   et.DefaultRentalPriceWeek,
			ReplacementValue:  et.DefaultReplacementValue,
			IsActive:          true,
			CreatedAt:         now,
			UpdatedAt:         now,
		}

		if err := s.equipmentRepo.Create(ctx, equipment); err != nil {
			s.logger.Error().Err(err).
				Str("equipment_type_id", typeID.String()).
				Str("tenant_id", tenantID.String()).
				Int("item_number", num).
				Msg("failed to create equipment item during bulk create")
			return created, fmt.Errorf("bulk create – item %d: %w", num, err)
		}

		created = append(created, equipment)
	}

	s.logger.Info().
		Str("equipment_type_id", typeID.String()).
		Str("tenant_id", tenantID.String()).
		Int("count", len(created)).
		Msg("bulk equipment creation completed")

	return created, nil
}

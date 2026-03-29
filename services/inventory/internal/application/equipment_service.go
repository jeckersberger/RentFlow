package application

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/services/inventory/internal/domain"
)

// ---------------------------------------------------------------------------
// Request DTOs
// ---------------------------------------------------------------------------

// CreateEquipmentRequest holds the data needed to create a new equipment item.
type CreateEquipmentRequest struct {
	CategoryID       *uuid.UUID      `json:"category_id,omitempty"`
	EquipmentTypeID  *uuid.UUID      `json:"equipment_type_id,omitempty"`
	Name             string          `json:"name"`
	Description      string          `json:"description,omitempty"`
	SKU              string          `json:"sku,omitempty"`
	Barcode          string          `json:"barcode,omitempty"`
	SerialNumber     string          `json:"serial_number,omitempty"`
	RFIDTag          string          `json:"rfid_tag,omitempty"`
	QuantityTotal    int             `json:"quantity_total"`
	RentalPriceDay   int64           `json:"rental_price_day"`
	RentalPriceWeek  int64           `json:"rental_price_week"`
	ReplacementValue int64           `json:"replacement_value"`
	WeightGrams      *int            `json:"weight_grams,omitempty"`
	WidthMM          *int            `json:"width_mm,omitempty"`
	HeightMM         *int            `json:"height_mm,omitempty"`
	DepthMM          *int            `json:"depth_mm,omitempty"`
	PurchasePrice    int64           `json:"purchase_price"`
	Manufacturer     string          `json:"manufacturer,omitempty"`
	Model            string          `json:"model,omitempty"`
	CustomFields     json.RawMessage `json:"custom_fields,omitempty"`
	Notes            string          `json:"notes,omitempty"`
}

// UpdateEquipmentRequest holds optional fields for patching an equipment item.
type UpdateEquipmentRequest struct {
	Name             *string    `json:"name,omitempty"`
	Description      *string    `json:"description,omitempty"`
	CategoryID       *uuid.UUID `json:"category_id,omitempty"`
	SKU              *string    `json:"sku,omitempty"`
	Barcode          *string    `json:"barcode,omitempty"`
	SerialNumber     *string    `json:"serial_number,omitempty"`
	QuantityTotal    *int       `json:"quantity_total,omitempty"`
	RentalPriceDay   *int64     `json:"rental_price_day,omitempty"`
	RentalPriceWeek  *int64     `json:"rental_price_week,omitempty"`
	ReplacementValue *int64     `json:"replacement_value,omitempty"`
	WeightGrams      *int       `json:"weight_grams,omitempty"`
	Manufacturer     *string    `json:"manufacturer,omitempty"`
	Model            *string    `json:"model,omitempty"`
	Notes            *string    `json:"notes,omitempty"`
}

// UpdateStatusRequest carries the new status value.
type UpdateStatusRequest struct {
	Status string `json:"status"`
}

// UpdateConditionRequest carries the new condition value.
type UpdateConditionRequest struct {
	Condition string `json:"condition"`
}

// AssignRFIDRequest carries the RFID tag to assign.
type AssignRFIDRequest struct {
	RFIDTag string `json:"rfid_tag"`
}

// ---------------------------------------------------------------------------
// Service
// ---------------------------------------------------------------------------

// EquipmentService implements the application-level use cases for equipment.
type EquipmentService struct {
	equipmentRepo domain.EquipmentRepository
	historyRepo   domain.EquipmentHistoryRepository
	logger        zerolog.Logger
}

// NewEquipmentService constructs a new EquipmentService.
func NewEquipmentService(
	equipmentRepo domain.EquipmentRepository,
	historyRepo domain.EquipmentHistoryRepository,
	logger zerolog.Logger,
) *EquipmentService {
	return &EquipmentService{
		equipmentRepo: equipmentRepo,
		historyRepo:   historyRepo,
		logger:        logger.With().Str("service", "equipment").Logger(),
	}
}

// Create validates the request and persists a new equipment item.
func (s *EquipmentService) Create(
	ctx context.Context,
	tenantID uuid.UUID,
	userID uuid.UUID,
	req CreateEquipmentRequest,
) (*domain.Equipment, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("equipment name is required")
	}
	if req.QuantityTotal < 1 {
		return nil, fmt.Errorf("quantity_total must be at least 1")
	}

	now := time.Now()
	equipment := &domain.Equipment{
		ID:                uuid.New(),
		TenantID:          tenantID,
		CategoryID:        req.CategoryID,
		EquipmentTypeID:   req.EquipmentTypeID,
		Name:              req.Name,
		Description:       req.Description,
		SKU:               req.SKU,
		Barcode:           req.Barcode,
		SerialNumber:      req.SerialNumber,
		RFIDTag:           req.RFIDTag,
		Status:            domain.StatusAvailable,
		Condition:         domain.ConditionOperational,
		QuantityTotal:     req.QuantityTotal,
		QuantityAvailable: req.QuantityTotal,
		RentalPriceDay:    req.RentalPriceDay,
		RentalPriceWeek:   req.RentalPriceWeek,
		ReplacementValue:  req.ReplacementValue,
		WeightGrams:       req.WeightGrams,
		WidthMM:           req.WidthMM,
		HeightMM:          req.HeightMM,
		DepthMM:           req.DepthMM,
		PurchasePrice:     req.PurchasePrice,
		Manufacturer:      req.Manufacturer,
		Model:             req.Model,
		CustomFields:      req.CustomFields,
		Notes:             req.Notes,
		IsActive:          true,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	if err := s.equipmentRepo.Create(ctx, equipment); err != nil {
		s.logger.Error().Err(err).Str("tenant_id", tenantID.String()).Msg("failed to create equipment")
		return nil, fmt.Errorf("create equipment: %w", err)
	}

	s.recordHistory(ctx, tenantID, equipment.ID, userID, domain.ActionCreated, map[string]string{
		"name": equipment.Name,
	})

	s.logger.Info().
		Str("equipment_id", equipment.ID.String()).
		Str("tenant_id", tenantID.String()).
		Msg("equipment created")

	return equipment, nil
}

// GetByID retrieves a single equipment item by ID within a tenant scope.
func (s *EquipmentService) GetByID(
	ctx context.Context,
	id uuid.UUID,
	tenantID uuid.UUID,
) (*domain.Equipment, error) {
	equipment, err := s.equipmentRepo.GetByID(ctx, id, tenantID)
	if err != nil {
		s.logger.Error().Err(err).
			Str("equipment_id", id.String()).
			Str("tenant_id", tenantID.String()).
			Msg("failed to get equipment by id")
		return nil, fmt.Errorf("get equipment by id: %w", err)
	}
	return equipment, nil
}

// List returns a filtered, paginated list of equipment for a tenant.
func (s *EquipmentService) List(
	ctx context.Context,
	tenantID uuid.UUID,
	filter domain.EquipmentFilter,
) ([]*domain.Equipment, int64, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PerPage < 1 {
		filter.PerPage = 20
	}

	items, total, err := s.equipmentRepo.List(ctx, tenantID, filter)
	if err != nil {
		s.logger.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Msg("failed to list equipment")
		return nil, 0, fmt.Errorf("list equipment: %w", err)
	}
	return items, total, nil
}

// Update patches an existing equipment item with non-nil fields from the request.
func (s *EquipmentService) Update(
	ctx context.Context,
	id uuid.UUID,
	tenantID uuid.UUID,
	userID uuid.UUID,
	req UpdateEquipmentRequest,
) (*domain.Equipment, error) {
	existing, err := s.equipmentRepo.GetByID(ctx, id, tenantID)
	if err != nil {
		s.logger.Error().Err(err).
			Str("equipment_id", id.String()).
			Str("tenant_id", tenantID.String()).
			Msg("failed to fetch equipment for update")
		return nil, fmt.Errorf("update equipment – fetch: %w", err)
	}

	if req.Name != nil {
		existing.Name = *req.Name
	}
	if req.Description != nil {
		existing.Description = *req.Description
	}
	if req.CategoryID != nil {
		existing.CategoryID = req.CategoryID
	}
	if req.SKU != nil {
		existing.SKU = *req.SKU
	}
	if req.Barcode != nil {
		existing.Barcode = *req.Barcode
	}
	if req.SerialNumber != nil {
		existing.SerialNumber = *req.SerialNumber
	}
	if req.QuantityTotal != nil {
		diff := *req.QuantityTotal - existing.QuantityTotal
		existing.QuantityTotal = *req.QuantityTotal
		existing.QuantityAvailable += diff
		if existing.QuantityAvailable < 0 {
			existing.QuantityAvailable = 0
		}
	}
	if req.RentalPriceDay != nil {
		existing.RentalPriceDay = *req.RentalPriceDay
	}
	if req.RentalPriceWeek != nil {
		existing.RentalPriceWeek = *req.RentalPriceWeek
	}
	if req.ReplacementValue != nil {
		existing.ReplacementValue = *req.ReplacementValue
	}
	if req.WeightGrams != nil {
		existing.WeightGrams = req.WeightGrams
	}
	if req.Manufacturer != nil {
		existing.Manufacturer = *req.Manufacturer
	}
	if req.Model != nil {
		existing.Model = *req.Model
	}
	if req.Notes != nil {
		existing.Notes = *req.Notes
	}

	existing.UpdatedAt = time.Now()

	if err := s.equipmentRepo.Update(ctx, existing); err != nil {
		s.logger.Error().Err(err).
			Str("equipment_id", id.String()).
			Str("tenant_id", tenantID.String()).
			Msg("failed to update equipment")
		return nil, fmt.Errorf("update equipment: %w", err)
	}

	s.recordHistory(ctx, tenantID, id, userID, domain.ActionUpdated, map[string]string{
		"name": existing.Name,
	})

	s.logger.Info().
		Str("equipment_id", id.String()).
		Str("tenant_id", tenantID.String()).
		Msg("equipment updated")

	return existing, nil
}

// UpdateStatus changes the equipment status after validation.
func (s *EquipmentService) UpdateStatus(
	ctx context.Context,
	id uuid.UUID,
	tenantID uuid.UUID,
	userID uuid.UUID,
	req UpdateStatusRequest,
) error {
	// Validate status using domain logic.
	eq := &domain.Equipment{}
	if !eq.ValidateStatus(req.Status) {
		return domain.ErrInvalidStatus
	}

	if err := s.equipmentRepo.UpdateStatus(ctx, id, tenantID, req.Status); err != nil {
		s.logger.Error().Err(err).
			Str("equipment_id", id.String()).
			Str("tenant_id", tenantID.String()).
			Str("status", req.Status).
			Msg("failed to update equipment status")
		return fmt.Errorf("update equipment status: %w", err)
	}

	s.recordHistory(ctx, tenantID, id, userID, domain.ActionStatusChanged, map[string]string{
		"status": req.Status,
	})

	s.logger.Info().
		Str("equipment_id", id.String()).
		Str("tenant_id", tenantID.String()).
		Str("status", req.Status).
		Msg("equipment status updated")

	return nil
}

// UpdateCondition changes the equipment condition after validation.
func (s *EquipmentService) UpdateCondition(
	ctx context.Context,
	id uuid.UUID,
	tenantID uuid.UUID,
	userID uuid.UUID,
	req UpdateConditionRequest,
) error {
	eq := &domain.Equipment{}
	if !eq.ValidateCondition(req.Condition) {
		return domain.ErrInvalidCondition
	}

	if err := s.equipmentRepo.UpdateCondition(ctx, id, tenantID, req.Condition); err != nil {
		s.logger.Error().Err(err).
			Str("equipment_id", id.String()).
			Str("tenant_id", tenantID.String()).
			Str("condition", req.Condition).
			Msg("failed to update equipment condition")
		return fmt.Errorf("update equipment condition: %w", err)
	}

	s.recordHistory(ctx, tenantID, id, userID, domain.ActionConditionChanged, map[string]string{
		"condition": req.Condition,
	})

	s.logger.Info().
		Str("equipment_id", id.String()).
		Str("tenant_id", tenantID.String()).
		Str("condition", req.Condition).
		Msg("equipment condition updated")

	return nil
}

// AssignRFID assigns an RFID tag to an equipment item.
func (s *EquipmentService) AssignRFID(
	ctx context.Context,
	id uuid.UUID,
	tenantID uuid.UUID,
	userID uuid.UUID,
	req AssignRFIDRequest,
) error {
	if req.RFIDTag == "" {
		return fmt.Errorf("rfid_tag is required")
	}

	if err := s.equipmentRepo.AssignRFID(ctx, id, tenantID, req.RFIDTag); err != nil {
		s.logger.Error().Err(err).
			Str("equipment_id", id.String()).
			Str("tenant_id", tenantID.String()).
			Str("rfid_tag", req.RFIDTag).
			Msg("failed to assign rfid tag")
		return fmt.Errorf("assign rfid tag: %w", err)
	}

	s.recordHistory(ctx, tenantID, id, userID, domain.ActionRFIDAssigned, map[string]string{
		"rfid_tag": req.RFIDTag,
	})

	s.logger.Info().
		Str("equipment_id", id.String()).
		Str("tenant_id", tenantID.String()).
		Str("rfid_tag", req.RFIDTag).
		Msg("rfid tag assigned")

	return nil
}

// Deactivate soft-deletes an equipment item by marking it inactive.
func (s *EquipmentService) Deactivate(
	ctx context.Context,
	id uuid.UUID,
	tenantID uuid.UUID,
	userID uuid.UUID,
) error {
	if err := s.equipmentRepo.Deactivate(ctx, id, tenantID); err != nil {
		s.logger.Error().Err(err).
			Str("equipment_id", id.String()).
			Str("tenant_id", tenantID.String()).
			Msg("failed to deactivate equipment")
		return fmt.Errorf("deactivate equipment: %w", err)
	}

	s.recordHistory(ctx, tenantID, id, userID, domain.ActionStatusChanged, map[string]string{
		"status": "deactivated",
	})

	s.logger.Info().
		Str("equipment_id", id.String()).
		Str("tenant_id", tenantID.String()).
		Msg("equipment deactivated")

	return nil
}

// Search performs a full-text search across equipment for a tenant.
func (s *EquipmentService) Search(
	ctx context.Context,
	tenantID uuid.UUID,
	query string,
	page int,
	perPage int,
) ([]*domain.Equipment, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}

	items, total, err := s.equipmentRepo.Search(ctx, tenantID, query, page, perPage)
	if err != nil {
		s.logger.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Str("query", query).
			Msg("failed to search equipment")
		return nil, 0, fmt.Errorf("search equipment: %w", err)
	}
	return items, total, nil
}

// GetByBarcode retrieves equipment by barcode within a tenant scope.
func (s *EquipmentService) GetByBarcode(
	ctx context.Context,
	tenantID uuid.UUID,
	barcode string,
) (*domain.Equipment, error) {
	equipment, err := s.equipmentRepo.GetByBarcode(ctx, tenantID, barcode)
	if err != nil {
		s.logger.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Str("barcode", barcode).
			Msg("failed to get equipment by barcode")
		return nil, fmt.Errorf("get equipment by barcode: %w", err)
	}
	return equipment, nil
}

// GetByRFID retrieves equipment by RFID tag within a tenant scope.
func (s *EquipmentService) GetByRFID(
	ctx context.Context,
	tenantID uuid.UUID,
	rfidTag string,
) (*domain.Equipment, error) {
	equipment, err := s.equipmentRepo.GetByRFID(ctx, tenantID, rfidTag)
	if err != nil {
		s.logger.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Str("rfid_tag", rfidTag).
			Msg("failed to get equipment by rfid")
		return nil, fmt.Errorf("get equipment by rfid: %w", err)
	}
	return equipment, nil
}

// GetHistory returns paginated audit history for a piece of equipment.
func (s *EquipmentService) GetHistory(
	ctx context.Context,
	tenantID uuid.UUID,
	equipmentID uuid.UUID,
	page int,
	perPage int,
) ([]*domain.EquipmentHistory, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}

	entries, total, err := s.historyRepo.ListByEquipment(ctx, equipmentID, tenantID, page, perPage)
	if err != nil {
		s.logger.Error().Err(err).
			Str("equipment_id", equipmentID.String()).
			Msg("failed to get equipment history")
		return nil, 0, fmt.Errorf("get equipment history: %w", err)
	}
	return entries, total, nil
}

// ResolveByIdentifier looks up equipment by barcode, serial_number, or rfid_tag.
// It tries each field in order and returns the first match.
func (s *EquipmentService) ResolveByIdentifier(
	ctx context.Context,
	tenantID uuid.UUID,
	identifier string,
) (*domain.Equipment, error) {
	if identifier == "" {
		return nil, fmt.Errorf("identifier is required")
	}

	// Try barcode first.
	equipment, err := s.equipmentRepo.GetByBarcode(ctx, tenantID, identifier)
	if err == nil {
		return equipment, nil
	}

	// Try RFID tag.
	equipment, err = s.equipmentRepo.GetByRFID(ctx, tenantID, identifier)
	if err == nil {
		return equipment, nil
	}

	// Try serial number via the Search repo method (exact match via custom query).
	equipment, err = s.equipmentRepo.GetBySerialNumber(ctx, tenantID, identifier)
	if err == nil {
		return equipment, nil
	}

	s.logger.Warn().
		Str("tenant_id", tenantID.String()).
		Str("identifier", identifier).
		Msg("equipment not found by any identifier")

	return nil, fmt.Errorf("equipment not found")
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

// recordHistory writes an audit trail entry. Errors are logged but not
// propagated to avoid failing the primary operation.
func (s *EquipmentService) recordHistory(
	ctx context.Context,
	tenantID uuid.UUID,
	equipmentID uuid.UUID,
	userID uuid.UUID,
	action string,
	details map[string]string,
) {
	detailsJSON, err := json.Marshal(details)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to marshal history details")
		return
	}

	entry := &domain.EquipmentHistory{
		ID:          uuid.New(),
		TenantID:    tenantID,
		EquipmentID: equipmentID,
		Action:      action,
		UserID:      &userID,
		Details:     detailsJSON,
		CreatedAt:   time.Now(),
	}

	if err := s.historyRepo.Record(ctx, entry); err != nil {
		s.logger.Error().Err(err).
			Str("equipment_id", equipmentID.String()).
			Str("action", action).
			Msg("failed to record equipment history")
	}
}

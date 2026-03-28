package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/services/warehouse/internal/domain"
)

// ---------------------------------------------------------------------------
// Request DTOs
// ---------------------------------------------------------------------------

// CreateWarehouseRequest holds the data needed to create a new warehouse.
type CreateWarehouseRequest struct {
	Name                string `json:"name"`
	Code                string `json:"code,omitempty"`
	Address             string `json:"address,omitempty"`
	CapacityDescription string `json:"capacity_description,omitempty"`
}

// ---------------------------------------------------------------------------
// Service
// ---------------------------------------------------------------------------

// WarehouseService implements the application-level use cases for warehouses.
type WarehouseService struct {
	warehouseRepo domain.WarehouseRepository
	logger        zerolog.Logger
}

// NewWarehouseService constructs a new WarehouseService.
func NewWarehouseService(
	warehouseRepo domain.WarehouseRepository,
	logger zerolog.Logger,
) *WarehouseService {
	return &WarehouseService{
		warehouseRepo: warehouseRepo,
		logger:        logger.With().Str("service", "warehouse").Logger(),
	}
}

// Create validates the request and persists a new warehouse.
func (s *WarehouseService) Create(
	ctx context.Context,
	tenantID uuid.UUID,
	req CreateWarehouseRequest,
) (*domain.Warehouse, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("warehouse name is required")
	}

	warehouse := &domain.Warehouse{
		ID:                  uuid.New(),
		TenantID:            tenantID,
		Name:                req.Name,
		Code:                req.Code,
		Address:             req.Address,
		CapacityDescription: req.CapacityDescription,
		IsActive:            true,
		CreatedAt:           time.Now(),
	}

	if err := s.warehouseRepo.Create(ctx, warehouse); err != nil {
		s.logger.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Msg("failed to create warehouse")
		return nil, fmt.Errorf("create warehouse: %w", err)
	}

	s.logger.Info().
		Str("warehouse_id", warehouse.ID.String()).
		Str("tenant_id", tenantID.String()).
		Msg("warehouse created")

	return warehouse, nil
}

// List returns all warehouses for a tenant.
func (s *WarehouseService) List(
	ctx context.Context,
	tenantID uuid.UUID,
) ([]*domain.Warehouse, error) {
	warehouses, err := s.warehouseRepo.List(ctx, tenantID)
	if err != nil {
		s.logger.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Msg("failed to list warehouses")
		return nil, fmt.Errorf("list warehouses: %w", err)
	}
	return warehouses, nil
}

// GetByID retrieves a single warehouse by ID within a tenant scope.
func (s *WarehouseService) GetByID(
	ctx context.Context,
	id uuid.UUID,
	tenantID uuid.UUID,
) (*domain.Warehouse, error) {
	warehouse, err := s.warehouseRepo.GetByID(ctx, id, tenantID)
	if err != nil {
		s.logger.Error().Err(err).
			Str("warehouse_id", id.String()).
			Str("tenant_id", tenantID.String()).
			Msg("failed to get warehouse by id")
		return nil, fmt.Errorf("get warehouse by id: %w", err)
	}
	return warehouse, nil
}

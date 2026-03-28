package application

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/services/transport/internal/domain"
)

// ---------------------------------------------------------------------------
// Request DTOs
// ---------------------------------------------------------------------------

// CreateVehicleRequest holds the data needed to create a vehicle.
type CreateVehicleRequest struct {
	Name                string `json:"name"`
	LicensePlate        string `json:"license_plate,omitempty"`
	Type                string `json:"type,omitempty"`
	CapacityKg          *int   `json:"capacity_kg,omitempty"`
	CapacityDescription string `json:"capacity_description,omitempty"`
	Notes               string `json:"notes,omitempty"`
}

// UpdateVehicleRequest holds the data needed to update a vehicle.
type UpdateVehicleRequest struct {
	Name                string `json:"name"`
	LicensePlate        string `json:"license_plate,omitempty"`
	Type                string `json:"type,omitempty"`
	CapacityKg          *int   `json:"capacity_kg,omitempty"`
	CapacityDescription string `json:"capacity_description,omitempty"`
	IsActive            *bool  `json:"is_active,omitempty"`
	Notes               string `json:"notes,omitempty"`
}

// ---------------------------------------------------------------------------
// Service
// ---------------------------------------------------------------------------

// VehicleService implements the application-level use cases for vehicles.
type VehicleService struct {
	vehicleRepo domain.VehicleRepository
	logger      zerolog.Logger
}

// NewVehicleService constructs a new VehicleService.
func NewVehicleService(
	vehicleRepo domain.VehicleRepository,
	logger zerolog.Logger,
) *VehicleService {
	return &VehicleService{
		vehicleRepo: vehicleRepo,
		logger:      logger.With().Str("service", "vehicle").Logger(),
	}
}

// Create creates a new vehicle for a tenant.
func (s *VehicleService) Create(
	ctx context.Context,
	tenantID uuid.UUID,
	req CreateVehicleRequest,
) (*domain.Vehicle, error) {
	if req.Name == "" {
		return nil, domain.ErrMissingVehicleName
	}

	vType := req.Type
	if vType == "" {
		vType = domain.VehicleTypeVan
	}

	vehicle := &domain.Vehicle{
		ID:                  uuid.New(),
		TenantID:            tenantID,
		Name:                req.Name,
		LicensePlate:        req.LicensePlate,
		Type:                vType,
		CapacityKg:          req.CapacityKg,
		CapacityDescription: req.CapacityDescription,
		IsActive:            true,
		Notes:               req.Notes,
	}

	if err := s.vehicleRepo.Create(ctx, vehicle); err != nil {
		s.logger.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Msg("failed to create vehicle")
		return nil, fmt.Errorf("create vehicle: %w", err)
	}

	s.logger.Info().
		Str("vehicle_id", vehicle.ID.String()).
		Str("tenant_id", tenantID.String()).
		Msg("vehicle created")

	return vehicle, nil
}

// List returns all vehicles for a tenant.
func (s *VehicleService) List(
	ctx context.Context,
	tenantID uuid.UUID,
) ([]*domain.Vehicle, error) {
	vehicles, err := s.vehicleRepo.List(ctx, tenantID)
	if err != nil {
		s.logger.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Msg("failed to list vehicles")
		return nil, fmt.Errorf("list vehicles: %w", err)
	}
	return vehicles, nil
}

// GetByID returns a vehicle by its ID within a tenant scope.
func (s *VehicleService) GetByID(
	ctx context.Context,
	id uuid.UUID,
	tenantID uuid.UUID,
) (*domain.Vehicle, error) {
	vehicle, err := s.vehicleRepo.GetByID(ctx, id, tenantID)
	if err != nil {
		s.logger.Error().Err(err).
			Str("vehicle_id", id.String()).
			Str("tenant_id", tenantID.String()).
			Msg("failed to get vehicle")
		return nil, fmt.Errorf("get vehicle: %w", err)
	}
	return vehicle, nil
}

// Update updates an existing vehicle.
func (s *VehicleService) Update(
	ctx context.Context,
	id uuid.UUID,
	tenantID uuid.UUID,
	req UpdateVehicleRequest,
) (*domain.Vehicle, error) {
	if req.Name == "" {
		return nil, domain.ErrMissingVehicleName
	}

	vehicle, err := s.vehicleRepo.GetByID(ctx, id, tenantID)
	if err != nil {
		return nil, fmt.Errorf("update vehicle: %w", err)
	}

	vehicle.Name = req.Name
	vehicle.LicensePlate = req.LicensePlate
	if req.Type != "" {
		vehicle.Type = req.Type
	}
	vehicle.CapacityKg = req.CapacityKg
	vehicle.CapacityDescription = req.CapacityDescription
	if req.IsActive != nil {
		vehicle.IsActive = *req.IsActive
	}
	vehicle.Notes = req.Notes

	if err := s.vehicleRepo.Update(ctx, vehicle); err != nil {
		s.logger.Error().Err(err).
			Str("vehicle_id", id.String()).
			Str("tenant_id", tenantID.String()).
			Msg("failed to update vehicle")
		return nil, fmt.Errorf("update vehicle: %w", err)
	}

	s.logger.Info().
		Str("vehicle_id", id.String()).
		Str("tenant_id", tenantID.String()).
		Msg("vehicle updated")

	return vehicle, nil
}

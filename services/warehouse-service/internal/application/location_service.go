package application

import (
	"context"
	"fmt"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/warehouse-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/warehouse-service/internal/ports"
)

type LocationService struct {
	locationRepo ports.LocationRepository
	logger       logger.Logger
}

func NewLocationService(
	locationRepo ports.LocationRepository,
	logger logger.Logger,
) *LocationService {
	return &LocationService{
		locationRepo: locationRepo,
		logger:       logger,
	}
}

func (s *LocationService) CreateLocation(ctx context.Context, cmd CreateLocationCommand) (*LocationDTO, error) {
	if cmd.TenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}
	if cmd.Name == "" {
		return nil, domain.NewDomainError("NAME_REQUIRED", "name is required", nil)
	}

	locationID := fmt.Sprintf("loc_%d", hashLocationString(cmd.TenantID+cmd.Name+time.Now().String()))
	location := domain.NewLocation(locationID, cmd.TenantID, cmd.Name, cmd.Type)
	location.ParentID = cmd.ParentID
	location.Capacity = cmd.Capacity
	location.Barcode = cmd.Barcode
	location.SortOrder = cmd.SortOrder

	// Build path if parent exists
	if cmd.ParentID != nil {
		parent, err := s.locationRepo.GetByID(ctx, cmd.TenantID, *cmd.ParentID)
		if err != nil {
			return nil, domain.NewDomainError("INVALID_PARENT", "parent location not found", err)
		}
		location.UpdatePath(parent.Path)
	} else {
		location.UpdatePath("")
	}

	// Validate
	if err := location.Validate(); err != nil {
		return nil, domain.NewDomainError("VALIDATION_ERROR", err.Error(), nil)
	}

	// Persist
	if err := s.locationRepo.Create(ctx, location); err != nil {
		return nil, domain.NewDomainError("CREATE_ERROR", "failed to create location", err)
	}

	s.logger.Info("Location created", "id", location.ID, "name", cmd.Name)
	return LocationToDTO(location), nil
}

func (s *LocationService) GetLocation(ctx context.Context, tenantID, locationID string) (*LocationDTO, error) {
	if tenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	location, err := s.locationRepo.GetByID(ctx, tenantID, locationID)
	if err != nil {
		return nil, domain.NewDomainError("NOT_FOUND", "location not found", err)
	}

	return LocationToDTO(location), nil
}

func (s *LocationService) ListLocations(ctx context.Context, tenantID string, limit, offset int) (*PaginatedResult, error) {
	if tenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	locations, total, err := s.locationRepo.List(ctx, tenantID, limit, offset)
	if err != nil {
		return nil, domain.NewDomainError("QUERY_ERROR", "failed to list locations", err)
	}

	dtos := make([]*LocationDTO, len(locations))
	for i, loc := range locations {
		dtos[i] = LocationToDTO(loc)
	}

	return &PaginatedResult{
		Data:   dtos,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}, nil
}

func (s *LocationService) GetLocationTree(ctx context.Context, tenantID string) (*LocationTreeNode, error) {
	if tenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	// Get all root locations
	rootLocs, err := s.locationRepo.ListByParent(ctx, tenantID, nil)
	if err != nil {
		return nil, domain.NewDomainError("QUERY_ERROR", "failed to get locations", err)
	}

	if len(rootLocs) == 0 {
		return nil, nil
	}

	// Build tree (simplified: returns first root with children)
	return s.buildLocationTree(ctx, tenantID, rootLocs[0])
}

func (s *LocationService) buildLocationTree(ctx context.Context, tenantID string, location *domain.Location) (*LocationTreeNode, error) {
	node := &LocationTreeNode{
		LocationDTO: LocationToDTO(location),
		Children:    make([]*LocationTreeNode, 0),
	}

	children, err := s.locationRepo.ListByParent(ctx, tenantID, &location.ID)
	if err != nil {
		return node, nil
	}

	for _, child := range children {
		childNode, _ := s.buildLocationTree(ctx, tenantID, child)
		node.Children = append(node.Children, childNode)
	}

	return node, nil
}

func (s *LocationService) UpdateLocation(ctx context.Context, cmd UpdateLocationCommand) (*LocationDTO, error) {
	if cmd.TenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	location, err := s.locationRepo.GetByID(ctx, cmd.TenantID, cmd.ID)
	if err != nil {
		return nil, domain.NewDomainError("NOT_FOUND", "location not found", err)
	}

	location.Name = cmd.Name
	location.Capacity = cmd.Capacity
	location.Barcode = cmd.Barcode
	location.SortOrder = cmd.SortOrder
	location.UpdatedAt = time.Now()

	if err := location.Validate(); err != nil {
		return nil, domain.NewDomainError("VALIDATION_ERROR", err.Error(), nil)
	}

	if err := s.locationRepo.Update(ctx, location); err != nil {
		return nil, domain.NewDomainError("UPDATE_ERROR", "failed to update location", err)
	}

	s.logger.Info("Location updated", "id", cmd.ID)
	return LocationToDTO(location), nil
}

func (s *LocationService) DeleteLocation(ctx context.Context, tenantID, locationID string) error {
	if tenantID == "" {
		return domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	// Verify exists
	_, err := s.locationRepo.GetByID(ctx, tenantID, locationID)
	if err != nil {
		return domain.NewDomainError("NOT_FOUND", "location not found", err)
	}

	if err := s.locationRepo.Delete(ctx, tenantID, locationID); err != nil {
		return domain.NewDomainError("DELETE_ERROR", "failed to delete location", err)
	}

	s.logger.Info("Location deleted", "id", locationID)
	return nil
}

func (s *LocationService) GetLocationOccupancy(ctx context.Context, tenantID, locationID string) (map[string]interface{}, error) {
	if tenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	location, err := s.locationRepo.GetByID(ctx, tenantID, locationID)
	if err != nil {
		return nil, domain.NewDomainError("NOT_FOUND", "location not found", err)
	}

	occupancy := map[string]interface{}{
		"location_id":   location.ID,
		"name":          location.Name,
		"capacity":      location.Capacity,
		"current_count": location.CurrentCount,
		"available":     location.Capacity - location.CurrentCount,
		"utilization":   float64(location.CurrentCount) / float64(location.Capacity) * 100,
	}

	return occupancy, nil
}

func hashLocationString(s string) int64 {
	h := int64(5381)
	for _, c := range s {
		h = ((h << 5) + h) + int64(c)
	}
	return h & 0x7FFFFFFFFFFFFFFF
}

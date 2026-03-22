package application

import (
	"context"
	"fmt"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/warehouse-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/warehouse-service/internal/ports"
)

type WarehouseService struct {
	warehouseRepo ports.WarehouseRepository
	zoneRepo      ports.ZoneRepository
	rackRepo      ports.RackRepository
	bayRepo       ports.BayRepository
	logger        logger.Logger
}

func NewWarehouseService(
	warehouseRepo ports.WarehouseRepository,
	zoneRepo ports.ZoneRepository,
	rackRepo ports.RackRepository,
	bayRepo ports.BayRepository,
	logger logger.Logger,
) *WarehouseService {
	return &WarehouseService{
		warehouseRepo: warehouseRepo,
		zoneRepo:      zoneRepo,
		rackRepo:      rackRepo,
		bayRepo:       bayRepo,
		logger:        logger,
	}
}

// Warehouse operations
func (s *WarehouseService) CreateWarehouse(ctx context.Context, cmd CreateWarehouseCommand) (*WarehouseDTO, error) {
	if cmd.TenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}
	if cmd.Name == "" {
		return nil, domain.NewDomainError("NAME_REQUIRED", "name is required", nil)
	}
	if cmd.Code == "" {
		return nil, domain.NewDomainError("CODE_REQUIRED", "code is required", nil)
	}

	id := fmt.Sprintf("wh_%d", hashString(cmd.TenantID+cmd.Code+time.Now().String()))
	warehouse := domain.NewWarehouse(id, cmd.TenantID, cmd.Name, cmd.Code)
	warehouse.Address = cmd.Address
	warehouse.City = cmd.City
	warehouse.PostalCode = cmd.PostalCode
	warehouse.Country = cmd.Country
	warehouse.Latitude = cmd.Latitude
	warehouse.Longitude = cmd.Longitude
	warehouse.TotalCapacity = cmd.TotalCapacity

	if err := warehouse.Validate(); err != nil {
		return nil, domain.NewDomainError("VALIDATION_ERROR", err.Error(), nil)
	}

	if err := s.warehouseRepo.Create(ctx, warehouse); err != nil {
		return nil, domain.NewDomainError("CREATE_ERROR", "failed to create warehouse", err)
	}

	s.logger.Info("Warehouse created", "id", warehouse.ID, "name", cmd.Name)
	return WarehouseToDTO(warehouse), nil
}

func (s *WarehouseService) GetWarehouse(ctx context.Context, tenantID, warehouseID string) (*WarehouseDTO, error) {
	if tenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	warehouse, err := s.warehouseRepo.GetByID(ctx, tenantID, warehouseID)
	if err != nil {
		return nil, domain.NewDomainError("NOT_FOUND", "warehouse not found", err)
	}

	return WarehouseToDTO(warehouse), nil
}

func (s *WarehouseService) ListWarehouses(ctx context.Context, tenantID string, limit, offset int) (*PaginatedResult, error) {
	if tenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	warehouses, total, err := s.warehouseRepo.List(ctx, tenantID, limit, offset)
	if err != nil {
		return nil, domain.NewDomainError("QUERY_ERROR", "failed to list warehouses", err)
	}

	dtos := make([]*WarehouseDTO, len(warehouses))
	for i, wh := range warehouses {
		dtos[i] = WarehouseToDTO(wh)
	}

	return &PaginatedResult{
		Data:   dtos,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}, nil
}

func (s *WarehouseService) UpdateWarehouse(ctx context.Context, cmd UpdateWarehouseCommand) (*WarehouseDTO, error) {
	if cmd.TenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	warehouse, err := s.warehouseRepo.GetByID(ctx, cmd.TenantID, cmd.ID)
	if err != nil {
		return nil, domain.NewDomainError("NOT_FOUND", "warehouse not found", err)
	}

	warehouse.Name = cmd.Name
	warehouse.Address = cmd.Address
	warehouse.City = cmd.City
	warehouse.PostalCode = cmd.PostalCode
	warehouse.Country = cmd.Country
	warehouse.Latitude = cmd.Latitude
	warehouse.Longitude = cmd.Longitude
	warehouse.TotalCapacity = cmd.TotalCapacity
	warehouse.Status = domain.WarehouseStatus(cmd.Status)
	warehouse.UpdatedAt = time.Now()

	if err := warehouse.Validate(); err != nil {
		return nil, domain.NewDomainError("VALIDATION_ERROR", err.Error(), nil)
	}

	if err := s.warehouseRepo.Update(ctx, warehouse); err != nil {
		return nil, domain.NewDomainError("UPDATE_ERROR", "failed to update warehouse", err)
	}

	s.logger.Info("Warehouse updated", "id", cmd.ID)
	return WarehouseToDTO(warehouse), nil
}

// Zone operations
func (s *WarehouseService) CreateZone(ctx context.Context, cmd CreateZoneCommand) (*ZoneDTO, error) {
	if cmd.TenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	// Verify warehouse exists
	_, err := s.warehouseRepo.GetByID(ctx, cmd.TenantID, cmd.WarehouseID)
	if err != nil {
		return nil, domain.NewDomainError("WAREHOUSE_NOT_FOUND", "warehouse not found", err)
	}

	id := fmt.Sprintf("zn_%d", hashString(cmd.TenantID+cmd.WarehouseID+cmd.Code+time.Now().String()))
	zone := domain.NewZone(id, cmd.TenantID, cmd.WarehouseID, cmd.Name, cmd.Code)
	zone.ZoneType = cmd.ZoneType
	zone.Description = cmd.Description
	zone.TemperatureMin = cmd.TemperatureMin
	zone.TemperatureMax = cmd.TemperatureMax
	zone.HumidityMin = cmd.HumidityMin
	zone.HumidityMax = cmd.HumidityMax
	zone.Capacity = cmd.Capacity
	zone.SortOrder = cmd.SortOrder

	if err := zone.Validate(); err != nil {
		return nil, domain.NewDomainError("VALIDATION_ERROR", err.Error(), nil)
	}

	if err := s.zoneRepo.Create(ctx, zone); err != nil {
		return nil, domain.NewDomainError("CREATE_ERROR", "failed to create zone", err)
	}

	s.logger.Info("Zone created", "id", zone.ID, "warehouse", cmd.WarehouseID)
	return ZoneToDTO(zone), nil
}

func (s *WarehouseService) ListZonesInWarehouse(ctx context.Context, tenantID, warehouseID string, limit, offset int) (*PaginatedResult, error) {
	if tenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	zones, total, err := s.zoneRepo.ListByWarehouse(ctx, tenantID, warehouseID, limit, offset)
	if err != nil {
		return nil, domain.NewDomainError("QUERY_ERROR", "failed to list zones", err)
	}

	dtos := make([]*ZoneDTO, len(zones))
	for i, z := range zones {
		dtos[i] = ZoneToDTO(z)
	}

	return &PaginatedResult{
		Data:   dtos,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}, nil
}

// Rack operations
func (s *WarehouseService) CreateRack(ctx context.Context, cmd CreateRackCommand) (*RackDTO, error) {
	if cmd.TenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	// Verify zone exists
	zone, err := s.zoneRepo.GetByID(ctx, cmd.TenantID, cmd.ZoneID)
	if err != nil {
		return nil, domain.NewDomainError("ZONE_NOT_FOUND", "zone not found", err)
	}

	id := fmt.Sprintf("rk_%d", hashString(cmd.TenantID+cmd.ZoneID+cmd.Code+time.Now().String()))
	rack := domain.NewRack(id, cmd.TenantID, cmd.ZoneID, zone.WarehouseID, cmd.Name, cmd.Code)
	rack.RackType = cmd.RackType
	rack.Aisle = cmd.Aisle
	rack.RowNumber = cmd.RowNumber
	rack.ColumnNumber = cmd.ColumnNumber
	rack.Capacity = cmd.Capacity
	rack.Height = cmd.Height
	rack.Width = cmd.Width
	rack.Depth = cmd.Depth
	rack.WeightCapacity = cmd.WeightCapacity
	rack.SortOrder = cmd.SortOrder

	if err := rack.Validate(); err != nil {
		return nil, domain.NewDomainError("VALIDATION_ERROR", err.Error(), nil)
	}

	if err := s.rackRepo.Create(ctx, rack); err != nil {
		return nil, domain.NewDomainError("CREATE_ERROR", "failed to create rack", err)
	}

	s.logger.Info("Rack created", "id", rack.ID, "zone", cmd.ZoneID)
	return RackToDTO(rack), nil
}

func (s *WarehouseService) ListRacksInZone(ctx context.Context, tenantID, zoneID string, limit, offset int) (*PaginatedResult, error) {
	if tenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	racks, total, err := s.rackRepo.ListByZone(ctx, tenantID, zoneID, limit, offset)
	if err != nil {
		return nil, domain.NewDomainError("QUERY_ERROR", "failed to list racks", err)
	}

	dtos := make([]*RackDTO, len(racks))
	for i, r := range racks {
		dtos[i] = RackToDTO(r)
	}

	return &PaginatedResult{
		Data:   dtos,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}, nil
}

// Bay operations
func (s *WarehouseService) CreateBay(ctx context.Context, cmd CreateBayCommand) (*BayDTO, error) {
	if cmd.TenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	// Verify rack exists
	rack, err := s.rackRepo.GetByID(ctx, cmd.TenantID, cmd.RackID)
	if err != nil {
		return nil, domain.NewDomainError("RACK_NOT_FOUND", "rack not found", err)
	}

	id := fmt.Sprintf("by_%d", hashString(cmd.TenantID+cmd.RackID+cmd.Code+time.Now().String()))
	bay := domain.NewBay(id, cmd.TenantID, cmd.RackID, rack.ZoneID, rack.WarehouseID, cmd.Name, cmd.Code)
	bay.BayNumber = cmd.BayNumber
	bay.BayLevel = cmd.BayLevel
	bay.Capacity = cmd.Capacity
	bay.WeightCapacity = cmd.WeightCapacity
	bay.SortOrder = cmd.SortOrder

	if err := bay.Validate(); err != nil {
		return nil, domain.NewDomainError("VALIDATION_ERROR", err.Error(), nil)
	}

	if err := s.bayRepo.Create(ctx, bay); err != nil {
		return nil, domain.NewDomainError("CREATE_ERROR", "failed to create bay", err)
	}

	s.logger.Info("Bay created", "id", bay.ID, "rack", cmd.RackID)
	return BayToDTO(bay), nil
}

func (s *WarehouseService) ListBaysInRack(ctx context.Context, tenantID, rackID string, limit, offset int) (*PaginatedResult, error) {
	if tenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	bays, total, err := s.bayRepo.ListByRack(ctx, tenantID, rackID, limit, offset)
	if err != nil {
		return nil, domain.NewDomainError("QUERY_ERROR", "failed to list bays", err)
	}

	dtos := make([]*BayDTO, len(bays))
	for i, b := range bays {
		dtos[i] = BayToDTO(b)
	}

	return &PaginatedResult{
		Data:   dtos,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}, nil
}

// hashString is defined in movement_service.go

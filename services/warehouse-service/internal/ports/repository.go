package ports

import (
	"context"

	"github.com/jeckersberger/rentflow/services/warehouse-service/internal/domain"
)

type LocationRepository interface {
	Create(ctx context.Context, location *domain.Location) error
	GetByID(ctx context.Context, tenantID, id string) (*domain.Location, error)
	List(ctx context.Context, tenantID string, limit, offset int) ([]*domain.Location, int, error)
	ListByParent(ctx context.Context, tenantID string, parentID *string) ([]*domain.Location, error)
	Update(ctx context.Context, location *domain.Location) error
	Delete(ctx context.Context, tenantID, id string) error
}

type MovementRepository interface {
	Create(ctx context.Context, movement *domain.Movement) error
	GetByID(ctx context.Context, tenantID, id string) (*domain.Movement, error)
	List(ctx context.Context, tenantID string, query *MovementListQuery) ([]*domain.Movement, int, error)
	GetByEquipmentID(ctx context.Context, tenantID, equipmentID string, limit, offset int) ([]*domain.Movement, int, error)
}

type MovementListQuery struct {
	EquipmentID  *string
	FromLocation *string
	ToLocation   *string
	MovementType *string
	Limit        int
	Offset       int
}

type InventoryCheckRepository interface {
	Create(ctx context.Context, check *domain.InventoryCheck) error
	GetByID(ctx context.Context, tenantID, id string) (*domain.InventoryCheck, error)
	List(ctx context.Context, tenantID string, limit, offset int) ([]*domain.InventoryCheck, int, error)
	Update(ctx context.Context, check *domain.InventoryCheck) error
}

type WarehouseRepository interface {
	Create(ctx context.Context, warehouse *domain.Warehouse) error
	GetByID(ctx context.Context, tenantID, id string) (*domain.Warehouse, error)
	List(ctx context.Context, tenantID string, limit, offset int) ([]*domain.Warehouse, int, error)
	Update(ctx context.Context, warehouse *domain.Warehouse) error
	Delete(ctx context.Context, tenantID, id string) error
}

type ZoneRepository interface {
	Create(ctx context.Context, zone *domain.Zone) error
	GetByID(ctx context.Context, tenantID, id string) (*domain.Zone, error)
	ListByWarehouse(ctx context.Context, tenantID, warehouseID string, limit, offset int) ([]*domain.Zone, int, error)
	Update(ctx context.Context, zone *domain.Zone) error
	Delete(ctx context.Context, tenantID, id string) error
}

type RackRepository interface {
	Create(ctx context.Context, rack *domain.Rack) error
	GetByID(ctx context.Context, tenantID, id string) (*domain.Rack, error)
	ListByZone(ctx context.Context, tenantID, zoneID string, limit, offset int) ([]*domain.Rack, int, error)
	Update(ctx context.Context, rack *domain.Rack) error
	Delete(ctx context.Context, tenantID, id string) error
}

type BayRepository interface {
	Create(ctx context.Context, bay *domain.Bay) error
	GetByID(ctx context.Context, tenantID, id string) (*domain.Bay, error)
	ListByRack(ctx context.Context, tenantID, rackID string, limit, offset int) ([]*domain.Bay, int, error)
	Update(ctx context.Context, bay *domain.Bay) error
	Delete(ctx context.Context, tenantID, id string) error
}

type StockLocationRepository interface {
	Create(ctx context.Context, location *domain.StockLocation) error
	GetByID(ctx context.Context, tenantID, id string) (*domain.StockLocation, error)
	GetByLocationCode(ctx context.Context, tenantID, code string) (*domain.StockLocation, error)
	ListByBay(ctx context.Context, tenantID, bayID string, limit, offset int) ([]*domain.StockLocation, int, error)
	ListByEquipment(ctx context.Context, tenantID, equipmentID string) ([]*domain.StockLocation, error)
	Update(ctx context.Context, location *domain.StockLocation) error
	Delete(ctx context.Context, tenantID, id string) error
}

package domain

import (
	"context"

	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// Repository ports (driven / secondary adapters)
// ---------------------------------------------------------------------------

// WarehouseRepository defines persistence operations for Warehouse aggregates.
type WarehouseRepository interface {
	Create(ctx context.Context, warehouse *Warehouse) error
	GetByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*Warehouse, error)
	List(ctx context.Context, tenantID uuid.UUID) ([]*Warehouse, error)
}

// ZoneRepository defines persistence operations for Zone aggregates.
type ZoneRepository interface {
	Create(ctx context.Context, zone *Zone) error
	List(ctx context.Context, tenantID uuid.UUID) ([]*Zone, error)
	GetByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*Zone, error)
	Update(ctx context.Context, zone *Zone) error
}

// RackRepository defines persistence operations for Rack aggregates.
type RackRepository interface {
	Create(ctx context.Context, rack *Rack) error
	List(ctx context.Context, tenantID uuid.UUID) ([]*Rack, error)
}

// StockLocationRepository defines persistence operations for StockLocation aggregates.
type StockLocationRepository interface {
	Create(ctx context.Context, location *StockLocation) error
	GetByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*StockLocation, error)
	List(ctx context.Context, tenantID uuid.UUID, page int, perPage int) ([]*StockLocation, int64, error)
}

// MovementRepository defines persistence operations for Movement aggregates.
type MovementRepository interface {
	Create(ctx context.Context, movement *Movement) error
	List(ctx context.Context, tenantID uuid.UUID, page int, perPage int) ([]*Movement, int64, error)
}

// InventoryCheckRepository defines persistence operations for InventoryCheck aggregates.
type InventoryCheckRepository interface {
	Create(ctx context.Context, check *InventoryCheck) error
	GetByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*InventoryCheck, error)
	List(ctx context.Context, tenantID uuid.UUID, page int, perPage int) ([]*InventoryCheck, int64, error)
	Complete(ctx context.Context, check *InventoryCheck) error
	CreateItem(ctx context.Context, item *InventoryCheckItem) error
	ScanItem(ctx context.Context, checkID uuid.UUID, equipmentID uuid.UUID) error
	GetItems(ctx context.Context, checkID uuid.UUID) ([]*InventoryCheckItem, error)
	CountExpected(ctx context.Context, checkID uuid.UUID) (int, error)
	CountFound(ctx context.Context, checkID uuid.UUID) (int, error)
}

package domain

import (
	"time"

	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// Inventory Check Status constants
// ---------------------------------------------------------------------------

const (
	CheckStatusInProgress = "in_progress"
	CheckStatusCompleted  = "completed"
)

// ---------------------------------------------------------------------------
// Domain models
// ---------------------------------------------------------------------------

// Warehouse represents a physical warehouse / storage location.
type Warehouse struct {
	ID                  uuid.UUID `json:"id"`
	TenantID            uuid.UUID `json:"tenant_id"`
	Name                string    `json:"name"`
	Code                string    `json:"code"`
	Address             string    `json:"address"`
	CapacityDescription string    `json:"capacity_description"`
	IsActive            bool      `json:"is_active"`
	CreatedAt           time.Time `json:"created_at"`
}

// Zone represents a logical zone within a warehouse.
type Zone struct {
	ID                uuid.UUID `json:"id"`
	WarehouseID       uuid.UUID `json:"warehouse_id"`
	TenantID          uuid.UUID `json:"tenant_id"`
	Name              string    `json:"name"`
	Code              string    `json:"code"`
	ClimateControlled bool      `json:"climate_controlled"`
	MaxWeightKg       *int      `json:"max_weight_kg,omitempty"`
	Notes             string    `json:"notes"`
	CreatedAt         time.Time `json:"created_at"`
}

// Rack represents a physical rack within a zone.
type Rack struct {
	ID          uuid.UUID `json:"id"`
	ZoneID      uuid.UUID `json:"zone_id"`
	TenantID    uuid.UUID `json:"tenant_id"`
	Name        string    `json:"name"`
	Code        string    `json:"code"`
	Levels      int       `json:"levels"`
	BaysPerLevel int      `json:"bays_per_level"`
	MaxWeightKg *int      `json:"max_weight_kg,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// StockLocation represents an individual storage slot (e.g. rack shelf position).
type StockLocation struct {
	ID          uuid.UUID  `json:"id"`
	RackID      *uuid.UUID `json:"rack_id,omitempty"`
	ZoneID      *uuid.UUID `json:"zone_id,omitempty"`
	TenantID    uuid.UUID  `json:"tenant_id"`
	Code        string     `json:"code"`
	Barcode     string     `json:"barcode"`
	Level       *int       `json:"level,omitempty"`
	Bay         *int       `json:"bay,omitempty"`
	MaxWeightKg *int       `json:"max_weight_kg,omitempty"`
	IsActive    bool       `json:"is_active"`
	CreatedAt   time.Time  `json:"created_at"`
}

// Movement represents a stock movement of equipment between locations.
type Movement struct {
	ID             uuid.UUID  `json:"id"`
	TenantID       uuid.UUID  `json:"tenant_id"`
	EquipmentID    uuid.UUID  `json:"equipment_id"`
	FromLocationID *uuid.UUID `json:"from_location_id,omitempty"`
	ToLocationID   *uuid.UUID `json:"to_location_id,omitempty"`
	Quantity       int        `json:"quantity"`
	Reason         string     `json:"reason"`
	UserID         *uuid.UUID `json:"user_id,omitempty"`
	Notes          string     `json:"notes"`
	CreatedAt      time.Time  `json:"created_at"`
}

// InventoryCheck represents a stock-take / inventory check session.
type InventoryCheck struct {
	ID               uuid.UUID  `json:"id"`
	TenantID         uuid.UUID  `json:"tenant_id"`
	ZoneID           *uuid.UUID `json:"zone_id,omitempty"`
	Status           string     `json:"status"`
	StartedBy        *uuid.UUID `json:"started_by,omitempty"`
	StartedAt        time.Time  `json:"started_at"`
	CompletedAt      *time.Time `json:"completed_at,omitempty"`
	ExpectedCount    int        `json:"expected_count"`
	ActualCount      int        `json:"actual_count"`
	DiscrepancyCount int        `json:"discrepancy_count"`
	Notes            string     `json:"notes"`
}

// InventoryCheckItem represents a single equipment entry within an inventory check.
type InventoryCheckItem struct {
	ID          uuid.UUID  `json:"id"`
	CheckID     uuid.UUID  `json:"check_id"`
	EquipmentID uuid.UUID  `json:"equipment_id"`
	Expected    bool       `json:"expected"`
	Found       bool       `json:"found"`
	ScannedAt   *time.Time `json:"scanned_at,omitempty"`
	Notes       string     `json:"notes"`
}

// InventoryCheckResult is the combined result of an inventory check with its items.
type InventoryCheckResult struct {
	Check *InventoryCheck       `json:"check"`
	Items []*InventoryCheckItem `json:"items"`
}

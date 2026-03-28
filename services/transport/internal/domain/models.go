package domain

import (
	"time"

	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// Transport Order type / status constants
// ---------------------------------------------------------------------------

const (
	OrderTypeDelivery = "delivery"
	OrderTypePickup   = "pickup"
	OrderTypeTransfer = "transfer"
)

var validOrderTypes = map[string]bool{
	OrderTypeDelivery: true,
	OrderTypePickup:   true,
	OrderTypeTransfer: true,
}

// ValidOrderType checks whether the given string is a valid order type.
func ValidOrderType(t string) bool {
	return validOrderTypes[t]
}

const (
	StatusPlanned    = "planned"
	StatusInTransit  = "in_transit"
	StatusCompleted  = "completed"
	StatusCancelled  = "cancelled"
)

var validStatuses = map[string]bool{
	StatusPlanned:   true,
	StatusInTransit: true,
	StatusCompleted: true,
	StatusCancelled: true,
}

// ValidStatus checks whether the given string is a valid order status.
func ValidStatus(s string) bool {
	return validStatuses[s]
}

// ---------------------------------------------------------------------------
// Vehicle type constants
// ---------------------------------------------------------------------------

const (
	VehicleTypeVan    = "van"
	VehicleTypeTruck  = "truck"
	VehicleTypeTrailer = "trailer"
	VehicleTypeCar    = "car"
)

// ---------------------------------------------------------------------------
// Domain models
// ---------------------------------------------------------------------------

// Vehicle represents a transport vehicle owned by the tenant.
type Vehicle struct {
	ID                  uuid.UUID  `json:"id"`
	TenantID            uuid.UUID  `json:"tenant_id"`
	Name                string     `json:"name"`
	LicensePlate        string     `json:"license_plate,omitempty"`
	Type                string     `json:"type"`
	CapacityKg          *int       `json:"capacity_kg,omitempty"`
	CapacityDescription string     `json:"capacity_description,omitempty"`
	IsActive            bool       `json:"is_active"`
	Notes               string     `json:"notes,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
}

// TransportOrder represents a transport order for equipment delivery/pickup.
type TransportOrder struct {
	ID              uuid.UUID  `json:"id"`
	TenantID        uuid.UUID  `json:"tenant_id"`
	ProjectID       *uuid.UUID `json:"project_id,omitempty"`
	VehicleID       *uuid.UUID `json:"vehicle_id,omitempty"`
	DriverID        *uuid.UUID `json:"driver_id,omitempty"`
	Type            string     `json:"type"`
	Status          string     `json:"status"`
	PickupAddress   string     `json:"pickup_address,omitempty"`
	DeliveryAddress string     `json:"delivery_address,omitempty"`
	ScheduledAt     *time.Time `json:"scheduled_at,omitempty"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
	Notes           string     `json:"notes,omitempty"`
	CreatedBy       *uuid.UUID `json:"created_by,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// TransportItem represents a single equipment item in a transport order.
type TransportItem struct {
	ID          uuid.UUID `json:"id"`
	OrderID     uuid.UUID `json:"order_id"`
	EquipmentID uuid.UUID `json:"equipment_id"`
	Quantity    int       `json:"quantity"`
	Notes       string    `json:"notes,omitempty"`
}

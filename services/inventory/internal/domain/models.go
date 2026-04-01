package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// Equipment Status constants
// ---------------------------------------------------------------------------

const (
	StatusAvailable     = "available"
	StatusReserved      = "reserved"
	StatusCheckedOut    = "checked_out"
	StatusInMaintenance = "in_maintenance"
	StatusDamaged       = "damaged"
	StatusRetired       = "retired"
)

// validStatuses is the set of all allowed equipment statuses.
var validStatuses = map[string]bool{
	StatusAvailable:     true,
	StatusReserved:      true,
	StatusCheckedOut:    true,
	StatusInMaintenance: true,
	StatusDamaged:       true,
	StatusRetired:       true,
}

// ---------------------------------------------------------------------------
// Equipment Condition constants
// ---------------------------------------------------------------------------

const (
	ConditionOperational    = "operational"
	ConditionGood           = "good"
	ConditionFair           = "fair"
	ConditionDamaged        = "damaged"
	ConditionDecommissioned = "decommissioned"
)

// validConditions is the set of all allowed equipment conditions.
var validConditions = map[string]bool{
	ConditionOperational:    true,
	ConditionGood:           true,
	ConditionFair:           true,
	ConditionDamaged:        true,
	ConditionDecommissioned: true,
}

// ---------------------------------------------------------------------------
// History Action constants
// ---------------------------------------------------------------------------

const (
	ActionCreated          = "created"
	ActionUpdated          = "updated"
	ActionStatusChanged    = "status_changed"
	ActionConditionChanged = "condition_changed"
	ActionRFIDAssigned     = "rfid_assigned"
	ActionCheckedOut       = "checked_out"
	ActionCheckedIn        = "checked_in"
)

// ---------------------------------------------------------------------------
// Domain models
// ---------------------------------------------------------------------------

// Category represents an equipment category with optional parent for nesting.
type Category struct {
	ID        uuid.UUID  `json:"id"`
	TenantID  uuid.UUID  `json:"tenant_id"`
	Name      string     `json:"name"`
	ParentID  *uuid.UUID `json:"parent_id,omitempty"`
	Icon      string     `json:"icon"`
	Color     string     `json:"color"`
	SortOrder int        `json:"sort_order"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// Equipment represents a single rentable piece of equipment.
type Equipment struct {
	ID              uuid.UUID  `json:"id"`
	TenantID        uuid.UUID  `json:"tenant_id"`
	CategoryID      *uuid.UUID `json:"category_id,omitempty"`
	EquipmentTypeID *uuid.UUID `json:"equipment_type_id,omitempty"`

	Name         string `json:"name"`
	Description  string `json:"description"`
	SKU          string `json:"sku"`
	Barcode      string `json:"barcode"`
	QRCode       string `json:"qr_code"`
	SerialNumber string `json:"serial_number"`
	RFIDTag      string `json:"rfid_tag"`

	Status    string `json:"status"`
	Condition string `json:"condition"`

	QuantityTotal     int `json:"quantity_total"`
	QuantityAvailable int `json:"quantity_available"`

	RentalPriceDay   int64 `json:"rental_price_day"`
	RentalPriceWeek  int64 `json:"rental_price_week"`
	ReplacementValue int64 `json:"replacement_value"`

	WeightGrams *int `json:"weight_grams,omitempty"`
	WidthMM     *int `json:"width_mm,omitempty"`
	HeightMM    *int `json:"height_mm,omitempty"`
	DepthMM     *int `json:"depth_mm,omitempty"`

	LocationID *uuid.UUID `json:"location_id,omitempty"`

	PurchaseDate  *time.Time `json:"purchase_date,omitempty"`
	PurchasePrice int64      `json:"purchase_price"`

	Manufacturer string `json:"manufacturer"`
	Model        string `json:"model"`
	ImageURL     string `json:"image_url"`

	CustomFields json.RawMessage `json:"custom_fields,omitempty"`
	Notes        string          `json:"notes"`

	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// IsAvailable returns true when the equipment is in available status and active.
func (e *Equipment) IsAvailable() bool {
	return e.Status == StatusAvailable && e.IsActive
}

// ValidateStatus checks whether the given string is a valid equipment status.
func (e *Equipment) ValidateStatus(s string) bool {
	return validStatuses[s]
}

// ValidateCondition checks whether the given string is a valid equipment condition.
func (e *Equipment) ValidateCondition(c string) bool {
	return validConditions[c]
}

// EquipmentType defines a template / type for equipment with default pricing.
type EquipmentType struct {
	ID         uuid.UUID  `json:"id"`
	TenantID   uuid.UUID  `json:"tenant_id"`
	CategoryID *uuid.UUID `json:"category_id,omitempty"`

	Name        string `json:"name"`
	Description string `json:"description"`

	DefaultRentalPriceDay   int64 `json:"default_rental_price_day"`
	DefaultRentalPriceWeek  int64 `json:"default_rental_price_week"`
	DefaultReplacementValue int64 `json:"default_replacement_value"`

	Specifications json.RawMessage `json:"specifications,omitempty"`

	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Flightcase represents a physical flight case that can hold multiple equipment items.
type Flightcase struct {
	ID          uuid.UUID `json:"id"`
	TenantID    uuid.UUID `json:"tenant_id"`
	Name        string    `json:"name"`
	Barcode     string    `json:"barcode"`
	QRCode      string    `json:"qr_code"`
	Description string    `json:"description"`
	WeightGrams *int      `json:"weight_grams,omitempty"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// FlightcaseItem links a piece of equipment to a flightcase.
type FlightcaseItem struct {
	ID          uuid.UUID `json:"id"`
	FlightcaseID uuid.UUID `json:"flightcase_id"`
	EquipmentID uuid.UUID `json:"equipment_id"`
	Quantity    int       `json:"quantity"`
	SortOrder   int       `json:"sort_order"`
	CreatedAt   time.Time `json:"created_at"`
}

// EquipmentHistory records an auditable change to a piece of equipment.
type EquipmentHistory struct {
	ID          uuid.UUID       `json:"id"`
	TenantID    uuid.UUID       `json:"tenant_id"`
	EquipmentID uuid.UUID       `json:"equipment_id"`
	Action      string          `json:"action"`
	UserID      *uuid.UUID      `json:"user_id,omitempty"`
	Details     json.RawMessage `json:"details,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
}

// ---------------------------------------------------------------------------
// Price Rule domain
// ---------------------------------------------------------------------------

// ---------------------------------------------------------------------------
// Availability domain
// ---------------------------------------------------------------------------

// AvailabilityFilter holds query parameters for the availability overview endpoint.
type AvailabilityFilter struct {
	From       time.Time  `json:"from"`
	To         time.Time  `json:"to"`
	CategoryID *uuid.UUID `json:"category_id,omitempty"`
}

// TypeAvailabilitySummary aggregates availability for a single equipment type.
type TypeAvailabilitySummary struct {
	EquipmentTypeID uuid.UUID `json:"equipment_type_id"`
	Name            string    `json:"name"`
	TotalQuantity   int       `json:"total_quantity"`
	Reserved        int       `json:"reserved"`
	InMaintenance   int       `json:"in_maintenance"`
	Available       int       `json:"available"`
	AvailabilityPct int       `json:"availability_pct"`
}

// EquipmentBooking represents a period during which a specific equipment item
// is reserved or checked out.
type EquipmentBooking struct {
	EquipmentID uuid.UUID `json:"equipment_id"`
	Status      string    `json:"status"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	ProjectID   *uuid.UUID `json:"project_id,omitempty"`
	ProjectName string    `json:"project_name,omitempty"`
}

// ---------------------------------------------------------------------------
// Price Rule domain
// ---------------------------------------------------------------------------

// PriceRule defines pricing tiers, quantity discounts, and seasonal surcharges
// for a piece of equipment or a category.
type PriceRule struct {
	ID                  uuid.UUID  `json:"id"`
	TenantID            uuid.UUID  `json:"tenant_id"`
	EquipmentID         *uuid.UUID `json:"equipment_id,omitempty"`
	CategoryID          *uuid.UUID `json:"category_id,omitempty"`
	BasePriceDay        int64      `json:"base_price_day"`
	BasePriceWeek       *int64     `json:"base_price_week,omitempty"`
	Tier2FromDays       *int       `json:"tier2_from_days,omitempty"`
	Tier2PriceDay       *int64     `json:"tier2_price_day,omitempty"`
	Tier3FromDays       *int       `json:"tier3_from_days,omitempty"`
	Tier3PriceDay       *int64     `json:"tier3_price_day,omitempty"`
	QtyDiscountThreshold *int      `json:"qty_discount_threshold,omitempty"`
	QtyDiscountPct      int        `json:"qty_discount_pct"`
	SeasonStart         *time.Time `json:"season_start,omitempty"`
	SeasonEnd           *time.Time `json:"season_end,omitempty"`
	SeasonSurchargePct  int        `json:"season_surcharge_pct"`
	IsActive            bool       `json:"is_active"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

package domain

import (
	"fmt"
	"time"
)

type WarehouseStatus string

const (
	WarehouseActive   WarehouseStatus = "active"
	WarehouseInactive WarehouseStatus = "inactive"
	WarehouseMaintenance WarehouseStatus = "maintenance"
)

// Warehouse represents a top-level warehouse facility
type Warehouse struct {
	ID               string
	TenantID         string
	Name             string
	Code             string
	Address          string
	City             string
	PostalCode       string
	Country          string
	Latitude         *float64
	Longitude        *float64
	TotalCapacity    int
	CurrentOccupancy int
	Status           WarehouseStatus
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// Zone represents an area within a warehouse (e.g., Receiving, Shipping)
type Zone struct {
	ID          string
	TenantID    string
	WarehouseID string
	Name        string
	Code        string
	ZoneType    string
	Description string
	TemperatureMin *float64
	TemperatureMax *float64
	HumidityMin    *float64
	HumidityMax    *float64
	Capacity    int
	Occupancy   int
	SortOrder   int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Rack represents a storage structure within a zone
type Rack struct {
	ID             string
	TenantID       string
	ZoneID         string
	WarehouseID    string
	Name           string
	Code           string
	RackType       string
	Aisle          string
	RowNumber      int
	ColumnNumber   int
	Capacity       int
	Occupancy      int
	Height         *float64
	Width          *float64
	Depth          *float64
	WeightCapacity *float64
	SortOrder      int
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// Bay represents a horizontal division on a rack (e.g., shelf level)
type Bay struct {
	ID             string
	TenantID       string
	RackID         string
	ZoneID         string
	WarehouseID    string
	Name           string
	Code           string
	BayNumber      int
	BayLevel       int
	Capacity       int
	Occupancy      int
	WeightCapacity *float64
	SortOrder      int
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// StockLocation represents an individual storage slot
type StockLocation struct {
	ID             string
	TenantID       string
	WarehouseID    string
	ZoneID         string
	RackID         string
	BayID          string
	LocationCode   string
	Barcode        string
	QRCodeLabel    string
	LocationType   string
	Capacity       int
	Occupancy      int
	WeightCapacity *float64
	CurrentWeight  *float64
	EquipmentID    *string
	IsAvailable    bool
	AccessLevel    string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// NewWarehouse creates a new warehouse domain object
func NewWarehouse(id, tenantID, name, code string) *Warehouse {
	return &Warehouse{
		ID:        id,
		TenantID:  tenantID,
		Name:      name,
		Code:      code,
		Status:    WarehouseActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// Validate validates warehouse
func (w *Warehouse) Validate() error {
	if w.TenantID == "" {
		return fmt.Errorf("tenant ID is required")
	}
	if w.Name == "" {
		return fmt.Errorf("name is required")
	}
	if w.Code == "" {
		return fmt.Errorf("code is required")
	}
	return nil
}

// NewZone creates a new zone domain object
func NewZone(id, tenantID, warehouseID, name, code string) *Zone {
	return &Zone{
		ID:          id,
		TenantID:    tenantID,
		WarehouseID: warehouseID,
		Name:        name,
		Code:        code,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// Validate validates zone
func (z *Zone) Validate() error {
	if z.TenantID == "" {
		return fmt.Errorf("tenant ID is required")
	}
	if z.WarehouseID == "" {
		return fmt.Errorf("warehouse ID is required")
	}
	if z.Name == "" {
		return fmt.Errorf("name is required")
	}
	if z.Code == "" {
		return fmt.Errorf("code is required")
	}
	return nil
}

// CanAddCapacity checks if capacity is available
func (z *Zone) CanAddCapacity(count int) bool {
	if z.Capacity == 0 {
		return true // unlimited capacity
	}
	return z.Occupancy+count <= z.Capacity
}

// NewRack creates a new rack domain object
func NewRack(id, tenantID, zoneID, warehouseID, name, code string) *Rack {
	return &Rack{
		ID:          id,
		TenantID:    tenantID,
		ZoneID:      zoneID,
		WarehouseID: warehouseID,
		Name:        name,
		Code:        code,
		RackType:    "standard",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// Validate validates rack
func (r *Rack) Validate() error {
	if r.TenantID == "" {
		return fmt.Errorf("tenant ID is required")
	}
	if r.ZoneID == "" {
		return fmt.Errorf("zone ID is required")
	}
	if r.Name == "" {
		return fmt.Errorf("name is required")
	}
	if r.Code == "" {
		return fmt.Errorf("code is required")
	}
	return nil
}

// CanAddCapacity checks if capacity is available
func (r *Rack) CanAddCapacity(count int) bool {
	if r.Capacity == 0 {
		return true
	}
	return r.Occupancy+count <= r.Capacity
}

// NewBay creates a new bay domain object
func NewBay(id, tenantID, rackID, zoneID, warehouseID, name, code string) *Bay {
	return &Bay{
		ID:          id,
		TenantID:    tenantID,
		RackID:      rackID,
		ZoneID:      zoneID,
		WarehouseID: warehouseID,
		Name:        name,
		Code:        code,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// Validate validates bay
func (b *Bay) Validate() error {
	if b.TenantID == "" {
		return fmt.Errorf("tenant ID is required")
	}
	if b.RackID == "" {
		return fmt.Errorf("rack ID is required")
	}
	if b.Name == "" {
		return fmt.Errorf("name is required")
	}
	if b.Code == "" {
		return fmt.Errorf("code is required")
	}
	return nil
}

// CanAddCapacity checks if capacity is available
func (b *Bay) CanAddCapacity(count int) bool {
	if b.Capacity == 0 {
		return true
	}
	return b.Occupancy+count <= b.Capacity
}

// NewStockLocation creates a new stock location domain object
func NewStockLocation(id, tenantID, warehouseID, zoneID, rackID, bayID, locationCode string) *StockLocation {
	return &StockLocation{
		ID:           id,
		TenantID:     tenantID,
		WarehouseID:  warehouseID,
		ZoneID:       zoneID,
		RackID:       rackID,
		BayID:        bayID,
		LocationCode: locationCode,
		Capacity:     1,
		IsAvailable:  true,
		AccessLevel:  "standard",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

// Validate validates stock location
func (sl *StockLocation) Validate() error {
	if sl.TenantID == "" {
		return fmt.Errorf("tenant ID is required")
	}
	if sl.WarehouseID == "" {
		return fmt.Errorf("warehouse ID is required")
	}
	if sl.LocationCode == "" {
		return fmt.Errorf("location code is required")
	}
	if sl.Capacity <= 0 {
		return fmt.Errorf("capacity must be greater than 0")
	}
	return nil
}

// CanAddEquipment checks if the location can hold more equipment
func (sl *StockLocation) CanAddEquipment() bool {
	return sl.Occupancy < sl.Capacity && sl.IsAvailable
}

// AddEquipment adds equipment to the location
func (sl *StockLocation) AddEquipment(equipmentID string, count int) error {
	if !sl.CanAddEquipment() {
		return fmt.Errorf("location is full or unavailable")
	}
	if sl.Occupancy+count > sl.Capacity {
		return fmt.Errorf("insufficient capacity: can fit %d more items", sl.Capacity-sl.Occupancy)
	}
	sl.EquipmentID = &equipmentID
	sl.Occupancy += count
	sl.UpdatedAt = time.Now()
	return nil
}

// RemoveEquipment removes equipment from location
func (sl *StockLocation) RemoveEquipment(count int) error {
	if sl.Occupancy < count {
		return fmt.Errorf("cannot remove more items than available")
	}
	sl.Occupancy -= count
	if sl.Occupancy == 0 {
		sl.EquipmentID = nil
	}
	sl.UpdatedAt = time.Now()
	return nil
}

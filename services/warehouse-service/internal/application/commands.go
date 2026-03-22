package application

import "github.com/jeckersberger/rentflow/services/warehouse-service/internal/domain"

// Warehouse Commands
type CreateWarehouseCommand struct {
	TenantID      string
	Name          string
	Code          string
	Address       string
	City          string
	PostalCode    string
	Country       string
	Latitude      *float64
	Longitude     *float64
	TotalCapacity int
}

type UpdateWarehouseCommand struct {
	ID            string
	TenantID      string
	Name          string
	Address       string
	City          string
	PostalCode    string
	Country       string
	Latitude      *float64
	Longitude     *float64
	TotalCapacity int
	Status        string
}

// Zone Commands
type CreateZoneCommand struct {
	TenantID       string
	WarehouseID    string
	Name           string
	Code           string
	ZoneType       string
	Description    string
	TemperatureMin *float64
	TemperatureMax *float64
	HumidityMin    *float64
	HumidityMax    *float64
	Capacity       int
	SortOrder      int
}

type UpdateZoneCommand struct {
	ID             string
	TenantID       string
	Name           string
	ZoneType       string
	Description    string
	TemperatureMin *float64
	TemperatureMax *float64
	HumidityMin    *float64
	HumidityMax    *float64
	Capacity       int
	SortOrder      int
}

// Rack Commands
type CreateRackCommand struct {
	TenantID       string
	ZoneID         string
	Name           string
	Code           string
	RackType       string
	Aisle          string
	RowNumber      int
	ColumnNumber   int
	Capacity       int
	Height         *float64
	Width          *float64
	Depth          *float64
	WeightCapacity *float64
	SortOrder      int
}

type UpdateRackCommand struct {
	ID             string
	TenantID       string
	Name           string
	RackType       string
	Aisle          string
	RowNumber      int
	ColumnNumber   int
	Capacity       int
	Height         *float64
	Width          *float64
	Depth          *float64
	WeightCapacity *float64
	SortOrder      int
}

// Bay Commands
type CreateBayCommand struct {
	TenantID       string
	RackID         string
	Name           string
	Code           string
	BayNumber      int
	BayLevel       int
	Capacity       int
	WeightCapacity *float64
	SortOrder      int
}

type UpdateBayCommand struct {
	ID             string
	TenantID       string
	Name           string
	BayNumber      int
	BayLevel       int
	Capacity       int
	WeightCapacity *float64
	SortOrder      int
}

// StockLocation Commands
type CreateStockLocationCommand struct {
	TenantID       string
	WarehouseID    string
	ZoneID         string
	RackID         string
	BayID          string
	LocationCode   string
	LocationType   string
	Capacity       int
	WeightCapacity *float64
	AccessLevel    string
}

// Inventory Commands
type StartInventoryCheckCommand struct {
	TenantID    string
	WarehouseID *string
	ZoneID      *string
	CheckType   string
	Name        string
	Description string
}

type ScanInventoryItemCommand struct {
	TenantID    string
	CheckID     string
	EquipmentID string
	LocationID  *string
	Notes       string
}

// Legacy ScanInventoryItemCommand - kept for backwards compatibility

type CompleteInventoryCheckCommand struct {
	TenantID    string
	CheckID     string
	CompletedBy string
}

// Legacy Location Commands (kept for backwards compatibility)
type CreateLocationCommand struct {
	TenantID  string
	Name      string
	Type      domain.LocationType
	ParentID  *string
	Capacity  int
	Barcode   string
	SortOrder int
}

type UpdateLocationCommand struct {
	ID        string
	TenantID  string
	Name      string
	Capacity  int
	Barcode   string
	SortOrder int
}

type DeleteLocationCommand struct {
	ID       string
	TenantID string
}

type RecordMovementCommand struct {
	TenantID       string
	EquipmentID    string
	FromLocationID *string
	ToLocationID   string
	MovementType   domain.MovementType
	Quantity       int
	Reason         string
	UserID         string
	ProjectID      *string
}

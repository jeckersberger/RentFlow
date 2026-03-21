package application

import "github.com/jeckersberger/rentflow/services/warehouse-service/internal/domain"

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

type StartInventoryCheckCommand struct {
	TenantID   string
	Name       string
	LocationID *string
}

type ScanInventoryItemCommand struct {
	TenantID       string
	CheckID        string
	EquipmentID    string
	Notes          string
}

type CompleteInventoryCheckCommand struct {
	TenantID string
	CheckID  string
}

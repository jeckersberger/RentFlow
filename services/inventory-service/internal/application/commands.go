package application

import (
	"time"

	"github.com/jeckersberger/rentflow/services/inventory-service/internal/domain"
)

// Equipment Commands

type CreateEquipmentCommand struct {
	TenantID        string
	Name            string
	Description     string
	CategoryID      string
	SKU             string
	SerialNumber    string
	Barcode         string
	PurchaseDate    *time.Time
	PurchasePrice   float64
	RentalPriceDay  float64
	RentalPriceWeek float64
	Weight          float64
	Dimensions      DimensionsDTO
	LocationID      string
	Tags            []string
	CustomFields    map[string]string
	CreatedByUserID string
}

type UpdateEquipmentCommand struct {
	ID              string
	TenantID        string
	Name            string
	Description     string
	CategoryID      string
	PurchaseDate    *time.Time
	PurchasePrice   float64
	RentalPriceDay  float64
	RentalPriceWeek float64
	Weight          float64
	Dimensions      DimensionsDTO
	Tags            []string
	CustomFields    map[string]string
}

type ChangeStatusCommand struct {
	ID       string
	TenantID string
	Status   domain.EquipmentStatus
	Reason   string
}

type UpdateConditionCommand struct {
	ID        string
	TenantID  string
	Condition domain.EquipmentCondition
}

type SetLocationCommand struct {
	ID         string
	TenantID   string
	LocationID string
}

type AddImageCommand struct {
	ID       string
	TenantID string
	ImageRef string
}

type DeleteEquipmentCommand struct {
	ID       string
	TenantID string
}

type CheckOutCommand struct {
	ID        string
	TenantID  string
	ProjectID string
	UserID    string
}

type CheckInCommand struct {
	ID       string
	TenantID string
	UserID   string
}

// Category Commands

type CreateCategoryCommand struct {
	TenantID        string
	Name            string
	ParentID        *string
	Icon            string
	Color           string
	SortOrder       int
	CreatedByUserID string
}

type UpdateCategoryCommand struct {
	ID        string
	TenantID  string
	Name      string
	Icon      string
	Color     string
	SortOrder int
}

type DeleteCategoryCommand struct {
	ID       string
	TenantID string
}

// Flightcase Commands

type CreateFlightcaseCommand struct {
	TenantID        string
	Name            string
	Description     string
	Barcode         string
	Weight          float64
	CreatedByUserID string
}

type UpdateFlightcaseCommand struct {
	ID          string
	TenantID    string
	Name        string
	Description string
	Weight      float64
}

type AddFlightcaseItemCommand struct {
	FlightcaseID string
	TenantID     string
	EquipmentID  string
	Quantity     int
}

type RemoveFlightcaseItemCommand struct {
	FlightcaseID string
	TenantID     string
	EquipmentID  string
	Quantity     int
}

type DeleteFlightcaseCommand struct {
	ID       string
	TenantID string
}

package domain

import (
	"context"

	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// Filter
// ---------------------------------------------------------------------------

// EquipmentFilter holds optional criteria for listing equipment.
type EquipmentFilter struct {
	CategoryID *uuid.UUID `json:"category_id,omitempty"`
	Status     *string    `json:"status,omitempty"`
	Condition  *string    `json:"condition,omitempty"`
	Search     *string    `json:"search,omitempty"`
	Page       int        `json:"page"`
	PerPage    int        `json:"per_page"`
}

// ---------------------------------------------------------------------------
// Repository ports (driven / secondary adapters)
// ---------------------------------------------------------------------------

// EquipmentRepository defines persistence operations for Equipment aggregates.
type EquipmentRepository interface {
	Create(ctx context.Context, equipment *Equipment) error
	GetByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*Equipment, error)
	GetByBarcode(ctx context.Context, tenantID uuid.UUID, barcode string) (*Equipment, error)
	GetByRFID(ctx context.Context, tenantID uuid.UUID, rfidTag string) (*Equipment, error)
	List(ctx context.Context, tenantID uuid.UUID, filter EquipmentFilter) ([]*Equipment, int64, error)
	Update(ctx context.Context, equipment *Equipment) error
	UpdateStatus(ctx context.Context, id uuid.UUID, tenantID uuid.UUID, status string) error
	UpdateCondition(ctx context.Context, id uuid.UUID, tenantID uuid.UUID, condition string) error
	AssignRFID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID, rfidTag string) error
	Deactivate(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error
	Search(ctx context.Context, tenantID uuid.UUID, query string, page int, perPage int) ([]*Equipment, int64, error)
}

// CategoryRepository defines persistence operations for Category aggregates.
type CategoryRepository interface {
	Create(ctx context.Context, category *Category) error
	GetByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*Category, error)
	List(ctx context.Context, tenantID uuid.UUID) ([]*Category, error)
	Update(ctx context.Context, category *Category) error
	Delete(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error
	HasChildren(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (bool, error)
	HasEquipment(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (bool, error)
}

// EquipmentTypeRepository defines persistence operations for EquipmentType aggregates.
type EquipmentTypeRepository interface {
	Create(ctx context.Context, equipmentType *EquipmentType) error
	GetByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*EquipmentType, error)
	List(ctx context.Context, tenantID uuid.UUID) ([]*EquipmentType, error)
	Update(ctx context.Context, equipmentType *EquipmentType) error
	Delete(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error
}

// FlightcaseRepository defines persistence operations for Flightcase aggregates.
type FlightcaseRepository interface {
	Create(ctx context.Context, flightcase *Flightcase) error
	GetByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*Flightcase, error)
	List(ctx context.Context, tenantID uuid.UUID, page int, perPage int) ([]*Flightcase, int64, error)
	Update(ctx context.Context, flightcase *Flightcase) error
	Delete(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error
	AddItem(ctx context.Context, item *FlightcaseItem) error
	RemoveItem(ctx context.Context, id uuid.UUID) error
	GetItems(ctx context.Context, flightcaseID uuid.UUID) ([]*FlightcaseItem, error)
}

// EquipmentHistoryRepository defines persistence operations for equipment audit history.
type EquipmentHistoryRepository interface {
	Record(ctx context.Context, entry *EquipmentHistory) error
	ListByEquipment(ctx context.Context, equipmentID uuid.UUID, page int, perPage int) ([]*EquipmentHistory, int64, error)
}

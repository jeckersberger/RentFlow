package domain

import (
	"context"

	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// Filters
// ---------------------------------------------------------------------------

// TransportOrderFilter holds optional criteria for listing transport orders.
type TransportOrderFilter struct {
	ProjectID *uuid.UUID `json:"project_id,omitempty"`
	Status    *string    `json:"status,omitempty"`
	Page      int        `json:"page"`
	PerPage   int        `json:"per_page"`
}

// ---------------------------------------------------------------------------
// Repository ports (driven / secondary adapters)
// ---------------------------------------------------------------------------

// VehicleRepository defines persistence operations for Vehicle aggregates.
type VehicleRepository interface {
	Create(ctx context.Context, vehicle *Vehicle) error
	List(ctx context.Context, tenantID uuid.UUID) ([]*Vehicle, error)
	GetByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*Vehicle, error)
	Update(ctx context.Context, vehicle *Vehicle) error
	Delete(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error
}

// TransportOrderRepository defines persistence operations for TransportOrder aggregates.
type TransportOrderRepository interface {
	Create(ctx context.Context, order *TransportOrder) error
	List(ctx context.Context, tenantID uuid.UUID, filter TransportOrderFilter) ([]*TransportOrder, int64, error)
	GetByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*TransportOrder, error)
	Update(ctx context.Context, order *TransportOrder) error
}

// TransportItemRepository defines persistence operations for TransportItem aggregates.
type TransportItemRepository interface {
	Create(ctx context.Context, item *TransportItem) error
	ListByOrder(ctx context.Context, orderID uuid.UUID) ([]*TransportItem, error)
}

// TransportCostRepository defines persistence operations for TransportCost aggregates.
type TransportCostRepository interface {
	Create(ctx context.Context, cost *TransportCost) error
	ListByOrder(ctx context.Context, orderID uuid.UUID) ([]*TransportCost, error)
}

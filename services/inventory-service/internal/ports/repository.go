package ports

import (
	"context"

	"github.com/jeckersberger/rentflow/services/inventory-service/internal/domain"
)

type EquipmentRepository interface {
	Create(ctx context.Context, equipment *domain.Equipment) error
	Update(ctx context.Context, equipment *domain.Equipment) error
	GetByID(ctx context.Context, tenantID, equipmentID string) (*domain.Equipment, error)
	GetByBarcode(ctx context.Context, tenantID, barcode string) (*domain.Equipment, error)
	GetByRfidTag(ctx context.Context, tenantID, rfidTag string) (*domain.Equipment, error)
	List(ctx context.Context, query *EquipmentListQuery) (*EquipmentListResult, error)
	Delete(ctx context.Context, tenantID, equipmentID string) error
	Search(ctx context.Context, tenantID, term string, limit, offset int) (*EquipmentListResult, error)
}

type CategoryRepository interface {
	Create(ctx context.Context, category *domain.Category) error
	Update(ctx context.Context, category *domain.Category) error
	GetByID(ctx context.Context, tenantID, categoryID string) (*domain.Category, error)
	List(ctx context.Context, tenantID string) ([]*domain.Category, error)
	Delete(ctx context.Context, tenantID, categoryID string) error
}

type FlightcaseRepository interface {
	Create(ctx context.Context, flightcase *domain.Flightcase) error
	Update(ctx context.Context, flightcase *domain.Flightcase) error
	GetByID(ctx context.Context, tenantID, flightcaseID string) (*domain.Flightcase, error)
	GetByBarcode(ctx context.Context, tenantID, barcode string) (*domain.Flightcase, error)
	List(ctx context.Context, tenantID string, limit, offset int) (*FlightcaseListResult, error)
	Delete(ctx context.Context, tenantID, flightcaseID string) error
}

type EquipmentListQuery struct {
	TenantID   string
	Status     *domain.EquipmentStatus
	CategoryID *string
	LocationID *string
	SearchTerm *string
	Limit      int
	Offset     int
}

type EquipmentListResult struct {
	Items  []*domain.Equipment
	Total  int64
	Limit  int
	Offset int
}

type FlightcaseListResult struct {
	Items  []*domain.Flightcase
	Total  int64
	Limit  int
	Offset int
}

type EquipmentHistoryRepository interface {
	Create(ctx context.Context, history *domain.EquipmentHistory) error
	GetByEquipmentID(ctx context.Context, tenantID, equipmentID string, limit, offset int) ([]*domain.EquipmentHistory, int64, error)
}

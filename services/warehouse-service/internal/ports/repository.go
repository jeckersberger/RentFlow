package ports

import (
	"context"

	"github.com/jeckersberger/rentflow/services/warehouse-service/internal/domain"
)

type LocationRepository interface {
	Create(ctx context.Context, location *domain.Location) error
	GetByID(ctx context.Context, tenantID, id string) (*domain.Location, error)
	List(ctx context.Context, tenantID string, limit, offset int) ([]*domain.Location, int, error)
	ListByParent(ctx context.Context, tenantID string, parentID *string) ([]*domain.Location, error)
	Update(ctx context.Context, location *domain.Location) error
	Delete(ctx context.Context, tenantID, id string) error
}

type MovementRepository interface {
	Create(ctx context.Context, movement *domain.Movement) error
	GetByID(ctx context.Context, tenantID, id string) (*domain.Movement, error)
	List(ctx context.Context, tenantID string, query *MovementListQuery) ([]*domain.Movement, int, error)
	GetByEquipmentID(ctx context.Context, tenantID, equipmentID string, limit, offset int) ([]*domain.Movement, int, error)
}

type MovementListQuery struct {
	EquipmentID  *string
	FromLocation *string
	ToLocation   *string
	MovementType *string
	Limit        int
	Offset       int
}

type InventoryCheckRepository interface {
	Create(ctx context.Context, check *domain.InventoryCheck) error
	GetByID(ctx context.Context, tenantID, id string) (*domain.InventoryCheck, error)
	List(ctx context.Context, tenantID string, limit, offset int) ([]*domain.InventoryCheck, int, error)
	Update(ctx context.Context, check *domain.InventoryCheck) error
}

package ports

import (
	"context"
	"time"

	"github.com/jeckersberger/rentflow/services/transport-service/internal/domain"
)

type VehicleRepository interface {
	Create(ctx context.Context, vehicle *domain.Vehicle) error
	GetByID(ctx context.Context, tenantID, id string) (*domain.Vehicle, error)
	ListByTenant(ctx context.Context, tenantID string) ([]*domain.Vehicle, error)
	Update(ctx context.Context, vehicle *domain.Vehicle) error
	Delete(ctx context.Context, tenantID, id string) error
}

type TourRepository interface {
	Create(ctx context.Context, tour *domain.Tour) error
	GetByID(ctx context.Context, tenantID, id string) (*domain.Tour, error)
	ListByTenant(ctx context.Context, tenantID string) ([]*domain.Tour, error)
	ListByDate(ctx context.Context, tenantID string, date time.Time) ([]*domain.Tour, error)
	Update(ctx context.Context, tour *domain.Tour) error
	Delete(ctx context.Context, tenantID, id string) error
}

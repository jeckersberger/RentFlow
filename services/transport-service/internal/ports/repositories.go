package ports

import (
	"context"

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
	Update(ctx context.Context, tour *domain.Tour) error
	Delete(ctx context.Context, tenantID, id string) error
}

type TourEquipmentRepository interface {
	Create(ctx context.Context, equipment *domain.TourEquipment) error
	GetByID(ctx context.Context, id string) (*domain.TourEquipment, error)
	ListByTour(ctx context.Context, tourID string) ([]*domain.TourEquipment, error)
	DeleteByTourAndEquipment(ctx context.Context, tourID, equipmentID string) error
	GetCapacityByTour(ctx context.Context, tourID string) (totalWeightKg, totalVolumeM3 float64, err error)
}

type DriverLogRepository interface {
	Create(ctx context.Context, log *domain.DriverLog) error
	GetByID(ctx context.Context, id string) (*domain.DriverLog, error)
	ListByTour(ctx context.Context, tourID string) ([]*domain.DriverLog, error)
	Update(ctx context.Context, log *domain.DriverLog) error
}

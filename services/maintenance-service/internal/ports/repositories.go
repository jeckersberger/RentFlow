package ports

import (
	"context"

	"github.com/jeckersberger/rentflow/services/maintenance-service/internal/domain"
)

type MaintenanceRecordRepository interface {
	Create(ctx context.Context, record *domain.MaintenanceRecord) error
	GetByID(ctx context.Context, tenantID, id string) (*domain.MaintenanceRecord, error)
	ListByTenant(ctx context.Context, tenantID string) ([]*domain.MaintenanceRecord, error)
	ListByEquipment(ctx context.Context, tenantID, equipmentID string) ([]*domain.MaintenanceRecord, error)
	ListOverdue(ctx context.Context, tenantID string) ([]*domain.MaintenanceRecord, error)
	Update(ctx context.Context, record *domain.MaintenanceRecord) error
	Delete(ctx context.Context, tenantID, id string) error
}

type MaintenanceScheduleRepository interface {
	Create(ctx context.Context, schedule *domain.MaintenanceSchedule) error
	GetByID(ctx context.Context, tenantID, id string) (*domain.MaintenanceSchedule, error)
	GetByEquipment(ctx context.Context, tenantID, equipmentID string) (*domain.MaintenanceSchedule, error)
	ListByTenant(ctx context.Context, tenantID string) ([]*domain.MaintenanceSchedule, error)
	Update(ctx context.Context, schedule *domain.MaintenanceSchedule) error
	Delete(ctx context.Context, tenantID, id string) error
}

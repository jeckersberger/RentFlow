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

type MaintenancePlanRepository interface {
	Create(ctx context.Context, plan *domain.MaintenancePlan) error
	GetByID(ctx context.Context, tenantID, id string) (*domain.MaintenancePlan, error)
	GetByEquipment(ctx context.Context, tenantID, equipmentID string) ([]*domain.MaintenancePlan, error)
	ListByTenant(ctx context.Context, tenantID string) ([]*domain.MaintenancePlan, error)
	Update(ctx context.Context, plan *domain.MaintenancePlan) error
	Delete(ctx context.Context, tenantID, id string) error
	ListDuePlans(ctx context.Context, tenantID string) ([]*domain.MaintenancePlan, error)
}

type MaintenanceTaskRepository interface {
	Create(ctx context.Context, task *domain.MaintenanceTask) error
	GetByID(ctx context.Context, tenantID, id string) (*domain.MaintenanceTask, error)
	ListByTenant(ctx context.Context, tenantID string) ([]*domain.MaintenanceTask, error)
	ListByPlan(ctx context.Context, tenantID, planID string) ([]*domain.MaintenanceTask, error)
	ListByStatus(ctx context.Context, tenantID string, status domain.TaskStatus) ([]*domain.MaintenanceTask, error)
	Update(ctx context.Context, task *domain.MaintenanceTask) error
	Delete(ctx context.Context, tenantID, id string) error
	ListDueTasks(ctx context.Context, tenantID string) ([]*domain.MaintenanceTask, error)
	ListOverdueTasks(ctx context.Context, tenantID string) ([]*domain.MaintenanceTask, error)
}

type ChecklistRepository interface {
	Create(ctx context.Context, checklist *domain.Checklist) error
	GetByID(ctx context.Context, tenantID, id string) (*domain.Checklist, error)
	ListByTenant(ctx context.Context, tenantID string) ([]*domain.Checklist, error)
	Update(ctx context.Context, checklist *domain.Checklist) error
	Delete(ctx context.Context, tenantID, id string) error
}

type ElectricalTestRepository interface {
	Create(ctx context.Context, test *domain.ElectricalTest) error
	GetByID(ctx context.Context, tenantID, id string) (*domain.ElectricalTest, error)
	ListByEquipment(ctx context.Context, tenantID, equipmentID string) ([]*domain.ElectricalTest, error)
	ListByTenant(ctx context.Context, tenantID string) ([]*domain.ElectricalTest, error)
	ListByTask(ctx context.Context, tenantID, taskID string) ([]*domain.ElectricalTest, error)
	Update(ctx context.Context, test *domain.ElectricalTest) error
	Delete(ctx context.Context, tenantID, id string) error
}

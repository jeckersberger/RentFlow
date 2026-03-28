package domain

import (
	"context"

	"github.com/google/uuid"
)

type ScheduleRepository interface {
	Create(ctx context.Context, schedule *MaintenanceSchedule) error
	GetByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*MaintenanceSchedule, error)
	List(ctx context.Context, tenantID uuid.UUID, filter ScheduleFilter) ([]*MaintenanceSchedule, int64, error)
	Update(ctx context.Context, schedule *MaintenanceSchedule) error
}

type TaskRepository interface {
	Create(ctx context.Context, task *MaintenanceTask) error
	GetByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*MaintenanceTask, error)
	List(ctx context.Context, tenantID uuid.UUID, filter TaskFilter) ([]*MaintenanceTask, int64, error)
	Update(ctx context.Context, task *MaintenanceTask) error
}

type LogRepository interface {
	Create(ctx context.Context, log *MaintenanceLog) error
	ListByTask(ctx context.Context, taskID uuid.UUID, tenantID uuid.UUID, filter LogFilter) ([]*MaintenanceLog, int64, error)
}

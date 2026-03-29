package domain

import (
	"context"

	"github.com/google/uuid"
)

type ReportDefinitionRepository interface {
	Create(ctx context.Context, def *ReportDefinition) error
	GetByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*ReportDefinition, error)
	List(ctx context.Context, tenantID uuid.UUID, filter DefinitionFilter) ([]*ReportDefinition, int64, error)
	Update(ctx context.Context, def *ReportDefinition) error
}

type ReportSnapshotRepository interface {
	Create(ctx context.Context, snap *ReportSnapshot) error
	GetByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*ReportSnapshot, error)
	List(ctx context.Context, tenantID uuid.UUID, filter SnapshotFilter) ([]*ReportSnapshot, int64, error)
}

type DashboardWidgetRepository interface {
	Create(ctx context.Context, widget *DashboardWidget) error
	GetByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*DashboardWidget, error)
	List(ctx context.Context, tenantID uuid.UUID, filter WidgetFilter) ([]*DashboardWidget, int64, error)
	Update(ctx context.Context, widget *DashboardWidget) error
	Delete(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error
}

// KPISnapshotRepository defines persistence operations for KPISnapshot aggregates.
type KPISnapshotRepository interface {
	Upsert(ctx context.Context, snap *KPISnapshot) error
	GetLatest(ctx context.Context, tenantID uuid.UUID) (*KPISnapshot, error)
	GetHistory(ctx context.Context, tenantID uuid.UUID, from, to string) ([]*KPISnapshot, error)
}

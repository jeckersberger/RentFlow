package ports

import (
	"context"
	"github.com/jeckersberger/rentflow/services/reporting-service/internal/domain"
)

type ReportRepository interface {
	Create(ctx context.Context, report *domain.Report) error
	GetByID(ctx context.Context, tenantID, id string) (*domain.Report, error)
	ListByTenant(ctx context.Context, tenantID string) ([]*domain.Report, error)
	Update(ctx context.Context, report *domain.Report) error
	Delete(ctx context.Context, tenantID, id string) error
}

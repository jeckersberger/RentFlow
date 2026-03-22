package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/jeckersberger/rentflow/services/audit-service/internal/domain"
)

type AuditRepository interface {
	Create(ctx context.Context, entry *domain.AuditEntry) (*domain.AuditEntry, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.AuditEntry, error)
	GetLastChecksum(ctx context.Context, tenantID uuid.UUID) (*string, error)
	ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*domain.AuditEntry, error)
	ListByEntity(ctx context.Context, tenantID uuid.UUID, entityType string, entityID uuid.UUID) ([]*domain.AuditEntry, error)
	ListByUser(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID) ([]*domain.AuditEntry, error)
	ListByDateRange(ctx context.Context, tenantID uuid.UUID, from, to interface{}) ([]*domain.AuditEntry, error)
	Pseudonymize(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID) error
	GetTodayCount(ctx context.Context, tenantID uuid.UUID) (int, error)
	GetAll(ctx context.Context, tenantID uuid.UUID) ([]*domain.AuditEntry, error)
}

type ExportRepository interface {
	Create(ctx context.Context, export *domain.AuditExport) (*domain.AuditExport, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.AuditExport, error)
	ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*domain.AuditExport, error)
	Update(ctx context.Context, export *domain.AuditExport) (*domain.AuditExport, error)
}

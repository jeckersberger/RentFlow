package ports

import (
	"context"
	"time"
	"github.com/jeckersberger/rentflow/services/audit-service/internal/domain"
)

type AuditEntryRepository interface {
	Create(ctx context.Context, entry *domain.AuditEntry) error
	GetByID(ctx context.Context, tenantID, id string) (*domain.AuditEntry, error)
	ListByTenant(ctx context.Context, tenantID string, fromTime, toTime time.Time) ([]*domain.AuditEntry, error)
	ListByEntity(ctx context.Context, tenantID, entityType, entityID string) ([]*domain.AuditEntry, error)
	ListByUser(ctx context.Context, tenantID, userID string) ([]*domain.AuditEntry, error)
	GetLastEntry(ctx context.Context, tenantID string) (*domain.AuditEntry, error)
}

type IntegrityCheckRepository interface {
	Create(ctx context.Context, check *domain.IntegrityCheck) error
	GetLatest(ctx context.Context, tenantID string) (*domain.IntegrityCheck, error)
}

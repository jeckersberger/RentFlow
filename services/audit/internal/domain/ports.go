package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type AuditLogRepository interface {
	Create(ctx context.Context, log *AuditLog) error
	GetByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*AuditLog, error)
	List(ctx context.Context, tenantID uuid.UUID, filter AuditLogFilter) ([]*AuditLog, int64, error)
	CreateWithHash(ctx context.Context, log *AuditLog) error
	GetLastChecksum(ctx context.Context, tenantID uuid.UUID) (string, error)
	ListBySequenceRange(ctx context.Context, tenantID uuid.UUID, fromSeq, toSeq int64) ([]*AuditLog, error)
	ListByDateRange(ctx context.Context, tenantID uuid.UUID, fromDate, toDate time.Time) ([]*AuditLog, error)
}

type AuditPolicyRepository interface {
	Create(ctx context.Context, policy *AuditPolicy) error
	GetByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*AuditPolicy, error)
	List(ctx context.Context, tenantID uuid.UUID, filter AuditPolicyFilter) ([]*AuditPolicy, int64, error)
	Update(ctx context.Context, policy *AuditPolicy) error
}

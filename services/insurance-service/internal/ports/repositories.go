package ports

import (
	"context"
	"time"

	"github.com/jeckersberger/rentflow/services/insurance-service/internal/domain"
)

type PolicyRepository interface {
	Create(ctx context.Context, policy *domain.Policy) error
	GetByID(ctx context.Context, tenantID, id string) (*domain.Policy, error)
	ListByTenant(ctx context.Context, tenantID string) ([]*domain.Policy, error)
	ListActive(ctx context.Context, tenantID string) ([]*domain.Policy, error)
	ListExpiring(ctx context.Context, tenantID string, days int) ([]*domain.Policy, error)
	Update(ctx context.Context, policy *domain.Policy) error
	Delete(ctx context.Context, tenantID, id string) error
}

type ClaimRepository interface {
	Create(ctx context.Context, claim *domain.Claim) error
	GetByID(ctx context.Context, tenantID, id string) (*domain.Claim, error)
	ListByTenant(ctx context.Context, tenantID string) ([]*domain.Claim, error)
	ListByEquipment(ctx context.Context, tenantID, equipmentID string) ([]*domain.Claim, error)
	ListByPolicy(ctx context.Context, tenantID, policyID string) ([]*domain.Claim, error)
	Update(ctx context.Context, claim *domain.Claim) error
	Delete(ctx context.Context, tenantID, id string) error
}

type RiskAssessmentRepository interface {
	Create(ctx context.Context, assessment *domain.RiskAssessment) error
	GetByEquipment(ctx context.Context, tenantID, equipmentID string) (*domain.RiskAssessment, error)
	ListByTenant(ctx context.Context, tenantID string) ([]*domain.RiskAssessment, error)
	Update(ctx context.Context, assessment *domain.RiskAssessment) error
	Delete(ctx context.Context, tenantID, equipmentID string) error
}

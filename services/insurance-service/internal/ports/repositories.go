package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/jeckersberger/rentflow/services/insurance-service/internal/domain"
)

// PolicyRepository defines the interface for policy persistence
type PolicyRepository interface {
	Create(ctx context.Context, policy *domain.Policy) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Policy, error)
	GetByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*domain.Policy, error)
	GetActiveByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*domain.Policy, error)
	Update(ctx context.Context, policy *domain.Policy) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetExpiringPolicies(ctx context.Context, daysUntilExpiry int) ([]*domain.Policy, error)
}

// ClaimRepository defines the interface for claim persistence
type ClaimRepository interface {
	Create(ctx context.Context, claim *domain.Claim) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Claim, error)
	GetByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*domain.Claim, error)
	GetByPolicyID(ctx context.Context, policyID uuid.UUID) ([]*domain.Claim, error)
	Update(ctx context.Context, claim *domain.Claim) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetClaimsByStatus(ctx context.Context, tenantID uuid.UUID, status domain.ClaimStatus) ([]*domain.Claim, error)
}

// ClaimItemRepository defines the interface for claim item persistence
type ClaimItemRepository interface {
	Create(ctx context.Context, item *domain.ClaimItem) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.ClaimItem, error)
	GetByClaimID(ctx context.Context, claimID uuid.UUID) ([]*domain.ClaimItem, error)
	Update(ctx context.Context, item *domain.ClaimItem) error
	Delete(ctx context.Context, id uuid.UUID) error
}

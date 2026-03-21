package ports

import (
	"context"

	"github.com/jeckersberger/rentflow/services/ai-service/internal/domain"
)

type AIRequestRepository interface {
	Create(ctx context.Context, req *domain.AIRequest) error
	GetByID(ctx context.Context, tenantID, id string) (*domain.AIRequest, error)
	ListByTenant(ctx context.Context, tenantID string) ([]*domain.AIRequest, error)
	Update(ctx context.Context, req *domain.AIRequest) error
}

type AnonymizationRuleRepository interface {
	Create(ctx context.Context, rule *domain.AnonymizationRule) error
	GetByID(ctx context.Context, tenantID, id string) (*domain.AnonymizationRule, error)
	ListByTenant(ctx context.Context, tenantID string) ([]*domain.AnonymizationRule, error)
	ListByType(ctx context.Context, tenantID string, ruleType domain.AnonymizationRuleType) ([]*domain.AnonymizationRule, error)
	Update(ctx context.Context, rule *domain.AnonymizationRule) error
	Delete(ctx context.Context, tenantID, id string) error
}

type AnonymizationMappingRepository interface {
	Create(ctx context.Context, mapping *domain.AnonymizationMapping) error
	GetByID(ctx context.Context, id string) (*domain.AnonymizationMapping, error)
	GetByOriginal(ctx context.Context, tenantID, original string) (*domain.AnonymizationMapping, error)
	Delete(ctx context.Context, tenantID, mappingID string) error
}

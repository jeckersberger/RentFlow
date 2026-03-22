package ports

import (
	"context"

	"github.com/jeckersberger/rentflow/services/ai-service/internal/domain"
)

// AIRequestRepository defines the interface for AI request data access
type AIRequestRepository interface {
	// FindByID retrieves an AI request by ID
	FindByID(ctx context.Context, id string) (*domain.AIRequest, error)

	// List retrieves AI requests in a tenant with pagination
	List(ctx context.Context, tenantID string, page, perPage int) ([]*domain.AIRequest, int, error)

	// ListByProvider retrieves requests for a specific provider
	ListByProvider(ctx context.Context, providerID string, page, perPage int) ([]*domain.AIRequest, int, error)

	// Save persists an AI request (creates or updates)
	Save(ctx context.Context, request *domain.AIRequest) error

	// Delete deletes an AI request
	Delete(ctx context.Context, id string) error
}

// AIFeedbackRepository defines the interface for AI feedback data access
type AIFeedbackRepository interface {
	// FindByID retrieves feedback by ID
	FindByID(ctx context.Context, id string) (*domain.AIFeedback, error)

	// ListByRequest retrieves feedback for a request
	ListByRequest(ctx context.Context, requestID string) ([]*domain.AIFeedback, error)

	// Save persists feedback (creates or updates)
	Save(ctx context.Context, feedback *domain.AIFeedback) error

	// Delete deletes feedback
	Delete(ctx context.Context, id string) error
}

// FewShotExampleRepository defines the interface for few-shot example data access
type FewShotExampleRepository interface {
	// FindByID retrieves an example by ID
	FindByID(ctx context.Context, id string) (*domain.FewShotExample, error)

	// ListByType retrieves examples for a request type
	ListByType(ctx context.Context, requestType domain.RequestType) ([]*domain.FewShotExample, error)

	// List retrieves all examples with pagination
	List(ctx context.Context, tenantID string, page, perPage int) ([]*domain.FewShotExample, int, error)

	// Save persists an example (creates or updates)
	Save(ctx context.Context, example *domain.FewShotExample) error

	// Delete deletes an example
	Delete(ctx context.Context, id string) error

	// IncrementUsageCount increments usage count for an example
	IncrementUsageCount(ctx context.Context, id string) error
}

// AIProviderRepository defines the interface for AI provider data access
type AIProviderRepository interface {
	// FindByID retrieves a provider by ID
	FindByID(ctx context.Context, id string) (*domain.AIProvider, error)

	// ListByTenant retrieves providers for a tenant
	ListByTenant(ctx context.Context, tenantID string) ([]*domain.AIProvider, error)

	// ListActive retrieves active providers for a tenant
	ListActive(ctx context.Context, tenantID string) ([]*domain.AIProvider, error)

	// Save persists a provider (creates or updates)
	Save(ctx context.Context, provider *domain.AIProvider) error

	// Delete deletes a provider
	Delete(ctx context.Context, id string) error
}

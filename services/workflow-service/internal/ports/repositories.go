package ports

import (
	"context"

	"github.com/jeckersberger/rentflow/services/workflow-service/internal/domain"
)

type WorkflowRepository interface {
	Create(ctx context.Context, workflow *domain.Workflow) error
	GetByID(ctx context.Context, tenantID, id string) (*domain.Workflow, error)
	ListByTenant(ctx context.Context, tenantID string) ([]*domain.Workflow, error)
	Update(ctx context.Context, workflow *domain.Workflow) error
	Delete(ctx context.Context, tenantID, id string) error
}

type WorkflowRunRepository interface {
	Create(ctx context.Context, run *domain.WorkflowRun) error
	GetByID(ctx context.Context, id string) (*domain.WorkflowRun, error)
	ListByTenant(ctx context.Context, tenantID string) ([]*domain.WorkflowRun, error)
	ListByWorkflow(ctx context.Context, workflowID string) ([]*domain.WorkflowRun, error)
	Update(ctx context.Context, run *domain.WorkflowRun) error
}

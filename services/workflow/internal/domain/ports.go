package domain

import (
	"context"

	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// Filters
// ---------------------------------------------------------------------------

// DefinitionFilter holds optional criteria for listing workflow definitions.
type DefinitionFilter struct {
	Page    int `json:"page"`
	PerPage int `json:"per_page"`
}

// InstanceFilter holds optional criteria for listing workflow instances.
type InstanceFilter struct {
	ReferenceID *uuid.UUID `json:"reference_id,omitempty"`
	Status      *string    `json:"status,omitempty"`
	Page        int        `json:"page"`
	PerPage     int        `json:"per_page"`
}

// ---------------------------------------------------------------------------
// Repository ports (driven / secondary adapters)
// ---------------------------------------------------------------------------

// WorkflowDefinitionRepository defines persistence operations for WorkflowDefinition aggregates.
type WorkflowDefinitionRepository interface {
	Create(ctx context.Context, def *WorkflowDefinition) error
	GetByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*WorkflowDefinition, error)
	List(ctx context.Context, tenantID uuid.UUID, filter DefinitionFilter) ([]*WorkflowDefinition, int64, error)
	Update(ctx context.Context, def *WorkflowDefinition) error
}

// WorkflowInstanceRepository defines persistence operations for WorkflowInstance aggregates.
type WorkflowInstanceRepository interface {
	Create(ctx context.Context, inst *WorkflowInstance) error
	GetByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*WorkflowInstance, error)
	List(ctx context.Context, tenantID uuid.UUID, filter InstanceFilter) ([]*WorkflowInstance, int64, error)
	Update(ctx context.Context, inst *WorkflowInstance) error
}

// WorkflowActionRepository defines persistence operations for WorkflowAction records.
type WorkflowActionRepository interface {
	Create(ctx context.Context, action *WorkflowAction) error
	ListByInstance(ctx context.Context, instanceID uuid.UUID) ([]*WorkflowAction, error)
}

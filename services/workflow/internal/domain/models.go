package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// Workflow type constants
// ---------------------------------------------------------------------------

const (
	TypeApproval   = "approval"
	TypeReview     = "review"
	TypeEscalation = "escalation"
)

// validTypes is the set of all allowed workflow definition types.
var validTypes = map[string]bool{
	TypeApproval:   true,
	TypeReview:     true,
	TypeEscalation: true,
}

// ValidType checks whether the given string is a valid workflow type.
func ValidType(t string) bool {
	return validTypes[t]
}

// ---------------------------------------------------------------------------
// Instance status constants
// ---------------------------------------------------------------------------

const (
	StatusPending   = "pending"
	StatusActive    = "active"
	StatusApproved  = "approved"
	StatusRejected  = "rejected"
	StatusEscalated = "escalated"
	StatusCancelled = "cancelled"
)

// validStatuses is the set of all allowed workflow instance statuses.
var validStatuses = map[string]bool{
	StatusPending:   true,
	StatusActive:    true,
	StatusApproved:  true,
	StatusRejected:  true,
	StatusEscalated: true,
	StatusCancelled: true,
}

// ValidStatus checks whether the given string is a valid instance status.
func ValidStatus(s string) bool {
	return validStatuses[s]
}

// ---------------------------------------------------------------------------
// Action constants
// ---------------------------------------------------------------------------

const (
	ActionApprove  = "approve"
	ActionReject   = "reject"
	ActionEscalate = "escalate"
)

// validActions is the set of all allowed workflow actions.
var validActions = map[string]bool{
	ActionApprove:  true,
	ActionReject:   true,
	ActionEscalate: true,
}

// ValidAction checks whether the given string is a valid workflow action.
func ValidAction(a string) bool {
	return validActions[a]
}

// ---------------------------------------------------------------------------
// Domain models
// ---------------------------------------------------------------------------

// WorkflowDefinition describes a reusable workflow template.
type WorkflowDefinition struct {
	ID        uuid.UUID        `json:"id"`
	TenantID  uuid.UUID        `json:"tenant_id"`
	Name      string           `json:"name"`
	Type      string           `json:"type"`
	Steps     json.RawMessage  `json:"steps"`
	IsActive  bool             `json:"is_active"`
	CreatedBy *uuid.UUID       `json:"created_by,omitempty"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
}

// WorkflowInstance represents a running workflow tied to a business entity.
type WorkflowInstance struct {
	ID            uuid.UUID        `json:"id"`
	DefinitionID  uuid.UUID        `json:"definition_id"`
	TenantID      uuid.UUID        `json:"tenant_id"`
	ReferenceID   *uuid.UUID       `json:"reference_id,omitempty"`
	ReferenceType string           `json:"reference_type,omitempty"`
	CurrentStep   int              `json:"current_step"`
	Status        string           `json:"status"`
	Data          json.RawMessage  `json:"data"`
	StartedBy     *uuid.UUID       `json:"started_by,omitempty"`
	StartedAt     time.Time        `json:"started_at"`
	CompletedAt   *time.Time       `json:"completed_at,omitempty"`
	Notes         string           `json:"notes,omitempty"`
}

// WorkflowAction records a single action taken on a workflow instance step.
type WorkflowAction struct {
	ID          uuid.UUID  `json:"id"`
	InstanceID  uuid.UUID  `json:"instance_id"`
	Step        int        `json:"step"`
	Action      string     `json:"action"`
	PerformedBy *uuid.UUID `json:"performed_by,omitempty"`
	Comment     string     `json:"comment,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

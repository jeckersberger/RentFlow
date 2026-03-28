package domain

import (
	"context"

	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// Repository ports (driven / secondary adapters)
// ---------------------------------------------------------------------------

// CrewMemberRepository defines persistence operations for CrewMember aggregates.
type CrewMemberRepository interface {
	Create(ctx context.Context, member *CrewMember) error
	GetByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*CrewMember, error)
	List(ctx context.Context, tenantID uuid.UUID) ([]*CrewMember, error)
	Update(ctx context.Context, member *CrewMember) error
	Delete(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error
}

// AssignmentRepository defines persistence operations for CrewAssignment aggregates.
type AssignmentRepository interface {
	Create(ctx context.Context, assignment *CrewAssignment) error
	ListByMember(ctx context.Context, crewMemberID uuid.UUID, tenantID uuid.UUID) ([]*CrewAssignment, error)
	ListAll(ctx context.Context, tenantID uuid.UUID, projectID *uuid.UUID) ([]*CrewAssignment, error)
}

// QualificationRepository defines persistence operations for CrewQualification aggregates.
type QualificationRepository interface {
	Create(ctx context.Context, qualification *CrewQualification) error
	ListByMember(ctx context.Context, crewMemberID uuid.UUID, tenantID uuid.UUID) ([]*CrewQualification, error)
}

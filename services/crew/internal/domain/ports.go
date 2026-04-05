package domain

import (
	"context"
	"time"

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

// TimeEntryRepository defines persistence operations for TimeEntry aggregates.
type TimeEntryRepository interface {
	Create(ctx context.Context, entry *TimeEntry) error
	Update(ctx context.Context, entry *TimeEntry) error
	GetActiveByMember(ctx context.Context, crewMemberID uuid.UUID, tenantID uuid.UUID) (*TimeEntry, error)
	ListByMember(ctx context.Context, crewMemberID uuid.UUID, tenantID uuid.UUID, from time.Time, to time.Time) ([]*TimeEntry, error)
	ListByProject(ctx context.Context, projectID uuid.UUID, tenantID uuid.UUID) ([]*TimeEntry, error)
	ListByTenant(ctx context.Context, tenantID uuid.UUID, from time.Time, to time.Time) ([]*TimeEntry, error)
}

// AvailabilityRepository defines persistence operations for CrewAvailability aggregates.
type AvailabilityRepository interface {
	Upsert(ctx context.Context, availability *CrewAvailability) error
	ListByMember(ctx context.Context, crewMemberID uuid.UUID, tenantID uuid.UUID, from string, to string) ([]*CrewAvailability, error)
	ListAll(ctx context.Context, tenantID uuid.UUID, from string, to string) ([]*CrewAvailability, error)
}

// AvailabilityBlockRepository defines persistence operations for AvailabilityBlock aggregates.
type AvailabilityBlockRepository interface {
	Create(ctx context.Context, block *AvailabilityBlock) error
	GetByID(ctx context.Context, id, tenantID uuid.UUID) (*AvailabilityBlock, error)
	ListByCrewMember(ctx context.Context, crewMemberID, tenantID uuid.UUID, from, to string) ([]*AvailabilityBlock, error)
	ListByDateRange(ctx context.Context, tenantID uuid.UUID, from, to string) ([]*AvailabilityBlock, error)
	Delete(ctx context.Context, id, tenantID uuid.UUID) error
}

// SkillMatchRepository defines queries for the skill-matching use case.
type SkillMatchRepository interface {
	// FindBySkill returns active crew members whose skills array contains the
	// given skill (case-insensitive). Returns the member ID, full name, and
	// whether the match is exact (true) or partial/ILIKE (false).
	FindBySkill(ctx context.Context, tenantID uuid.UUID, skill string) ([]SkillMatchRow, error)

	// BlockedMemberIDs returns the set of crew member IDs that have an
	// overlapping availability block within the given date range.
	BlockedMemberIDs(ctx context.Context, tenantID uuid.UUID, from, to string) (map[uuid.UUID]bool, error)

	// AssignedMemberIDs returns the set of crew member IDs that already have
	// a non-cancelled assignment overlapping the given date range, optionally
	// excluding a specific project.
	AssignedMemberIDs(ctx context.Context, tenantID uuid.UUID, from, to string, excludeProjectID *uuid.UUID) (map[uuid.UUID]bool, error)
}

// SkillMatchRow is a lightweight projection returned by SkillMatchRepository.FindBySkill.
type SkillMatchRow struct {
	CrewMemberID uuid.UUID
	FirstName    string
	LastName     string
	ExactMatch   bool
}

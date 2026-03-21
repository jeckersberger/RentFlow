package ports

import (
	"context"
	"time"

	"github.com/jeckersberger/rentflow/services/crew-service/internal/domain"
)

type CrewRepository interface {
	CreateCrewMember(ctx context.Context, member *domain.CrewMember) error
	UpdateCrewMember(ctx context.Context, member *domain.CrewMember) error
	GetCrewMember(ctx context.Context, tenantID, crewMemberID string) (*domain.CrewMember, error)
	ListCrewMembers(ctx context.Context, tenantID string, limit, offset int) ([]*domain.CrewMember, int64, error)
	DeleteCrewMember(ctx context.Context, tenantID, crewMemberID string) error
}

type TimeEntryRepository interface {
	CreateTimeEntry(ctx context.Context, entry *domain.TimeEntry) error
	UpdateTimeEntry(ctx context.Context, entry *domain.TimeEntry) error
	GetTimeEntry(ctx context.Context, tenantID, timeEntryID string) (*domain.TimeEntry, error)
	ListTimeEntries(ctx context.Context, tenantID string, limit, offset int) ([]*domain.TimeEntry, int64, error)
	ListByCrewMemberAndPeriod(ctx context.Context, tenantID, crewMemberID string, from, to time.Time) ([]*domain.TimeEntry, int64, error)
}

type AssignmentRepository interface {
	CreateAssignment(ctx context.Context, assignment *domain.Assignment) error
	GetAssignment(ctx context.Context, tenantID, assignmentID string) (*domain.Assignment, error)
	ListAssignments(ctx context.Context, tenantID string, limit, offset int) ([]*domain.Assignment, int64, error)
	GetByProject(ctx context.Context, tenantID, projectID string) ([]*domain.Assignment, error)
	DeleteAssignment(ctx context.Context, tenantID, assignmentID string) error
}

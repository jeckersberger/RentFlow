package ports

import (
	"context"

	"github.com/jeckersberger/rentflow/services/crew-service/internal/domain"
)

// CrewMemberRepository defines the interface for crew member data access
type CrewMemberRepository interface {
	// FindByID retrieves a crew member by ID
	FindByID(ctx context.Context, id string) (*domain.CrewMember, error)

	// FindByEmail retrieves a crew member by email within a tenant
	FindByEmail(ctx context.Context, tenantID, email string) (*domain.CrewMember, error)

	// List retrieves crew members in a tenant with pagination
	List(ctx context.Context, tenantID string, page, perPage int) ([]*domain.CrewMember, int, error)

	// Save persists a crew member (creates or updates)
	Save(ctx context.Context, member *domain.CrewMember) error

	// Delete deletes a crew member
	Delete(ctx context.Context, id string) error
}

// QualificationRepository defines the interface for qualification data access
type QualificationRepository interface {
	// FindByID retrieves a qualification by ID
	FindByID(ctx context.Context, id string) (*domain.Qualification, error)

	// ListByCrewMember retrieves qualifications for a crew member
	ListByCrewMember(ctx context.Context, crewMemberID string) ([]*domain.Qualification, error)

	// Save persists a qualification (creates or updates)
	Save(ctx context.Context, qualification *domain.Qualification) error

	// Delete deletes a qualification
	Delete(ctx context.Context, id string) error
}

// CrewAssignmentRepository defines the interface for crew assignment data access
type CrewAssignmentRepository interface {
	// FindByID retrieves an assignment by ID
	FindByID(ctx context.Context, id string) (*domain.CrewAssignment, error)

	// List retrieves assignments in a tenant with pagination
	List(ctx context.Context, tenantID string, page, perPage int) ([]*domain.CrewAssignment, int, error)

	// ListByCrewMember retrieves assignments for a crew member
	ListByCrewMember(ctx context.Context, crewMemberID string) ([]*domain.CrewAssignment, error)

	// ListByDateRange retrieves assignments within a date range
	ListByDateRange(ctx context.Context, tenantID string, startDate, endDate interface{}) ([]*domain.CrewAssignment, error)

	// Save persists an assignment (creates or updates)
	Save(ctx context.Context, assignment *domain.CrewAssignment) error

	// Delete deletes an assignment
	Delete(ctx context.Context, id string) error
}

// TimeRecordRepository defines the interface for time record data access
type TimeRecordRepository interface {
	// FindByID retrieves a time record by ID
	FindByID(ctx context.Context, id string) (*domain.TimeRecord, error)

	// ListByCrewMember retrieves time records for a crew member
	ListByCrewMember(ctx context.Context, crewMemberID string) ([]*domain.TimeRecord, error)

	// ListByCrewMemberAndDate retrieves time records for a crew member on a specific date
	ListByCrewMemberAndDate(ctx context.Context, crewMemberID string, date interface{}) ([]*domain.TimeRecord, error)

	// Save persists a time record (creates or updates)
	Save(ctx context.Context, record *domain.TimeRecord) error

	// Delete deletes a time record
	Delete(ctx context.Context, id string) error
}

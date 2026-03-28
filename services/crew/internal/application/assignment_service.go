package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/services/crew/internal/domain"
)

// ---------------------------------------------------------------------------
// Request DTOs
// ---------------------------------------------------------------------------

// CreateAssignmentRequest holds the data needed to create a new crew assignment.
type CreateAssignmentRequest struct {
	ProjectID    uuid.UUID `json:"project_id"`
	Role         string    `json:"role,omitempty"`
	StartDate    string    `json:"start_date,omitempty"`
	EndDate      string    `json:"end_date,omitempty"`
	HoursPlanned float64  `json:"hours_planned,omitempty"`
	Status       string    `json:"status,omitempty"`
	Notes        string    `json:"notes,omitempty"`
}

// ---------------------------------------------------------------------------
// Service
// ---------------------------------------------------------------------------

// AssignmentService implements the application-level use cases for crew assignments.
type AssignmentService struct {
	assignmentRepo domain.AssignmentRepository
	memberRepo     domain.CrewMemberRepository
	logger         zerolog.Logger
}

// NewAssignmentService constructs a new AssignmentService.
func NewAssignmentService(
	assignmentRepo domain.AssignmentRepository,
	memberRepo domain.CrewMemberRepository,
	logger zerolog.Logger,
) *AssignmentService {
	return &AssignmentService{
		assignmentRepo: assignmentRepo,
		memberRepo:     memberRepo,
		logger:         logger.With().Str("service", "assignment").Logger(),
	}
}

// Create validates the request and persists a new assignment for a crew member.
func (s *AssignmentService) Create(
	ctx context.Context,
	crewMemberID uuid.UUID,
	tenantID uuid.UUID,
	req CreateAssignmentRequest,
) (*domain.CrewAssignment, error) {
	// Verify that the crew member exists.
	if _, err := s.memberRepo.GetByID(ctx, crewMemberID, tenantID); err != nil {
		return nil, fmt.Errorf("create assignment – member: %w", err)
	}

	status := req.Status
	if status == "" {
		status = domain.AssignmentStatusPlanned
	}

	assignment := &domain.CrewAssignment{
		ID:           uuid.New(),
		CrewMemberID: crewMemberID,
		ProjectID:    req.ProjectID,
		TenantID:     tenantID,
		Role:         req.Role,
		StartDate:    req.StartDate,
		EndDate:      req.EndDate,
		HoursPlanned: req.HoursPlanned,
		HoursActual:  0,
		Status:       status,
		Notes:        req.Notes,
		CreatedAt:    time.Now(),
	}

	if err := s.assignmentRepo.Create(ctx, assignment); err != nil {
		s.logger.Error().Err(err).
			Str("crew_member_id", crewMemberID.String()).
			Str("project_id", req.ProjectID.String()).
			Str("tenant_id", tenantID.String()).
			Msg("failed to create assignment")
		return nil, fmt.Errorf("create assignment: %w", err)
	}

	s.logger.Info().
		Str("assignment_id", assignment.ID.String()).
		Str("crew_member_id", crewMemberID.String()).
		Str("tenant_id", tenantID.String()).
		Msg("assignment created")

	return assignment, nil
}

// ListByMember returns all assignments for a specific crew member.
func (s *AssignmentService) ListByMember(
	ctx context.Context,
	crewMemberID uuid.UUID,
	tenantID uuid.UUID,
) ([]*domain.CrewAssignment, error) {
	assignments, err := s.assignmentRepo.ListByMember(ctx, crewMemberID, tenantID)
	if err != nil {
		s.logger.Error().Err(err).
			Str("crew_member_id", crewMemberID.String()).
			Str("tenant_id", tenantID.String()).
			Msg("failed to list assignments for member")
		return nil, fmt.Errorf("list assignments by member: %w", err)
	}
	return assignments, nil
}

// ListAll returns all assignments for a tenant, optionally filtered by project_id.
func (s *AssignmentService) ListAll(
	ctx context.Context,
	tenantID uuid.UUID,
	projectID *uuid.UUID,
) ([]*domain.CrewAssignment, error) {
	assignments, err := s.assignmentRepo.ListAll(ctx, tenantID, projectID)
	if err != nil {
		s.logger.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Msg("failed to list all assignments")
		return nil, fmt.Errorf("list all assignments: %w", err)
	}
	return assignments, nil
}

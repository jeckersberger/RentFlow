package application

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/crew-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/crew-service/internal/ports"
)

// AssignmentService handles crew assignment business logic
type AssignmentService struct {
	assignmentRepo ports.CrewAssignmentRepository
	crewRepo       ports.CrewMemberRepository
	logger         logger.Logger
}

// NewAssignmentService creates a new assignment service
func NewAssignmentService(
	assignmentRepo ports.CrewAssignmentRepository,
	crewRepo ports.CrewMemberRepository,
	log logger.Logger,
) *AssignmentService {
	return &AssignmentService{
		assignmentRepo: assignmentRepo,
		crewRepo:       crewRepo,
		logger:         log,
	}
}

// CreateAssignment creates a new crew assignment
func (s *AssignmentService) CreateAssignment(ctx context.Context, cmd CreateCrewAssignmentCommand) (*CrewAssignmentDTO, error) {
	if cmd.TenantID == "" {
		return nil, domain.ErrTenantIDRequired
	}
	if cmd.CrewMemberID == "" {
		return nil, domain.ErrCrewMemberIDRequired
	}

	// Parse dates
	startDate, err := time.Parse(time.RFC3339, cmd.StartDate)
	if err != nil {
		s.logger.Error("invalid start date", err)
		return nil, domain.ErrInvalidDateRange
	}
	endDate, err := time.Parse(time.RFC3339, cmd.EndDate)
	if err != nil {
		s.logger.Error("invalid end date", err)
		return nil, domain.ErrInvalidDateRange
	}

	if !startDate.Before(endDate) {
		return nil, domain.ErrInvalidDateRange
	}

	// Verify crew member exists
	member, err := s.crewRepo.FindByID(ctx, cmd.CrewMemberID)
	if err != nil {
		s.logger.Error("failed to find crew member", err)
		return nil, err
	}
	if member == nil {
		return nil, domain.ErrCrewMemberNotFound
	}

	// Check for conflicts
	assignments, err := s.assignmentRepo.ListByCrewMember(ctx, cmd.CrewMemberID)
	if err != nil {
		s.logger.Error("failed to list assignments", err)
		return nil, err
	}

	assignment := &domain.CrewAssignment{
		ID:           uuid.New().String(),
		TenantID:     cmd.TenantID,
		CrewMemberID: cmd.CrewMemberID,
		ProjectID:    cmd.ProjectID,
		TourID:       cmd.TourID,
		Role:         cmd.Role,
		StartDate:    startDate,
		EndDate:      endDate,
		Status:       domain.AssignmentStatusPlanned,
		Notes:        cmd.Notes,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}

	if cmd.Status != "" {
		assignment.Status = domain.AssignmentStatus(cmd.Status)
	}

	if assignment.HasConflict(assignments) {
		return nil, domain.ErrConflictDetected
	}

	if err := s.assignmentRepo.Save(ctx, assignment); err != nil {
		s.logger.Error("failed to save assignment", err)
		return nil, err
	}

	return ToCrewAssignmentDTO(assignment), nil
}

// GetAssignment retrieves an assignment by ID
func (s *AssignmentService) GetAssignment(ctx context.Context, id string) (*CrewAssignmentDTO, error) {
	assignment, err := s.assignmentRepo.FindByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to find assignment", err, "id", id)
		return nil, err
	}
	if assignment == nil {
		return nil, domain.ErrAssignmentNotFound
	}
	return ToCrewAssignmentDTO(assignment), nil
}

// UpdateAssignment updates a crew assignment
func (s *AssignmentService) UpdateAssignment(ctx context.Context, cmd UpdateCrewAssignmentCommand) (*CrewAssignmentDTO, error) {
	assignment, err := s.assignmentRepo.FindByID(ctx, cmd.ID)
	if err != nil {
		s.logger.Error("failed to find assignment", err, "id", cmd.ID)
		return nil, err
	}
	if assignment == nil {
		return nil, domain.ErrAssignmentNotFound
	}

	// Parse dates
	startDate, err := time.Parse(time.RFC3339, cmd.StartDate)
	if err != nil {
		s.logger.Error("invalid start date", err)
		return nil, domain.ErrInvalidDateRange
	}
	endDate, err := time.Parse(time.RFC3339, cmd.EndDate)
	if err != nil {
		s.logger.Error("invalid end date", err)
		return nil, domain.ErrInvalidDateRange
	}

	if !startDate.Before(endDate) {
		return nil, domain.ErrInvalidDateRange
	}

	// Check for conflicts with other assignments (excluding this one)
	assignments, err := s.assignmentRepo.ListByCrewMember(ctx, cmd.CrewMemberID)
	if err != nil {
		s.logger.Error("failed to list assignments", err)
		return nil, err
	}

	assignment.Role = cmd.Role
	assignment.ProjectID = cmd.ProjectID
	assignment.TourID = cmd.TourID
	assignment.StartDate = startDate
	assignment.EndDate = endDate
	assignment.Status = domain.AssignmentStatus(cmd.Status)
	assignment.Notes = cmd.Notes
	assignment.UpdatedAt = time.Now().UTC()

	if assignment.HasConflict(assignments) {
		return nil, domain.ErrConflictDetected
	}

	if err := s.assignmentRepo.Save(ctx, assignment); err != nil {
		s.logger.Error("failed to update assignment", err)
		return nil, err
	}

	return ToCrewAssignmentDTO(assignment), nil
}

// DeleteAssignment deletes a crew assignment
func (s *AssignmentService) DeleteAssignment(ctx context.Context, id string) error {
	assignment, err := s.assignmentRepo.FindByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to find assignment", err, "id", id)
		return err
	}
	if assignment == nil {
		return domain.ErrAssignmentNotFound
	}

	if err := s.assignmentRepo.Delete(ctx, id); err != nil {
		s.logger.Error("failed to delete assignment", err)
		return err
	}

	return nil
}

// ListAssignments lists crew assignments for a tenant
func (s *AssignmentService) ListAssignments(ctx context.Context, tenantID string, page, perPage int) (*PaginatedResult, error) {
	assignments, total, err := s.assignmentRepo.List(ctx, tenantID, page, perPage)
	if err != nil {
		s.logger.Error("failed to list assignments", err)
		return nil, err
	}

	dtos := make([]interface{}, len(assignments))
	for i, assignment := range assignments {
		dtos[i] = ToCrewAssignmentDTO(assignment)
	}

	totalPages := (total + perPage - 1) / perPage
	return &PaginatedResult{
		Data:       dtos,
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

// DetectConflicts detects assignment conflicts
func (s *AssignmentService) DetectConflicts(ctx context.Context, crewMemberID string, startDate, endDate time.Time) ([]ConflictDTO, error) {
	assignments, err := s.assignmentRepo.ListByCrewMember(ctx, crewMemberID)
	if err != nil {
		s.logger.Error("failed to list assignments", err)
		return nil, err
	}

	member, err := s.crewRepo.FindByID(ctx, crewMemberID)
	if err != nil {
		s.logger.Error("failed to find crew member", err)
		return nil, err
	}

	var conflicts []ConflictDTO
	for _, a := range assignments {
		if a.Status == domain.AssignmentStatusCancelled {
			continue
		}
		// Check for overlap
		if !(endDate.Before(a.StartDate) || startDate.After(a.EndDate)) {
			var name string
			if member != nil {
				name = member.FirstName + " " + member.LastName
			}
			conflicts = append(conflicts, ConflictDTO{
				CrewMemberID:    crewMemberID,
				CrewMemberName:  name,
				ConflictingDates: a.StartDate.Format("2006-01-02") + " to " + a.EndDate.Format("2006-01-02"),
				ExistingAssignment: ToCrewAssignmentDTO(a),
			})
		}
	}

	return conflicts, nil
}

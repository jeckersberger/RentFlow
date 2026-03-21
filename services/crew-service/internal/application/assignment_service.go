package application

import (
	"context"
	"fmt"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/crew-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/crew-service/internal/ports"
)

type AssignmentService struct {
	assignRepo ports.AssignmentRepository
	logger     *logger.Logger
}

func NewAssignmentService(assignRepo ports.AssignmentRepository, logger *logger.Logger) *AssignmentService {
	return &AssignmentService{
		assignRepo: assignRepo,
		logger:     logger,
	}
}

func (s *AssignmentService) CreateAssignment(ctx context.Context, cmd CreateAssignmentCommand) (*AssignmentDTO, error) {
	if cmd.TenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	aID := fmt.Sprintf("assign_%d", hashString(cmd.TenantID+cmd.CrewMemberID+cmd.ProjectID))
	a := domain.NewAssignment(aID, cmd.TenantID, cmd.CrewMemberID, cmd.ProjectID, cmd.Role, cmd.StartDate, cmd.EndDate)

	if err := a.Validate(); err != nil {
		return nil, domain.NewDomainError("VALIDATION_ERROR", err.Error(), nil)
	}

	if err := s.assignRepo.CreateAssignment(ctx, a); err != nil {
		return nil, domain.NewDomainError("CREATE_FAILED", "failed to create assignment", err)
	}

	return AssignmentToDTO(a), nil
}

func (s *AssignmentService) GetAssignment(ctx context.Context, tenantID, assignmentID string) (*AssignmentDTO, error) {
	if tenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	a, err := s.assignRepo.GetAssignment(ctx, tenantID, assignmentID)
	if err != nil {
		return nil, domain.NewDomainError("NOT_FOUND", "assignment not found", err)
	}

	return AssignmentToDTO(a), nil
}

func (s *AssignmentService) ListAssignments(ctx context.Context, tenantID string, limit, offset int) ([]*AssignmentDTO, int64, error) {
	if tenantID == "" {
		return nil, 0, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	assignments, total, err := s.assignRepo.ListAssignments(ctx, tenantID, limit, offset)
	if err != nil {
		return nil, 0, domain.NewDomainError("LIST_FAILED", "failed to list assignments", err)
	}

	dtos := make([]*AssignmentDTO, len(assignments))
	for i, a := range assignments {
		dtos[i] = AssignmentToDTO(a)
	}

	return dtos, total, nil
}

func (s *AssignmentService) GetByProject(ctx context.Context, tenantID, projectID string) ([]*AssignmentDTO, error) {
	if tenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	assignments, err := s.assignRepo.GetByProject(ctx, tenantID, projectID)
	if err != nil {
		return nil, domain.NewDomainError("QUERY_FAILED", "failed to get assignments", err)
	}

	dtos := make([]*AssignmentDTO, len(assignments))
	for i, a := range assignments {
		dtos[i] = AssignmentToDTO(a)
	}

	return dtos, nil
}

func (s *AssignmentService) DeleteAssignment(ctx context.Context, tenantID, assignmentID string) error {
	if tenantID == "" {
		return domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	if err := s.assignRepo.DeleteAssignment(ctx, tenantID, assignmentID); err != nil {
		return domain.NewDomainError("DELETE_FAILED", "failed to delete assignment", err)
	}

	return nil
}

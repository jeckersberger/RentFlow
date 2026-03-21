package application

import (
	"context"
	"fmt"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/project-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/project-service/internal/ports"
)

type ProjectService struct {
	repo   ports.ProjectRepository
	logger *logger.Logger
}

func NewProjectService(repo ports.ProjectRepository, logger *logger.Logger) *ProjectService {
	return &ProjectService{
		repo:   repo,
		logger: logger,
	}
}

func (s *ProjectService) CreateProject(ctx context.Context, cmd CreateProjectCommand) (*ProjectDTO, error) {
	if cmd.TenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}
	if cmd.Name == "" {
		return nil, domain.NewDomainError("NAME_REQUIRED", "project name is required", nil)
	}
	if cmd.ClientName == "" {
		return nil, domain.NewDomainError("CLIENT_REQUIRED", "client name is required", nil)
	}

	projectID := fmt.Sprintf("proj_%d", hashString(cmd.TenantID+cmd.Name))

	project := domain.NewProject(projectID, cmd.TenantID, cmd.Name, cmd.ClientName, cmd.CreatedByUserID)
	project.Description = cmd.Description
	project.ClientEmail = cmd.ClientEmail
	project.ClientPhone = cmd.ClientPhone
	project.ClientAddress = domain.Address{
		Street:      cmd.ClientAddress.Street,
		City:        cmd.ClientAddress.City,
		State:       cmd.ClientAddress.State,
		PostalCode:  cmd.ClientAddress.PostalCode,
		Country:     cmd.ClientAddress.Country,
		Coordinates: cmd.ClientAddress.Coordinates,
	}
	project.VenueAddress = domain.Address{
		Street:      cmd.VenueAddress.Street,
		City:        cmd.VenueAddress.City,
		State:       cmd.VenueAddress.State,
		PostalCode:  cmd.VenueAddress.PostalCode,
		Country:     cmd.VenueAddress.Country,
		Coordinates: cmd.VenueAddress.Coordinates,
	}
	project.StartDate = cmd.StartDate
	project.EndDate = cmd.EndDate
	project.SetupDate = cmd.SetupDate
	project.TeardownDate = cmd.TeardownDate
	project.Budget = cmd.Budget
	project.Currency = cmd.Currency
	project.Notes = cmd.Notes
	project.Tags = cmd.Tags

	if err := project.Validate(); err != nil {
		return nil, domain.NewDomainError("VALIDATION_ERROR", err.Error(), nil)
	}

	if err := s.repo.Create(ctx, project); err != nil {
		return nil, domain.NewDomainError("CREATE_ERROR", "failed to create project", err)
	}

	s.logger.Info("Project created", "id", project.ID, "tenant_id", cmd.TenantID)
	return ProjectToDTO(project), nil
}

func (s *ProjectService) UpdateProject(ctx context.Context, cmd UpdateProjectCommand) (*ProjectDTO, error) {
	if cmd.TenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	project, err := s.repo.GetByID(ctx, cmd.TenantID, cmd.ID)
	if err != nil {
		return nil, domain.NewDomainError("NOT_FOUND", "project not found", err)
	}

	project.Name = cmd.Name
	project.Description = cmd.Description
	project.ClientName = cmd.ClientName
	project.ClientEmail = cmd.ClientEmail
	project.ClientPhone = cmd.ClientPhone
	project.ClientAddress = domain.Address{
		Street:      cmd.ClientAddress.Street,
		City:        cmd.ClientAddress.City,
		State:       cmd.ClientAddress.State,
		PostalCode:  cmd.ClientAddress.PostalCode,
		Country:     cmd.ClientAddress.Country,
		Coordinates: cmd.ClientAddress.Coordinates,
	}
	project.VenueAddress = domain.Address{
		Street:      cmd.VenueAddress.Street,
		City:        cmd.VenueAddress.City,
		State:       cmd.VenueAddress.State,
		PostalCode:  cmd.VenueAddress.PostalCode,
		Country:     cmd.VenueAddress.Country,
		Coordinates: cmd.VenueAddress.Coordinates,
	}
	project.StartDate = cmd.StartDate
	project.EndDate = cmd.EndDate
	project.SetupDate = cmd.SetupDate
	project.TeardownDate = cmd.TeardownDate
	project.Budget = cmd.Budget
	project.Currency = cmd.Currency
	project.Notes = cmd.Notes
	project.Tags = cmd.Tags

	if err := s.repo.Update(ctx, project); err != nil {
		return nil, domain.NewDomainError("UPDATE_ERROR", "failed to update project", err)
	}

	s.logger.Info("Project updated", "id", cmd.ID, "tenant_id", cmd.TenantID)
	return ProjectToDTO(project), nil
}

func (s *ProjectService) GetProject(ctx context.Context, tenantID, projectID string) (*ProjectDTO, error) {
	if tenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	project, err := s.repo.GetByID(ctx, tenantID, projectID)
	if err != nil {
		return nil, domain.NewDomainError("NOT_FOUND", "project not found", err)
	}

	return ProjectToDTO(project), nil
}

func (s *ProjectService) ListProjects(ctx context.Context, query ListProjectsQuery) (*PaginatedResult, error) {
	if query.TenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	result, err := s.repo.List(ctx, query.TenantID, query.Limit, query.Offset)
	if err != nil {
		return nil, domain.NewDomainError("QUERY_ERROR", "failed to list projects", err)
	}

	dtos := make([]*ProjectDTO, len(result.Items))
	for i, project := range result.Items {
		dtos[i] = ProjectToDTO(project)
	}

	return &PaginatedResult{
		Data:   dtos,
		Total:  result.Total,
		Limit:  result.Limit,
		Offset: result.Offset,
	}, nil
}

func (s *ProjectService) SearchProjects(ctx context.Context, query SearchProjectsQuery) (*PaginatedResult, error) {
	if query.TenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	result, err := s.repo.Search(ctx, query.TenantID, query.SearchTerm, query.Limit, query.Offset)
	if err != nil {
		return nil, domain.NewDomainError("SEARCH_ERROR", "failed to search projects", err)
	}

	dtos := make([]*ProjectDTO, len(result.Items))
	for i, project := range result.Items {
		dtos[i] = ProjectToDTO(project)
	}

	return &PaginatedResult{
		Data:   dtos,
		Total:  result.Total,
		Limit:  result.Limit,
		Offset: result.Offset,
	}, nil
}

func (s *ProjectService) ChangeStatus(ctx context.Context, cmd ChangeProjectStatusCommand) error {
	if cmd.TenantID == "" {
		return domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	project, err := s.repo.GetByID(ctx, cmd.TenantID, cmd.ID)
	if err != nil {
		return domain.NewDomainError("NOT_FOUND", "project not found", err)
	}

	if err := project.ChangeStatus(cmd.Status); err != nil {
		return domain.NewDomainError("INVALID_STATUS", err.Error(), nil)
	}

	if err := s.repo.Update(ctx, project); err != nil {
		return domain.NewDomainError("UPDATE_ERROR", "failed to update status", err)
	}

	s.logger.Info("Project status changed", "id", cmd.ID, "status", cmd.Status)
	return nil
}

func (s *ProjectService) SetProjectManager(ctx context.Context, cmd SetProjectManagerCommand) error {
	if cmd.TenantID == "" {
		return domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	project, err := s.repo.GetByID(ctx, cmd.TenantID, cmd.ID)
	if err != nil {
		return domain.NewDomainError("NOT_FOUND", "project not found", err)
	}

	if err := project.SetProjectManager(cmd.ManagerID); err != nil {
		return domain.NewDomainError("INVALID_INPUT", err.Error(), nil)
	}

	if err := s.repo.Update(ctx, project); err != nil {
		return domain.NewDomainError("UPDATE_ERROR", "failed to update project manager", err)
	}

	s.logger.Info("Project manager set", "id", cmd.ID, "manager_id", cmd.ManagerID)
	return nil
}

func (s *ProjectService) DeleteProject(ctx context.Context, tenantID, projectID string) error {
	if tenantID == "" {
		return domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	_, err := s.repo.GetByID(ctx, tenantID, projectID)
	if err != nil {
		return domain.NewDomainError("NOT_FOUND", "project not found", err)
	}

	if err := s.repo.Delete(ctx, tenantID, projectID); err != nil {
		return domain.NewDomainError("DELETE_ERROR", "failed to delete project", err)
	}

	s.logger.Info("Project deleted", "id", projectID, "tenant_id", tenantID)
	return nil
}

func hashString(s string) int64 {
	h := int64(5381)
	for _, c := range s {
		h = ((h << 5) + h) + int64(c)
	}
	return h & 0x7FFFFFFFFFFFFFFF
}

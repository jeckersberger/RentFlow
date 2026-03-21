package application

import (
	"context"
	"fmt"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/project-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/project-service/internal/ports"
)

type PacklistService struct {
	repo   ports.PacklistRepository
	logger *logger.Logger
}

func NewPacklistService(repo ports.PacklistRepository, logger *logger.Logger) *PacklistService {
	return &PacklistService{
		repo:   repo,
		logger: logger,
	}
}

func (s *PacklistService) CreatePacklist(ctx context.Context, cmd CreatePacklistCommand) (*PacklistDTO, error) {
	if cmd.TenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}
	if cmd.ProjectID == "" {
		return nil, domain.NewDomainError("PROJECT_REQUIRED", "project ID is required", nil)
	}
	if cmd.Name == "" {
		return nil, domain.NewDomainError("NAME_REQUIRED", "packlist name is required", nil)
	}

	packlistID := fmt.Sprintf("pklist_%d", hashString(cmd.TenantID+cmd.ProjectID+cmd.Name))

	packlist := domain.NewPacklist(packlistID, cmd.ProjectID, cmd.TenantID, cmd.Name)

	if err := packlist.Validate(); err != nil {
		return nil, domain.NewDomainError("VALIDATION_ERROR", err.Error(), nil)
	}

	if err := s.repo.Create(ctx, packlist); err != nil {
		return nil, domain.NewDomainError("CREATE_ERROR", "failed to create packlist", err)
	}

	s.logger.Info("Packlist created", "id", packlist.ID, "tenant_id", cmd.TenantID)
	return PacklistToDTO(packlist), nil
}

func (s *PacklistService) GetPacklist(ctx context.Context, tenantID, packlistID string) (*PacklistDTO, error) {
	if tenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	packlist, err := s.repo.GetByID(ctx, tenantID, packlistID)
	if err != nil {
		return nil, domain.NewDomainError("NOT_FOUND", "packlist not found", err)
	}

	return PacklistToDTO(packlist), nil
}

func (s *PacklistService) ListByProject(ctx context.Context, query ListPacklistsQuery) (*PaginatedResult, error) {
	if query.TenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}
	if query.ProjectID == "" {
		return nil, domain.NewDomainError("PROJECT_REQUIRED", "project ID is required", nil)
	}

	result, err := s.repo.ListByProjectID(ctx, query.TenantID, query.ProjectID, query.Limit, query.Offset)
	if err != nil {
		return nil, domain.NewDomainError("QUERY_ERROR", "failed to list packlists", err)
	}

	dtos := make([]*PacklistDTO, len(result.Items))
	for i, packlist := range result.Items {
		dtos[i] = PacklistToDTO(packlist)
	}

	return &PaginatedResult{
		Data:   dtos,
		Total:  result.Total,
		Limit:  result.Limit,
		Offset: result.Offset,
	}, nil
}

func (s *PacklistService) AddItem(ctx context.Context, cmd AddPacklistItemCommand) (*PacklistDTO, error) {
	if cmd.TenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	packlist, err := s.repo.GetByID(ctx, cmd.TenantID, cmd.PacklistID)
	if err != nil {
		return nil, domain.NewDomainError("NOT_FOUND", "packlist not found", err)
	}

	itemID := fmt.Sprintf("pitem_%d", hashString(cmd.PacklistID+cmd.EquipmentID))
	if err := packlist.AddItem(itemID, cmd.EquipmentID, cmd.EquipmentName, cmd.Quantity); err != nil {
		return nil, domain.NewDomainError("INVALID_INPUT", err.Error(), nil)
	}

	if err := s.repo.Update(ctx, packlist); err != nil {
		return nil, domain.NewDomainError("UPDATE_ERROR", "failed to add item", err)
	}

	s.logger.Info("Item added to packlist", "packlist_id", cmd.PacklistID, "equipment_id", cmd.EquipmentID)
	return PacklistToDTO(packlist), nil
}

func (s *PacklistService) RemoveItem(ctx context.Context, cmd RemovePacklistItemCommand) (*PacklistDTO, error) {
	if cmd.TenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	packlist, err := s.repo.GetByID(ctx, cmd.TenantID, cmd.PacklistID)
	if err != nil {
		return nil, domain.NewDomainError("NOT_FOUND", "packlist not found", err)
	}

	if err := packlist.RemoveItem(cmd.EquipmentID); err != nil {
		return nil, domain.NewDomainError("INVALID_INPUT", err.Error(), nil)
	}

	if err := s.repo.Update(ctx, packlist); err != nil {
		return nil, domain.NewDomainError("UPDATE_ERROR", "failed to remove item", err)
	}

	s.logger.Info("Item removed from packlist", "packlist_id", cmd.PacklistID, "equipment_id", cmd.EquipmentID)
	return PacklistToDTO(packlist), nil
}

func (s *PacklistService) MarkItemPacked(ctx context.Context, cmd MarkItemPackedCommand) (*PacklistDTO, error) {
	if cmd.TenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	packlist, err := s.repo.GetByID(ctx, cmd.TenantID, cmd.PacklistID)
	if err != nil {
		return nil, domain.NewDomainError("NOT_FOUND", "packlist not found", err)
	}

	if err := packlist.MarkItemPacked(cmd.EquipmentID, cmd.QuantityPacked); err != nil {
		return nil, domain.NewDomainError("INVALID_INPUT", err.Error(), nil)
	}

	if err := s.repo.Update(ctx, packlist); err != nil {
		return nil, domain.NewDomainError("UPDATE_ERROR", "failed to mark item packed", err)
	}

	s.logger.Info("Item marked packed", "packlist_id", cmd.PacklistID, "equipment_id", cmd.EquipmentID)
	return PacklistToDTO(packlist), nil
}

func (s *PacklistService) MarkItemReturned(ctx context.Context, cmd MarkItemReturnedCommand) (*PacklistDTO, error) {
	if cmd.TenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	packlist, err := s.repo.GetByID(ctx, cmd.TenantID, cmd.PacklistID)
	if err != nil {
		return nil, domain.NewDomainError("NOT_FOUND", "packlist not found", err)
	}

	if err := packlist.MarkItemReturned(cmd.EquipmentID, cmd.QuantityReturned); err != nil {
		return nil, domain.NewDomainError("INVALID_INPUT", err.Error(), nil)
	}

	if err := s.repo.Update(ctx, packlist); err != nil {
		return nil, domain.NewDomainError("UPDATE_ERROR", "failed to mark item returned", err)
	}

	s.logger.Info("Item marked returned", "packlist_id", cmd.PacklistID, "equipment_id", cmd.EquipmentID)
	return PacklistToDTO(packlist), nil
}

func (s *PacklistService) ChangeStatus(ctx context.Context, cmd ChangePacklistStatusCommand) error {
	if cmd.TenantID == "" {
		return domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	packlist, err := s.repo.GetByID(ctx, cmd.TenantID, cmd.ID)
	if err != nil {
		return domain.NewDomainError("NOT_FOUND", "packlist not found", err)
	}

	if err := packlist.ChangeStatus(cmd.Status); err != nil {
		return domain.NewDomainError("INVALID_STATUS", err.Error(), nil)
	}

	if err := s.repo.Update(ctx, packlist); err != nil {
		return domain.NewDomainError("UPDATE_ERROR", "failed to update status", err)
	}

	s.logger.Info("Packlist status changed", "id", cmd.ID, "status", cmd.Status)
	return nil
}

func (s *PacklistService) DeletePacklist(ctx context.Context, tenantID, packlistID string) error {
	if tenantID == "" {
		return domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	_, err := s.repo.GetByID(ctx, tenantID, packlistID)
	if err != nil {
		return domain.NewDomainError("NOT_FOUND", "packlist not found", err)
	}

	if err := s.repo.Delete(ctx, tenantID, packlistID); err != nil {
		return domain.NewDomainError("DELETE_ERROR", "failed to delete packlist", err)
	}

	s.logger.Info("Packlist deleted", "id", packlistID, "tenant_id", tenantID)
	return nil
}

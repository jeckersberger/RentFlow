package application

import (
	"context"
	"fmt"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/inventory-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/inventory-service/internal/ports"
)

type CategoryService struct {
	catRepo ports.CategoryRepository
	logger  logger.Logger
}

func NewCategoryService(catRepo ports.CategoryRepository, logger logger.Logger) *CategoryService {
	return &CategoryService{
		catRepo: catRepo,
		logger:  logger,
	}
}

func (s *CategoryService) CreateCategory(ctx context.Context, cmd CreateCategoryCommand) (*CategoryDTO, error) {
	if cmd.TenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}
	if cmd.Name == "" {
		return nil, domain.NewDomainError("NAME_REQUIRED", "category name is required", nil)
	}

	// Verify parent exists if specified
	if cmd.ParentID != nil && *cmd.ParentID != "" {
		_, err := s.catRepo.GetByID(ctx, cmd.TenantID, *cmd.ParentID)
		if err != nil {
			return nil, domain.NewDomainError("INVALID_PARENT", "parent category not found", err)
		}
	}

	// Generate category ID
	categoryID := fmt.Sprintf("cat_%d", hashString(cmd.TenantID+cmd.Name))

	// Create new category
	cat := domain.NewCategory(categoryID, cmd.TenantID, cmd.Name)
	cat.ParentID = cmd.ParentID
	cat.Icon = cmd.Icon
	cat.Color = cmd.Color
	cat.SortOrder = cmd.SortOrder
	cat.CreatedByUserID = cmd.CreatedByUserID

	// Validate
	if err := cat.Validate(); err != nil {
		return nil, domain.NewDomainError("VALIDATION_ERROR", err.Error(), nil)
	}

	// Persist
	if err := s.catRepo.Create(ctx, cat); err != nil {
		return nil, domain.NewDomainError("CREATE_ERROR", "failed to create category", err)
	}

	s.logger.Info("Category created", "id", cat.ID, "tenant_id", cmd.TenantID)
	return CategoryToDTO(cat), nil
}

func (s *CategoryService) UpdateCategory(ctx context.Context, cmd UpdateCategoryCommand) (*CategoryDTO, error) {
	if cmd.TenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	cat, err := s.catRepo.GetByID(ctx, cmd.TenantID, cmd.ID)
	if err != nil {
		return nil, domain.NewDomainError("NOT_FOUND", "category not found", err)
	}

	// Update fields
	cat.Name = cmd.Name
	cat.Icon = cmd.Icon
	cat.Color = cmd.Color
	cat.SortOrder = cmd.SortOrder

	// Persist
	if err := s.catRepo.Update(ctx, cat); err != nil {
		return nil, domain.NewDomainError("UPDATE_ERROR", "failed to update category", err)
	}

	s.logger.Info("Category updated", "id", cmd.ID, "tenant_id", cmd.TenantID)
	return CategoryToDTO(cat), nil
}

func (s *CategoryService) GetCategory(ctx context.Context, tenantID, categoryID string) (*CategoryDTO, error) {
	if tenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	cat, err := s.catRepo.GetByID(ctx, tenantID, categoryID)
	if err != nil {
		return nil, domain.NewDomainError("NOT_FOUND", "category not found", err)
	}

	return CategoryToDTO(cat), nil
}

func (s *CategoryService) ListCategories(ctx context.Context, tenantID string) ([]*CategoryDTO, error) {
	if tenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	categories, err := s.catRepo.List(ctx, tenantID)
	if err != nil {
		return nil, domain.NewDomainError("QUERY_ERROR", "failed to list categories", err)
	}

	dtos := make([]*CategoryDTO, len(categories))
	for i, cat := range categories {
		dtos[i] = CategoryToDTO(cat)
	}

	return dtos, nil
}

func (s *CategoryService) DeleteCategory(ctx context.Context, tenantID, categoryID string) error {
	if tenantID == "" {
		return domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	// Verify exists
	_, err := s.catRepo.GetByID(ctx, tenantID, categoryID)
	if err != nil {
		return domain.NewDomainError("NOT_FOUND", "category not found", err)
	}

	if err := s.catRepo.Delete(ctx, tenantID, categoryID); err != nil {
		return domain.NewDomainError("DELETE_ERROR", "failed to delete category", err)
	}

	s.logger.Info("Category deleted", "id", categoryID, "tenant_id", tenantID)
	return nil
}

// GetCategoryRepo returns the category repository for use by other services
func (s *CategoryService) GetCategoryRepo() ports.CategoryRepository {
	return s.catRepo
}

package application

import (
	"context"
	"fmt"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/expense-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/expense-service/internal/ports"
)

type CategoryService struct {
	catRepo ports.CategoryRepository
	logger  *logger.Logger
}

func NewCategoryService(catRepo ports.CategoryRepository, logger *logger.Logger) *CategoryService {
	return &CategoryService{
		catRepo: catRepo,
		logger:  logger,
	}
}

func (s *CategoryService) CreateCategory(ctx context.Context, cmd CreateExpenseCategoryCommand) (*ExpenseCategoryDTO, error) {
	if cmd.TenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	catID := fmt.Sprintf("cat_%d", hashString(cmd.TenantID+cmd.Name))
	cat := domain.NewExpenseCategory(catID, cmd.TenantID, cmd.Name, cmd.SKR03Code, cmd.SKR04Code)
	cat.IsDefault = cmd.IsDefault

	if err := cat.Validate(); err != nil {
		return nil, domain.NewDomainError("VALIDATION_ERROR", err.Error(), nil)
	}

	if err := s.catRepo.CreateCategory(ctx, cat); err != nil {
		return nil, domain.NewDomainError("CREATE_FAILED", "failed to create category", err)
	}

	return CategoryToDTO(cat), nil
}

func (s *CategoryService) GetCategory(ctx context.Context, tenantID, categoryID string) (*ExpenseCategoryDTO, error) {
	if tenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	cat, err := s.catRepo.GetCategory(ctx, tenantID, categoryID)
	if err != nil {
		return nil, domain.NewDomainError("NOT_FOUND", "category not found", err)
	}

	return CategoryToDTO(cat), nil
}

func (s *CategoryService) ListCategories(ctx context.Context, tenantID string, limit, offset int) ([]*ExpenseCategoryDTO, int64, error) {
	if tenantID == "" {
		return nil, 0, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	categories, total, err := s.catRepo.ListCategories(ctx, tenantID, limit, offset)
	if err != nil {
		return nil, 0, domain.NewDomainError("LIST_FAILED", "failed to list categories", err)
	}

	dtos := make([]*ExpenseCategoryDTO, len(categories))
	for i, c := range categories {
		dtos[i] = CategoryToDTO(c)
	}

	return dtos, total, nil
}

func (s *CategoryService) UpdateCategory(ctx context.Context, tenantID, categoryID, name, skr03Code, skr04Code string) (*ExpenseCategoryDTO, error) {
	if tenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	cat, err := s.catRepo.GetCategory(ctx, tenantID, categoryID)
	if err != nil {
		return nil, domain.NewDomainError("NOT_FOUND", "category not found", err)
	}

	if name != "" {
		cat.Name = name
	}
	if skr03Code != "" {
		cat.SKR03Code = skr03Code
	}
	if skr04Code != "" {
		cat.SKR04Code = skr04Code
	}

	if err := s.catRepo.UpdateCategory(ctx, cat); err != nil {
		return nil, domain.NewDomainError("UPDATE_FAILED", "failed to update category", err)
	}

	return CategoryToDTO(cat), nil
}

func (s *CategoryService) DeleteCategory(ctx context.Context, tenantID, categoryID string) error {
	if tenantID == "" {
		return domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	if err := s.catRepo.DeleteCategory(ctx, tenantID, categoryID); err != nil {
		return domain.NewDomainError("DELETE_FAILED", "failed to delete category", err)
	}

	return nil
}

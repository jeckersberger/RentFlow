package application

import (
	"context"
	"fmt"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/expense-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/expense-service/internal/ports"
)

type BudgetService struct {
	budgetRepo ports.BudgetRepository
	logger     logger.Logger
}

func NewBudgetService(budgetRepo ports.BudgetRepository, logger logger.Logger) *BudgetService {
	return &BudgetService{
		budgetRepo: budgetRepo,
		logger:     logger,
	}
}

func (s *BudgetService) CreateBudget(ctx context.Context, cmd CreateBudgetCommand) (*BudgetDTO, error) {
	if cmd.TenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	budID := fmt.Sprintf("bud_%d", hashString(cmd.TenantID+cmd.CategoryID+cmd.ProjectID))
	period := domain.BudgetPeriod(cmd.Period)
	budget := domain.NewBudget(budID, cmd.TenantID, cmd.CategoryID, cmd.ProjectID, period, cmd.Amount)

	if err := budget.Validate(); err != nil {
		return nil, domain.NewDomainError("VALIDATION_ERROR", err.Error(), nil)
	}

	if err := s.budgetRepo.CreateBudget(ctx, budget); err != nil {
		return nil, domain.NewDomainError("CREATE_FAILED", "failed to create budget", err)
	}

	return BudgetToDTO(budget), nil
}

func (s *BudgetService) GetBudget(ctx context.Context, tenantID, budgetID string) (*BudgetDTO, error) {
	if tenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	budget, err := s.budgetRepo.GetBudget(ctx, tenantID, budgetID)
	if err != nil {
		return nil, domain.NewDomainError("NOT_FOUND", "budget not found", err)
	}

	return BudgetToDTO(budget), nil
}

func (s *BudgetService) ListBudgets(ctx context.Context, tenantID string, limit, offset int) ([]*BudgetDTO, int64, error) {
	if tenantID == "" {
		return nil, 0, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	budgets, total, err := s.budgetRepo.ListBudgets(ctx, tenantID, limit, offset)
	if err != nil {
		return nil, 0, domain.NewDomainError("LIST_FAILED", "failed to list budgets", err)
	}

	dtos := make([]*BudgetDTO, len(budgets))
	for i, b := range budgets {
		dtos[i] = BudgetToDTO(b)
	}

	return dtos, total, nil
}

func (s *BudgetService) GetBudgetStatus(ctx context.Context, tenantID, categoryID, period string) (*BudgetStatusDTO, error) {
	if tenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	budget, err := s.budgetRepo.GetByCategory(ctx, tenantID, categoryID, period)
	if err != nil {
		return nil, domain.NewDomainError("NOT_FOUND", "budget not found", err)
	}

	return BudgetToStatusDTO(budget), nil
}

func (s *BudgetService) DeleteBudget(ctx context.Context, tenantID, budgetID string) error {
	if tenantID == "" {
		return domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	if err := s.budgetRepo.DeleteBudget(ctx, tenantID, budgetID); err != nil {
		return domain.NewDomainError("DELETE_FAILED", "failed to delete budget", err)
	}

	return nil
}

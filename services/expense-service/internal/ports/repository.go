package ports

import (
	"context"

	"github.com/jeckersberger/rentflow/services/expense-service/internal/domain"
)

type ExpenseRepository interface {
	CreateExpense(ctx context.Context, expense *domain.Expense) error
	UpdateExpense(ctx context.Context, expense *domain.Expense) error
	GetExpense(ctx context.Context, tenantID, expenseID string) (*domain.Expense, error)
	ListExpenses(ctx context.Context, tenantID string, limit, offset int) ([]*domain.Expense, int64, error)
	DeleteExpense(ctx context.Context, tenantID, expenseID string) error
}

type CategoryRepository interface {
	CreateCategory(ctx context.Context, category *domain.ExpenseCategory) error
	UpdateCategory(ctx context.Context, category *domain.ExpenseCategory) error
	GetCategory(ctx context.Context, tenantID, categoryID string) (*domain.ExpenseCategory, error)
	ListCategories(ctx context.Context, tenantID string, limit, offset int) ([]*domain.ExpenseCategory, int64, error)
	DeleteCategory(ctx context.Context, tenantID, categoryID string) error
}

type BudgetRepository interface {
	CreateBudget(ctx context.Context, budget *domain.Budget) error
	GetBudget(ctx context.Context, tenantID, budgetID string) (*domain.Budget, error)
	ListBudgets(ctx context.Context, tenantID string, limit, offset int) ([]*domain.Budget, int64, error)
	GetByCategory(ctx context.Context, tenantID, categoryID, period string) (*domain.Budget, error)
	DeleteBudget(ctx context.Context, tenantID, budgetID string) error
}

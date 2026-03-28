package domain

import (
	"context"

	"github.com/google/uuid"
)

type ExpenseCategoryRepository interface {
	Create(ctx context.Context, category *ExpenseCategory) error
	GetByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*ExpenseCategory, error)
	List(ctx context.Context, tenantID uuid.UUID) ([]*ExpenseCategory, error)
	Update(ctx context.Context, category *ExpenseCategory) error
}

type ExpenseRepository interface {
	Create(ctx context.Context, expense *Expense) error
	GetByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*Expense, error)
	List(ctx context.Context, tenantID uuid.UUID, filter ExpenseFilter) ([]*Expense, int64, error)
	Update(ctx context.Context, expense *Expense) error
}

type ExpenseReceiptRepository interface {
	Create(ctx context.Context, receipt *ExpenseReceipt) error
	ListByExpense(ctx context.Context, expenseID uuid.UUID, tenantID uuid.UUID) ([]*ExpenseReceipt, error)
}

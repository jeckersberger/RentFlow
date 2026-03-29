package domain

import "errors"

var (
	ErrExpenseNotFound          = errors.New("expense not found")
	ErrCategoryNotFound         = errors.New("expense category not found")
	ErrReceiptNotFound          = errors.New("expense receipt not found")
	ErrAlreadyApproved          = errors.New("expense is already approved")
	ErrRecurringExpenseNotFound = errors.New("recurring expense not found")
	ErrBudgetNotFound           = errors.New("budget not found")
)

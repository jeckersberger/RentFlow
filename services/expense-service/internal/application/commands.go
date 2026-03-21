package application

import "time"

type CreateExpenseCommand struct {
	TenantID      string
	Vendor        string
	Amount        float64
	Currency      string
	TaxRate       float64
	CategoryCode  string
	Date          time.Time
	PaymentMethod string
	ReceiptRef    string
	ProjectID     string
	Notes         string
}

type UpdateExpenseCommand struct {
	ID            string
	TenantID      string
	Vendor        string
	Amount        float64
	CategoryCode  string
	PaymentMethod string
	Notes         string
}

type CreateExpenseCategoryCommand struct {
	TenantID  string
	Name      string
	SKR03Code string
	SKR04Code string
	IsDefault bool
}

type CreateBudgetCommand struct {
	TenantID   string
	CategoryID string
	ProjectID  string
	Period     string
	Amount     float64
}

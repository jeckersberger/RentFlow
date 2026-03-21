package application

import (
	"time"

	"github.com/jeckersberger/rentflow/services/expense-service/internal/domain"
)

type ExpenseDTO struct {
	ID            string            `json:"id"`
	TenantID      string            `json:"tenant_id"`
	Vendor        string            `json:"vendor"`
	Amount        float64           `json:"amount"`
	Currency      string            `json:"currency"`
	TaxRate       float64           `json:"tax_rate"`
	TaxAmount     float64           `json:"tax_amount"`
	NetAmount     float64           `json:"net_amount"`
	CategoryCode  string            `json:"category_code"`
	Date          time.Time         `json:"date"`
	PaymentMethod string            `json:"payment_method"`
	ReceiptRef    string            `json:"receipt_ref"`
	OCRData       map[string]string `json:"ocr_data"`
	Status        string            `json:"status"`
	ProjectID     string            `json:"project_id"`
	ApprovedBy    string            `json:"approved_by"`
	Notes         string            `json:"notes"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
}

type ExpenseCategoryDTO struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id"`
	Name      string    `json:"name"`
	SKR03Code string    `json:"skr03_code"`
	SKR04Code string    `json:"skr04_code"`
	IsDefault bool      `json:"is_default"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type BudgetDTO struct {
	ID         string    `json:"id"`
	TenantID   string    `json:"tenant_id"`
	CategoryID string    `json:"category_id"`
	ProjectID  string    `json:"project_id"`
	Period     string    `json:"period"`
	Amount     float64   `json:"amount"`
	Spent      float64   `json:"spent"`
	Remaining  float64   `json:"remaining"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type BudgetStatusDTO struct {
	BudgetID   string  `json:"budget_id"`
	CategoryID string  `json:"category_id"`
	Amount     float64 `json:"amount"`
	Spent      float64 `json:"spent"`
	Remaining  float64 `json:"remaining"`
	IsOver     bool    `json:"is_over"`
	Percentage float64 `json:"percentage"`
}

func ExpenseToDTO(e *domain.Expense) *ExpenseDTO {
	return &ExpenseDTO{
		ID:            e.ID,
		TenantID:      e.TenantID,
		Vendor:        e.Vendor,
		Amount:        e.Amount,
		Currency:      e.Currency,
		TaxRate:       e.TaxRate,
		TaxAmount:     e.TaxAmount,
		NetAmount:     e.NetAmount,
		CategoryCode:  e.CategoryCode,
		Date:          e.Date,
		PaymentMethod: e.PaymentMethod,
		ReceiptRef:    e.ReceiptRef,
		OCRData:       e.OCRData,
		Status:        string(e.Status),
		ProjectID:     e.ProjectID,
		ApprovedBy:    e.ApprovedBy,
		Notes:         e.Notes,
		CreatedAt:     e.CreatedAt,
		UpdatedAt:     e.UpdatedAt,
	}
}

func CategoryToDTO(c *domain.ExpenseCategory) *ExpenseCategoryDTO {
	return &ExpenseCategoryDTO{
		ID:        c.ID,
		TenantID:  c.TenantID,
		Name:      c.Name,
		SKR03Code: c.SKR03Code,
		SKR04Code: c.SKR04Code,
		IsDefault: c.IsDefault,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

func BudgetToDTO(b *domain.Budget) *BudgetDTO {
	return &BudgetDTO{
		ID:         b.ID,
		TenantID:   b.TenantID,
		CategoryID: b.CategoryID,
		ProjectID:  b.ProjectID,
		Period:     string(b.Period),
		Amount:     b.Amount,
		Spent:      b.Spent,
		Remaining:  b.RemainingBudget(),
		CreatedAt:  b.CreatedAt,
		UpdatedAt:  b.UpdatedAt,
	}
}

func BudgetToStatusDTO(b *domain.Budget) *BudgetStatusDTO {
	percentage := 0.0
	if b.Amount > 0 {
		percentage = (b.Spent / b.Amount) * 100
	}
	return &BudgetStatusDTO{
		BudgetID:   b.ID,
		CategoryID: b.CategoryID,
		Amount:     b.Amount,
		Spent:      b.Spent,
		Remaining:  b.RemainingBudget(),
		IsOver:     b.IsOverBudget(),
		Percentage: percentage,
	}
}

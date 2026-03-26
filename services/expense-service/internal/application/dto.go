package application

import (
	"time"

	"github.com/jeckersberger/rentflow/services/expense-service/internal/domain"
)

type ExpenseDTO struct {
	ID            string            `json:"id"`
	TenantID      string            `json:"tenant_id"`
	Type          string            `json:"type"`
	Vendor        string            `json:"vendor"`
	VendorAddress string            `json:"vendor_address,omitempty"`
	VendorVATID   string            `json:"vendor_vat_id,omitempty"`
	VendorIBAN    string            `json:"vendor_iban,omitempty"`
	Amount        float64           `json:"amount"`
	Currency      string            `json:"currency"`
	TaxRate       float64           `json:"tax_rate"`
	TaxAmount     float64           `json:"tax_amount"`
	NetAmount     float64           `json:"net_amount"`
	CategoryCode  string            `json:"category_code"`
	BookingAccount string           `json:"booking_account,omitempty"`
	Date          time.Time         `json:"date"`
	ServicePeriodFrom *time.Time    `json:"service_period_from,omitempty"`
	ServicePeriodTo   *time.Time    `json:"service_period_to,omitempty"`
	DueDate       *time.Time        `json:"due_date,omitempty"`
	DiscountPercent float64         `json:"discount_percent,omitempty"`
	DiscountDays    int             `json:"discount_days,omitempty"`
	DiscountDeadline *time.Time     `json:"discount_deadline,omitempty"`
	PaymentMethod string            `json:"payment_method"`
	PaymentStatus string            `json:"payment_status"`
	PaidAt        *time.Time        `json:"paid_at,omitempty"`
	InvoiceNumber string            `json:"invoice_number,omitempty"`
	ReceiptRef    string            `json:"receipt_ref"`
	ReceiptChecksum string          `json:"receipt_checksum,omitempty"`
	ReceiptNASPath  string          `json:"receipt_nas_path,omitempty"`
	OCRData       map[string]string `json:"ocr_data"`
	OCRConfidence float64           `json:"ocr_confidence,omitempty"`
	Source        string            `json:"source"`
	EmailRef      string            `json:"email_ref,omitempty"`
	Status        string            `json:"status"`
	ProjectID     string            `json:"project_id"`
	ApprovedBy    string            `json:"approved_by"`
	Notes         string            `json:"notes"`
	// Bewirtungsbeleg
	EntertainmentLocation string   `json:"entertainment_location,omitempty"`
	EntertainmentReason   string   `json:"entertainment_reason,omitempty"`
	EntertainmentGuests   string   `json:"entertainment_guests,omitempty"`
	EntertainmentTip      float64  `json:"entertainment_tip,omitempty"`
	EntertainmentDeductible float64 `json:"entertainment_deductible,omitempty"`
	EntertainmentNonDeductible float64 `json:"entertainment_non_deductible,omitempty"`
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
		ID:                  e.ID,
		TenantID:            e.TenantID,
		Type:                string(e.Type),
		Vendor:              e.Vendor,
		VendorAddress:       e.VendorAddress,
		VendorVATID:         e.VendorVATID,
		VendorIBAN:          e.VendorIBAN,
		Amount:              e.Amount,
		Currency:            e.Currency,
		TaxRate:             e.TaxRate,
		TaxAmount:           e.TaxAmount,
		NetAmount:           e.NetAmount,
		CategoryCode:        e.CategoryCode,
		BookingAccount:      e.BookingAccount,
		Date:                e.Date,
		ServicePeriodFrom:   e.ServicePeriodFrom,
		ServicePeriodTo:     e.ServicePeriodTo,
		DueDate:             e.DueDate,
		DiscountPercent:     e.DiscountPercent,
		DiscountDays:        e.DiscountDays,
		DiscountDeadline:    e.DiscountDeadline,
		PaymentMethod:       e.PaymentMethod,
		PaymentStatus:       string(e.PaymentStatus),
		PaidAt:              e.PaidAt,
		InvoiceNumber:       e.InvoiceNumber,
		ReceiptRef:          e.ReceiptRef,
		ReceiptChecksum:     e.ReceiptChecksum,
		ReceiptNASPath:      e.ReceiptNASPath,
		OCRData:             e.OCRData,
		OCRConfidence:       e.OCRConfidence,
		Source:              e.Source,
		EmailRef:            e.EmailRef,
		Status:              string(e.Status),
		ProjectID:           e.ProjectID,
		ApprovedBy:          e.ApprovedBy,
		Notes:               e.Notes,
		EntertainmentLocation: e.EntertainmentLocation,
		EntertainmentReason:   e.EntertainmentReason,
		EntertainmentGuests:   e.EntertainmentGuests,
		EntertainmentTip:      e.EntertainmentTip,
		EntertainmentDeductible: e.EntertainmentDeductible,
		EntertainmentNonDeductible: e.EntertainmentNonDeductible,
		CreatedAt:           e.CreatedAt,
		UpdatedAt:           e.UpdatedAt,
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

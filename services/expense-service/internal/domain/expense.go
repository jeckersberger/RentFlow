package domain

import (
	"fmt"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/events"
)

type ExpenseStatus string

const (
	StatusDraft    ExpenseStatus = "draft"
	StatusPending  ExpenseStatus = "pending"
	StatusApproved ExpenseStatus = "approved"
	StatusRejected ExpenseStatus = "rejected"
	StatusExported ExpenseStatus = "exported"
)

type Expense struct {
	events.AggregateRoot
	TenantID      string
	Vendor        string
	Amount        float64
	Currency      string
	TaxRate       float64
	TaxAmount     float64
	NetAmount     float64
	CategoryCode  string
	Date          time.Time
	PaymentMethod string
	ReceiptRef    string
	OCRData       map[string]string
	Status        ExpenseStatus
	ProjectID     string
	ApprovedBy    string
	Notes         string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func NewExpense(id, tenantID, vendor, currency, categoryCode, paymentMethod string, amount float64, date time.Time) *Expense {
	now := time.Now()
	return &Expense{
		AggregateRoot: *events.NewAggregateRoot(id, "expense"),
		TenantID:      tenantID,
		Vendor:        vendor,
		Amount:        amount,
		Currency:      currency,
		TaxRate:       0.19,
		CategoryCode:  categoryCode,
		Date:          date,
		PaymentMethod: paymentMethod,
		Status:        StatusDraft,
		OCRData:       make(map[string]string),
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

func (e *Expense) CalculateTax() {
	e.TaxAmount = e.Amount * e.TaxRate
	e.NetAmount = e.Amount - e.TaxAmount
}

func (e *Expense) Approve(approvedBy string) error {
	if e.Status != StatusPending {
		return fmt.Errorf("can only approve pending expenses")
	}
	e.Status = StatusApproved
	e.ApprovedBy = approvedBy
	e.UpdatedAt = time.Now()
	return nil
}

func (e *Expense) Reject() error {
	if e.Status != StatusPending {
		return fmt.Errorf("can only reject pending expenses")
	}
	e.Status = StatusRejected
	e.UpdatedAt = time.Now()
	return nil
}

func (e *Expense) Submit() error {
	if e.Status != StatusDraft {
		return fmt.Errorf("can only submit draft expenses")
	}
	e.Status = StatusPending
	e.UpdatedAt = time.Now()
	return nil
}

func (e *Expense) Validate() error {
	if e.ID == "" {
		return fmt.Errorf("expense ID cannot be empty")
	}
	if e.TenantID == "" {
		return fmt.Errorf("tenant ID cannot be empty")
	}
	if e.Vendor == "" {
		return fmt.Errorf("vendor cannot be empty")
	}
	if e.Amount <= 0 {
		return fmt.Errorf("amount must be positive")
	}
	return nil
}

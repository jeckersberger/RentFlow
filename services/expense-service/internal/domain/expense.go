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

type ExpenseType string

const (
	ExpenseTypeInvoice      ExpenseType = "invoice"
	ExpenseTypeReceipt      ExpenseType = "receipt"
	ExpenseTypeEntertainment ExpenseType = "entertainment"
)

type PaymentStatus string

const (
	PaymentStatusOpen     PaymentStatus = "open"
	PaymentStatusPaid     PaymentStatus = "paid"
	PaymentStatusOverdue  PaymentStatus = "overdue"
)

type Expense struct {
	events.AggregateRoot
	TenantID      string
	Type          ExpenseType
	Vendor        string
	VendorAddress string
	VendorVATID   string // USt-IdNr / Steuernummer
	VendorIBAN    string
	Amount        float64
	Currency      string
	TaxRate       float64
	TaxAmount     float64
	NetAmount     float64
	CategoryCode  string
	BookingAccount string // SKR03/04 Buchungskonto
	Date          time.Time
	ServicePeriodFrom *time.Time
	ServicePeriodTo   *time.Time
	DueDate       *time.Time // Zahlungsziel
	DiscountPercent float64   // Skonto %
	DiscountDays    int       // Skonto Frist in Tagen
	DiscountDeadline *time.Time
	PaymentMethod string
	PaymentStatus PaymentStatus
	PaidAt        *time.Time
	InvoiceNumber string     // Rechnungsnummer
	ReceiptRef    string
	ReceiptChecksum string   // SHA-256 des Originals
	ReceiptNASPath  string   // Pfad auf NAS
	OCRData       map[string]string
	OCRConfidence float64    // KI-Konfidenz 0-1
	Source        string     // "manual", "email_auto", "photo"
	EmailRef      string     // Referenz zur Original-Mail
	Status        ExpenseStatus
	ProjectID     string
	ApprovedBy    string
	Notes         string
	// Bewirtungsbeleg-Felder (§4 Abs. 5 EStG)
	EntertainmentLocation string
	EntertainmentReason   string
	EntertainmentGuests   string  // Komma-separiert
	EntertainmentTip      float64
	EntertainmentDeductible float64 // 70% vom Betrag
	EntertainmentNonDeductible float64 // 30%
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func NewExpense(id, tenantID, vendor, currency, categoryCode, paymentMethod string, amount float64, date time.Time) *Expense {
	now := time.Now()
	return &Expense{
		AggregateRoot: *events.NewAggregateRoot(id, "expense"),
		TenantID:      tenantID,
		Type:          ExpenseTypeInvoice,
		Vendor:        vendor,
		Amount:        amount,
		Currency:      currency,
		TaxRate:       0.19,
		CategoryCode:  categoryCode,
		Date:          date,
		PaymentMethod: paymentMethod,
		PaymentStatus: PaymentStatusOpen,
		Status:        StatusDraft,
		Source:        "manual",
		OCRData:       make(map[string]string),
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

// CalculateEntertainmentSplit berechnet die 70/30-Aufteilung fuer Bewirtungsbelege
func (e *Expense) CalculateEntertainmentSplit() {
	if e.Type != ExpenseTypeEntertainment {
		return
	}
	base := e.Amount - e.EntertainmentTip
	e.EntertainmentDeductible = base * 0.70
	e.EntertainmentNonDeductible = base * 0.30
}

// CalculateDiscountDeadline berechnet die Skonto-Frist
func (e *Expense) CalculateDiscountDeadline() {
	if e.DiscountDays > 0 && e.Date != (time.Time{}) {
		deadline := e.Date.AddDate(0, 0, e.DiscountDays)
		e.DiscountDeadline = &deadline
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

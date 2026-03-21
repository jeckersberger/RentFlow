package application

import (
	"time"

	"github.com/jeckersberger/rentflow/services/invoice-service/internal/domain"
)

// Invoice Commands

type CreateInvoiceCommand struct {
	TenantID      string
	ProjectID     *string
	ClientName    string
	ClientAddress domain.Address
	ClientEmail   string
	ClientTaxID   string
	Items         []CreateInvoiceItemCommand
	TaxRate       float64
	Currency      string
	IssueDate     time.Time
	DueDate       time.Time
	Notes         string
	InternalNotes string
	CreatedByUserID string
}

type CreateInvoiceItemCommand struct {
	Description string
	Quantity    float64
	Unit        string
	UnitPrice   float64
	EquipmentID *string
}

type UpdateInvoiceCommand struct {
	ID            string
	TenantID      string
	ClientName    string
	ClientAddress domain.Address
	ClientEmail   string
	ClientTaxID   string
	TaxRate       float64
	DueDate       time.Time
	Notes         string
	InternalNotes string
}

type SendInvoiceCommand struct {
	ID       string
	TenantID string
	Email    string
}

type MarkInvoicePaidCommand struct {
	ID             string
	TenantID       string
	PaymentMethod  string
	PaymentRef     string
	PaymentDate    *time.Time
}

type CancelInvoiceCommand struct {
	ID       string
	TenantID string
	Reason   string
}

type CreditInvoiceCommand struct {
	ID       string
	TenantID string
	CreditID string
}

type AddInvoiceItemCommand struct {
	InvoiceID   string
	TenantID    string
	Description string
	Quantity    float64
	Unit        string
	UnitPrice   float64
	EquipmentID *string
}

type RemoveInvoiceItemCommand struct {
	InvoiceID string
	TenantID  string
	ItemID    string
}

type GenerateInvoicePDFCommand struct {
	ID       string
	TenantID string
}

// Quote Commands

type CreateQuoteCommand struct {
	TenantID      string
	ProjectID     *string
	ClientName    string
	ClientAddress domain.Address
	ClientEmail   string
	Items         []CreateInvoiceItemCommand
	TaxRate       float64
	Currency      string
	ValidDays     int
	Notes         string
	CreatedByUserID string
}

type UpdateQuoteCommand struct {
	ID            string
	TenantID      string
	ClientName    string
	ClientAddress domain.Address
	ClientEmail   string
	TaxRate       float64
	ValidUntil    time.Time
	Notes         string
}

type SendQuoteCommand struct {
	ID       string
	TenantID string
	Email    string
}

type AcceptQuoteCommand struct {
	ID       string
	TenantID string
}

type RejectQuoteCommand struct {
	ID       string
	TenantID string
	Reason   string
}

type ConvertQuoteToInvoiceCommand struct {
	QuoteID     string
	TenantID    string
	IssueDate   time.Time
	DueDate     time.Time
	InternalNotes string
}

type AddQuoteItemCommand struct {
	QuoteID     string
	TenantID    string
	Description string
	Quantity    float64
	Unit        string
	UnitPrice   float64
	EquipmentID *string
}

type RemoveQuoteItemCommand struct {
	QuoteID  string
	TenantID string
	ItemID   string
}

// Dunning Commands

type CreateDunningCommand struct {
	TenantID    string
	InvoiceID   string
	Level       int
	DaysToAdd   int
	CustomFee   *float64
}

type SendDunningCommand struct {
	ID       string
	TenantID string
	Email    string
}

type CreateDunningReminderCommand struct {
	TenantID  string
	InvoiceID string
}

package application

import (
	"time"

	"github.com/jeckersberger/rentflow/services/invoice-service/internal/domain"
)

// Invoice Commands

type CreateInvoiceCommand struct {
	TenantID           string              `json:"tenant_id"`
	ProjectID          *string             `json:"project_id"`
	ClientName         string              `json:"client_name"`
	ClientAddress      domain.Address      `json:"client_address"`
	ClientEmail        string              `json:"client_email"`
	ClientTaxID        string              `json:"client_tax_id"`
	Items              []CreateInvoiceItemCommand `json:"items"`
	TaxRate            float64             `json:"tax_rate"`
	IsKleinunternehmer bool                `json:"is_kleinunternehmer"`
	Currency           string              `json:"currency"`
	IssueDate          time.Time           `json:"issue_date"`
	DueDate            time.Time           `json:"due_date"`
	Notes              string              `json:"notes"`
	InternalNotes      string              `json:"internal_notes"`
	CreatedByUserID    string              `json:"created_by_user_id"`
}

type CreateInvoiceItemCommand struct {
	Description string  `json:"description"`
	Quantity    float64 `json:"quantity"`
	Unit        string  `json:"unit"`
	UnitPrice   float64 `json:"unit_price"`
	TaxRate     float64 `json:"tax_rate"`
	EquipmentID *string `json:"equipment_id"`
}

type UpdateInvoiceCommand struct {
	ID            string         `json:"id"`
	TenantID      string         `json:"tenant_id"`
	ClientName    string         `json:"client_name"`
	ClientAddress domain.Address `json:"client_address"`
	ClientEmail   string         `json:"client_email"`
	ClientTaxID   string         `json:"client_tax_id"`
	TaxRate       float64        `json:"tax_rate"`
	DueDate       time.Time      `json:"due_date"`
	Notes         string         `json:"notes"`
	InternalNotes string         `json:"internal_notes"`
}

type SendInvoiceCommand struct {
	ID       string
	TenantID string
	Email    string
}

type MarkInvoicePaidCommand struct {
	ID            string
	TenantID      string
	PaymentMethod string
	PaymentRef    string
	PaymentDate   *time.Time
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

type RecordPaymentCommand struct {
	ID       string
	TenantID string
	Amount   float64
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

// Credit Note Commands

type CreateCreditNoteCommand struct {
	InvoiceID   string
	TenantID    string
	Reason      string
	Items       []CreateCreditNoteItemCommand
	TaxRate     float64
	Currency    string
}

type CreateCreditNoteItemCommand struct {
	Description string
	Quantity    float64
	Unit        string
	UnitPrice   float64
}

// Quote Commands

type CreateQuoteCommand struct {
	TenantID        string
	ProjectID       *string
	ClientName      string
	ClientAddress   domain.Address
	ClientEmail     string
	Items           []CreateInvoiceItemCommand
	TaxRate         float64
	Currency        string
	ValidDays       int
	Notes           string
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

type ConfirmQuoteCommand struct {
	ID       string
	TenantID string
}

type RejectQuoteCommand struct {
	ID       string
	TenantID string
	Reason   string
}

type ConvertQuoteToInvoiceCommand struct {
	QuoteID       string
	TenantID      string
	IssueDate     time.Time
	DueDate       time.Time
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
	TenantID  string
	InvoiceID string
	Level     int
	DaysToAdd int
	CustomFee *float64
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

package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	InvoiceTypeInvoice    = "invoice"
	InvoiceTypePartial    = "partial"
	InvoiceTypeAdvance    = "advance"
	InvoiceTypeCreditNote = "credit_note"
	InvoiceTypeReversal   = "reversal"
	InvoiceTypeProforma   = "proforma"
)

const (
	StatusDraft       = "draft"
	StatusFinalized   = "finalized"
	StatusSent        = "sent"
	StatusPartialPaid = "partial_paid"
	StatusPaid        = "paid"
	StatusOverdue     = "overdue"
	StatusCancelled   = "cancelled"
)

type Invoice struct {
	ID               uuid.UUID  `json:"id"`
	TenantID         uuid.UUID  `json:"tenant_id"`
	InvoiceNumber    string     `json:"invoice_number"`
	InvoiceType      string     `json:"invoice_type"`
	Status           string     `json:"status"`
	CustomerName     string     `json:"customer_name"`
	CustomerEmail    string     `json:"customer_email"`
	CustomerAddress  string     `json:"customer_address"`
	InvoiceDate      string     `json:"invoice_date"`
	DueDate          string     `json:"due_date"`
	VatRate          int64      `json:"vat_rate"`
	Kleinunternehmer bool       `json:"kleinunternehmer"`
	TotalNet         int64      `json:"total_net"`
	TotalVat         int64      `json:"total_vat"`
	TotalGross       int64      `json:"total_gross"`
	AmountPaid       int64      `json:"amount_paid"`
	Notes            string     `json:"notes"`
	Hash             string     `json:"hash"`
	PreviousHash     string     `json:"previous_hash"`
	FinalizedAt      *time.Time `json:"finalized_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type InvoiceItem struct {
	ID          uuid.UUID `json:"id"`
	TenantID    uuid.UUID `json:"tenant_id"`
	InvoiceID   uuid.UUID `json:"invoice_id"`
	Description string    `json:"description"`
	Quantity    int64     `json:"quantity"`
	Unit        string    `json:"unit"`
	UnitPrice   int64     `json:"unit_price"`
	VatRate     int64     `json:"vat_rate"`
	Position    int       `json:"position"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Payment struct {
	ID            uuid.UUID `json:"id"`
	TenantID      uuid.UUID `json:"tenant_id"`
	InvoiceID     uuid.UUID `json:"invoice_id"`
	Amount        int64     `json:"amount"`
	PaymentDate   string    `json:"payment_date"`
	PaymentMethod string    `json:"payment_method"`
	Reference     string    `json:"reference"`
	CreatedAt     time.Time `json:"created_at"`
}

type NumberSequence struct {
	ID         uuid.UUID `json:"id"`
	TenantID   uuid.UUID `json:"tenant_id"`
	Prefix     string    `json:"prefix"`
	Year       int       `json:"year"`
	LastNumber int64     `json:"last_number"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type InvoiceFilter struct {
	Page    int    `json:"page"`
	PerPage int    `json:"per_page"`
	Status  string `json:"status,omitempty"`
}

// ---------------------------------------------------------------------------
// Quote (Angebot) domain
// ---------------------------------------------------------------------------

const (
	QuoteStatusDraft     = "draft"
	QuoteStatusSent      = "sent"
	QuoteStatusAccepted  = "accepted"
	QuoteStatusDeclined  = "declined"
	QuoteStatusExpired   = "expired"
	QuoteStatusCancelled = "cancelled"
)

type Quote struct {
	ID                 uuid.UUID  `json:"id"`
	TenantID           uuid.UUID  `json:"tenant_id"`
	QuoteNumber        string     `json:"quote_number"`
	Status             string     `json:"status"`
	CustomerName       string     `json:"customer_name"`
	CustomerEmail      string     `json:"customer_email"`
	CustomerAddress    string     `json:"customer_address"`
	ProjectID          *uuid.UUID `json:"project_id,omitempty"`
	Subject            string     `json:"subject"`
	IntroText          string     `json:"intro_text"`
	OutroText          string     `json:"outro_text"`
	QuoteDate          string     `json:"quote_date"`
	ValidUntil         string     `json:"valid_until"`
	VatRate            int64      `json:"vat_rate"`
	Kleinunternehmer   bool       `json:"kleinunternehmer"`
	TotalNet           int64      `json:"total_net"`
	TotalVat           int64      `json:"total_vat"`
	TotalGross         int64      `json:"total_gross"`
	PaymentTermsDays   int        `json:"payment_terms_days"`
	DiscountPct        int64      `json:"discount_pct"`
	Notes              string     `json:"notes"`
	ConvertedInvoiceID *uuid.UUID `json:"converted_invoice_id,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

type QuoteItem struct {
	ID          uuid.UUID `json:"id"`
	TenantID    uuid.UUID `json:"tenant_id"`
	QuoteID     uuid.UUID `json:"quote_id"`
	Description string    `json:"description"`
	Quantity    int64     `json:"quantity"`
	Unit        string    `json:"unit"`
	UnitPrice   int64     `json:"unit_price"`
	DiscountPct int64     `json:"discount_pct"`
	Position    int       `json:"position"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type QuoteFilter struct {
	Page    int
	PerPage int
	Status  string
}

// ---------------------------------------------------------------------------
// Dunning (Mahnwesen) domain
// ---------------------------------------------------------------------------

const (
	DunningLevelReminder  = "reminder"
	DunningLevelDunning1  = "dunning_1"
	DunningLevelDunning2  = "dunning_2"
	DunningLevelDunning3  = "dunning_3"
)

const (
	DunningStatusPending   = "pending"
	DunningStatusSent      = "sent"
	DunningStatusCancelled = "cancelled"
)

type DunningEntry struct {
	ID        uuid.UUID  `json:"id"`
	TenantID  uuid.UUID  `json:"tenant_id"`
	InvoiceID uuid.UUID  `json:"invoice_id"`
	Level     string     `json:"level"`
	FeeCents  int64      `json:"fee_cents"`
	SentAt    *time.Time `json:"sent_at,omitempty"`
	Status    string     `json:"status"`
	Notes     string     `json:"notes"`
	CreatedAt time.Time  `json:"created_at"`
}

type DunningConfig struct {
	ID           uuid.UUID `json:"id"`
	TenantID     uuid.UUID `json:"tenant_id"`
	ReminderDays int       `json:"reminder_days"`
	Dunning1Days int       `json:"dunning1_days"`
	Dunning2Days int       `json:"dunning2_days"`
	ReminderFee  int64     `json:"reminder_fee"`
	Dunning1Fee  int64     `json:"dunning1_fee"`
	Dunning2Fee  int64     `json:"dunning2_fee"`
	AutoSend     bool      `json:"auto_send"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// OverdueInvoice wraps an invoice with its suggested dunning level and fee.
type OverdueInvoice struct {
	Invoice        *Invoice `json:"invoice"`
	DaysOverdue    int      `json:"days_overdue"`
	SuggestedLevel string   `json:"suggested_level"`
	SuggestedFee   int64    `json:"suggested_fee"`
}

// ---------------------------------------------------------------------------
// Bank Transaction domain
// ---------------------------------------------------------------------------

type BankTransaction struct {
	ID               uuid.UUID  `json:"id"`
	TenantID         uuid.UUID  `json:"tenant_id"`
	BookingDate      string     `json:"booking_date"`
	ValueDate        string     `json:"value_date"`
	Amount           int64      `json:"amount"`
	Currency         string     `json:"currency"`
	Reference        string     `json:"reference"`
	CounterpartyName string     `json:"counterparty_name"`
	CounterpartyIBAN string     `json:"counterparty_iban"`
	MatchedInvoiceID *uuid.UUID `json:"matched_invoice_id,omitempty"`
	MatchConfidence  string     `json:"match_confidence"`
	ImportSource     string     `json:"import_source"`
	ImportBatchID    *uuid.UUID `json:"import_batch_id,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
}

package domain

import "time"

// Invoice Events

type InvoiceCreatedEvent struct {
	TenantID       string
	InvoiceNumber  string
	ClientName     string
	ClientEmail    string
	Total          float64
	Currency       string
	CreatedAt      time.Time
}

type InvoiceSentEvent struct {
	TenantID      string
	InvoiceNumber string
	SentAt        time.Time
	SentTo        string
}

type InvoicePaidEvent struct {
	TenantID      string
	InvoiceNumber string
	PaidAt        time.Time
	PaymentMethod string
	PaymentRef    string
}

type InvoiceCancelledEvent struct {
	TenantID      string
	InvoiceNumber string
	CancelledAt   time.Time
	Reason        string
}

type InvoiceCreditedEvent struct {
	TenantID      string
	InvoiceNumber string
	CreditedAt    time.Time
	CreditNoteID  string
}

// Quote Events

type QuoteCreatedEvent struct {
	TenantID   string
	QuoteNumber string
	ClientName string
	Total      float64
	ValidUntil time.Time
	CreatedAt  time.Time
}

type QuoteSentEvent struct {
	TenantID   string
	QuoteNumber string
	SentAt     time.Time
	SentTo     string
}

type QuoteAcceptedEvent struct {
	TenantID    string
	QuoteNumber string
	AcceptedAt  time.Time
}

type QuoteRejectedEvent struct {
	TenantID    string
	QuoteNumber string
	RejectedAt  time.Time
	Reason      string
}

type QuoteExpiredEvent struct {
	TenantID    string
	QuoteNumber string
	ExpiredAt   time.Time
}

type QuoteConvertedToInvoiceEvent struct {
	TenantID      string
	QuoteNumber   string
	InvoiceNumber string
	ConvertedAt   time.Time
}

// Dunning Events

type DunningCreatedEvent struct {
	TenantID      string
	DunningID     string
	InvoiceNumber string
	Level         int
	DueDate       time.Time
	Fee           float64
	CreatedAt     time.Time
}

type DunningSentEvent struct {
	TenantID      string
	DunningID     string
	InvoiceNumber string
	SentAt        time.Time
	SentTo        string
	Level         int
}

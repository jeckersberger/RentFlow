package domain

import (
	"fmt"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/events"
)

// QuoteStatus represents the state of a quote
type QuoteStatus string

const (
	QuoteDraft    QuoteStatus = "draft"
	QuoteSent     QuoteStatus = "sent"
	QuoteAccepted QuoteStatus = "accepted"
	QuoteRejected QuoteStatus = "rejected"
	QuoteExpired  QuoteStatus = "expired"
)

// IsValidStatus checks if status is valid
func (s QuoteStatus) IsValidStatus() bool {
	switch s {
	case QuoteDraft, QuoteSent, QuoteAccepted, QuoteRejected, QuoteExpired:
		return true
	default:
		return false
	}
}

// Quote represents an estimate/Angebot
type Quote struct {
	events.AggregateRoot
	TenantID    string
	QuoteNumber string
	ProjectID   *string
	ClientName  string
	ClientEmail string
	ClientAddress Address
	Items       []InvoiceItem // reuse same item structure
	SubTotal    float64
	TaxRate     float64
	TaxAmount   float64
	Total       float64
	Currency    string
	Status      QuoteStatus
	ValidUntil  time.Time
	Notes       string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewQuote creates a new quote aggregate
func NewQuote(id, tenantID, quoteNumber, clientName, clientEmail string, validDays int) *Quote {
	now := time.Now()
	return &Quote{
		AggregateRoot: *events.NewAggregateRoot(id, "quote"),
		TenantID:      tenantID,
		QuoteNumber:   quoteNumber,
		ClientName:    clientName,
		ClientEmail:   clientEmail,
		Status:        QuoteDraft,
		Currency:      "EUR",
		TaxRate:       19, // Default 19% VAT
		ValidUntil:    now.AddDate(0, 0, validDays),
		Items:         []InvoiceItem{},
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

// AddItem adds a line item to the quote
func (q *Quote) AddItem(item InvoiceItem) error {
	if item.Description == "" {
		return ErrInvalidInput
	}
	if item.Quantity <= 0 || item.UnitPrice < 0 {
		return ErrInvalidInput
	}

	item.TotalPrice = item.Quantity * item.UnitPrice
	item.TaxRate = q.TaxRate
	item.ID = fmt.Sprintf("item_%d", len(q.Items)+1)

	q.Items = append(q.Items, item)
	q.UpdatedAt = time.Now()
	return nil
}

// RemoveItem removes a line item from the quote
func (q *Quote) RemoveItem(itemID string) error {
	for idx, item := range q.Items {
		if item.ID == itemID {
			q.Items = append(q.Items[:idx], q.Items[idx+1:]...)
			q.UpdatedAt = time.Now()
			return nil
		}
	}
	return ErrInvalidInput
}

// SetTaxRate sets the VAT rate
func (q *Quote) SetTaxRate(rate float64) error {
	switch rate {
	case 0, 7, 19:
		q.TaxRate = rate
		q.UpdatedAt = time.Now()
		return nil
	default:
		return ErrInvalidTaxRate
	}
}

// CalculateTotals recalculates SubTotal, TaxAmount, and Total
func (q *Quote) CalculateTotals() error {
	if len(q.Items) == 0 {
		return ErrNoItems
	}

	q.SubTotal = 0
	for _, item := range q.Items {
		q.SubTotal += item.TotalPrice
	}

	q.TaxAmount = q.SubTotal * (q.TaxRate / 100)
	q.Total = q.SubTotal + q.TaxAmount

	if q.Total <= 0 {
		return ErrZeroAmount
	}

	q.UpdatedAt = time.Now()
	return nil
}

// Validate checks the quote is valid
func (q *Quote) Validate() error {
	if q.TenantID == "" {
		return ErrTenantIDRequired
	}
	if q.QuoteNumber == "" {
		return ErrInvalidInput
	}
	if q.ClientName == "" {
		return ErrInvalidInput
	}
	if len(q.Items) == 0 {
		return ErrNoItems
	}
	if !q.TaxRate.IsValidTaxRate() {
		return ErrInvalidTaxRate
	}
	if q.Total <= 0 {
		return ErrZeroAmount
	}
	if !q.Status.IsValidStatus() {
		return ErrInvalidStatus
	}
	return nil
}

// Send transitions quote from Draft to Sent
func (q *Quote) Send(email string) error {
	if q.Status != QuoteDraft {
		return ErrInvalidTransition
	}
	q.Status = QuoteSent
	q.UpdatedAt = time.Now()

	event := QuoteSentEvent{
		TenantID:    q.TenantID,
		QuoteNumber: q.QuoteNumber,
		SentAt:      q.UpdatedAt,
		SentTo:      email,
	}
	q.Apply(event)
	return nil
}

// Accept transitions quote to Accepted
func (q *Quote) Accept() error {
	if time.Now().After(q.ValidUntil) {
		q.Status = QuoteExpired
		return ErrInvalidTransition
	}
	if q.Status != QuoteSent && q.Status != QuoteDraft {
		return ErrInvalidTransition
	}
	q.Status = QuoteAccepted
	q.UpdatedAt = time.Now()

	event := QuoteAcceptedEvent{
		TenantID:    q.TenantID,
		QuoteNumber: q.QuoteNumber,
		AcceptedAt:  q.UpdatedAt,
	}
	q.Apply(event)
	return nil
}

// Reject transitions quote to Rejected
func (q *Quote) Reject(reason string) error {
	if q.Status == QuoteAccepted {
		return ErrInvalidTransition
	}
	q.Status = QuoteRejected
	q.Notes = fmt.Sprintf("Rejected: %s", reason)
	q.UpdatedAt = time.Now()

	event := QuoteRejectedEvent{
		TenantID:    q.TenantID,
		QuoteNumber: q.QuoteNumber,
		RejectedAt:  q.UpdatedAt,
		Reason:      reason,
	}
	q.Apply(event)
	return nil
}

// IsExpired checks if quote is past validity date
func (q *Quote) IsExpired() bool {
	return time.Now().After(q.ValidUntil) && q.Status != QuoteAccepted && q.Status != QuoteRejected
}

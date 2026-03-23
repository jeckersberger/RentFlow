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
	QuoteConfirmed QuoteStatus = "confirmed"
	QuoteRejected QuoteStatus = "rejected"
	QuoteExpired  QuoteStatus = "expired"
)

// IsValidStatus checks if status is valid
func (s QuoteStatus) IsValidStatus() bool {
	switch s {
	case QuoteDraft, QuoteSent, QuoteAccepted, QuoteConfirmed, QuoteRejected, QuoteExpired:
		return true
	default:
		return false
	}
}

// Quote represents an estimate/Angebot
type Quote struct {
	events.AggregateRoot
	TenantID      string
	QuoteNumber   string
	ProjectID     *string
	ClientName    string
	ClientEmail   string
	ClientAddress Address
	Items         []InvoiceItem // reuse same item structure
	SubTotal      float64
	TaxRate       TaxRate
	TaxAmount     float64
	Total         float64
	Currency      string
	Status        QuoteStatus
	ValidUntil    time.Time
	Notes         string
	CreatedAt     time.Time
	UpdatedAt     time.Time
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
	if item.TaxRate == 0 && q.TaxRate != 0 {
		item.TaxRate = q.TaxRate
	}
	item.TaxAmount = item.TotalPrice * (float64(item.TaxRate) / 100)
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
	taxRate := TaxRate(rate)
	if !taxRate.IsValidTaxRate() {
		return ErrInvalidTaxRate
	}
	q.TaxRate = taxRate
	q.UpdatedAt = time.Now()
	return nil
}

// CalculateTotals recalculates SubTotal, TaxAmount, and Total using per-item tax
func (q *Quote) CalculateTotals() error {
	if len(q.Items) == 0 {
		return ErrNoItems
	}

	q.SubTotal = 0
	q.TaxAmount = 0
	for idx := range q.Items {
		q.Items[idx].TotalPrice = q.Items[idx].Quantity * q.Items[idx].UnitPrice
		q.Items[idx].TaxAmount = q.Items[idx].TotalPrice * (float64(q.Items[idx].TaxRate) / 100)
		q.SubTotal += q.Items[idx].TotalPrice
		q.TaxAmount += q.Items[idx].TaxAmount
	}

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
	eventData, err := events.NewEventData("QuoteSent", event, nil)
	if err != nil {
		return err
	}
	q.Apply(*eventData)
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
	eventData, err := events.NewEventData("QuoteAccepted", event, nil)
	if err != nil {
		return err
	}
	q.Apply(*eventData)
	return nil
}

// Confirm transitions quote from Accepted to Confirmed (Auftragsbestätigung)
func (q *Quote) Confirm() error {
	if q.Status != QuoteAccepted {
		return ErrInvalidTransition
	}
	q.Status = QuoteConfirmed
	q.UpdatedAt = time.Now()

	event := QuoteConfirmedEvent{
		TenantID:    q.TenantID,
		QuoteNumber: q.QuoteNumber,
		ConfirmedAt: q.UpdatedAt,
	}
	eventData, err := events.NewEventData("QuoteConfirmed", event, nil)
	if err != nil {
		return err
	}
	q.Apply(*eventData)
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
	eventData, err := events.NewEventData("QuoteRejected", event, nil)
	if err != nil {
		return err
	}
	q.Apply(*eventData)
	return nil
}

// IsExpired checks if quote is past validity date
func (q *Quote) IsExpired() bool {
	return time.Now().After(q.ValidUntil) && q.Status != QuoteAccepted && q.Status != QuoteRejected
}

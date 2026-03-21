package domain

import (
	"fmt"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/events"
)

// CreditNoteStatus represents the state of a credit note
type CreditNoteStatus string

const (
	CreditNoteDraft  CreditNoteStatus = "draft"
	CreditNoteIssued CreditNoteStatus = "issued"
)

// IsValidStatus checks if status is valid
func (s CreditNoteStatus) IsValidStatus() bool {
	switch s {
	case CreditNoteDraft, CreditNoteIssued:
		return true
	default:
		return false
	}
}

// CreditNote represents a credit note (Gutschrift) aggregate
type CreditNote struct {
	events.AggregateRoot
	TenantID              string
	CreditNoteNumber      string
	OriginalInvoiceID     string
	OriginalInvoiceNumber string
	ClientName            string
	ClientEmail           string
	Items                 []CreditNoteItem
	SubTotal              float64
	TaxRate               TaxRate
	TaxAmount             float64
	Total                 float64
	Currency              string
	Reason                string
	Status                CreditNoteStatus
	IssuedAt              *time.Time
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

// CreditNoteItem represents a line item on a credit note
type CreditNoteItem struct {
	ID          string
	Description string
	Quantity    float64
	Unit        string
	UnitPrice   float64
	TotalPrice  float64
	TaxRate     TaxRate
}

// NewCreditNote creates a new credit note aggregate
func NewCreditNote(id, tenantID, creditNoteNumber, originalInvoiceID, originalInvoiceNumber, clientName, clientEmail string) *CreditNote {
	now := time.Now()
	return &CreditNote{
		AggregateRoot:         *events.NewAggregateRoot(id, "credit_note"),
		TenantID:              tenantID,
		CreditNoteNumber:      creditNoteNumber,
		OriginalInvoiceID:     originalInvoiceID,
		OriginalInvoiceNumber: originalInvoiceNumber,
		ClientName:            clientName,
		ClientEmail:           clientEmail,
		Status:                CreditNoteDraft,
		Currency:              "EUR",
		Items:                 []CreditNoteItem{},
		CreatedAt:             now,
		UpdatedAt:             now,
	}
}

// AddItem adds a line item to the credit note
func (c *CreditNote) AddItem(item CreditNoteItem) error {
	if item.Description == "" {
		return ErrInvalidInput
	}
	if item.Quantity <= 0 || item.UnitPrice < 0 {
		return ErrInvalidInput
	}

	item.TotalPrice = item.Quantity * item.UnitPrice
	item.TaxRate = c.TaxRate
	item.ID = fmt.Sprintf("item_%d", len(c.Items)+1)

	c.Items = append(c.Items, item)
	c.UpdatedAt = time.Now()
	return nil
}

// SetTaxRate sets the VAT rate
func (c *CreditNote) SetTaxRate(rate float64) error {
	taxRate := TaxRate(rate)
	if !taxRate.IsValidTaxRate() {
		return ErrInvalidTaxRate
	}
	c.TaxRate = taxRate
	c.UpdatedAt = time.Now()
	return nil
}

// CalculateTotals recalculates SubTotal, TaxAmount, and Total
func (c *CreditNote) CalculateTotals() error {
	if len(c.Items) == 0 {
		return ErrNoItems
	}

	c.SubTotal = 0
	for _, item := range c.Items {
		c.SubTotal += item.TotalPrice
	}

	c.TaxAmount = c.SubTotal * (float64(c.TaxRate) / 100)
	c.Total = c.SubTotal + c.TaxAmount

	if c.Total <= 0 {
		return ErrZeroAmount
	}

	c.UpdatedAt = time.Now()
	return nil
}

// Validate checks the credit note is valid for persistence
func (c *CreditNote) Validate() error {
	if c.TenantID == "" {
		return ErrTenantIDRequired
	}
	if c.CreditNoteNumber == "" {
		return ErrInvalidInput
	}
	if c.ClientName == "" {
		return ErrInvalidInput
	}
	if len(c.Items) == 0 {
		return ErrNoItems
	}
	if !c.TaxRate.IsValidTaxRate() {
		return ErrInvalidTaxRate
	}
	if c.Total <= 0 {
		return ErrZeroAmount
	}
	if !c.Status.IsValidStatus() {
		return ErrInvalidStatus
	}
	return nil
}

// Issue transitions credit note to Issued
func (c *CreditNote) Issue() error {
	if c.Status == CreditNoteIssued {
		return ErrInvalidTransition
	}
	now := time.Now()
	c.Status = CreditNoteIssued
	c.IssuedAt = &now
	c.UpdatedAt = now

	// Record event
	event := CreditNoteIssuedEvent{
		TenantID:              c.TenantID,
		CreditNoteNumber:      c.CreditNoteNumber,
		OriginalInvoiceNumber: c.OriginalInvoiceNumber,
		IssuedAt:              now,
		Total:                 c.Total,
	}
	eventData, err := events.NewEventData("CreditNoteIssued", event, nil)
	if err != nil {
		return err
	}
	c.Apply(*eventData)
	return nil
}

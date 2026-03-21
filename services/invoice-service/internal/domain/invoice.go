package domain

import (
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/events"
)

// TaxRate represents a tax rate as a percentage
type TaxRate float64

// IsValidTaxRate checks if the tax rate is valid (0, 7, or 19%)
func (t TaxRate) IsValidTaxRate() bool {
	switch t {
	case 0, 7, 19:
		return true
	default:
		return false
	}
}

// InvoiceStatus represents the state of an invoice
type InvoiceStatus string

const (
	InvoiceDraft        InvoiceStatus = "draft"
	InvoiceSent         InvoiceStatus = "sent"
	InvoiceOverdue      InvoiceStatus = "overdue"
	InvoicePaid         InvoiceStatus = "paid"
	InvoicePartiallyPaid InvoiceStatus = "partially_paid"
	InvoiceCancelled    InvoiceStatus = "cancelled"
	InvoiceCredited     InvoiceStatus = "credited"
)

// IsValidStatus checks if status is valid
func (s InvoiceStatus) IsValidStatus() bool {
	switch s {
	case InvoiceDraft, InvoiceSent, InvoiceOverdue, InvoicePaid, InvoicePartiallyPaid, InvoiceCancelled, InvoiceCredited:
		return true
	default:
		return false
	}
}

// Invoice represents a GoBD-compliant invoice aggregate
type Invoice struct {
	events.AggregateRoot
	TenantID        string
	InvoiceNumber   string
	ProjectID       *string
	ClientName      string
	ClientAddress   Address
	ClientEmail     string
	ClientTaxID     string // USt-IdNr
	Items           []InvoiceItem
	SubTotal        float64
	TaxRate         TaxRate // 0, 7, or 19%
	TaxAmount       float64
	Total           float64
	Currency        string
	Status          InvoiceStatus
	IssueDate       time.Time
	DueDate         time.Time
	PaidDate        *time.Time
	PaymentMethod   string
	PaymentRef      string // Bank reference
	PaidAmount      float64
	RemainingAmount float64
	Notes           string
	InternalNotes   string
	PDFRef          string // file reference to generated PDF
	Hash            string // SHA-256 for GoBD immutability
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// InvoiceItem represents a single line item on an invoice
type InvoiceItem struct {
	ID          string
	Description string
	Quantity    float64
	Unit        string // "Stück", "Tag", "Pauschal", etc.
	UnitPrice   float64
	TotalPrice  float64
	TaxRate     TaxRate
	EquipmentID *string // optional link to equipment
}

// Address represents a customer address
type Address struct {
	Street   string
	City     string
	PostCode string
	Country  string
}

// NewInvoice creates a new invoice aggregate
func NewInvoice(id, tenantID, invoiceNumber, clientName, clientEmail string) *Invoice {
	now := time.Now()
	return &Invoice{
		AggregateRoot: *events.NewAggregateRoot(id, "invoice"),
		TenantID:      tenantID,
		InvoiceNumber: invoiceNumber,
		ClientName:    clientName,
		ClientEmail:   clientEmail,
		Status:        InvoiceDraft,
		Currency:      "EUR",
		IssueDate:     now,
		DueDate:       now.AddDate(0, 0, 14), // Default 14 days payment term
		Items:         []InvoiceItem{},
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

// AddItem adds a line item to the invoice
func (i *Invoice) AddItem(item InvoiceItem) error {
	if item.Description == "" {
		return ErrInvalidInput
	}
	if item.Quantity <= 0 || item.UnitPrice < 0 {
		return ErrInvalidInput
	}

	item.TotalPrice = item.Quantity * item.UnitPrice
	item.TaxRate = i.TaxRate
	item.ID = fmt.Sprintf("item_%d", len(i.Items)+1)

	i.Items = append(i.Items, item)
	i.UpdatedAt = time.Now()
	return nil
}

// RemoveItem removes a line item from the invoice
func (i *Invoice) RemoveItem(itemID string) error {
	for idx, item := range i.Items {
		if item.ID == itemID {
			i.Items = append(i.Items[:idx], i.Items[idx+1:]...)
			i.UpdatedAt = time.Now()
			return nil
		}
	}
	return ErrInvalidInput
}

// SetTaxRate sets the VAT rate (0, 7, or 19%)
func (i *Invoice) SetTaxRate(rate float64) error {
	taxRate := TaxRate(rate)
	if !taxRate.IsValidTaxRate() {
		return ErrInvalidTaxRate
	}
	i.TaxRate = taxRate
	i.UpdatedAt = time.Now()
	return nil
}

// CalculateTotals recalculates SubTotal, TaxAmount, and Total
// Based on item quantities and the invoice's tax rate
func (i *Invoice) CalculateTotals() error {
	if len(i.Items) == 0 {
		return ErrNoItems
	}

	i.SubTotal = 0
	for _, item := range i.Items {
		i.SubTotal += item.TotalPrice
	}

	// Calculate tax: tax = subtotal * (rate / 100)
	i.TaxAmount = i.SubTotal * (float64(i.TaxRate) / 100)
	i.Total = i.SubTotal + i.TaxAmount

	if i.Total <= 0 {
		return ErrZeroAmount
	}

	i.UpdatedAt = time.Now()
	return nil
}

// Validate checks the invoice is valid for persistence
func (i *Invoice) Validate() error {
	if i.TenantID == "" {
		return ErrTenantIDRequired
	}
	if i.InvoiceNumber == "" {
		return ErrInvalidInput
	}
	if i.ClientName == "" {
		return ErrInvalidInput
	}
	if len(i.Items) == 0 {
		return ErrNoItems
	}
	if !i.TaxRate.IsValidTaxRate() {
		return ErrInvalidTaxRate
	}
	if i.Total <= 0 {
		return ErrZeroAmount
	}
	if !i.Status.IsValidStatus() {
		return ErrInvalidStatus
	}
	// Initialize RemainingAmount if not yet set
	if i.RemainingAmount == 0 && i.Status != InvoicePaid && i.Status != InvoiceCredited && i.Status != InvoiceCancelled {
		i.RemainingAmount = i.Total
	}
	return nil
}

// Send transitions invoice from Draft to Sent
func (i *Invoice) Send(email string) error {
	if i.Status != InvoiceDraft {
		return ErrInvalidTransition
	}
	i.Status = InvoiceSent
	i.UpdatedAt = time.Now()

	// Record event
	event := InvoiceSentEvent{
		TenantID:      i.TenantID,
		InvoiceNumber: i.InvoiceNumber,
		SentAt:        i.UpdatedAt,
		SentTo:        email,
	}
	eventData, err := events.NewEventData("InvoiceSent", event, nil)
	if err != nil {
		return err
	}
	i.Apply(*eventData)
	return nil
}

// MarkPaid transitions invoice to Paid
func (i *Invoice) MarkPaid(paymentMethod, paymentRef string) error {
	if i.Status == InvoicePaid || i.Status == InvoiceCancelled {
		return ErrInvalidTransition
	}
	now := time.Now()
	i.Status = InvoicePaid
	i.PaidDate = &now
	i.PaymentMethod = paymentMethod
	i.PaymentRef = paymentRef
	i.PaidAmount = i.Total
	i.RemainingAmount = 0
	i.UpdatedAt = now

	// Record event
	event := InvoicePaidEvent{
		TenantID:      i.TenantID,
		InvoiceNumber: i.InvoiceNumber,
		PaidAt:        now,
		PaymentMethod: paymentMethod,
		PaymentRef:    paymentRef,
	}
	eventData, err := events.NewEventData("InvoicePaid", event, nil)
	if err != nil {
		return err
	}
	i.Apply(*eventData)
	return nil
}

// RecordPayment records a partial payment on the invoice
func (i *Invoice) RecordPayment(amount float64) error {
	if amount <= 0 {
		return ErrInvalidInput
	}
	if i.Status == InvoiceCancelled || i.Status == InvoiceCredited {
		return ErrInvalidTransition
	}
	if i.Status == InvoicePaid {
		return ErrInvalidTransition
	}
	if amount > i.RemainingAmount {
		return ErrInvalidInput
	}

	i.PaidAmount += amount
	i.RemainingAmount = i.Total - i.PaidAmount

	// Update status based on remaining amount
	if i.RemainingAmount <= 0 {
		i.Status = InvoicePaid
		now := time.Now()
		i.PaidDate = &now
	} else {
		i.Status = InvoicePartiallyPaid
	}

	i.UpdatedAt = time.Now()

	// Record event
	event := InvoicePartialPaymentEvent{
		TenantID:      i.TenantID,
		InvoiceNumber: i.InvoiceNumber,
		PaidAt:        i.UpdatedAt,
		PaidAmount:    amount,
		TotalPaidAmount: i.PaidAmount,
		RemainingAmount: i.RemainingAmount,
	}
	eventData, err := events.NewEventData("InvoicePartialPayment", event, nil)
	if err != nil {
		return err
	}
	i.Apply(*eventData)
	return nil
}

// Cancel transitions invoice to Cancelled
func (i *Invoice) Cancel(reason string) error {
	if i.Status == InvoicePaid || i.Status == InvoiceCancelled {
		return ErrInvalidTransition
	}
	i.Status = InvoiceCancelled
	i.InternalNotes = fmt.Sprintf("Cancelled: %s", reason)
	i.UpdatedAt = time.Now()

	// Record event
	event := InvoiceCancelledEvent{
		TenantID:      i.TenantID,
		InvoiceNumber: i.InvoiceNumber,
		CancelledAt:   i.UpdatedAt,
		Reason:        reason,
	}
	eventData, err := events.NewEventData("InvoiceCancelled", event, nil)
	if err != nil {
		return err
	}
	i.Apply(*eventData)
	return nil
}

// Credit creates a credit note (negative invoice)
func (i *Invoice) Credit(creditID string) error {
	if i.Status == InvoiceCancelled || i.Status == InvoiceCredited {
		return ErrInvalidTransition
	}
	i.Status = InvoiceCredited
	i.UpdatedAt = time.Now()

	// Record event
	event := InvoiceCreditedEvent{
		TenantID:      i.TenantID,
		InvoiceNumber: i.InvoiceNumber,
		CreditedAt:    i.UpdatedAt,
		CreditNoteID:  creditID,
	}
	eventData, err := events.NewEventData("InvoiceCredited", event, nil)
	if err != nil {
		return err
	}
	i.Apply(*eventData)
	return nil
}

// ComputeHash returns SHA-256 hash of invoice content for GoBD compliance
func (i *Invoice) ComputeHash() string {
	// Hash includes all financial data to ensure immutability
	data := fmt.Sprintf(
		"%s|%s|%f|%f|%f|%s|%s",
		i.InvoiceNumber,
		i.ClientName,
		i.SubTotal,
		i.TaxAmount,
		i.Total,
		i.IssueDate.Format("2006-01-02"),
		i.DueDate.Format("2006-01-02"),
	)
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash)
}

// VerifyHash verifies the invoice hash for integrity
func (i *Invoice) VerifyHash() bool {
	return i.Hash == i.ComputeHash()
}

// IsFinalized checks if invoice cannot be modified
func (i *Invoice) IsFinalized() bool {
	return i.Status != InvoiceDraft
}

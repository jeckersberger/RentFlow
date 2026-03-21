package domain

import (
	"fmt"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/events"
)

// DunningEntry represents a payment reminder/dunning notice
type DunningEntry struct {
	events.AggregateRoot
	TenantID      string
	InvoiceID     string
	InvoiceNumber string
	Level         int       // 1 = Zahlungserinnerung, 2 = 1. Mahnung, 3 = 2. Mahnung
	SentAt        time.Time
	DueDate       time.Time
	Fee           float64   // Mahngebühr (dunning fee)
	Notes         string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// DunningLevel constants
const (
	DunningLevelPaymentReminder = 1 // Zahlungserinnerung
	DunningLevelFirstNotice     = 2 // 1. Mahnung
	DunningLevelSecondNotice    = 3 // 2. Mahnung
)

// DunningFees in EUR (configurable per tenant)
var DefaultDunningFees = map[int]float64{
	DunningLevelPaymentReminder: 0,    // No fee for reminder
	DunningLevelFirstNotice:     5.00, // €5 for first notice
	DunningLevelSecondNotice:    10.00, // €10 for second notice
}

// NewDunningEntry creates a new dunning entry
func NewDunningEntry(
	id, tenantID, invoiceID, invoiceNumber string,
	level int,
	daysToAdd int,
	fee float64,
) *DunningEntry {
	now := time.Now()
	return &DunningEntry{
		AggregateRoot: *events.NewAggregateRoot(id, "dunning"),
		TenantID:      tenantID,
		InvoiceID:     invoiceID,
		InvoiceNumber: invoiceNumber,
		Level:         level,
		DueDate:       now.AddDate(0, 0, daysToAdd),
		Fee:           fee,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

// Validate checks the dunning entry is valid
func (d *DunningEntry) Validate() error {
	if d.TenantID == "" {
		return ErrTenantIDRequired
	}
	if d.InvoiceID == "" {
		return ErrInvalidInput
	}
	if d.InvoiceNumber == "" {
		return ErrInvalidInput
	}
	if d.Level < 1 || d.Level > 3 {
		return ErrInvalidInput
	}
	if d.Fee < 0 {
		return ErrInvalidInput
	}
	return nil
}

// Send marks the dunning entry as sent
func (d *DunningEntry) Send(email string) error {
	d.SentAt = time.Now()
	d.UpdatedAt = d.SentAt

	event := DunningSentEvent{
		TenantID:      d.TenantID,
		DunningID:     d.ID,
		InvoiceNumber: d.InvoiceNumber,
		SentAt:        d.SentAt,
		SentTo:        email,
		Level:         d.Level,
	}
	eventData, err := events.NewEventData("DunningSent", event, nil)
	if err != nil {
		return err
	}
	d.Apply(*eventData)
	return nil
}

// GetDunningLevelDescription returns a human-readable description
func GetDunningLevelDescription(level int) string {
	switch level {
	case DunningLevelPaymentReminder:
		return "Zahlungserinnerung"
	case DunningLevelFirstNotice:
		return "1. Mahnung"
	case DunningLevelSecondNotice:
		return "2. Mahnung"
	default:
		return fmt.Sprintf("Mahnung Level %d", level)
	}
}

// DunningService helper to create appropriate dunning entries
type DunningCalculator struct {
	DaysBetweenReminders int
	Fees                 map[int]float64
}

// NewDunningCalculator creates a new dunning calculator
func NewDunningCalculator() *DunningCalculator {
	return &DunningCalculator{
		DaysBetweenReminders: 7,
		Fees:                 DefaultDunningFees,
	}
}

// CalculateNextDunningLevel determines the next dunning level
// Returns the level and days from due date
func (dc *DunningCalculator) CalculateNextDunningLevel(daysPastDue int) (int, int) {
	if daysPastDue <= 0 {
		return 0, 0 // Not overdue
	} else if daysPastDue <= 14 {
		return DunningLevelPaymentReminder, 0 // Send reminder immediately
	} else if daysPastDue <= 21 {
		return DunningLevelFirstNotice, 14 // 14 days past due
	} else {
		return DunningLevelSecondNotice, 21 // 21 days past due
	}
}

// GetFeeForLevel returns the dunning fee for a given level
func (dc *DunningCalculator) GetFeeForLevel(level int) float64 {
	if fee, ok := dc.Fees[level]; ok {
		return fee
	}
	return 0
}

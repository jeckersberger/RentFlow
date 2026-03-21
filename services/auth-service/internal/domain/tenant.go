package domain

import (
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/events"
)

// AggregateRoot is the base for event sourcing
type AggregateRoot struct {
	ID      string
	Type    string
	Version int64
	Changes []interface{}
}

// TenantStatus represents the status of a tenant
type TenantStatus string

const (
	TenantStatusActive   TenantStatus = "active"
	TenantStatusInactive TenantStatus = "inactive"
)

// Address represents a physical address
type Address struct {
	Street  string
	City    string
	ZIP     string
	Country string
}

// TenantSettings holds tenant-specific configuration
type TenantSettings struct {
	DefaultLanguage string
	Currency        string
	TaxRate         float64
	InvoicePrefix   string
}

// Tenant represents a tenant (organization) in the system
type Tenant struct {
	AggregateRoot
	Name       string
	Slug       string
	Address    Address
	Logo       string
	Settings   TenantSettings
	Status     TenantStatus
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// NewTenant creates a new tenant aggregate
func NewTenant(id, name, slug string) *Tenant {
	return &Tenant{
		AggregateRoot: AggregateRoot{
			ID:      id,
			Type:    "Tenant",
			Version: 0,
		},
		Name:   name,
		Slug:   slug,
		Status: TenantStatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Settings: TenantSettings{
			DefaultLanguage: "de",
			Currency:        "EUR",
			TaxRate:         19.00,
			InvoicePrefix:   "RF",
		},
	}
}

// UpdateTenant updates tenant details
func (t *Tenant) UpdateTenant(name, slug string, settings TenantSettings) error {
	t.Name = name
	t.Slug = slug
	t.Settings = settings
	t.UpdatedAt = time.Now()

	data, _ := NewEventData("TenantUpdated", map[string]interface{}{
		"id":   t.ID,
		"name": t.Name,
		"slug": t.Slug,
	}, nil)
	t.Apply(*data)

	return nil
}

// Deactivate deactivates the tenant
func (t *Tenant) Deactivate() error {
	t.Status = TenantStatusInactive
	t.UpdatedAt = time.Now()

	data, _ := NewEventData("TenantDeactivated", map[string]interface{}{
		"id": t.ID,
	}, nil)
	t.Apply(*data)

	return nil
}

// Apply adds an event to the uncommitted changes
func (t *Tenant) Apply(event events.EventData) {
	t.Changes = append(t.Changes, event)
}

package domain

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// Customer struct construction and zero-value behavior
// ---------------------------------------------------------------------------

func TestCustomerConstruction(t *testing.T) {
	tenantID := uuid.New()
	customerID := uuid.New()
	now := time.Now()

	c := Customer{
		ID:                    customerID,
		TenantID:              tenantID,
		CompanyName:           "Test GmbH",
		CustomerNumber:        "KD-001",
		Email:                 "test@example.com",
		Phone:                 "+49 123 456",
		BillingAddressStreet:  "Teststr. 1",
		BillingAddressCity:    "Berlin",
		BillingAddressZip:     "10115",
		BillingAddressCountry: "DE",
		TaxID:                 "DE123456789",
		IsActive:              true,
		CreatedAt:             now,
		UpdatedAt:             now,
	}

	if c.ID != customerID {
		t.Errorf("Customer.ID = %v, want %v", c.ID, customerID)
	}
	if c.TenantID != tenantID {
		t.Errorf("Customer.TenantID = %v, want %v", c.TenantID, tenantID)
	}
	if c.CompanyName != "Test GmbH" {
		t.Errorf("Customer.CompanyName = %q, want %q", c.CompanyName, "Test GmbH")
	}
	if !c.IsActive {
		t.Error("Customer.IsActive = false, want true")
	}
}

func TestCustomerZeroValue(t *testing.T) {
	var c Customer

	if c.ID != uuid.Nil {
		t.Errorf("zero Customer.ID = %v, want uuid.Nil", c.ID)
	}
	if c.CompanyName != "" {
		t.Errorf("zero Customer.CompanyName = %q, want empty", c.CompanyName)
	}
	if c.IsActive {
		t.Error("zero Customer.IsActive = true, want false")
	}
}

// ---------------------------------------------------------------------------
// Contact struct construction
// ---------------------------------------------------------------------------

func TestContactConstruction(t *testing.T) {
	tenantID := uuid.New()
	customerID := uuid.New()
	contactID := uuid.New()

	c := Contact{
		ID:         contactID,
		TenantID:   tenantID,
		CustomerID: customerID,
		FirstName:  "Max",
		LastName:   "Mustermann",
		Email:      "max@example.com",
		Phone:      "+49 123 456",
		Position:   "Geschaeftsfuehrer",
		IsPrimary:  true,
	}

	if c.ID != contactID {
		t.Errorf("Contact.ID = %v, want %v", c.ID, contactID)
	}
	if c.CustomerID != customerID {
		t.Errorf("Contact.CustomerID = %v, want %v", c.CustomerID, customerID)
	}
	if c.FirstName != "Max" {
		t.Errorf("Contact.FirstName = %q, want %q", c.FirstName, "Max")
	}
	if c.LastName != "Mustermann" {
		t.Errorf("Contact.LastName = %q, want %q", c.LastName, "Mustermann")
	}
	if !c.IsPrimary {
		t.Error("Contact.IsPrimary = false, want true")
	}
}

// ---------------------------------------------------------------------------
// ContactNote struct construction
// ---------------------------------------------------------------------------

func TestContactNoteConstruction(t *testing.T) {
	contactID := uuid.New()
	userID := uuid.New()

	note := ContactNote{
		ID:        uuid.New(),
		TenantID:  uuid.New(),
		ContactID: contactID,
		Content:   "Bevorzugt E-Mail-Kommunikation",
		CreatedBy: &userID,
	}

	if note.ContactID != contactID {
		t.Errorf("ContactNote.ContactID = %v, want %v", note.ContactID, contactID)
	}
	if note.Content != "Bevorzugt E-Mail-Kommunikation" {
		t.Errorf("ContactNote.Content = %q, unexpected", note.Content)
	}
	if note.CreatedBy == nil || *note.CreatedBy != userID {
		t.Errorf("ContactNote.CreatedBy = %v, want %v", note.CreatedBy, userID)
	}
}

func TestContactNoteCreatedByNilable(t *testing.T) {
	note := ContactNote{
		ID:      uuid.New(),
		Content: "System generated note",
	}

	if note.CreatedBy != nil {
		t.Errorf("ContactNote.CreatedBy = %v, want nil for system notes", note.CreatedBy)
	}
}

// ---------------------------------------------------------------------------
// CustomerFilter defaults
// ---------------------------------------------------------------------------

func TestCustomerFilterDefaults(t *testing.T) {
	f := CustomerFilter{}

	if f.Page != 0 {
		t.Errorf("CustomerFilter zero value Page = %d, want 0", f.Page)
	}
	if f.PerPage != 0 {
		t.Errorf("CustomerFilter zero value PerPage = %d, want 0", f.PerPage)
	}
	if f.Active != nil {
		t.Errorf("CustomerFilter zero value Active = %v, want nil", f.Active)
	}
}

func TestCustomerFilterWithActiveFlag(t *testing.T) {
	active := true
	f := CustomerFilter{
		Page:    1,
		PerPage: 20,
		Active:  &active,
	}

	if f.Page != 1 {
		t.Errorf("CustomerFilter.Page = %d, want 1", f.Page)
	}
	if f.Active == nil || *f.Active != true {
		t.Errorf("CustomerFilter.Active = %v, want *true", f.Active)
	}
}

// ---------------------------------------------------------------------------
// Domain error sentinels
// ---------------------------------------------------------------------------

func TestDomainErrors(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		wantMsg string
	}{
		{"ErrCustomerNotFound", ErrCustomerNotFound, "customer not found"},
		{"ErrContactNotFound", ErrContactNotFound, "contact not found"},
		{"ErrNoteNotFound", ErrNoteNotFound, "contact note not found"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err == nil {
				t.Fatalf("%s is nil", tt.name)
			}
			if tt.err.Error() != tt.wantMsg {
				t.Errorf("%s.Error() = %q, want %q", tt.name, tt.err.Error(), tt.wantMsg)
			}
		})
	}
}

func TestDomainErrorsAreDistinct(t *testing.T) {
	allErrors := []error{
		ErrCustomerNotFound,
		ErrContactNotFound,
		ErrNoteNotFound,
	}

	for i := 0; i < len(allErrors); i++ {
		for j := i + 1; j < len(allErrors); j++ {
			if errors.Is(allErrors[i], allErrors[j]) {
				t.Errorf("errors at index %d and %d should be distinct", i, j)
			}
		}
	}
}

// ---------------------------------------------------------------------------
// Customer address fields: billing vs shipping
// ---------------------------------------------------------------------------

func TestCustomerSeparateBillingAndShipping(t *testing.T) {
	c := Customer{
		BillingAddressStreet:   "Rechnungsstr. 1",
		BillingAddressCity:     "Berlin",
		BillingAddressZip:      "10115",
		BillingAddressCountry:  "DE",
		ShippingAddressStreet:  "Lieferstr. 2",
		ShippingAddressCity:    "Hamburg",
		ShippingAddressZip:     "20095",
		ShippingAddressCountry: "DE",
	}

	if c.BillingAddressStreet == c.ShippingAddressStreet {
		t.Error("billing and shipping streets should be different in this test")
	}
	if c.BillingAddressCity == c.ShippingAddressCity {
		t.Error("billing and shipping cities should be different in this test")
	}
}

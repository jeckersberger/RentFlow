package domain

import (
	"time"

	"github.com/google/uuid"
)

// Customer represents a rental customer (company or individual).
type Customer struct {
	ID                     uuid.UUID `json:"id"`
	TenantID               uuid.UUID `json:"tenant_id"`
	CompanyName            string    `json:"company_name"`
	CustomerNumber         string    `json:"customer_number"`
	Email                  string    `json:"email"`
	Phone                  string    `json:"phone"`
	Mobile                 string    `json:"mobile"`
	Website                string    `json:"website"`
	BillingAddressStreet   string    `json:"billing_address_street"`
	BillingAddressCity     string    `json:"billing_address_city"`
	BillingAddressZip      string    `json:"billing_address_zip"`
	BillingAddressCountry  string    `json:"billing_address_country"`
	ShippingAddressStreet  string    `json:"shipping_address_street"`
	ShippingAddressCity    string    `json:"shipping_address_city"`
	ShippingAddressZip     string    `json:"shipping_address_zip"`
	ShippingAddressCountry string    `json:"shipping_address_country"`
	TaxID                  string    `json:"tax_id"`
	Notes                  string    `json:"notes"`
	IsActive               bool      `json:"is_active"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
}

// Contact represents a contact person associated with a customer.
type Contact struct {
	ID         uuid.UUID `json:"id"`
	TenantID   uuid.UUID `json:"tenant_id"`
	CustomerID uuid.UUID `json:"customer_id"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	Email      string    `json:"email"`
	Phone      string    `json:"phone"`
	Mobile     string    `json:"mobile"`
	Position   string    `json:"position"`
	IsPrimary  bool      `json:"is_primary"`
	Notes      string    `json:"notes"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// ContactNote represents a note attached to a contact.
type ContactNote struct {
	ID        uuid.UUID  `json:"id"`
	TenantID  uuid.UUID  `json:"tenant_id"`
	ContactID uuid.UUID  `json:"contact_id"`
	Content   string     `json:"content"`
	CreatedBy *uuid.UUID `json:"created_by,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// CustomerFilter holds filter/pagination parameters for listing customers.
type CustomerFilter struct {
	Page    int    `json:"page"`
	PerPage int    `json:"per_page"`
	Active  *bool  `json:"active,omitempty"`
	Search  string `json:"search,omitempty"`
}

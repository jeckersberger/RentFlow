package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Tenant represents a company/organization using the EquipFlow platform.
type Tenant struct {
	ID              uuid.UUID       `json:"id"`
	Name            string          `json:"name"`
	Slug            string          `json:"slug"`
	Email           string          `json:"email,omitempty"`
	Phone           string          `json:"phone,omitempty"`
	AddressStreet   string          `json:"address_street,omitempty"`
	AddressCity     string          `json:"address_city,omitempty"`
	AddressZip      string          `json:"address_zip,omitempty"`
	AddressCountry  string          `json:"address_country"`
	TaxNumber       string          `json:"tax_number,omitempty"`
	VatID           string          `json:"vat_id,omitempty"`
	IBAN            string          `json:"iban,omitempty"`
	BIC             string          `json:"bic,omitempty"`
	BankName        string          `json:"bank_name,omitempty"`
	LogoURL         string          `json:"logo_url,omitempty"`
	Currency        string          `json:"currency"`
	Locale          string          `json:"locale"`
	Timezone        string          `json:"timezone"`
	Industry        string          `json:"industry"`
	InvoicePrefix   string          `json:"invoice_prefix"`
	InvoiceCounter  int             `json:"invoice_counter"`
	Settings        json.RawMessage `json:"settings"`
	IndustryProfile json.RawMessage `json:"industry_profile"`
	IsActive        bool            `json:"is_active"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

// TenantConfig stores per-tenant key-value configuration.
type TenantConfig struct {
	ID        uuid.UUID       `json:"id"`
	TenantID  uuid.UUID       `json:"tenant_id"`
	Key       string          `json:"key"`
	Value     json.RawMessage `json:"value"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// SetupState tracks whether the initial system setup has been completed.
type SetupState struct {
	ID          uuid.UUID  `json:"id"`
	TokenHash   string     `json:"setup_token_hash"`
	IsComplete  bool       `json:"is_complete"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	CompletedBy *uuid.UUID `json:"completed_by,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

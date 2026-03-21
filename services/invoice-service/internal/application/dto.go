package application

import (
	"time"

	"github.com/jeckersberger/rentflow/services/invoice-service/internal/domain"
)

// InvoiceDTO for API responses
type InvoiceDTO struct {
	ID            string           `json:"id"`
	TenantID      string           `json:"tenant_id"`
	InvoiceNumber string           `json:"invoice_number"`
	ProjectID     *string          `json:"project_id,omitempty"`
	ClientName    string           `json:"client_name"`
	ClientAddress AddressDTO       `json:"client_address"`
	ClientEmail   string           `json:"client_email"`
	ClientTaxID   string           `json:"client_tax_id"`
	Items         []InvoiceItemDTO `json:"items"`
	SubTotal      float64          `json:"sub_total"`
	TaxRate       float64          `json:"tax_rate"`
	TaxAmount     float64          `json:"tax_amount"`
	Total         float64          `json:"total"`
	Currency      string           `json:"currency"`
	Status        string           `json:"status"`
	IssueDate     time.Time        `json:"issue_date"`
	DueDate       time.Time        `json:"due_date"`
	PaidDate      *time.Time       `json:"paid_date,omitempty"`
	PaymentMethod string           `json:"payment_method"`
	PaymentRef    string           `json:"payment_ref"`
	Notes         string           `json:"notes"`
	InternalNotes string           `json:"internal_notes"`
	PDFRef        string           `json:"pdf_ref"`
	Hash          string           `json:"hash"`
	CreatedAt     time.Time        `json:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at"`
}

// InvoiceItemDTO represents a line item
type InvoiceItemDTO struct {
	ID          string  `json:"id"`
	Description string  `json:"description"`
	Quantity    float64 `json:"quantity"`
	Unit        string  `json:"unit"`
	UnitPrice   float64 `json:"unit_price"`
	TotalPrice  float64 `json:"total_price"`
	TaxRate     float64 `json:"tax_rate"`
	EquipmentID *string `json:"equipment_id,omitempty"`
}

// AddressDTO for client addresses
type AddressDTO struct {
	Street   string `json:"street"`
	City     string `json:"city"`
	PostCode string `json:"post_code"`
	Country  string `json:"country"`
}

// QuoteDTO for API responses
type QuoteDTO struct {
	ID            string           `json:"id"`
	TenantID      string           `json:"tenant_id"`
	QuoteNumber   string           `json:"quote_number"`
	ProjectID     *string          `json:"project_id,omitempty"`
	ClientName    string           `json:"client_name"`
	ClientAddress AddressDTO       `json:"client_address"`
	ClientEmail   string           `json:"client_email"`
	Items         []InvoiceItemDTO `json:"items"`
	SubTotal      float64          `json:"sub_total"`
	TaxRate       float64          `json:"tax_rate"`
	TaxAmount     float64          `json:"tax_amount"`
	Total         float64          `json:"total"`
	Currency      string           `json:"currency"`
	Status        string           `json:"status"`
	ValidUntil    time.Time        `json:"valid_until"`
	Notes         string           `json:"notes"`
	CreatedAt     time.Time        `json:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at"`
}

// DunningDTO for API responses
type DunningDTO struct {
	ID            string    `json:"id"`
	TenantID      string    `json:"tenant_id"`
	InvoiceID     string    `json:"invoice_id"`
	InvoiceNumber string    `json:"invoice_number"`
	Level         int       `json:"level"`
	LevelName     string    `json:"level_name"`
	SentAt        time.Time `json:"sent_at"`
	DueDate       time.Time `json:"due_date"`
	Fee           float64   `json:"fee"`
	Notes         string    `json:"notes"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// PaginatedResult for list responses
type PaginatedResult struct {
	Data   interface{} `json:"data"`
	Total  int64       `json:"total"`
	Limit  int         `json:"limit"`
	Offset int         `json:"offset"`
}

// DATEVExportRow for German tax export
type DATEVExportRow struct {
	Umsatz       float64 `json:"umsatz"`       // Amount
	SollHaben    string  `json:"soll_haben"`   // "S" or "H"
	WKZUmsatz    string  `json:"wkz_umsatz"`   // EUR
	Konto        string  `json:"konto"`        // GL account (SKR03: 8400, 8300, etc.)
	Gegenkonto   string  `json:"gegenkonto"`   // AR account
	Belegdatum   string  `json:"belegdatum"`   // DDMM
	Belegnummer  string  `json:"belegnummer"`  // Invoice number
	Buchungstext string  `json:"buchungstext"` // Booking text
}

// Converter functions

// InvoiceToDTO converts domain model to DTO
func InvoiceToDTO(inv *domain.Invoice) *InvoiceDTO {
	items := make([]InvoiceItemDTO, len(inv.Items))
	for i, item := range inv.Items {
		items[i] = InvoiceItemDTO{
			ID:          item.ID,
			Description: item.Description,
			Quantity:    item.Quantity,
			Unit:        item.Unit,
			UnitPrice:   item.UnitPrice,
			TotalPrice:  item.TotalPrice,
			TaxRate:     float64(item.TaxRate),
			EquipmentID: item.EquipmentID,
		}
	}

	return &InvoiceDTO{
		ID:            inv.ID,
		TenantID:      inv.TenantID,
		InvoiceNumber: inv.InvoiceNumber,
		ProjectID:     inv.ProjectID,
		ClientName:    inv.ClientName,
		ClientAddress: AddressDTO{
			Street:   inv.ClientAddress.Street,
			City:     inv.ClientAddress.City,
			PostCode: inv.ClientAddress.PostCode,
			Country:  inv.ClientAddress.Country,
		},
		ClientEmail:   inv.ClientEmail,
		ClientTaxID:   inv.ClientTaxID,
		Items:         items,
		SubTotal:      inv.SubTotal,
		TaxRate:       float64(inv.TaxRate),
		TaxAmount:     inv.TaxAmount,
		Total:         inv.Total,
		Currency:      inv.Currency,
		Status:        string(inv.Status),
		IssueDate:     inv.IssueDate,
		DueDate:       inv.DueDate,
		PaidDate:      inv.PaidDate,
		PaymentMethod: inv.PaymentMethod,
		PaymentRef:    inv.PaymentRef,
		Notes:         inv.Notes,
		InternalNotes: inv.InternalNotes,
		PDFRef:        inv.PDFRef,
		Hash:          inv.Hash,
		CreatedAt:     inv.CreatedAt,
		UpdatedAt:     inv.UpdatedAt,
	}
}

// QuoteToDTO converts domain model to DTO
func QuoteToDTO(quote *domain.Quote) *QuoteDTO {
	items := make([]InvoiceItemDTO, len(quote.Items))
	for i, item := range quote.Items {
		items[i] = InvoiceItemDTO{
			ID:          item.ID,
			Description: item.Description,
			Quantity:    item.Quantity,
			Unit:        item.Unit,
			UnitPrice:   item.UnitPrice,
			TotalPrice:  item.TotalPrice,
			TaxRate:     float64(item.TaxRate),
			EquipmentID: item.EquipmentID,
		}
	}

	return &QuoteDTO{
		ID:          quote.ID,
		TenantID:    quote.TenantID,
		QuoteNumber: quote.QuoteNumber,
		ProjectID:   quote.ProjectID,
		ClientName:  quote.ClientName,
		ClientAddress: AddressDTO{
			Street:   quote.ClientAddress.Street,
			City:     quote.ClientAddress.City,
			PostCode: quote.ClientAddress.PostCode,
			Country:  quote.ClientAddress.Country,
		},
		ClientEmail: quote.ClientEmail,
		Items:       items,
		SubTotal:    quote.SubTotal,
		TaxRate:     float64(quote.TaxRate),
		TaxAmount:   quote.TaxAmount,
		Total:       quote.Total,
		Currency:    quote.Currency,
		Status:      string(quote.Status),
		ValidUntil:  quote.ValidUntil,
		Notes:       quote.Notes,
		CreatedAt:   quote.CreatedAt,
		UpdatedAt:   quote.UpdatedAt,
	}
}

// DunningToDTO converts domain model to DTO
func DunningToDTO(dunning *domain.DunningEntry) *DunningDTO {
	return &DunningDTO{
		ID:            dunning.ID,
		TenantID:      dunning.TenantID,
		InvoiceID:     dunning.InvoiceID,
		InvoiceNumber: dunning.InvoiceNumber,
		Level:         dunning.Level,
		LevelName:     domain.GetDunningLevelDescription(dunning.Level),
		SentAt:        dunning.SentAt,
		DueDate:       dunning.DueDate,
		Fee:           dunning.Fee,
		Notes:         dunning.Notes,
		CreatedAt:     dunning.CreatedAt,
		UpdatedAt:     dunning.UpdatedAt,
	}
}

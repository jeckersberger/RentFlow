package application

import "time"

type CreateExpenseCommand struct {
	TenantID        string
	Type            string
	Vendor          string
	VendorAddress   string
	VendorVATID     string
	VendorIBAN      string
	Amount          float64
	Currency        string
	TaxRate         float64
	CategoryCode    string
	BookingAccount  string
	Date            time.Time
	ServicePeriodFrom *time.Time
	ServicePeriodTo   *time.Time
	DueDate         *time.Time
	DiscountPercent float64
	DiscountDays    int
	PaymentMethod   string
	InvoiceNumber   string
	ReceiptRef      string
	ProjectID       string
	Notes           string
	Source          string // "manual", "email_auto", "photo"
	EmailRef        string
	// Bewirtungsbeleg
	EntertainmentLocation string
	EntertainmentReason   string
	EntertainmentGuests   string
	EntertainmentTip      float64
	// KI-Felder (werden vom ai-service befuellt)
	OCRData       map[string]string
	OCRConfidence float64
	ReceiptChecksum string
	ReceiptNASPath  string
}

type UpdateExpenseCommand struct {
	ID              string
	TenantID        string
	Type            string
	Vendor          string
	VendorAddress   string
	VendorVATID     string
	VendorIBAN      string
	Amount          float64
	CategoryCode    string
	BookingAccount  string
	DueDate         *time.Time
	DiscountPercent float64
	DiscountDays    int
	PaymentMethod   string
	PaymentStatus   string
	InvoiceNumber   string
	Notes           string
	// Bewirtungsbeleg
	EntertainmentLocation string
	EntertainmentReason   string
	EntertainmentGuests   string
	EntertainmentTip      float64
}

type CreateExpenseCategoryCommand struct {
	TenantID  string
	Name      string
	SKR03Code string
	SKR04Code string
	IsDefault bool
}

type CreateBudgetCommand struct {
	TenantID   string
	CategoryID string
	ProjectID  string
	Period     string
	Amount     float64
}

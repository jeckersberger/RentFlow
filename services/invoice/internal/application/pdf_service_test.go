package application

import (
	"bytes"
	"testing"

	"github.com/jeckersberger/EquipFlow/services/invoice/internal/domain"
)

// ---------------------------------------------------------------------------
// centsToEur
// ---------------------------------------------------------------------------

func TestCentsToEur(t *testing.T) {
	tests := []struct {
		name  string
		cents int64
		want  string
	}{
		{name: "zero", cents: 0, want: "0,00 EUR"},
		{name: "one cent", cents: 1, want: "0,01 EUR"},
		{name: "one euro", cents: 100, want: "1,00 EUR"},
		{name: "typical price", cents: 5990, want: "59,90 EUR"},
		{name: "large amount", cents: 150000, want: "1500,00 EUR"},
		{name: "negative amount", cents: -250, want: "-2,50 EUR"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := centsToEur(tt.cents)
			if got != tt.want {
				t.Errorf("centsToEur(%d) = %q, want %q", tt.cents, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// formatDate
// ---------------------------------------------------------------------------

func TestFormatDate(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "ISO date", input: "2026-03-15", want: "15.03.2026"},
		{name: "RFC3339", input: "2026-12-01T10:30:00Z", want: "01.12.2026"},
		{name: "unparseable returns original", input: "not-a-date", want: "not-a-date"},
		{name: "new year", input: "2027-01-01", want: "01.01.2027"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatDate(tt.input)
			if got != tt.want {
				t.Errorf("formatDate(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// encode (umlaut replacement)
// ---------------------------------------------------------------------------

func TestEncode(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "ae", input: "Ger\u00e4t", want: "Geraet"},
		{name: "oe", input: "Gr\u00f6\u00dfe", want: "Groesse"},
		{name: "ue", input: "\u00dcbung", want: "Uebung"},
		{name: "Ae", input: "\u00c4nderung", want: "Aenderung"},
		{name: "Oe", input: "\u00d6ffnung", want: "Oeffnung"},
		{name: "Ue", input: "\u00dcbersicht", want: "Uebersicht"},
		{name: "ss", input: "Stra\u00dfe", want: "Strasse"},
		{name: "euro sign", input: "100\u20ac", want: "100EUR"},
		{name: "no special chars", input: "Hello World", want: "Hello World"},
		{name: "multiple umlauts", input: "f\u00fcr \u00c4nderungen \u00f6ffnen", want: "fuer Aenderungen oeffnen"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := encode(tt.input)
			if got != tt.want {
				t.Errorf("encode(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// GenerateInvoicePDF
// ---------------------------------------------------------------------------

func TestGenerateInvoicePDF_NonEmptyOutput(t *testing.T) {
	data := InvoicePDFData{
		Invoice: &domain.Invoice{
			InvoiceNumber:   "RE-2026-0001",
			InvoiceType:     "invoice",
			Status:          "draft",
			CustomerName:    "Test GmbH",
			CustomerAddress: "Teststrasse 1, 12345 Berlin",
			InvoiceDate:     "2026-03-15",
			DueDate:         "2026-04-15",
			VatRate:         1900,
			TotalNet:        10000,
			TotalVat:        1900,
			TotalGross:      11900,
		},
		Items: []*domain.InvoiceItem{
			{
				Position:    1,
				Description: "Mikrofon Shure SM58",
				Quantity:    2,
				Unit:        "Stk",
				UnitPrice:   5000,
			},
		},
		Company: CompanyInfo{
			Name:      "JE Sound & Light",
			Street:    "Musterweg 42",
			City:      "12345 Musterstadt",
			Phone:     "+49 123 456789",
			Email:     "info@example.com",
			TaxNumber: "12/345/67890",
			IBAN:      "DE89370400440532013000",
			BIC:       "COBADEFFXXX",
			BankName:  "Commerzbank",
		},
	}

	var buf bytes.Buffer
	err := GenerateInvoicePDF(&buf, data)
	if err != nil {
		t.Fatalf("GenerateInvoicePDF() returned error: %v", err)
	}

	if buf.Len() == 0 {
		t.Fatal("GenerateInvoicePDF() produced empty output")
	}

	// A valid PDF starts with %PDF
	header := buf.Bytes()[:4]
	if string(header) != "%PDF" {
		t.Errorf("expected PDF header %%PDF, got %q", string(header))
	}
}

func TestGenerateInvoicePDF_Kleinunternehmer_NoVATLine(t *testing.T) {
	// When Kleinunternehmer is true the PDF should contain the
	// paragraph 19 UStG notice rather than an explicit VAT line.
	// We cannot easily inspect gofpdf text content, but we verify
	// generation succeeds without error, which exercises that branch.
	data := InvoicePDFData{
		Invoice: &domain.Invoice{
			InvoiceNumber:    "RE-2026-0002",
			InvoiceType:      "invoice",
			Status:           "finalized",
			CustomerName:     "Kunde ABC",
			CustomerAddress:  "Beispielstr. 5\n54321 Stadt",
			InvoiceDate:      "2026-03-20",
			VatRate:          0,
			Kleinunternehmer: true,
			TotalNet:         50000,
			TotalVat:         0,
			TotalGross:       50000,
		},
		Items: []*domain.InvoiceItem{
			{
				Position:    1,
				Description: "Lichtpult",
				Quantity:    1,
				Unit:        "Stk",
				UnitPrice:   50000,
			},
		},
		Company: CompanyInfo{
			Name:  "JE Sound & Light",
			Email: "test@example.com",
		},
	}

	var buf bytes.Buffer
	err := GenerateInvoicePDF(&buf, data)
	if err != nil {
		t.Fatalf("GenerateInvoicePDF() (Kleinunternehmer) returned error: %v", err)
	}

	if buf.Len() == 0 {
		t.Fatal("GenerateInvoicePDF() produced empty output for Kleinunternehmer invoice")
	}
}

// ---------------------------------------------------------------------------
// statusLabel
// ---------------------------------------------------------------------------

func TestStatusLabel(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"draft", "Entwurf"},
		{"finalized", "Finalisiert"},
		{"sent", "Versendet"},
		{"paid", "Bezahlt"},
		{"partial_paid", "Teilweise bezahlt"},
		{"overdue", "Ueberfaellig"},
		{"cancelled", "Storniert"},
		{"unknown_status", "unknown_status"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := statusLabel(tt.input)
			if got != tt.want {
				t.Errorf("statusLabel(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

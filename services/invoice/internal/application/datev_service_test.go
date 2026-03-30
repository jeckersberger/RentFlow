package application

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/services/invoice/internal/domain"
)

// ---------------------------------------------------------------------------
// writeHeader — DATEV header format
// ---------------------------------------------------------------------------

func TestWriteHeader(t *testing.T) {
	logger := zerolog.Nop()
	svc := NewDatevService(nil, logger)

	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 3, 31, 0, 0, 0, 0, time.UTC)

	var buf bytes.Buffer
	svc.writeHeader(&buf, from, to)

	got := buf.String()

	// Must start with "EXTF" and contain the date range + year
	if !strings.HasPrefix(got, "\"EXTF\"") {
		t.Errorf("header should start with \"EXTF\", got: %q", got)
	}
	if !strings.Contains(got, "\"20260101\"") {
		t.Errorf("header should contain from-date 20260101, got: %q", got)
	}
	if !strings.Contains(got, "\"20260331\"") {
		t.Errorf("header should contain to-date 20260331, got: %q", got)
	}
	if !strings.Contains(got, "2026") {
		t.Errorf("header should contain year 2026, got: %q", got)
	}
	if !strings.HasSuffix(got, "\r\n") {
		t.Errorf("header should end with CRLF, got: %q", got)
	}
}

// ---------------------------------------------------------------------------
// revenueAccount — SKR03 account mapping
// ---------------------------------------------------------------------------

func TestRevenueAccount_SKR03(t *testing.T) {
	logger := zerolog.Nop()
	svc := NewDatevService(nil, logger)

	tests := []struct {
		name             string
		vatRate          int64
		kleinunternehmer bool
		want             int
	}{
		{name: "19% VAT", vatRate: 1900, kleinunternehmer: false, want: 8400},
		{name: "7% VAT", vatRate: 700, kleinunternehmer: false, want: 8300},
		{name: "Kleinunternehmer", vatRate: 0, kleinunternehmer: true, want: 8195},
		{name: "0% default", vatRate: 0, kleinunternehmer: false, want: 8195},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inv := &domain.Invoice{
				VatRate:          tt.vatRate,
				Kleinunternehmer: tt.kleinunternehmer,
			}
			got := svc.revenueAccount(inv, "skr03")
			if got != tt.want {
				t.Errorf("revenueAccount(skr03, vat=%d, ku=%v) = %d, want %d",
					tt.vatRate, tt.kleinunternehmer, got, tt.want)
			}
		})
	}
}

func TestRevenueAccount_SKR04(t *testing.T) {
	logger := zerolog.Nop()
	svc := NewDatevService(nil, logger)

	tests := []struct {
		name             string
		vatRate          int64
		kleinunternehmer bool
		want             int
	}{
		{name: "19% VAT", vatRate: 1900, kleinunternehmer: false, want: 4400},
		{name: "7% VAT", vatRate: 700, kleinunternehmer: false, want: 4300},
		{name: "Kleinunternehmer", vatRate: 0, kleinunternehmer: true, want: 4185},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inv := &domain.Invoice{
				VatRate:          tt.vatRate,
				Kleinunternehmer: tt.kleinunternehmer,
			}
			got := svc.revenueAccount(inv, "skr04")
			if got != tt.want {
				t.Errorf("revenueAccount(skr04, vat=%d, ku=%v) = %d, want %d",
					tt.vatRate, tt.kleinunternehmer, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// formatDatevAmount — cents to DATEV amount string
// ---------------------------------------------------------------------------

func TestFormatDatevAmount(t *testing.T) {
	tests := []struct {
		name  string
		cents int64
		want  string
	}{
		{name: "zero", cents: 0, want: "0,00"},
		{name: "one cent", cents: 1, want: "0,01"},
		{name: "one euro", cents: 100, want: "1,00"},
		{name: "typical amount", cents: 11900, want: "119,00"},
		{name: "large amount", cents: 150000, want: "1500,00"},
		{name: "negative", cents: -5000, want: "-50,00"},
		{name: "cents remainder", cents: 12345, want: "123,45"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatDatevAmount(tt.cents)
			if got != tt.want {
				t.Errorf("formatDatevAmount(%d) = %q, want %q", tt.cents, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// formatDatevDate — DDMM format
// ---------------------------------------------------------------------------

func TestFormatDatevDate(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "March 15", input: "2026-03-15", want: "1503"},
		{name: "January 1", input: "2026-01-01", want: "0101"},
		{name: "December 31", input: "2026-12-31", want: "3112"},
		{name: "invalid date", input: "not-a-date", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatDatevDate(tt.input)
			if got != tt.want {
				t.Errorf("formatDatevDate(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// taxKey
// ---------------------------------------------------------------------------

func TestTaxKey(t *testing.T) {
	logger := zerolog.Nop()
	svc := NewDatevService(nil, logger)

	tests := []struct {
		name             string
		vatRate          int64
		kleinunternehmer bool
		want             string
	}{
		{name: "19%", vatRate: 1900, kleinunternehmer: false, want: "3"},
		{name: "7%", vatRate: 700, kleinunternehmer: false, want: "2"},
		{name: "Kleinunternehmer", vatRate: 0, kleinunternehmer: true, want: ""},
		{name: "0% no KU", vatRate: 0, kleinunternehmer: false, want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inv := &domain.Invoice{
				VatRate:          tt.vatRate,
				Kleinunternehmer: tt.kleinunternehmer,
			}
			got := svc.taxKey(inv)
			if got != tt.want {
				t.Errorf("taxKey(vat=%d, ku=%v) = %q, want %q",
					tt.vatRate, tt.kleinunternehmer, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// customerAccount — deterministic hash
// ---------------------------------------------------------------------------

func TestCustomerAccount_IsStable(t *testing.T) {
	logger := zerolog.Nop()
	svc := NewDatevService(nil, logger)

	inv := &domain.Invoice{CustomerName: "Test GmbH"}

	first := svc.customerAccount(inv)
	second := svc.customerAccount(inv)

	if first != second {
		t.Errorf("customerAccount should be deterministic: got %d then %d", first, second)
	}

	if first < 10000 || first >= 100000 {
		t.Errorf("customerAccount should be in [10000, 100000), got %d", first)
	}
}

// ---------------------------------------------------------------------------
// writeInvoiceRow — full row format
// ---------------------------------------------------------------------------

func TestWriteInvoiceRow(t *testing.T) {
	logger := zerolog.Nop()
	svc := NewDatevService(nil, logger)

	inv := &domain.Invoice{
		InvoiceNumber:    "RE-2026-0001",
		CustomerName:     "Kunde",
		InvoiceDate:      "2026-03-15",
		VatRate:          1900,
		Kleinunternehmer: false,
		TotalGross:       11900,
	}

	var buf bytes.Buffer
	svc.writeInvoiceRow(&buf, inv, "skr03")

	got := buf.String()

	// Should contain amount, S for Soll, revenue account 8400
	if !strings.Contains(got, "119,00") {
		t.Errorf("row should contain amount 119,00, got: %q", got)
	}
	if !strings.Contains(got, ";S;") {
		t.Errorf("row should contain ;S; for Soll, got: %q", got)
	}
	if !strings.Contains(got, ";8400;") {
		t.Errorf("row should contain revenue account 8400, got: %q", got)
	}
	if !strings.Contains(got, "1503") {
		t.Errorf("row should contain DDMM date 1503, got: %q", got)
	}
	if !strings.Contains(got, "RE-2026-0001") {
		t.Errorf("row should contain invoice number, got: %q", got)
	}
	if !strings.HasSuffix(got, "\r\n") {
		t.Errorf("row should end with CRLF, got: %q", got)
	}
}

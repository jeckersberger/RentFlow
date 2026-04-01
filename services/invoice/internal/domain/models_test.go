package domain

import (
	"errors"
	"testing"
)

// ---------------------------------------------------------------------------
// Invoice type constants
// ---------------------------------------------------------------------------

func TestInvoiceTypeConstants(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{"InvoiceTypeInvoice", InvoiceTypeInvoice, "invoice"},
		{"InvoiceTypePartial", InvoiceTypePartial, "partial"},
		{"InvoiceTypeAdvance", InvoiceTypeAdvance, "advance"},
		{"InvoiceTypeCreditNote", InvoiceTypeCreditNote, "credit_note"},
		{"InvoiceTypeReversal", InvoiceTypeReversal, "reversal"},
		{"InvoiceTypeProforma", InvoiceTypeProforma, "proforma"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("%s = %q, want %q", tt.name, tt.got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Invoice status constants
// ---------------------------------------------------------------------------

func TestInvoiceStatusConstants(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{"StatusDraft", StatusDraft, "draft"},
		{"StatusFinalized", StatusFinalized, "finalized"},
		{"StatusSent", StatusSent, "sent"},
		{"StatusPartialPaid", StatusPartialPaid, "partial_paid"},
		{"StatusPaid", StatusPaid, "paid"},
		{"StatusOverdue", StatusOverdue, "overdue"},
		{"StatusCancelled", StatusCancelled, "cancelled"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("%s = %q, want %q", tt.name, tt.got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Quote status constants
// ---------------------------------------------------------------------------

func TestQuoteStatusConstants(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{"QuoteStatusDraft", QuoteStatusDraft, "draft"},
		{"QuoteStatusSent", QuoteStatusSent, "sent"},
		{"QuoteStatusAccepted", QuoteStatusAccepted, "accepted"},
		{"QuoteStatusDeclined", QuoteStatusDeclined, "declined"},
		{"QuoteStatusExpired", QuoteStatusExpired, "expired"},
		{"QuoteStatusCancelled", QuoteStatusCancelled, "cancelled"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("%s = %q, want %q", tt.name, tt.got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Dunning level and status constants
// ---------------------------------------------------------------------------

func TestDunningConstants(t *testing.T) {
	// Levels
	levelTests := []struct {
		name string
		got  string
		want string
	}{
		{"DunningLevelReminder", DunningLevelReminder, "reminder"},
		{"DunningLevelDunning1", DunningLevelDunning1, "dunning_1"},
		{"DunningLevelDunning2", DunningLevelDunning2, "dunning_2"},
		{"DunningLevelDunning3", DunningLevelDunning3, "dunning_3"},
	}

	for _, tt := range levelTests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("%s = %q, want %q", tt.name, tt.got, tt.want)
			}
		})
	}

	// Statuses
	statusTests := []struct {
		name string
		got  string
		want string
	}{
		{"DunningStatusPending", DunningStatusPending, "pending"},
		{"DunningStatusSent", DunningStatusSent, "sent"},
		{"DunningStatusCancelled", DunningStatusCancelled, "cancelled"},
	}

	for _, tt := range statusTests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("%s = %q, want %q", tt.name, tt.got, tt.want)
			}
		})
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
		{"ErrInvoiceNotFound", ErrInvoiceNotFound, "invoice not found"},
		{"ErrInvoiceNotDraft", ErrInvoiceNotDraft, "invoice is not in draft status"},
		{"ErrInvoiceAlreadyPaid", ErrInvoiceAlreadyPaid, "invoice is already fully paid"},
		{"ErrInvalidInvoiceType", ErrInvalidInvoiceType, "invalid invoice type"},
		{"ErrPaymentExceedsTotal", ErrPaymentExceedsTotal, "payment exceeds remaining balance"},
		{"ErrItemNotFound", ErrItemNotFound, "invoice item not found"},
		{"ErrQuoteNotFound", ErrQuoteNotFound, "quote not found"},
		{"ErrQuoteNotDraft", ErrQuoteNotDraft, "quote is not in draft status"},
		{"ErrQuoteAlreadyConverted", ErrQuoteAlreadyConverted, "quote has already been converted to an invoice"},
		{"ErrDunningConfigNotFound", ErrDunningConfigNotFound, "dunning config not found"},
		{"ErrInvoiceNotOverdue", ErrInvoiceNotOverdue, "invoice is not overdue"},
		{"ErrBankTxNotFound", ErrBankTxNotFound, "bank transaction not found"},
		{"ErrBankTxAlreadyMatched", ErrBankTxAlreadyMatched, "bank transaction is already matched"},
		{"ErrInvalidPercentage", ErrInvalidPercentage, "Prozentsatz muss zwischen 1 und 100 liegen"},
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

// ---------------------------------------------------------------------------
// Domain errors are distinct (no accidental aliasing)
// ---------------------------------------------------------------------------

func TestDomainErrorsAreDistinct(t *testing.T) {
	allErrors := []error{
		ErrInvoiceNotFound,
		ErrInvoiceNotDraft,
		ErrInvoiceAlreadyPaid,
		ErrInvalidInvoiceType,
		ErrPaymentExceedsTotal,
		ErrItemNotFound,
		ErrQuoteNotFound,
		ErrQuoteNotDraft,
		ErrQuoteAlreadyConverted,
		ErrDunningConfigNotFound,
		ErrInvoiceNotOverdue,
		ErrBankTxNotFound,
		ErrBankTxAlreadyMatched,
		ErrInvalidPercentage,
	}

	for i := 0; i < len(allErrors); i++ {
		for j := i + 1; j < len(allErrors); j++ {
			if errors.Is(allErrors[i], allErrors[j]) {
				t.Errorf("errors at index %d and %d should be distinct but errors.Is returned true", i, j)
			}
		}
	}
}

// ---------------------------------------------------------------------------
// InvoiceItem line total calculation (Quantity * UnitPrice)
// ---------------------------------------------------------------------------

func TestInvoiceItemLineTotal(t *testing.T) {
	tests := []struct {
		name      string
		quantity  int64
		unitPrice int64
		wantTotal int64
	}{
		{name: "single item", quantity: 1, unitPrice: 5000, wantTotal: 5000},
		{name: "multiple items", quantity: 3, unitPrice: 2500, wantTotal: 7500},
		{name: "zero quantity", quantity: 0, unitPrice: 5000, wantTotal: 0},
		{name: "zero price", quantity: 5, unitPrice: 0, wantTotal: 0},
		{name: "large quantity", quantity: 100, unitPrice: 15000, wantTotal: 1500000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := &InvoiceItem{
				Quantity:  tt.quantity,
				UnitPrice: tt.unitPrice,
			}
			got := item.Quantity * item.UnitPrice
			if got != tt.wantTotal {
				t.Errorf("InvoiceItem{Qty: %d, Price: %d} line total = %d, want %d",
					tt.quantity, tt.unitPrice, got, tt.wantTotal)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// VAT calculation logic (mirrors Finalize logic)
// ---------------------------------------------------------------------------

func TestVATCalculation(t *testing.T) {
	tests := []struct {
		name             string
		totalNet         int64
		vatRate          int64
		kleinunternehmer bool
		wantVat          int64
		wantGross        int64
	}{
		{
			name:     "standard 19% VAT on 10000 cents",
			totalNet: 10000, vatRate: 1900, kleinunternehmer: false,
			wantVat: 1900, wantGross: 11900,
		},
		{
			name:     "7% VAT on 10000 cents",
			totalNet: 10000, vatRate: 700, kleinunternehmer: false,
			wantVat: 700, wantGross: 10700,
		},
		{
			name:     "Kleinunternehmer no VAT",
			totalNet: 50000, vatRate: 0, kleinunternehmer: true,
			wantVat: 0, wantGross: 50000,
		},
		{
			name:     "zero net amount",
			totalNet: 0, vatRate: 1900, kleinunternehmer: false,
			wantVat: 0, wantGross: 0,
		},
		{
			name:     "large amount 19%",
			totalNet: 1000000, vatRate: 1900, kleinunternehmer: false,
			wantVat: 190000, wantGross: 1190000,
		},
		{
			name:     "rounding: 19% on 101 cents",
			totalNet: 101, vatRate: 1900, kleinunternehmer: false,
			// 101 * 1900 / 10000 = 19.19 -> int truncates to 19
			wantVat: 19, wantGross: 120,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var totalVat int64
			if !tt.kleinunternehmer && tt.vatRate > 0 {
				totalVat = tt.totalNet * tt.vatRate / 10000
			}
			totalGross := tt.totalNet + totalVat

			if totalVat != tt.wantVat {
				t.Errorf("VAT = %d, want %d", totalVat, tt.wantVat)
			}
			if totalGross != tt.wantGross {
				t.Errorf("Gross = %d, want %d", totalGross, tt.wantGross)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// InvoiceFilter defaults
// ---------------------------------------------------------------------------

func TestInvoiceFilterDefaults(t *testing.T) {
	f := InvoiceFilter{}
	if f.Page != 0 {
		t.Errorf("InvoiceFilter zero value Page = %d, want 0", f.Page)
	}
	if f.PerPage != 0 {
		t.Errorf("InvoiceFilter zero value PerPage = %d, want 0", f.PerPage)
	}
	if f.Status != "" {
		t.Errorf("InvoiceFilter zero value Status = %q, want empty", f.Status)
	}
}

package domain

import (
	"math"
	"testing"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/events"
)

func floatEq(a, b float64) bool {
	return math.Abs(a-b) < 0.001
}

func newTestExpense() *Expense {
	return NewExpense(
		"exp_123",
		"tenant_1",
		"Thomann GmbH",
		"EUR",
		"4900",
		"bank_transfer",
		119.00,
		time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
	)
}

// --- NewExpense defaults ---

func TestNewExpense_Defaults(t *testing.T) {
	exp := newTestExpense()

	if exp.ID != "exp_123" {
		t.Errorf("expected ID exp_123, got %s", exp.ID)
	}
	if exp.TenantID != "tenant_1" {
		t.Errorf("expected TenantID tenant_1, got %s", exp.TenantID)
	}
	if exp.Vendor != "Thomann GmbH" {
		t.Errorf("expected Vendor Thomann GmbH, got %s", exp.Vendor)
	}
	if exp.Amount != 119.00 {
		t.Errorf("expected Amount 119.00, got %f", exp.Amount)
	}
	if exp.Currency != "EUR" {
		t.Errorf("expected Currency EUR, got %s", exp.Currency)
	}
	if exp.Status != StatusDraft {
		t.Errorf("expected Status draft, got %s", exp.Status)
	}
	if exp.PaymentStatus != PaymentStatusOpen {
		t.Errorf("expected PaymentStatus open, got %s", exp.PaymentStatus)
	}
	if exp.Type != ExpenseTypeInvoice {
		t.Errorf("expected Type invoice, got %s", exp.Type)
	}
	if exp.TaxRate != 0.19 {
		t.Errorf("expected default TaxRate 0.19, got %f", exp.TaxRate)
	}
	if exp.Source != "manual" {
		t.Errorf("expected Source manual, got %s", exp.Source)
	}
	if exp.OCRData == nil {
		t.Error("expected OCRData to be initialized, got nil")
	}
	if exp.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
	if exp.UpdatedAt.IsZero() {
		t.Error("expected UpdatedAt to be set")
	}
}

func TestNewExpense_AggregateRoot(t *testing.T) {
	exp := newTestExpense()
	if exp.AggregateRoot.ID != "exp_123" {
		t.Errorf("expected AggregateRoot.ID exp_123, got %s", exp.AggregateRoot.ID)
	}
	if exp.AggregateRoot.Type != "expense" {
		t.Errorf("expected AggregateRoot.Type expense, got %s", exp.AggregateRoot.Type)
	}
}

// --- CalculateTax ---

func TestCalculateTax_19Percent(t *testing.T) {
	exp := newTestExpense()
	exp.TaxRate = 0.19
	exp.Amount = 119.00
	exp.CalculateTax()

	expectedTax := 119.00 * 0.19
	expectedNet := 119.00 - expectedTax
	if exp.TaxAmount != expectedTax {
		t.Errorf("expected TaxAmount %f, got %f", expectedTax, exp.TaxAmount)
	}
	if exp.NetAmount != expectedNet {
		t.Errorf("expected NetAmount %f, got %f", expectedNet, exp.NetAmount)
	}
}

func TestCalculateTax_7Percent(t *testing.T) {
	exp := newTestExpense()
	exp.TaxRate = 0.07
	exp.Amount = 107.00
	exp.CalculateTax()

	expectedTax := 107.00 * 0.07
	expectedNet := 107.00 - expectedTax
	if !floatEq(exp.TaxAmount, expectedTax) {
		t.Errorf("expected TaxAmount %f, got %f", expectedTax, exp.TaxAmount)
	}
	if !floatEq(exp.NetAmount, expectedNet) {
		t.Errorf("expected NetAmount %f, got %f", expectedNet, exp.NetAmount)
	}
}

func TestCalculateTax_0Percent(t *testing.T) {
	exp := newTestExpense()
	exp.TaxRate = 0.0
	exp.Amount = 100.00
	exp.CalculateTax()

	if exp.TaxAmount != 0.0 {
		t.Errorf("expected TaxAmount 0, got %f", exp.TaxAmount)
	}
	if exp.NetAmount != 100.00 {
		t.Errorf("expected NetAmount 100.00, got %f", exp.NetAmount)
	}
}

// --- CalculateEntertainmentSplit ---

func TestCalculateEntertainmentSplit_70_30(t *testing.T) {
	exp := newTestExpense()
	exp.Type = ExpenseTypeEntertainment
	exp.Amount = 200.00
	exp.EntertainmentTip = 20.00
	exp.CalculateEntertainmentSplit()

	base := 200.00 - 20.00 // 180
	expectedDeductible := base * 0.70
	expectedNonDeductible := base * 0.30

	if exp.EntertainmentDeductible != expectedDeductible {
		t.Errorf("expected Deductible %f, got %f", expectedDeductible, exp.EntertainmentDeductible)
	}
	if exp.EntertainmentNonDeductible != expectedNonDeductible {
		t.Errorf("expected NonDeductible %f, got %f", expectedNonDeductible, exp.EntertainmentNonDeductible)
	}
}

func TestCalculateEntertainmentSplit_NoTip(t *testing.T) {
	exp := newTestExpense()
	exp.Type = ExpenseTypeEntertainment
	exp.Amount = 100.00
	exp.EntertainmentTip = 0
	exp.CalculateEntertainmentSplit()

	if exp.EntertainmentDeductible != 70.00 {
		t.Errorf("expected Deductible 70.00, got %f", exp.EntertainmentDeductible)
	}
	if exp.EntertainmentNonDeductible != 30.00 {
		t.Errorf("expected NonDeductible 30.00, got %f", exp.EntertainmentNonDeductible)
	}
}

func TestCalculateEntertainmentSplit_NotEntertainmentType(t *testing.T) {
	exp := newTestExpense()
	exp.Type = ExpenseTypeInvoice
	exp.Amount = 100.00
	exp.CalculateEntertainmentSplit()

	if exp.EntertainmentDeductible != 0 {
		t.Errorf("expected Deductible 0 for non-entertainment, got %f", exp.EntertainmentDeductible)
	}
	if exp.EntertainmentNonDeductible != 0 {
		t.Errorf("expected NonDeductible 0 for non-entertainment, got %f", exp.EntertainmentNonDeductible)
	}
}

// --- CalculateDiscountDeadline ---

func TestCalculateDiscountDeadline(t *testing.T) {
	exp := newTestExpense()
	exp.DiscountDays = 10
	exp.CalculateDiscountDeadline()

	expected := exp.Date.AddDate(0, 0, 10)
	if exp.DiscountDeadline == nil {
		t.Fatal("expected DiscountDeadline to be set")
	}
	if !exp.DiscountDeadline.Equal(expected) {
		t.Errorf("expected DiscountDeadline %v, got %v", expected, *exp.DiscountDeadline)
	}
}

func TestCalculateDiscountDeadline_ZeroDays(t *testing.T) {
	exp := newTestExpense()
	exp.DiscountDays = 0
	exp.CalculateDiscountDeadline()

	if exp.DiscountDeadline != nil {
		t.Errorf("expected DiscountDeadline nil when DiscountDays=0, got %v", exp.DiscountDeadline)
	}
}

// --- Validate ---

func TestValidate_OK(t *testing.T) {
	exp := newTestExpense()
	if err := exp.Validate(); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestValidate_EmptyID(t *testing.T) {
	exp := newTestExpense()
	exp.AggregateRoot = *events.NewAggregateRoot("", "expense")
	if err := exp.Validate(); err == nil {
		t.Error("expected error for empty ID")
	}
}

func TestValidate_EmptyTenantID(t *testing.T) {
	exp := newTestExpense()
	exp.TenantID = ""
	if err := exp.Validate(); err == nil {
		t.Error("expected error for empty TenantID")
	}
}

func TestValidate_EmptyVendor(t *testing.T) {
	exp := newTestExpense()
	exp.Vendor = ""
	if err := exp.Validate(); err == nil {
		t.Error("expected error for empty Vendor")
	}
}

func TestValidate_ZeroAmount(t *testing.T) {
	exp := newTestExpense()
	exp.Amount = 0
	if err := exp.Validate(); err == nil {
		t.Error("expected error for zero Amount")
	}
}

func TestValidate_NegativeAmount(t *testing.T) {
	exp := newTestExpense()
	exp.Amount = -50
	if err := exp.Validate(); err == nil {
		t.Error("expected error for negative Amount")
	}
}

// --- Status transitions ---

func TestSubmit_FromDraft(t *testing.T) {
	exp := newTestExpense()
	if err := exp.Submit(); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if exp.Status != StatusPending {
		t.Errorf("expected status pending, got %s", exp.Status)
	}
}

func TestSubmit_FromPending_Fails(t *testing.T) {
	exp := newTestExpense()
	_ = exp.Submit()
	if err := exp.Submit(); err == nil {
		t.Error("expected error when submitting from pending")
	}
}

func TestSubmit_FromApproved_Fails(t *testing.T) {
	exp := newTestExpense()
	_ = exp.Submit()
	_ = exp.Approve("admin")
	if err := exp.Submit(); err == nil {
		t.Error("expected error when submitting from approved")
	}
}

func TestApprove_FromPending(t *testing.T) {
	exp := newTestExpense()
	_ = exp.Submit()
	if err := exp.Approve("admin@test.com"); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if exp.Status != StatusApproved {
		t.Errorf("expected status approved, got %s", exp.Status)
	}
	if exp.ApprovedBy != "admin@test.com" {
		t.Errorf("expected ApprovedBy admin@test.com, got %s", exp.ApprovedBy)
	}
}

func TestApprove_FromDraft_Fails(t *testing.T) {
	exp := newTestExpense()
	if err := exp.Approve("admin"); err == nil {
		t.Error("expected error when approving from draft")
	}
}

func TestApprove_FromRejected_Fails(t *testing.T) {
	exp := newTestExpense()
	_ = exp.Submit()
	_ = exp.Reject()
	if err := exp.Approve("admin"); err == nil {
		t.Error("expected error when approving from rejected")
	}
}

func TestReject_FromPending(t *testing.T) {
	exp := newTestExpense()
	_ = exp.Submit()
	if err := exp.Reject(); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if exp.Status != StatusRejected {
		t.Errorf("expected status rejected, got %s", exp.Status)
	}
}

func TestReject_FromDraft_Fails(t *testing.T) {
	exp := newTestExpense()
	if err := exp.Reject(); err == nil {
		t.Error("expected error when rejecting from draft")
	}
}

func TestReject_FromApproved_Fails(t *testing.T) {
	exp := newTestExpense()
	_ = exp.Submit()
	_ = exp.Approve("admin")
	if err := exp.Reject(); err == nil {
		t.Error("expected error when rejecting from approved")
	}
}

// --- Full workflow: Draft -> Pending -> Approved ---

func TestFullWorkflow_DraftToPendingToApproved(t *testing.T) {
	exp := newTestExpense()

	if exp.Status != StatusDraft {
		t.Fatalf("initial status should be draft, got %s", exp.Status)
	}

	if err := exp.Submit(); err != nil {
		t.Fatalf("submit failed: %v", err)
	}
	if exp.Status != StatusPending {
		t.Fatalf("status after submit should be pending, got %s", exp.Status)
	}

	if err := exp.Approve("manager@test.com"); err != nil {
		t.Fatalf("approve failed: %v", err)
	}
	if exp.Status != StatusApproved {
		t.Fatalf("status after approve should be approved, got %s", exp.Status)
	}
}

func TestFullWorkflow_DraftToPendingToRejected(t *testing.T) {
	exp := newTestExpense()
	_ = exp.Submit()

	if err := exp.Reject(); err != nil {
		t.Fatalf("reject failed: %v", err)
	}
	if exp.Status != StatusRejected {
		t.Fatalf("status after reject should be rejected, got %s", exp.Status)
	}
}

// ===========================================================================
// Table-Driven: CalculateTax
// ===========================================================================

func TestCalculateTax_TableDriven(t *testing.T) {
	tests := []struct {
		name          string
		amount        float64
		taxRate       float64
		wantTaxAmount float64
		wantNetAmount float64
	}{
		{
			name:          "19% standard rate on 100 EUR",
			amount:        100.00,
			taxRate:       0.19,
			wantTaxAmount: 19.00,
			wantNetAmount: 81.00,
		},
		{
			name:          "7% reduced rate on 200 EUR",
			amount:        200.00,
			taxRate:       0.07,
			wantTaxAmount: 14.00,
			wantNetAmount: 186.00,
		},
		{
			name:          "0% Kleinunternehmer on 500 EUR",
			amount:        500.00,
			taxRate:       0.00,
			wantTaxAmount: 0.00,
			wantNetAmount: 500.00,
		},
		{
			name:          "19% on fractional amount",
			amount:        119.00,
			taxRate:       0.19,
			wantTaxAmount: 22.61,
			wantNetAmount: 96.39,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			exp := newTestExpense()
			exp.Amount = tc.amount
			exp.TaxRate = tc.taxRate
			exp.CalculateTax()

			if !floatEq(exp.TaxAmount, tc.wantTaxAmount) {
				t.Errorf("TaxAmount: expected %.2f, got %.2f", tc.wantTaxAmount, exp.TaxAmount)
			}
			if !floatEq(exp.NetAmount, tc.wantNetAmount) {
				t.Errorf("NetAmount: expected %.2f, got %.2f", tc.wantNetAmount, exp.NetAmount)
			}
		})
	}
}

// ===========================================================================
// Table-Driven: CalculateEntertainmentSplit
// ===========================================================================

func TestCalculateEntertainmentSplit_TableDriven(t *testing.T) {
	tests := []struct {
		name              string
		amount            float64
		tip               float64
		wantDeductible    float64
		wantNonDeductible float64
	}{
		{
			name:              "100 EUR no tip: 70/30",
			amount:            100.00,
			tip:               0.00,
			wantDeductible:    70.00,
			wantNonDeductible: 30.00,
		},
		{
			name:              "110 EUR with 10 EUR tip: base 100, 70/30",
			amount:            110.00,
			tip:               10.00,
			wantDeductible:    70.00,
			wantNonDeductible: 30.00,
		},
		{
			name:              "500 EUR with 50 EUR tip: base 450, 315/135",
			amount:            500.00,
			tip:               50.00,
			wantDeductible:    315.00,
			wantNonDeductible: 135.00,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			exp := newTestExpense()
			exp.Type = ExpenseTypeEntertainment
			exp.Amount = tc.amount
			exp.EntertainmentTip = tc.tip
			exp.CalculateEntertainmentSplit()

			if !floatEq(exp.EntertainmentDeductible, tc.wantDeductible) {
				t.Errorf("Deductible: expected %.2f, got %.2f", tc.wantDeductible, exp.EntertainmentDeductible)
			}
			if !floatEq(exp.EntertainmentNonDeductible, tc.wantNonDeductible) {
				t.Errorf("NonDeductible: expected %.2f, got %.2f", tc.wantNonDeductible, exp.EntertainmentNonDeductible)
			}
		})
	}
}

// ===========================================================================
// Table-Driven: Validate with missing fields
// ===========================================================================

func TestValidate_MissingFields_TableDriven(t *testing.T) {
	tests := []struct {
		name      string
		modify    func(*Expense)
		wantError string
	}{
		{
			name:      "valid expense passes",
			modify:    func(e *Expense) {},
			wantError: "",
		},
		{
			name:      "empty ID fails",
			modify:    func(e *Expense) { e.AggregateRoot = *events.NewAggregateRoot("", "expense") },
			wantError: "expense ID cannot be empty",
		},
		{
			name:      "empty tenant ID fails",
			modify:    func(e *Expense) { e.TenantID = "" },
			wantError: "tenant ID cannot be empty",
		},
		{
			name:      "empty vendor fails",
			modify:    func(e *Expense) { e.Vendor = "" },
			wantError: "vendor cannot be empty",
		},
		{
			name:      "zero amount fails",
			modify:    func(e *Expense) { e.Amount = 0 },
			wantError: "amount must be positive",
		},
		{
			name:      "negative amount fails",
			modify:    func(e *Expense) { e.Amount = -10 },
			wantError: "amount must be positive",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			exp := newTestExpense()
			tc.modify(exp)
			err := exp.Validate()

			if tc.wantError == "" {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
			} else {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tc.wantError)
				}
				if err.Error() != tc.wantError {
					t.Errorf("expected error %q, got %q", tc.wantError, err.Error())
				}
			}
		})
	}
}

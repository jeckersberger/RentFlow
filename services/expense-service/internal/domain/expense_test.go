package domain

import (
	"testing"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/events"
)

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
	if exp.TaxAmount != expectedTax {
		t.Errorf("expected TaxAmount %f, got %f", expectedTax, exp.TaxAmount)
	}
	if exp.NetAmount != expectedNet {
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

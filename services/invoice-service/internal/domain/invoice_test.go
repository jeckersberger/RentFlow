package domain

import (
	"testing"
	"time"
)

// --- Helper ---

func newTestInvoice() *Invoice {
	inv := NewInvoice("inv_001", "tenant_1", "RE-2026-001", "Musterfirma GmbH", "info@musterfirma.de")
	return inv
}

func addTestItem(inv *Invoice, desc string, qty, price float64, taxRate TaxRate) {
	_ = inv.AddItem(InvoiceItem{
		Description: desc,
		Quantity:    qty,
		UnitPrice:   price,
		TaxRate:     taxRate,
	})
}

// --- NewInvoice ---

func TestNewInvoice_Defaults(t *testing.T) {
	inv := newTestInvoice()

	if inv.ID != "inv_001" {
		t.Errorf("expected ID inv_001, got %s", inv.ID)
	}
	if inv.TenantID != "tenant_1" {
		t.Errorf("expected TenantID tenant_1, got %s", inv.TenantID)
	}
	if inv.InvoiceNumber != "RE-2026-001" {
		t.Errorf("expected InvoiceNumber RE-2026-001, got %s", inv.InvoiceNumber)
	}
	if inv.ClientName != "Musterfirma GmbH" {
		t.Errorf("expected ClientName Musterfirma GmbH, got %s", inv.ClientName)
	}
	if inv.Status != InvoiceDraft {
		t.Errorf("expected Status draft, got %s", inv.Status)
	}
	if inv.Currency != "EUR" {
		t.Errorf("expected Currency EUR, got %s", inv.Currency)
	}
	if len(inv.Items) != 0 {
		t.Errorf("expected 0 items, got %d", len(inv.Items))
	}
	if inv.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
	// DueDate should be 14 days after IssueDate
	expectedDue := inv.IssueDate.AddDate(0, 0, 14)
	if !inv.DueDate.Equal(expectedDue) {
		t.Errorf("expected DueDate %v, got %v", expectedDue, inv.DueDate)
	}
}

// --- AddItem ---

func TestAddItem_Success(t *testing.T) {
	inv := newTestInvoice()
	err := inv.AddItem(InvoiceItem{
		Description: "PA-Anlage Vermietung",
		Quantity:    2,
		UnitPrice:   150.00,
		TaxRate:     19,
	})
	if err != nil {
		t.Fatalf("AddItem failed: %v", err)
	}
	if len(inv.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(inv.Items))
	}
	item := inv.Items[0]
	if item.TotalPrice != 300.00 {
		t.Errorf("expected TotalPrice 300.00, got %f", item.TotalPrice)
	}
	if item.TaxAmount != 57.00 {
		t.Errorf("expected TaxAmount 57.00, got %f", item.TaxAmount)
	}
	if item.ID != "item_1" {
		t.Errorf("expected ID item_1, got %s", item.ID)
	}
}

func TestAddItem_FallbackToInvoiceTaxRate(t *testing.T) {
	inv := newTestInvoice()
	inv.TaxRate = 19
	err := inv.AddItem(InvoiceItem{
		Description: "Lichtanlage",
		Quantity:    1,
		UnitPrice:   500.00,
		// TaxRate 0 -> should fall back to invoice default 19
	})
	if err != nil {
		t.Fatalf("AddItem failed: %v", err)
	}
	if inv.Items[0].TaxRate != 19 {
		t.Errorf("expected item TaxRate 19, got %f", float64(inv.Items[0].TaxRate))
	}
}

func TestAddItem_EmptyDescription_Fails(t *testing.T) {
	inv := newTestInvoice()
	err := inv.AddItem(InvoiceItem{Description: "", Quantity: 1, UnitPrice: 10})
	if err != ErrInvalidInput {
		t.Errorf("expected ErrInvalidInput, got %v", err)
	}
}

func TestAddItem_ZeroQuantity_Fails(t *testing.T) {
	inv := newTestInvoice()
	err := inv.AddItem(InvoiceItem{Description: "Test", Quantity: 0, UnitPrice: 10})
	if err != ErrInvalidInput {
		t.Errorf("expected ErrInvalidInput, got %v", err)
	}
}

func TestAddItem_NegativePrice_Fails(t *testing.T) {
	inv := newTestInvoice()
	err := inv.AddItem(InvoiceItem{Description: "Test", Quantity: 1, UnitPrice: -5})
	if err != ErrInvalidInput {
		t.Errorf("expected ErrInvalidInput, got %v", err)
	}
}

// --- RemoveItem ---

func TestRemoveItem_Success(t *testing.T) {
	inv := newTestInvoice()
	addTestItem(inv, "Item A", 1, 100, 19)
	addTestItem(inv, "Item B", 1, 200, 19)

	err := inv.RemoveItem("item_1")
	if err != nil {
		t.Fatalf("RemoveItem failed: %v", err)
	}
	if len(inv.Items) != 1 {
		t.Errorf("expected 1 item after removal, got %d", len(inv.Items))
	}
	if inv.Items[0].Description != "Item B" {
		t.Errorf("expected remaining item to be Item B, got %s", inv.Items[0].Description)
	}
}

func TestRemoveItem_NotFound_Fails(t *testing.T) {
	inv := newTestInvoice()
	err := inv.RemoveItem("nonexistent")
	if err != ErrInvalidInput {
		t.Errorf("expected ErrInvalidInput, got %v", err)
	}
}

// --- SetTaxRate ---

func TestSetTaxRate_Valid(t *testing.T) {
	inv := newTestInvoice()
	for _, rate := range []float64{0, 7, 19} {
		if err := inv.SetTaxRate(rate); err != nil {
			t.Errorf("SetTaxRate(%f) failed: %v", rate, err)
		}
	}
}

func TestSetTaxRate_Invalid(t *testing.T) {
	inv := newTestInvoice()
	for _, rate := range []float64{5, 10, 20, -1} {
		if err := inv.SetTaxRate(rate); err != ErrInvalidTaxRate {
			t.Errorf("SetTaxRate(%f) expected ErrInvalidTaxRate, got %v", rate, err)
		}
	}
}

// --- CalculateTotals ---

func TestCalculateTotals_SingleItem_19Percent(t *testing.T) {
	inv := newTestInvoice()
	addTestItem(inv, "Mischpult Vermietung", 1, 100.00, 19)

	err := inv.CalculateTotals()
	if err != nil {
		t.Fatalf("CalculateTotals failed: %v", err)
	}
	if inv.SubTotal != 100.00 {
		t.Errorf("expected SubTotal 100.00, got %f", inv.SubTotal)
	}
	if inv.TaxAmount != 19.00 {
		t.Errorf("expected TaxAmount 19.00, got %f", inv.TaxAmount)
	}
	if inv.Total != 119.00 {
		t.Errorf("expected Total 119.00, got %f", inv.Total)
	}
}

func TestCalculateTotals_MultipleItems_MixedTax(t *testing.T) {
	inv := newTestInvoice()
	addTestItem(inv, "PA Vermietung", 2, 100.00, 19)  // 200 + 38 tax
	addTestItem(inv, "Lieferpauschale", 1, 50.00, 7)   // 50 + 3.5 tax

	err := inv.CalculateTotals()
	if err != nil {
		t.Fatalf("CalculateTotals failed: %v", err)
	}

	expectedSub := 250.00
	expectedTax := 41.50
	expectedTotal := 291.50

	if inv.SubTotal != expectedSub {
		t.Errorf("expected SubTotal %f, got %f", expectedSub, inv.SubTotal)
	}
	if inv.TaxAmount != expectedTax {
		t.Errorf("expected TaxAmount %f, got %f", expectedTax, inv.TaxAmount)
	}
	if inv.Total != expectedTotal {
		t.Errorf("expected Total %f, got %f", expectedTotal, inv.Total)
	}
}

func TestCalculateTotals_0PercentTax(t *testing.T) {
	inv := newTestInvoice()
	addTestItem(inv, "Beratung", 1, 500.00, 0)

	err := inv.CalculateTotals()
	if err != nil {
		t.Fatalf("CalculateTotals failed: %v", err)
	}

	if inv.TaxAmount != 0 {
		t.Errorf("expected TaxAmount 0, got %f", inv.TaxAmount)
	}
	if inv.Total != 500.00 {
		t.Errorf("expected Total 500.00, got %f", inv.Total)
	}
}

func TestCalculateTotals_NoItems_Fails(t *testing.T) {
	inv := newTestInvoice()
	err := inv.CalculateTotals()
	if err != ErrNoItems {
		t.Errorf("expected ErrNoItems, got %v", err)
	}
}

func TestCalculateTotals_Kleinunternehmer(t *testing.T) {
	inv := newTestInvoice()
	inv.IsKleinunternehmer = true
	addTestItem(inv, "DJ Set", 1, 300.00, 19) // Tax should be zeroed

	err := inv.CalculateTotals()
	if err != nil {
		t.Fatalf("CalculateTotals failed: %v", err)
	}

	if inv.TaxAmount != 0 {
		t.Errorf("expected TaxAmount 0 for Kleinunternehmer, got %f", inv.TaxAmount)
	}
	if inv.Total != 300.00 {
		t.Errorf("expected Total 300.00 for Kleinunternehmer, got %f", inv.Total)
	}
	if inv.Items[0].TaxRate != 0 {
		t.Errorf("expected item TaxRate 0, got %f", float64(inv.Items[0].TaxRate))
	}
	if inv.KleinunternehmerText == "" {
		t.Error("expected KleinunternehmerText to be set")
	}
}

// --- Status transitions ---

func TestSend_FromDraft(t *testing.T) {
	inv := newTestInvoice()
	err := inv.Send("info@musterfirma.de")
	if err != nil {
		t.Fatalf("Send failed: %v", err)
	}
	if inv.Status != InvoiceSent {
		t.Errorf("expected status sent, got %s", inv.Status)
	}
}

func TestSend_FromSent_Fails(t *testing.T) {
	inv := newTestInvoice()
	_ = inv.Send("info@musterfirma.de")
	err := inv.Send("info@musterfirma.de")
	if err != ErrInvalidTransition {
		t.Errorf("expected ErrInvalidTransition, got %v", err)
	}
}

func TestMarkPaid_FromSent(t *testing.T) {
	inv := newTestInvoice()
	addTestItem(inv, "Vermietung", 1, 100, 19)
	_ = inv.CalculateTotals()
	_ = inv.Send("info@test.de")

	err := inv.MarkPaid("bank_transfer", "REF-123")
	if err != nil {
		t.Fatalf("MarkPaid failed: %v", err)
	}
	if inv.Status != InvoicePaid {
		t.Errorf("expected status paid, got %s", inv.Status)
	}
	if inv.PaidDate == nil {
		t.Error("expected PaidDate to be set")
	}
	if inv.PaymentMethod != "bank_transfer" {
		t.Errorf("expected PaymentMethod bank_transfer, got %s", inv.PaymentMethod)
	}
	if inv.PaidAmount != inv.Total {
		t.Errorf("expected PaidAmount %f, got %f", inv.Total, inv.PaidAmount)
	}
	if inv.RemainingAmount != 0 {
		t.Errorf("expected RemainingAmount 0, got %f", inv.RemainingAmount)
	}
}

func TestMarkPaid_FromPaid_Fails(t *testing.T) {
	inv := newTestInvoice()
	addTestItem(inv, "Vermietung", 1, 100, 19)
	_ = inv.CalculateTotals()
	_ = inv.Send("info@test.de")
	_ = inv.MarkPaid("bank", "REF")

	err := inv.MarkPaid("bank", "REF2")
	if err != ErrInvalidTransition {
		t.Errorf("expected ErrInvalidTransition, got %v", err)
	}
}

func TestMarkPaid_FromCancelled_Fails(t *testing.T) {
	inv := newTestInvoice()
	addTestItem(inv, "Vermietung", 1, 100, 19)
	_ = inv.CalculateTotals()
	_ = inv.Send("info@test.de")
	_ = inv.Cancel("Fehler")

	err := inv.MarkPaid("bank", "REF")
	if err != ErrInvalidTransition {
		t.Errorf("expected ErrInvalidTransition, got %v", err)
	}
}

func TestCancel_FromSent(t *testing.T) {
	inv := newTestInvoice()
	_ = inv.Send("info@test.de")

	err := inv.Cancel("Storno auf Kundenwunsch")
	if err != nil {
		t.Fatalf("Cancel failed: %v", err)
	}
	if inv.Status != InvoiceCancelled {
		t.Errorf("expected status cancelled, got %s", inv.Status)
	}
}

func TestCancel_FromPaid_Fails(t *testing.T) {
	inv := newTestInvoice()
	addTestItem(inv, "Vermietung", 1, 100, 19)
	_ = inv.CalculateTotals()
	_ = inv.Send("info@test.de")
	_ = inv.MarkPaid("bank", "REF")

	err := inv.Cancel("reason")
	if err != ErrInvalidTransition {
		t.Errorf("expected ErrInvalidTransition, got %v", err)
	}
}

func TestCancel_FromCancelled_Fails(t *testing.T) {
	inv := newTestInvoice()
	_ = inv.Send("info@test.de")
	_ = inv.Cancel("reason1")

	err := inv.Cancel("reason2")
	if err != ErrInvalidTransition {
		t.Errorf("expected ErrInvalidTransition, got %v", err)
	}
}

func TestCredit_FromSent(t *testing.T) {
	inv := newTestInvoice()
	_ = inv.Send("info@test.de")

	err := inv.Credit("credit_001")
	if err != nil {
		t.Fatalf("Credit failed: %v", err)
	}
	if inv.Status != InvoiceCredited {
		t.Errorf("expected status credited, got %s", inv.Status)
	}
}

func TestCredit_FromCancelled_Fails(t *testing.T) {
	inv := newTestInvoice()
	_ = inv.Send("info@test.de")
	_ = inv.Cancel("reason")

	err := inv.Credit("credit_001")
	if err != ErrInvalidTransition {
		t.Errorf("expected ErrInvalidTransition, got %v", err)
	}
}

// --- RecordPayment (partial) ---

func TestRecordPayment_Partial(t *testing.T) {
	inv := newTestInvoice()
	addTestItem(inv, "PA Vermietung", 1, 1000, 19)
	_ = inv.CalculateTotals()
	inv.RemainingAmount = inv.Total
	_ = inv.Send("info@test.de")

	err := inv.RecordPayment(500.00)
	if err != nil {
		t.Fatalf("RecordPayment failed: %v", err)
	}
	if inv.Status != InvoicePartiallyPaid {
		t.Errorf("expected status partially_paid, got %s", inv.Status)
	}
	if inv.PaidAmount != 500.00 {
		t.Errorf("expected PaidAmount 500.00, got %f", inv.PaidAmount)
	}
	if inv.RemainingAmount != inv.Total-500.00 {
		t.Errorf("expected RemainingAmount %f, got %f", inv.Total-500.00, inv.RemainingAmount)
	}
}

func TestRecordPayment_FullPayment(t *testing.T) {
	inv := newTestInvoice()
	addTestItem(inv, "PA Vermietung", 1, 100, 19)
	_ = inv.CalculateTotals()
	inv.RemainingAmount = inv.Total
	_ = inv.Send("info@test.de")

	err := inv.RecordPayment(inv.Total)
	if err != nil {
		t.Fatalf("RecordPayment failed: %v", err)
	}
	if inv.Status != InvoicePaid {
		t.Errorf("expected status paid after full payment, got %s", inv.Status)
	}
}

func TestRecordPayment_ExceedsRemaining_Fails(t *testing.T) {
	inv := newTestInvoice()
	addTestItem(inv, "Vermietung", 1, 100, 19)
	_ = inv.CalculateTotals()
	inv.RemainingAmount = inv.Total
	_ = inv.Send("info@test.de")

	err := inv.RecordPayment(inv.Total + 1)
	if err != ErrInvalidInput {
		t.Errorf("expected ErrInvalidInput, got %v", err)
	}
}

func TestRecordPayment_ZeroAmount_Fails(t *testing.T) {
	inv := newTestInvoice()
	addTestItem(inv, "Vermietung", 1, 100, 19)
	_ = inv.CalculateTotals()
	inv.RemainingAmount = inv.Total
	_ = inv.Send("info@test.de")

	err := inv.RecordPayment(0)
	if err != ErrInvalidInput {
		t.Errorf("expected ErrInvalidInput, got %v", err)
	}
}

// --- Validate ---

func TestValidate_OK(t *testing.T) {
	inv := newTestInvoice()
	inv.TaxRate = 19
	addTestItem(inv, "Vermietung", 1, 100, 19)
	_ = inv.CalculateTotals()

	if err := inv.Validate(); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestValidate_EmptyTenantID(t *testing.T) {
	inv := newTestInvoice()
	inv.TenantID = ""
	if err := inv.Validate(); err != ErrTenantIDRequired {
		t.Errorf("expected ErrTenantIDRequired, got %v", err)
	}
}

func TestValidate_EmptyInvoiceNumber(t *testing.T) {
	inv := newTestInvoice()
	inv.InvoiceNumber = ""
	if err := inv.Validate(); err != ErrInvalidInput {
		t.Errorf("expected ErrInvalidInput, got %v", err)
	}
}

func TestValidate_EmptyClientName(t *testing.T) {
	inv := newTestInvoice()
	inv.ClientName = ""
	if err := inv.Validate(); err != ErrInvalidInput {
		t.Errorf("expected ErrInvalidInput, got %v", err)
	}
}

func TestValidate_NoItems(t *testing.T) {
	inv := newTestInvoice()
	if err := inv.Validate(); err != ErrNoItems {
		t.Errorf("expected ErrNoItems, got %v", err)
	}
}

func TestValidate_ZeroTotal(t *testing.T) {
	inv := newTestInvoice()
	addTestItem(inv, "Vermietung", 1, 100, 19)
	_ = inv.CalculateTotals()
	inv.Total = 0
	inv.TaxRate = 19

	if err := inv.Validate(); err != ErrZeroAmount {
		t.Errorf("expected ErrZeroAmount, got %v", err)
	}
}

func TestValidate_InvalidTaxRate(t *testing.T) {
	inv := newTestInvoice()
	addTestItem(inv, "Vermietung", 1, 100, 19)
	_ = inv.CalculateTotals()
	inv.TaxRate = 15 // invalid

	if err := inv.Validate(); err != ErrInvalidTaxRate {
		t.Errorf("expected ErrInvalidTaxRate, got %v", err)
	}
}

// --- GoBD Hash ---

func TestComputeHash_Deterministic(t *testing.T) {
	inv := newTestInvoice()
	inv.IssueDate = time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	inv.DueDate = time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)
	addTestItem(inv, "Vermietung", 1, 100, 19)
	_ = inv.CalculateTotals()

	hash1 := inv.ComputeHash()
	hash2 := inv.ComputeHash()

	if hash1 == "" {
		t.Error("expected non-empty hash")
	}
	if hash1 != hash2 {
		t.Errorf("hash should be deterministic, got %s vs %s", hash1, hash2)
	}
}

func TestComputeHash_ChangesOnDataChange(t *testing.T) {
	inv := newTestInvoice()
	inv.IssueDate = time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	inv.DueDate = time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)
	addTestItem(inv, "Vermietung", 1, 100, 19)
	_ = inv.CalculateTotals()
	hash1 := inv.ComputeHash()

	// Change financial data
	inv.SubTotal = 200.00
	hash2 := inv.ComputeHash()

	if hash1 == hash2 {
		t.Error("hash should change when financial data changes")
	}
}

func TestComputeHash_IncludesKleinunternehmer(t *testing.T) {
	inv := newTestInvoice()
	inv.IssueDate = time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	inv.DueDate = time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)
	addTestItem(inv, "Vermietung", 1, 100, 19)
	_ = inv.CalculateTotals()
	hash1 := inv.ComputeHash()

	inv.IsKleinunternehmer = true
	hash2 := inv.ComputeHash()

	if hash1 == hash2 {
		t.Error("hash should change when Kleinunternehmer flag changes")
	}
}

func TestVerifyHash_Valid(t *testing.T) {
	inv := newTestInvoice()
	inv.IssueDate = time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	inv.DueDate = time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)
	addTestItem(inv, "Vermietung", 1, 100, 19)
	_ = inv.CalculateTotals()
	inv.Hash = inv.ComputeHash()

	if !inv.VerifyHash() {
		t.Error("expected VerifyHash to return true for valid hash")
	}
}

func TestVerifyHash_Invalid(t *testing.T) {
	inv := newTestInvoice()
	inv.IssueDate = time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	inv.DueDate = time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)
	addTestItem(inv, "Vermietung", 1, 100, 19)
	_ = inv.CalculateTotals()
	inv.Hash = "tampered_hash"

	if inv.VerifyHash() {
		t.Error("expected VerifyHash to return false for tampered hash")
	}
}

// --- IsFinalized ---

func TestIsFinalized_Draft(t *testing.T) {
	inv := newTestInvoice()
	if inv.IsFinalized() {
		t.Error("draft invoice should not be finalized")
	}
}

func TestIsFinalized_Sent(t *testing.T) {
	inv := newTestInvoice()
	_ = inv.Send("info@test.de")
	if !inv.IsFinalized() {
		t.Error("sent invoice should be finalized")
	}
}

// --- TaxRate.IsValidTaxRate ---

func TestIsValidTaxRate(t *testing.T) {
	tests := []struct {
		rate  TaxRate
		valid bool
	}{
		{0, true},
		{7, true},
		{19, true},
		{5, false},
		{10, false},
		{20, false},
	}
	for _, tc := range tests {
		if tc.rate.IsValidTaxRate() != tc.valid {
			t.Errorf("TaxRate(%f).IsValidTaxRate() = %v, want %v", float64(tc.rate), !tc.valid, tc.valid)
		}
	}
}

// --- InvoiceStatus.IsValidStatus ---

func TestIsValidStatus(t *testing.T) {
	validStatuses := []InvoiceStatus{
		InvoiceDraft, InvoiceSent, InvoiceOverdue, InvoicePaid,
		InvoicePartiallyPaid, InvoiceCancelled, InvoiceCredited,
	}
	for _, s := range validStatuses {
		if !s.IsValidStatus() {
			t.Errorf("expected %s to be valid", s)
		}
	}
	invalid := InvoiceStatus("unknown")
	if invalid.IsValidStatus() {
		t.Error("expected 'unknown' to be invalid")
	}
}

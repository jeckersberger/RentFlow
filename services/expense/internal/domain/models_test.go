package domain

import (
	"testing"

	"github.com/google/uuid"
)

func TestExpenseStatusConstants(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{"Pending", StatusPending, "pending"},
		{"Approved", StatusApproved, "approved"},
		{"Rejected", StatusRejected, "rejected"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("%s = %q, want %q", tt.name, tt.got, tt.want)
			}
		})
	}
}

func TestExpenseAmountInCents(t *testing.T) {
	e := &Expense{
		ID:          uuid.New(),
		Description: "Kabeltrommel 50m",
		Amount:      8990, // 89.90 EUR
		Currency:    "EUR",
		Status:      StatusPending,
	}

	if e.Amount != 8990 {
		t.Errorf("Amount = %d, want 8990", e.Amount)
	}
}

func TestBudgetAlertThreshold(t *testing.T) {
	b := &Budget{
		Name:              "Monatliches Materialbudget",
		Amount:            500000, // 5000 EUR
		AlertThresholdPct: 80,
		IsActive:          true,
		SpentAmount:       420000, // 4200 EUR = 84%
	}

	pct := int(b.SpentAmount * 100 / b.Amount)
	if pct != 84 {
		t.Errorf("spent percentage = %d, want 84", pct)
	}
	if pct < b.AlertThresholdPct {
		t.Error("should be above alert threshold")
	}
}

func TestBudgetBelowThreshold(t *testing.T) {
	b := &Budget{
		Amount:            1000000,
		AlertThresholdPct: 80,
		SpentAmount:       500000, // 50%
	}

	pct := int(b.SpentAmount * 100 / b.Amount)
	if pct >= b.AlertThresholdPct {
		t.Error("should be below alert threshold at 50%")
	}
}

func TestRecurringExpenseFrequency(t *testing.T) {
	frequencies := []string{"monthly", "quarterly", "yearly"}
	for _, f := range frequencies {
		r := &RecurringExpense{
			Name:      "Miete",
			Amount:    300000,
			Frequency: f,
			IsActive:  true,
		}
		if r.Frequency == "" {
			t.Errorf("Frequency should not be empty for %s", f)
		}
	}
}

func TestExpenseReceiptFileSize(t *testing.T) {
	receipt := &ExpenseReceipt{
		FileName: "beleg.pdf",
		FileSize: 1500000, // 1.5 MB
		MimeType: "application/pdf",
	}

	maxSize := 10 * 1024 * 1024 // 10 MB
	if receipt.FileSize > maxSize {
		t.Errorf("FileSize %d exceeds max %d", receipt.FileSize, maxSize)
	}
}

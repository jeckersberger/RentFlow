package application

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/expense-service/internal/domain"
)

// --- Mock Repository ---

type mockExpenseRepo struct {
	expenses map[string]*domain.Expense
	createFn func(ctx context.Context, expense *domain.Expense) error
	updateFn func(ctx context.Context, expense *domain.Expense) error
}

func newMockExpenseRepo() *mockExpenseRepo {
	return &mockExpenseRepo{
		expenses: make(map[string]*domain.Expense),
	}
}

func (m *mockExpenseRepo) CreateExpense(ctx context.Context, expense *domain.Expense) error {
	if m.createFn != nil {
		return m.createFn(ctx, expense)
	}
	m.expenses[expense.ID] = expense
	return nil
}

func (m *mockExpenseRepo) UpdateExpense(ctx context.Context, expense *domain.Expense) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, expense)
	}
	m.expenses[expense.ID] = expense
	return nil
}

func (m *mockExpenseRepo) GetExpense(ctx context.Context, tenantID, expenseID string) (*domain.Expense, error) {
	exp, ok := m.expenses[expenseID]
	if !ok || exp.TenantID != tenantID {
		return nil, fmt.Errorf("not found")
	}
	return exp, nil
}

func (m *mockExpenseRepo) ListExpenses(ctx context.Context, tenantID string, limit, offset int) ([]*domain.Expense, int64, error) {
	var result []*domain.Expense
	for _, e := range m.expenses {
		if e.TenantID == tenantID {
			result = append(result, e)
		}
	}
	return result, int64(len(result)), nil
}

func (m *mockExpenseRepo) DeleteExpense(ctx context.Context, tenantID, expenseID string) error {
	delete(m.expenses, expenseID)
	return nil
}

// --- Mock Logger ---

type mockLogger struct{}

func (l *mockLogger) Debug(msg string, args ...interface{})          {}
func (l *mockLogger) Info(msg string, args ...interface{})           {}
func (l *mockLogger) Warn(msg string, args ...interface{})           {}
func (l *mockLogger) Error(msg string, args ...interface{})          {}
func (l *mockLogger) Fatal(msg string, args ...interface{})          {}
func (l *mockLogger) WithPrefix(prefix string) logger.Logger         { return l }
func (l *mockLogger) WithCorrelationID(id string) logger.Logger      { return l }
func (l *mockLogger) WithRequestID(id string) logger.Logger          { return l }
func (l *mockLogger) WithField(key string, value interface{}) logger.Logger { return l }

// --- Test helpers ---

func newTestService() (*ExpenseService, *mockExpenseRepo) {
	repo := newMockExpenseRepo()
	svc := NewExpenseService(repo, &mockLogger{})
	return svc, repo
}

func baseCreateCmd() CreateExpenseCommand {
	return CreateExpenseCommand{
		TenantID:      "tenant_1",
		Vendor:        "Thomann GmbH",
		Amount:        119.00,
		Currency:      "EUR",
		CategoryCode:  "4900",
		PaymentMethod: "bank_transfer",
		Date:          time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
	}
}

// --- CreateExpense Tests ---

func TestCreateExpense_Basic(t *testing.T) {
	svc, repo := newTestService()
	ctx := context.Background()

	cmd := baseCreateCmd()
	dto, err := svc.CreateExpense(ctx, cmd)
	if err != nil {
		t.Fatalf("CreateExpense failed: %v", err)
	}

	if dto.Vendor != "Thomann GmbH" {
		t.Errorf("expected Vendor Thomann GmbH, got %s", dto.Vendor)
	}
	if dto.Status != "draft" {
		t.Errorf("expected Status draft, got %s", dto.Status)
	}
	if dto.PaymentStatus != "open" {
		t.Errorf("expected PaymentStatus open, got %s", dto.PaymentStatus)
	}
	if len(repo.expenses) != 1 {
		t.Errorf("expected 1 expense in repo, got %d", len(repo.expenses))
	}
}

func TestCreateExpense_TaxCalculation(t *testing.T) {
	svc, _ := newTestService()
	ctx := context.Background()

	cmd := baseCreateCmd()
	cmd.TaxRate = 0.19
	cmd.Amount = 100.00
	dto, err := svc.CreateExpense(ctx, cmd)
	if err != nil {
		t.Fatalf("CreateExpense failed: %v", err)
	}

	if dto.TaxAmount != 19.00 {
		t.Errorf("expected TaxAmount 19.00, got %f", dto.TaxAmount)
	}
	if dto.NetAmount != 81.00 {
		t.Errorf("expected NetAmount 81.00, got %f", dto.NetAmount)
	}
}

func TestCreateExpense_Entertainment(t *testing.T) {
	svc, _ := newTestService()
	ctx := context.Background()

	cmd := baseCreateCmd()
	cmd.Type = "entertainment"
	cmd.Amount = 200.00
	cmd.EntertainmentLocation = "Restaurant Muehlbauer"
	cmd.EntertainmentReason = "Kundentermin Projekt Alpha"
	cmd.EntertainmentGuests = "Max Mustermann, Erika Muster"
	cmd.EntertainmentTip = 20.00

	dto, err := svc.CreateExpense(ctx, cmd)
	if err != nil {
		t.Fatalf("CreateExpense failed: %v", err)
	}

	if dto.Type != "entertainment" {
		t.Errorf("expected Type entertainment, got %s", dto.Type)
	}
	if dto.EntertainmentLocation != "Restaurant Muehlbauer" {
		t.Errorf("expected location Restaurant Muehlbauer, got %s", dto.EntertainmentLocation)
	}

	// 70/30 split: base = 200 - 20 = 180
	expectedDeductible := 180.0 * 0.70
	expectedNonDeductible := 180.0 * 0.30
	if dto.EntertainmentDeductible != expectedDeductible {
		t.Errorf("expected Deductible %f, got %f", expectedDeductible, dto.EntertainmentDeductible)
	}
	if dto.EntertainmentNonDeductible != expectedNonDeductible {
		t.Errorf("expected NonDeductible %f, got %f", expectedNonDeductible, dto.EntertainmentNonDeductible)
	}
}

func TestCreateExpense_Skonto(t *testing.T) {
	svc, _ := newTestService()
	ctx := context.Background()

	cmd := baseCreateCmd()
	cmd.DiscountPercent = 2.0
	cmd.DiscountDays = 14

	dto, err := svc.CreateExpense(ctx, cmd)
	if err != nil {
		t.Fatalf("CreateExpense failed: %v", err)
	}

	if dto.DiscountPercent != 2.0 {
		t.Errorf("expected DiscountPercent 2.0, got %f", dto.DiscountPercent)
	}
	if dto.DiscountDays != 14 {
		t.Errorf("expected DiscountDays 14, got %d", dto.DiscountDays)
	}
	if dto.DiscountDeadline == nil {
		t.Fatal("expected DiscountDeadline to be set")
	}
	expected := cmd.Date.AddDate(0, 0, 14)
	if !dto.DiscountDeadline.Equal(expected) {
		t.Errorf("expected DiscountDeadline %v, got %v", expected, *dto.DiscountDeadline)
	}
}

func TestCreateExpense_OCRFields(t *testing.T) {
	svc, _ := newTestService()
	ctx := context.Background()

	cmd := baseCreateCmd()
	cmd.Source = "photo"
	cmd.OCRData = map[string]string{"vendor": "Thomann", "total": "119.00"}
	cmd.OCRConfidence = 0.95
	cmd.ReceiptChecksum = "sha256:abc123"
	cmd.ReceiptNASPath = "/nas/receipts/2026/03/exp_123.pdf"

	dto, err := svc.CreateExpense(ctx, cmd)
	if err != nil {
		t.Fatalf("CreateExpense failed: %v", err)
	}

	if dto.Source != "photo" {
		t.Errorf("expected Source photo, got %s", dto.Source)
	}
	if dto.OCRConfidence != 0.95 {
		t.Errorf("expected OCRConfidence 0.95, got %f", dto.OCRConfidence)
	}
	if dto.ReceiptChecksum != "sha256:abc123" {
		t.Errorf("expected ReceiptChecksum sha256:abc123, got %s", dto.ReceiptChecksum)
	}
}

func TestCreateExpense_EmptyTenant_Fails(t *testing.T) {
	svc, _ := newTestService()
	ctx := context.Background()

	cmd := baseCreateCmd()
	cmd.TenantID = ""

	_, err := svc.CreateExpense(ctx, cmd)
	if err == nil {
		t.Error("expected error for empty TenantID")
	}
}

func TestCreateExpense_RepoError(t *testing.T) {
	svc, repo := newTestService()
	repo.createFn = func(ctx context.Context, expense *domain.Expense) error {
		return fmt.Errorf("db connection failed")
	}
	ctx := context.Background()

	cmd := baseCreateCmd()
	_, err := svc.CreateExpense(ctx, cmd)
	if err == nil {
		t.Error("expected error when repo fails")
	}
}

// --- UpdateExpense Tests ---

func TestUpdateExpense_PaymentStatus(t *testing.T) {
	svc, _ := newTestService()
	ctx := context.Background()

	// First create
	cmd := baseCreateCmd()
	dto, _ := svc.CreateExpense(ctx, cmd)

	// Then update payment status to paid
	updateCmd := UpdateExpenseCommand{
		ID:            dto.ID,
		TenantID:      "tenant_1",
		PaymentStatus: "paid",
	}
	updated, err := svc.UpdateExpense(ctx, updateCmd)
	if err != nil {
		t.Fatalf("UpdateExpense failed: %v", err)
	}

	if updated.PaymentStatus != "paid" {
		t.Errorf("expected PaymentStatus paid, got %s", updated.PaymentStatus)
	}
	if updated.PaidAt == nil {
		t.Error("expected PaidAt to be set when status changes to paid")
	}
}

func TestUpdateExpense_VendorFields(t *testing.T) {
	svc, _ := newTestService()
	ctx := context.Background()

	cmd := baseCreateCmd()
	dto, _ := svc.CreateExpense(ctx, cmd)

	updateCmd := UpdateExpenseCommand{
		ID:            dto.ID,
		TenantID:      "tenant_1",
		VendorAddress: "Treppendorf 30, 96138 Burgebrach",
		VendorVATID:   "DE123456789",
		VendorIBAN:    "DE89370400440532013000",
	}
	updated, err := svc.UpdateExpense(ctx, updateCmd)
	if err != nil {
		t.Fatalf("UpdateExpense failed: %v", err)
	}

	if updated.VendorAddress != "Treppendorf 30, 96138 Burgebrach" {
		t.Errorf("unexpected VendorAddress: %s", updated.VendorAddress)
	}
	if updated.VendorVATID != "DE123456789" {
		t.Errorf("unexpected VendorVATID: %s", updated.VendorVATID)
	}
	if updated.VendorIBAN != "DE89370400440532013000" {
		t.Errorf("unexpected VendorIBAN: %s", updated.VendorIBAN)
	}
}

func TestUpdateExpense_EntertainmentUpdate(t *testing.T) {
	svc, _ := newTestService()
	ctx := context.Background()

	cmd := baseCreateCmd()
	cmd.Type = "entertainment"
	cmd.Amount = 100.00
	dto, _ := svc.CreateExpense(ctx, cmd)

	updateCmd := UpdateExpenseCommand{
		ID:                    dto.ID,
		TenantID:              "tenant_1",
		EntertainmentLocation: "Ristorante Roma",
		EntertainmentReason:   "Teammeeting",
		EntertainmentGuests:   "Hans, Peter",
		EntertainmentTip:      10.00,
	}
	updated, err := svc.UpdateExpense(ctx, updateCmd)
	if err != nil {
		t.Fatalf("UpdateExpense failed: %v", err)
	}

	// base = 100 - 10 = 90
	if updated.EntertainmentDeductible != 63.0 {
		t.Errorf("expected Deductible 63.0, got %f", updated.EntertainmentDeductible)
	}
	if updated.EntertainmentNonDeductible != 27.0 {
		t.Errorf("expected NonDeductible 27.0, got %f", updated.EntertainmentNonDeductible)
	}
}

func TestUpdateExpense_EmptyTenant_Fails(t *testing.T) {
	svc, _ := newTestService()
	ctx := context.Background()

	_, err := svc.UpdateExpense(ctx, UpdateExpenseCommand{ID: "x", TenantID: ""})
	if err == nil {
		t.Error("expected error for empty TenantID")
	}
}

func TestUpdateExpense_NotFound_Fails(t *testing.T) {
	svc, _ := newTestService()
	ctx := context.Background()

	_, err := svc.UpdateExpense(ctx, UpdateExpenseCommand{ID: "nonexistent", TenantID: "tenant_1"})
	if err == nil {
		t.Error("expected error for nonexistent expense")
	}
}

// --- Approve/Reject Workflow Tests ---

func TestApproveExpense_Success(t *testing.T) {
	svc, repo := newTestService()
	ctx := context.Background()

	// Create and move to pending
	cmd := baseCreateCmd()
	dto, _ := svc.CreateExpense(ctx, cmd)
	exp := repo.expenses[dto.ID]
	_ = exp.Submit()

	result, err := svc.ApproveExpense(ctx, "tenant_1", dto.ID, "admin@test.com")
	if err != nil {
		t.Fatalf("ApproveExpense failed: %v", err)
	}
	if result.Status != "approved" {
		t.Errorf("expected status approved, got %s", result.Status)
	}
	if result.ApprovedBy != "admin@test.com" {
		t.Errorf("expected ApprovedBy admin@test.com, got %s", result.ApprovedBy)
	}
}

func TestApproveExpense_FromDraft_Fails(t *testing.T) {
	svc, _ := newTestService()
	ctx := context.Background()

	cmd := baseCreateCmd()
	dto, _ := svc.CreateExpense(ctx, cmd)

	_, err := svc.ApproveExpense(ctx, "tenant_1", dto.ID, "admin")
	if err == nil {
		t.Error("expected error when approving draft expense")
	}
}

func TestApproveExpense_EmptyTenant_Fails(t *testing.T) {
	svc, _ := newTestService()
	ctx := context.Background()

	_, err := svc.ApproveExpense(ctx, "", "exp_1", "admin")
	if err == nil {
		t.Error("expected error for empty TenantID")
	}
}

func TestRejectExpense_Success(t *testing.T) {
	svc, repo := newTestService()
	ctx := context.Background()

	cmd := baseCreateCmd()
	dto, _ := svc.CreateExpense(ctx, cmd)
	exp := repo.expenses[dto.ID]
	_ = exp.Submit()

	result, err := svc.RejectExpense(ctx, "tenant_1", dto.ID)
	if err != nil {
		t.Fatalf("RejectExpense failed: %v", err)
	}
	if result.Status != "rejected" {
		t.Errorf("expected status rejected, got %s", result.Status)
	}
}

func TestRejectExpense_FromDraft_Fails(t *testing.T) {
	svc, _ := newTestService()
	ctx := context.Background()

	cmd := baseCreateCmd()
	dto, _ := svc.CreateExpense(ctx, cmd)

	_, err := svc.RejectExpense(ctx, "tenant_1", dto.ID)
	if err == nil {
		t.Error("expected error when rejecting draft expense")
	}
}

func TestRejectExpense_EmptyTenant_Fails(t *testing.T) {
	svc, _ := newTestService()
	ctx := context.Background()

	_, err := svc.RejectExpense(ctx, "", "exp_1")
	if err == nil {
		t.Error("expected error for empty TenantID")
	}
}

// --- Repo update error ---

func TestApproveExpense_RepoUpdateError(t *testing.T) {
	svc, repo := newTestService()
	ctx := context.Background()

	cmd := baseCreateCmd()
	dto, _ := svc.CreateExpense(ctx, cmd)
	exp := repo.expenses[dto.ID]
	_ = exp.Submit()

	repo.updateFn = func(ctx context.Context, expense *domain.Expense) error {
		return fmt.Errorf("update failed")
	}

	_, err := svc.ApproveExpense(ctx, "tenant_1", dto.ID, "admin")
	if err == nil {
		t.Error("expected error when repo update fails")
	}
}

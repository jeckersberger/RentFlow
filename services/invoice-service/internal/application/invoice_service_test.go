package application

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/invoice-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/invoice-service/internal/ports"
)

// ===========================================================================
// Mock implementations
// ===========================================================================

// --- mockLogger ------------------------------------------------------------

type mockLoggerImpl struct{}

func (m *mockLoggerImpl) Debug(msg string, args ...interface{})                 {}
func (m *mockLoggerImpl) Info(msg string, args ...interface{})                  {}
func (m *mockLoggerImpl) Warn(msg string, args ...interface{})                  {}
func (m *mockLoggerImpl) Error(msg string, args ...interface{})                 {}
func (m *mockLoggerImpl) Fatal(msg string, args ...interface{})                 {}
func (m *mockLoggerImpl) WithPrefix(prefix string) logger.Logger                { return m }
func (m *mockLoggerImpl) WithCorrelationID(id string) logger.Logger             { return m }
func (m *mockLoggerImpl) WithRequestID(id string) logger.Logger                 { return m }
func (m *mockLoggerImpl) WithField(key string, value interface{}) logger.Logger { return m }

var _ logger.Logger = (*mockLoggerImpl)(nil)

// --- mockInvoiceRepository -------------------------------------------------

type mockInvoiceRepository struct {
	invoices map[string]*domain.Invoice
}

func newMockInvoiceRepo() *mockInvoiceRepository {
	return &mockInvoiceRepository{invoices: make(map[string]*domain.Invoice)}
}

func (r *mockInvoiceRepository) Create(ctx context.Context, invoice *domain.Invoice) error {
	r.invoices[invoice.ID] = invoice
	return nil
}

func (r *mockInvoiceRepository) GetByID(ctx context.Context, tenantID, invoiceID string) (*domain.Invoice, error) {
	inv, ok := r.invoices[invoiceID]
	if !ok || inv.TenantID != tenantID {
		return nil, fmt.Errorf("invoice not found")
	}
	return inv, nil
}

func (r *mockInvoiceRepository) GetByNumber(ctx context.Context, tenantID, invoiceNumber string) (*domain.Invoice, error) {
	for _, inv := range r.invoices {
		if inv.InvoiceNumber == invoiceNumber && inv.TenantID == tenantID {
			return inv, nil
		}
	}
	return nil, fmt.Errorf("invoice not found")
}

func (r *mockInvoiceRepository) List(ctx context.Context, query *ports.InvoiceListQuery) (*ports.InvoiceListResult, error) {
	var items []*domain.Invoice
	for _, inv := range r.invoices {
		if inv.TenantID == query.TenantID {
			items = append(items, inv)
		}
	}
	return &ports.InvoiceListResult{Items: items, Total: int64(len(items)), Limit: query.Limit, Offset: query.Offset}, nil
}

func (r *mockInvoiceRepository) Update(ctx context.Context, invoice *domain.Invoice) error {
	r.invoices[invoice.ID] = invoice
	return nil
}

func (r *mockInvoiceRepository) Delete(ctx context.Context, tenantID, invoiceID string) error {
	delete(r.invoices, invoiceID)
	return nil
}

func (r *mockInvoiceRepository) GetNextSequenceNumber(ctx context.Context, tenantID string) (int, error) {
	return len(r.invoices) + 1, nil
}

func (r *mockInvoiceRepository) GetOverdueInvoices(ctx context.Context, tenantID string) ([]*domain.Invoice, error) {
	return nil, nil
}

var _ ports.InvoiceRepository = (*mockInvoiceRepository)(nil)

// --- mockNumberSequenceRepository ------------------------------------------

type mockNumberSequenceRepository struct {
	counters map[string]int // key: tenantID + sequenceType
}

func newMockSeqRepo() *mockNumberSequenceRepository {
	return &mockNumberSequenceRepository{counters: make(map[string]int)}
}

func (r *mockNumberSequenceRepository) GetNextNumber(ctx context.Context, tenantID, sequenceType string) (int, error) {
	key := tenantID + ":" + sequenceType
	r.counters[key]++
	return r.counters[key], nil
}

func (r *mockNumberSequenceRepository) ResetSequence(ctx context.Context, tenantID, sequenceType string) error {
	key := tenantID + ":" + sequenceType
	r.counters[key] = 0
	return nil
}

var _ ports.NumberSequenceRepository = (*mockNumberSequenceRepository)(nil)

// ===========================================================================
// Test helpers
// ===========================================================================

const testTenantID = "tenant-test-1"

func newTestInvoiceService() (*InvoiceService, *mockInvoiceRepository, *mockNumberSequenceRepository) {
	invRepo := newMockInvoiceRepo()
	seqRepo := newMockSeqRepo()
	log := &mockLoggerImpl{}
	svc := NewInvoiceService(invRepo, seqRepo, log)
	return svc, invRepo, seqRepo
}

func validCreateCmd() CreateInvoiceCommand {
	return CreateInvoiceCommand{
		TenantID:   testTenantID,
		ClientName: "Musterfirma GmbH",
		ClientEmail: "info@musterfirma.de",
		Items: []CreateInvoiceItemCommand{
			{
				Description: "PA-Anlage Vermietung",
				Quantity:    2,
				Unit:        "Tag",
				UnitPrice:   150.00,
				TaxRate:     19,
			},
		},
		TaxRate:   19,
		Currency:  "EUR",
		IssueDate: time.Now(),
		DueDate:   time.Now().AddDate(0, 0, 14),
	}
}

// ===========================================================================
// CreateInvoice Tests (Table-Driven)
// ===========================================================================

func TestCreateInvoice_TableDriven(t *testing.T) {
	tests := []struct {
		name       string
		modifyCmd  func(*CreateInvoiceCommand)
		wantErrMsg string // empty means no error expected
	}{
		{
			name:       "valid invoice is created successfully",
			modifyCmd:  func(cmd *CreateInvoiceCommand) {},
			wantErrMsg: "",
		},
		{
			name: "missing client name returns error",
			modifyCmd: func(cmd *CreateInvoiceCommand) {
				cmd.ClientName = ""
			},
			wantErrMsg: "INVALID_INPUT",
		},
		{
			name: "no items returns error",
			modifyCmd: func(cmd *CreateInvoiceCommand) {
				cmd.Items = nil
			},
			wantErrMsg: "NO_ITEMS",
		},
		{
			name: "missing tenant ID returns error",
			modifyCmd: func(cmd *CreateInvoiceCommand) {
				cmd.TenantID = ""
			},
			wantErrMsg: "TENANT_REQUIRED",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, _, _ := newTestInvoiceService()
			ctx := context.Background()

			cmd := validCreateCmd()
			tc.modifyCmd(&cmd)

			dto, err := svc.CreateInvoice(ctx, cmd)

			if tc.wantErrMsg != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tc.wantErrMsg)
				}
				domErr, ok := err.(*domain.DomainError)
				if !ok {
					t.Fatalf("expected DomainError, got %T: %v", err, err)
				}
				if domErr.Code != tc.wantErrMsg {
					t.Errorf("expected error code %q, got %q", tc.wantErrMsg, domErr.Code)
				}
				if dto != nil {
					t.Error("expected nil DTO on error")
				}
			} else {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				if dto == nil {
					t.Fatal("expected non-nil DTO")
				}
				if dto.ClientName != cmd.ClientName {
					t.Errorf("expected client name %q, got %q", cmd.ClientName, dto.ClientName)
				}
				if dto.Status != "draft" {
					t.Errorf("expected status 'draft', got %q", dto.Status)
				}
				if dto.InvoiceNumber == "" {
					t.Error("expected non-empty invoice number")
				}
				if dto.Total <= 0 {
					t.Errorf("expected positive total, got %f", dto.Total)
				}
			}
		})
	}
}

func TestCreateInvoice_ValidData(t *testing.T) {
	svc, invRepo, _ := newTestInvoiceService()
	ctx := context.Background()

	cmd := validCreateCmd()
	dto, err := svc.CreateInvoice(ctx, cmd)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify the invoice was persisted
	if len(invRepo.invoices) != 1 {
		t.Errorf("expected 1 invoice in repo, got %d", len(invRepo.invoices))
	}

	// Verify DTO fields
	if dto.TenantID != testTenantID {
		t.Errorf("expected tenant ID %q, got %q", testTenantID, dto.TenantID)
	}
	if dto.Currency != "EUR" {
		t.Errorf("expected currency EUR, got %q", dto.Currency)
	}
	if len(dto.Items) != 1 {
		t.Errorf("expected 1 item, got %d", len(dto.Items))
	}

	// Expected: 2 * 150 = 300 net, 300 * 0.19 = 57 tax, total = 357
	if dto.SubTotal != 300.00 {
		t.Errorf("expected subtotal 300.00, got %f", dto.SubTotal)
	}
	if dto.TaxAmount != 57.00 {
		t.Errorf("expected tax amount 57.00, got %f", dto.TaxAmount)
	}
	if dto.Total != 357.00 {
		t.Errorf("expected total 357.00, got %f", dto.Total)
	}
	if dto.Hash == "" {
		t.Error("expected non-empty GoBD hash")
	}
}

func TestCreateInvoice_SequentialNumbers(t *testing.T) {
	svc, _, _ := newTestInvoiceService()
	ctx := context.Background()

	// Create two invoices and verify sequential numbering
	cmd1 := validCreateCmd()
	dto1, err := svc.CreateInvoice(ctx, cmd1)
	if err != nil {
		t.Fatalf("first invoice failed: %v", err)
	}

	cmd2 := validCreateCmd()
	cmd2.ClientName = "Andere Firma AG"
	dto2, err := svc.CreateInvoice(ctx, cmd2)
	if err != nil {
		t.Fatalf("second invoice failed: %v", err)
	}

	// Numbers should be different and sequential
	if dto1.InvoiceNumber == dto2.InvoiceNumber {
		t.Error("invoice numbers should be different")
	}

	// Both should follow the RF-YEAR-NNNN pattern
	year := time.Now().Year()
	expected1 := fmt.Sprintf("RF-%d-0001", year)
	expected2 := fmt.Sprintf("RF-%d-0002", year)
	if dto1.InvoiceNumber != expected1 {
		t.Errorf("expected first invoice number %q, got %q", expected1, dto1.InvoiceNumber)
	}
	if dto2.InvoiceNumber != expected2 {
		t.Errorf("expected second invoice number %q, got %q", expected2, dto2.InvoiceNumber)
	}
}

func TestCreateInvoice_Kleinunternehmer(t *testing.T) {
	svc, _, _ := newTestInvoiceService()
	ctx := context.Background()

	cmd := validCreateCmd()
	cmd.IsKleinunternehmer = true
	cmd.TaxRate = 0

	dto, err := svc.CreateInvoice(ctx, cmd)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if dto.TaxAmount != 0 {
		t.Errorf("expected 0 tax for Kleinunternehmer, got %f", dto.TaxAmount)
	}
	if dto.Total != dto.SubTotal {
		t.Errorf("expected total == subtotal for Kleinunternehmer, got total=%f subtotal=%f", dto.Total, dto.SubTotal)
	}
	if !dto.IsKleinunternehmer {
		t.Error("expected IsKleinunternehmer flag to be true")
	}
}

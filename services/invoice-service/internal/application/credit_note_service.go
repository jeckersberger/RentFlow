package application

import (
	"context"
	"fmt"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/invoice-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/invoice-service/internal/ports"
)

type CreditNoteService struct {
	creditNoteRepo ports.CreditNoteRepository
	invoiceRepo    ports.InvoiceRepository
	seqRepo        ports.NumberSequenceRepository
	logger         logger.Logger
}

func NewCreditNoteService(
	creditNoteRepo ports.CreditNoteRepository,
	invoiceRepo ports.InvoiceRepository,
	seqRepo ports.NumberSequenceRepository,
	logger logger.Logger,
) *CreditNoteService {
	return &CreditNoteService{
		creditNoteRepo: creditNoteRepo,
		invoiceRepo:    invoiceRepo,
		seqRepo:        seqRepo,
		logger:         logger,
	}
}

// CreateCreditNote creates a credit note for an invoice
func (s *CreditNoteService) CreateCreditNote(ctx context.Context, cmd CreateCreditNoteCommand) (*CreditNoteDTO, error) {
	if cmd.TenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}
	if cmd.InvoiceID == "" {
		return nil, domain.NewDomainError("INVALID_INPUT", "invoice ID is required", nil)
	}
	if len(cmd.Items) == 0 {
		return nil, domain.NewDomainError("NO_ITEMS", "credit note must have at least one item", nil)
	}

	// Get the original invoice
	invoice, err := s.invoiceRepo.GetByID(ctx, cmd.TenantID, cmd.InvoiceID)
	if err != nil {
		return nil, domain.NewDomainError("NOT_FOUND", "invoice not found", err)
	}

	// Get next credit note number
	seq, err := s.seqRepo.GetNextNumber(ctx, cmd.TenantID, "credit_note")
	if err != nil {
		return nil, domain.NewDomainError("SEQUENCE_ERROR", "failed to get next credit note number", err)
	}

	now := time.Now()
	creditNoteNumber := fmt.Sprintf("GS-%d-%04d", now.Year(), seq)
	creditNoteID := fmt.Sprintf("cn_%d_%04d", now.Unix(), seq)

	// Create new credit note
	creditNote := domain.NewCreditNote(
		creditNoteID,
		cmd.TenantID,
		creditNoteNumber,
		invoice.ID,
		invoice.InvoiceNumber,
		invoice.ClientName,
		invoice.ClientEmail,
	)
	creditNote.Currency = cmd.Currency
	if creditNote.Currency == "" {
		creditNote.Currency = invoice.Currency
	}
	creditNote.Reason = cmd.Reason

	// Set tax rate
	taxRate := cmd.TaxRate
	if taxRate == 0 {
		taxRate = float64(invoice.TaxRate)
	}
	if err := creditNote.SetTaxRate(taxRate); err != nil {
		return nil, domain.NewDomainError("INVALID_TAX_RATE", err.Error(), err)
	}

	// Add items
	for _, item := range cmd.Items {
		creditNoteItem := domain.CreditNoteItem{
			Description: item.Description,
			Quantity:    item.Quantity,
			Unit:        item.Unit,
			UnitPrice:   item.UnitPrice,
		}
		if err := creditNote.AddItem(creditNoteItem); err != nil {
			return nil, domain.NewDomainError("INVALID_ITEM", err.Error(), err)
		}
	}

	// Calculate totals
	if err := creditNote.CalculateTotals(); err != nil {
		return nil, domain.NewDomainError("CALCULATION_ERROR", err.Error(), err)
	}

	// Validate
	if err := creditNote.Validate(); err != nil {
		return nil, domain.NewDomainError("VALIDATION_ERROR", err.Error(), err)
	}

	// Persist credit note
	if err := s.creditNoteRepo.Create(ctx, creditNote); err != nil {
		return nil, domain.NewDomainError("CREATE_ERROR", "failed to create credit note", err)
	}

	// Mark original invoice as credited
	if err := invoice.Credit(creditNoteID); err != nil {
		return nil, domain.NewDomainError("INVALID_TRANSITION", err.Error(), err)
	}

	if err := s.invoiceRepo.Update(ctx, invoice); err != nil {
		s.logger.Error("Failed to update invoice after credit note creation", err)
		// Don't fail - credit note was created successfully
	}

	s.logger.Info("Credit note created", "id", creditNote.ID, "number", creditNoteNumber, "tenant_id", cmd.TenantID, "invoice_id", cmd.InvoiceID)
	return CreditNoteToDTO(creditNote), nil
}

// GetCreditNote retrieves a credit note by ID
func (s *CreditNoteService) GetCreditNote(ctx context.Context, tenantID, creditNoteID string) (*CreditNoteDTO, error) {
	if tenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	creditNote, err := s.creditNoteRepo.GetByID(ctx, tenantID, creditNoteID)
	if err != nil {
		return nil, domain.NewDomainError("NOT_FOUND", "credit note not found", err)
	}

	return CreditNoteToDTO(creditNote), nil
}

// IssueCreditNote transitions a credit note to Issued
func (s *CreditNoteService) IssueCreditNote(ctx context.Context, tenantID, creditNoteID string) (*CreditNoteDTO, error) {
	if tenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	creditNote, err := s.creditNoteRepo.GetByID(ctx, tenantID, creditNoteID)
	if err != nil {
		return nil, domain.NewDomainError("NOT_FOUND", "credit note not found", err)
	}

	if err := creditNote.Issue(); err != nil {
		return nil, domain.NewDomainError("INVALID_TRANSITION", err.Error(), err)
	}

	if err := s.creditNoteRepo.Update(ctx, creditNote); err != nil {
		return nil, domain.NewDomainError("UPDATE_ERROR", "failed to issue credit note", err)
	}

	s.logger.Info("Credit note issued", "id", creditNoteID, "number", creditNote.CreditNoteNumber)
	return CreditNoteToDTO(creditNote), nil
}

package application

import (
	"context"
	"fmt"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/invoice-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/invoice-service/internal/ports"
)

type InvoiceService struct {
	invoiceRepo ports.InvoiceRepository
	seqRepo     ports.NumberSequenceRepository
	logger      *logger.Logger
}

func NewInvoiceService(
	invoiceRepo ports.InvoiceRepository,
	seqRepo ports.NumberSequenceRepository,
	logger *logger.Logger,
) *InvoiceService {
	return &InvoiceService{
		invoiceRepo: invoiceRepo,
		seqRepo:     seqRepo,
		logger:      logger,
	}
}

// CreateInvoice creates a new invoice
func (s *InvoiceService) CreateInvoice(ctx context.Context, cmd CreateInvoiceCommand) (*InvoiceDTO, error) {
	if cmd.TenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}
	if cmd.ClientName == "" {
		return nil, domain.NewDomainError("INVALID_INPUT", "client name is required", nil)
	}
	if len(cmd.Items) == 0 {
		return nil, domain.NewDomainError("NO_ITEMS", "invoice must have at least one item", nil)
	}

	// Get next invoice number (sequential, no gaps - GoBD requirement)
	seq, err := s.seqRepo.GetNextNumber(ctx, cmd.TenantID, "invoice")
	if err != nil {
		return nil, domain.NewDomainError("SEQUENCE_ERROR", "failed to get next invoice number", err)
	}

	now := time.Now()
	invoiceNumber := fmt.Sprintf("RF-%d-%04d", now.Year(), seq)
	invoiceID := fmt.Sprintf("inv_%d_%04d", now.Unix(), seq)

	// Create new invoice aggregate
	invoice := domain.NewInvoice(invoiceID, cmd.TenantID, invoiceNumber, cmd.ClientName, cmd.ClientEmail)
	invoice.ProjectID = cmd.ProjectID
	invoice.ClientAddress = cmd.ClientAddress
	invoice.ClientTaxID = cmd.ClientTaxID
	invoice.Currency = cmd.Currency
	invoice.IssueDate = cmd.IssueDate
	invoice.DueDate = cmd.DueDate
	invoice.Notes = cmd.Notes
	invoice.InternalNotes = cmd.InternalNotes

	// Set tax rate (validate 0, 7, or 19%)
	if err := invoice.SetTaxRate(cmd.TaxRate); err != nil {
		return nil, domain.NewDomainError("INVALID_TAX_RATE", err.Error(), err)
	}

	// Add items
	for _, item := range cmd.Items {
		invoiceItem := domain.InvoiceItem{
			Description: item.Description,
			Quantity:    item.Quantity,
			Unit:        item.Unit,
			UnitPrice:   item.UnitPrice,
			EquipmentID: item.EquipmentID,
		}
		if err := invoice.AddItem(invoiceItem); err != nil {
			return nil, domain.NewDomainError("INVALID_ITEM", err.Error(), err)
		}
	}

	// Calculate totals
	if err := invoice.CalculateTotals(); err != nil {
		return nil, domain.NewDomainError("CALCULATION_ERROR", err.Error(), err)
	}

	// Compute GoBD hash
	invoice.Hash = invoice.ComputeHash()

	// Validate before persistence
	if err := invoice.Validate(); err != nil {
		return nil, domain.NewDomainError("VALIDATION_ERROR", err.Error(), err)
	}

	// Persist
	if err := s.invoiceRepo.Create(ctx, invoice); err != nil {
		return nil, domain.NewDomainError("CREATE_ERROR", "failed to create invoice", err)
	}

	s.logger.Info("Invoice created", "id", invoice.ID, "number", invoiceNumber, "tenant_id", cmd.TenantID)
	return InvoiceToDTO(invoice), nil
}

// GetInvoice retrieves an invoice by ID
func (s *InvoiceService) GetInvoice(ctx context.Context, tenantID, invoiceID string) (*InvoiceDTO, error) {
	if tenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	invoice, err := s.invoiceRepo.GetByID(ctx, tenantID, invoiceID)
	if err != nil {
		return nil, domain.NewDomainError("NOT_FOUND", "invoice not found", err)
	}

	return InvoiceToDTO(invoice), nil
}

// GetInvoiceByNumber retrieves an invoice by number
func (s *InvoiceService) GetInvoiceByNumber(ctx context.Context, tenantID, invoiceNumber string) (*InvoiceDTO, error) {
	if tenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	invoice, err := s.invoiceRepo.GetByNumber(ctx, tenantID, invoiceNumber)
	if err != nil {
		return nil, domain.NewDomainError("NOT_FOUND", "invoice not found", err)
	}

	return InvoiceToDTO(invoice), nil
}

// ListInvoices lists invoices with pagination and filtering
func (s *InvoiceService) ListInvoices(ctx context.Context, query ListInvoicesQuery) (*PaginatedResult, error) {
	if query.TenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	var status *domain.InvoiceStatus
	if query.Status != nil {
		s := domain.InvoiceStatus(*query.Status)
		status = &s
	}

	repoQuery := &ports.InvoiceListQuery{
		TenantID:   query.TenantID,
		Status:     status,
		ClientName: query.ClientName,
		FromDate:   query.FromDate,
		ToDate:     query.ToDate,
		Limit:      query.Limit,
		Offset:     query.Offset,
	}

	result, err := s.invoiceRepo.List(ctx, repoQuery)
	if err != nil {
		return nil, domain.NewDomainError("QUERY_ERROR", "failed to list invoices", err)
	}

	dtos := make([]interface{}, len(result.Items))
	for i, invoice := range result.Items {
		dtos[i] = InvoiceToDTO(invoice)
	}

	return &PaginatedResult{
		Data:   dtos,
		Total:  result.Total,
		Limit:  result.Limit,
		Offset: result.Offset,
	}, nil
}

// SendInvoice transitions invoice to Sent
func (s *InvoiceService) SendInvoice(ctx context.Context, cmd SendInvoiceCommand) (*InvoiceDTO, error) {
	if cmd.TenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	invoice, err := s.invoiceRepo.GetByID(ctx, cmd.TenantID, cmd.ID)
	if err != nil {
		return nil, domain.NewDomainError("NOT_FOUND", "invoice not found", err)
	}

	// Verify hash for finalized invoices (GoBD compliance)
	if invoice.IsFinalized() && !invoice.VerifyHash() {
		return nil, domain.NewDomainError("INTEGRITY_CHECK", "invoice hash mismatch - possible tampering detected", nil)
	}

	if err := invoice.Send(cmd.Email); err != nil {
		return nil, domain.NewDomainError("INVALID_TRANSITION", err.Error(), err)
	}

	if err := s.invoiceRepo.Update(ctx, invoice); err != nil {
		return nil, domain.NewDomainError("UPDATE_ERROR", "failed to send invoice", err)
	}

	s.logger.Info("Invoice sent", "id", cmd.ID, "number", invoice.InvoiceNumber, "email", cmd.Email)
	return InvoiceToDTO(invoice), nil
}

// MarkInvoicePaid transitions invoice to Paid
func (s *InvoiceService) MarkInvoicePaid(ctx context.Context, cmd MarkInvoicePaidCommand) (*InvoiceDTO, error) {
	if cmd.TenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	invoice, err := s.invoiceRepo.GetByID(ctx, cmd.TenantID, cmd.ID)
	if err != nil {
		return nil, domain.NewDomainError("NOT_FOUND", "invoice not found", err)
	}

	// Verify hash for finalized invoices
	if invoice.IsFinalized() && !invoice.VerifyHash() {
		return nil, domain.NewDomainError("INTEGRITY_CHECK", "invoice hash mismatch", nil)
	}

	if err := invoice.MarkPaid(cmd.PaymentMethod, cmd.PaymentRef); err != nil {
		return nil, domain.NewDomainError("INVALID_TRANSITION", err.Error(), err)
	}

	if err := s.invoiceRepo.Update(ctx, invoice); err != nil {
		return nil, domain.NewDomainError("UPDATE_ERROR", "failed to mark invoice paid", err)
	}

	s.logger.Info("Invoice marked paid", "id", cmd.ID, "number", invoice.InvoiceNumber)
	return InvoiceToDTO(invoice), nil
}

// CancelInvoice transitions invoice to Cancelled
func (s *InvoiceService) CancelInvoice(ctx context.Context, cmd CancelInvoiceCommand) (*InvoiceDTO, error) {
	if cmd.TenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	invoice, err := s.invoiceRepo.GetByID(ctx, cmd.TenantID, cmd.ID)
	if err != nil {
		return nil, domain.NewDomainError("NOT_FOUND", "invoice not found", err)
	}

	// Verify hash
	if invoice.IsFinalized() && !invoice.VerifyHash() {
		return nil, domain.NewDomainError("INTEGRITY_CHECK", "invoice hash mismatch", nil)
	}

	if err := invoice.Cancel(cmd.Reason); err != nil {
		return nil, domain.NewDomainError("INVALID_TRANSITION", err.Error(), err)
	}

	if err := s.invoiceRepo.Update(ctx, invoice); err != nil {
		return nil, domain.NewDomainError("UPDATE_ERROR", "failed to cancel invoice", err)
	}

	s.logger.Info("Invoice cancelled", "id", cmd.ID, "number", invoice.InvoiceNumber, "reason", cmd.Reason)
	return InvoiceToDTO(invoice), nil
}

// CreditInvoice creates a credit note
func (s *InvoiceService) CreditInvoice(ctx context.Context, cmd CreditInvoiceCommand) (*InvoiceDTO, error) {
	if cmd.TenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	invoice, err := s.invoiceRepo.GetByID(ctx, cmd.TenantID, cmd.ID)
	if err != nil {
		return nil, domain.NewDomainError("NOT_FOUND", "invoice not found", err)
	}

	// Verify hash
	if invoice.IsFinalized() && !invoice.VerifyHash() {
		return nil, domain.NewDomainError("INTEGRITY_CHECK", "invoice hash mismatch", nil)
	}

	if err := invoice.Credit(cmd.CreditID); err != nil {
		return nil, domain.NewDomainError("INVALID_TRANSITION", err.Error(), err)
	}

	if err := s.invoiceRepo.Update(ctx, invoice); err != nil {
		return nil, domain.NewDomainError("UPDATE_ERROR", "failed to credit invoice", err)
	}

	s.logger.Info("Invoice credited", "id", cmd.ID, "number", invoice.InvoiceNumber, "credit_id", cmd.CreditID)
	return InvoiceToDTO(invoice), nil
}

// AddItem adds a line item to an invoice
func (s *InvoiceService) AddItem(ctx context.Context, cmd AddInvoiceItemCommand) (*InvoiceDTO, error) {
	if cmd.TenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	invoice, err := s.invoiceRepo.GetByID(ctx, cmd.TenantID, cmd.InvoiceID)
	if err != nil {
		return nil, domain.NewDomainError("NOT_FOUND", "invoice not found", err)
	}

	// Can only add items to draft invoices
	if invoice.IsFinalized() {
		return nil, domain.NewDomainError("CANNOT_MODIFY", "cannot modify finalized invoice", nil)
	}

	item := domain.InvoiceItem{
		Description: cmd.Description,
		Quantity:    cmd.Quantity,
		Unit:        cmd.Unit,
		UnitPrice:   cmd.UnitPrice,
		EquipmentID: cmd.EquipmentID,
	}

	if err := invoice.AddItem(item); err != nil {
		return nil, domain.NewDomainError("INVALID_ITEM", err.Error(), err)
	}

	// Recalculate totals
	if err := invoice.CalculateTotals(); err != nil {
		return nil, domain.NewDomainError("CALCULATION_ERROR", err.Error(), err)
	}

	if err := s.invoiceRepo.Update(ctx, invoice); err != nil {
		return nil, domain.NewDomainError("UPDATE_ERROR", "failed to add item", err)
	}

	s.logger.Info("Item added to invoice", "invoice_id", cmd.InvoiceID, "item_id", len(invoice.Items))
	return InvoiceToDTO(invoice), nil
}

// RemoveItem removes a line item from an invoice
func (s *InvoiceService) RemoveItem(ctx context.Context, cmd RemoveInvoiceItemCommand) (*InvoiceDTO, error) {
	if cmd.TenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	invoice, err := s.invoiceRepo.GetByID(ctx, cmd.TenantID, cmd.InvoiceID)
	if err != nil {
		return nil, domain.NewDomainError("NOT_FOUND", "invoice not found", err)
	}

	if invoice.IsFinalized() {
		return nil, domain.NewDomainError("CANNOT_MODIFY", "cannot modify finalized invoice", nil)
	}

	if err := invoice.RemoveItem(cmd.ItemID); err != nil {
		return nil, domain.NewDomainError("INVALID_ITEM", err.Error(), err)
	}

	// Recalculate totals
	if err := invoice.CalculateTotals(); err != nil {
		return nil, domain.NewDomainError("CALCULATION_ERROR", err.Error(), err)
	}

	if err := s.invoiceRepo.Update(ctx, invoice); err != nil {
		return nil, domain.NewDomainError("UPDATE_ERROR", "failed to remove item", err)
	}

	s.logger.Info("Item removed from invoice", "invoice_id", cmd.InvoiceID, "item_id", cmd.ItemID)
	return InvoiceToDTO(invoice), nil
}

// GetOverdueInvoices returns all overdue invoices for a tenant
func (s *InvoiceService) GetOverdueInvoices(ctx context.Context, tenantID string) ([]*InvoiceDTO, error) {
	if tenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	invoices, err := s.invoiceRepo.GetOverdueInvoices(ctx, tenantID)
	if err != nil {
		return nil, domain.NewDomainError("QUERY_ERROR", "failed to get overdue invoices", err)
	}

	dtos := make([]*InvoiceDTO, len(invoices))
	for i, invoice := range invoices {
		dtos[i] = InvoiceToDTO(invoice)
	}

	return dtos, nil
}

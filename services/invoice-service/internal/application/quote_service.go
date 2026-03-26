package application

import (
	"context"
	"fmt"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/events"
	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/invoice-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/invoice-service/internal/ports"
)

type QuoteService struct {
	quoteRepo    ports.QuoteRepository
	invoiceRepo  ports.InvoiceRepository
	seqRepo      ports.NumberSequenceRepository
	pdfGenerator ports.PDFGenerator
	logger       logger.Logger
}

// SetPDFGenerator setzt den PDF-Generator für Angebots-PDFs
func (s *QuoteService) SetPDFGenerator(gen ports.PDFGenerator) {
	s.pdfGenerator = gen
}

func NewQuoteService(
	quoteRepo ports.QuoteRepository,
	invoiceRepo ports.InvoiceRepository,
	seqRepo ports.NumberSequenceRepository,
	logger logger.Logger,
) *QuoteService {
	return &QuoteService{
		quoteRepo:   quoteRepo,
		invoiceRepo: invoiceRepo,
		seqRepo:     seqRepo,
		logger:      logger,
	}
}

// CreateQuote creates a new quote
func (s *QuoteService) CreateQuote(ctx context.Context, cmd CreateQuoteCommand) (*QuoteDTO, error) {
	if cmd.TenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}
	if cmd.ClientName == "" {
		return nil, domain.NewDomainError("INVALID_INPUT", "client name is required", nil)
	}
	if len(cmd.Items) == 0 {
		return nil, domain.NewDomainError("NO_ITEMS", "quote must have at least one item", nil)
	}

	// Get next quote number
	seq, err := s.seqRepo.GetNextNumber(ctx, cmd.TenantID, "quote")
	if err != nil {
		return nil, domain.NewDomainError("SEQUENCE_ERROR", "failed to get next quote number", err)
	}

	now := time.Now()
	quoteNumber := fmt.Sprintf("AN-%d-%04d", now.Year(), seq)
	quoteID := fmt.Sprintf("quote_%d_%04d", now.Unix(), seq)

	validDays := cmd.ValidDays
	if validDays == 0 {
		validDays = 30 // Default 30 days
	}

	// Create new quote
	quote := domain.NewQuote(quoteID, cmd.TenantID, quoteNumber, cmd.ClientName, cmd.ClientEmail, validDays)
	quote.ProjectID = cmd.ProjectID
	quote.ClientAddress = cmd.ClientAddress
	quote.Currency = cmd.Currency
	quote.Notes = cmd.Notes

	// Set tax rate
	if err := quote.SetTaxRate(cmd.TaxRate); err != nil {
		return nil, domain.NewDomainError("INVALID_TAX_RATE", err.Error(), err)
	}

	// Add items
	for _, item := range cmd.Items {
		quoteItem := domain.InvoiceItem{
			Description: item.Description,
			Quantity:    item.Quantity,
			Unit:        item.Unit,
			UnitPrice:   item.UnitPrice,
			EquipmentID: item.EquipmentID,
		}
		if err := quote.AddItem(quoteItem); err != nil {
			return nil, domain.NewDomainError("INVALID_ITEM", err.Error(), err)
		}
	}

	// Calculate totals
	if err := quote.CalculateTotals(); err != nil {
		return nil, domain.NewDomainError("CALCULATION_ERROR", err.Error(), err)
	}

	// Validate
	if err := quote.Validate(); err != nil {
		return nil, domain.NewDomainError("VALIDATION_ERROR", err.Error(), err)
	}

	// Persist
	if err := s.quoteRepo.Create(ctx, quote); err != nil {
		return nil, domain.NewDomainError("CREATE_ERROR", "failed to create quote", err)
	}

	s.logger.Info("Quote created", "id", quote.ID, "number", quoteNumber, "tenant_id", cmd.TenantID)
	return QuoteToDTO(quote), nil
}

// GetQuote retrieves a quote by ID
func (s *QuoteService) GetQuote(ctx context.Context, tenantID, quoteID string) (*QuoteDTO, error) {
	if tenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	quote, err := s.quoteRepo.GetByID(ctx, tenantID, quoteID)
	if err != nil {
		return nil, domain.NewDomainError("NOT_FOUND", "quote not found", err)
	}

	return QuoteToDTO(quote), nil
}

// GetQuoteByNumber retrieves a quote by number
func (s *QuoteService) GetQuoteByNumber(ctx context.Context, tenantID, quoteNumber string) (*QuoteDTO, error) {
	if tenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	quote, err := s.quoteRepo.GetByNumber(ctx, tenantID, quoteNumber)
	if err != nil {
		return nil, domain.NewDomainError("NOT_FOUND", "quote not found", err)
	}

	return QuoteToDTO(quote), nil
}

// ListQuotes lists quotes with pagination and filtering
func (s *QuoteService) ListQuotes(ctx context.Context, query ListQuotesQuery) (*PaginatedResult, error) {
	if query.TenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	var status *domain.QuoteStatus
	if query.Status != nil {
		s := domain.QuoteStatus(*query.Status)
		status = &s
	}

	repoQuery := &ports.QuoteListQuery{
		TenantID:   query.TenantID,
		Status:     status,
		ClientName: query.ClientName,
		FromDate:   query.FromDate,
		ToDate:     query.ToDate,
		Limit:      query.Limit,
		Offset:     query.Offset,
	}

	result, err := s.quoteRepo.List(ctx, repoQuery)
	if err != nil {
		return nil, domain.NewDomainError("QUERY_ERROR", "failed to list quotes", err)
	}

	dtos := make([]interface{}, len(result.Items))
	for i, quote := range result.Items {
		dtos[i] = QuoteToDTO(quote)
	}

	return &PaginatedResult{
		Data:   dtos,
		Total:  result.Total,
		Limit:  result.Limit,
		Offset: result.Offset,
	}, nil
}

// SendQuote sends a quote to the client
func (s *QuoteService) SendQuote(ctx context.Context, cmd SendQuoteCommand) (*QuoteDTO, error) {
	if cmd.TenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	quote, err := s.quoteRepo.GetByID(ctx, cmd.TenantID, cmd.ID)
	if err != nil {
		return nil, domain.NewDomainError("NOT_FOUND", "quote not found", err)
	}

	if err := quote.Send(cmd.Email); err != nil {
		return nil, domain.NewDomainError("INVALID_TRANSITION", err.Error(), err)
	}

	if err := s.quoteRepo.Update(ctx, quote); err != nil {
		return nil, domain.NewDomainError("UPDATE_ERROR", "failed to send quote", err)
	}

	s.logger.Info("Quote sent", "id", cmd.ID, "number", quote.QuoteNumber, "email", cmd.Email)
	return QuoteToDTO(quote), nil
}

// AcceptQuote accepts a quote
func (s *QuoteService) AcceptQuote(ctx context.Context, cmd AcceptQuoteCommand) (*QuoteDTO, error) {
	if cmd.TenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	quote, err := s.quoteRepo.GetByID(ctx, cmd.TenantID, cmd.ID)
	if err != nil {
		return nil, domain.NewDomainError("NOT_FOUND", "quote not found", err)
	}

	if err := quote.Accept(); err != nil {
		return nil, domain.NewDomainError("INVALID_TRANSITION", err.Error(), err)
	}

	if err := s.quoteRepo.Update(ctx, quote); err != nil {
		return nil, domain.NewDomainError("UPDATE_ERROR", "failed to accept quote", err)
	}

	s.logger.Info("Quote accepted", "id", cmd.ID, "number", quote.QuoteNumber)
	return QuoteToDTO(quote), nil
}

// ConfirmQuote confirms a quote (Auftragsbestätigung)
func (s *QuoteService) ConfirmQuote(ctx context.Context, cmd ConfirmQuoteCommand) (*QuoteDTO, error) {
	if cmd.TenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	quote, err := s.quoteRepo.GetByID(ctx, cmd.TenantID, cmd.ID)
	if err != nil {
		return nil, domain.NewDomainError("NOT_FOUND", "quote not found", err)
	}

	if err := quote.Confirm(); err != nil {
		return nil, domain.NewDomainError("INVALID_TRANSITION", err.Error(), err)
	}

	if err := s.quoteRepo.Update(ctx, quote); err != nil {
		return nil, domain.NewDomainError("UPDATE_ERROR", "failed to confirm quote", err)
	}

	s.logger.Info("Quote confirmed", "id", cmd.ID, "number", quote.QuoteNumber)
	return QuoteToDTO(quote), nil
}

// RejectQuote rejects a quote
func (s *QuoteService) RejectQuote(ctx context.Context, cmd RejectQuoteCommand) (*QuoteDTO, error) {
	if cmd.TenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	quote, err := s.quoteRepo.GetByID(ctx, cmd.TenantID, cmd.ID)
	if err != nil {
		return nil, domain.NewDomainError("NOT_FOUND", "quote not found", err)
	}

	if err := quote.Reject(cmd.Reason); err != nil {
		return nil, domain.NewDomainError("INVALID_TRANSITION", err.Error(), err)
	}

	if err := s.quoteRepo.Update(ctx, quote); err != nil {
		return nil, domain.NewDomainError("UPDATE_ERROR", "failed to reject quote", err)
	}

	s.logger.Info("Quote rejected", "id", cmd.ID, "number", quote.QuoteNumber, "reason", cmd.Reason)
	return QuoteToDTO(quote), nil
}

// ConvertQuoteToInvoice converts an accepted or confirmed quote to an invoice
func (s *QuoteService) ConvertQuoteToInvoice(ctx context.Context, cmd ConvertQuoteToInvoiceCommand) (*InvoiceDTO, error) {
	if cmd.TenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	quote, err := s.quoteRepo.GetByID(ctx, cmd.TenantID, cmd.QuoteID)
	if err != nil {
		return nil, domain.NewDomainError("NOT_FOUND", "quote not found", err)
	}

	if quote.Status != domain.QuoteAccepted && quote.Status != domain.QuoteConfirmed {
		return nil, domain.NewDomainError("INVALID_STATUS", "only accepted or confirmed quotes can be converted to invoices", nil)
	}

	// Get next invoice number
	seq, err := s.seqRepo.GetNextNumber(ctx, cmd.TenantID, "invoice")
	if err != nil {
		return nil, domain.NewDomainError("SEQUENCE_ERROR", "failed to get next invoice number", err)
	}

	now := time.Now()
	invoiceNumber := fmt.Sprintf("EF-%d-%04d", now.Year(), seq)
	invoiceID := fmt.Sprintf("inv_%d_%04d", now.Unix(), seq)

	// Create invoice from quote
	invoice := domain.NewInvoice(invoiceID, cmd.TenantID, invoiceNumber, quote.ClientName, quote.ClientEmail)
	invoice.ProjectID = quote.ProjectID
	invoice.ClientAddress = quote.ClientAddress
	invoice.Currency = quote.Currency
	invoice.IssueDate = cmd.IssueDate
	invoice.DueDate = cmd.DueDate
	invoice.Notes = quote.Notes
	invoice.InternalNotes = cmd.InternalNotes

	// Set tax rate
	if err := invoice.SetTaxRate(float64(quote.TaxRate)); err != nil {
		return nil, domain.NewDomainError("INVALID_TAX_RATE", err.Error(), err)
	}

	// Copy items from quote
	for _, quoteItem := range quote.Items {
		invoiceItem := domain.InvoiceItem{
			Description: quoteItem.Description,
			Quantity:    quoteItem.Quantity,
			Unit:        quoteItem.Unit,
			UnitPrice:   quoteItem.UnitPrice,
			EquipmentID: quoteItem.EquipmentID,
		}
		if err := invoice.AddItem(invoiceItem); err != nil {
			return nil, domain.NewDomainError("INVALID_ITEM", err.Error(), err)
		}
	}

	// Calculate totals
	if err := invoice.CalculateTotals(); err != nil {
		return nil, domain.NewDomainError("CALCULATION_ERROR", err.Error(), err)
	}

	// Compute hash
	invoice.Hash = invoice.ComputeHash()

	// Validate
	if err := invoice.Validate(); err != nil {
		return nil, domain.NewDomainError("VALIDATION_ERROR", err.Error(), err)
	}

	// Persist invoice
	if err := s.invoiceRepo.Create(ctx, invoice); err != nil {
		return nil, domain.NewDomainError("CREATE_ERROR", "failed to create invoice from quote", err)
	}

	// Record conversion event on quote
	convEvent := domain.QuoteConvertedToInvoiceEvent{
		TenantID:      quote.TenantID,
		QuoteNumber:   quote.QuoteNumber,
		InvoiceNumber: invoiceNumber,
		ConvertedAt:   now,
	}
	eventData, _ := events.NewEventData("QuoteConvertedToInvoice", convEvent, nil)
	quote.Apply(*eventData)
	quote.UpdatedAt = now

	if err := s.quoteRepo.Update(ctx, quote); err != nil {
		// Log but don't fail the invoice creation
		s.logger.Error("Failed to update quote after conversion", err)
	}

	s.logger.Info("Quote converted to invoice", "quote_id", cmd.QuoteID, "invoice_id", invoiceID)
	return InvoiceToDTO(invoice), nil
}

// AddItem adds a line item to a quote
func (s *QuoteService) AddItem(ctx context.Context, cmd AddQuoteItemCommand) (*QuoteDTO, error) {
	if cmd.TenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	quote, err := s.quoteRepo.GetByID(ctx, cmd.TenantID, cmd.QuoteID)
	if err != nil {
		return nil, domain.NewDomainError("NOT_FOUND", "quote not found", err)
	}

	// Can only add items to draft quotes
	if quote.Status != domain.QuoteDraft {
		return nil, domain.NewDomainError("CANNOT_MODIFY", "cannot modify quote in current status", nil)
	}

	item := domain.InvoiceItem{
		Description: cmd.Description,
		Quantity:    cmd.Quantity,
		Unit:        cmd.Unit,
		UnitPrice:   cmd.UnitPrice,
		EquipmentID: cmd.EquipmentID,
	}

	if err := quote.AddItem(item); err != nil {
		return nil, domain.NewDomainError("INVALID_ITEM", err.Error(), err)
	}

	// Recalculate totals
	if err := quote.CalculateTotals(); err != nil {
		return nil, domain.NewDomainError("CALCULATION_ERROR", err.Error(), err)
	}

	if err := s.quoteRepo.Update(ctx, quote); err != nil {
		return nil, domain.NewDomainError("UPDATE_ERROR", "failed to add item", err)
	}

	s.logger.Info("Item added to quote", "quote_id", cmd.QuoteID)
	return QuoteToDTO(quote), nil
}

// RemoveItem removes a line item from a quote
func (s *QuoteService) RemoveItem(ctx context.Context, cmd RemoveQuoteItemCommand) (*QuoteDTO, error) {
	if cmd.TenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	quote, err := s.quoteRepo.GetByID(ctx, cmd.TenantID, cmd.QuoteID)
	if err != nil {
		return nil, domain.NewDomainError("NOT_FOUND", "quote not found", err)
	}

	if quote.Status != domain.QuoteDraft {
		return nil, domain.NewDomainError("CANNOT_MODIFY", "cannot modify quote in current status", nil)
	}

	if err := quote.RemoveItem(cmd.ItemID); err != nil {
		return nil, domain.NewDomainError("INVALID_ITEM", err.Error(), err)
	}

	// Recalculate totals
	if err := quote.CalculateTotals(); err != nil {
		return nil, domain.NewDomainError("CALCULATION_ERROR", err.Error(), err)
	}

	if err := s.quoteRepo.Update(ctx, quote); err != nil {
		return nil, domain.NewDomainError("UPDATE_ERROR", "failed to remove item", err)
	}

	s.logger.Info("Item removed from quote", "quote_id", cmd.QuoteID)
	return QuoteToDTO(quote), nil
}

// GenerateQuoteHTML generiert druckbares Angebots-HTML
func (s *QuoteService) GenerateQuoteHTML(ctx context.Context, tenantID, quoteID string) (string, error) {
	quote, err := s.quoteRepo.GetByID(ctx, tenantID, quoteID)
	if err != nil {
		return "", domain.NewDomainError("NOT_FOUND", "quote not found", err)
	}

	html := buildQuoteHTML(quote)
	return html, nil
}

// GenerateQuotePDF generiert ein PDF-Byte-Array des Angebots mittels chromedp
func (s *QuoteService) GenerateQuotePDF(ctx context.Context, tenantID, quoteID string) ([]byte, error) {
	quote, err := s.quoteRepo.GetByID(ctx, tenantID, quoteID)
	if err != nil {
		return nil, domain.NewDomainError("NOT_FOUND", "quote not found", err)
	}

	html := buildQuoteHTML(quote)

	if s.pdfGenerator != nil {
		pdfBytes, err := s.pdfGenerator.GeneratePDF(ctx, html)
		if err != nil {
			s.logger.Error("Quote PDF generation failed, falling back to HTML", err)
			return []byte(html), nil
		}
		return pdfBytes, nil
	}

	return []byte(html), nil
}

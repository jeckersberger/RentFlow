package application

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/invoice-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/invoice-service/internal/ports"
)

type InvoiceService struct {
	invoiceRepo  ports.InvoiceRepository
	seqRepo      ports.NumberSequenceRepository
	pdfGenerator ports.PDFGenerator
	emailSender  ports.EmailSender
	logger       logger.Logger
}

func NewInvoiceService(
	invoiceRepo ports.InvoiceRepository,
	seqRepo ports.NumberSequenceRepository,
	logger logger.Logger,
) *InvoiceService {
	return &InvoiceService{
		invoiceRepo: invoiceRepo,
		seqRepo:     seqRepo,
		logger:      logger,
	}
}

// SetPDFGenerator setzt den PDF-Generator (optional, für chromedp-basierte PDF-Erzeugung)
func (s *InvoiceService) SetPDFGenerator(gen ports.PDFGenerator) {
	s.pdfGenerator = gen
}

// SetEmailSender setzt den Email-Sender (optional, für SMTP-basierten E-Mail-Versand)
func (s *InvoiceService) SetEmailSender(sender ports.EmailSender) {
	s.emailSender = sender
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

	// Kleinunternehmerregelung (§19 UStG)
	invoice.IsKleinunternehmer = cmd.IsKleinunternehmer

	// Add items
	for _, item := range cmd.Items {
		invoiceItem := domain.InvoiceItem{
			Description: item.Description,
			Quantity:    item.Quantity,
			Unit:        item.Unit,
			UnitPrice:   item.UnitPrice,
			TaxRate:     domain.TaxRate(item.TaxRate),
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

// UpdateInvoice updates a draft invoice
func (s *InvoiceService) UpdateInvoice(ctx context.Context, cmd UpdateInvoiceCommand) (*InvoiceDTO, error) {
	if cmd.TenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	invoice, err := s.invoiceRepo.GetByID(ctx, cmd.TenantID, cmd.ID)
	if err != nil {
		return nil, domain.NewDomainError("NOT_FOUND", "invoice not found", err)
	}

	if invoice.IsFinalized() {
		return nil, domain.NewDomainError("CANNOT_MODIFY", "cannot modify finalized invoice", nil)
	}

	if cmd.ClientName != "" {
		invoice.ClientName = cmd.ClientName
	}
	if cmd.ClientEmail != "" {
		invoice.ClientEmail = cmd.ClientEmail
	}
	if cmd.ClientTaxID != "" {
		invoice.ClientTaxID = cmd.ClientTaxID
	}
	invoice.ClientAddress = cmd.ClientAddress
	if cmd.TaxRate > 0 {
		if err := invoice.SetTaxRate(cmd.TaxRate); err != nil {
			return nil, domain.NewDomainError("INVALID_TAX_RATE", err.Error(), err)
		}
	}
	if !cmd.DueDate.IsZero() {
		invoice.DueDate = cmd.DueDate
	}
	invoice.Notes = cmd.Notes
	invoice.InternalNotes = cmd.InternalNotes

	if err := invoice.CalculateTotals(); err != nil {
		return nil, domain.NewDomainError("CALCULATION_ERROR", err.Error(), err)
	}

	invoice.Hash = invoice.ComputeHash()
	invoice.UpdatedAt = time.Now()

	if err := s.invoiceRepo.Update(ctx, invoice); err != nil {
		return nil, domain.NewDomainError("UPDATE_ERROR", "failed to update invoice", err)
	}

	s.logger.Info("Invoice updated", "id", cmd.ID, "tenant_id", cmd.TenantID)
	return InvoiceToDTO(invoice), nil
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

// RecordPayment records a partial payment on an invoice
func (s *InvoiceService) RecordPayment(ctx context.Context, cmd RecordPaymentCommand) (*InvoiceDTO, error) {
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

	if err := invoice.RecordPayment(cmd.Amount); err != nil {
		return nil, domain.NewDomainError("INVALID_PAYMENT", err.Error(), err)
	}

	if err := s.invoiceRepo.Update(ctx, invoice); err != nil {
		return nil, domain.NewDomainError("UPDATE_ERROR", "failed to record payment", err)
	}

	s.logger.Info("Payment recorded", "id", cmd.ID, "number", invoice.InvoiceNumber, "amount", cmd.Amount, "remaining", invoice.RemainingAmount)
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

// GetOpenInvoices returns summary of open and overdue invoices
func (s *InvoiceService) GetOpenInvoices(ctx context.Context, tenantID string) (*OpenInvoicesSummaryDTO, error) {
	if tenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	invoices, err := s.invoiceRepo.GetOverdueInvoices(ctx, tenantID)
	if err != nil {
		return nil, domain.NewDomainError("QUERY_ERROR", "failed to get open invoices", err)
	}

	dtos := make([]*InvoiceDTO, len(invoices))
	var totalOpen, totalOverdue float64
	var countOpen, countOverdue int

	now := time.Now()
	for i, inv := range invoices {
		dtos[i] = InvoiceToDTO(inv)
		if inv.Status == domain.InvoiceOverdue || inv.DueDate.Before(now) {
			totalOverdue += inv.Total
			countOverdue++
		} else {
			totalOpen += inv.Total
			countOpen++
		}
	}

	return &OpenInvoicesSummaryDTO{
		TotalOpen:    totalOpen,
		TotalOverdue: totalOverdue,
		CountOpen:    countOpen,
		CountOverdue: countOverdue,
		Invoices:     dtos,
	}, nil
}

// GenerateInvoiceHTML generiert druckbares Rechnungs-HTML
func (s *InvoiceService) GenerateInvoiceHTML(ctx context.Context, tenantID, invoiceID string) (string, error) {
	invoice, err := s.invoiceRepo.GetByID(ctx, tenantID, invoiceID)
	if err != nil {
		return "", domain.NewDomainError("NOT_FOUND", "invoice not found", err)
	}

	html := buildInvoiceHTML(invoice)
	return html, nil
}

// GenerateInvoicePDF generiert ein professionelles PDF der Rechnung.
// Nutzt go-pdf/fpdf für reine Go-basierte PDF-Erzeugung (kein Chrome nötig).
func (s *InvoiceService) GenerateInvoicePDF(ctx context.Context, tenantID, invoiceID string) ([]byte, error) {
	invoice, err := s.invoiceRepo.GetByID(ctx, tenantID, invoiceID)
	if err != nil {
		return nil, domain.NewDomainError("NOT_FOUND", "invoice not found", err)
	}

	// Wenn ein PDF-Generator verfügbar ist, strukturiertes PDF erzeugen
	if s.pdfGenerator != nil {
		pdfBytes, err := s.pdfGenerator.GeneratePDF(ctx, buildInvoiceHTML(invoice))
		if err != nil {
			s.logger.Error("PDF generation failed, falling back to HTML", err)
			return []byte(buildInvoiceHTML(invoice)), nil
		}
		return pdfBytes, nil
	}

	// Fallback: HTML als Byte-Array
	s.logger.Warn("No PDF generator configured, returning HTML")
	return []byte(buildInvoiceHTML(invoice)), nil
}

// SendInvoiceWithPDF versendet eine Rechnung per E-Mail mit PDF-Anhang.
// Generiert das PDF, ändert den Status auf "sent" und verschickt die E-Mail.
func (s *InvoiceService) SendInvoiceWithPDF(ctx context.Context, cmd SendInvoiceCommand) (*InvoiceDTO, error) {
	if cmd.TenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	invoice, err := s.invoiceRepo.GetByID(ctx, cmd.TenantID, cmd.ID)
	if err != nil {
		return nil, domain.NewDomainError("NOT_FOUND", "invoice not found", err)
	}

	// GoBD-Integritätsprüfung
	if invoice.IsFinalized() && !invoice.VerifyHash() {
		return nil, domain.NewDomainError("INTEGRITY_CHECK", "invoice hash mismatch - possible tampering detected", nil)
	}

	// PDF generieren
	pdfBytes, err := s.GenerateInvoicePDF(ctx, cmd.TenantID, cmd.ID)
	if err != nil {
		return nil, domain.NewDomainError("PDF_ERROR", "failed to generate invoice PDF", err)
	}

	// Status auf "sent" setzen
	if err := invoice.Send(cmd.Email); err != nil {
		return nil, domain.NewDomainError("INVALID_TRANSITION", err.Error(), err)
	}

	if err := s.invoiceRepo.Update(ctx, invoice); err != nil {
		return nil, domain.NewDomainError("UPDATE_ERROR", "failed to update invoice status", err)
	}

	// E-Mail mit PDF-Anhang versenden (wenn Email-Sender konfiguriert)
	if s.emailSender != nil {
		subject := fmt.Sprintf("Rechnung %s", invoice.InvoiceNumber)
		body := fmt.Sprintf(`<html><body>
			<p>Sehr geehrte/r %s,</p>
			<p>anbei erhalten Sie die Rechnung <strong>%s</strong> über <strong>%s %.2f</strong>.</p>
			<p>Zahlbar bis: <strong>%s</strong></p>
			<p>Bei Fragen stehen wir Ihnen gerne zur Verfügung.</p>
			<p>Mit freundlichen Grüßen<br>Ihr EquipFlow-Team</p>
		</body></html>`,
			invoice.ClientName,
			invoice.InvoiceNumber,
			invoice.Currency, invoice.Total,
			invoice.DueDate.Format("02.01.2006"),
		)
		attachmentName := fmt.Sprintf("Rechnung_%s.pdf", invoice.InvoiceNumber)

		if err := s.emailSender.SendWithAttachment(cmd.Email, subject, body, pdfBytes, attachmentName); err != nil {
			s.logger.Error("Email sending failed", err, "invoice_id", cmd.ID, "email", cmd.Email)
			// Fehler loggen, aber Status bleibt auf "sent" (Rechnung wurde korrekt finalisiert)
			// Der Nutzer kann die E-Mail manuell erneut versenden
		} else {
			s.logger.Info("Invoice email sent", "id", cmd.ID, "number", invoice.InvoiceNumber, "email", cmd.Email)
		}
	} else {
		s.logger.Warn("No email sender configured, invoice marked as sent without email delivery")
	}

	s.logger.Info("Invoice sent", "id", cmd.ID, "number", invoice.InvoiceNumber, "email", cmd.Email)
	return InvoiceToDTO(invoice), nil
}

// CreateInvoiceFromProject creates invoice from project reference
func (s *InvoiceService) CreateInvoiceFromProject(ctx context.Context, tenantID, projectID string, clientName string) (*InvoiceDTO, error) {
	if tenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}
	if projectID == "" {
		return nil, domain.NewDomainError("INVALID_INPUT", "project ID is required", nil)
	}

	seq, err := s.seqRepo.GetNextNumber(ctx, tenantID, "invoice")
	if err != nil {
		return nil, domain.NewDomainError("SEQUENCE_ERROR", "failed to get next invoice number", err)
	}

	now := time.Now()
	invoiceNumber := fmt.Sprintf("RF-%d-%04d", now.Year(), seq)

	invoice := domain.NewInvoice(fmt.Sprintf("inv_%d", hashString(tenantID+invoiceNumber)), tenantID, invoiceNumber, clientName, "")
	invoice.ProjectID = &projectID
	invoice.Status = domain.InvoiceDraft
	invoice.IssueDate = now
	invoice.DueDate = now.AddDate(0, 1, 0)
	invoice.Currency = "EUR"

	if err := invoice.Validate(); err != nil {
		return nil, domain.NewDomainError("VALIDATION_ERROR", err.Error(), nil)
	}

	if err := s.invoiceRepo.Create(ctx, invoice); err != nil {
		return nil, domain.NewDomainError("CREATE_ERROR", "failed to create invoice", err)
	}

	s.logger.Info("Invoice created from project", "id", invoice.ID, "project_id", projectID, "tenant_id", tenantID)
	return InvoiceToDTO(invoice), nil
}

func buildInvoiceHTML(invoice *domain.Invoice) string {
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Invoice ` + invoice.InvoiceNumber + `</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 850px; margin: 0 auto; padding: 20px; }
        .header { margin-bottom: 40px; border-bottom: 2px solid #333; padding-bottom: 20px; }
        .company-info { margin-bottom: 40px; }
        .invoice-details { display: flex; justify-content: space-between; margin-bottom: 40px; }
        .detail-section { width: 45%; }
        .detail-label { font-weight: bold; color: #666; font-size: 12px; }
        .detail-value { margin-bottom: 15px; }
        table { width: 100%; border-collapse: collapse; margin: 30px 0; }
        th { background-color: #f0f0f0; border-bottom: 2px solid #333; padding: 10px; text-align: left; }
        td { border-bottom: 1px solid #ddd; padding: 10px; }
        .total-section { float: right; width: 300px; margin-top: 20px; }
        .total-row { display: flex; justify-content: space-between; padding: 10px 0; }
        .total-amount { font-weight: bold; font-size: 18px; border-top: 2px solid #333; padding-top: 10px; }
        .footer { margin-top: 60px; border-top: 1px solid #ddd; padding-top: 20px; font-size: 12px; color: #666; }
        @media print { body { margin: 0; padding: 0; } .no-print { display: none; } }
    </style>
</head>
<body>
    <div class="header">
        <h1>INVOICE</h1>
        <p style="margin: 0; font-size: 14px;">Invoice #` + invoice.InvoiceNumber + `</p>
    </div>

    <div class="invoice-details">
        <div class="detail-section">
            <div class="detail-label">INVOICE TO:</div>
            <div class="detail-value">
                <strong>` + invoice.ClientName + `</strong><br>
                ` + invoice.ClientAddress.Street + `<br>
                ` + invoice.ClientAddress.PostCode + ` ` + invoice.ClientAddress.City + `<br>
                ` + invoice.ClientAddress.Country + `<br>
                ` + invoice.ClientEmail + `
            </div>
        </div>
        <div class="detail-section">
            <div class="detail-row"><div class="detail-label">Invoice Date:</div><div>` + invoice.IssueDate.Format("2006-01-02") + `</div></div>
            <div class="detail-row"><div class="detail-label">Due Date:</div><div>` + invoice.DueDate.Format("2006-01-02") + `</div></div>
            <div class="detail-row"><div class="detail-label">Status:</div><div>` + string(invoice.Status) + `</div></div>
        </div>
    </div>

    <table>
        <thead>
            <tr>
                <th style="width: 40%;">Description</th>
                <th style="width: 10%; text-align: right;">Qty</th>
                <th style="width: 15%; text-align: right;">Unit Price</th>` +
		func() string {
			if !invoice.IsKleinunternehmer {
				return `
                <th style="width: 10%; text-align: right;">MwSt. %</th>
                <th style="width: 10%; text-align: right;">MwSt.</th>`
			}
			return ""
		}() + `
                <th style="width: 15%; text-align: right;">Total</th>
            </tr>
        </thead>
        <tbody>`

	for _, item := range invoice.Items {
		html += `<tr>
            <td>` + item.Description + `</td>
            <td style="text-align: right;">` + fmt.Sprintf("%.2f", item.Quantity) + ` ` + item.Unit + `</td>
            <td style="text-align: right;">` + invoice.Currency + ` ` + fmt.Sprintf("%.2f", item.UnitPrice) + `</td>`
		if !invoice.IsKleinunternehmer {
			html += `
            <td style="text-align: right;">` + fmt.Sprintf("%.0f", float64(item.TaxRate)) + `%</td>
            <td style="text-align: right;">` + invoice.Currency + ` ` + fmt.Sprintf("%.2f", item.TaxAmount) + `</td>`
		}
		html += `
            <td style="text-align: right;">` + invoice.Currency + ` ` + fmt.Sprintf("%.2f", item.TotalPrice) + `</td>
        </tr>`
	}

	html += `</tbody>
    </table>

    <div class="total-section">
        <div class="total-row">
            <span>Subtotal:</span>
            <span>` + invoice.Currency + ` ` + fmt.Sprintf("%.2f", invoice.SubTotal) + `</span>
        </div>`

	if !invoice.IsKleinunternehmer {
		html += `
        <div class="total-row">
            <span>Tax:</span>
            <span>` + invoice.Currency + ` ` + fmt.Sprintf("%.2f", invoice.TaxAmount) + `</span>
        </div>`
	}

	html += `
        <div class="total-row total-amount">
            <span>TOTAL:</span>
            <span>` + invoice.Currency + ` ` + fmt.Sprintf("%.2f", invoice.Total) + `</span>
        </div>
    </div>`

	if invoice.IsKleinunternehmer && invoice.KleinunternehmerText != "" {
		html += `
    <div style="margin-top: 20px; padding: 10px; background: #f0f7ff; border: 1px solid #b3d4fc; border-radius: 4px; font-size: 12px;">
        <strong>Hinweis:</strong> ` + invoice.KleinunternehmerText + `
    </div>`
	}

	html += `

    <div class="footer">
        <p><strong>Payment Terms:</strong> Due by ` + invoice.DueDate.Format("2006-01-02") + `</p>
        <p><strong>Notes:</strong> ` + invoice.Notes + `</p>
        <p style="margin-top: 20px; color: #999;">This is an automated invoice. For questions, please contact our accounting department.</p>
    </div>
</body>
</html>`

	return html
}

func buildQuoteHTML(quote *domain.Quote) string {
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Quote ` + quote.QuoteNumber + `</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 850px; margin: 0 auto; padding: 20px; }
        .header { margin-bottom: 40px; border-bottom: 2px solid #333; padding-bottom: 20px; }
        table { width: 100%; border-collapse: collapse; margin: 30px 0; }
        th { background-color: #f0f0f0; border-bottom: 2px solid #333; padding: 10px; text-align: left; }
        td { border-bottom: 1px solid #ddd; padding: 10px; }
        .total-section { float: right; width: 300px; margin-top: 20px; }
        .total-row { display: flex; justify-content: space-between; padding: 10px 0; }
        @media print { body { margin: 0; padding: 0; } }
    </style>
</head>
<body>
    <div class="header">
        <h1>QUOTE / ANGEBOT</h1>
        <p>Quote #` + quote.QuoteNumber + ` | Valid until ` + quote.ValidUntil.Format("2006-01-02") + `</p>
    </div>

    <div class="client-info">
        <strong>` + quote.ClientName + `</strong><br>
        ` + quote.ClientAddress.Street + `<br>
        ` + quote.ClientAddress.PostCode + ` ` + quote.ClientAddress.City + `
    </div>

    <table>
        <thead>
            <tr>
                <th>Description</th>
                <th style="text-align: right;">Qty</th>
                <th style="text-align: right;">Unit Price</th>
                <th style="text-align: right;">Total</th>
            </tr>
        </thead>
        <tbody>`

	for _, item := range quote.Items {
		html += `<tr>
            <td>` + item.Description + `</td>
            <td style="text-align: right;">` + fmt.Sprintf("%.2f", item.Quantity) + `</td>
            <td style="text-align: right;">` + quote.Currency + ` ` + fmt.Sprintf("%.2f", item.UnitPrice) + `</td>
            <td style="text-align: right;">` + quote.Currency + ` ` + fmt.Sprintf("%.2f", item.TotalPrice) + `</td>
        </tr>`
	}

	html += `</tbody>
    </table>

    <div class="total-section">
        <div class="total-row"><span>Subtotal:</span><span>` + quote.Currency + ` ` + fmt.Sprintf("%.2f", quote.SubTotal) + `</span></div>
        <div class="total-row"><span>Tax:</span><span>` + quote.Currency + ` ` + fmt.Sprintf("%.2f", quote.TaxAmount) + `</span></div>
        <div class="total-row"><span><strong>TOTAL:</strong></span><span><strong>` + quote.Currency + ` ` + fmt.Sprintf("%.2f", quote.Total) + `</strong></span></div>
    </div>
</body>
</html>`

	return html
}

// DeliveryNoteData holds all data for the delivery note template
type DeliveryNoteData struct {
	CompanyName        string             `json:"company_name"`
	CompanyStreet      string             `json:"company_street"`
	CompanyPostCode    string             `json:"company_post_code"`
	CompanyCity        string             `json:"company_city"`
	CompanyPhone       string             `json:"company_phone"`
	CompanyEmail       string             `json:"company_email"`
	CompanyWebsite     string             `json:"company_website"`
	ManagingDirector   string             `json:"managing_director"`
	LogoURL            string             `json:"logo_url"`
	ClientName         string             `json:"client_name"`
	ClientStreet       string             `json:"client_street"`
	ClientPostCode     string             `json:"client_post_code"`
	ClientCity         string             `json:"client_city"`
	ContactPerson      string             `json:"contact_person"`
	ContactPhone       string             `json:"contact_phone"`
	DeliveryNoteNumber string             `json:"delivery_note_number"`
	DeliveryDate       string             `json:"delivery_date"`
	ProjectReference   string             `json:"project_reference"`
	ProjectName        string             `json:"project_name"`
	InvoiceNumber      string             `json:"invoice_number"`
	DeliveryLocation   string             `json:"delivery_location"`
	ReturnDate         string             `json:"return_date"`
	Notes              string             `json:"notes"`
	Items              []DeliveryNoteItem `json:"items"`
}

// DeliveryNoteItem represents a single item on the delivery note
type DeliveryNoteItem struct {
	Description       string `json:"description"`
	Category          string `json:"category"`
	SerialNumber      string `json:"serial_number"`
	QuantityFormatted string `json:"quantity_formatted"`
	Unit              string `json:"unit"`
	Condition         string `json:"condition"`
	Note              string `json:"note"`
}

// GenerateDeliveryNoteHTML renders the delivery_note.html template with the given data
func (s *InvoiceService) GenerateDeliveryNoteHTML(ctx context.Context, data *DeliveryNoteData) (string, error) {
	tmplPath := "services/invoice-service/internal/infrastructure/pdf/templates/delivery_note.html"

	funcMap := template.FuncMap{
		"inc": func(i int) int { return i + 1 },
	}

	tmpl, err := template.New("delivery_note.html").Funcs(funcMap).ParseFiles(tmplPath)
	if err != nil {
		return "", fmt.Errorf("failed to parse delivery note template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute delivery note template: %w", err)
	}

	return buf.String(), nil
}

// GenerateDeliveryNotePDF generates a PDF from the delivery note template
func (s *InvoiceService) GenerateDeliveryNotePDF(ctx context.Context, data *DeliveryNoteData) ([]byte, error) {
	html, err := s.GenerateDeliveryNoteHTML(ctx, data)
	if err != nil {
		return nil, err
	}

	if s.pdfGenerator != nil {
		pdfBytes, err := s.pdfGenerator.GeneratePDF(ctx, html)
		if err != nil {
			s.logger.Error("Delivery note PDF generation failed, falling back to HTML", err)
			return []byte(html), nil
		}
		return pdfBytes, nil
	}

	return []byte(html), nil
}

func hashString(s string) int64 {
	h := int64(5381)
	for _, c := range s {
		h = ((h << 5) + h) + int64(c)
	}
	return h & 0x7FFFFFFFFFFFFFFF
}

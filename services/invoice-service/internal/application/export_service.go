package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/invoice-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/invoice-service/internal/ports"
)

type ExportService struct {
	invoiceRepo ports.InvoiceRepository
	logger      *logger.Logger
}

func NewExportService(
	invoiceRepo ports.InvoiceRepository,
	logger *logger.Logger,
) *ExportService {
	return &ExportService{
		invoiceRepo: invoiceRepo,
		logger:      logger,
	}
}

// ExportDATEV exports invoices in German DATEV format for accounting software
// Uses SKR03 (Standardkontenrahmen) accounting scheme
func (s *ExportService) ExportDATEV(ctx context.Context, query ExportDATEVQuery) ([]*DATEVExportRow, error) {
	if query.TenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	fromDate := query.FromDate
	toDate := query.ToDate

	listQuery := &ports.InvoiceListQuery{
		TenantID: query.TenantID,
		FromDate: &fromDate,
		ToDate:   &toDate,
		Limit:    1000,
		Offset:   0,
	}

	result, err := s.invoiceRepo.List(ctx, listQuery)
	if err != nil {
		return nil, domain.NewDomainError("QUERY_ERROR", "failed to list invoices", err)
	}

	var rows []*DATEVExportRow

	// SKR03 GL account mappings for revenues (Erlöse)
	// 8400 = Sales 19% VAT
	// 8300 = Sales 7% VAT
	// 8000 = Sales 0% VAT
	// AR accounts typically use 1200 = Forderungen (Receivables)

	for _, invoice := range result.Items {
		if invoice.Status == domain.InvoicePaid || invoice.Status == domain.InvoiceDraft {
			continue // Only export sent/overdue invoices
		}

		// Determine GL account based on tax rate
		glAccount := s.getGLAccountForTaxRate(invoice.TaxRate)

		// Revenue entry
		row := &DATEVExportRow{
			Umsatz:       invoice.Total,
			SollHaben:    "S", // Soll (debit) for receivables
			WKZUmsatz:    invoice.Currency,
			Konto:        "1200", // AR account (Forderungen)
			Gegenkonto:   glAccount,
			Belegdatum:   invoice.IssueDate.Format("0102"), // DDMM format
			Belegnummer:  invoice.InvoiceNumber,
			Buchungstext: fmt.Sprintf("Rechnung %s - %s", invoice.InvoiceNumber, invoice.ClientName),
		}
		rows = append(rows, row)

		// Revenue account entry (opposite)
		revRow := &DATEVExportRow{
			Umsatz:       invoice.SubTotal,
			SollHaben:    "H", // Haben (credit) for revenue
			WKZUmsatz:    invoice.Currency,
			Konto:        glAccount,
			Gegenkonto:   "1200",
			Belegdatum:   invoice.IssueDate.Format("0102"),
			Belegnummer:  invoice.InvoiceNumber,
			Buchungstext: fmt.Sprintf("Rechnung %s - %s", invoice.InvoiceNumber, invoice.ClientName),
		}
		rows = append(rows, revRow)

		// VAT liability entry if tax > 0
		if invoice.TaxAmount > 0 {
			vatAccount := s.getVATAccountForRate(invoice.TaxRate)
			vatRow := &DATEVExportRow{
				Umsatz:       invoice.TaxAmount,
				SollHaben:    "H", // Haben (credit) for VAT liability
				WKZUmsatz:    invoice.Currency,
				Konto:        vatAccount,
				Gegenkonto:   glAccount,
				Belegdatum:   invoice.IssueDate.Format("0102"),
				Belegnummer:  invoice.InvoiceNumber,
				Buchungstext: fmt.Sprintf("USt. %d%% Rechnung %s", int(invoice.TaxRate), invoice.InvoiceNumber),
			}
			rows = append(rows, vatRow)
		}
	}

	s.logger.Info("DATEV export generated", "tenant_id", query.TenantID, "rows", len(rows))
	return rows, nil
}

// ExportCSV exports invoices/quotes in CSV format
func (s *ExportService) ExportCSV(ctx context.Context, query ExportCSVQuery) (string, error) {
	if query.TenantID == "" {
		return "", domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	listQuery := &ports.InvoiceListQuery{
		TenantID: query.TenantID,
		FromDate: &query.FromDate,
		ToDate:   &query.ToDate,
		Limit:    10000,
		Offset:   0,
	}

	result, err := s.invoiceRepo.List(ctx, listQuery)
	if err != nil {
		return "", domain.NewDomainError("QUERY_ERROR", "failed to list invoices", err)
	}

	var sb strings.Builder
	sb.WriteString("Rechnungsnummer,Datum,Fälligkeitsdatum,Kunde,Betrag,USt.,Gesamtbetrag,Status,Zahlungsart\n")

	for _, invoice := range result.Items {
		// Escape CSV fields
		clientName := strings.ReplaceAll(invoice.ClientName, "\"", "\"\"")

		sb.WriteString(fmt.Sprintf(
			"\"%s\",\"%s\",\"%s\",\"%s\",%.2f,%.2f,%.2f,\"%s\",\"%s\"\n",
			invoice.InvoiceNumber,
			invoice.IssueDate.Format("2006-01-02"),
			invoice.DueDate.Format("2006-01-02"),
			clientName,
			invoice.SubTotal,
			invoice.TaxAmount,
			invoice.Total,
			string(invoice.Status),
			invoice.PaymentMethod,
		))
	}

	s.logger.Info("CSV export generated", "tenant_id", query.TenantID, "records", len(result.Items))
	return sb.String(), nil
}

// ExportWISO exports in WISO format (popular German accounting software)
func (s *ExportService) ExportWISO(ctx context.Context, tenantID, fromDate, toDate string) (string, error) {
	if tenantID == "" {
		return "", domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	listQuery := &ports.InvoiceListQuery{
		TenantID: tenantID,
		FromDate: &fromDate,
		ToDate:   &toDate,
		Limit:    10000,
		Offset:   0,
	}

	result, err := s.invoiceRepo.List(ctx, listQuery)
	if err != nil {
		return "", domain.NewDomainError("QUERY_ERROR", "failed to list invoices", err)
	}

	// WISO format is a simplified tab-separated format
	var sb strings.Builder
	sb.WriteString("Belegnummer\tDatum\tKunde\tBetrag\tUSt.-Satz\tUSt.\tKonto\tGegenkonto\n")

	for _, invoice := range result.Items {
		glAccount := s.getGLAccountForTaxRate(invoice.TaxRate)
		sb.WriteString(fmt.Sprintf(
			"%s\t%s\t%s\t%.2f\t%d%%\t%.2f\t%s\t1200\n",
			invoice.InvoiceNumber,
			invoice.IssueDate.Format("02.01.2006"), // German date format
			invoice.ClientName,
			invoice.SubTotal,
			int(invoice.TaxRate),
			invoice.TaxAmount,
			glAccount,
		))
	}

	s.logger.Info("WISO export generated", "tenant_id", tenantID, "records", len(result.Items))
	return sb.String(), nil
}

// Helper functions

// getGLAccountForTaxRate returns the correct GL account for a tax rate (SKR03)
func (s *ExportService) getGLAccountForTaxRate(rate float64) string {
	switch rate {
	case 19:
		return "8400" // Sales 19%
	case 7:
		return "8300" // Sales 7%
	case 0:
		return "8000" // Sales 0%
	default:
		return "8100" // Other sales
	}
}

// getVATAccountForRate returns the VAT liability account
func (s *ExportService) getVATAccountForRate(rate float64) string {
	// SKR03 VAT accounts
	// 3801 = VAT 19%
	// 3806 = VAT 7%
	// 3811 = VAT other
	switch rate {
	case 19:
		return "3801"
	case 7:
		return "3806"
	default:
		return "3811"
	}
}

// CalculateDueAmounts calculates totals for a date range
func (s *ExportService) CalculateDueAmounts(ctx context.Context, tenantID, fromDate, toDate string) (map[string]interface{}, error) {
	if tenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	listQuery := &ports.InvoiceListQuery{
		TenantID: tenantID,
		FromDate: &fromDate,
		ToDate:   &toDate,
		Limit:    10000,
		Offset:   0,
	}

	result, err := s.invoiceRepo.List(ctx, listQuery)
	if err != nil {
		return nil, domain.NewDomainError("QUERY_ERROR", "failed to list invoices", err)
	}

	totals := map[string]interface{}{
		"total_invoiced": 0.0,
		"total_paid":     0.0,
		"total_overdue":  0.0,
		"tax_collected":  0.0,
		"count":          0,
	}

	now := time.Now()

	for _, invoice := range result.Items {
		count := totals["count"].(int)
		totals["count"] = count + 1

		total := totals["total_invoiced"].(float64)
		totals["total_invoiced"] = total + invoice.Total

		tax := totals["tax_collected"].(float64)
		totals["tax_collected"] = tax + invoice.TaxAmount

		if invoice.Status == domain.InvoicePaid {
			paid := totals["total_paid"].(float64)
			totals["total_paid"] = paid + invoice.Total
		}

		if now.After(invoice.DueDate) && invoice.Status != domain.InvoicePaid {
			overdue := totals["total_overdue"].(float64)
			totals["total_overdue"] = overdue + invoice.Total
		}
	}

	return totals, nil
}

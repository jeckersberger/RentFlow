package application

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/services/invoice/internal/domain"
)

// ---------------------------------------------------------------------------
// Request DTO
// ---------------------------------------------------------------------------

// DatevExportRequest holds the parameters for a DATEV export.
type DatevExportRequest struct {
	FromDate string `json:"from_date"` // YYYY-MM-DD
	ToDate   string `json:"to_date"`   // YYYY-MM-DD
	Format   string `json:"format"`    // "skr03" or "skr04"
}

// ---------------------------------------------------------------------------
// Service
// ---------------------------------------------------------------------------

// DatevService generates DATEV-compatible CSV exports.
type DatevService struct {
	invoiceRepo domain.InvoiceRepository
	logger      zerolog.Logger
}

// NewDatevService creates a new DatevService.
func NewDatevService(invoiceRepo domain.InvoiceRepository, logger zerolog.Logger) *DatevService {
	return &DatevService{
		invoiceRepo: invoiceRepo,
		logger:      logger.With().Str("service", "datev").Logger(),
	}
}

// ExportCSV generates a DATEV-compatible CSV for the given date range and SKR format.
func (s *DatevService) ExportCSV(ctx context.Context, tenantID uuid.UUID, req DatevExportRequest) ([]byte, error) {
	format := strings.ToLower(req.Format)
	if format == "" {
		format = "skr03"
	}
	if format != "skr03" && format != "skr04" {
		return nil, fmt.Errorf("Ungueltiges Format: %s (erlaubt: skr03, skr04)", format)
	}

	fromDate, err := time.Parse("2006-01-02", req.FromDate)
	if err != nil {
		return nil, fmt.Errorf("Ungueltiges Von-Datum: %s", req.FromDate)
	}

	toDate, err := time.Parse("2006-01-02", req.ToDate)
	if err != nil {
		return nil, fmt.Errorf("Ungueltiges Bis-Datum: %s", req.ToDate)
	}

	// Load all finalized/sent/paid invoices in the period
	invoices, err := s.loadInvoicesInRange(ctx, tenantID, fromDate, toDate)
	if err != nil {
		return nil, fmt.Errorf("Rechnungen laden: %w", err)
	}

	var buf bytes.Buffer

	// Write DATEV header
	s.writeHeader(&buf, fromDate, toDate)

	// Write column headers
	buf.WriteString("Umsatz;Soll/Haben;Konto;Gegenkonto;BU-Schluessel;Belegdatum;Belegnummer;Buchungstext\r\n")

	// Write invoice rows
	for _, inv := range invoices {
		s.writeInvoiceRow(&buf, inv, format)
	}

	s.logger.Info().
		Str("tenant_id", tenantID.String()).
		Str("format", format).
		Int("invoices", len(invoices)).
		Msg("DATEV export generated")

	return buf.Bytes(), nil
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

func (s *DatevService) writeHeader(buf *bytes.Buffer, from, to time.Time) {
	// DATEV header metadata line
	buf.WriteString(fmt.Sprintf(
		"\"EXTF\";510;21;\"Buchungsstapel\";7;;;\"%s\";\"%s\";%d;;;;\r\n",
		from.Format("20060102"),
		to.Format("20060102"),
		from.Year(),
	))
}

func (s *DatevService) writeInvoiceRow(buf *bytes.Buffer, inv *domain.Invoice, format string) {
	// Determine revenue account based on VAT rate and Kleinunternehmer status
	revenueAccount := s.revenueAccount(inv, format)

	// Customer account: 10000 + sequential (we use a hash of customer name for simplicity)
	customerAccount := s.customerAccount(inv)

	// Format amount: cents to EUR with comma as decimal separator, no thousand separators
	amountStr := formatDatevAmount(inv.TotalGross)

	// Soll/Haben: S = Soll (Debit) for receivable
	sollHaben := "S"

	// BU-Schluessel (tax key)
	buSchluessel := s.taxKey(inv)

	// Belegdatum: DDMM format
	belegdatum := formatDatevDate(inv.InvoiceDate)

	// Buchungstext
	buchungstext := fmt.Sprintf("Rechnung %s %s", inv.InvoiceNumber, inv.CustomerName)
	if len(buchungstext) > 60 {
		buchungstext = buchungstext[:60]
	}

	buf.WriteString(fmt.Sprintf(
		"%s;%s;%d;%d;%s;%s;%s;%s\r\n",
		amountStr,
		sollHaben,
		customerAccount,
		revenueAccount,
		buSchluessel,
		belegdatum,
		inv.InvoiceNumber,
		buchungstext,
	))
}

func (s *DatevService) revenueAccount(inv *domain.Invoice, format string) int {
	if inv.Kleinunternehmer {
		if format == "skr04" {
			return 4185 // SKR04 Kleinunternehmer
		}
		return 8195 // SKR03 Kleinunternehmer
	}

	// Determine by VAT rate
	switch inv.VatRate {
	case 1900:
		if format == "skr04" {
			return 4400 // SKR04 Erloese 19%
		}
		return 8400 // SKR03 Erloese 19%
	case 700:
		if format == "skr04" {
			return 4300 // SKR04 Erloese 7%
		}
		return 8300 // SKR03 Erloese 7%
	default:
		if format == "skr04" {
			return 4185
		}
		return 8195 // Default: tax-free
	}
}

func (s *DatevService) customerAccount(inv *domain.Invoice) int {
	// DATEV convention: customer accounts start at 10000
	// We use a simple hash of customer name to generate a stable account number
	hash := 0
	for _, c := range inv.CustomerName {
		hash = (hash*31 + int(c)) % 90000
	}
	return 10000 + hash
}

func (s *DatevService) taxKey(inv *domain.Invoice) string {
	if inv.Kleinunternehmer {
		return "" // No tax key for Kleinunternehmer
	}
	switch inv.VatRate {
	case 1900:
		return "3" // Automatik USt 19%
	case 700:
		return "2" // Automatik USt 7%
	default:
		return ""
	}
}

func (s *DatevService) loadInvoicesInRange(ctx context.Context, tenantID uuid.UUID, from, to time.Time) ([]*domain.Invoice, error) {
	var allInvoices []*domain.Invoice

	// Load invoices with relevant statuses
	for _, status := range []string{domain.StatusFinalized, domain.StatusSent, domain.StatusPartialPaid, domain.StatusPaid, domain.StatusOverdue} {
		filter := domain.InvoiceFilter{
			Page:    1,
			PerPage: 1000,
			Status:  status,
		}
		invoices, _, err := s.invoiceRepo.List(ctx, tenantID, filter)
		if err != nil {
			return nil, err
		}

		for _, inv := range invoices {
			invDate, parseErr := time.Parse("2006-01-02", inv.InvoiceDate)
			if parseErr != nil {
				continue
			}
			// Check if invoice date is within range (inclusive)
			if (invDate.Equal(from) || invDate.After(from)) && (invDate.Equal(to) || invDate.Before(to)) {
				allInvoices = append(allInvoices, inv)
			}
		}
	}

	return allInvoices, nil
}

// formatDatevAmount converts cents to DATEV amount format (e.g., 12345 -> "123,45").
func formatDatevAmount(cents int64) string {
	negative := cents < 0
	if negative {
		cents = -cents
	}
	eur := cents / 100
	ct := cents % 100
	s := fmt.Sprintf("%d,%02d", eur, ct)
	if negative {
		s = "-" + s
	}
	return s
}

// formatDatevDate converts YYYY-MM-DD to DDMM (4 digits, DATEV format).
func formatDatevDate(dateStr string) string {
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%02d%02d", t.Day(), t.Month())
}

package application

import (
	"context"
	"encoding/csv"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/services/invoice/internal/domain"
)

// ---------------------------------------------------------------------------
// Request DTOs
// ---------------------------------------------------------------------------

// ConfirmMatchRequest holds the data for manually confirming a bank transaction match.
type ConfirmMatchRequest struct {
	TransactionID uuid.UUID `json:"transaction_id"`
	InvoiceID     uuid.UUID `json:"invoice_id"`
}

// ---------------------------------------------------------------------------
// Service
// ---------------------------------------------------------------------------

// BankingService implements bank import and auto-matching use cases.
type BankingService struct {
	bankRepo    domain.BankRepository
	invoiceRepo domain.InvoiceRepository
	paymentRepo domain.PaymentRepository
	logger      zerolog.Logger
}

// NewBankingService creates a new BankingService.
func NewBankingService(
	bankRepo domain.BankRepository,
	invoiceRepo domain.InvoiceRepository,
	paymentRepo domain.PaymentRepository,
	logger zerolog.Logger,
) *BankingService {
	return &BankingService{
		bankRepo:    bankRepo,
		invoiceRepo: invoiceRepo,
		paymentRepo: paymentRepo,
		logger:      logger.With().Str("service", "banking").Logger(),
	}
}

// ImportCSV parses a German bank CSV (semicolon-separated) and creates BankTransactions.
// Expected columns: Buchungstag;Wertstellung;Betrag;Waehrung;Verwendungszweck;Auftraggeber;IBAN
func (s *BankingService) ImportCSV(ctx context.Context, tenantID uuid.UUID, csvData []byte) ([]*domain.BankTransaction, error) {
	reader := csv.NewReader(strings.NewReader(string(csvData)))
	reader.Comma = ';'
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true

	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrCSVParseError, err)
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("CSV-Datei enthaelt keine Daten")
	}

	batchID := uuid.New()
	var transactions []*domain.BankTransaction

	// Skip header row (index 0)
	for i, row := range records[1:] {
		if len(row) < 7 {
			s.logger.Warn().Int("line", i+2).Msg("CSV-Zeile uebersprungen: zu wenige Spalten")
			continue
		}

		bookingDate := parseGermanDate(strings.TrimSpace(row[0]))
		if bookingDate == "" {
			s.logger.Warn().Int("line", i+2).Str("raw", row[0]).Msg("Buchungstag konnte nicht gelesen werden")
			continue
		}

		valueDate := parseGermanDate(strings.TrimSpace(row[1]))

		amount, parseErr := parseGermanAmount(strings.TrimSpace(row[2]))
		if parseErr != nil {
			s.logger.Warn().Int("line", i+2).Str("raw", row[2]).Msg("Betrag konnte nicht gelesen werden")
			continue
		}

		currency := strings.TrimSpace(row[3])
		if currency == "" {
			currency = "EUR"
		}

		reference := strings.TrimSpace(row[4])
		counterpartyName := strings.TrimSpace(row[5])
		counterpartyIBAN := strings.TrimSpace(row[6])

		tx := &domain.BankTransaction{
			ID:               uuid.New(),
			TenantID:         tenantID,
			BookingDate:      bookingDate,
			ValueDate:        valueDate,
			Amount:           amount,
			Currency:         currency,
			Reference:        reference,
			CounterpartyName: counterpartyName,
			CounterpartyIBAN: counterpartyIBAN,
			MatchConfidence:  "none",
			ImportSource:     "csv",
			ImportBatchID:    &batchID,
		}

		if err := s.bankRepo.Create(ctx, tx); err != nil {
			return nil, fmt.Errorf("bank transaction erstellen (Zeile %d): %w", i+2, err)
		}
		transactions = append(transactions, tx)
	}

	s.logger.Info().
		Str("tenant_id", tenantID.String()).
		Str("batch_id", batchID.String()).
		Int("count", len(transactions)).
		Msg("bank transactions imported")

	return transactions, nil
}

// AutoMatch tries to match unmatched bank transactions against open invoices.
func (s *BankingService) AutoMatch(ctx context.Context, tenantID uuid.UUID) (int, error) {
	// Get unmatched transactions
	matchedFalse := false
	transactions, err := s.bankRepo.List(ctx, tenantID, &matchedFalse, 500)
	if err != nil {
		return 0, fmt.Errorf("auto-match: Transaktionen laden: %w", err)
	}

	// Get open invoices (finalized, sent, partial_paid)
	openInvoices, err := s.getOpenInvoices(ctx, tenantID)
	if err != nil {
		return 0, fmt.Errorf("auto-match: offene Rechnungen laden: %w", err)
	}

	matched := 0
	for _, tx := range transactions {
		// Only match incoming payments (positive amounts)
		if tx.Amount <= 0 {
			continue
		}

		invoiceID, confidence := s.findBestMatch(tx, openInvoices)
		if invoiceID == nil {
			continue
		}

		if err := s.bankRepo.UpdateMatch(ctx, tx.ID, tenantID, *invoiceID, confidence); err != nil {
			s.logger.Error().Err(err).
				Str("tx_id", tx.ID.String()).
				Msg("auto-match: Zuordnung speichern fehlgeschlagen")
			continue
		}
		matched++
	}

	s.logger.Info().
		Str("tenant_id", tenantID.String()).
		Int("matched", matched).
		Int("total_unmatched", len(transactions)).
		Msg("auto-match completed")

	return matched, nil
}

// ConfirmMatch manually confirms a match between a bank transaction and an invoice,
// and records a payment on the invoice.
func (s *BankingService) ConfirmMatch(ctx context.Context, tenantID uuid.UUID, req ConfirmMatchRequest) error {
	// Verify invoice exists
	inv, err := s.invoiceRepo.GetByID(ctx, req.InvoiceID, tenantID)
	if err != nil {
		return fmt.Errorf("confirm-match: Rechnung nicht gefunden: %w", err)
	}

	// Update the bank transaction match
	if err := s.bankRepo.UpdateMatch(ctx, req.TransactionID, tenantID, req.InvoiceID, "confirmed"); err != nil {
		return fmt.Errorf("confirm-match: Zuordnung speichern: %w", err)
	}

	// Get the transaction to read the amount
	transactions, err := s.bankRepo.List(ctx, tenantID, nil, 1000)
	if err != nil {
		return fmt.Errorf("confirm-match: Transaktion laden: %w", err)
	}

	var txAmount int64
	for _, tx := range transactions {
		if tx.ID == req.TransactionID {
			txAmount = tx.Amount
			break
		}
	}

	if txAmount <= 0 {
		return nil // No payment to record for negative amounts
	}

	// Record payment on the invoice
	remaining := inv.TotalGross - inv.AmountPaid
	paymentAmount := txAmount
	if paymentAmount > remaining {
		paymentAmount = remaining
	}

	if paymentAmount > 0 {
		payment := &domain.Payment{
			ID:            uuid.New(),
			TenantID:      tenantID,
			InvoiceID:     req.InvoiceID,
			Amount:        paymentAmount,
			PaymentDate:   time.Now().Format("2006-01-02"),
			PaymentMethod: "bank_transfer",
			Reference:     fmt.Sprintf("Bank-Import TX %s", req.TransactionID.String()[:8]),
		}

		if err := s.paymentRepo.Create(ctx, payment); err != nil {
			return fmt.Errorf("confirm-match: Zahlung erstellen: %w", err)
		}

		newAmountPaid := inv.AmountPaid + paymentAmount
		if err := s.invoiceRepo.UpdateAmountPaid(ctx, req.InvoiceID, tenantID, newAmountPaid); err != nil {
			return fmt.Errorf("confirm-match: Betrag aktualisieren: %w", err)
		}

		newStatus := domain.StatusPartialPaid
		if newAmountPaid >= inv.TotalGross {
			newStatus = domain.StatusPaid
		}
		if err := s.invoiceRepo.UpdateStatus(ctx, req.InvoiceID, tenantID, newStatus); err != nil {
			return fmt.Errorf("confirm-match: Status aktualisieren: %w", err)
		}
	}

	s.logger.Info().
		Str("tx_id", req.TransactionID.String()).
		Str("invoice_id", req.InvoiceID.String()).
		Int64("amount", txAmount).
		Msg("bank match confirmed and payment recorded")

	return nil
}

// ListTransactions returns bank transactions with optional match filter.
func (s *BankingService) ListTransactions(ctx context.Context, tenantID uuid.UUID, matched *bool) ([]*domain.BankTransaction, error) {
	return s.bankRepo.List(ctx, tenantID, matched, 500)
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

func (s *BankingService) getOpenInvoices(ctx context.Context, tenantID uuid.UUID) ([]*domain.Invoice, error) {
	var allOpen []*domain.Invoice

	for _, status := range []string{domain.StatusFinalized, domain.StatusSent, domain.StatusPartialPaid, domain.StatusOverdue} {
		filter := domain.InvoiceFilter{
			Page:    1,
			PerPage: 500,
			Status:  status,
		}
		invoices, _, err := s.invoiceRepo.List(ctx, tenantID, filter)
		if err != nil {
			return nil, err
		}
		allOpen = append(allOpen, invoices...)
	}

	return allOpen, nil
}

func (s *BankingService) findBestMatch(tx *domain.BankTransaction, invoices []*domain.Invoice) (*uuid.UUID, string) {
	reference := strings.ToUpper(tx.Reference)

	// 1. Exact match: amount matches total_gross AND reference contains invoice_number
	for _, inv := range invoices {
		remaining := inv.TotalGross - inv.AmountPaid
		if remaining <= 0 {
			continue
		}
		invNum := strings.ToUpper(inv.InvoiceNumber)
		if tx.Amount == remaining && strings.Contains(reference, invNum) {
			return &inv.ID, "high"
		}
	}

	// 2. Amount match: amount matches remaining within 1 cent tolerance
	for _, inv := range invoices {
		remaining := inv.TotalGross - inv.AmountPaid
		if remaining <= 0 {
			continue
		}
		if math.Abs(float64(tx.Amount-remaining)) <= 1 {
			return &inv.ID, "medium"
		}
	}

	// 3. Name match: counterparty_name contains customer_name
	counterpartyUpper := strings.ToUpper(tx.CounterpartyName)
	if counterpartyUpper != "" {
		for _, inv := range invoices {
			remaining := inv.TotalGross - inv.AmountPaid
			if remaining <= 0 {
				continue
			}
			customerUpper := strings.ToUpper(inv.CustomerName)
			if customerUpper != "" && strings.Contains(counterpartyUpper, customerUpper) {
				return &inv.ID, "low"
			}
		}
	}

	return nil, ""
}

// parseGermanDate parses DD.MM.YYYY or YYYY-MM-DD into YYYY-MM-DD.
func parseGermanDate(s string) string {
	if s == "" {
		return ""
	}

	// Try DD.MM.YYYY first
	t, err := time.Parse("02.01.2006", s)
	if err == nil {
		return t.Format("2006-01-02")
	}

	// Try YYYY-MM-DD
	t, err = time.Parse("2006-01-02", s)
	if err == nil {
		return t.Format("2006-01-02")
	}

	return ""
}

// parseGermanAmount parses German-formatted amount (e.g. "1.234,56" or "-1234,56") into int64 cents.
func parseGermanAmount(s string) (int64, error) {
	if s == "" {
		return 0, fmt.Errorf("leerer Betrag")
	}

	// Remove thousand separators (dots in German format)
	s = strings.ReplaceAll(s, ".", "")
	// Replace decimal comma with dot
	s = strings.ReplaceAll(s, ",", ".")
	// Remove whitespace
	s = strings.TrimSpace(s)

	var amount float64
	_, err := fmt.Sscanf(s, "%f", &amount)
	if err != nil {
		return 0, fmt.Errorf("Betrag '%s' ungueltig: %w", s, err)
	}

	// Convert to cents
	cents := int64(math.Round(amount * 100))
	return cents, nil
}

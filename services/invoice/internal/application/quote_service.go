package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	apperrors "github.com/jeckersberger/EquipFlow/pkg/common/errors"
	"github.com/jeckersberger/EquipFlow/services/invoice/internal/domain"
)

// ---------------------------------------------------------------------------
// Request DTOs
// ---------------------------------------------------------------------------

type CreateQuoteRequest struct {
	CustomerName     string     `json:"customer_name"`
	CustomerEmail    string     `json:"customer_email"`
	CustomerAddress  string     `json:"customer_address"`
	ProjectID        *uuid.UUID `json:"project_id,omitempty"`
	Subject          string     `json:"subject"`
	IntroText        string     `json:"intro_text"`
	OutroText        string     `json:"outro_text"`
	QuoteDate        string     `json:"quote_date"`
	ValidUntil       string     `json:"valid_until"`
	VatRate          *int64     `json:"vat_rate"`
	Kleinunternehmer bool       `json:"kleinunternehmer"`
	PaymentTermsDays *int       `json:"payment_terms_days"`
	DiscountPct      int64      `json:"discount_pct"`
	Notes            string     `json:"notes"`
}

type UpdateQuoteRequest struct {
	CustomerName     *string    `json:"customer_name"`
	CustomerEmail    *string    `json:"customer_email"`
	CustomerAddress  *string    `json:"customer_address"`
	ProjectID        *uuid.UUID `json:"project_id"`
	Subject          *string    `json:"subject"`
	IntroText        *string    `json:"intro_text"`
	OutroText        *string    `json:"outro_text"`
	QuoteDate        *string    `json:"quote_date"`
	ValidUntil       *string    `json:"valid_until"`
	VatRate          *int64     `json:"vat_rate"`
	PaymentTermsDays *int       `json:"payment_terms_days"`
	DiscountPct      *int64     `json:"discount_pct"`
	Notes            *string    `json:"notes"`
}

type CreateQuoteItemRequest struct {
	Description string `json:"description"`
	Quantity    int64  `json:"quantity"`
	Unit        string `json:"unit"`
	UnitPrice   int64  `json:"unit_price"`
	DiscountPct int64  `json:"discount_pct"`
	Position    int    `json:"position"`
}

type UpdateQuoteStatusRequest struct {
	Status string `json:"status"`
}

// ---------------------------------------------------------------------------
// Service
// ---------------------------------------------------------------------------

type QuoteService struct {
	quoteRepo domain.QuoteRepository
	itemRepo  domain.QuoteItemRepository
	seqRepo   domain.NumberSequenceRepository
	// invoiceRepo + invoiceItemRepo needed for ConvertToInvoice
	invoiceRepo     domain.InvoiceRepository
	invoiceItemRepo domain.InvoiceItemRepository
	logger          zerolog.Logger
}

func NewQuoteService(
	quoteRepo domain.QuoteRepository,
	itemRepo domain.QuoteItemRepository,
	seqRepo domain.NumberSequenceRepository,
	invoiceRepo domain.InvoiceRepository,
	invoiceItemRepo domain.InvoiceItemRepository,
	logger zerolog.Logger,
) *QuoteService {
	return &QuoteService{
		quoteRepo:       quoteRepo,
		itemRepo:        itemRepo,
		seqRepo:         seqRepo,
		invoiceRepo:     invoiceRepo,
		invoiceItemRepo: invoiceItemRepo,
		logger:          logger.With().Str("service", "quote").Logger(),
	}
}

// Create creates a new quote (Angebot) in draft status.
func (s *QuoteService) Create(ctx context.Context, tenantID uuid.UUID, req CreateQuoteRequest) (*domain.Quote, error) {
	if req.CustomerName == "" {
		return nil, fmt.Errorf("customer_name is required")
	}

	now := time.Now()
	year := now.Year()

	seqNum, err := s.seqRepo.NextNumber(ctx, tenantID, "AN", year)
	if err != nil {
		return nil, fmt.Errorf("generate quote number: %w", err)
	}
	quoteNumber := fmt.Sprintf("AN-%d-%05d", year, seqNum)

	vatRate := int64(1900)
	if req.VatRate != nil {
		vatRate = *req.VatRate
	}
	if req.Kleinunternehmer {
		vatRate = 0
	}

	quoteDate := now.Format("2006-01-02")
	if req.QuoteDate != "" {
		quoteDate = req.QuoteDate
	}

	validUntil := ""
	if req.ValidUntil != "" {
		validUntil = req.ValidUntil
	} else {
		t, parseErr := time.Parse("2006-01-02", quoteDate)
		if parseErr == nil {
			validUntil = t.AddDate(0, 0, 30).Format("2006-01-02")
		}
	}

	paymentTermsDays := 14
	if req.PaymentTermsDays != nil {
		paymentTermsDays = *req.PaymentTermsDays
	}

	introText := "Gerne unterbreiten wir Ihnen folgendes Angebot:"
	if req.IntroText != "" {
		introText = req.IntroText
	}

	outroText := "Wir freuen uns auf Ihre Rueckmeldung."
	if req.OutroText != "" {
		outroText = req.OutroText
	}

	quote := &domain.Quote{
		ID:               uuid.New(),
		TenantID:         tenantID,
		QuoteNumber:      quoteNumber,
		Status:           domain.QuoteStatusDraft,
		CustomerName:     req.CustomerName,
		CustomerEmail:    req.CustomerEmail,
		CustomerAddress:  req.CustomerAddress,
		ProjectID:        req.ProjectID,
		Subject:          req.Subject,
		IntroText:        introText,
		OutroText:        outroText,
		QuoteDate:        quoteDate,
		ValidUntil:       validUntil,
		VatRate:          vatRate,
		Kleinunternehmer: req.Kleinunternehmer,
		PaymentTermsDays: paymentTermsDays,
		DiscountPct:      req.DiscountPct,
		Notes:            req.Notes,
	}

	if err := s.quoteRepo.Create(ctx, quote); err != nil {
		return nil, fmt.Errorf("create quote: %w", err)
	}

	s.logger.Info().Str("quote_id", quote.ID.String()).Str("number", quoteNumber).Msg("quote created")
	return quote, nil
}

// GetByID returns a single quote by ID.
func (s *QuoteService) GetByID(ctx context.Context, id, tenantID uuid.UUID) (*domain.Quote, error) {
	return s.quoteRepo.GetByID(ctx, id, tenantID)
}

// List returns a paginated list of quotes for a tenant.
func (s *QuoteService) List(ctx context.Context, tenantID uuid.UUID, filter domain.QuoteFilter) ([]*domain.Quote, int64, error) {
	return s.quoteRepo.List(ctx, tenantID, filter)
}

// Update modifies a quote that is still in draft status.
func (s *QuoteService) Update(ctx context.Context, id, tenantID uuid.UUID, req UpdateQuoteRequest) (*domain.Quote, error) {
	existing, err := s.quoteRepo.GetByID(ctx, id, tenantID)
	if err != nil {
		return nil, err
	}
	if existing.Status != domain.QuoteStatusDraft {
		return nil, apperrors.Wrap(apperrors.ErrBadRequest, domain.ErrQuoteNotDraft.Error())
	}

	if req.CustomerName != nil {
		existing.CustomerName = *req.CustomerName
	}
	if req.CustomerEmail != nil {
		existing.CustomerEmail = *req.CustomerEmail
	}
	if req.CustomerAddress != nil {
		existing.CustomerAddress = *req.CustomerAddress
	}
	if req.ProjectID != nil {
		existing.ProjectID = req.ProjectID
	}
	if req.Subject != nil {
		existing.Subject = *req.Subject
	}
	if req.IntroText != nil {
		existing.IntroText = *req.IntroText
	}
	if req.OutroText != nil {
		existing.OutroText = *req.OutroText
	}
	if req.QuoteDate != nil {
		existing.QuoteDate = *req.QuoteDate
	}
	if req.ValidUntil != nil {
		existing.ValidUntil = *req.ValidUntil
	}
	if req.VatRate != nil {
		existing.VatRate = *req.VatRate
	}
	if req.PaymentTermsDays != nil {
		existing.PaymentTermsDays = *req.PaymentTermsDays
	}
	if req.DiscountPct != nil {
		existing.DiscountPct = *req.DiscountPct
	}
	if req.Notes != nil {
		existing.Notes = *req.Notes
	}

	if err := s.quoteRepo.Update(ctx, existing); err != nil {
		return nil, err
	}

	return existing, nil
}

// AddItem adds a line item to a quote in draft status.
func (s *QuoteService) AddItem(ctx context.Context, quoteID, tenantID uuid.UUID, req CreateQuoteItemRequest) (*domain.QuoteItem, error) {
	quote, err := s.quoteRepo.GetByID(ctx, quoteID, tenantID)
	if err != nil {
		return nil, err
	}
	if quote.Status != domain.QuoteStatusDraft {
		return nil, apperrors.Wrap(apperrors.ErrBadRequest, domain.ErrQuoteNotDraft.Error())
	}

	if req.Description == "" {
		return nil, fmt.Errorf("description is required")
	}
	if req.Quantity <= 0 {
		req.Quantity = 1
	}
	if req.Unit == "" {
		req.Unit = "Stueck"
	}

	item := &domain.QuoteItem{
		ID:          uuid.New(),
		TenantID:    tenantID,
		QuoteID:     quoteID,
		Description: req.Description,
		Quantity:    req.Quantity,
		Unit:        req.Unit,
		UnitPrice:   req.UnitPrice,
		DiscountPct: req.DiscountPct,
		Position:    req.Position,
	}

	if err := s.itemRepo.Create(ctx, item); err != nil {
		return nil, fmt.Errorf("add quote item: %w", err)
	}

	// Recalculate totals
	if err := s.recalculateTotals(ctx, quote); err != nil {
		s.logger.Warn().Err(err).Str("quote_id", quoteID.String()).Msg("failed to recalculate totals after adding item")
	}

	return item, nil
}

// ListItems returns all line items for a quote.
func (s *QuoteService) ListItems(ctx context.Context, quoteID, tenantID uuid.UUID) ([]*domain.QuoteItem, error) {
	return s.itemRepo.ListByQuote(ctx, quoteID, tenantID)
}

// RemoveItem deletes a line item from a quote.
func (s *QuoteService) RemoveItem(ctx context.Context, itemID, tenantID uuid.UUID) error {
	return s.itemRepo.Delete(ctx, itemID, tenantID)
}

// UpdateStatus changes the status of a quote.
func (s *QuoteService) UpdateStatus(ctx context.Context, id, tenantID uuid.UUID, status string) error {
	switch status {
	case domain.QuoteStatusDraft, domain.QuoteStatusSent, domain.QuoteStatusAccepted,
		domain.QuoteStatusDeclined, domain.QuoteStatusExpired, domain.QuoteStatusCancelled:
		// valid
	default:
		return fmt.Errorf("invalid quote status: %s", status)
	}

	return s.quoteRepo.UpdateStatus(ctx, id, tenantID, status)
}

// ConvertToInvoice creates a new Invoice from the quote data and marks the quote as accepted.
func (s *QuoteService) ConvertToInvoice(ctx context.Context, quoteID, tenantID uuid.UUID) (*domain.Invoice, error) {
	// 1. Load the quote
	quote, err := s.quoteRepo.GetByID(ctx, quoteID, tenantID)
	if err != nil {
		return nil, err
	}

	if quote.ConvertedInvoiceID != nil {
		return nil, apperrors.Wrap(apperrors.ErrConflict, domain.ErrQuoteAlreadyConverted.Error())
	}

	// 2. Load quote items
	quoteItems, err := s.itemRepo.ListByQuote(ctx, quoteID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("convert - load items: %w", err)
	}

	// 3. Generate invoice number
	now := time.Now()
	year := now.Year()
	seqNum, err := s.seqRepo.NextNumber(ctx, tenantID, "RE", year)
	if err != nil {
		return nil, fmt.Errorf("convert - generate invoice number: %w", err)
	}
	invoiceNumber := fmt.Sprintf("RE-%d-%05d", year, seqNum)

	// 4. Calculate totals from items
	var totalNet int64
	for _, item := range quoteItems {
		lineNet := item.Quantity * item.UnitPrice
		if item.DiscountPct > 0 {
			lineNet = lineNet * (10000 - item.DiscountPct*100) / 10000
		}
		totalNet += lineNet
	}

	// Apply quote-level discount
	if quote.DiscountPct > 0 {
		totalNet = totalNet * (10000 - quote.DiscountPct*100) / 10000
	}

	var totalVat int64
	if !quote.Kleinunternehmer && quote.VatRate > 0 {
		totalVat = totalNet * quote.VatRate / 10000
	}
	totalGross := totalNet + totalVat

	// Calculate due date from payment terms
	invoiceDate := now.Format("2006-01-02")
	dueDate := now.AddDate(0, 0, quote.PaymentTermsDays).Format("2006-01-02")

	// 5. Create the invoice
	invoice := &domain.Invoice{
		ID:               uuid.New(),
		TenantID:         tenantID,
		InvoiceNumber:    invoiceNumber,
		InvoiceType:      domain.InvoiceTypeInvoice,
		Status:           domain.StatusDraft,
		CustomerName:     quote.CustomerName,
		CustomerEmail:    quote.CustomerEmail,
		CustomerAddress:  quote.CustomerAddress,
		InvoiceDate:      invoiceDate,
		DueDate:          dueDate,
		VatRate:          quote.VatRate,
		Kleinunternehmer: quote.Kleinunternehmer,
		TotalNet:         totalNet,
		TotalVat:         totalVat,
		TotalGross:       totalGross,
		Notes:            quote.Notes,
	}

	if err := s.invoiceRepo.Create(ctx, invoice); err != nil {
		return nil, fmt.Errorf("convert - create invoice: %w", err)
	}

	// 6. Create invoice items from quote items
	for _, qi := range quoteItems {
		invItem := &domain.InvoiceItem{
			ID:          uuid.New(),
			TenantID:    tenantID,
			InvoiceID:   invoice.ID,
			Description: qi.Description,
			Quantity:    qi.Quantity,
			Unit:        qi.Unit,
			UnitPrice:   qi.UnitPrice,
			Position:    qi.Position,
		}
		if err := s.invoiceItemRepo.Create(ctx, invItem); err != nil {
			return nil, fmt.Errorf("convert - create invoice item: %w", err)
		}
	}

	// 7. Update quote: mark as accepted and store converted_invoice_id
	quote.Status = domain.QuoteStatusAccepted
	quote.ConvertedInvoiceID = &invoice.ID
	if err := s.quoteRepo.Update(ctx, quote); err != nil {
		return nil, fmt.Errorf("convert - update quote: %w", err)
	}

	s.logger.Info().
		Str("quote_id", quoteID.String()).
		Str("invoice_id", invoice.ID.String()).
		Str("invoice_number", invoiceNumber).
		Msg("quote converted to invoice")

	return invoice, nil
}

// recalculateTotals recalculates net/vat/gross from current items and persists them.
func (s *QuoteService) recalculateTotals(ctx context.Context, quote *domain.Quote) error {
	items, err := s.itemRepo.ListByQuote(ctx, quote.ID, quote.TenantID)
	if err != nil {
		return err
	}

	var totalNet int64
	for _, item := range items {
		lineNet := item.Quantity * item.UnitPrice
		if item.DiscountPct > 0 {
			lineNet = lineNet * (10000 - item.DiscountPct*100) / 10000
		}
		totalNet += lineNet
	}

	if quote.DiscountPct > 0 {
		totalNet = totalNet * (10000 - quote.DiscountPct*100) / 10000
	}

	var totalVat int64
	if !quote.Kleinunternehmer && quote.VatRate > 0 {
		totalVat = totalNet * quote.VatRate / 10000
	}

	quote.TotalNet = totalNet
	quote.TotalVat = totalVat
	quote.TotalGross = totalNet + totalVat

	return s.quoteRepo.Update(ctx, quote)
}

package application

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
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

type CreateInvoiceRequest struct {
	CustomerName     string            `json:"customer_name"`
	CustomerEmail    string            `json:"customer_email"`
	CustomerAddress  string            `json:"customer_address"`
	InvoiceDate      string            `json:"invoice_date"`
	DueDate          string            `json:"due_date"`
	InvoiceType      string            `json:"invoice_type"`
	VatRate          *int64            `json:"vat_rate"`
	Kleinunternehmer bool              `json:"kleinunternehmer"`
	Notes            string            `json:"notes"`
	Items            []AddItemRequest  `json:"items,omitempty"`
}

type UpdateInvoiceRequest struct {
	CustomerName    *string `json:"customer_name"`
	CustomerEmail   *string `json:"customer_email"`
	CustomerAddress *string `json:"customer_address"`
	InvoiceDate     *string `json:"invoice_date"`
	DueDate         *string `json:"due_date"`
	Notes           *string `json:"notes"`
	VatRate         *int64  `json:"vat_rate"`
}

type AddItemRequest struct {
	Description string `json:"description"`
	Quantity    int64  `json:"quantity"`
	Unit        string `json:"unit"`
	UnitPrice   int64  `json:"unit_price"`
	VatRate     int64  `json:"vat_rate"`
	Position    int    `json:"position"`
}

type AddPaymentRequest struct {
	Amount        int64  `json:"amount"`
	PaymentDate   string `json:"payment_date"`
	PaymentMethod string `json:"payment_method"`
	Reference     string `json:"reference"`
}

// ---------------------------------------------------------------------------
// Service
// ---------------------------------------------------------------------------

type InvoiceService struct {
	invoiceRepo domain.InvoiceRepository
	itemRepo    domain.InvoiceItemRepository
	seqRepo     domain.NumberSequenceRepository
	paymentRepo domain.PaymentRepository
	logger      zerolog.Logger
}

func NewInvoiceService(
	invoiceRepo domain.InvoiceRepository,
	itemRepo domain.InvoiceItemRepository,
	seqRepo domain.NumberSequenceRepository,
	paymentRepo domain.PaymentRepository,
	logger zerolog.Logger,
) *InvoiceService {
	return &InvoiceService{
		invoiceRepo: invoiceRepo,
		itemRepo:    itemRepo,
		seqRepo:     seqRepo,
		paymentRepo: paymentRepo,
		logger:      logger.With().Str("service", "invoice").Logger(),
	}
}

func (s *InvoiceService) Create(ctx context.Context, tenantID uuid.UUID, req CreateInvoiceRequest) (*domain.Invoice, error) {
	if req.CustomerName == "" {
		return nil, fmt.Errorf("customer_name is required")
	}

	invType := domain.InvoiceTypeInvoice
	if req.InvoiceType != "" {
		invType = req.InvoiceType
	}

	prefix := "RE"
	switch invType {
	case domain.InvoiceTypeInvoice:
		prefix = "RE"
	case domain.InvoiceTypeCreditNote:
		prefix = "GS"
	case domain.InvoiceTypeReversal:
		prefix = "ST"
	case domain.InvoiceTypeProforma:
		prefix = "PF"
	case domain.InvoiceTypeAdvance:
		prefix = "AR"
	case domain.InvoiceTypePartial:
		prefix = "TR"
	default:
		return nil, fmt.Errorf("invalid invoice type")
	}

	now := time.Now()
	year := now.Year()

	seqNum, err := s.seqRepo.NextNumber(ctx, tenantID, prefix, year)
	if err != nil {
		return nil, fmt.Errorf("generate invoice number: %w", err)
	}
	invoiceNumber := fmt.Sprintf("%s-%d-%05d", prefix, year, seqNum)

	vatRate := int64(1900)
	if req.VatRate != nil {
		vatRate = *req.VatRate
	}
	if req.Kleinunternehmer {
		vatRate = 0
	}

	invoiceDate := now.Format("2006-01-02")
	if req.InvoiceDate != "" {
		invoiceDate = req.InvoiceDate
	}

	dueDate := ""
	if req.DueDate != "" {
		dueDate = req.DueDate
	} else {
		t, parseErr := time.Parse("2006-01-02", invoiceDate)
		if parseErr == nil {
			dueDate = t.AddDate(0, 0, 30).Format("2006-01-02")
		}
	}

	invoice := &domain.Invoice{
		ID:               uuid.New(),
		TenantID:         tenantID,
		InvoiceNumber:    invoiceNumber,
		InvoiceType:      invType,
		Status:           domain.StatusDraft,
		CustomerName:     req.CustomerName,
		CustomerEmail:    req.CustomerEmail,
		CustomerAddress:  req.CustomerAddress,
		InvoiceDate:      invoiceDate,
		DueDate:          dueDate,
		VatRate:          vatRate,
		Kleinunternehmer: req.Kleinunternehmer,
		Notes:            req.Notes,
	}

	if err := s.invoiceRepo.Create(ctx, invoice); err != nil {
		return nil, fmt.Errorf("create invoice: %w", err)
	}

	// Process inline items if provided
	if len(req.Items) > 0 {
		var totalNet int64
		for i, itemReq := range req.Items {
			item := &domain.InvoiceItem{
				ID:          uuid.New(),
				TenantID:    tenantID,
				InvoiceID:   invoice.ID,
				Description: itemReq.Description,
				Quantity:    itemReq.Quantity,
				Unit:        itemReq.Unit,
				UnitPrice:   itemReq.UnitPrice,
				Position:    itemReq.Position,
			}
			if item.Position == 0 {
				item.Position = i + 1
			}
			if item.Quantity == 0 {
				item.Quantity = 1
			}
			if item.Unit == "" {
				item.Unit = "Stueck"
			}
			if err := s.itemRepo.Create(ctx, item); err != nil {
				return nil, fmt.Errorf("create invoice item %d: %w", i+1, err)
			}
			totalNet += item.Quantity * item.UnitPrice
		}

		var totalVat int64
		if !invoice.Kleinunternehmer && invoice.VatRate > 0 {
			totalVat = totalNet * invoice.VatRate / 10000
		}
		invoice.TotalNet = totalNet
		invoice.TotalVat = totalVat
		invoice.TotalGross = totalNet + totalVat

		if err := s.invoiceRepo.UpdateTotals(ctx, invoice.ID, tenantID, totalNet, totalVat, totalNet+totalVat); err != nil {
			s.logger.Warn().Err(err).Msg("failed to update invoice totals after inline items")
		}
	}

	s.logger.Info().Str("invoice_id", invoice.ID.String()).Str("number", invoiceNumber).Msg("invoice created")
	return invoice, nil
}

func (s *InvoiceService) GetByID(ctx context.Context, id, tenantID uuid.UUID) (*domain.Invoice, error) {
	return s.invoiceRepo.GetByID(ctx, id, tenantID)
}

func (s *InvoiceService) List(ctx context.Context, tenantID uuid.UUID, filter domain.InvoiceFilter) ([]*domain.Invoice, int64, error) {
	return s.invoiceRepo.List(ctx, tenantID, filter)
}

func (s *InvoiceService) Update(ctx context.Context, id, tenantID uuid.UUID, req UpdateInvoiceRequest) (*domain.Invoice, error) {
	existing, err := s.invoiceRepo.GetByID(ctx, id, tenantID)
	if err != nil {
		return nil, err
	}
	if existing.Status != domain.StatusDraft {
		return nil, apperrors.Wrap(apperrors.ErrBadRequest, domain.ErrInvoiceNotDraft.Error())
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
	if req.Notes != nil {
		existing.Notes = *req.Notes
	}
	if req.VatRate != nil {
		existing.VatRate = *req.VatRate
	}
	if req.InvoiceDate != nil {
		existing.InvoiceDate = *req.InvoiceDate
	}
	if req.DueDate != nil {
		existing.DueDate = *req.DueDate
	}

	if err := s.invoiceRepo.Update(ctx, existing); err != nil {
		return nil, err
	}

	return existing, nil
}

func (s *InvoiceService) Delete(ctx context.Context, id, tenantID uuid.UUID) error {
	return s.invoiceRepo.Delete(ctx, id, tenantID)
}

func (s *InvoiceService) Finalize(ctx context.Context, id, tenantID uuid.UUID) (*domain.Invoice, error) {
	inv, err := s.invoiceRepo.GetByID(ctx, id, tenantID)
	if err != nil {
		return nil, err
	}
	if inv.Status != domain.StatusDraft {
		return nil, apperrors.Wrap(apperrors.ErrBadRequest, domain.ErrInvoiceNotDraft.Error())
	}

	items, err := s.itemRepo.ListByInvoice(ctx, inv.ID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("finalize - items: %w", err)
	}

	var totalNet int64
	for _, item := range items {
		totalNet += item.Quantity * item.UnitPrice
	}

	var totalVat int64
	if !inv.Kleinunternehmer && inv.VatRate > 0 {
		totalVat = totalNet * inv.VatRate / 10000
	}
	totalGross := totalNet + totalVat

	prevHash, err := s.invoiceRepo.GetLastHash(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("finalize - prev hash: %w", err)
	}

	hashInput := fmt.Sprintf("%s|%d|%d|%d|%s", inv.InvoiceNumber, totalNet, totalVat, totalGross, prevHash)
	h := sha256.Sum256([]byte(hashInput))

	now := time.Now()
	inv.TotalNet = totalNet
	inv.TotalVat = totalVat
	inv.TotalGross = totalGross
	inv.Hash = hex.EncodeToString(h[:])
	inv.PreviousHash = prevHash
	inv.Status = domain.StatusFinalized
	inv.FinalizedAt = &now

	if err := s.invoiceRepo.Update(ctx, inv); err != nil {
		return nil, fmt.Errorf("finalize - update: %w", err)
	}

	s.logger.Info().Str("invoice_id", id.String()).Str("hash", inv.Hash).Msg("invoice finalized")
	return inv, nil
}

func (s *InvoiceService) Search(ctx context.Context, tenantID uuid.UUID, query string, page, perPage int) ([]*domain.Invoice, int64, error) {
	return s.invoiceRepo.Search(ctx, tenantID, query, page, perPage)
}

// --- Items ---

func (s *InvoiceService) AddItem(ctx context.Context, invoiceID, tenantID uuid.UUID, req AddItemRequest) (*domain.InvoiceItem, error) {
	inv, err := s.invoiceRepo.GetByID(ctx, invoiceID, tenantID)
	if err != nil {
		return nil, err
	}
	if inv.Status != domain.StatusDraft {
		return nil, apperrors.Wrap(apperrors.ErrBadRequest, domain.ErrInvoiceNotDraft.Error())
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

	vatRate := req.VatRate
	if vatRate == 0 {
		vatRate = inv.VatRate
	}

	item := &domain.InvoiceItem{
		ID:          uuid.New(),
		TenantID:    tenantID,
		InvoiceID:   invoiceID,
		Description: req.Description,
		Quantity:    req.Quantity,
		Unit:        req.Unit,
		UnitPrice:   req.UnitPrice,
		VatRate:     vatRate,
		Position:    req.Position,
	}

	if err := s.itemRepo.Create(ctx, item); err != nil {
		return nil, fmt.Errorf("add item: %w", err)
	}

	return item, nil
}

func (s *InvoiceService) ListItems(ctx context.Context, invoiceID, tenantID uuid.UUID) ([]*domain.InvoiceItem, error) {
	return s.itemRepo.ListByInvoice(ctx, invoiceID, tenantID)
}

func (s *InvoiceService) RemoveItem(ctx context.Context, itemID, tenantID uuid.UUID) error {
	return s.itemRepo.Delete(ctx, itemID, tenantID)
}

// --- Payments ---

func (s *InvoiceService) AddPayment(ctx context.Context, invoiceID, tenantID uuid.UUID, req AddPaymentRequest) (*domain.Payment, error) {
	inv, err := s.invoiceRepo.GetByID(ctx, invoiceID, tenantID)
	if err != nil {
		return nil, err
	}

	switch inv.Status {
	case domain.StatusFinalized, domain.StatusSent, domain.StatusPartialPaid:
		// OK
	default:
		return nil, fmt.Errorf("invoice cannot accept payments")
	}

	if req.Amount <= 0 {
		return nil, fmt.Errorf("amount must be positive")
	}

	remaining := inv.TotalGross - inv.AmountPaid
	if req.Amount > remaining {
		return nil, fmt.Errorf("payment exceeds remaining balance")
	}

	paymentMethod := "bank_transfer"
	if req.PaymentMethod != "" {
		paymentMethod = req.PaymentMethod
	}
	paymentDate := time.Now().Format("2006-01-02")
	if req.PaymentDate != "" {
		paymentDate = req.PaymentDate
	}

	payment := &domain.Payment{
		ID:            uuid.New(),
		TenantID:      tenantID,
		InvoiceID:     invoiceID,
		Amount:        req.Amount,
		PaymentDate:   paymentDate,
		PaymentMethod: paymentMethod,
		Reference:     req.Reference,
	}

	if err := s.paymentRepo.Create(ctx, payment); err != nil {
		return nil, fmt.Errorf("add payment: %w", err)
	}

	newAmountPaid := inv.AmountPaid + req.Amount
	if err := s.invoiceRepo.UpdateAmountPaid(ctx, invoiceID, tenantID, newAmountPaid); err != nil {
		return nil, fmt.Errorf("update amount_paid: %w", err)
	}

	newStatus := domain.StatusPartialPaid
	if newAmountPaid >= inv.TotalGross {
		newStatus = domain.StatusPaid
	}
	if err := s.invoiceRepo.UpdateStatus(ctx, invoiceID, tenantID, newStatus); err != nil {
		return nil, fmt.Errorf("update status: %w", err)
	}

	s.logger.Info().Str("invoice_id", invoiceID.String()).Int64("amount", req.Amount).Msg("payment recorded")
	return payment, nil
}

func (s *InvoiceService) ListPayments(ctx context.Context, invoiceID, tenantID uuid.UUID) ([]*domain.Payment, error) {
	return s.paymentRepo.ListByInvoice(ctx, invoiceID, tenantID)
}

// --- Partial Invoices (Teilrechnungen) ---

// CreatePartialInvoiceRequest holds the data for creating a partial invoice.
type CreatePartialInvoiceRequest struct {
	Percentage int `json:"percentage"`
}

// CreatePartialInvoice creates a new partial invoice (Teilrechnung) based on an existing invoice.
// The partial invoice takes a percentage of the original total.
func (s *InvoiceService) CreatePartialInvoice(ctx context.Context, originalInvoiceID, tenantID uuid.UUID, percentage int) (*domain.Invoice, error) {
	if percentage < 1 || percentage > 100 {
		return nil, domain.ErrInvalidPercentage
	}

	// Load the original invoice
	original, err := s.invoiceRepo.GetByID(ctx, originalInvoiceID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("Teilrechnung: Original-Rechnung nicht gefunden: %w", err)
	}

	// Generate partial invoice number
	now := time.Now()
	year := now.Year()

	seqNum, err := s.seqRepo.NextNumber(ctx, tenantID, "TR", year)
	if err != nil {
		return nil, fmt.Errorf("Teilrechnung: Rechnungsnummer generieren: %w", err)
	}
	invoiceNumber := fmt.Sprintf("TR-%d-%05d", year, seqNum)

	// Calculate amounts as percentage of original
	totalNet := original.TotalNet * int64(percentage) / 100
	var totalVat int64
	if !original.Kleinunternehmer && original.VatRate > 0 {
		totalVat = totalNet * original.VatRate / 10000
	}
	totalGross := totalNet + totalVat

	invoiceDate := now.Format("2006-01-02")
	dueDate := now.AddDate(0, 0, 30).Format("2006-01-02")

	notes := fmt.Sprintf("Teilrechnung (%d%%) zu %s", percentage, original.InvoiceNumber)
	if original.Notes != "" {
		notes = notes + "\n" + original.Notes
	}

	partial := &domain.Invoice{
		ID:               uuid.New(),
		TenantID:         tenantID,
		InvoiceNumber:    invoiceNumber,
		InvoiceType:      domain.InvoiceTypePartial,
		Status:           domain.StatusDraft,
		CustomerName:     original.CustomerName,
		CustomerEmail:    original.CustomerEmail,
		CustomerAddress:  original.CustomerAddress,
		InvoiceDate:      invoiceDate,
		DueDate:          dueDate,
		VatRate:          original.VatRate,
		Kleinunternehmer: original.Kleinunternehmer,
		TotalNet:         totalNet,
		TotalVat:         totalVat,
		TotalGross:       totalGross,
		Notes:            notes,
	}

	if err := s.invoiceRepo.Create(ctx, partial); err != nil {
		return nil, fmt.Errorf("Teilrechnung erstellen: %w", err)
	}

	// Copy items from original, scaled by percentage
	items, err := s.itemRepo.ListByInvoice(ctx, originalInvoiceID, tenantID)
	if err != nil {
		s.logger.Warn().Err(err).Msg("Teilrechnung: Positionen konnten nicht kopiert werden")
	} else {
		for _, item := range items {
			partialItem := &domain.InvoiceItem{
				ID:          uuid.New(),
				TenantID:    tenantID,
				InvoiceID:   partial.ID,
				Description: fmt.Sprintf("%s (%d%%)", item.Description, percentage),
				Quantity:    item.Quantity,
				Unit:        item.Unit,
				UnitPrice:   item.UnitPrice * int64(percentage) / 100,
				Position:    item.Position,
			}
			if err := s.itemRepo.Create(ctx, partialItem); err != nil {
				s.logger.Warn().Err(err).Msg("Teilrechnung: Position konnte nicht kopiert werden")
			}
		}
	}

	s.logger.Info().
		Str("original_id", originalInvoiceID.String()).
		Str("partial_id", partial.ID.String()).
		Str("number", invoiceNumber).
		Int("percentage", percentage).
		Msg("partial invoice created")

	return partial, nil
}

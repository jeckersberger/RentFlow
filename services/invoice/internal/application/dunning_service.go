package application

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/services/invoice/internal/domain"
)

// ---------------------------------------------------------------------------
// Request DTOs
// ---------------------------------------------------------------------------

// CreateReminderRequest holds the data needed to create a dunning entry.
type CreateReminderRequest struct {
	InvoiceID uuid.UUID `json:"invoice_id"`
	Level     string    `json:"level"`
	Notes     string    `json:"notes"`
}

// UpdateDunningConfigRequest holds the data needed to update dunning configuration.
type UpdateDunningConfigRequest struct {
	ReminderDays *int   `json:"reminder_days"`
	Dunning1Days *int   `json:"dunning1_days"`
	Dunning2Days *int   `json:"dunning2_days"`
	ReminderFee  *int64 `json:"reminder_fee"`
	Dunning1Fee  *int64 `json:"dunning1_fee"`
	Dunning2Fee  *int64 `json:"dunning2_fee"`
	AutoSend     *bool  `json:"auto_send"`
}

// ---------------------------------------------------------------------------
// Service
// ---------------------------------------------------------------------------

// DunningService implements the application-level use cases for dunning (Mahnwesen).
type DunningService struct {
	dunningRepo domain.DunningRepository
	invoiceRepo domain.InvoiceRepository
	logger      zerolog.Logger
}

// NewDunningService constructs a new DunningService.
func NewDunningService(
	dunningRepo domain.DunningRepository,
	invoiceRepo domain.InvoiceRepository,
	logger zerolog.Logger,
) *DunningService {
	return &DunningService{
		dunningRepo: dunningRepo,
		invoiceRepo: invoiceRepo,
		logger:      logger.With().Str("service", "dunning").Logger(),
	}
}

// GetOverdueInvoices finds sent/partial_paid invoices past their due date and
// calculates the suggested dunning level based on tenant config.
func (s *DunningService) GetOverdueInvoices(ctx context.Context, tenantID uuid.UUID) ([]*domain.OverdueInvoice, error) {
	invoices, err := s.dunningRepo.ListOverdue(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("get overdue invoices: %w", err)
	}

	cfg, err := s.dunningRepo.GetConfig(ctx, tenantID)
	if err != nil {
		// Use defaults if no config exists
		cfg = s.defaultConfig(tenantID)
	}

	now := time.Now()
	var result []*domain.OverdueInvoice

	for _, inv := range invoices {
		dueDate, parseErr := time.Parse("2006-01-02", inv.DueDate)
		if parseErr != nil {
			continue
		}

		daysOverdue := int(math.Floor(now.Sub(dueDate).Hours() / 24))
		if daysOverdue < 1 {
			continue
		}

		level, fee := s.determineDunningLevel(daysOverdue, cfg)

		result = append(result, &domain.OverdueInvoice{
			Invoice:        inv,
			DaysOverdue:    daysOverdue,
			SuggestedLevel: level,
			SuggestedFee:   fee,
		})
	}

	return result, nil
}

// CreateReminder creates a dunning entry for a given invoice.
func (s *DunningService) CreateReminder(ctx context.Context, tenantID uuid.UUID, req CreateReminderRequest) (*domain.DunningEntry, error) {
	// Validate invoice exists and belongs to tenant
	inv, err := s.invoiceRepo.GetByID(ctx, req.InvoiceID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("create reminder - invoice: %w", err)
	}

	// Invoice must be in a payable state
	switch inv.Status {
	case domain.StatusSent, domain.StatusPartialPaid, domain.StatusOverdue:
		// OK
	default:
		return nil, fmt.Errorf("invoice status '%s' does not allow dunning", inv.Status)
	}

	// Validate level
	level := req.Level
	if level == "" {
		level = domain.DunningLevelReminder
	}
	if !isValidDunningLevel(level) {
		return nil, fmt.Errorf("invalid dunning level: %s", level)
	}

	// Get fee from config
	cfg, err := s.dunningRepo.GetConfig(ctx, tenantID)
	if err != nil {
		cfg = s.defaultConfig(tenantID)
	}
	fee := s.feeForLevel(level, cfg)

	now := time.Now()
	entry := &domain.DunningEntry{
		ID:        uuid.New(),
		TenantID:  tenantID,
		InvoiceID: req.InvoiceID,
		Level:     level,
		FeeCents:  fee,
		SentAt:    &now,
		Status:    domain.DunningStatusPending,
		Notes:     req.Notes,
	}

	if err := s.dunningRepo.Create(ctx, entry); err != nil {
		return nil, fmt.Errorf("create reminder: %w", err)
	}

	// Update invoice status to overdue
	if inv.Status == domain.StatusSent || inv.Status == domain.StatusPartialPaid {
		_ = s.invoiceRepo.UpdateStatus(ctx, inv.ID, tenantID, domain.StatusOverdue)
	}

	s.logger.Info().
		Str("invoice_id", req.InvoiceID.String()).
		Str("level", level).
		Int64("fee_cents", fee).
		Msg("dunning entry created")

	return entry, nil
}

// ListByInvoice returns all dunning entries for a given invoice.
func (s *DunningService) ListByInvoice(ctx context.Context, invoiceID uuid.UUID, tenantID uuid.UUID) ([]*domain.DunningEntry, error) {
	return s.dunningRepo.ListByInvoice(ctx, invoiceID, tenantID)
}

// GetConfig retrieves the dunning configuration for a tenant.
func (s *DunningService) GetConfig(ctx context.Context, tenantID uuid.UUID) (*domain.DunningConfig, error) {
	cfg, err := s.dunningRepo.GetConfig(ctx, tenantID)
	if err != nil {
		// Return defaults if not found
		return s.defaultConfig(tenantID), nil
	}
	return cfg, nil
}

// UpdateConfig updates the dunning configuration for a tenant.
func (s *DunningService) UpdateConfig(ctx context.Context, tenantID uuid.UUID, req UpdateDunningConfigRequest) (*domain.DunningConfig, error) {
	existing, err := s.dunningRepo.GetConfig(ctx, tenantID)
	if err != nil {
		// Create new config with defaults
		existing = s.defaultConfig(tenantID)
	}

	if req.ReminderDays != nil {
		existing.ReminderDays = *req.ReminderDays
	}
	if req.Dunning1Days != nil {
		existing.Dunning1Days = *req.Dunning1Days
	}
	if req.Dunning2Days != nil {
		existing.Dunning2Days = *req.Dunning2Days
	}
	if req.ReminderFee != nil {
		existing.ReminderFee = *req.ReminderFee
	}
	if req.Dunning1Fee != nil {
		existing.Dunning1Fee = *req.Dunning1Fee
	}
	if req.Dunning2Fee != nil {
		existing.Dunning2Fee = *req.Dunning2Fee
	}
	if req.AutoSend != nil {
		existing.AutoSend = *req.AutoSend
	}

	if err := s.dunningRepo.UpsertConfig(ctx, existing); err != nil {
		return nil, fmt.Errorf("update dunning config: %w", err)
	}

	s.logger.Info().Str("tenant_id", tenantID.String()).Msg("dunning config updated")
	return existing, nil
}

// RunDunningCheck auto-detects overdue invoices and determines the correct
// dunning level based on days overdue and config. Returns newly created entries.
func (s *DunningService) RunDunningCheck(ctx context.Context, tenantID uuid.UUID) ([]*domain.DunningEntry, error) {
	overdue, err := s.GetOverdueInvoices(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("run dunning check: %w", err)
	}

	var created []*domain.DunningEntry
	for _, o := range overdue {
		// Check if there is already a dunning entry at this level
		existing, listErr := s.dunningRepo.ListByInvoice(ctx, o.Invoice.ID, tenantID)
		if listErr != nil {
			s.logger.Error().Err(listErr).
				Str("invoice_id", o.Invoice.ID.String()).
				Msg("failed to list existing dunning entries")
			continue
		}

		if s.hasEntryAtLevel(existing, o.SuggestedLevel) {
			continue
		}

		entry, createErr := s.CreateReminder(ctx, tenantID, CreateReminderRequest{
			InvoiceID: o.Invoice.ID,
			Level:     o.SuggestedLevel,
			Notes:     fmt.Sprintf("Automatisch erstellt - %d Tage ueberfaellig", o.DaysOverdue),
		})
		if createErr != nil {
			s.logger.Error().Err(createErr).
				Str("invoice_id", o.Invoice.ID.String()).
				Msg("failed to create dunning entry")
			continue
		}

		created = append(created, entry)
	}

	s.logger.Info().
		Str("tenant_id", tenantID.String()).
		Int("new_entries", len(created)).
		Msg("dunning check completed")

	return created, nil
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

func (s *DunningService) defaultConfig(tenantID uuid.UUID) *domain.DunningConfig {
	return &domain.DunningConfig{
		ID:           uuid.New(),
		TenantID:     tenantID,
		ReminderDays: 7,
		Dunning1Days: 14,
		Dunning2Days: 28,
		ReminderFee:  0,
		Dunning1Fee:  500,
		Dunning2Fee:  1000,
		AutoSend:     false,
	}
}

func (s *DunningService) determineDunningLevel(daysOverdue int, cfg *domain.DunningConfig) (string, int64) {
	if daysOverdue >= cfg.Dunning2Days {
		return domain.DunningLevelDunning2, cfg.Dunning2Fee
	}
	if daysOverdue >= cfg.Dunning1Days {
		return domain.DunningLevelDunning1, cfg.Dunning1Fee
	}
	if daysOverdue >= cfg.ReminderDays {
		return domain.DunningLevelReminder, cfg.ReminderFee
	}
	return domain.DunningLevelReminder, cfg.ReminderFee
}

func (s *DunningService) feeForLevel(level string, cfg *domain.DunningConfig) int64 {
	switch level {
	case domain.DunningLevelReminder:
		return cfg.ReminderFee
	case domain.DunningLevelDunning1:
		return cfg.Dunning1Fee
	case domain.DunningLevelDunning2:
		return cfg.Dunning2Fee
	case domain.DunningLevelDunning3:
		return cfg.Dunning2Fee // Dunning 3 uses same fee as Dunning 2
	default:
		return 0
	}
}

func (s *DunningService) hasEntryAtLevel(entries []*domain.DunningEntry, level string) bool {
	for _, e := range entries {
		if e.Level == level && e.Status != domain.DunningStatusCancelled {
			return true
		}
	}
	return false
}

func isValidDunningLevel(level string) bool {
	switch level {
	case domain.DunningLevelReminder,
		domain.DunningLevelDunning1,
		domain.DunningLevelDunning2,
		domain.DunningLevelDunning3:
		return true
	default:
		return false
	}
}

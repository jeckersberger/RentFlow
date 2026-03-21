package application

import (
	"context"
	"fmt"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/invoice-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/invoice-service/internal/ports"
)

type DunningService struct {
	dunningRepo ports.DunningRepository
	invoiceRepo ports.InvoiceRepository
	logger      *logger.Logger
}

func NewDunningService(
	dunningRepo ports.DunningRepository,
	invoiceRepo ports.InvoiceRepository,
	logger *logger.Logger,
) *DunningService {
	return &DunningService{
		dunningRepo: dunningRepo,
		invoiceRepo: invoiceRepo,
		logger:      logger,
	}
}

// CreateReminder creates a dunning reminder for an overdue invoice
func (s *DunningService) CreateReminder(ctx context.Context, cmd CreateDunningCommand) (*DunningDTO, error) {
	if cmd.TenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	invoice, err := s.invoiceRepo.GetByID(ctx, cmd.TenantID, cmd.InvoiceID)
	if err != nil {
		return nil, domain.NewDomainError("NOT_FOUND", "invoice not found", err)
	}

	if invoice.Status == domain.InvoicePaid || invoice.Status == domain.InvoiceCancelled {
		return nil, domain.NewDomainError("INVALID_STATUS", "cannot create dunning for paid or cancelled invoice", nil)
	}

	// Check if past due date
	if time.Now().Before(invoice.DueDate) {
		return nil, domain.NewDomainError("NOT_OVERDUE", "invoice is not yet overdue", nil)
	}

	// Validate level
	if cmd.Level < 1 || cmd.Level > 3 {
		cmd.Level = domain.DunningLevelPaymentReminder
	}

	// Get fee
	calc := domain.NewDunningCalculator()
	fee := calc.GetFeeForLevel(cmd.Level)
	if cmd.CustomFee != nil {
		fee = *cmd.CustomFee
	}

	// Create dunning entry
	now := time.Now()
	dunningID := fmt.Sprintf("dun_%d_%d", now.Unix(), cmd.Level)

	dunning := domain.NewDunningEntry(
		dunningID,
		cmd.TenantID,
		cmd.InvoiceID,
		invoice.InvoiceNumber,
		cmd.Level,
		cmd.DaysToAdd,
		fee,
	)

	if err := dunning.Validate(); err != nil {
		return nil, domain.NewDomainError("VALIDATION_ERROR", err.Error(), err)
	}

	if err := s.dunningRepo.Create(ctx, dunning); err != nil {
		return nil, domain.NewDomainError("CREATE_ERROR", "failed to create dunning reminder", err)
	}

	s.logger.Info("Dunning reminder created", "id", dunning.ID, "invoice_number", invoice.InvoiceNumber, "level", cmd.Level)
	return DunningToDTO(dunning), nil
}

// SendReminder sends a dunning reminder
func (s *DunningService) SendReminder(ctx context.Context, cmd SendDunningCommand) (*DunningDTO, error) {
	if cmd.TenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	dunning, err := s.dunningRepo.GetByID(ctx, cmd.TenantID, cmd.ID)
	if err != nil {
		return nil, domain.NewDomainError("NOT_FOUND", "dunning entry not found", err)
	}

	if err := dunning.Send(cmd.Email); err != nil {
		return nil, domain.NewDomainError("SEND_ERROR", "failed to send dunning reminder", err)
	}

	if err := s.dunningRepo.Update(ctx, dunning); err != nil {
		return nil, domain.NewDomainError("UPDATE_ERROR", "failed to update dunning entry", err)
	}

	s.logger.Info("Dunning reminder sent", "id", cmd.ID, "email", cmd.Email, "level", dunning.Level)
	return DunningToDTO(dunning), nil
}

// GetOverdueInvoices returns all overdue invoices for dunning processing
func (s *DunningService) GetOverdueInvoices(ctx context.Context, tenantID string) ([]*InvoiceDTO, error) {
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

// GetDunningHistory returns all dunning entries for an invoice
func (s *DunningService) GetDunningHistory(ctx context.Context, tenantID, invoiceID string) ([]*DunningDTO, error) {
	if tenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	entries, err := s.dunningRepo.List(ctx, tenantID, invoiceID)
	if err != nil {
		return nil, domain.NewDomainError("QUERY_ERROR", "failed to get dunning history", err)
	}

	dtos := make([]*DunningDTO, len(entries))
	for i, entry := range entries {
		dtos[i] = DunningToDTO(entry)
	}

	return dtos, nil
}

// ProcessAutoDunning automatically creates and sends dunning notices for overdue invoices
// This would typically be called by a scheduled job
func (s *DunningService) ProcessAutoDunning(ctx context.Context, tenantID string) ([]string, error) {
	if tenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	overdue, err := s.invoiceRepo.GetOverdueInvoices(ctx, tenantID)
	if err != nil {
		return nil, domain.NewDomainError("QUERY_ERROR", "failed to get overdue invoices", err)
	}

	var dunningIDs []string
	calc := domain.NewDunningCalculator()

	for _, invoice := range overdue {
		daysPastDue := int(time.Since(invoice.DueDate).Hours() / 24)

		// Determine which dunning level to send
		level, _ := calc.CalculateNextDunningLevel(daysPastDue)
		if level == 0 {
			continue // Not yet time for first dunning
		}

		// Check if dunning at this level already exists
		existing, err := s.dunningRepo.List(ctx, tenantID, invoice.ID)
		if err == nil && len(existing) > 0 {
			// Check if we already have this level
			hasLevel := false
			for _, entry := range existing {
				if entry.Level == level {
					hasLevel = true
					break
				}
			}
			if hasLevel {
				continue // Already sent this level
			}
		}

		// Create dunning entry
		fee := calc.GetFeeForLevel(level)
		now := time.Now()
		dunningID := fmt.Sprintf("dun_%d_%d", now.Unix(), level)

		dunning := domain.NewDunningEntry(
			dunningID,
			tenantID,
			invoice.ID,
			invoice.InvoiceNumber,
			level,
			7, // 7 days to pay
			fee,
		)

		if err := s.dunningRepo.Create(ctx, dunning); err != nil {
			s.logger.Error("Failed to create auto dunning", err)
			continue
		}

		dunningIDs = append(dunningIDs, dunning.ID)
		s.logger.Info("Auto dunning created", "dunning_id", dunning.ID, "invoice_number", invoice.InvoiceNumber, "level", level)
	}

	return dunningIDs, nil
}

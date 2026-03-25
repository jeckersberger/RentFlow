package application

import (
	"context"
	"fmt"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/expense-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/expense-service/internal/ports"
)

type ExpenseService struct {
	expRepo ports.ExpenseRepository
	logger  logger.Logger
}

func NewExpenseService(expRepo ports.ExpenseRepository, logger logger.Logger) *ExpenseService {
	return &ExpenseService{
		expRepo: expRepo,
		logger:  logger,
	}
}

func (s *ExpenseService) CreateExpense(ctx context.Context, cmd CreateExpenseCommand) (*ExpenseDTO, error) {
	if cmd.TenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	expID := fmt.Sprintf("exp_%d", hashString(cmd.TenantID+cmd.Vendor+cmd.Date.String()))
	exp := domain.NewExpense(expID, cmd.TenantID, cmd.Vendor, cmd.Currency, cmd.CategoryCode, cmd.PaymentMethod, cmd.Amount, cmd.Date)
	exp.ReceiptRef = cmd.ReceiptRef
	exp.ProjectID = cmd.ProjectID
	exp.Notes = cmd.Notes

	// Neue Felder
	if cmd.Type != "" {
		exp.Type = domain.ExpenseType(cmd.Type)
	}
	exp.VendorAddress = cmd.VendorAddress
	exp.VendorVATID = cmd.VendorVATID
	exp.VendorIBAN = cmd.VendorIBAN
	exp.BookingAccount = cmd.BookingAccount
	exp.ServicePeriodFrom = cmd.ServicePeriodFrom
	exp.ServicePeriodTo = cmd.ServicePeriodTo
	exp.DueDate = cmd.DueDate
	exp.DiscountPercent = cmd.DiscountPercent
	exp.DiscountDays = cmd.DiscountDays
	exp.InvoiceNumber = cmd.InvoiceNumber
	exp.ReceiptChecksum = cmd.ReceiptChecksum
	exp.ReceiptNASPath = cmd.ReceiptNASPath
	exp.OCRConfidence = cmd.OCRConfidence
	exp.EmailRef = cmd.EmailRef
	if cmd.Source != "" {
		exp.Source = cmd.Source
	}
	if cmd.OCRData != nil {
		exp.OCRData = cmd.OCRData
	}

	// Bewirtungsbeleg
	exp.EntertainmentLocation = cmd.EntertainmentLocation
	exp.EntertainmentReason = cmd.EntertainmentReason
	exp.EntertainmentGuests = cmd.EntertainmentGuests
	exp.EntertainmentTip = cmd.EntertainmentTip

	if cmd.TaxRate > 0 {
		exp.TaxRate = cmd.TaxRate
	}
	exp.CalculateTax()
	exp.CalculateEntertainmentSplit()
	exp.CalculateDiscountDeadline()

	if err := exp.Validate(); err != nil {
		return nil, domain.NewDomainError("VALIDATION_ERROR", err.Error(), nil)
	}

	if err := s.expRepo.CreateExpense(ctx, exp); err != nil {
		return nil, domain.NewDomainError("CREATE_FAILED", "failed to create expense", err)
	}

	return ExpenseToDTO(exp), nil
}

func (s *ExpenseService) GetExpense(ctx context.Context, tenantID, expenseID string) (*ExpenseDTO, error) {
	if tenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	exp, err := s.expRepo.GetExpense(ctx, tenantID, expenseID)
	if err != nil {
		return nil, domain.NewDomainError("NOT_FOUND", "expense not found", err)
	}

	return ExpenseToDTO(exp), nil
}

func (s *ExpenseService) ListExpenses(ctx context.Context, tenantID string, limit, offset int) ([]*ExpenseDTO, int64, error) {
	if tenantID == "" {
		return nil, 0, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	expenses, total, err := s.expRepo.ListExpenses(ctx, tenantID, limit, offset)
	if err != nil {
		return nil, 0, domain.NewDomainError("LIST_FAILED", "failed to list expenses", err)
	}

	dtos := make([]*ExpenseDTO, len(expenses))
	for i, e := range expenses {
		dtos[i] = ExpenseToDTO(e)
	}

	return dtos, total, nil
}

func (s *ExpenseService) UpdateExpense(ctx context.Context, cmd UpdateExpenseCommand) (*ExpenseDTO, error) {
	if cmd.TenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	exp, err := s.expRepo.GetExpense(ctx, cmd.TenantID, cmd.ID)
	if err != nil {
		return nil, domain.NewDomainError("NOT_FOUND", "expense not found", err)
	}

	if cmd.Type != "" {
		exp.Type = domain.ExpenseType(cmd.Type)
	}
	if cmd.Vendor != "" {
		exp.Vendor = cmd.Vendor
	}
	if cmd.VendorAddress != "" {
		exp.VendorAddress = cmd.VendorAddress
	}
	if cmd.VendorVATID != "" {
		exp.VendorVATID = cmd.VendorVATID
	}
	if cmd.VendorIBAN != "" {
		exp.VendorIBAN = cmd.VendorIBAN
	}
	if cmd.Amount > 0 {
		exp.Amount = cmd.Amount
		exp.CalculateTax()
	}
	if cmd.CategoryCode != "" {
		exp.CategoryCode = cmd.CategoryCode
	}
	if cmd.BookingAccount != "" {
		exp.BookingAccount = cmd.BookingAccount
	}
	if cmd.DueDate != nil {
		exp.DueDate = cmd.DueDate
	}
	if cmd.DiscountPercent > 0 {
		exp.DiscountPercent = cmd.DiscountPercent
	}
	if cmd.DiscountDays > 0 {
		exp.DiscountDays = cmd.DiscountDays
		exp.CalculateDiscountDeadline()
	}
	if cmd.PaymentMethod != "" {
		exp.PaymentMethod = cmd.PaymentMethod
	}
	if cmd.PaymentStatus != "" {
		exp.PaymentStatus = domain.PaymentStatus(cmd.PaymentStatus)
		if cmd.PaymentStatus == "paid" {
			now := time.Now()
			exp.PaidAt = &now
		}
	}
	if cmd.InvoiceNumber != "" {
		exp.InvoiceNumber = cmd.InvoiceNumber
	}
	if cmd.Notes != "" {
		exp.Notes = cmd.Notes
	}
	// Bewirtungsbeleg
	if cmd.EntertainmentLocation != "" {
		exp.EntertainmentLocation = cmd.EntertainmentLocation
	}
	if cmd.EntertainmentReason != "" {
		exp.EntertainmentReason = cmd.EntertainmentReason
	}
	if cmd.EntertainmentGuests != "" {
		exp.EntertainmentGuests = cmd.EntertainmentGuests
	}
	if cmd.EntertainmentTip > 0 {
		exp.EntertainmentTip = cmd.EntertainmentTip
	}
	exp.CalculateEntertainmentSplit()

	if err := s.expRepo.UpdateExpense(ctx, exp); err != nil {
		return nil, domain.NewDomainError("UPDATE_FAILED", "failed to update expense", err)
	}

	return ExpenseToDTO(exp), nil
}

func (s *ExpenseService) ApproveExpense(ctx context.Context, tenantID, expenseID, approvedBy string) (*ExpenseDTO, error) {
	if tenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	exp, err := s.expRepo.GetExpense(ctx, tenantID, expenseID)
	if err != nil {
		return nil, domain.NewDomainError("NOT_FOUND", "expense not found", err)
	}

	if err := exp.Approve(approvedBy); err != nil {
		return nil, domain.NewDomainError("APPROVE_FAILED", err.Error(), nil)
	}

	if err := s.expRepo.UpdateExpense(ctx, exp); err != nil {
		return nil, domain.NewDomainError("UPDATE_FAILED", "failed to approve expense", err)
	}

	return ExpenseToDTO(exp), nil
}

func (s *ExpenseService) RejectExpense(ctx context.Context, tenantID, expenseID string) (*ExpenseDTO, error) {
	if tenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	exp, err := s.expRepo.GetExpense(ctx, tenantID, expenseID)
	if err != nil {
		return nil, domain.NewDomainError("NOT_FOUND", "expense not found", err)
	}

	if err := exp.Reject(); err != nil {
		return nil, domain.NewDomainError("REJECT_FAILED", err.Error(), nil)
	}

	if err := s.expRepo.UpdateExpense(ctx, exp); err != nil {
		return nil, domain.NewDomainError("UPDATE_FAILED", "failed to reject expense", err)
	}

	return ExpenseToDTO(exp), nil
}

func (s *ExpenseService) DeleteExpense(ctx context.Context, tenantID, expenseID string) error {
	if tenantID == "" {
		return domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	if err := s.expRepo.DeleteExpense(ctx, tenantID, expenseID); err != nil {
		return domain.NewDomainError("DELETE_FAILED", "failed to delete expense", err)
	}

	return nil
}

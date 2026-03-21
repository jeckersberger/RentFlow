package application

import (
	"context"
	"fmt"

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

	if cmd.TaxRate > 0 {
		exp.TaxRate = cmd.TaxRate
	}
	exp.CalculateTax()

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

	if cmd.Vendor != "" {
		exp.Vendor = cmd.Vendor
	}
	if cmd.Amount > 0 {
		exp.Amount = cmd.Amount
		exp.CalculateTax()
	}
	if cmd.CategoryCode != "" {
		exp.CategoryCode = cmd.CategoryCode
	}
	if cmd.PaymentMethod != "" {
		exp.PaymentMethod = cmd.PaymentMethod
	}
	if cmd.Notes != "" {
		exp.Notes = cmd.Notes
	}

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

package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/services/expense/internal/domain"
)

// ---------------------------------------------------------------------------
// Request DTOs
// ---------------------------------------------------------------------------

type CreateExpenseRequest struct {
	ProjectID     string `json:"project_id"`
	CategoryID    string `json:"category_id"`
	Description   string `json:"description"`
	Amount        int64  `json:"amount"`
	Currency      string `json:"currency"`
	ExpenseDate   string `json:"expense_date"`
	ReceiptNumber string `json:"receipt_number"`
	Vendor        string `json:"vendor"`
	Notes         string `json:"notes"`
}

type UpdateExpenseRequest struct {
	ProjectID     *string `json:"project_id"`
	CategoryID    *string `json:"category_id"`
	Description   *string `json:"description"`
	Amount        *int64  `json:"amount"`
	Currency      *string `json:"currency"`
	ExpenseDate   *string `json:"expense_date"`
	ReceiptNumber *string `json:"receipt_number"`
	Vendor        *string `json:"vendor"`
	Notes         *string `json:"notes"`
}

// ---------------------------------------------------------------------------
// Service
// ---------------------------------------------------------------------------

type ExpenseService struct {
	repo   domain.ExpenseRepository
	logger zerolog.Logger
}

func NewExpenseService(repo domain.ExpenseRepository, logger zerolog.Logger) *ExpenseService {
	return &ExpenseService{
		repo:   repo,
		logger: logger.With().Str("service", "expense").Logger(),
	}
}

func (s *ExpenseService) Create(ctx context.Context, tenantID, userID uuid.UUID, req CreateExpenseRequest) (*domain.Expense, error) {
	if req.Description == "" {
		return nil, fmt.Errorf("description is required")
	}
	if req.Amount <= 0 {
		return nil, fmt.Errorf("amount must be positive")
	}

	currency := "EUR"
	if req.Currency != "" {
		currency = req.Currency
	}

	expenseDate := time.Now().Format("2006-01-02")
	if req.ExpenseDate != "" {
		expenseDate = req.ExpenseDate
	}

	expense := &domain.Expense{
		ID:            uuid.New(),
		TenantID:      tenantID,
		Description:   req.Description,
		Amount:        req.Amount,
		Currency:      currency,
		ExpenseDate:   expenseDate,
		ReceiptNumber: req.ReceiptNumber,
		Vendor:        req.Vendor,
		Status:        domain.StatusPending,
		Notes:         req.Notes,
		CreatedBy:     userID,
	}

	if req.ProjectID != "" {
		pid, err := uuid.Parse(req.ProjectID)
		if err != nil {
			return nil, fmt.Errorf("invalid project_id: %w", err)
		}
		expense.ProjectID = pid
	}

	if req.CategoryID != "" {
		cid, err := uuid.Parse(req.CategoryID)
		if err != nil {
			return nil, fmt.Errorf("invalid category_id: %w", err)
		}
		expense.CategoryID = cid
	}

	if err := s.repo.Create(ctx, expense); err != nil {
		return nil, fmt.Errorf("create expense: %w", err)
	}

	s.logger.Info().Str("expense_id", expense.ID.String()).Int64("amount", expense.Amount).Msg("expense created")
	return expense, nil
}

func (s *ExpenseService) GetByID(ctx context.Context, id, tenantID uuid.UUID) (*domain.Expense, error) {
	return s.repo.GetByID(ctx, id, tenantID)
}

func (s *ExpenseService) List(ctx context.Context, tenantID uuid.UUID, filter domain.ExpenseFilter) ([]*domain.Expense, int64, error) {
	return s.repo.List(ctx, tenantID, filter)
}

func (s *ExpenseService) Update(ctx context.Context, id, tenantID uuid.UUID, req UpdateExpenseRequest) (*domain.Expense, error) {
	existing, err := s.repo.GetByID(ctx, id, tenantID)
	if err != nil {
		return nil, err
	}

	if req.Description != nil {
		existing.Description = *req.Description
	}
	if req.Amount != nil {
		existing.Amount = *req.Amount
	}
	if req.Currency != nil {
		existing.Currency = *req.Currency
	}
	if req.ExpenseDate != nil {
		existing.ExpenseDate = *req.ExpenseDate
	}
	if req.ReceiptNumber != nil {
		existing.ReceiptNumber = *req.ReceiptNumber
	}
	if req.Vendor != nil {
		existing.Vendor = *req.Vendor
	}
	if req.Notes != nil {
		existing.Notes = *req.Notes
	}
	if req.ProjectID != nil {
		if *req.ProjectID != "" {
			pid, parseErr := uuid.Parse(*req.ProjectID)
			if parseErr != nil {
				return nil, fmt.Errorf("invalid project_id: %w", parseErr)
			}
			existing.ProjectID = pid
		} else {
			existing.ProjectID = uuid.Nil
		}
	}
	if req.CategoryID != nil {
		if *req.CategoryID != "" {
			cid, parseErr := uuid.Parse(*req.CategoryID)
			if parseErr != nil {
				return nil, fmt.Errorf("invalid category_id: %w", parseErr)
			}
			existing.CategoryID = cid
		} else {
			existing.CategoryID = uuid.Nil
		}
	}

	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, err
	}

	return existing, nil
}

func (s *ExpenseService) Approve(ctx context.Context, id, tenantID, approverID uuid.UUID) (*domain.Expense, error) {
	existing, err := s.repo.GetByID(ctx, id, tenantID)
	if err != nil {
		return nil, err
	}

	if existing.Status == domain.StatusApproved {
		return nil, fmt.Errorf("expense is already approved")
	}

	now := time.Now()
	existing.Status = domain.StatusApproved
	existing.ApprovedBy = approverID
	existing.ApprovedAt = &now

	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, err
	}

	s.logger.Info().Str("expense_id", id.String()).Str("approved_by", approverID.String()).Msg("expense approved")
	return existing, nil
}

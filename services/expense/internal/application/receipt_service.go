package application

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/services/expense/internal/domain"
)

// ---------------------------------------------------------------------------
// Request DTOs
// ---------------------------------------------------------------------------

type AddReceiptRequest struct {
	FileName string `json:"file_name"`
	FilePath string `json:"file_path"`
	FileSize int    `json:"file_size"`
	MimeType string `json:"mime_type"`
}

// ---------------------------------------------------------------------------
// Service
// ---------------------------------------------------------------------------

type ReceiptService struct {
	receiptRepo domain.ExpenseReceiptRepository
	expenseRepo domain.ExpenseRepository
	logger      zerolog.Logger
}

func NewReceiptService(receiptRepo domain.ExpenseReceiptRepository, expenseRepo domain.ExpenseRepository, logger zerolog.Logger) *ReceiptService {
	return &ReceiptService{
		receiptRepo: receiptRepo,
		expenseRepo: expenseRepo,
		logger:      logger.With().Str("service", "expense-receipt").Logger(),
	}
}

func (s *ReceiptService) AddReceipt(ctx context.Context, expenseID, tenantID, userID uuid.UUID, req AddReceiptRequest) (*domain.ExpenseReceipt, error) {
	// Verify expense exists and belongs to tenant.
	if _, err := s.expenseRepo.GetByID(ctx, expenseID, tenantID); err != nil {
		return nil, err
	}

	if req.FileName == "" {
		return nil, fmt.Errorf("file_name is required")
	}
	if req.FilePath == "" {
		return nil, fmt.Errorf("file_path is required")
	}

	receipt := &domain.ExpenseReceipt{
		ID:         uuid.New(),
		ExpenseID:  expenseID,
		TenantID:   tenantID,
		FileName:   req.FileName,
		FilePath:   req.FilePath,
		FileSize:   req.FileSize,
		MimeType:   req.MimeType,
		UploadedBy: userID,
	}

	if err := s.receiptRepo.Create(ctx, receipt); err != nil {
		return nil, fmt.Errorf("add receipt: %w", err)
	}

	s.logger.Info().Str("receipt_id", receipt.ID.String()).Str("expense_id", expenseID.String()).Msg("receipt added")
	return receipt, nil
}

func (s *ReceiptService) ListReceipts(ctx context.Context, expenseID, tenantID uuid.UUID) ([]*domain.ExpenseReceipt, error) {
	// Verify expense exists and belongs to tenant.
	if _, err := s.expenseRepo.GetByID(ctx, expenseID, tenantID); err != nil {
		return nil, err
	}

	return s.receiptRepo.ListByExpense(ctx, expenseID, tenantID)
}

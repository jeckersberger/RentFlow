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

// CreateRecurringRequest holds the data for creating a recurring expense.
type CreateRecurringRequest struct {
	Name       string `json:"name"`
	CategoryID string `json:"category_id"`
	Amount     int64  `json:"amount"`
	Frequency  string `json:"frequency"`
	NextDate   string `json:"next_date"`
}

// UpdateRecurringRequest holds the data for updating a recurring expense.
type UpdateRecurringRequest struct {
	Name       *string `json:"name"`
	CategoryID *string `json:"category_id"`
	Amount     *int64  `json:"amount"`
	Frequency  *string `json:"frequency"`
	NextDate   *string `json:"next_date"`
	IsActive   *bool   `json:"is_active"`
}

// ---------------------------------------------------------------------------
// Service
// ---------------------------------------------------------------------------

// RecurringService implements application-level use cases for recurring expenses.
type RecurringService struct {
	repo   domain.RecurringExpenseRepository
	logger zerolog.Logger
}

// NewRecurringService constructs a new RecurringService.
func NewRecurringService(repo domain.RecurringExpenseRepository, logger zerolog.Logger) *RecurringService {
	return &RecurringService{
		repo:   repo,
		logger: logger.With().Str("service", "recurring").Logger(),
	}
}

// Create creates a new recurring expense.
func (s *RecurringService) Create(ctx context.Context, tenantID uuid.UUID, req CreateRecurringRequest) (*domain.RecurringExpense, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if req.Amount <= 0 {
		return nil, fmt.Errorf("amount must be positive")
	}

	frequency := "monthly"
	if req.Frequency != "" {
		frequency = req.Frequency
	}

	rec := &domain.RecurringExpense{
		ID:        uuid.New(),
		TenantID:  tenantID,
		Name:      req.Name,
		Amount:    req.Amount,
		Frequency: frequency,
		NextDate:  req.NextDate,
		IsActive:  true,
	}

	if req.CategoryID != "" {
		cid, err := uuid.Parse(req.CategoryID)
		if err != nil {
			return nil, fmt.Errorf("invalid category_id: %w", err)
		}
		rec.CategoryID = cid
	}

	if err := s.repo.Create(ctx, rec); err != nil {
		return nil, fmt.Errorf("create recurring expense: %w", err)
	}

	s.logger.Info().Str("recurring_id", rec.ID.String()).Msg("recurring expense created")
	return rec, nil
}

// List returns a paginated list of recurring expenses.
func (s *RecurringService) List(ctx context.Context, tenantID uuid.UUID, filter domain.RecurringExpenseFilter) ([]*domain.RecurringExpense, int64, error) {
	return s.repo.List(ctx, tenantID, filter)
}

// Update updates an existing recurring expense.
func (s *RecurringService) Update(ctx context.Context, id, tenantID uuid.UUID, req UpdateRecurringRequest) (*domain.RecurringExpense, error) {
	existing, err := s.repo.GetByID(ctx, id, tenantID)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		existing.Name = *req.Name
	}
	if req.Amount != nil {
		existing.Amount = *req.Amount
	}
	if req.Frequency != nil {
		existing.Frequency = *req.Frequency
	}
	if req.NextDate != nil {
		existing.NextDate = *req.NextDate
	}
	if req.IsActive != nil {
		existing.IsActive = *req.IsActive
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

// Delete removes a recurring expense.
func (s *RecurringService) Delete(ctx context.Context, id, tenantID uuid.UUID) error {
	return s.repo.Delete(ctx, id, tenantID)
}

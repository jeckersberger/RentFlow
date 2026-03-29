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

// CreateBudgetRequest holds the data for creating a budget.
type CreateBudgetRequest struct {
	Name              string `json:"name"`
	ScopeType         string `json:"scope_type"`
	ScopeID           string `json:"scope_id"`
	PeriodType        string `json:"period_type"`
	Amount            int64  `json:"amount"`
	AlertThresholdPct *int   `json:"alert_threshold_pct"`
}

// UpdateBudgetRequest holds the data for updating a budget.
type UpdateBudgetRequest struct {
	Name              *string `json:"name"`
	ScopeType         *string `json:"scope_type"`
	ScopeID           *string `json:"scope_id"`
	PeriodType        *string `json:"period_type"`
	Amount            *int64  `json:"amount"`
	AlertThresholdPct *int    `json:"alert_threshold_pct"`
	IsActive          *bool   `json:"is_active"`
}

// ---------------------------------------------------------------------------
// Service
// ---------------------------------------------------------------------------

// BudgetService implements application-level use cases for budgets.
type BudgetService struct {
	repo   domain.BudgetRepository
	logger zerolog.Logger
}

// NewBudgetService constructs a new BudgetService.
func NewBudgetService(repo domain.BudgetRepository, logger zerolog.Logger) *BudgetService {
	return &BudgetService{
		repo:   repo,
		logger: logger.With().Str("service", "budget").Logger(),
	}
}

// Create creates a new budget.
func (s *BudgetService) Create(ctx context.Context, tenantID uuid.UUID, req CreateBudgetRequest) (*domain.Budget, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if req.Amount <= 0 {
		return nil, fmt.Errorf("amount must be positive")
	}

	scopeType := "global"
	if req.ScopeType != "" {
		scopeType = req.ScopeType
	}
	periodType := "monthly"
	if req.PeriodType != "" {
		periodType = req.PeriodType
	}
	alertPct := 80
	if req.AlertThresholdPct != nil {
		alertPct = *req.AlertThresholdPct
	}

	budget := &domain.Budget{
		ID:                uuid.New(),
		TenantID:          tenantID,
		Name:              req.Name,
		ScopeType:         scopeType,
		PeriodType:        periodType,
		Amount:            req.Amount,
		AlertThresholdPct: alertPct,
		IsActive:          true,
	}

	if req.ScopeID != "" {
		sid, err := uuid.Parse(req.ScopeID)
		if err != nil {
			return nil, fmt.Errorf("invalid scope_id: %w", err)
		}
		budget.ScopeID = sid
	}

	if err := s.repo.Create(ctx, budget); err != nil {
		return nil, fmt.Errorf("create budget: %w", err)
	}

	s.logger.Info().Str("budget_id", budget.ID.String()).Int64("amount", budget.Amount).Msg("budget created")
	return budget, nil
}

// List returns a paginated list of budgets with current spent amounts.
func (s *BudgetService) List(ctx context.Context, tenantID uuid.UUID, filter domain.BudgetFilter) ([]*domain.Budget, int64, error) {
	return s.repo.List(ctx, tenantID, filter)
}

// Update updates an existing budget.
func (s *BudgetService) Update(ctx context.Context, id, tenantID uuid.UUID, req UpdateBudgetRequest) (*domain.Budget, error) {
	existing, err := s.repo.GetByID(ctx, id, tenantID)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		existing.Name = *req.Name
	}
	if req.ScopeType != nil {
		existing.ScopeType = *req.ScopeType
	}
	if req.PeriodType != nil {
		existing.PeriodType = *req.PeriodType
	}
	if req.Amount != nil {
		existing.Amount = *req.Amount
	}
	if req.AlertThresholdPct != nil {
		existing.AlertThresholdPct = *req.AlertThresholdPct
	}
	if req.IsActive != nil {
		existing.IsActive = *req.IsActive
	}
	if req.ScopeID != nil {
		if *req.ScopeID != "" {
			sid, parseErr := uuid.Parse(*req.ScopeID)
			if parseErr != nil {
				return nil, fmt.Errorf("invalid scope_id: %w", parseErr)
			}
			existing.ScopeID = sid
		} else {
			existing.ScopeID = uuid.Nil
		}
	}

	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, err
	}

	return existing, nil
}

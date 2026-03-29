package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	apperrors "github.com/jeckersberger/EquipFlow/pkg/common/errors"
	"github.com/jeckersberger/EquipFlow/services/expense/internal/domain"
)

const budgetColumns = `
	id, tenant_id, name, scope_type, scope_id,
	period_type, amount, alert_threshold_pct, is_active, created_at`

// BudgetRepo implements domain.BudgetRepository using PostgreSQL.
type BudgetRepo struct {
	pool *pgxpool.Pool
}

// NewBudgetRepo creates a new BudgetRepo.
func NewBudgetRepo(pool *pgxpool.Pool) *BudgetRepo {
	return &BudgetRepo{pool: pool}
}

func scanBudget(row pgx.Row) (*domain.Budget, error) {
	b := &domain.Budget{}
	var scopeID *uuid.UUID

	err := row.Scan(
		&b.ID, &b.TenantID, &b.Name, &b.ScopeType, &scopeID,
		&b.PeriodType, &b.Amount, &b.AlertThresholdPct, &b.IsActive, &b.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	if scopeID != nil {
		b.ScopeID = *scopeID
	}

	return b, nil
}

func (r *BudgetRepo) Create(ctx context.Context, budget *domain.Budget) error {
	query := `
		INSERT INTO budgets (
			id, tenant_id, name, scope_type, scope_id,
			period_type, amount, alert_threshold_pct, is_active
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING created_at`

	err := r.pool.QueryRow(ctx, query,
		budget.ID, budget.TenantID, budget.Name, budget.ScopeType, nilUUID(budget.ScopeID),
		budget.PeriodType, budget.Amount, budget.AlertThresholdPct, budget.IsActive,
	).Scan(&budget.CreatedAt)
	if err != nil {
		return fmt.Errorf("budget_repo: create: %w", err)
	}
	return nil
}

func (r *BudgetRepo) GetByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*domain.Budget, error) {
	query := fmt.Sprintf(`SELECT %s FROM budgets WHERE id = $1 AND tenant_id = $2`, budgetColumns)
	b, err := scanBudget(r.pool.QueryRow(ctx, query, id, tenantID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("budget_repo: get_by_id: %w", err)
	}
	return b, nil
}

func (r *BudgetRepo) List(ctx context.Context, tenantID uuid.UUID, filter domain.BudgetFilter) ([]*domain.Budget, int64, error) {
	var total int64
	countQuery := `SELECT COUNT(*) FROM budgets WHERE tenant_id = $1`
	err := r.pool.QueryRow(ctx, countQuery, tenantID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("budget_repo: list count: %w", err)
	}

	page := filter.Page
	if page < 1 {
		page = 1
	}
	perPage := filter.PerPage
	if perPage < 1 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	dataQuery := fmt.Sprintf(
		`SELECT %s FROM budgets WHERE tenant_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		budgetColumns,
	)

	rows, err := r.pool.Query(ctx, dataQuery, tenantID, perPage, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("budget_repo: list query: %w", err)
	}
	defer rows.Close()

	var items []*domain.Budget
	for rows.Next() {
		b, scanErr := scanBudget(rows)
		if scanErr != nil {
			return nil, 0, fmt.Errorf("budget_repo: list scan: %w", scanErr)
		}
		items = append(items, b)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("budget_repo: list rows: %w", err)
	}

	// Calculate spent amount for each budget based on current period.
	for _, b := range items {
		spent, spentErr := r.calculateSpent(ctx, tenantID, b)
		if spentErr == nil {
			b.SpentAmount = spent
		}
	}

	return items, total, nil
}

func (r *BudgetRepo) Update(ctx context.Context, budget *domain.Budget) error {
	query := `
		UPDATE budgets SET
			name = $3, scope_type = $4, scope_id = $5,
			period_type = $6, amount = $7, alert_threshold_pct = $8, is_active = $9
		WHERE id = $1 AND tenant_id = $2
		RETURNING created_at`

	err := r.pool.QueryRow(ctx, query,
		budget.ID, budget.TenantID,
		budget.Name, budget.ScopeType, nilUUID(budget.ScopeID),
		budget.PeriodType, budget.Amount, budget.AlertThresholdPct, budget.IsActive,
	).Scan(&budget.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apperrors.ErrNotFound
		}
		return fmt.Errorf("budget_repo: update: %w", err)
	}
	return nil
}

// calculateSpent queries the sum of approved expenses for the current period of this budget.
func (r *BudgetRepo) calculateSpent(ctx context.Context, tenantID uuid.UUID, budget *domain.Budget) (int64, error) {
	var periodStart time.Time
	now := time.Now()

	switch budget.PeriodType {
	case "monthly":
		periodStart = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	case "quarterly":
		quarter := (now.Month() - 1) / 3 * 3
		periodStart = time.Date(now.Year(), quarter+1, 1, 0, 0, 0, 0, now.Location())
	case "yearly":
		periodStart = time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
	default:
		periodStart = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	}

	query := `SELECT COALESCE(SUM(amount), 0) FROM expenses
		WHERE tenant_id = $1 AND expense_date >= $2 AND status != 'rejected'`
	args := []interface{}{tenantID, periodStart.Format("2006-01-02")}

	if budget.ScopeType == "category" && budget.ScopeID != uuid.Nil {
		query += ` AND category_id = $3`
		args = append(args, budget.ScopeID)
	} else if budget.ScopeType == "project" && budget.ScopeID != uuid.Nil {
		query += ` AND project_id = $3`
		args = append(args, budget.ScopeID)
	}

	var spent int64
	err := r.pool.QueryRow(ctx, query, args...).Scan(&spent)
	if err != nil {
		return 0, fmt.Errorf("budget_repo: calculate_spent: %w", err)
	}
	return spent, nil
}

// Ensure interface compliance at compile time.
var _ domain.BudgetRepository = (*BudgetRepo)(nil)

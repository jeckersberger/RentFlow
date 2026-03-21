package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jeckersberger/rentflow/pkg/common/database"
	"github.com/jeckersberger/rentflow/services/expense-service/internal/domain"
)

type BudgetPostgres struct {
	db *database.PostgresPool
}

func NewBudgetPostgres(db *database.PostgresPool) *BudgetPostgres {
	return &BudgetPostgres{db: db}
}

func (r *BudgetPostgres) CreateBudget(ctx context.Context, budget *domain.Budget) error {
	query := `
		INSERT INTO expenses.budgets (
			id, tenant_id, category_id, project_id, period, amount, spent, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		)
	`

	_, err := r.db.Exec(ctx, query,
		budget.ID, budget.TenantID, budget.CategoryID, budget.ProjectID, string(budget.Period),
		budget.Amount, budget.Spent, budget.CreatedAt, budget.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create budget: %w", err)
	}

	return nil
}

func (r *BudgetPostgres) GetBudget(ctx context.Context, tenantID, budgetID string) (*domain.Budget, error) {
	query := `
		SELECT id, tenant_id, category_id, project_id, period, amount, spent, created_at, updated_at
		FROM expenses.budgets
		WHERE id = $1 AND tenant_id = $2
	`

	var budget domain.Budget
	var period string

	err := r.db.QueryRow(ctx, query, budgetID, tenantID).Scan(
		&budget.ID, &budget.TenantID, &budget.CategoryID, &budget.ProjectID, &period,
		&budget.Amount, &budget.Spent, &budget.CreatedAt, &budget.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("budget not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get budget: %w", err)
	}

	budget.Period = domain.BudgetPeriod(period)

	return &budget, nil
}

func (r *BudgetPostgres) ListBudgets(ctx context.Context, tenantID string, limit, offset int) ([]*domain.Budget, int64, error) {
	countQuery := `SELECT COUNT(*) FROM expenses.budgets WHERE tenant_id = $1`
	var total int64
	if err := r.db.QueryRow(ctx, countQuery, tenantID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count budgets: %w", err)
	}

	query := `
		SELECT id, tenant_id, category_id, project_id, period, amount, spent, created_at, updated_at
		FROM expenses.budgets
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list budgets: %w", err)
	}
	defer rows.Close()

	var budgets []*domain.Budget
	for rows.Next() {
		var budget domain.Budget
		var period string

		err := rows.Scan(
			&budget.ID, &budget.TenantID, &budget.CategoryID, &budget.ProjectID, &period,
			&budget.Amount, &budget.Spent, &budget.CreatedAt, &budget.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan budget: %w", err)
		}

		budget.Period = domain.BudgetPeriod(period)
		budgets = append(budgets, &budget)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows error: %w", err)
	}

	return budgets, total, nil
}

func (r *BudgetPostgres) GetByCategory(ctx context.Context, tenantID, categoryID, period string) (*domain.Budget, error) {
	query := `
		SELECT id, tenant_id, category_id, project_id, period, amount, spent, created_at, updated_at
		FROM expenses.budgets
		WHERE tenant_id = $1 AND category_id = $2 AND period = $3
		LIMIT 1
	`

	var budget domain.Budget
	var periodVal string

	err := r.db.QueryRow(ctx, query, tenantID, categoryID, period).Scan(
		&budget.ID, &budget.TenantID, &budget.CategoryID, &budget.ProjectID, &periodVal,
		&budget.Amount, &budget.Spent, &budget.CreatedAt, &budget.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("budget not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get budget: %w", err)
	}

	budget.Period = domain.BudgetPeriod(periodVal)

	return &budget, nil
}

func (r *BudgetPostgres) DeleteBudget(ctx context.Context, tenantID, budgetID string) error {
	query := `DELETE FROM expenses.budgets WHERE id = $1 AND tenant_id = $2`

	result, err := r.db.Exec(ctx, query, budgetID, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete budget: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("budget not found")
	}

	return nil
}

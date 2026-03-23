package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jeckersberger/rentflow/pkg/common/database"
	"github.com/jeckersberger/rentflow/services/expense-service/internal/domain"
)

type CategoryPostgres struct {
	db *database.PostgresPool
}

func NewCategoryPostgres(db *database.PostgresPool) *CategoryPostgres {
	return &CategoryPostgres{db: db}
}

func (r *CategoryPostgres) CreateCategory(ctx context.Context, cat *domain.ExpenseCategory) error {
	query := `
		INSERT INTO expenses.expense_categories (
			id, tenant_id, name, skr03_code, skr04_code, is_default, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8
		)
	`

	_, err := r.db.Exec(ctx, query,
		cat.ID, cat.TenantID, cat.Name, cat.SKR03Code, cat.SKR04Code, cat.IsDefault,
		cat.CreatedAt, cat.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create category: %w", err)
	}

	return nil
}

func (r *CategoryPostgres) UpdateCategory(ctx context.Context, cat *domain.ExpenseCategory) error {
	query := `
		UPDATE expenses.expense_categories SET
			name = $3, skr03_code = $4, skr04_code = $5, is_default = $6, updated_at = $7
		WHERE id = $1 AND tenant_id = $2
	`

	result, err := r.db.Exec(ctx, query,
		cat.ID, cat.TenantID, cat.Name, cat.SKR03Code, cat.SKR04Code, cat.IsDefault, cat.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update category: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("category not found")
	}

	return nil
}

func (r *CategoryPostgres) GetCategory(ctx context.Context, tenantID, categoryID string) (*domain.ExpenseCategory, error) {
	query := `
		SELECT id, tenant_id, name, COALESCE(skr03_code, ''), COALESCE(skr04_code, ''), is_default, created_at, updated_at
		FROM expenses.expense_categories
		WHERE id = $1 AND tenant_id = $2
	`

	var cat domain.ExpenseCategory

	err := r.db.QueryRow(ctx, query, categoryID, tenantID).Scan(
		&cat.ID, &cat.TenantID, &cat.Name, &cat.SKR03Code, &cat.SKR04Code, &cat.IsDefault,
		&cat.CreatedAt, &cat.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("category not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get category: %w", err)
	}

	return &cat, nil
}

func (r *CategoryPostgres) ListCategories(ctx context.Context, tenantID string, limit, offset int) ([]*domain.ExpenseCategory, int64, error) {
	countQuery := `SELECT COUNT(*) FROM expenses.expense_categories WHERE tenant_id = $1`
	var total int64
	if err := r.db.QueryRow(ctx, countQuery, tenantID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count categories: %w", err)
	}

	query := `
		SELECT id, tenant_id, name, COALESCE(skr03_code, ''), COALESCE(skr04_code, ''), is_default, created_at, updated_at
		FROM expenses.expense_categories
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list categories: %w", err)
	}
	defer rows.Close()

	var categories []*domain.ExpenseCategory
	for rows.Next() {
		var cat domain.ExpenseCategory

		err := rows.Scan(
			&cat.ID, &cat.TenantID, &cat.Name, &cat.SKR03Code, &cat.SKR04Code, &cat.IsDefault,
			&cat.CreatedAt, &cat.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan category: %w", err)
		}

		categories = append(categories, &cat)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows error: %w", err)
	}

	return categories, total, nil
}

func (r *CategoryPostgres) DeleteCategory(ctx context.Context, tenantID, categoryID string) error {
	query := `DELETE FROM expenses.expense_categories WHERE id = $1 AND tenant_id = $2`

	result, err := r.db.Exec(ctx, query, categoryID, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("category not found")
	}

	return nil
}

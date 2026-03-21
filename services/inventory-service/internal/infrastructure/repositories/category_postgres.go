package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jeckersberger/rentflow/pkg/common/database"
	"github.com/jeckersberger/rentflow/services/inventory-service/internal/domain"
)

type CategoryPostgres struct {
	db *database.PostgresPool
}

func NewCategoryPostgres(db *database.PostgresPool) *CategoryPostgres {
	return &CategoryPostgres{db: db}
}

func (r *CategoryPostgres) Create(ctx context.Context, cat *domain.Category) error {
	query := `
		INSERT INTO inventory.categories (
			id, tenant_id, name, parent_id, icon, color, sort_order, created_at, created_by_user_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.Exec(ctx, query,
		cat.ID, cat.TenantID, cat.Name, cat.ParentID, cat.Icon, cat.Color,
		cat.SortOrder, cat.CreatedAt, cat.CreatedByUserID,
	)

	if err != nil {
		return fmt.Errorf("failed to create category: %w", err)
	}

	return nil
}

func (r *CategoryPostgres) Update(ctx context.Context, cat *domain.Category) error {
	query := `
		UPDATE inventory.categories SET
			name = $3, icon = $4, color = $5, sort_order = $6
		WHERE id = $1 AND tenant_id = $2
	`

	result, err := r.db.Exec(ctx, query,
		cat.ID, cat.TenantID, cat.Name, cat.Icon, cat.Color, cat.SortOrder,
	)

	if err != nil {
		return fmt.Errorf("failed to update category: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("category not found: %s", cat.ID)
	}

	return nil
}

func (r *CategoryPostgres) GetByID(ctx context.Context, tenantID, categoryID string) (*domain.Category, error) {
	query := `
		SELECT id, tenant_id, name, parent_id, icon, color, sort_order, created_at, created_by_user_id
		FROM inventory.categories
		WHERE id = $1 AND tenant_id = $2
	`

	row := r.db.QueryRow(ctx, query, categoryID, tenantID)
	return r.scanCategory(row)
}

func (r *CategoryPostgres) List(ctx context.Context, tenantID string) ([]*domain.Category, error) {
	query := `
		SELECT id, tenant_id, name, parent_id, icon, color, sort_order, created_at, created_by_user_id
		FROM inventory.categories
		WHERE tenant_id = $1
		ORDER BY sort_order ASC, created_at ASC
	`

	rows, err := r.db.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to query categories: %w", err)
	}
	defer rows.Close()

	var categories []*domain.Category
	for rows.Next() {
		cat, err := r.scanCategoryRow(rows)
		if err != nil {
			return nil, err
		}
		categories = append(categories, cat)
	}

	return categories, nil
}

func (r *CategoryPostgres) Delete(ctx context.Context, tenantID, categoryID string) error {
	query := "DELETE FROM inventory.categories WHERE id = $1 AND tenant_id = $2"
	_, err := r.db.Exec(ctx, query, categoryID, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}
	return nil
}

// Helper methods

func (r *CategoryPostgres) scanCategory(row *sql.Row) (*domain.Category, error) {
	cat := &domain.Category{}

	err := row.Scan(
		&cat.ID, &cat.TenantID, &cat.Name, &cat.ParentID, &cat.Icon, &cat.Color,
		&cat.SortOrder, &cat.CreatedAt, &cat.CreatedByUserID,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrCategoryNotFound
		}
		return nil, fmt.Errorf("failed to scan category: %w", err)
	}

	return cat, nil
}

func (r *CategoryPostgres) scanCategoryRow(rows interface {
	Scan(...interface{}) error
}) (*domain.Category, error) {
	cat := &domain.Category{}

	err := rows.Scan(
		&cat.ID, &cat.TenantID, &cat.Name, &cat.ParentID, &cat.Icon, &cat.Color,
		&cat.SortOrder, &cat.CreatedAt, &cat.CreatedByUserID,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to scan category: %w", err)
	}

	return cat, nil
}

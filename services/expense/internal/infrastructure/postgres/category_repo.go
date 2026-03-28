package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	apperrors "github.com/jeckersberger/EquipFlow/pkg/common/errors"
	"github.com/jeckersberger/EquipFlow/services/expense/internal/domain"
)

type CategoryRepo struct {
	pool *pgxpool.Pool
}

func NewCategoryRepo(pool *pgxpool.Pool) *CategoryRepo {
	return &CategoryRepo{pool: pool}
}

func (r *CategoryRepo) Create(ctx context.Context, category *domain.ExpenseCategory) error {
	query := `
		INSERT INTO expense_categories (id, tenant_id, name, code, description, is_active)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at`

	err := r.pool.QueryRow(ctx, query,
		category.ID, category.TenantID, category.Name,
		nilIfEmpty(category.Code), nilIfEmpty(category.Description), category.IsActive,
	).Scan(&category.CreatedAt)
	if err != nil {
		return fmt.Errorf("category_repo: create: %w", err)
	}
	return nil
}

func (r *CategoryRepo) GetByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*domain.ExpenseCategory, error) {
	query := `SELECT id, tenant_id, name, code, description, is_active, created_at
		FROM expense_categories WHERE id = $1 AND tenant_id = $2`

	cat := &domain.ExpenseCategory{}
	var code, description *string
	err := r.pool.QueryRow(ctx, query, id, tenantID).Scan(
		&cat.ID, &cat.TenantID, &cat.Name, &code, &description, &cat.IsActive, &cat.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("category_repo: get_by_id: %w", err)
	}
	cat.Code = derefString(code)
	cat.Description = derefString(description)
	return cat, nil
}

func (r *CategoryRepo) List(ctx context.Context, tenantID uuid.UUID) ([]*domain.ExpenseCategory, error) {
	query := `SELECT id, tenant_id, name, code, description, is_active, created_at
		FROM expense_categories WHERE tenant_id = $1 ORDER BY name ASC`

	rows, err := r.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("category_repo: list: %w", err)
	}
	defer rows.Close()

	var items []*domain.ExpenseCategory
	for rows.Next() {
		cat := &domain.ExpenseCategory{}
		var code, description *string
		if err := rows.Scan(
			&cat.ID, &cat.TenantID, &cat.Name, &code, &description, &cat.IsActive, &cat.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("category_repo: list scan: %w", err)
		}
		cat.Code = derefString(code)
		cat.Description = derefString(description)
		items = append(items, cat)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("category_repo: list rows: %w", err)
	}
	return items, nil
}

func (r *CategoryRepo) Update(ctx context.Context, category *domain.ExpenseCategory) error {
	query := `
		UPDATE expense_categories SET
			name = $3, code = $4, description = $5, is_active = $6
		WHERE id = $1 AND tenant_id = $2`

	tag, err := r.pool.Exec(ctx, query,
		category.ID, category.TenantID,
		category.Name, nilIfEmpty(category.Code), nilIfEmpty(category.Description), category.IsActive,
	)
	if err != nil {
		return fmt.Errorf("category_repo: update: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

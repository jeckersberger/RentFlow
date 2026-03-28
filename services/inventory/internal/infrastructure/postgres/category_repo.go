package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jeckersberger/EquipFlow/services/inventory/internal/domain"

	apperrors "github.com/jeckersberger/EquipFlow/pkg/common/errors"
)

// categoryColumns lists all columns of the categories table.
const categoryColumns = `id, tenant_id, name, parent_id, icon, color, sort_order, created_at, updated_at`

// CategoryRepo implements domain.CategoryRepository using PostgreSQL.
type CategoryRepo struct {
	pool *pgxpool.Pool
}

// NewCategoryRepo creates a new CategoryRepo.
func NewCategoryRepo(pool *pgxpool.Pool) *CategoryRepo {
	return &CategoryRepo{pool: pool}
}

// scanCategory scans a single category row into a domain.Category.
func scanCategory(row pgx.Row) (*domain.Category, error) {
	c := &domain.Category{}
	var icon, color *string

	err := row.Scan(
		&c.ID, &c.TenantID, &c.Name, &c.ParentID,
		&icon, &color, &c.SortOrder,
		&c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if icon != nil {
		c.Icon = *icon
	}
	if color != nil {
		c.Color = *color
	}
	return c, nil
}

// Create inserts a new category and scans back the generated fields.
func (r *CategoryRepo) Create(ctx context.Context, category *domain.Category) error {
	query := `
		INSERT INTO categories (
			id, tenant_id, name, parent_id, icon, color, sort_order
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		) RETURNING id, created_at, updated_at`

	err := r.pool.QueryRow(ctx, query,
		category.ID, category.TenantID, category.Name, category.ParentID,
		nilIfEmpty(category.Icon), nilIfEmpty(category.Color), category.SortOrder,
	).Scan(&category.ID, &category.CreatedAt, &category.UpdatedAt)
	if err != nil {
		return fmt.Errorf("category_repo: create: %w", err)
	}
	return nil
}

// GetByID retrieves a category by primary key scoped to a tenant.
func (r *CategoryRepo) GetByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*domain.Category, error) {
	query := fmt.Sprintf(`SELECT %s FROM categories WHERE id = $1 AND tenant_id = $2`, categoryColumns)
	c, err := scanCategory(r.pool.QueryRow(ctx, query, id, tenantID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("category_repo: get_by_id: %w", err)
	}
	return c, nil
}

// List returns all categories for a tenant ordered by sort_order then name.
func (r *CategoryRepo) List(ctx context.Context, tenantID uuid.UUID) ([]*domain.Category, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM categories WHERE tenant_id = $1 ORDER BY sort_order ASC, name ASC`,
		categoryColumns,
	)

	rows, err := r.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("category_repo: list query: %w", err)
	}
	defer rows.Close()

	var categories []*domain.Category
	for rows.Next() {
		c, scanErr := scanCategory(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("category_repo: list scan: %w", scanErr)
		}
		categories = append(categories, c)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("category_repo: list rows: %w", err)
	}
	return categories, nil
}

// Update modifies an existing category.
func (r *CategoryRepo) Update(ctx context.Context, category *domain.Category) error {
	query := `
		UPDATE categories SET
			name = $3, parent_id = $4, icon = $5, color = $6, sort_order = $7,
			updated_at = NOW()
		WHERE id = $1 AND tenant_id = $2
		RETURNING updated_at`

	err := r.pool.QueryRow(ctx, query,
		category.ID, category.TenantID,
		category.Name, category.ParentID,
		nilIfEmpty(category.Icon), nilIfEmpty(category.Color), category.SortOrder,
	).Scan(&category.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apperrors.ErrNotFound
		}
		return fmt.Errorf("category_repo: update: %w", err)
	}
	return nil
}

// Delete removes a category by primary key scoped to a tenant.
func (r *CategoryRepo) Delete(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error {
	tag, err := r.pool.Exec(ctx,
		`DELETE FROM categories WHERE id = $1 AND tenant_id = $2`,
		id, tenantID,
	)
	if err != nil {
		return fmt.Errorf("category_repo: delete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

// HasChildren checks whether a category has any child categories.
func (r *CategoryRepo) HasChildren(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM categories WHERE parent_id = $1 AND tenant_id = $2)`,
		id, tenantID,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("category_repo: has_children: %w", err)
	}
	return exists, nil
}

// HasEquipment checks whether any equipment is assigned to a category.
func (r *CategoryRepo) HasEquipment(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM equipment WHERE category_id = $1 AND tenant_id = $2)`,
		id, tenantID,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("category_repo: has_equipment: %w", err)
	}
	return exists, nil
}

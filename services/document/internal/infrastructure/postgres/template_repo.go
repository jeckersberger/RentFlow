package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jeckersberger/EquipFlow/services/document/internal/domain"

	apperrors "github.com/jeckersberger/EquipFlow/pkg/common/errors"
)

// templateColumns lists all columns of the document_templates table for consistent scanning.
const templateColumns = `
	id, tenant_id, name, type, content, variables,
	is_default, is_active, created_at, updated_at`

// TemplateRepo implements domain.DocumentTemplateRepository using PostgreSQL.
type TemplateRepo struct {
	pool *pgxpool.Pool
}

// NewTemplateRepo creates a new TemplateRepo.
func NewTemplateRepo(pool *pgxpool.Pool) *TemplateRepo {
	return &TemplateRepo{pool: pool}
}

// scanTemplate scans a single document_templates row into a domain.DocumentTemplate.
func scanTemplate(row pgx.Row) (*domain.DocumentTemplate, error) {
	t := &domain.DocumentTemplate{}
	err := row.Scan(
		&t.ID, &t.TenantID, &t.Name, &t.Type, &t.Content, &t.Variables,
		&t.IsDefault, &t.IsActive, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return t, nil
}

// Create inserts a new document template.
func (r *TemplateRepo) Create(ctx context.Context, tpl *domain.DocumentTemplate) error {
	query := `
		INSERT INTO document_templates (
			id, tenant_id, name, type, content, variables,
			is_default, is_active
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING created_at, updated_at`

	err := r.pool.QueryRow(ctx, query,
		tpl.ID, tpl.TenantID, tpl.Name, tpl.Type, tpl.Content, tpl.Variables,
		tpl.IsDefault, tpl.IsActive,
	).Scan(&tpl.CreatedAt, &tpl.UpdatedAt)
	if err != nil {
		return fmt.Errorf("template_repo: create: %w", err)
	}
	return nil
}

// GetByID retrieves a document template by its primary key within a tenant scope.
func (r *TemplateRepo) GetByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*domain.DocumentTemplate, error) {
	query := fmt.Sprintf(`SELECT %s FROM document_templates WHERE id = $1 AND tenant_id = $2`, templateColumns)
	t, err := scanTemplate(r.pool.QueryRow(ctx, query, id, tenantID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("template_repo: get_by_id: %w", err)
	}
	return t, nil
}

// List returns all document templates for a tenant.
func (r *TemplateRepo) List(ctx context.Context, tenantID uuid.UUID) ([]*domain.DocumentTemplate, error) {
	query := fmt.Sprintf(`SELECT %s FROM document_templates WHERE tenant_id = $1 ORDER BY created_at DESC`, templateColumns)

	rows, err := r.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("template_repo: list query: %w", err)
	}
	defer rows.Close()

	var items []*domain.DocumentTemplate
	for rows.Next() {
		t, scanErr := scanTemplate(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("template_repo: list scan: %w", scanErr)
		}
		items = append(items, t)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("template_repo: list rows: %w", err)
	}
	return items, nil
}

// Update updates an existing document template.
func (r *TemplateRepo) Update(ctx context.Context, tpl *domain.DocumentTemplate) error {
	query := `
		UPDATE document_templates SET
			name = $3, type = $4, content = $5, variables = $6,
			is_default = $7, is_active = $8, updated_at = NOW()
		WHERE id = $1 AND tenant_id = $2
		RETURNING updated_at`

	err := r.pool.QueryRow(ctx, query,
		tpl.ID, tpl.TenantID, tpl.Name, tpl.Type, tpl.Content, tpl.Variables,
		tpl.IsDefault, tpl.IsActive,
	).Scan(&tpl.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apperrors.ErrNotFound
		}
		return fmt.Errorf("template_repo: update: %w", err)
	}
	return nil
}

// Ensure interface compliance at compile time.
var _ domain.DocumentTemplateRepository = (*TemplateRepo)(nil)

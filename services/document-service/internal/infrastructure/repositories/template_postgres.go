package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/lib/pq"
	"github.com/jeckersberger/rentflow/pkg/common/database"
	"github.com/jeckersberger/rentflow/services/document-service/internal/domain"
)

type TemplatePostgres struct {
	db *database.PostgresPool
}

func NewTemplatePostgres(db *database.PostgresPool) *TemplatePostgres {
	return &TemplatePostgres{db: db}
}

func (r *TemplatePostgres) Create(ctx context.Context, tpl *domain.Template) error {
	query := `
		INSERT INTO documents.templates (
			id, tenant_id, name, type, content, variables, is_default, created_at, updated_at, created_by
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		)
	`

	vars := pq.StringArray(tpl.Variables)

	_, err := r.db.Exec(ctx, query,
		tpl.ID, tpl.TenantID, tpl.Name, string(tpl.Type), tpl.Content, vars, tpl.IsDefault,
		tpl.CreatedAt, tpl.UpdatedAt, tpl.CreatedBy,
	)

	if err != nil {
		return fmt.Errorf("failed to create template: %w", err)
	}

	return nil
}

func (r *TemplatePostgres) Update(ctx context.Context, tpl *domain.Template) error {
	query := `
		UPDATE documents.templates SET
			name = $3, type = $4, content = $5, variables = $6, is_default = $7, updated_at = $8
		WHERE id = $1 AND tenant_id = $2
	`

	vars := pq.StringArray(tpl.Variables)

	result, err := r.db.Exec(ctx, query,
		tpl.ID, tpl.TenantID, tpl.Name, string(tpl.Type), tpl.Content, vars, tpl.IsDefault, tpl.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update template: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("template not found")
	}

	return nil
}

func (r *TemplatePostgres) GetByID(ctx context.Context, tenantID, templateID string) (*domain.Template, error) {
	query := `
		SELECT id, tenant_id, name, type, content, variables, is_default, created_at, updated_at, created_by
		FROM documents.templates
		WHERE id = $1 AND tenant_id = $2
	`

	var tpl domain.Template
	var vars pq.StringArray

	err := r.db.QueryRow(ctx, query, templateID, tenantID).Scan(
		&tpl.ID, &tpl.TenantID, &tpl.Name, (*string)(&tpl.Type), &tpl.Content, &vars, &tpl.IsDefault,
		&tpl.CreatedAt, &tpl.UpdatedAt, &tpl.CreatedBy,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("template not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	tpl.Variables = fromStringArray(vars)

	return &tpl, nil
}

func (r *TemplatePostgres) List(ctx context.Context, tenantID string, limit, offset int) ([]*domain.Template, int64, error) {
	countQuery := `SELECT COUNT(*) FROM documents.templates WHERE tenant_id = $1`
	var total int64
	if err := r.db.QueryRow(ctx, countQuery, tenantID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count templates: %w", err)
	}

	query := `
		SELECT id, tenant_id, name, type, content, variables, is_default, created_at, updated_at, created_by
		FROM documents.templates
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list templates: %w", err)
	}
	defer rows.Close()

	var tpls []*domain.Template
	for rows.Next() {
		var tpl domain.Template
		var vars pq.StringArray

		err := rows.Scan(
			&tpl.ID, &tpl.TenantID, &tpl.Name, (*string)(&tpl.Type), &tpl.Content, &vars, &tpl.IsDefault,
			&tpl.CreatedAt, &tpl.UpdatedAt, &tpl.CreatedBy,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan template: %w", err)
		}

		tpl.Variables = fromStringArray(vars)
		tpls = append(tpls, &tpl)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows error: %w", err)
	}

	return tpls, total, nil
}

func (r *TemplatePostgres) Delete(ctx context.Context, tenantID, templateID string) error {
	query := `DELETE FROM documents.templates WHERE id = $1 AND tenant_id = $2`

	result, err := r.db.Exec(ctx, query, templateID, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete template: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("template not found")
	}

	return nil
}

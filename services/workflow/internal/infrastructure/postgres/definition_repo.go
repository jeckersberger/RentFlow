package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	apperrors "github.com/jeckersberger/EquipFlow/pkg/common/errors"
	"github.com/jeckersberger/EquipFlow/services/workflow/internal/domain"
)

// definitionColumns lists all columns of the workflow_definitions table for consistent scanning.
const definitionColumns = `
	id, tenant_id, name, type, steps, is_active,
	created_by, created_at, updated_at`

// DefinitionRepo implements domain.WorkflowDefinitionRepository using PostgreSQL.
type DefinitionRepo struct {
	pool *pgxpool.Pool
}

// NewDefinitionRepo creates a new DefinitionRepo.
func NewDefinitionRepo(pool *pgxpool.Pool) *DefinitionRepo {
	return &DefinitionRepo{pool: pool}
}

// scanDefinition scans a single workflow_definitions row into a domain.WorkflowDefinition.
func scanDefinition(row pgx.Row) (*domain.WorkflowDefinition, error) {
	d := &domain.WorkflowDefinition{}
	var (
		name     *string
		wfType   *string
		isActive *bool
	)

	err := row.Scan(
		&d.ID, &d.TenantID, &name, &wfType, &d.Steps, &isActive,
		&d.CreatedBy, &d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if name != nil {
		d.Name = *name
	}
	if wfType != nil {
		d.Type = *wfType
	}
	if isActive != nil {
		d.IsActive = *isActive
	}

	return d, nil
}

// Create inserts a new workflow definition record.
func (r *DefinitionRepo) Create(ctx context.Context, def *domain.WorkflowDefinition) error {
	query := `
		INSERT INTO workflow_definitions (
			id, tenant_id, name, type, steps, is_active, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at, updated_at`

	err := r.pool.QueryRow(ctx, query,
		def.ID, def.TenantID, def.Name, def.Type,
		def.Steps, def.IsActive, def.CreatedBy,
	).Scan(&def.CreatedAt, &def.UpdatedAt)
	if err != nil {
		return fmt.Errorf("definition_repo: create: %w", err)
	}
	return nil
}

// GetByID retrieves a workflow definition by its primary key within a tenant scope.
func (r *DefinitionRepo) GetByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*domain.WorkflowDefinition, error) {
	query := fmt.Sprintf(`SELECT %s FROM workflow_definitions WHERE id = $1 AND tenant_id = $2`, definitionColumns)
	d, err := scanDefinition(r.pool.QueryRow(ctx, query, id, tenantID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("definition_repo: get_by_id: %w", err)
	}
	return d, nil
}

// List returns a paginated list of workflow definitions for a tenant plus total count.
func (r *DefinitionRepo) List(ctx context.Context, tenantID uuid.UUID, filter domain.DefinitionFilter) ([]*domain.WorkflowDefinition, int64, error) {
	// Count query.
	var total int64
	countQuery := `SELECT COUNT(*) FROM workflow_definitions WHERE tenant_id = $1`
	err := r.pool.QueryRow(ctx, countQuery, tenantID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("definition_repo: list count: %w", err)
	}

	// Data query with pagination.
	offset := (filter.Page - 1) * filter.PerPage
	dataQuery := fmt.Sprintf(
		`SELECT %s FROM workflow_definitions WHERE tenant_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		definitionColumns,
	)

	rows, err := r.pool.Query(ctx, dataQuery, tenantID, filter.PerPage, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("definition_repo: list query: %w", err)
	}
	defer rows.Close()

	var items []*domain.WorkflowDefinition
	for rows.Next() {
		d, scanErr := scanDefinition(rows)
		if scanErr != nil {
			return nil, 0, fmt.Errorf("definition_repo: list scan: %w", scanErr)
		}
		items = append(items, d)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("definition_repo: list rows: %w", err)
	}
	return items, total, nil
}

// Update updates an existing workflow definition.
func (r *DefinitionRepo) Update(ctx context.Context, def *domain.WorkflowDefinition) error {
	query := `
		UPDATE workflow_definitions
		SET name = $3, type = $4, steps = $5, is_active = $6, updated_at = NOW()
		WHERE id = $1 AND tenant_id = $2
		RETURNING updated_at`

	err := r.pool.QueryRow(ctx, query,
		def.ID, def.TenantID, def.Name, def.Type,
		def.Steps, def.IsActive,
	).Scan(&def.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apperrors.ErrNotFound
		}
		return fmt.Errorf("definition_repo: update: %w", err)
	}
	return nil
}

// Ensure interface compliance at compile time.
var _ domain.WorkflowDefinitionRepository = (*DefinitionRepo)(nil)

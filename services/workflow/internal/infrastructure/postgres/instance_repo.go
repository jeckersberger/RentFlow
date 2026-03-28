package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	apperrors "github.com/jeckersberger/EquipFlow/pkg/common/errors"
	"github.com/jeckersberger/EquipFlow/services/workflow/internal/domain"
)

// instanceColumns lists all columns of the workflow_instances table for consistent scanning.
const instanceColumns = `
	id, definition_id, tenant_id, reference_id, reference_type,
	current_step, status, data, started_by, started_at, completed_at, notes`

// InstanceRepo implements domain.WorkflowInstanceRepository using PostgreSQL.
type InstanceRepo struct {
	pool *pgxpool.Pool
}

// NewInstanceRepo creates a new InstanceRepo.
func NewInstanceRepo(pool *pgxpool.Pool) *InstanceRepo {
	return &InstanceRepo{pool: pool}
}

// scanInstance scans a single workflow_instances row into a domain.WorkflowInstance.
func scanInstance(row pgx.Row) (*domain.WorkflowInstance, error) {
	i := &domain.WorkflowInstance{}
	var (
		referenceType *string
		status        *string
		notes         *string
	)

	err := row.Scan(
		&i.ID, &i.DefinitionID, &i.TenantID, &i.ReferenceID, &referenceType,
		&i.CurrentStep, &status, &i.Data, &i.StartedBy, &i.StartedAt,
		&i.CompletedAt, &notes,
	)
	if err != nil {
		return nil, err
	}

	if referenceType != nil {
		i.ReferenceType = *referenceType
	}
	if status != nil {
		i.Status = *status
	}
	if notes != nil {
		i.Notes = *notes
	}

	return i, nil
}

// Create inserts a new workflow instance record.
func (r *InstanceRepo) Create(ctx context.Context, inst *domain.WorkflowInstance) error {
	query := `
		INSERT INTO workflow_instances (
			id, definition_id, tenant_id, reference_id, reference_type,
			current_step, status, data, started_by, notes
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING started_at`

	err := r.pool.QueryRow(ctx, query,
		inst.ID, inst.DefinitionID, inst.TenantID,
		inst.ReferenceID, nilIfEmpty(inst.ReferenceType),
		inst.CurrentStep, inst.Status, inst.Data,
		inst.StartedBy, nilIfEmpty(inst.Notes),
	).Scan(&inst.StartedAt)
	if err != nil {
		return fmt.Errorf("instance_repo: create: %w", err)
	}
	return nil
}

// GetByID retrieves a workflow instance by its primary key within a tenant scope.
func (r *InstanceRepo) GetByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*domain.WorkflowInstance, error) {
	query := fmt.Sprintf(`SELECT %s FROM workflow_instances WHERE id = $1 AND tenant_id = $2`, instanceColumns)
	i, err := scanInstance(r.pool.QueryRow(ctx, query, id, tenantID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("instance_repo: get_by_id: %w", err)
	}
	return i, nil
}

// List returns a filtered, paginated list of workflow instances for a tenant plus total count.
func (r *InstanceRepo) List(ctx context.Context, tenantID uuid.UUID, filter domain.InstanceFilter) ([]*domain.WorkflowInstance, int64, error) {
	// Build dynamic WHERE clause.
	conditions := []string{"tenant_id = $1"}
	args := []interface{}{tenantID}
	argIdx := 2

	if filter.ReferenceID != nil {
		conditions = append(conditions, fmt.Sprintf("reference_id = $%d", argIdx))
		args = append(args, *filter.ReferenceID)
		argIdx++
	}
	if filter.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, *filter.Status)
		argIdx++
	}

	where := strings.Join(conditions, " AND ")

	// Count query.
	var total int64
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM workflow_instances WHERE %s`, where)
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("instance_repo: list count: %w", err)
	}

	// Data query with pagination.
	offset := (filter.Page - 1) * filter.PerPage
	dataQuery := fmt.Sprintf(
		`SELECT %s FROM workflow_instances WHERE %s ORDER BY started_at DESC LIMIT $%d OFFSET $%d`,
		instanceColumns, where, argIdx, argIdx+1,
	)
	args = append(args, filter.PerPage, offset)

	rows, err := r.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("instance_repo: list query: %w", err)
	}
	defer rows.Close()

	var items []*domain.WorkflowInstance
	for rows.Next() {
		i, scanErr := scanInstance(rows)
		if scanErr != nil {
			return nil, 0, fmt.Errorf("instance_repo: list scan: %w", scanErr)
		}
		items = append(items, i)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("instance_repo: list rows: %w", err)
	}
	return items, total, nil
}

// Update updates an existing workflow instance (status, current_step, completed_at, notes).
func (r *InstanceRepo) Update(ctx context.Context, inst *domain.WorkflowInstance) error {
	query := `
		UPDATE workflow_instances
		SET current_step = $3, status = $4, data = $5, completed_at = $6, notes = $7
		WHERE id = $1 AND tenant_id = $2`

	tag, err := r.pool.Exec(ctx, query,
		inst.ID, inst.TenantID,
		inst.CurrentStep, inst.Status, inst.Data,
		inst.CompletedAt, nilIfEmpty(inst.Notes),
	)
	if err != nil {
		return fmt.Errorf("instance_repo: update: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

// nilIfEmpty returns nil if the string is empty, otherwise a pointer to it.
func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// Ensure interface compliance at compile time.
var _ domain.WorkflowInstanceRepository = (*InstanceRepo)(nil)

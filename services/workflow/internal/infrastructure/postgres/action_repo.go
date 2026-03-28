package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jeckersberger/EquipFlow/services/workflow/internal/domain"
)

// actionColumns lists all columns of the workflow_actions table for consistent scanning.
const actionColumns = `id, instance_id, step, action, performed_by, comment, created_at`

// ActionRepo implements domain.WorkflowActionRepository using PostgreSQL.
type ActionRepo struct {
	pool *pgxpool.Pool
}

// NewActionRepo creates a new ActionRepo.
func NewActionRepo(pool *pgxpool.Pool) *ActionRepo {
	return &ActionRepo{pool: pool}
}

// scanAction scans a single workflow_actions row into a domain.WorkflowAction.
func scanAction(row pgx.Row) (*domain.WorkflowAction, error) {
	a := &domain.WorkflowAction{}
	var (
		action  *string
		comment *string
	)

	err := row.Scan(
		&a.ID, &a.InstanceID, &a.Step, &action,
		&a.PerformedBy, &comment, &a.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	if action != nil {
		a.Action = *action
	}
	if comment != nil {
		a.Comment = *comment
	}

	return a, nil
}

// Create inserts a new workflow action record.
func (r *ActionRepo) Create(ctx context.Context, action *domain.WorkflowAction) error {
	query := `
		INSERT INTO workflow_actions (
			id, instance_id, step, action, performed_by, comment
		) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at`

	err := r.pool.QueryRow(ctx, query,
		action.ID, action.InstanceID, action.Step,
		action.Action, action.PerformedBy, nilIfEmpty(action.Comment),
	).Scan(&action.CreatedAt)
	if err != nil {
		return fmt.Errorf("action_repo: create: %w", err)
	}
	return nil
}

// ListByInstance returns all actions for a workflow instance, ordered by step/time.
func (r *ActionRepo) ListByInstance(ctx context.Context, instanceID uuid.UUID) ([]*domain.WorkflowAction, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM workflow_actions WHERE instance_id = $1 ORDER BY step ASC, created_at ASC`,
		actionColumns,
	)

	rows, err := r.pool.Query(ctx, query, instanceID)
	if err != nil {
		return nil, fmt.Errorf("action_repo: list_by_instance query: %w", err)
	}
	defer rows.Close()

	var items []*domain.WorkflowAction
	for rows.Next() {
		a, scanErr := scanAction(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("action_repo: list_by_instance scan: %w", scanErr)
		}
		items = append(items, a)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("action_repo: list_by_instance rows: %w", err)
	}
	return items, nil
}

// Ensure interface compliance at compile time.
var _ domain.WorkflowActionRepository = (*ActionRepo)(nil)

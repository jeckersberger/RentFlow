package repositories

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/jeckersberger/rentflow/pkg/common/database"
	"github.com/jeckersberger/rentflow/services/workflow-service/internal/domain"
)

type WorkflowPostgres struct {
	db *database.PostgresPool
}

func NewWorkflowPostgres(db *database.PostgresPool) *WorkflowPostgres {
	return &WorkflowPostgres{db: db}
}

func (r *WorkflowPostgres) Create(ctx context.Context, workflow *domain.Workflow) error {
	stepsJSON, _ := json.Marshal(workflow.Steps)
	query := `
		INSERT INTO workflows
		(id, tenant_id, name, description, trigger_event, steps, is_active, version, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.db.Exec(ctx, query,
		workflow.ID, workflow.TenantID, workflow.Name, workflow.Description,
		workflow.TriggerEvent, stepsJSON, workflow.IsActive, workflow.Version,
		workflow.CreatedAt, workflow.UpdatedAt,
	)
	return err
}

func (r *WorkflowPostgres) GetByID(ctx context.Context, tenantID, id string) (*domain.Workflow, error) {
	query := `
		SELECT id, tenant_id, name, description, trigger_event, steps, is_active, version, created_at, updated_at
		FROM workflows
		WHERE id = $1 AND tenant_id = $2
	`
	var wf domain.Workflow
	var stepsJSON []byte

	err := r.db.QueryRow(ctx, query, id, tenantID).Scan(
		&wf.ID, &wf.TenantID, &wf.Name, &wf.Description, &wf.TriggerEvent,
		&stepsJSON, &wf.IsActive, &wf.Version, &wf.CreatedAt, &wf.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	_ = json.Unmarshal(stepsJSON, &wf.Steps)
	return &wf, nil
}

func (r *WorkflowPostgres) ListByTenant(ctx context.Context, tenantID string) ([]*domain.Workflow, error) {
	query := `
		SELECT id, tenant_id, name, description, trigger_event, steps, is_active, version, created_at, updated_at
		FROM workflows
		WHERE tenant_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var workflows []*domain.Workflow
	for rows.Next() {
		var wf domain.Workflow
		var stepsJSON []byte

		err := rows.Scan(
			&wf.ID, &wf.TenantID, &wf.Name, &wf.Description, &wf.TriggerEvent,
			&stepsJSON, &wf.IsActive, &wf.Version, &wf.CreatedAt, &wf.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		_ = json.Unmarshal(stepsJSON, &wf.Steps)
		workflows = append(workflows, &wf)
	}
	return workflows, rows.Err()
}

func (r *WorkflowPostgres) Update(ctx context.Context, workflow *domain.Workflow) error {
	stepsJSON, _ := json.Marshal(workflow.Steps)
	query := `
		UPDATE workflows
		SET name = $1, description = $2, trigger_event = $3, steps = $4,
		    is_active = $5, version = $6, updated_at = $7
		WHERE id = $8 AND tenant_id = $9
	`
	_, err := r.db.Exec(ctx, query,
		workflow.Name, workflow.Description, workflow.TriggerEvent, stepsJSON,
		workflow.IsActive, workflow.Version, workflow.UpdatedAt, workflow.ID, workflow.TenantID,
	)
	return err
}

func (r *WorkflowPostgres) Delete(ctx context.Context, tenantID, id string) error {
	query := `DELETE FROM workflows WHERE id = $1 AND tenant_id = $2`
	_, err := r.db.Exec(ctx, query, id, tenantID)
	return err
}

type WorkflowRunPostgres struct {
	db *database.PostgresPool
}

func NewWorkflowRunPostgres(db *database.PostgresPool) *WorkflowRunPostgres {
	return &WorkflowRunPostgres{db: db}
}

func (r *WorkflowRunPostgres) Create(ctx context.Context, run *domain.WorkflowRun) error {
	query := `
		INSERT INTO workflow_runs
		(id, workflow_id, trigger_event_id, status, current_step, context, started_at, completed_at, error)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.db.Exec(ctx, query,
		run.ID, run.WorkflowID, run.TriggerEventID, run.Status, run.CurrentStep,
		run.Context, run.StartedAt, run.CompletedAt, run.Error,
	)
	return err
}

func (r *WorkflowRunPostgres) GetByID(ctx context.Context, id string) (*domain.WorkflowRun, error) {
	query := `
		SELECT id, workflow_id, trigger_event_id, status, current_step, context, started_at, completed_at, error
		FROM workflow_runs
		WHERE id = $1
	`
	var run domain.WorkflowRun
	err := r.db.QueryRow(ctx, query, id).Scan(
		&run.ID, &run.WorkflowID, &run.TriggerEventID, &run.Status, &run.CurrentStep,
		&run.Context, &run.StartedAt, &run.CompletedAt, &run.Error,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &run, nil
}

func (r *WorkflowRunPostgres) ListByTenant(ctx context.Context, tenantID string) ([]*domain.WorkflowRun, error) {
	query := `
		SELECT wr.id, wr.workflow_id, wr.trigger_event_id, wr.status, wr.current_step, wr.context, wr.started_at, wr.completed_at, wr.error
		FROM workflow_runs wr
		INNER JOIN workflows w ON wr.workflow_id = w.id
		WHERE w.tenant_id = $1
		ORDER BY wr.started_at DESC
	`
	rows, err := r.db.Query(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var runs []*domain.WorkflowRun
	for rows.Next() {
		var run domain.WorkflowRun
		err := rows.Scan(
			&run.ID, &run.WorkflowID, &run.TriggerEventID, &run.Status, &run.CurrentStep,
			&run.Context, &run.StartedAt, &run.CompletedAt, &run.Error,
		)
		if err != nil {
			return nil, err
		}
		runs = append(runs, &run)
	}
	return runs, rows.Err()
}

func (r *WorkflowRunPostgres) ListByWorkflow(ctx context.Context, workflowID string) ([]*domain.WorkflowRun, error) {
	query := `
		SELECT id, workflow_id, trigger_event_id, status, current_step, context, started_at, completed_at, error
		FROM workflow_runs
		WHERE workflow_id = $1
		ORDER BY started_at DESC
	`
	rows, err := r.db.Query(ctx, query, workflowID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var runs []*domain.WorkflowRun
	for rows.Next() {
		var run domain.WorkflowRun
		err := rows.Scan(
			&run.ID, &run.WorkflowID, &run.TriggerEventID, &run.Status, &run.CurrentStep,
			&run.Context, &run.StartedAt, &run.CompletedAt, &run.Error,
		)
		if err != nil {
			return nil, err
		}
		runs = append(runs, &run)
	}
	return runs, rows.Err()
}

func (r *WorkflowRunPostgres) Update(ctx context.Context, run *domain.WorkflowRun) error {
	query := `
		UPDATE workflow_runs
		SET status = $1, current_step = $2, context = $3, completed_at = $4, error = $5
		WHERE id = $6
	`
	_, err := r.db.Exec(ctx, query,
		run.Status, run.CurrentStep, run.Context, run.CompletedAt, run.Error, run.ID,
	)
	return err
}

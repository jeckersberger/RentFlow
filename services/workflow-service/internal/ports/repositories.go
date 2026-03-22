package ports

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/workflow-service/internal/domain"
)

type WorkflowDefinitionRepository interface {
	Create(ctx context.Context, wd *domain.WorkflowDefinition) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.WorkflowDefinition, error)
	ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*domain.WorkflowDefinition, error)
	ListTemplates(ctx context.Context) ([]*domain.WorkflowDefinition, error)
	Update(ctx context.Context, wd *domain.WorkflowDefinition) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type WorkflowInstanceRepository interface {
	Create(ctx context.Context, wi *domain.WorkflowInstance) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.WorkflowInstance, error)
	ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*domain.WorkflowInstance, error)
	ListByDefinition(ctx context.Context, definitionID uuid.UUID) ([]*domain.WorkflowInstance, error)
	Update(ctx context.Context, wi *domain.WorkflowInstance) error
	CountByStatus(ctx context.Context, tenantID uuid.UUID, status domain.Status) (int, error)
	CountCompletedSince(ctx context.Context, tenantID uuid.UUID, since time.Time) (int, error)
}

type WorkflowStepRepository interface {
	Create(ctx context.Context, ws *domain.WorkflowStep) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.WorkflowStep, error)
	ListByDefinition(ctx context.Context, definitionID uuid.UUID) ([]*domain.WorkflowStep, error)
	Update(ctx context.Context, ws *domain.WorkflowStep) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// PostgreSQL Repositories

type postgresWorkflowDefinitionRepository struct {
	db  *sql.DB
	log logger.Logger
}

func NewPostgresWorkflowDefinitionRepository(db *sql.DB, log logger.Logger) WorkflowDefinitionRepository {
	return &postgresWorkflowDefinitionRepository{db: db, log: log}
}

func (r *postgresWorkflowDefinitionRepository) Create(ctx context.Context, wd *domain.WorkflowDefinition) error {
	query := `
		INSERT INTO workflow_definitions 
		(id, tenant_id, name, description, trigger_type, trigger_config, is_active, is_template, template_category, version)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.db.ExecContext(ctx, query,
		wd.ID, wd.TenantID, wd.Name, wd.Description, wd.TriggerType, wd.TriggerConfig,
		wd.IsActive, wd.IsTemplate, wd.TemplateCategory, wd.Version)
	return err
}

func (r *postgresWorkflowDefinitionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.WorkflowDefinition, error) {
	query := `
		SELECT id, tenant_id, name, description, trigger_type, trigger_config, is_active, is_template, template_category, version, created_at, updated_at
		FROM workflow_definitions WHERE id = $1
	`
	wd := &domain.WorkflowDefinition{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&wd.ID, &wd.TenantID, &wd.Name, &wd.Description, &wd.TriggerType, &wd.TriggerConfig,
		&wd.IsActive, &wd.IsTemplate, &wd.TemplateCategory, &wd.Version, &wd.CreatedAt, &wd.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, domain.ErrWorkflowNotFound
	}
	return wd, err
}

func (r *postgresWorkflowDefinitionRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*domain.WorkflowDefinition, error) {
	query := `
		SELECT id, tenant_id, name, description, trigger_type, trigger_config, is_active, is_template, template_category, version, created_at, updated_at
		FROM workflow_definitions WHERE tenant_id = $1 ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var wds []*domain.WorkflowDefinition
	for rows.Next() {
		wd := &domain.WorkflowDefinition{}
		err := rows.Scan(&wd.ID, &wd.TenantID, &wd.Name, &wd.Description, &wd.TriggerType, &wd.TriggerConfig,
			&wd.IsActive, &wd.IsTemplate, &wd.TemplateCategory, &wd.Version, &wd.CreatedAt, &wd.UpdatedAt)
		if err != nil {
			return nil, err
		}
		wds = append(wds, wd)
	}
	return wds, rows.Err()
}

func (r *postgresWorkflowDefinitionRepository) ListTemplates(ctx context.Context) ([]*domain.WorkflowDefinition, error) {
	query := `
		SELECT id, tenant_id, name, description, trigger_type, trigger_config, is_active, is_template, template_category, version, created_at, updated_at
		FROM workflow_definitions WHERE is_template = true AND is_active = true ORDER BY template_category, name
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var wds []*domain.WorkflowDefinition
	for rows.Next() {
		wd := &domain.WorkflowDefinition{}
		err := rows.Scan(&wd.ID, &wd.TenantID, &wd.Name, &wd.Description, &wd.TriggerType, &wd.TriggerConfig,
			&wd.IsActive, &wd.IsTemplate, &wd.TemplateCategory, &wd.Version, &wd.CreatedAt, &wd.UpdatedAt)
		if err != nil {
			return nil, err
		}
		wds = append(wds, wd)
	}
	return wds, rows.Err()
}

func (r *postgresWorkflowDefinitionRepository) Update(ctx context.Context, wd *domain.WorkflowDefinition) error {
	query := `
		UPDATE workflow_definitions 
		SET name = $1, description = $2, trigger_type = $3, trigger_config = $4, is_active = $5, version = $6, updated_at = NOW()
		WHERE id = $7
	`
	_, err := r.db.ExecContext(ctx, query,
		wd.Name, wd.Description, wd.TriggerType, wd.TriggerConfig, wd.IsActive, wd.Version, wd.ID)
	return err
}

func (r *postgresWorkflowDefinitionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := "DELETE FROM workflow_definitions WHERE id = $1"
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

type postgresWorkflowInstanceRepository struct {
	db  *sql.DB
	log logger.Logger
}

func NewPostgresWorkflowInstanceRepository(db *sql.DB, log logger.Logger) WorkflowInstanceRepository {
	return &postgresWorkflowInstanceRepository{db: db, log: log}
}

func (r *postgresWorkflowInstanceRepository) Create(ctx context.Context, wi *domain.WorkflowInstance) error {
	query := `
		INSERT INTO workflow_instances (id, tenant_id, definition_id, status, trigger_data, context_data, current_step_index)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.ExecContext(ctx, query,
		wi.ID, wi.TenantID, wi.DefinitionID, wi.Status, wi.TriggerData, wi.ContextData, wi.CurrentStepIdx)
	return err
}

func (r *postgresWorkflowInstanceRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.WorkflowInstance, error) {
	query := `
		SELECT id, tenant_id, definition_id, status, trigger_data, context_data, current_step_index, started_at, completed_at, error_message
		FROM workflow_instances WHERE id = $1
	`
	wi := &domain.WorkflowInstance{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&wi.ID, &wi.TenantID, &wi.DefinitionID, &wi.Status, &wi.TriggerData, &wi.ContextData,
		&wi.CurrentStepIdx, &wi.StartedAt, &wi.CompletedAt, &wi.ErrorMessage)
	if err == sql.ErrNoRows {
		return nil, domain.ErrWorkflowInstanceNotFound
	}
	return wi, err
}

func (r *postgresWorkflowInstanceRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*domain.WorkflowInstance, error) {
	query := `
		SELECT id, tenant_id, definition_id, status, trigger_data, context_data, current_step_index, started_at, completed_at, error_message
		FROM workflow_instances WHERE tenant_id = $1 ORDER BY started_at DESC LIMIT 100
	`
	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var wis []*domain.WorkflowInstance
	for rows.Next() {
		wi := &domain.WorkflowInstance{}
		err := rows.Scan(&wi.ID, &wi.TenantID, &wi.DefinitionID, &wi.Status, &wi.TriggerData, &wi.ContextData,
			&wi.CurrentStepIdx, &wi.StartedAt, &wi.CompletedAt, &wi.ErrorMessage)
		if err != nil {
			return nil, err
		}
		wis = append(wis, wi)
	}
	return wis, rows.Err()
}

func (r *postgresWorkflowInstanceRepository) ListByDefinition(ctx context.Context, definitionID uuid.UUID) ([]*domain.WorkflowInstance, error) {
	query := `
		SELECT id, tenant_id, definition_id, status, trigger_data, context_data, current_step_index, started_at, completed_at, error_message
		FROM workflow_instances WHERE definition_id = $1 ORDER BY started_at DESC LIMIT 100
	`
	rows, err := r.db.QueryContext(ctx, query, definitionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var wis []*domain.WorkflowInstance
	for rows.Next() {
		wi := &domain.WorkflowInstance{}
		err := rows.Scan(&wi.ID, &wi.TenantID, &wi.DefinitionID, &wi.Status, &wi.TriggerData, &wi.ContextData,
			&wi.CurrentStepIdx, &wi.StartedAt, &wi.CompletedAt, &wi.ErrorMessage)
		if err != nil {
			return nil, err
		}
		wis = append(wis, wi)
	}
	return wis, rows.Err()
}

func (r *postgresWorkflowInstanceRepository) Update(ctx context.Context, wi *domain.WorkflowInstance) error {
	query := `
		UPDATE workflow_instances 
		SET status = $1, trigger_data = $2, context_data = $3, current_step_index = $4, completed_at = $5, error_message = $6
		WHERE id = $7
	`
	_, err := r.db.ExecContext(ctx, query,
		wi.Status, wi.TriggerData, wi.ContextData, wi.CurrentStepIdx, wi.CompletedAt, wi.ErrorMessage, wi.ID)
	return err
}

func (r *postgresWorkflowInstanceRepository) CountByStatus(ctx context.Context, tenantID uuid.UUID, status domain.Status) (int, error) {
	query := "SELECT COUNT(*) FROM workflow_instances WHERE tenant_id = $1 AND status = $2"
	var count int
	err := r.db.QueryRowContext(ctx, query, tenantID, status).Scan(&count)
	return count, err
}

func (r *postgresWorkflowInstanceRepository) CountCompletedSince(ctx context.Context, tenantID uuid.UUID, since time.Time) (int, error) {
	query := "SELECT COUNT(*) FROM workflow_instances WHERE tenant_id = $1 AND status = $2 AND completed_at >= $3"
	var count int
	err := r.db.QueryRowContext(ctx, query, tenantID, domain.StatusCompleted, since).Scan(&count)
	return count, err
}

type postgresWorkflowStepRepository struct {
	db  *sql.DB
	log logger.Logger
}

func NewPostgresWorkflowStepRepository(db *sql.DB, log logger.Logger) WorkflowStepRepository {
	return &postgresWorkflowStepRepository{db: db, log: log}
}

func (r *postgresWorkflowStepRepository) Create(ctx context.Context, ws *domain.WorkflowStep) error {
	query := `
		INSERT INTO workflow_steps (id, definition_id, step_index, name, action_type, action_config, on_success_step, on_failure_step, timeout_seconds, retry_count)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.db.ExecContext(ctx, query,
		ws.ID, ws.DefinitionID, ws.StepIndex, ws.Name, ws.ActionType, ws.ActionConfig,
		ws.OnSuccessStep, ws.OnFailureStep, ws.TimeoutSecs, ws.RetryCount)
	return err
}

func (r *postgresWorkflowStepRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.WorkflowStep, error) {
	query := `
		SELECT id, definition_id, step_index, name, action_type, action_config, on_success_step, on_failure_step, timeout_seconds, retry_count, created_at
		FROM workflow_steps WHERE id = $1
	`
	ws := &domain.WorkflowStep{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&ws.ID, &ws.DefinitionID, &ws.StepIndex, &ws.Name, &ws.ActionType, &ws.ActionConfig,
		&ws.OnSuccessStep, &ws.OnFailureStep, &ws.TimeoutSecs, &ws.RetryCount, &ws.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, domain.ErrWorkflowStepNotFound
	}
	return ws, err
}

func (r *postgresWorkflowStepRepository) ListByDefinition(ctx context.Context, definitionID uuid.UUID) ([]*domain.WorkflowStep, error) {
	query := `
		SELECT id, definition_id, step_index, name, action_type, action_config, on_success_step, on_failure_step, timeout_seconds, retry_count, created_at
		FROM workflow_steps WHERE definition_id = $1 ORDER BY step_index ASC
	`
	rows, err := r.db.QueryContext(ctx, query, definitionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var wss []*domain.WorkflowStep
	for rows.Next() {
		ws := &domain.WorkflowStep{}
		err := rows.Scan(&ws.ID, &ws.DefinitionID, &ws.StepIndex, &ws.Name, &ws.ActionType, &ws.ActionConfig,
			&ws.OnSuccessStep, &ws.OnFailureStep, &ws.TimeoutSecs, &ws.RetryCount, &ws.CreatedAt)
		if err != nil {
			return nil, err
		}
		wss = append(wss, ws)
	}
	return wss, rows.Err()
}

func (r *postgresWorkflowStepRepository) Update(ctx context.Context, ws *domain.WorkflowStep) error {
	query := `
		UPDATE workflow_steps 
		SET name = $1, action_type = $2, action_config = $3, on_success_step = $4, on_failure_step = $5, timeout_seconds = $6, retry_count = $7
		WHERE id = $8
	`
	_, err := r.db.ExecContext(ctx, query,
		ws.Name, ws.ActionType, ws.ActionConfig, ws.OnSuccessStep, ws.OnFailureStep, ws.TimeoutSecs, ws.RetryCount, ws.ID)
	return err
}

func (r *postgresWorkflowStepRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := "DELETE FROM workflow_steps WHERE id = $1"
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

package persistence

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jeckersberger/rentflow/services/reporting-service/internal/domain"
)

type PostgresReportRepository struct {
	db *sql.DB
}

func NewPostgresReportRepository(db *sql.DB) *PostgresReportRepository {
	return &PostgresReportRepository{db: db}
}

func (r *PostgresReportRepository) CreateDefinition(ctx context.Context, def *domain.ReportDefinition) error {
	query := `
		INSERT INTO report_definitions (
			id, tenant_id, name, report_type, description, parameters, 
			schedule_cron, email_recipients, format, is_active, created_by, 
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	_, err := r.db.ExecContext(ctx, query,
		def.ID, def.TenantID, def.Name, def.ReportType, def.Description,
		def.Parameters, def.ScheduleCron, def.EmailRecipients, def.Format,
		def.IsActive, def.CreatedBy, def.CreatedAt, def.UpdatedAt,
	)

	return err
}

func (r *PostgresReportRepository) UpdateDefinition(ctx context.Context, def *domain.ReportDefinition) error {
	query := `
		UPDATE report_definitions SET 
			name = $1, description = $2, parameters = $3, schedule_cron = $4,
			email_recipients = $5, format = $6, is_active = $7, updated_at = $8
		WHERE id = $9 AND tenant_id = $10
	`

	result, err := r.db.ExecContext(ctx, query,
		def.Name, def.Description, def.Parameters, def.ScheduleCron,
		def.EmailRecipients, def.Format, def.IsActive, def.UpdatedAt,
		def.ID, def.TenantID,
	)

	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *PostgresReportRepository) GetDefinitionByID(ctx context.Context, tenantID, defID uuid.UUID) (*domain.ReportDefinition, error) {
	query := `
		SELECT id, tenant_id, name, report_type, description, parameters,
			   schedule_cron, email_recipients, format, is_active, created_by,
			   created_at, updated_at
		FROM report_definitions
		WHERE id = $1 AND tenant_id = $2
	`

	def := &domain.ReportDefinition{}
	err := r.db.QueryRowContext(ctx, query, defID, tenantID).Scan(
		&def.ID, &def.TenantID, &def.Name, &def.ReportType, &def.Description,
		&def.Parameters, &def.ScheduleCron, &def.EmailRecipients, &def.Format,
		&def.IsActive, &def.CreatedBy, &def.CreatedAt, &def.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return def, nil
}

func (r *PostgresReportRepository) ListDefinitions(ctx context.Context, tenantID uuid.UUID) ([]domain.ReportDefinition, error) {
	query := `
		SELECT id, tenant_id, name, report_type, description, parameters,
			   schedule_cron, email_recipients, format, is_active, created_by,
			   created_at, updated_at
		FROM report_definitions
		WHERE tenant_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var definitions []domain.ReportDefinition
	for rows.Next() {
		def := domain.ReportDefinition{}
		err := rows.Scan(
			&def.ID, &def.TenantID, &def.Name, &def.ReportType, &def.Description,
			&def.Parameters, &def.ScheduleCron, &def.EmailRecipients, &def.Format,
			&def.IsActive, &def.CreatedBy, &def.CreatedAt, &def.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		definitions = append(definitions, def)
	}

	return definitions, rows.Err()
}

func (r *PostgresReportRepository) CreateRun(ctx context.Context, run *domain.ReportRun) error {
	query := `
		INSERT INTO report_runs (
			id, tenant_id, report_definition_id, status, parameters_used,
			period_start, period_end, file_path, file_size, error_message,
			started_at, completed_at, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	_, err := r.db.ExecContext(ctx, query,
		run.ID, run.TenantID, run.ReportDefinitionID, run.Status, run.ParametersUsed,
		run.PeriodStart, run.PeriodEnd, run.FilePath, run.FileSize, run.ErrorMessage,
		run.StartedAt, run.CompletedAt, run.CreatedAt,
	)

	return err
}

func (r *PostgresReportRepository) UpdateRun(ctx context.Context, run *domain.ReportRun) error {
	query := `
		UPDATE report_runs SET 
			status = $1, file_path = $2, file_size = $3, error_message = $4,
			started_at = $5, completed_at = $6
		WHERE id = $7 AND tenant_id = $8
	`

	result, err := r.db.ExecContext(ctx, query,
		run.Status, run.FilePath, run.FileSize, run.ErrorMessage,
		run.StartedAt, run.CompletedAt, run.ID, run.TenantID,
	)

	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *PostgresReportRepository) GetRunByID(ctx context.Context, tenantID, runID uuid.UUID) (*domain.ReportRun, error) {
	query := `
		SELECT id, tenant_id, report_definition_id, status, parameters_used,
			   period_start, period_end, file_path, file_size, error_message,
			   started_at, completed_at, created_at
		FROM report_runs
		WHERE id = $1 AND tenant_id = $2
	`

	run := &domain.ReportRun{}
	err := r.db.QueryRowContext(ctx, query, runID, tenantID).Scan(
		&run.ID, &run.TenantID, &run.ReportDefinitionID, &run.Status, &run.ParametersUsed,
		&run.PeriodStart, &run.PeriodEnd, &run.FilePath, &run.FileSize, &run.ErrorMessage,
		&run.StartedAt, &run.CompletedAt, &run.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return run, nil
}

func (r *PostgresReportRepository) ListRuns(ctx context.Context, tenantID, defID uuid.UUID) ([]domain.ReportRun, error) {
	var query string
	var args []interface{}

	if defID == uuid.Nil {
		query = `
			SELECT id, tenant_id, report_definition_id, status, parameters_used,
				   period_start, period_end, file_path, file_size, error_message,
				   started_at, completed_at, created_at
			FROM report_runs
			WHERE tenant_id = $1
			ORDER BY created_at DESC
		`
		args = []interface{}{tenantID}
	} else {
		query = `
			SELECT id, tenant_id, report_definition_id, status, parameters_used,
				   period_start, period_end, file_path, file_size, error_message,
				   started_at, completed_at, created_at
			FROM report_runs
			WHERE tenant_id = $1 AND report_definition_id = $2
			ORDER BY created_at DESC
		`
		args = []interface{}{tenantID, defID}
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var runs []domain.ReportRun
	for rows.Next() {
		run := domain.ReportRun{}
		err := rows.Scan(
			&run.ID, &run.TenantID, &run.ReportDefinitionID, &run.Status, &run.ParametersUsed,
			&run.PeriodStart, &run.PeriodEnd, &run.FilePath, &run.FileSize, &run.ErrorMessage,
			&run.StartedAt, &run.CompletedAt, &run.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan run: %w", err)
		}
		runs = append(runs, run)
	}

	return runs, rows.Err()
}

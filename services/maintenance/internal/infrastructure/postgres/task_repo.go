package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	apperrors "github.com/jeckersberger/EquipFlow/pkg/common/errors"
	"github.com/jeckersberger/EquipFlow/services/maintenance/internal/domain"
)

const taskColumns = `
	id, tenant_id, schedule_id, equipment_id, title, description,
	status, priority, assigned_to, due_date, completed_at,
	cost, notes, created_by, created_at, updated_at`

type TaskRepo struct {
	pool *pgxpool.Pool
}

func NewTaskRepo(pool *pgxpool.Pool) *TaskRepo {
	return &TaskRepo{pool: pool}
}

func scanTask(row pgx.Row) (*domain.MaintenanceTask, error) {
	t := &domain.MaintenanceTask{}
	var (
		description *string
		dueDate     *time.Time
		notes       *string
	)

	err := row.Scan(
		&t.ID, &t.TenantID, &t.ScheduleID, &t.EquipmentID,
		&t.Title, &description,
		&t.Status, &t.Priority, &t.AssignedTo,
		&dueDate, &t.CompletedAt,
		&t.Cost, &notes, &t.CreatedBy,
		&t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if description != nil {
		t.Description = *description
	}
	if dueDate != nil {
		t.DueDate = dueDate.Format("2006-01-02")
	}
	if notes != nil {
		t.Notes = *notes
	}

	return t, nil
}

func (r *TaskRepo) Create(ctx context.Context, task *domain.MaintenanceTask) error {
	query := `
		INSERT INTO maintenance_tasks (
			id, tenant_id, schedule_id, equipment_id, title, description,
			status, priority, assigned_to, due_date,
			cost, notes, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING created_at, updated_at`

	err := r.pool.QueryRow(ctx, query,
		task.ID, task.TenantID, task.ScheduleID, task.EquipmentID,
		task.Title, nilIfEmpty(task.Description),
		task.Status, task.Priority, task.AssignedTo,
		nilIfEmpty(task.DueDate),
		task.Cost, nilIfEmpty(task.Notes), task.CreatedBy,
	).Scan(&task.CreatedAt, &task.UpdatedAt)
	if err != nil {
		return fmt.Errorf("task_repo: create: %w", err)
	}
	return nil
}

func (r *TaskRepo) GetByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*domain.MaintenanceTask, error) {
	query := fmt.Sprintf(`SELECT %s FROM maintenance_tasks WHERE id = $1 AND tenant_id = $2`, taskColumns)
	task, err := scanTask(r.pool.QueryRow(ctx, query, id, tenantID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("task_repo: get_by_id: %w", err)
	}
	return task, nil
}

func (r *TaskRepo) List(ctx context.Context, tenantID uuid.UUID, filter domain.TaskFilter) ([]*domain.MaintenanceTask, int64, error) {
	conditions := []string{"tenant_id = $1"}
	args := []interface{}{tenantID}
	argIdx := 2

	if filter.EquipmentID != nil {
		conditions = append(conditions, fmt.Sprintf("equipment_id = $%d", argIdx))
		args = append(args, *filter.EquipmentID)
		argIdx++
	}
	if filter.Status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, filter.Status)
		argIdx++
	}

	where := strings.Join(conditions, " AND ")

	var total int64
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM maintenance_tasks WHERE %s`, where)
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("task_repo: list count: %w", err)
	}

	page := filter.Page
	if page < 1 {
		page = 1
	}
	perPage := filter.PerPage
	if perPage < 1 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	dataQuery := fmt.Sprintf(
		`SELECT %s FROM maintenance_tasks WHERE %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		taskColumns, where, argIdx, argIdx+1,
	)
	args = append(args, perPage, offset)

	rows, err := r.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("task_repo: list query: %w", err)
	}
	defer rows.Close()

	var items []*domain.MaintenanceTask
	for rows.Next() {
		task, scanErr := scanTask(rows)
		if scanErr != nil {
			return nil, 0, fmt.Errorf("task_repo: list scan: %w", scanErr)
		}
		items = append(items, task)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("task_repo: list rows: %w", err)
	}
	return items, total, nil
}

func (r *TaskRepo) Update(ctx context.Context, task *domain.MaintenanceTask) error {
	query := `
		UPDATE maintenance_tasks SET
			title = $3, description = $4, status = $5, priority = $6,
			assigned_to = $7, due_date = $8, completed_at = $9,
			cost = $10, notes = $11, updated_at = NOW()
		WHERE id = $1 AND tenant_id = $2
		RETURNING updated_at`

	err := r.pool.QueryRow(ctx, query,
		task.ID, task.TenantID,
		task.Title, nilIfEmpty(task.Description),
		task.Status, task.Priority,
		task.AssignedTo, nilIfEmpty(task.DueDate),
		task.CompletedAt, task.Cost, nilIfEmpty(task.Notes),
	).Scan(&task.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apperrors.ErrNotFound
		}
		return fmt.Errorf("task_repo: update: %w", err)
	}
	return nil
}

func (r *TaskRepo) Delete(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error {
	query := `DELETE FROM maintenance_tasks WHERE id = $1 AND tenant_id = $2`
	tag, err := r.pool.Exec(ctx, query, id, tenantID)
	if err != nil {
		return fmt.Errorf("task_repo: delete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

package repositories

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/database"
	"github.com/jeckersberger/rentflow/services/maintenance-service/internal/domain"
)

type MaintenanceTaskPostgres struct {
	db *database.PostgresPool
}

func NewMaintenanceTaskPostgres(db *database.PostgresPool) *MaintenanceTaskPostgres {
	return &MaintenanceTaskPostgres{db: db}
}

func (r *MaintenanceTaskPostgres) Create(ctx context.Context, task *domain.MaintenanceTask) error {
	query := `
		INSERT INTO maintenance_tasks (id, tenant_id, plan_id, equipment_id, assigned_to, status, priority, scheduled_at, started_at, completed_at, notes, checklist_data, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`
	var checklistData interface{}
	if task.ChecklistData != nil {
		data, _ := json.Marshal(task.ChecklistData)
		checklistData = string(data)
	}

	_, err := r.db.Exec(ctx, query,
		task.ID, task.TenantID, task.PlanID, task.EquipmentID, task.AssignedTo, task.Status, task.Priority,
		task.ScheduledAt, task.StartedAt, task.CompletedAt, task.Notes, checklistData, task.CreatedAt, task.UpdatedAt,
	)
	return err
}

func (r *MaintenanceTaskPostgres) GetByID(ctx context.Context, tenantID, id string) (*domain.MaintenanceTask, error) {
	query := `
		SELECT id, tenant_id, plan_id, equipment_id, assigned_to, status, priority, scheduled_at, started_at, completed_at, notes, checklist_data, created_at, updated_at
		FROM maintenance_tasks WHERE id = $1 AND tenant_id = $2
	`
	row := r.db.QueryRow(ctx, query, id, tenantID)

	task := &domain.MaintenanceTask{}
	var checklistDataStr *string
	err := row.Scan(&task.ID, &task.TenantID, &task.PlanID, &task.EquipmentID, &task.AssignedTo, &task.Status, &task.Priority,
		&task.ScheduledAt, &task.StartedAt, &task.CompletedAt, &task.Notes, &checklistDataStr, &task.CreatedAt, &task.UpdatedAt)
	if err != nil {
		return nil, err
	}

	if checklistDataStr != nil {
		var cd domain.ChecklistData
		if err := json.Unmarshal([]byte(*checklistDataStr), &cd); err == nil {
			task.ChecklistData = &cd
		}
	}

	return task, nil
}

func (r *MaintenanceTaskPostgres) ListByTenant(ctx context.Context, tenantID string) ([]*domain.MaintenanceTask, error) {
	query := `
		SELECT id, tenant_id, plan_id, equipment_id, assigned_to, status, priority, scheduled_at, started_at, completed_at, notes, checklist_data, created_at, updated_at
		FROM maintenance_tasks WHERE tenant_id = $1 ORDER BY scheduled_at DESC
	`
	rows, err := r.db.Query(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*domain.MaintenanceTask
	for rows.Next() {
		task := &domain.MaintenanceTask{}
		var checklistDataStr *string
		err := rows.Scan(&task.ID, &task.TenantID, &task.PlanID, &task.EquipmentID, &task.AssignedTo, &task.Status, &task.Priority,
			&task.ScheduledAt, &task.StartedAt, &task.CompletedAt, &task.Notes, &checklistDataStr, &task.CreatedAt, &task.UpdatedAt)
		if err != nil {
			return nil, err
		}
		if checklistDataStr != nil {
			var cd domain.ChecklistData
			if err := json.Unmarshal([]byte(*checklistDataStr), &cd); err == nil {
				task.ChecklistData = &cd
			}
		}
		tasks = append(tasks, task)
	}

	return tasks, rows.Err()
}

func (r *MaintenanceTaskPostgres) ListByPlan(ctx context.Context, tenantID, planID string) ([]*domain.MaintenanceTask, error) {
	query := `
		SELECT id, tenant_id, plan_id, equipment_id, assigned_to, status, priority, scheduled_at, started_at, completed_at, notes, checklist_data, created_at, updated_at
		FROM maintenance_tasks WHERE tenant_id = $1 AND plan_id = $2 ORDER BY scheduled_at DESC
	`
	rows, err := r.db.Query(ctx, query, tenantID, planID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*domain.MaintenanceTask
	for rows.Next() {
		task := &domain.MaintenanceTask{}
		var checklistDataStr *string
		err := rows.Scan(&task.ID, &task.TenantID, &task.PlanID, &task.EquipmentID, &task.AssignedTo, &task.Status, &task.Priority,
			&task.ScheduledAt, &task.StartedAt, &task.CompletedAt, &task.Notes, &checklistDataStr, &task.CreatedAt, &task.UpdatedAt)
		if err != nil {
			return nil, err
		}
		if checklistDataStr != nil {
			var cd domain.ChecklistData
			if err := json.Unmarshal([]byte(*checklistDataStr), &cd); err == nil {
				task.ChecklistData = &cd
			}
		}
		tasks = append(tasks, task)
	}

	return tasks, rows.Err()
}

func (r *MaintenanceTaskPostgres) ListByStatus(ctx context.Context, tenantID string, status domain.TaskStatus) ([]*domain.MaintenanceTask, error) {
	query := `
		SELECT id, tenant_id, plan_id, equipment_id, assigned_to, status, priority, scheduled_at, started_at, completed_at, notes, checklist_data, created_at, updated_at
		FROM maintenance_tasks WHERE tenant_id = $1 AND status = $2 ORDER BY scheduled_at DESC
	`
	rows, err := r.db.Query(ctx, query, tenantID, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*domain.MaintenanceTask
	for rows.Next() {
		task := &domain.MaintenanceTask{}
		var checklistDataStr *string
		err := rows.Scan(&task.ID, &task.TenantID, &task.PlanID, &task.EquipmentID, &task.AssignedTo, &task.Status, &task.Priority,
			&task.ScheduledAt, &task.StartedAt, &task.CompletedAt, &task.Notes, &checklistDataStr, &task.CreatedAt, &task.UpdatedAt)
		if err != nil {
			return nil, err
		}
		if checklistDataStr != nil {
			var cd domain.ChecklistData
			if err := json.Unmarshal([]byte(*checklistDataStr), &cd); err == nil {
				task.ChecklistData = &cd
			}
		}
		tasks = append(tasks, task)
	}

	return tasks, rows.Err()
}

func (r *MaintenanceTaskPostgres) Update(ctx context.Context, task *domain.MaintenanceTask) error {
	query := `
		UPDATE maintenance_tasks
		SET status = $1, assigned_to = $2, started_at = $3, completed_at = $4, notes = $5, checklist_data = $6, updated_at = $7
		WHERE id = $8 AND tenant_id = $9
	`
	var checklistData interface{}
	if task.ChecklistData != nil {
		data, _ := json.Marshal(task.ChecklistData)
		checklistData = string(data)
	}

	_, err := r.db.Exec(ctx, query,
		task.Status, task.AssignedTo, task.StartedAt, task.CompletedAt, task.Notes, checklistData, task.UpdatedAt, task.ID, task.TenantID,
	)
	return err
}

func (r *MaintenanceTaskPostgres) Delete(ctx context.Context, tenantID, id string) error {
	query := `DELETE FROM maintenance_tasks WHERE id = $1 AND tenant_id = $2`
	_, err := r.db.Exec(ctx, query, id, tenantID)
	return err
}

func (r *MaintenanceTaskPostgres) ListDueTasks(ctx context.Context, tenantID string) ([]*domain.MaintenanceTask, error) {
	query := `
		SELECT id, tenant_id, plan_id, equipment_id, assigned_to, status, priority, scheduled_at, started_at, completed_at, notes, checklist_data, created_at, updated_at
		FROM maintenance_tasks WHERE tenant_id = $1 AND status IN ('planned', 'in_progress') AND scheduled_at <= $2 AND scheduled_at > $3 ORDER BY scheduled_at ASC
	`
	now := time.Now()
	rows, err := r.db.Query(ctx, query, tenantID, now, now.AddDate(0, 0, -30))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*domain.MaintenanceTask
	for rows.Next() {
		task := &domain.MaintenanceTask{}
		var checklistDataStr *string
		err := rows.Scan(&task.ID, &task.TenantID, &task.PlanID, &task.EquipmentID, &task.AssignedTo, &task.Status, &task.Priority,
			&task.ScheduledAt, &task.StartedAt, &task.CompletedAt, &task.Notes, &checklistDataStr, &task.CreatedAt, &task.UpdatedAt)
		if err != nil {
			return nil, err
		}
		if checklistDataStr != nil {
			var cd domain.ChecklistData
			if err := json.Unmarshal([]byte(*checklistDataStr), &cd); err == nil {
				task.ChecklistData = &cd
			}
		}
		tasks = append(tasks, task)
	}

	return tasks, rows.Err()
}

func (r *MaintenanceTaskPostgres) ListOverdueTasks(ctx context.Context, tenantID string) ([]*domain.MaintenanceTask, error) {
	query := `
		SELECT id, tenant_id, plan_id, equipment_id, assigned_to, status, priority, scheduled_at, started_at, completed_at, notes, checklist_data, created_at, updated_at
		FROM maintenance_tasks WHERE tenant_id = $1 AND status IN ('planned', 'in_progress') AND scheduled_at < $2 ORDER BY scheduled_at ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID, time.Now())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*domain.MaintenanceTask
	for rows.Next() {
		task := &domain.MaintenanceTask{}
		var checklistDataStr *string
		err := rows.Scan(&task.ID, &task.TenantID, &task.PlanID, &task.EquipmentID, &task.AssignedTo, &task.Status, &task.Priority,
			&task.ScheduledAt, &task.StartedAt, &task.CompletedAt, &task.Notes, &checklistDataStr, &task.CreatedAt, &task.UpdatedAt)
		if err != nil {
			return nil, err
		}
		if checklistDataStr != nil {
			var cd domain.ChecklistData
			if err := json.Unmarshal([]byte(*checklistDataStr), &cd); err == nil {
				task.ChecklistData = &cd
			}
		}
		tasks = append(tasks, task)
	}

	return tasks, rows.Err()
}

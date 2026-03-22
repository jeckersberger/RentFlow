package repositories

import (
	"context"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/database"
	"github.com/jeckersberger/rentflow/services/maintenance-service/internal/domain"
)

type MaintenancePlanPostgres struct {
	db *database.PostgresPool
}

func NewMaintenancePlanPostgres(db *database.PostgresPool) *MaintenancePlanPostgres {
	return &MaintenancePlanPostgres{db: db}
}

func (r *MaintenancePlanPostgres) Create(ctx context.Context, plan *domain.MaintenancePlan) error {
	query := `
		INSERT INTO maintenance_plans (id, tenant_id, equipment_id, plan_type, interval_days, interval_hours, name, description, checklist_template_id, last_executed_at, next_due_at, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`
	_, err := r.db.Exec(ctx, query,
		plan.ID, plan.TenantID, plan.EquipmentID, plan.PlanType, plan.IntervalDays, plan.IntervalHours,
		plan.Name, plan.Description, plan.ChecklistTemplateID, plan.LastExecutedAt, plan.NextDueAt,
		plan.IsActive, plan.CreatedAt, plan.UpdatedAt,
	)
	return err
}

func (r *MaintenancePlanPostgres) GetByID(ctx context.Context, tenantID, id string) (*domain.MaintenancePlan, error) {
	query := `
		SELECT id, tenant_id, equipment_id, plan_type, interval_days, interval_hours, name, description, checklist_template_id, last_executed_at, next_due_at, is_active, created_at, updated_at
		FROM maintenance_plans WHERE id = $1 AND tenant_id = $2
	`
	row := r.db.QueryRow(ctx, query, id, tenantID)

	plan := &domain.MaintenancePlan{}
	err := row.Scan(&plan.ID, &plan.TenantID, &plan.EquipmentID, &plan.PlanType, &plan.IntervalDays, &plan.IntervalHours,
		&plan.Name, &plan.Description, &plan.ChecklistTemplateID, &plan.LastExecutedAt, &plan.NextDueAt, &plan.IsActive, &plan.CreatedAt, &plan.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return plan, nil
}

func (r *MaintenancePlanPostgres) GetByEquipment(ctx context.Context, tenantID, equipmentID string) ([]*domain.MaintenancePlan, error) {
	query := `
		SELECT id, tenant_id, equipment_id, plan_type, interval_days, interval_hours, name, description, checklist_template_id, last_executed_at, next_due_at, is_active, created_at, updated_at
		FROM maintenance_plans WHERE tenant_id = $1 AND equipment_id = $2 AND is_active = true ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, tenantID, equipmentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plans []*domain.MaintenancePlan
	for rows.Next() {
		plan := &domain.MaintenancePlan{}
		err := rows.Scan(&plan.ID, &plan.TenantID, &plan.EquipmentID, &plan.PlanType, &plan.IntervalDays, &plan.IntervalHours,
			&plan.Name, &plan.Description, &plan.ChecklistTemplateID, &plan.LastExecutedAt, &plan.NextDueAt, &plan.IsActive, &plan.CreatedAt, &plan.UpdatedAt)
		if err != nil {
			return nil, err
		}
		plans = append(plans, plan)
	}

	return plans, rows.Err()
}

func (r *MaintenancePlanPostgres) ListByTenant(ctx context.Context, tenantID string) ([]*domain.MaintenancePlan, error) {
	query := `
		SELECT id, tenant_id, equipment_id, plan_type, interval_days, interval_hours, name, description, checklist_template_id, last_executed_at, next_due_at, is_active, created_at, updated_at
		FROM maintenance_plans WHERE tenant_id = $1 AND is_active = true ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plans []*domain.MaintenancePlan
	for rows.Next() {
		plan := &domain.MaintenancePlan{}
		err := rows.Scan(&plan.ID, &plan.TenantID, &plan.EquipmentID, &plan.PlanType, &plan.IntervalDays, &plan.IntervalHours,
			&plan.Name, &plan.Description, &plan.ChecklistTemplateID, &plan.LastExecutedAt, &plan.NextDueAt, &plan.IsActive, &plan.CreatedAt, &plan.UpdatedAt)
		if err != nil {
			return nil, err
		}
		plans = append(plans, plan)
	}

	return plans, rows.Err()
}

func (r *MaintenancePlanPostgres) Update(ctx context.Context, plan *domain.MaintenancePlan) error {
	query := `
		UPDATE maintenance_plans
		SET name = $1, description = $2, interval_days = $3, interval_hours = $4, checklist_template_id = $5, last_executed_at = $6, next_due_at = $7, is_active = $8, updated_at = $9
		WHERE id = $10 AND tenant_id = $11
	`
	_, err := r.db.Exec(ctx, query,
		plan.Name, plan.Description, plan.IntervalDays, plan.IntervalHours, plan.ChecklistTemplateID,
		plan.LastExecutedAt, plan.NextDueAt, plan.IsActive, plan.UpdatedAt, plan.ID, plan.TenantID,
	)
	return err
}

func (r *MaintenancePlanPostgres) Delete(ctx context.Context, tenantID, id string) error {
	query := `UPDATE maintenance_plans SET is_active = false WHERE id = $1 AND tenant_id = $2`
	_, err := r.db.Exec(ctx, query, id, tenantID)
	return err
}

func (r *MaintenancePlanPostgres) ListDuePlans(ctx context.Context, tenantID string) ([]*domain.MaintenancePlan, error) {
	query := `
		SELECT id, tenant_id, equipment_id, plan_type, interval_days, interval_hours, name, description, checklist_template_id, last_executed_at, next_due_at, is_active, created_at, updated_at
		FROM maintenance_plans WHERE tenant_id = $1 AND is_active = true AND next_due_at IS NOT NULL AND next_due_at <= $2 ORDER BY next_due_at ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID, time.Now())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plans []*domain.MaintenancePlan
	for rows.Next() {
		plan := &domain.MaintenancePlan{}
		err := rows.Scan(&plan.ID, &plan.TenantID, &plan.EquipmentID, &plan.PlanType, &plan.IntervalDays, &plan.IntervalHours,
			&plan.Name, &plan.Description, &plan.ChecklistTemplateID, &plan.LastExecutedAt, &plan.NextDueAt, &plan.IsActive, &plan.CreatedAt, &plan.UpdatedAt)
		if err != nil {
			return nil, err
		}
		plans = append(plans, plan)
	}

	return plans, rows.Err()
}

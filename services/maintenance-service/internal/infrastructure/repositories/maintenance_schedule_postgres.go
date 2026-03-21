package repositories

import (
	"context"

	"github.com/jeckersberger/rentflow/pkg/common/database"
	"github.com/jeckersberger/rentflow/services/maintenance-service/internal/domain"
)

type MaintenanceSchedulePostgres struct {
	db *database.PostgresPool
}

func NewMaintenanceSchedulePostgres(db *database.PostgresPool) *MaintenanceSchedulePostgres {
	return &MaintenanceSchedulePostgres{db: db}
}

func (r *MaintenanceSchedulePostgres) Create(ctx context.Context, schedule *domain.MaintenanceSchedule) error {
	query := `
		INSERT INTO maintenance_schedules (id, tenant_id, equipment_id, interval_days, last_performed, next_due, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.Exec(ctx, query,
		schedule.ID, schedule.TenantID, schedule.EquipmentID, schedule.IntervalDays, schedule.LastPerformed, schedule.NextDue, schedule.CreatedAt, schedule.UpdatedAt,
	)
	return err
}

func (r *MaintenanceSchedulePostgres) GetByID(ctx context.Context, tenantID, id string) (*domain.MaintenanceSchedule, error) {
	query := `
		SELECT id, tenant_id, equipment_id, interval_days, last_performed, next_due, created_at, updated_at
		FROM maintenance_schedules WHERE id = $1 AND tenant_id = $2
	`
	row := r.db.QueryRow(ctx, query, id, tenantID)

	schedule := &domain.MaintenanceSchedule{}
	err := row.Scan(&schedule.ID, &schedule.TenantID, &schedule.EquipmentID, &schedule.IntervalDays,
		&schedule.LastPerformed, &schedule.NextDue, &schedule.CreatedAt, &schedule.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return schedule, nil
}

func (r *MaintenanceSchedulePostgres) GetByEquipment(ctx context.Context, tenantID, equipmentID string) (*domain.MaintenanceSchedule, error) {
	query := `
		SELECT id, tenant_id, equipment_id, interval_days, last_performed, next_due, created_at, updated_at
		FROM maintenance_schedules WHERE tenant_id = $1 AND equipment_id = $2
	`
	row := r.db.QueryRow(ctx, query, tenantID, equipmentID)

	schedule := &domain.MaintenanceSchedule{}
	err := row.Scan(&schedule.ID, &schedule.TenantID, &schedule.EquipmentID, &schedule.IntervalDays,
		&schedule.LastPerformed, &schedule.NextDue, &schedule.CreatedAt, &schedule.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return schedule, nil
}

func (r *MaintenanceSchedulePostgres) ListByTenant(ctx context.Context, tenantID string) ([]*domain.MaintenanceSchedule, error) {
	query := `
		SELECT id, tenant_id, equipment_id, interval_days, last_performed, next_due, created_at, updated_at
		FROM maintenance_schedules WHERE tenant_id = $1 ORDER BY next_due ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schedules []*domain.MaintenanceSchedule
	for rows.Next() {
		schedule := &domain.MaintenanceSchedule{}
		err := rows.Scan(&schedule.ID, &schedule.TenantID, &schedule.EquipmentID, &schedule.IntervalDays,
			&schedule.LastPerformed, &schedule.NextDue, &schedule.CreatedAt, &schedule.UpdatedAt)
		if err != nil {
			return nil, err
		}
		schedules = append(schedules, schedule)
	}

	return schedules, rows.Err()
}

func (r *MaintenanceSchedulePostgres) Update(ctx context.Context, schedule *domain.MaintenanceSchedule) error {
	query := `
		UPDATE maintenance_schedules
		SET interval_days = $1, last_performed = $2, next_due = $3, updated_at = $4
		WHERE id = $5 AND tenant_id = $6
	`
	_, err := r.db.Exec(ctx, query,
		schedule.IntervalDays, schedule.LastPerformed, schedule.NextDue, schedule.UpdatedAt, schedule.ID, schedule.TenantID,
	)
	return err
}

func (r *MaintenanceSchedulePostgres) Delete(ctx context.Context, tenantID, id string) error {
	query := `DELETE FROM maintenance_schedules WHERE id = $1 AND tenant_id = $2`
	_, err := r.db.Exec(ctx, query, id, tenantID)
	return err
}

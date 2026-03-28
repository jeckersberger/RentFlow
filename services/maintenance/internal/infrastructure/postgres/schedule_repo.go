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
	"github.com/jeckersberger/EquipFlow/services/maintenance/internal/domain"
)

const scheduleColumns = `
	id, tenant_id, equipment_id, name, interval_days,
	last_performed_at, next_due_at, is_active, notes, created_at`

type ScheduleRepo struct {
	pool *pgxpool.Pool
}

func NewScheduleRepo(pool *pgxpool.Pool) *ScheduleRepo {
	return &ScheduleRepo{pool: pool}
}

func scanSchedule(row pgx.Row) (*domain.MaintenanceSchedule, error) {
	s := &domain.MaintenanceSchedule{}
	var notes *string

	err := row.Scan(
		&s.ID, &s.TenantID, &s.EquipmentID, &s.Name, &s.IntervalDays,
		&s.LastPerformedAt, &s.NextDueAt, &s.IsActive, &notes, &s.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	if notes != nil {
		s.Notes = *notes
	}

	return s, nil
}

func (r *ScheduleRepo) Create(ctx context.Context, schedule *domain.MaintenanceSchedule) error {
	query := `
		INSERT INTO maintenance_schedules (
			id, tenant_id, equipment_id, name, interval_days,
			next_due_at, is_active, notes
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING created_at`

	err := r.pool.QueryRow(ctx, query,
		schedule.ID, schedule.TenantID, schedule.EquipmentID,
		schedule.Name, schedule.IntervalDays,
		schedule.NextDueAt, schedule.IsActive, nilIfEmpty(schedule.Notes),
	).Scan(&schedule.CreatedAt)
	if err != nil {
		return fmt.Errorf("schedule_repo: create: %w", err)
	}
	return nil
}

func (r *ScheduleRepo) GetByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*domain.MaintenanceSchedule, error) {
	query := fmt.Sprintf(`SELECT %s FROM maintenance_schedules WHERE id = $1 AND tenant_id = $2`, scheduleColumns)
	schedule, err := scanSchedule(r.pool.QueryRow(ctx, query, id, tenantID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("schedule_repo: get_by_id: %w", err)
	}
	return schedule, nil
}

func (r *ScheduleRepo) List(ctx context.Context, tenantID uuid.UUID, filter domain.ScheduleFilter) ([]*domain.MaintenanceSchedule, int64, error) {
	conditions := []string{"tenant_id = $1"}
	args := []interface{}{tenantID}
	argIdx := 2

	if filter.EquipmentID != nil {
		conditions = append(conditions, fmt.Sprintf("equipment_id = $%d", argIdx))
		args = append(args, *filter.EquipmentID)
		argIdx++
	}

	where := strings.Join(conditions, " AND ")

	var total int64
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM maintenance_schedules WHERE %s`, where)
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("schedule_repo: list count: %w", err)
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
		`SELECT %s FROM maintenance_schedules WHERE %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		scheduleColumns, where, argIdx, argIdx+1,
	)
	args = append(args, perPage, offset)

	rows, err := r.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("schedule_repo: list query: %w", err)
	}
	defer rows.Close()

	var items []*domain.MaintenanceSchedule
	for rows.Next() {
		schedule, scanErr := scanSchedule(rows)
		if scanErr != nil {
			return nil, 0, fmt.Errorf("schedule_repo: list scan: %w", scanErr)
		}
		items = append(items, schedule)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("schedule_repo: list rows: %w", err)
	}
	return items, total, nil
}

func (r *ScheduleRepo) Update(ctx context.Context, schedule *domain.MaintenanceSchedule) error {
	query := `
		UPDATE maintenance_schedules SET
			name = $3, interval_days = $4, next_due_at = $5,
			is_active = $6, notes = $7
		WHERE id = $1 AND tenant_id = $2
		RETURNING created_at`

	err := r.pool.QueryRow(ctx, query,
		schedule.ID, schedule.TenantID,
		schedule.Name, schedule.IntervalDays, schedule.NextDueAt,
		schedule.IsActive, nilIfEmpty(schedule.Notes),
	).Scan(&schedule.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apperrors.ErrNotFound
		}
		return fmt.Errorf("schedule_repo: update: %w", err)
	}
	return nil
}

func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

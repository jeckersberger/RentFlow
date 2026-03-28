package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jeckersberger/EquipFlow/services/maintenance/internal/domain"
)

const logColumns = `id, task_id, tenant_id, action, performed_by, notes, created_at`

type LogRepo struct {
	pool *pgxpool.Pool
}

func NewLogRepo(pool *pgxpool.Pool) *LogRepo {
	return &LogRepo{pool: pool}
}

func scanLog(row pgx.Row) (*domain.MaintenanceLog, error) {
	l := &domain.MaintenanceLog{}
	var notes *string

	err := row.Scan(
		&l.ID, &l.TaskID, &l.TenantID, &l.Action,
		&l.PerformedBy, &notes, &l.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	if notes != nil {
		l.Notes = *notes
	}

	return l, nil
}

func (r *LogRepo) Create(ctx context.Context, logEntry *domain.MaintenanceLog) error {
	query := `
		INSERT INTO maintenance_logs (
			id, task_id, tenant_id, action, performed_by, notes
		) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at`

	err := r.pool.QueryRow(ctx, query,
		logEntry.ID, logEntry.TaskID, logEntry.TenantID,
		logEntry.Action, logEntry.PerformedBy, nilIfEmpty(logEntry.Notes),
	).Scan(&logEntry.CreatedAt)
	if err != nil {
		return fmt.Errorf("log_repo: create: %w", err)
	}
	return nil
}

func (r *LogRepo) ListByTask(ctx context.Context, taskID uuid.UUID, tenantID uuid.UUID, filter domain.LogFilter) ([]*domain.MaintenanceLog, int64, error) {
	var total int64
	countQuery := `SELECT COUNT(*) FROM maintenance_logs WHERE task_id = $1 AND tenant_id = $2`
	err := r.pool.QueryRow(ctx, countQuery, taskID, tenantID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("log_repo: list count: %w", err)
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
		`SELECT %s FROM maintenance_logs WHERE task_id = $1 AND tenant_id = $2 ORDER BY created_at DESC LIMIT $3 OFFSET $4`,
		logColumns,
	)

	rows, err := r.pool.Query(ctx, dataQuery, taskID, tenantID, perPage, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("log_repo: list query: %w", err)
	}
	defer rows.Close()

	var items []*domain.MaintenanceLog
	for rows.Next() {
		logEntry, scanErr := scanLog(rows)
		if scanErr != nil {
			return nil, 0, fmt.Errorf("log_repo: list scan: %w", scanErr)
		}
		items = append(items, logEntry)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("log_repo: list rows: %w", err)
	}
	return items, total, nil
}

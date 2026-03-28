package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	apperrors "github.com/jeckersberger/EquipFlow/pkg/common/errors"
	"github.com/jeckersberger/EquipFlow/services/reporting/internal/domain"
)

const snapshotColumns = `
	id, definition_id, tenant_id, title, data, file_path, file_size,
	format, period_start, period_end, generated_by, generated_at`

type SnapshotRepo struct {
	pool *pgxpool.Pool
}

func NewSnapshotRepo(pool *pgxpool.Pool) *SnapshotRepo {
	return &SnapshotRepo{pool: pool}
}

func scanSnapshot(row pgx.Row) (*domain.ReportSnapshot, error) {
	snap := &domain.ReportSnapshot{}
	var (
		data        []byte
		filePath    *string
		periodStart *time.Time
		periodEnd   *time.Time
		generatedBy *uuid.UUID
	)

	err := row.Scan(
		&snap.ID, &snap.DefinitionID, &snap.TenantID, &snap.Title,
		&data, &filePath, &snap.FileSize,
		&snap.Format, &periodStart, &periodEnd, &generatedBy, &snap.GeneratedAt,
	)
	if err != nil {
		return nil, err
	}

	if data != nil {
		snap.Data = json.RawMessage(data)
	} else {
		snap.Data = json.RawMessage("{}")
	}
	if filePath != nil {
		snap.FilePath = *filePath
	}
	if periodStart != nil {
		snap.PeriodStart = periodStart.Format("2006-01-02")
	}
	if periodEnd != nil {
		snap.PeriodEnd = periodEnd.Format("2006-01-02")
	}
	snap.GeneratedBy = generatedBy

	return snap, nil
}

func (r *SnapshotRepo) Create(ctx context.Context, snap *domain.ReportSnapshot) error {
	query := `
		INSERT INTO report_snapshots (
			id, definition_id, tenant_id, title, data, file_path, file_size,
			format, period_start, period_end, generated_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING generated_at`

	err := r.pool.QueryRow(ctx, query,
		snap.ID, snap.DefinitionID, snap.TenantID, snap.Title,
		snap.Data, nilIfEmpty(snap.FilePath), snap.FileSize,
		snap.Format, nilIfEmpty(snap.PeriodStart), nilIfEmpty(snap.PeriodEnd),
		snap.GeneratedBy,
	).Scan(&snap.GeneratedAt)
	if err != nil {
		return fmt.Errorf("snapshot_repo: create: %w", err)
	}
	return nil
}

func (r *SnapshotRepo) GetByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*domain.ReportSnapshot, error) {
	query := fmt.Sprintf(`SELECT %s FROM report_snapshots WHERE id = $1 AND tenant_id = $2`, snapshotColumns)
	snap, err := scanSnapshot(r.pool.QueryRow(ctx, query, id, tenantID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("snapshot_repo: get_by_id: %w", err)
	}
	return snap, nil
}

func (r *SnapshotRepo) List(ctx context.Context, tenantID uuid.UUID, filter domain.SnapshotFilter) ([]*domain.ReportSnapshot, int64, error) {
	conditions := []string{"tenant_id = $1"}
	args := []interface{}{tenantID}
	argIdx := 2

	if filter.DefinitionID != nil {
		conditions = append(conditions, fmt.Sprintf("definition_id = $%d", argIdx))
		args = append(args, *filter.DefinitionID)
		argIdx++
	}

	where := strings.Join(conditions, " AND ")

	var total int64
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM report_snapshots WHERE %s`, where)
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("snapshot_repo: list count: %w", err)
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
		`SELECT %s FROM report_snapshots WHERE %s ORDER BY generated_at DESC LIMIT $%d OFFSET $%d`,
		snapshotColumns, where, argIdx, argIdx+1,
	)
	args = append(args, perPage, offset)

	rows, err := r.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("snapshot_repo: list query: %w", err)
	}
	defer rows.Close()

	var items []*domain.ReportSnapshot
	for rows.Next() {
		snap, scanErr := scanSnapshot(rows)
		if scanErr != nil {
			return nil, 0, fmt.Errorf("snapshot_repo: list scan: %w", scanErr)
		}
		items = append(items, snap)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("snapshot_repo: list rows: %w", err)
	}
	return items, total, nil
}

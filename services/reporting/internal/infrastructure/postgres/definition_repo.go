package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	apperrors "github.com/jeckersberger/EquipFlow/pkg/common/errors"
	"github.com/jeckersberger/EquipFlow/services/reporting/internal/domain"
)

const definitionColumns = `
	id, tenant_id, name, type, query_config, schedule, format,
	is_active, created_by, created_at, updated_at`

type DefinitionRepo struct {
	pool *pgxpool.Pool
}

func NewDefinitionRepo(pool *pgxpool.Pool) *DefinitionRepo {
	return &DefinitionRepo{pool: pool}
}

func scanDefinition(row pgx.Row) (*domain.ReportDefinition, error) {
	def := &domain.ReportDefinition{}
	var (
		queryConfig []byte
		schedule    *string
		createdBy   *uuid.UUID
	)

	err := row.Scan(
		&def.ID, &def.TenantID, &def.Name, &def.Type, &queryConfig, &schedule, &def.Format,
		&def.IsActive, &createdBy, &def.CreatedAt, &def.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if queryConfig != nil {
		def.QueryConfig = json.RawMessage(queryConfig)
	} else {
		def.QueryConfig = json.RawMessage("{}")
	}
	if schedule != nil {
		def.Schedule = *schedule
	}
	def.CreatedBy = createdBy

	return def, nil
}

func (r *DefinitionRepo) Create(ctx context.Context, def *domain.ReportDefinition) error {
	query := `
		INSERT INTO report_definitions (
			id, tenant_id, name, type, query_config, schedule, format,
			is_active, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING created_at, updated_at`

	err := r.pool.QueryRow(ctx, query,
		def.ID, def.TenantID, def.Name, def.Type, def.QueryConfig,
		nilIfEmpty(def.Schedule), def.Format, def.IsActive, def.CreatedBy,
	).Scan(&def.CreatedAt, &def.UpdatedAt)
	if err != nil {
		return fmt.Errorf("definition_repo: create: %w", err)
	}
	return nil
}

func (r *DefinitionRepo) GetByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*domain.ReportDefinition, error) {
	query := fmt.Sprintf(`SELECT %s FROM report_definitions WHERE id = $1 AND tenant_id = $2`, definitionColumns)
	def, err := scanDefinition(r.pool.QueryRow(ctx, query, id, tenantID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("definition_repo: get_by_id: %w", err)
	}
	return def, nil
}

func (r *DefinitionRepo) List(ctx context.Context, tenantID uuid.UUID, filter domain.DefinitionFilter) ([]*domain.ReportDefinition, int64, error) {
	conditions := []string{"tenant_id = $1"}
	args := []interface{}{tenantID}
	argIdx := 2

	if filter.Type != "" {
		conditions = append(conditions, fmt.Sprintf("type = $%d", argIdx))
		args = append(args, filter.Type)
		argIdx++
	}

	where := strings.Join(conditions, " AND ")

	var total int64
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM report_definitions WHERE %s`, where)
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("definition_repo: list count: %w", err)
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
		`SELECT %s FROM report_definitions WHERE %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		definitionColumns, where, argIdx, argIdx+1,
	)
	args = append(args, perPage, offset)

	rows, err := r.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("definition_repo: list query: %w", err)
	}
	defer rows.Close()

	var items []*domain.ReportDefinition
	for rows.Next() {
		def, scanErr := scanDefinition(rows)
		if scanErr != nil {
			return nil, 0, fmt.Errorf("definition_repo: list scan: %w", scanErr)
		}
		items = append(items, def)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("definition_repo: list rows: %w", err)
	}
	return items, total, nil
}

func (r *DefinitionRepo) Update(ctx context.Context, def *domain.ReportDefinition) error {
	query := `
		UPDATE report_definitions SET
			name = $3, type = $4, query_config = $5, schedule = $6,
			format = $7, is_active = $8, updated_at = NOW()
		WHERE id = $1 AND tenant_id = $2
		RETURNING updated_at`

	err := r.pool.QueryRow(ctx, query,
		def.ID, def.TenantID,
		def.Name, def.Type, def.QueryConfig, nilIfEmpty(def.Schedule),
		def.Format, def.IsActive,
	).Scan(&def.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apperrors.ErrNotFound
		}
		return fmt.Errorf("definition_repo: update: %w", err)
	}
	return nil
}

func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

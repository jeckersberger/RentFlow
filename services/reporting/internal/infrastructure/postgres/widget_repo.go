package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	apperrors "github.com/jeckersberger/EquipFlow/pkg/common/errors"
	"github.com/jeckersberger/EquipFlow/services/reporting/internal/domain"
)

const widgetColumns = `
	id, tenant_id, name, type, config, position,
	is_active, created_by, created_at, updated_at`

type WidgetRepo struct {
	pool *pgxpool.Pool
}

func NewWidgetRepo(pool *pgxpool.Pool) *WidgetRepo {
	return &WidgetRepo{pool: pool}
}

func scanWidget(row pgx.Row) (*domain.DashboardWidget, error) {
	w := &domain.DashboardWidget{}
	var (
		config    []byte
		createdBy *uuid.UUID
	)

	err := row.Scan(
		&w.ID, &w.TenantID, &w.Name, &w.Type, &config, &w.Position,
		&w.IsActive, &createdBy, &w.CreatedAt, &w.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if config != nil {
		w.Config = json.RawMessage(config)
	} else {
		w.Config = json.RawMessage("{}")
	}
	w.CreatedBy = createdBy

	return w, nil
}

func (r *WidgetRepo) Create(ctx context.Context, widget *domain.DashboardWidget) error {
	query := `
		INSERT INTO dashboard_widgets (
			id, tenant_id, name, type, config, position, is_active, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING created_at, updated_at`

	err := r.pool.QueryRow(ctx, query,
		widget.ID, widget.TenantID, widget.Name, widget.Type,
		widget.Config, widget.Position, widget.IsActive, widget.CreatedBy,
	).Scan(&widget.CreatedAt, &widget.UpdatedAt)
	if err != nil {
		return fmt.Errorf("widget_repo: create: %w", err)
	}
	return nil
}

func (r *WidgetRepo) GetByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*domain.DashboardWidget, error) {
	query := fmt.Sprintf(`SELECT %s FROM dashboard_widgets WHERE id = $1 AND tenant_id = $2`, widgetColumns)
	w, err := scanWidget(r.pool.QueryRow(ctx, query, id, tenantID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("widget_repo: get_by_id: %w", err)
	}
	return w, nil
}

func (r *WidgetRepo) List(ctx context.Context, tenantID uuid.UUID, filter domain.WidgetFilter) ([]*domain.DashboardWidget, int64, error) {
	var total int64
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM dashboard_widgets WHERE tenant_id = $1`, tenantID,
	).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("widget_repo: list count: %w", err)
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
		`SELECT %s FROM dashboard_widgets WHERE tenant_id = $1 ORDER BY position ASC, created_at DESC LIMIT $2 OFFSET $3`,
		widgetColumns,
	)

	rows, err := r.pool.Query(ctx, dataQuery, tenantID, perPage, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("widget_repo: list query: %w", err)
	}
	defer rows.Close()

	var items []*domain.DashboardWidget
	for rows.Next() {
		w, scanErr := scanWidget(rows)
		if scanErr != nil {
			return nil, 0, fmt.Errorf("widget_repo: list scan: %w", scanErr)
		}
		items = append(items, w)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("widget_repo: list rows: %w", err)
	}
	return items, total, nil
}

func (r *WidgetRepo) Update(ctx context.Context, widget *domain.DashboardWidget) error {
	query := `
		UPDATE dashboard_widgets SET
			name = $3, type = $4, config = $5, position = $6,
			is_active = $7, updated_at = NOW()
		WHERE id = $1 AND tenant_id = $2
		RETURNING updated_at`

	err := r.pool.QueryRow(ctx, query,
		widget.ID, widget.TenantID,
		widget.Name, widget.Type, widget.Config, widget.Position, widget.IsActive,
	).Scan(&widget.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apperrors.ErrNotFound
		}
		return fmt.Errorf("widget_repo: update: %w", err)
	}
	return nil
}

func (r *WidgetRepo) Delete(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error {
	tag, err := r.pool.Exec(ctx,
		`DELETE FROM dashboard_widgets WHERE id = $1 AND tenant_id = $2`,
		id, tenantID,
	)
	if err != nil {
		return fmt.Errorf("widget_repo: delete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	apperrors "github.com/jeckersberger/EquipFlow/pkg/common/errors"
	"github.com/jeckersberger/EquipFlow/services/expense/internal/domain"
)

const recurringColumns = `
	id, tenant_id, name, category_id, amount,
	frequency, next_date, is_active, created_at`

// RecurringRepo implements domain.RecurringExpenseRepository using PostgreSQL.
type RecurringRepo struct {
	pool *pgxpool.Pool
}

// NewRecurringRepo creates a new RecurringRepo.
func NewRecurringRepo(pool *pgxpool.Pool) *RecurringRepo {
	return &RecurringRepo{pool: pool}
}

func scanRecurring(row pgx.Row) (*domain.RecurringExpense, error) {
	rec := &domain.RecurringExpense{}
	var (
		categoryID *uuid.UUID
		nextDate   *time.Time
	)

	err := row.Scan(
		&rec.ID, &rec.TenantID, &rec.Name, &categoryID, &rec.Amount,
		&rec.Frequency, &nextDate, &rec.IsActive, &rec.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	if categoryID != nil {
		rec.CategoryID = *categoryID
	}
	if nextDate != nil {
		rec.NextDate = nextDate.Format("2006-01-02")
	}

	return rec, nil
}

func (r *RecurringRepo) Create(ctx context.Context, rec *domain.RecurringExpense) error {
	query := `
		INSERT INTO recurring_expenses (
			id, tenant_id, name, category_id, amount,
			frequency, next_date, is_active
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING created_at`

	err := r.pool.QueryRow(ctx, query,
		rec.ID, rec.TenantID, rec.Name, nilUUID(rec.CategoryID), rec.Amount,
		rec.Frequency, nilIfEmpty(rec.NextDate), rec.IsActive,
	).Scan(&rec.CreatedAt)
	if err != nil {
		return fmt.Errorf("recurring_repo: create: %w", err)
	}
	return nil
}

func (r *RecurringRepo) GetByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*domain.RecurringExpense, error) {
	query := fmt.Sprintf(`SELECT %s FROM recurring_expenses WHERE id = $1 AND tenant_id = $2`, recurringColumns)
	rec, err := scanRecurring(r.pool.QueryRow(ctx, query, id, tenantID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("recurring_repo: get_by_id: %w", err)
	}
	return rec, nil
}

func (r *RecurringRepo) List(ctx context.Context, tenantID uuid.UUID, filter domain.RecurringExpenseFilter) ([]*domain.RecurringExpense, int64, error) {
	var total int64
	countQuery := `SELECT COUNT(*) FROM recurring_expenses WHERE tenant_id = $1`
	err := r.pool.QueryRow(ctx, countQuery, tenantID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("recurring_repo: list count: %w", err)
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
		`SELECT %s FROM recurring_expenses WHERE tenant_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		recurringColumns,
	)

	rows, err := r.pool.Query(ctx, dataQuery, tenantID, perPage, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("recurring_repo: list query: %w", err)
	}
	defer rows.Close()

	var items []*domain.RecurringExpense
	for rows.Next() {
		rec, scanErr := scanRecurring(rows)
		if scanErr != nil {
			return nil, 0, fmt.Errorf("recurring_repo: list scan: %w", scanErr)
		}
		items = append(items, rec)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("recurring_repo: list rows: %w", err)
	}
	return items, total, nil
}

func (r *RecurringRepo) Update(ctx context.Context, rec *domain.RecurringExpense) error {
	query := `
		UPDATE recurring_expenses SET
			name = $3, category_id = $4, amount = $5,
			frequency = $6, next_date = $7, is_active = $8
		WHERE id = $1 AND tenant_id = $2
		RETURNING created_at`

	err := r.pool.QueryRow(ctx, query,
		rec.ID, rec.TenantID,
		rec.Name, nilUUID(rec.CategoryID), rec.Amount,
		rec.Frequency, nilIfEmpty(rec.NextDate), rec.IsActive,
	).Scan(&rec.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apperrors.ErrNotFound
		}
		return fmt.Errorf("recurring_repo: update: %w", err)
	}
	return nil
}

func (r *RecurringRepo) Delete(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error {
	query := `DELETE FROM recurring_expenses WHERE id = $1 AND tenant_id = $2`
	ct, err := r.pool.Exec(ctx, query, id, tenantID)
	if err != nil {
		return fmt.Errorf("recurring_repo: delete: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

// Ensure interface compliance at compile time.
var _ domain.RecurringExpenseRepository = (*RecurringRepo)(nil)

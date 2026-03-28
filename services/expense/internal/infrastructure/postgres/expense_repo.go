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
	"github.com/jeckersberger/EquipFlow/services/expense/internal/domain"
)

const expenseColumns = `
	id, tenant_id, project_id, category_id, description,
	amount, currency, expense_date, receipt_number, vendor,
	status, approved_by, approved_at, notes, created_by,
	created_at, updated_at`

type ExpenseRepo struct {
	pool *pgxpool.Pool
}

func NewExpenseRepo(pool *pgxpool.Pool) *ExpenseRepo {
	return &ExpenseRepo{pool: pool}
}

func scanExpense(row pgx.Row) (*domain.Expense, error) {
	exp := &domain.Expense{}
	var (
		projectID     *uuid.UUID
		categoryID    *uuid.UUID
		receiptNumber *string
		vendor        *string
		approvedBy    *uuid.UUID
		notes         *string
		createdBy     *uuid.UUID
		expenseDate   time.Time
	)

	err := row.Scan(
		&exp.ID, &exp.TenantID, &projectID, &categoryID, &exp.Description,
		&exp.Amount, &exp.Currency, &expenseDate, &receiptNumber, &vendor,
		&exp.Status, &approvedBy, &exp.ApprovedAt, &notes, &createdBy,
		&exp.CreatedAt, &exp.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	exp.ExpenseDate = expenseDate.Format("2006-01-02")
	if projectID != nil {
		exp.ProjectID = *projectID
	}
	if categoryID != nil {
		exp.CategoryID = *categoryID
	}
	exp.ReceiptNumber = derefString(receiptNumber)
	exp.Vendor = derefString(vendor)
	if approvedBy != nil {
		exp.ApprovedBy = *approvedBy
	}
	exp.Notes = derefString(notes)
	if createdBy != nil {
		exp.CreatedBy = *createdBy
	}

	return exp, nil
}

func (r *ExpenseRepo) Create(ctx context.Context, expense *domain.Expense) error {
	query := `
		INSERT INTO expenses (
			id, tenant_id, project_id, category_id, description,
			amount, currency, expense_date, receipt_number, vendor,
			status, approved_by, approved_at, notes, created_by
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9, $10,
			$11, $12, $13, $14, $15
		) RETURNING created_at, updated_at`

	err := r.pool.QueryRow(ctx, query,
		expense.ID, expense.TenantID, nilUUID(expense.ProjectID), nilUUID(expense.CategoryID), expense.Description,
		expense.Amount, expense.Currency, expense.ExpenseDate, nilIfEmpty(expense.ReceiptNumber), nilIfEmpty(expense.Vendor),
		expense.Status, nilUUID(expense.ApprovedBy), expense.ApprovedAt, nilIfEmpty(expense.Notes), nilUUID(expense.CreatedBy),
	).Scan(&expense.CreatedAt, &expense.UpdatedAt)
	if err != nil {
		return fmt.Errorf("expense_repo: create: %w", err)
	}
	return nil
}

func (r *ExpenseRepo) GetByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*domain.Expense, error) {
	query := fmt.Sprintf(`SELECT %s FROM expenses WHERE id = $1 AND tenant_id = $2`, expenseColumns)
	exp, err := scanExpense(r.pool.QueryRow(ctx, query, id, tenantID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("expense_repo: get_by_id: %w", err)
	}
	return exp, nil
}

func (r *ExpenseRepo) List(ctx context.Context, tenantID uuid.UUID, filter domain.ExpenseFilter) ([]*domain.Expense, int64, error) {
	conditions := []string{"tenant_id = $1"}
	args := []interface{}{tenantID}
	argIdx := 2

	if filter.ProjectID != "" {
		conditions = append(conditions, fmt.Sprintf("project_id = $%d", argIdx))
		args = append(args, filter.ProjectID)
		argIdx++
	}
	if filter.Status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, filter.Status)
		argIdx++
	}

	where := strings.Join(conditions, " AND ")

	var total int64
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM expenses WHERE %s`, where)
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("expense_repo: list count: %w", err)
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
		`SELECT %s FROM expenses WHERE %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		expenseColumns, where, argIdx, argIdx+1,
	)
	args = append(args, perPage, offset)

	rows, err := r.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("expense_repo: list query: %w", err)
	}
	defer rows.Close()

	var items []*domain.Expense
	for rows.Next() {
		exp, scanErr := scanExpense(rows)
		if scanErr != nil {
			return nil, 0, fmt.Errorf("expense_repo: list scan: %w", scanErr)
		}
		items = append(items, exp)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("expense_repo: list rows: %w", err)
	}
	return items, total, nil
}

func (r *ExpenseRepo) Update(ctx context.Context, expense *domain.Expense) error {
	query := `
		UPDATE expenses SET
			project_id = $3, category_id = $4, description = $5,
			amount = $6, currency = $7, expense_date = $8,
			receipt_number = $9, vendor = $10,
			status = $11, approved_by = $12, approved_at = $13,
			notes = $14, updated_at = NOW()
		WHERE id = $1 AND tenant_id = $2
		RETURNING updated_at`

	err := r.pool.QueryRow(ctx, query,
		expense.ID, expense.TenantID,
		nilUUID(expense.ProjectID), nilUUID(expense.CategoryID), expense.Description,
		expense.Amount, expense.Currency, expense.ExpenseDate,
		nilIfEmpty(expense.ReceiptNumber), nilIfEmpty(expense.Vendor),
		expense.Status, nilUUID(expense.ApprovedBy), expense.ApprovedAt,
		nilIfEmpty(expense.Notes),
	).Scan(&expense.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apperrors.ErrNotFound
		}
		return fmt.Errorf("expense_repo: update: %w", err)
	}
	return nil
}

func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func nilUUID(id uuid.UUID) *uuid.UUID {
	if id == uuid.Nil {
		return nil
	}
	return &id
}

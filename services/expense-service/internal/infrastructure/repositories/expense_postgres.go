package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jeckersberger/rentflow/pkg/common/database"
	"github.com/jeckersberger/rentflow/services/expense-service/internal/domain"
)

type ExpensePostgres struct {
	db *database.PostgresPool
}

func NewExpensePostgres(db *database.PostgresPool) *ExpensePostgres {
	return &ExpensePostgres{db: db}
}

func (r *ExpensePostgres) CreateExpense(ctx context.Context, exp *domain.Expense) error {
	query := `
		INSERT INTO expenses.expenses (
			id, tenant_id, vendor, amount, currency, tax_rate, tax_amount, net_amount,
			category_code, date, payment_method, receipt_ref, ocr_data, status, project_id,
			approved_by, notes, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19
		)
	`

	ocrData := toJSONB(exp.OCRData)

	_, err := r.db.Exec(ctx, query,
		exp.ID, exp.TenantID, exp.Vendor, exp.Amount, exp.Currency, exp.TaxRate, exp.TaxAmount,
		exp.NetAmount, exp.CategoryCode, exp.Date, exp.PaymentMethod, exp.ReceiptRef, ocrData,
		string(exp.Status), exp.ProjectID, exp.ApprovedBy, exp.Notes, exp.CreatedAt, exp.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create expense: %w", err)
	}

	return nil
}

func (r *ExpensePostgres) UpdateExpense(ctx context.Context, exp *domain.Expense) error {
	query := `
		UPDATE expenses.expenses SET
			vendor = $3, amount = $4, currency = $5, tax_rate = $6, tax_amount = $7,
			net_amount = $8, category_code = $9, payment_method = $10, receipt_ref = $11,
			ocr_data = $12, status = $13, project_id = $14, approved_by = $15, notes = $16,
			updated_at = $17
		WHERE id = $1 AND tenant_id = $2
	`

	ocrData := toJSONB(exp.OCRData)

	result, err := r.db.Exec(ctx, query,
		exp.ID, exp.TenantID, exp.Vendor, exp.Amount, exp.Currency, exp.TaxRate, exp.TaxAmount,
		exp.NetAmount, exp.CategoryCode, exp.PaymentMethod, exp.ReceiptRef, ocrData,
		string(exp.Status), exp.ProjectID, exp.ApprovedBy, exp.Notes, exp.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update expense: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("expense not found")
	}

	return nil
}

func (r *ExpensePostgres) GetExpense(ctx context.Context, tenantID, expenseID string) (*domain.Expense, error) {
	query := `
		SELECT id, tenant_id, vendor, amount, currency, tax_rate, tax_amount, net_amount,
		       category_code, date, payment_method, receipt_ref, ocr_data, status, project_id,
		       approved_by, notes, created_at, updated_at
		FROM expenses.expenses
		WHERE id = $1 AND tenant_id = $2
	`

	var exp domain.Expense
	var ocrDataJSON []byte

	err := r.db.QueryRow(ctx, query, expenseID, tenantID).Scan(
		&exp.ID, &exp.TenantID, &exp.Vendor, &exp.Amount, &exp.Currency, &exp.TaxRate,
		&exp.TaxAmount, &exp.NetAmount, &exp.CategoryCode, &exp.Date, &exp.PaymentMethod,
		&exp.ReceiptRef, &ocrDataJSON, (*string)(&exp.Status), &exp.ProjectID,
		&exp.ApprovedBy, &exp.Notes, &exp.CreatedAt, &exp.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("expense not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get expense: %w", err)
	}

	exp.OCRData = fromJSONB(ocrDataJSON)

	return &exp, nil
}

func (r *ExpensePostgres) ListExpenses(ctx context.Context, tenantID string, limit, offset int) ([]*domain.Expense, int64, error) {
	countQuery := `SELECT COUNT(*) FROM expenses.expenses WHERE tenant_id = $1`
	var total int64
	if err := r.db.QueryRow(ctx, countQuery, tenantID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count expenses: %w", err)
	}

	query := `
		SELECT id, tenant_id, vendor, amount, currency, tax_rate, tax_amount, net_amount,
		       category_code, date, payment_method, receipt_ref, ocr_data, status, project_id,
		       approved_by, notes, created_at, updated_at
		FROM expenses.expenses
		WHERE tenant_id = $1
		ORDER BY date DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list expenses: %w", err)
	}
	defer rows.Close()

	var expenses []*domain.Expense
	for rows.Next() {
		var exp domain.Expense
		var ocrDataJSON []byte

		err := rows.Scan(
			&exp.ID, &exp.TenantID, &exp.Vendor, &exp.Amount, &exp.Currency, &exp.TaxRate,
			&exp.TaxAmount, &exp.NetAmount, &exp.CategoryCode, &exp.Date, &exp.PaymentMethod,
			&exp.ReceiptRef, &ocrDataJSON, (*string)(&exp.Status), &exp.ProjectID,
			&exp.ApprovedBy, &exp.Notes, &exp.CreatedAt, &exp.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan expense: %w", err)
		}

		exp.OCRData = fromJSONB(ocrDataJSON)
		expenses = append(expenses, &exp)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows error: %w", err)
	}

	return expenses, total, nil
}

func (r *ExpensePostgres) DeleteExpense(ctx context.Context, tenantID, expenseID string) error {
	query := `DELETE FROM expenses.expenses WHERE id = $1 AND tenant_id = $2`

	result, err := r.db.Exec(ctx, query, expenseID, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete expense: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("expense not found")
	}

	return nil
}

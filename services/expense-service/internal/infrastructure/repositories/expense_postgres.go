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
			id, tenant_id, type, vendor, vendor_address, vendor_vat_id, vendor_iban,
			amount, currency, tax_rate, tax_amount, net_amount,
			category_code, booking_account, date, service_period_from, service_period_to,
			due_date, discount_percent, discount_days, discount_deadline,
			payment_method, payment_status, paid_at, invoice_number,
			receipt_ref, receipt_checksum, receipt_nas_path,
			ocr_data, ocr_confidence, source, email_ref,
			status, project_id, approved_by, notes,
			entertainment_location, entertainment_reason, entertainment_guests,
			entertainment_tip, entertainment_deductible, entertainment_non_deductible,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17,
			$18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31, $32,
			$33, $34, $35, $36, $37, $38, $39, $40, $41, $42, $43, $44
		)
	`

	ocrData := toJSONB(exp.OCRData)

	_, err := r.db.Exec(ctx, query,
		exp.ID, exp.TenantID, string(exp.Type), exp.Vendor, exp.VendorAddress, exp.VendorVATID, exp.VendorIBAN,
		exp.Amount, exp.Currency, exp.TaxRate, exp.TaxAmount, exp.NetAmount,
		exp.CategoryCode, exp.BookingAccount, exp.Date, exp.ServicePeriodFrom, exp.ServicePeriodTo,
		exp.DueDate, exp.DiscountPercent, exp.DiscountDays, exp.DiscountDeadline,
		exp.PaymentMethod, string(exp.PaymentStatus), exp.PaidAt, exp.InvoiceNumber,
		exp.ReceiptRef, exp.ReceiptChecksum, exp.ReceiptNASPath,
		ocrData, exp.OCRConfidence, exp.Source, exp.EmailRef,
		string(exp.Status), exp.ProjectID, exp.ApprovedBy, exp.Notes,
		exp.EntertainmentLocation, exp.EntertainmentReason, exp.EntertainmentGuests,
		exp.EntertainmentTip, exp.EntertainmentDeductible, exp.EntertainmentNonDeductible,
		exp.CreatedAt, exp.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create expense: %w", err)
	}

	return nil
}

func (r *ExpensePostgres) UpdateExpense(ctx context.Context, exp *domain.Expense) error {
	query := `
		UPDATE expenses.expenses SET
			type = $3, vendor = $4, vendor_address = $5, vendor_vat_id = $6, vendor_iban = $7,
			amount = $8, currency = $9, tax_rate = $10, tax_amount = $11, net_amount = $12,
			category_code = $13, booking_account = $14,
			service_period_from = $15, service_period_to = $16,
			due_date = $17, discount_percent = $18, discount_days = $19, discount_deadline = $20,
			payment_method = $21, payment_status = $22, paid_at = $23, invoice_number = $24,
			receipt_ref = $25, receipt_checksum = $26, receipt_nas_path = $27,
			ocr_data = $28, ocr_confidence = $29, source = $30, email_ref = $31,
			status = $32, project_id = $33, approved_by = $34, notes = $35,
			entertainment_location = $36, entertainment_reason = $37, entertainment_guests = $38,
			entertainment_tip = $39, entertainment_deductible = $40, entertainment_non_deductible = $41,
			updated_at = $42
		WHERE id = $1 AND tenant_id = $2
	`

	ocrData := toJSONB(exp.OCRData)

	result, err := r.db.Exec(ctx, query,
		exp.ID, exp.TenantID,
		string(exp.Type), exp.Vendor, exp.VendorAddress, exp.VendorVATID, exp.VendorIBAN,
		exp.Amount, exp.Currency, exp.TaxRate, exp.TaxAmount, exp.NetAmount,
		exp.CategoryCode, exp.BookingAccount,
		exp.ServicePeriodFrom, exp.ServicePeriodTo,
		exp.DueDate, exp.DiscountPercent, exp.DiscountDays, exp.DiscountDeadline,
		exp.PaymentMethod, string(exp.PaymentStatus), exp.PaidAt, exp.InvoiceNumber,
		exp.ReceiptRef, exp.ReceiptChecksum, exp.ReceiptNASPath,
		ocrData, exp.OCRConfidence, exp.Source, exp.EmailRef,
		string(exp.Status), exp.ProjectID, exp.ApprovedBy, exp.Notes,
		exp.EntertainmentLocation, exp.EntertainmentReason, exp.EntertainmentGuests,
		exp.EntertainmentTip, exp.EntertainmentDeductible, exp.EntertainmentNonDeductible,
		exp.UpdatedAt,
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
		SELECT id, tenant_id, COALESCE(type, 'invoice'), COALESCE(vendor, ''),
		       COALESCE(vendor_address, ''), COALESCE(vendor_vat_id, ''), COALESCE(vendor_iban, ''),
		       amount, COALESCE(currency, 'EUR'), tax_rate, tax_amount, net_amount,
		       COALESCE(category_code, ''), COALESCE(booking_account, ''),
		       date, service_period_from, service_period_to,
		       due_date, COALESCE(discount_percent, 0), COALESCE(discount_days, 0), discount_deadline,
		       COALESCE(payment_method, ''), COALESCE(payment_status, 'open'), paid_at,
		       COALESCE(invoice_number, ''),
		       COALESCE(receipt_ref, ''), COALESCE(receipt_checksum, ''), COALESCE(receipt_nas_path, ''),
		       ocr_data, COALESCE(ocr_confidence, 0), COALESCE(source, 'manual'), COALESCE(email_ref, ''),
		       COALESCE(status, 'draft'), COALESCE(project_id, ''),
		       COALESCE(approved_by, ''), COALESCE(notes, ''),
		       COALESCE(entertainment_location, ''), COALESCE(entertainment_reason, ''),
		       COALESCE(entertainment_guests, ''), COALESCE(entertainment_tip, 0),
		       COALESCE(entertainment_deductible, 0), COALESCE(entertainment_non_deductible, 0),
		       created_at, updated_at
		FROM expenses.expenses
		WHERE id = $1 AND tenant_id = $2
	`

	var exp domain.Expense
	var ocrDataJSON []byte
	var expType, paymentStatus string

	err := r.db.QueryRow(ctx, query, expenseID, tenantID).Scan(
		&exp.ID, &exp.TenantID, &expType, &exp.Vendor,
		&exp.VendorAddress, &exp.VendorVATID, &exp.VendorIBAN,
		&exp.Amount, &exp.Currency, &exp.TaxRate, &exp.TaxAmount, &exp.NetAmount,
		&exp.CategoryCode, &exp.BookingAccount,
		&exp.Date, &exp.ServicePeriodFrom, &exp.ServicePeriodTo,
		&exp.DueDate, &exp.DiscountPercent, &exp.DiscountDays, &exp.DiscountDeadline,
		&exp.PaymentMethod, &paymentStatus, &exp.PaidAt,
		&exp.InvoiceNumber,
		&exp.ReceiptRef, &exp.ReceiptChecksum, &exp.ReceiptNASPath,
		&ocrDataJSON, &exp.OCRConfidence, &exp.Source, &exp.EmailRef,
		(*string)(&exp.Status), &exp.ProjectID,
		&exp.ApprovedBy, &exp.Notes,
		&exp.EntertainmentLocation, &exp.EntertainmentReason,
		&exp.EntertainmentGuests, &exp.EntertainmentTip,
		&exp.EntertainmentDeductible, &exp.EntertainmentNonDeductible,
		&exp.CreatedAt, &exp.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("expense not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get expense: %w", err)
	}

	exp.Type = domain.ExpenseType(expType)
	exp.PaymentStatus = domain.PaymentStatus(paymentStatus)
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
		SELECT id, tenant_id, COALESCE(type, 'invoice'), COALESCE(vendor, ''),
		       COALESCE(vendor_address, ''), COALESCE(vendor_vat_id, ''), COALESCE(vendor_iban, ''),
		       amount, COALESCE(currency, 'EUR'), tax_rate, tax_amount, net_amount,
		       COALESCE(category_code, ''), COALESCE(booking_account, ''),
		       date, service_period_from, service_period_to,
		       due_date, COALESCE(discount_percent, 0), COALESCE(discount_days, 0), discount_deadline,
		       COALESCE(payment_method, ''), COALESCE(payment_status, 'open'), paid_at,
		       COALESCE(invoice_number, ''),
		       COALESCE(receipt_ref, ''), COALESCE(receipt_checksum, ''), COALESCE(receipt_nas_path, ''),
		       ocr_data, COALESCE(ocr_confidence, 0), COALESCE(source, 'manual'), COALESCE(email_ref, ''),
		       COALESCE(status, 'draft'), COALESCE(project_id, ''),
		       COALESCE(approved_by, ''), COALESCE(notes, ''),
		       COALESCE(entertainment_location, ''), COALESCE(entertainment_reason, ''),
		       COALESCE(entertainment_guests, ''), COALESCE(entertainment_tip, 0),
		       COALESCE(entertainment_deductible, 0), COALESCE(entertainment_non_deductible, 0),
		       created_at, updated_at
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
		var expType, paymentStatus string

		err := rows.Scan(
			&exp.ID, &exp.TenantID, &expType, &exp.Vendor,
			&exp.VendorAddress, &exp.VendorVATID, &exp.VendorIBAN,
			&exp.Amount, &exp.Currency, &exp.TaxRate, &exp.TaxAmount, &exp.NetAmount,
			&exp.CategoryCode, &exp.BookingAccount,
			&exp.Date, &exp.ServicePeriodFrom, &exp.ServicePeriodTo,
			&exp.DueDate, &exp.DiscountPercent, &exp.DiscountDays, &exp.DiscountDeadline,
			&exp.PaymentMethod, &paymentStatus, &exp.PaidAt,
			&exp.InvoiceNumber,
			&exp.ReceiptRef, &exp.ReceiptChecksum, &exp.ReceiptNASPath,
			&ocrDataJSON, &exp.OCRConfidence, &exp.Source, &exp.EmailRef,
			(*string)(&exp.Status), &exp.ProjectID,
			&exp.ApprovedBy, &exp.Notes,
			&exp.EntertainmentLocation, &exp.EntertainmentReason,
			&exp.EntertainmentGuests, &exp.EntertainmentTip,
			&exp.EntertainmentDeductible, &exp.EntertainmentNonDeductible,
			&exp.CreatedAt, &exp.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan expense: %w", err)
		}

		exp.Type = domain.ExpenseType(expType)
		exp.PaymentStatus = domain.PaymentStatus(paymentStatus)
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

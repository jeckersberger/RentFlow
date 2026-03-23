package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/jeckersberger/rentflow/pkg/common/database"
	"github.com/jeckersberger/rentflow/services/invoice-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/invoice-service/internal/ports"
)

type InvoicePostgres struct {
	db *database.PostgresPool
}

func NewInvoicePostgres(db *database.PostgresPool) *InvoicePostgres {
	return &InvoicePostgres{db: db}
}

func (r *InvoicePostgres) Create(ctx context.Context, inv *domain.Invoice) error {
	query := `
		INSERT INTO invoice.invoices (
			id, tenant_id, invoice_number, project_id, client_name, client_address_street,
			client_address_city, client_address_postcode, client_address_country,
			client_email, client_tax_id, sub_total, tax_rate, tax_amount, total,
			currency, status, issue_date, due_date, paid_date, payment_method,
			payment_ref, notes, internal_notes, pdf_ref, hash, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15,
			$16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28
		)
	`

	_, err := r.db.Exec(ctx, query,
		inv.ID, inv.TenantID, inv.InvoiceNumber, inv.ProjectID, inv.ClientName,
		inv.ClientAddress.Street, inv.ClientAddress.City, inv.ClientAddress.PostCode,
		inv.ClientAddress.Country, inv.ClientEmail, inv.ClientTaxID,
		inv.SubTotal, inv.TaxRate, inv.TaxAmount, inv.Total, inv.Currency,
		string(inv.Status), inv.IssueDate, inv.DueDate, inv.PaidDate,
		inv.PaymentMethod, inv.PaymentRef, inv.Notes, inv.InternalNotes,
		inv.PDFRef, inv.Hash, inv.CreatedAt, inv.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create invoice: %w", err)
	}

	// Insert items
	for _, item := range inv.Items {
		itemQuery := `
			INSERT INTO invoice.invoice_items (
				id, invoice_id, description, quantity, unit, unit_price, total_price, tax_rate, equipment_id
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`
		_, err := r.db.Exec(ctx, itemQuery,
			item.ID, inv.ID, item.Description, item.Quantity, item.Unit,
			item.UnitPrice, item.TotalPrice, item.TaxRate, item.EquipmentID,
		)
		if err != nil {
			return fmt.Errorf("failed to create invoice item: %w", err)
		}
	}

	return nil
}

// invoiceSelectColumns returns the SELECT column list with COALESCE for nullable string fields
const invoiceSelectColumns = `id, tenant_id, invoice_number, project_id, client_name, client_address_street,
		       client_address_city, client_address_postcode, client_address_country,
		       client_email, client_tax_id, sub_total, tax_rate, tax_amount, total,
		       currency, status, issue_date, due_date, paid_date,
		       COALESCE(payment_method, '') as payment_method,
		       COALESCE(payment_ref, '') as payment_ref,
		       COALESCE(notes, '') as notes,
		       COALESCE(internal_notes, '') as internal_notes,
		       COALESCE(pdf_ref, '') as pdf_ref,
		       hash, created_at, updated_at`

func (r *InvoicePostgres) scanInvoice(scanner interface{ Scan(...interface{}) error }) (*domain.Invoice, error) {
	inv := &domain.Invoice{}
	err := scanner.Scan(
		&inv.ID, &inv.TenantID, &inv.InvoiceNumber, &inv.ProjectID, &inv.ClientName,
		&inv.ClientAddress.Street, &inv.ClientAddress.City, &inv.ClientAddress.PostCode,
		&inv.ClientAddress.Country, &inv.ClientEmail, &inv.ClientTaxID,
		&inv.SubTotal, &inv.TaxRate, &inv.TaxAmount, &inv.Total, &inv.Currency,
		&inv.Status, &inv.IssueDate, &inv.DueDate, &inv.PaidDate,
		&inv.PaymentMethod, &inv.PaymentRef, &inv.Notes, &inv.InternalNotes,
		&inv.PDFRef, &inv.Hash, &inv.CreatedAt, &inv.UpdatedAt,
	)
	return inv, err
}

func (r *InvoicePostgres) GetByID(ctx context.Context, tenantID, invoiceID string) (*domain.Invoice, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM invoice.invoices
		WHERE id = $1 AND tenant_id = $2
	`, invoiceSelectColumns)

	row := r.db.QueryRow(ctx, query, invoiceID, tenantID)

	inv, err := r.scanInvoice(row)
	if err == sql.ErrNoRows {
		return nil, domain.ErrInvoiceNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}

	// Load items
	items, err := r.getInvoiceItems(ctx, invoiceID)
	if err != nil {
		return nil, fmt.Errorf("failed to load invoice items: %w", err)
	}
	inv.Items = items

	return inv, nil
}

func (r *InvoicePostgres) GetByNumber(ctx context.Context, tenantID, invoiceNumber string) (*domain.Invoice, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM invoice.invoices
		WHERE invoice_number = $1 AND tenant_id = $2
	`, invoiceSelectColumns)

	row := r.db.QueryRow(ctx, query, invoiceNumber, tenantID)

	inv, err := r.scanInvoice(row)
	if err == sql.ErrNoRows {
		return nil, domain.ErrInvoiceNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}

	items, err := r.getInvoiceItems(ctx, inv.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load invoice items: %w", err)
	}
	inv.Items = items

	return inv, nil
}

func (r *InvoicePostgres) List(ctx context.Context, query *ports.InvoiceListQuery) (*ports.InvoiceListResult, error) {
	// Build WHERE clause
	var whereConditions []string
	var args []interface{}
	argCount := 1

	whereConditions = append(whereConditions, fmt.Sprintf("tenant_id = $%d", argCount))
	args = append(args, query.TenantID)
	argCount++

	if query.Status != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("status = $%d", argCount))
		args = append(args, string(*query.Status))
		argCount++
	}

	if query.ClientName != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("client_name ILIKE $%d", argCount))
		args = append(args, "%"+*query.ClientName+"%")
		argCount++
	}

	if query.FromDate != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("issue_date >= $%d::date", argCount))
		args = append(args, *query.FromDate)
		argCount++
	}

	if query.ToDate != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("issue_date <= $%d::date", argCount))
		args = append(args, *query.ToDate)
		argCount++
	}

	whereClause := strings.Join(whereConditions, " AND ")

	// Count query
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM invoice.invoices WHERE %s", whereClause)
	var total int64
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count invoices: %w", err)
	}

	// List query
	listQuery := fmt.Sprintf(`
		SELECT %s
		FROM invoice.invoices
		WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, invoiceSelectColumns, whereClause, argCount, argCount+1)

	args = append(args, query.Limit, query.Offset)

	rows, err := r.db.Query(ctx, listQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list invoices: %w", err)
	}
	defer rows.Close()

	var invoices []*domain.Invoice
	for rows.Next() {
		inv, err := r.scanInvoice(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan invoice: %w", err)
		}

		items, err := r.getInvoiceItems(ctx, inv.ID)
		if err == nil {
			inv.Items = items
		}

		invoices = append(invoices, inv)
	}

	return &ports.InvoiceListResult{
		Items:  invoices,
		Total:  total,
		Limit:  query.Limit,
		Offset: query.Offset,
	}, nil
}

func (r *InvoicePostgres) Update(ctx context.Context, inv *domain.Invoice) error {
	// Cannot update finalized invoices (GoBD compliance)
	if inv.Status != domain.InvoiceDraft {
		// Verify hash on finalized invoices
		if !inv.VerifyHash() {
			return fmt.Errorf("cannot update invoice: hash mismatch")
		}
	}

	query := `
		UPDATE invoice.invoices SET
			client_name = $3, client_address_street = $4, client_address_city = $5,
			client_address_postcode = $6, client_address_country = $7, client_email = $8,
			client_tax_id = $9, sub_total = $10, tax_rate = $11, tax_amount = $12,
			total = $13, currency = $14, status = $15, issue_date = $16, due_date = $17,
			paid_date = $18, payment_method = $19, payment_ref = $20, notes = $21,
			internal_notes = $22, pdf_ref = $23, hash = $24, updated_at = $25
		WHERE id = $1 AND tenant_id = $2
	`

	_, err := r.db.Exec(ctx, query,
		inv.ID, inv.TenantID, inv.ClientName,
		inv.ClientAddress.Street, inv.ClientAddress.City, inv.ClientAddress.PostCode,
		inv.ClientAddress.Country, inv.ClientEmail, inv.ClientTaxID,
		inv.SubTotal, inv.TaxRate, inv.TaxAmount, inv.Total, inv.Currency,
		string(inv.Status), inv.IssueDate, inv.DueDate, inv.PaidDate,
		inv.PaymentMethod, inv.PaymentRef, inv.Notes, inv.InternalNotes,
		inv.PDFRef, inv.Hash, inv.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update invoice: %w", err)
	}

	// Clear and re-insert items
	_, err = r.db.Exec(ctx, "DELETE FROM invoice.invoice_items WHERE invoice_id = $1", inv.ID)
	if err != nil {
		return fmt.Errorf("failed to delete old items: %w", err)
	}

	for _, item := range inv.Items {
		itemQuery := `
			INSERT INTO invoice.invoice_items (
				id, invoice_id, description, quantity, unit, unit_price, total_price, tax_rate, equipment_id
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`
		_, err := r.db.Exec(ctx, itemQuery,
			item.ID, inv.ID, item.Description, item.Quantity, item.Unit,
			item.UnitPrice, item.TotalPrice, item.TaxRate, item.EquipmentID,
		)
		if err != nil {
			return fmt.Errorf("failed to update invoice item: %w", err)
		}
	}

	return nil
}

func (r *InvoicePostgres) Delete(ctx context.Context, tenantID, invoiceID string) error {
	// Soft delete by marking as cancelled
	query := `
		UPDATE invoice.invoices
		SET status = 'cancelled', updated_at = NOW()
		WHERE id = $1 AND tenant_id = $2
	`

	_, err := r.db.Exec(ctx, query, invoiceID, tenantID)
	return err
}

// Helper methods

func (r *InvoicePostgres) getInvoiceItems(ctx context.Context, invoiceID string) ([]domain.InvoiceItem, error) {
	query := `
		SELECT id, description, quantity, unit, unit_price, total_price, tax_rate, equipment_id
		FROM invoice.invoice_items
		WHERE invoice_id = $1
		ORDER BY id
	`

	rows, err := r.db.Query(ctx, query, invoiceID)
	if err != nil {
		return nil, fmt.Errorf("failed to query items: %w", err)
	}
	defer rows.Close()

	var items []domain.InvoiceItem
	for rows.Next() {
		item := domain.InvoiceItem{}
		err := rows.Scan(
			&item.ID, &item.Description, &item.Quantity, &item.Unit,
			&item.UnitPrice, &item.TotalPrice, &item.TaxRate, &item.EquipmentID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan item: %w", err)
		}
		items = append(items, item)
	}

	return items, nil
}

// GetNextSequenceNumber returns the next sequential invoice number for a tenant
func (r *InvoicePostgres) GetNextSequenceNumber(ctx context.Context, tenantID string) (int, error) {
	var nextNum int

	query := `
		INSERT INTO invoice.number_sequences (tenant_id, sequence_type, next_value, created_at)
		VALUES ($1, 'invoice', 2, NOW())
		ON CONFLICT (tenant_id, sequence_type)
		DO UPDATE SET next_value = invoice.number_sequences.next_value + 1, updated_at = NOW()
		RETURNING next_value - 1
	`

	err := r.db.QueryRow(ctx, query, tenantID).Scan(&nextNum)
	if err != nil {
		return 0, fmt.Errorf("failed to get next invoice sequence number: %w", err)
	}

	return nextNum, nil
}

func (r *InvoicePostgres) GetOverdueInvoices(ctx context.Context, tenantID string) ([]*domain.Invoice, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM invoice.invoices
		WHERE tenant_id = $1 AND status IN ('sent', 'overdue')
		ORDER BY due_date ASC
	`, invoiceSelectColumns)

	rows, err := r.db.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get overdue invoices: %w", err)
	}
	defer rows.Close()

	invoices := make([]*domain.Invoice, 0)
	for rows.Next() {
		inv, err := r.scanInvoice(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan invoice: %w", err)
		}
		invoices = append(invoices, inv)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating invoices: %w", err)
	}

	return invoices, nil
}

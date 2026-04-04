package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jeckersberger/EquipFlow/services/invoice/internal/domain"

	apperrors "github.com/jeckersberger/EquipFlow/pkg/common/errors"
)

const invoiceColumns = `
	id, tenant_id, invoice_number, invoice_type, status,
	customer_name, customer_email, customer_address,
	invoice_date, due_date, vat_rate, kleinunternehmer,
	total_net, total_vat, total_gross, amount_paid,
	notes, hash, previous_hash, finalized_at,
	created_at, updated_at`

type InvoiceRepo struct {
	pool *pgxpool.Pool
}

func NewInvoiceRepo(pool *pgxpool.Pool) *InvoiceRepo {
	return &InvoiceRepo{pool: pool}
}

func scanInvoice(row pgx.Row) (*domain.Invoice, error) {
	inv := &domain.Invoice{}
	var (
		customerEmail   *string
		customerAddress *string
		invoiceDate     time.Time
		dueDate         *time.Time
		notes           *string
		hash            *string
		previousHash    *string
	)

	err := row.Scan(
		&inv.ID, &inv.TenantID, &inv.InvoiceNumber, &inv.InvoiceType, &inv.Status,
		&inv.CustomerName, &customerEmail, &customerAddress,
		&invoiceDate, &dueDate, &inv.VatRate, &inv.Kleinunternehmer,
		&inv.TotalNet, &inv.TotalVat, &inv.TotalGross, &inv.AmountPaid,
		&notes, &hash, &previousHash, &inv.FinalizedAt,
		&inv.CreatedAt, &inv.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	inv.InvoiceDate = invoiceDate.Format("2006-01-02")
	if dueDate != nil {
		inv.DueDate = dueDate.Format("2006-01-02")
	}
	inv.CustomerEmail = derefString(customerEmail)
	inv.CustomerAddress = derefString(customerAddress)
	inv.Notes = derefString(notes)
	inv.Hash = derefString(hash)
	inv.PreviousHash = derefString(previousHash)

	return inv, nil
}

func (r *InvoiceRepo) Create(ctx context.Context, invoice *domain.Invoice) error {
	query := `
		INSERT INTO invoices (
			id, tenant_id, invoice_number, invoice_type, status,
			customer_name, customer_email, customer_address,
			invoice_date, due_date, vat_rate, kleinunternehmer,
			total_net, total_vat, total_gross, amount_paid,
			notes, hash, previous_hash, finalized_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8,
			$9, $10, $11, $12,
			$13, $14, $15, $16,
			$17, $18, $19, $20
		) RETURNING created_at, updated_at`

	err := r.pool.QueryRow(ctx, query,
		invoice.ID, invoice.TenantID, invoice.InvoiceNumber, invoice.InvoiceType, invoice.Status,
		invoice.CustomerName, nilIfEmpty(invoice.CustomerEmail), nilIfEmpty(invoice.CustomerAddress),
		invoice.InvoiceDate, nilIfEmpty(invoice.DueDate), invoice.VatRate, invoice.Kleinunternehmer,
		invoice.TotalNet, invoice.TotalVat, invoice.TotalGross, invoice.AmountPaid,
		nilIfEmpty(invoice.Notes), nilIfEmpty(invoice.Hash), nilIfEmpty(invoice.PreviousHash), invoice.FinalizedAt,
	).Scan(&invoice.CreatedAt, &invoice.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return apperrors.ErrConflict
		}
		return fmt.Errorf("invoice_repo: create: %w", err)
	}
	return nil
}

func (r *InvoiceRepo) GetByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*domain.Invoice, error) {
	query := fmt.Sprintf(`SELECT %s FROM invoices WHERE id = $1 AND tenant_id = $2`, invoiceColumns)
	inv, err := scanInvoice(r.pool.QueryRow(ctx, query, id, tenantID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("invoice_repo: get_by_id: %w", err)
	}
	return inv, nil
}

func (r *InvoiceRepo) List(ctx context.Context, tenantID uuid.UUID, filter domain.InvoiceFilter) ([]*domain.Invoice, int64, error) {
	conditions := []string{"tenant_id = $1"}
	args := []interface{}{tenantID}
	argIdx := 2

	if filter.Status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, filter.Status)
		argIdx++
	}

	where := strings.Join(conditions, " AND ")

	var total int64
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM invoices WHERE %s`, where)
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("invoice_repo: list count: %w", err)
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
		`SELECT %s FROM invoices WHERE %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		invoiceColumns, where, argIdx, argIdx+1,
	)
	args = append(args, perPage, offset)

	rows, err := r.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("invoice_repo: list query: %w", err)
	}
	defer rows.Close()

	var items []*domain.Invoice
	for rows.Next() {
		inv, scanErr := scanInvoice(rows)
		if scanErr != nil {
			return nil, 0, fmt.Errorf("invoice_repo: list scan: %w", scanErr)
		}
		items = append(items, inv)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("invoice_repo: list rows: %w", err)
	}
	return items, total, nil
}

func (r *InvoiceRepo) Update(ctx context.Context, invoice *domain.Invoice) error {
	query := `
		UPDATE invoices SET
			customer_name = $3, customer_email = $4, customer_address = $5,
			invoice_date = $6, due_date = $7, vat_rate = $8, kleinunternehmer = $9,
			total_net = $10, total_vat = $11, total_gross = $12, amount_paid = $13,
			notes = $14, status = $15,
			hash = $16, previous_hash = $17, finalized_at = $18,
			updated_at = NOW()
		WHERE id = $1 AND tenant_id = $2
		RETURNING updated_at`

	err := r.pool.QueryRow(ctx, query,
		invoice.ID, invoice.TenantID,
		invoice.CustomerName, nilIfEmpty(invoice.CustomerEmail), nilIfEmpty(invoice.CustomerAddress),
		invoice.InvoiceDate, nilIfEmpty(invoice.DueDate), invoice.VatRate, invoice.Kleinunternehmer,
		invoice.TotalNet, invoice.TotalVat, invoice.TotalGross, invoice.AmountPaid,
		nilIfEmpty(invoice.Notes), invoice.Status,
		nilIfEmpty(invoice.Hash), nilIfEmpty(invoice.PreviousHash), invoice.FinalizedAt,
	).Scan(&invoice.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apperrors.ErrNotFound
		}
		return fmt.Errorf("invoice_repo: update: %w", err)
	}
	return nil
}

func (r *InvoiceRepo) UpdateStatus(ctx context.Context, id uuid.UUID, tenantID uuid.UUID, status string) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE invoices SET status = $3, updated_at = NOW() WHERE id = $1 AND tenant_id = $2`,
		id, tenantID, status,
	)
	if err != nil {
		return fmt.Errorf("invoice_repo: update_status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

func (r *InvoiceRepo) Delete(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error {
	tag, err := r.pool.Exec(ctx,
		`DELETE FROM invoices WHERE id = $1 AND tenant_id = $2 AND status = 'draft'`,
		id, tenantID,
	)
	if err != nil {
		return fmt.Errorf("invoice_repo: delete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

func (r *InvoiceRepo) UpdateAmountPaid(ctx context.Context, id uuid.UUID, tenantID uuid.UUID, amount int64) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE invoices SET amount_paid = $3, updated_at = NOW() WHERE id = $1 AND tenant_id = $2`,
		id, tenantID, amount,
	)
	if err != nil {
		return fmt.Errorf("invoice_repo: update_amount_paid: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

func (r *InvoiceRepo) UpdateTotals(ctx context.Context, id uuid.UUID, tenantID uuid.UUID, totalNet, totalVat, totalGross int64) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE invoices SET total_net = $3, total_vat = $4, total_gross = $5, updated_at = NOW() WHERE id = $1 AND tenant_id = $2`,
		id, tenantID, totalNet, totalVat, totalGross,
	)
	if err != nil {
		return fmt.Errorf("invoice_repo: update_totals: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

func (r *InvoiceRepo) GetLastHash(ctx context.Context, tenantID uuid.UUID) (string, error) {
	var hash *string
	err := r.pool.QueryRow(ctx,
		`SELECT hash FROM invoices WHERE tenant_id = $1 AND hash IS NOT NULL ORDER BY finalized_at DESC LIMIT 1`,
		tenantID,
	).Scan(&hash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", nil
		}
		return "", fmt.Errorf("invoice_repo: get_last_hash: %w", err)
	}
	return derefString(hash), nil
}

func (r *InvoiceRepo) Search(ctx context.Context, tenantID uuid.UUID, query string, page int, perPage int) ([]*domain.Invoice, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	offset := (page - 1) * perPage
	pattern := "%" + query + "%"

	searchCondition := `tenant_id = $1 AND (invoice_number ILIKE $2 OR customer_name ILIKE $2 OR customer_email ILIKE $2)`

	var total int64
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM invoices WHERE %s`, searchCondition)
	err := r.pool.QueryRow(ctx, countQuery, tenantID, pattern).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("invoice_repo: search count: %w", err)
	}

	dataQuery := fmt.Sprintf(
		`SELECT %s FROM invoices WHERE %s ORDER BY created_at DESC LIMIT $3 OFFSET $4`,
		invoiceColumns, searchCondition,
	)

	rows, err := r.pool.Query(ctx, dataQuery, tenantID, pattern, perPage, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("invoice_repo: search query: %w", err)
	}
	defer rows.Close()

	var items []*domain.Invoice
	for rows.Next() {
		inv, scanErr := scanInvoice(rows)
		if scanErr != nil {
			return nil, 0, fmt.Errorf("invoice_repo: search scan: %w", scanErr)
		}
		items = append(items, inv)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("invoice_repo: search rows: %w", err)
	}
	return items, total, nil
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

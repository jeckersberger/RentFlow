package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jeckersberger/EquipFlow/pkg/common/database"
	"github.com/jeckersberger/EquipFlow/services/invoice/internal/domain"
)

// paymentColumns lists all columns of the payments table.
const paymentColumns = `
	id, tenant_id, invoice_id, amount, payment_date,
	payment_method, reference, created_at`

// PaymentRepo implements domain.PaymentRepository using PostgreSQL.
type PaymentRepo struct {
	pool *pgxpool.Pool
	db   database.DBTX
}

// NewPaymentRepo creates a new PaymentRepo.
func NewPaymentRepo(pool *pgxpool.Pool) *PaymentRepo {
	return &PaymentRepo{pool: pool}
}

// WithTx returns a new PaymentRepo that runs queries against the given transaction.
func (r *PaymentRepo) WithTx(tx database.DBTX) domain.PaymentRepository {
	return &PaymentRepo{pool: r.pool, db: tx}
}

func (r *PaymentRepo) conn() database.DBTX {
	if r.db != nil {
		return r.db
	}
	return r.pool
}

// scanPayment scans a single payment row into a domain.Payment.
func scanPayment(row pgx.Row) (*domain.Payment, error) {
	p := &domain.Payment{}
	var (
		reference   *string
		paymentDate time.Time
	)

	err := row.Scan(
		&p.ID, &p.TenantID, &p.InvoiceID, &p.Amount, &paymentDate,
		&p.PaymentMethod, &reference, &p.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	p.PaymentDate = paymentDate.Format("2006-01-02")
	p.Reference = derefString(reference)

	return p, nil
}

// Create inserts a new payment record and scans back the generated fields.
func (r *PaymentRepo) Create(ctx context.Context, payment *domain.Payment) error {
	query := `
		INSERT INTO payments (
			id, tenant_id, invoice_id, amount, payment_date,
			payment_method, reference
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7
		) RETURNING created_at`

	err := r.conn().QueryRow(ctx, query,
		payment.ID, payment.TenantID, payment.InvoiceID,
		payment.Amount, payment.PaymentDate,
		payment.PaymentMethod, nilIfEmpty(payment.Reference),
	).Scan(&payment.CreatedAt)
	if err != nil {
		return fmt.Errorf("payment_repo: create: %w", err)
	}
	return nil
}

// ListByInvoice returns all payments for a given invoice scoped to a tenant, ordered by payment_date.
func (r *PaymentRepo) ListByInvoice(ctx context.Context, invoiceID uuid.UUID, tenantID uuid.UUID) ([]*domain.Payment, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM payments WHERE invoice_id = $1 AND tenant_id = $2 ORDER BY payment_date ASC, created_at ASC`,
		paymentColumns,
	)

	rows, err := r.conn().Query(ctx, query, invoiceID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("payment_repo: list_by_invoice query: %w", err)
	}
	defer rows.Close()

	var items []*domain.Payment
	for rows.Next() {
		p, scanErr := scanPayment(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("payment_repo: list_by_invoice scan: %w", scanErr)
		}
		items = append(items, p)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("payment_repo: list_by_invoice rows: %w", err)
	}
	return items, nil
}

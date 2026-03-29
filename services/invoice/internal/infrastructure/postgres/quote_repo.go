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

const quoteColumns = `
	id, tenant_id, quote_number, status,
	customer_name, customer_email, customer_address,
	project_id, subject, intro_text, outro_text,
	quote_date, valid_until, vat_rate, kleinunternehmer,
	total_net, total_vat, total_gross,
	payment_terms_days, discount_pct, notes,
	converted_invoice_id, created_at, updated_at`

// QuoteRepo implements domain.QuoteRepository using PostgreSQL.
type QuoteRepo struct {
	pool *pgxpool.Pool
}

// NewQuoteRepo creates a new QuoteRepo.
func NewQuoteRepo(pool *pgxpool.Pool) *QuoteRepo {
	return &QuoteRepo{pool: pool}
}

// scanQuote scans a single quote row into a domain.Quote.
func scanQuote(row pgx.Row) (*domain.Quote, error) {
	q := &domain.Quote{}
	var (
		customerEmail    *string
		customerAddress  *string
		projectID        *uuid.UUID
		subject          *string
		introText        *string
		outroText        *string
		quoteDate        time.Time
		validUntil       *time.Time
		notes            *string
		convertedInvID   *uuid.UUID
		paymentTermsDays *int
		discountPct      *int64
	)

	err := row.Scan(
		&q.ID, &q.TenantID, &q.QuoteNumber, &q.Status,
		&q.CustomerName, &customerEmail, &customerAddress,
		&projectID, &subject, &introText, &outroText,
		&quoteDate, &validUntil, &q.VatRate, &q.Kleinunternehmer,
		&q.TotalNet, &q.TotalVat, &q.TotalGross,
		&paymentTermsDays, &discountPct, &notes,
		&convertedInvID, &q.CreatedAt, &q.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	q.QuoteDate = quoteDate.Format("2006-01-02")
	if validUntil != nil {
		q.ValidUntil = validUntil.Format("2006-01-02")
	}
	q.CustomerEmail = derefString(customerEmail)
	q.CustomerAddress = derefString(customerAddress)
	q.ProjectID = projectID
	q.Subject = derefString(subject)
	q.IntroText = derefString(introText)
	q.OutroText = derefString(outroText)
	q.Notes = derefString(notes)
	q.ConvertedInvoiceID = convertedInvID
	if paymentTermsDays != nil {
		q.PaymentTermsDays = *paymentTermsDays
	}
	if discountPct != nil {
		q.DiscountPct = *discountPct
	}

	return q, nil
}

// Create inserts a new quote and scans back the generated fields.
func (r *QuoteRepo) Create(ctx context.Context, q *domain.Quote) error {
	query := `
		INSERT INTO quotes (
			id, tenant_id, quote_number, status,
			customer_name, customer_email, customer_address,
			project_id, subject, intro_text, outro_text,
			quote_date, valid_until, vat_rate, kleinunternehmer,
			total_net, total_vat, total_gross,
			payment_terms_days, discount_pct, notes,
			converted_invoice_id
		) VALUES (
			$1, $2, $3, $4,
			$5, $6, $7,
			$8, $9, $10, $11,
			$12, $13, $14, $15,
			$16, $17, $18,
			$19, $20, $21,
			$22
		) RETURNING created_at, updated_at`

	err := r.pool.QueryRow(ctx, query,
		q.ID, q.TenantID, q.QuoteNumber, q.Status,
		q.CustomerName, nilIfEmpty(q.CustomerEmail), nilIfEmpty(q.CustomerAddress),
		q.ProjectID, nilIfEmpty(q.Subject), nilIfEmpty(q.IntroText), nilIfEmpty(q.OutroText),
		q.QuoteDate, nilIfEmpty(q.ValidUntil), q.VatRate, q.Kleinunternehmer,
		q.TotalNet, q.TotalVat, q.TotalGross,
		q.PaymentTermsDays, q.DiscountPct, nilIfEmpty(q.Notes),
		q.ConvertedInvoiceID,
	).Scan(&q.CreatedAt, &q.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return apperrors.ErrConflict
		}
		return fmt.Errorf("quote_repo: create: %w", err)
	}
	return nil
}

// GetByID returns a single quote by ID scoped to a tenant.
func (r *QuoteRepo) GetByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*domain.Quote, error) {
	query := fmt.Sprintf(`SELECT %s FROM quotes WHERE id = $1 AND tenant_id = $2`, quoteColumns)
	q, err := scanQuote(r.pool.QueryRow(ctx, query, id, tenantID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("quote_repo: get_by_id: %w", err)
	}
	return q, nil
}

// List returns a paginated list of quotes for a tenant, optionally filtered by status.
func (r *QuoteRepo) List(ctx context.Context, tenantID uuid.UUID, filter domain.QuoteFilter) ([]*domain.Quote, int64, error) {
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
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM quotes WHERE %s`, where)
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("quote_repo: list count: %w", err)
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
		`SELECT %s FROM quotes WHERE %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		quoteColumns, where, argIdx, argIdx+1,
	)
	args = append(args, perPage, offset)

	rows, err := r.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("quote_repo: list query: %w", err)
	}
	defer rows.Close()

	var items []*domain.Quote
	for rows.Next() {
		q, scanErr := scanQuote(rows)
		if scanErr != nil {
			return nil, 0, fmt.Errorf("quote_repo: list scan: %w", scanErr)
		}
		items = append(items, q)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("quote_repo: list rows: %w", err)
	}
	return items, total, nil
}

// Update persists all mutable fields of a quote.
func (r *QuoteRepo) Update(ctx context.Context, q *domain.Quote) error {
	query := `
		UPDATE quotes SET
			customer_name = $3, customer_email = $4, customer_address = $5,
			project_id = $6, subject = $7, intro_text = $8, outro_text = $9,
			quote_date = $10, valid_until = $11, vat_rate = $12, kleinunternehmer = $13,
			total_net = $14, total_vat = $15, total_gross = $16,
			payment_terms_days = $17, discount_pct = $18, notes = $19,
			status = $20, converted_invoice_id = $21,
			updated_at = NOW()
		WHERE id = $1 AND tenant_id = $2
		RETURNING updated_at`

	err := r.pool.QueryRow(ctx, query,
		q.ID, q.TenantID,
		q.CustomerName, nilIfEmpty(q.CustomerEmail), nilIfEmpty(q.CustomerAddress),
		q.ProjectID, nilIfEmpty(q.Subject), nilIfEmpty(q.IntroText), nilIfEmpty(q.OutroText),
		q.QuoteDate, nilIfEmpty(q.ValidUntil), q.VatRate, q.Kleinunternehmer,
		q.TotalNet, q.TotalVat, q.TotalGross,
		q.PaymentTermsDays, q.DiscountPct, nilIfEmpty(q.Notes),
		q.Status, q.ConvertedInvoiceID,
	).Scan(&q.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apperrors.ErrNotFound
		}
		return fmt.Errorf("quote_repo: update: %w", err)
	}
	return nil
}

// UpdateStatus changes only the status column of a quote.
func (r *QuoteRepo) UpdateStatus(ctx context.Context, id uuid.UUID, tenantID uuid.UUID, status string) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE quotes SET status = $3, updated_at = NOW() WHERE id = $1 AND tenant_id = $2`,
		id, tenantID, status,
	)
	if err != nil {
		return fmt.Errorf("quote_repo: update_status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

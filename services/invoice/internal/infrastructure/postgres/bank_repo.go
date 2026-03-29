package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	apperrors "github.com/jeckersberger/EquipFlow/pkg/common/errors"
	"github.com/jeckersberger/EquipFlow/services/invoice/internal/domain"
)

// bankTxColumns lists all columns of the bank_transactions table.
const bankTxColumns = `
	id, tenant_id, booking_date, value_date, amount,
	currency, reference, counterparty_name, counterparty_iban,
	matched_invoice_id, match_confidence, import_source,
	import_batch_id, created_at`

// BankRepo implements domain.BankRepository using PostgreSQL.
type BankRepo struct {
	pool *pgxpool.Pool
}

// NewBankRepo creates a new BankRepo.
func NewBankRepo(pool *pgxpool.Pool) *BankRepo {
	return &BankRepo{pool: pool}
}

// Create inserts a new bank transaction.
func (r *BankRepo) Create(ctx context.Context, tx *domain.BankTransaction) error {
	query := `
		INSERT INTO bank_transactions (
			id, tenant_id, booking_date, value_date, amount,
			currency, reference, counterparty_name, counterparty_iban,
			matched_invoice_id, match_confidence, import_source,
			import_batch_id
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9,
			$10, $11, $12,
			$13
		) RETURNING created_at`

	err := r.pool.QueryRow(ctx, query,
		tx.ID, tx.TenantID, tx.BookingDate, nilIfEmpty(tx.ValueDate), tx.Amount,
		tx.Currency, nilIfEmpty(tx.Reference), nilIfEmpty(tx.CounterpartyName), nilIfEmpty(tx.CounterpartyIBAN),
		tx.MatchedInvoiceID, tx.MatchConfidence, nilIfEmpty(tx.ImportSource),
		tx.ImportBatchID,
	).Scan(&tx.CreatedAt)
	if err != nil {
		return fmt.Errorf("bank_repo: create: %w", err)
	}
	return nil
}

// List returns bank transactions for a tenant, optionally filtered by match status.
func (r *BankRepo) List(ctx context.Context, tenantID uuid.UUID, matched *bool, limit int) ([]*domain.BankTransaction, error) {
	if limit <= 0 {
		limit = 100
	}

	where := "tenant_id = $1"
	args := []interface{}{tenantID}
	argIdx := 2

	if matched != nil {
		if *matched {
			where += fmt.Sprintf(" AND matched_invoice_id IS NOT NULL AND match_confidence != $%d", argIdx)
			args = append(args, "none")
		} else {
			where += fmt.Sprintf(" AND (matched_invoice_id IS NULL OR match_confidence = $%d)", argIdx)
			args = append(args, "none")
		}
		argIdx++
	}

	query := fmt.Sprintf(
		`SELECT %s FROM bank_transactions WHERE %s ORDER BY booking_date DESC, created_at DESC LIMIT $%d`,
		bankTxColumns, where, argIdx,
	)
	args = append(args, limit)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("bank_repo: list: %w", err)
	}
	defer rows.Close()

	var items []*domain.BankTransaction
	for rows.Next() {
		tx := &domain.BankTransaction{}
		var (
			bookingDate      time.Time
			valueDate        *time.Time
			reference        *string
			counterpartyName *string
			counterpartyIBAN *string
			importSource     *string
		)

		err := rows.Scan(
			&tx.ID, &tx.TenantID, &bookingDate, &valueDate, &tx.Amount,
			&tx.Currency, &reference, &counterpartyName, &counterpartyIBAN,
			&tx.MatchedInvoiceID, &tx.MatchConfidence, &importSource,
			&tx.ImportBatchID, &tx.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("bank_repo: list scan: %w", err)
		}

		tx.BookingDate = bookingDate.Format("2006-01-02")
		if valueDate != nil {
			tx.ValueDate = valueDate.Format("2006-01-02")
		}
		tx.Reference = derefString(reference)
		tx.CounterpartyName = derefString(counterpartyName)
		tx.CounterpartyIBAN = derefString(counterpartyIBAN)
		tx.ImportSource = derefString(importSource)

		items = append(items, tx)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("bank_repo: list rows: %w", err)
	}
	return items, nil
}

// UpdateMatch sets the matched invoice and confidence on a bank transaction.
func (r *BankRepo) UpdateMatch(ctx context.Context, id uuid.UUID, tenantID uuid.UUID, invoiceID uuid.UUID, confidence string) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE bank_transactions SET matched_invoice_id = $3, match_confidence = $4 WHERE id = $1 AND tenant_id = $2`,
		id, tenantID, invoiceID, confidence,
	)
	if err != nil {
		return fmt.Errorf("bank_repo: update_match: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

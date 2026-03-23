package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/database"
	"github.com/jeckersberger/rentflow/services/invoice-service/internal/domain"
)

type CreditNotePostgres struct {
	db *database.PostgresPool
}

func NewCreditNotePostgres(db *database.PostgresPool) *CreditNotePostgres {
	return &CreditNotePostgres{db: db}
}

func (r *CreditNotePostgres) Create(ctx context.Context, cn *domain.CreditNote) error {
	itemsJSON, err := json.Marshal(cn.Items)
	if err != nil {
		return fmt.Errorf("failed to marshal items: %w", err)
	}

	query := `INSERT INTO invoice.credit_notes (
		id, tenant_id, credit_note_number, original_invoice_id, original_invoice_number,
		client_name, client_email, items, subtotal, tax_rate, tax_amount, total,
		currency, reason, status, issued_at, created_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)`

	_, err = r.db.Exec(ctx, query,
		cn.ID, cn.TenantID, cn.CreditNoteNumber, cn.OriginalInvoiceID, cn.OriginalInvoiceNumber,
		cn.ClientName, cn.ClientEmail, string(itemsJSON), cn.SubTotal, cn.TaxRate, cn.TaxAmount, cn.Total,
		cn.Currency, cn.Reason, cn.Status, cn.IssuedAt, cn.CreatedAt,
	)
	return err
}

func (r *CreditNotePostgres) GetByID(ctx context.Context, tenantID, creditNoteID string) (*domain.CreditNote, error) {
	query := `SELECT id, tenant_id, credit_note_number, original_invoice_id, original_invoice_number,
		client_name, client_email, items, subtotal, tax_rate, tax_amount, total,
		currency, reason, status, issued_at, created_at
		FROM invoice.credit_notes WHERE tenant_id = $1 AND id = $2`

	row := r.db.QueryRow(ctx, query, tenantID, creditNoteID)

	cn := &domain.CreditNote{}
	var itemsJSON string
	var issuedAt *time.Time

	err := row.Scan(
		&cn.ID, &cn.TenantID, &cn.CreditNoteNumber, &cn.OriginalInvoiceID, &cn.OriginalInvoiceNumber,
		&cn.ClientName, &cn.ClientEmail, &itemsJSON, &cn.SubTotal, &cn.TaxRate, &cn.TaxAmount, &cn.Total,
		&cn.Currency, &cn.Reason, &cn.Status, &issuedAt, &cn.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("credit note not found: %w", err)
	}

	cn.IssuedAt = issuedAt
	if err := json.Unmarshal([]byte(itemsJSON), &cn.Items); err != nil {
		return nil, fmt.Errorf("failed to unmarshal items: %w", err)
	}

	return cn, nil
}

func (r *CreditNotePostgres) Update(ctx context.Context, cn *domain.CreditNote) error {
	itemsJSON, err := json.Marshal(cn.Items)
	if err != nil {
		return fmt.Errorf("failed to marshal items: %w", err)
	}

	query := `UPDATE invoice.credit_notes SET
		status = $1, issued_at = $2, items = $3, subtotal = $4, tax_amount = $5, total = $6, reason = $7
		WHERE tenant_id = $8 AND id = $9`

	_, err = r.db.Exec(ctx, query,
		cn.Status, cn.IssuedAt, string(itemsJSON), cn.SubTotal, cn.TaxAmount, cn.Total, cn.Reason,
		cn.TenantID, cn.ID,
	)
	return err
}

func (r *CreditNotePostgres) Delete(ctx context.Context, tenantID, creditNoteID string) error {
	query := `DELETE FROM invoice.credit_notes WHERE tenant_id = $1 AND id = $2`
	_, err := r.db.Exec(ctx, query, tenantID, creditNoteID)
	return err
}

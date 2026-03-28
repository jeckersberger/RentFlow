package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jeckersberger/EquipFlow/services/invoice/internal/domain"

	apperrors "github.com/jeckersberger/EquipFlow/pkg/common/errors"
)

// invoiceItemColumns lists all columns of the invoice_items table.
const invoiceItemColumns = `
	id, tenant_id, invoice_id, description, quantity,
	unit, unit_price, position, created_at, updated_at`

// ItemRepo implements domain.InvoiceItemRepository using PostgreSQL.
type ItemRepo struct {
	pool *pgxpool.Pool
}

// NewItemRepo creates a new ItemRepo.
func NewItemRepo(pool *pgxpool.Pool) *ItemRepo {
	return &ItemRepo{pool: pool}
}

// scanInvoiceItem scans a single invoice_item row into a domain.InvoiceItem.
func scanInvoiceItem(row pgx.Row) (*domain.InvoiceItem, error) {
	item := &domain.InvoiceItem{}

	err := row.Scan(
		&item.ID, &item.TenantID, &item.InvoiceID, &item.Description, &item.Quantity,
		&item.Unit, &item.UnitPrice, &item.Position, &item.CreatedAt, &item.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return item, nil
}

// Create inserts a new invoice item and scans back the generated fields.
func (r *ItemRepo) Create(ctx context.Context, item *domain.InvoiceItem) error {
	query := `
		INSERT INTO invoice_items (
			id, tenant_id, invoice_id, description, quantity,
			unit, unit_price, position
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8
		) RETURNING created_at, updated_at`

	err := r.pool.QueryRow(ctx, query,
		item.ID, item.TenantID, item.InvoiceID, item.Description, item.Quantity,
		item.Unit, item.UnitPrice, item.Position,
	).Scan(&item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		return fmt.Errorf("item_repo: create: %w", err)
	}
	return nil
}

// ListByInvoice returns all items for a given invoice scoped to a tenant, ordered by position.
func (r *ItemRepo) ListByInvoice(ctx context.Context, invoiceID uuid.UUID, tenantID uuid.UUID) ([]*domain.InvoiceItem, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM invoice_items WHERE invoice_id = $1 AND tenant_id = $2 ORDER BY position ASC`,
		invoiceItemColumns,
	)

	rows, err := r.pool.Query(ctx, query, invoiceID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("item_repo: list_by_invoice query: %w", err)
	}
	defer rows.Close()

	var items []*domain.InvoiceItem
	for rows.Next() {
		item, scanErr := scanInvoiceItem(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("item_repo: list_by_invoice scan: %w", scanErr)
		}
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("item_repo: list_by_invoice rows: %w", err)
	}
	return items, nil
}

// Delete removes a single invoice item by ID scoped to a tenant.
func (r *ItemRepo) Delete(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error {
	tag, err := r.pool.Exec(ctx,
		`DELETE FROM invoice_items WHERE id = $1 AND tenant_id = $2`,
		id, tenantID,
	)
	if err != nil {
		return fmt.Errorf("item_repo: delete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

// DeleteAllByInvoice removes all items belonging to an invoice scoped to a tenant.
func (r *ItemRepo) DeleteAllByInvoice(ctx context.Context, invoiceID uuid.UUID, tenantID uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM invoice_items WHERE invoice_id = $1 AND tenant_id = $2`,
		invoiceID, tenantID,
	)
	if err != nil {
		return fmt.Errorf("item_repo: delete_all_by_invoice: %w", err)
	}
	return nil
}

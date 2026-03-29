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

// quoteItemColumns lists all columns of the quote_items table.
const quoteItemColumns = `
	id, tenant_id, quote_id, description, quantity,
	unit, unit_price, discount_pct, position, created_at, updated_at`

// QuoteItemRepo implements domain.QuoteItemRepository using PostgreSQL.
type QuoteItemRepo struct {
	pool *pgxpool.Pool
}

// NewQuoteItemRepo creates a new QuoteItemRepo.
func NewQuoteItemRepo(pool *pgxpool.Pool) *QuoteItemRepo {
	return &QuoteItemRepo{pool: pool}
}

// scanQuoteItem scans a single quote_item row into a domain.QuoteItem.
func scanQuoteItem(row pgx.Row) (*domain.QuoteItem, error) {
	item := &domain.QuoteItem{}
	var discountPct *int64

	err := row.Scan(
		&item.ID, &item.TenantID, &item.QuoteID, &item.Description, &item.Quantity,
		&item.Unit, &item.UnitPrice, &discountPct, &item.Position, &item.CreatedAt, &item.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if discountPct != nil {
		item.DiscountPct = *discountPct
	}

	return item, nil
}

// Create inserts a new quote item and scans back the generated fields.
func (r *QuoteItemRepo) Create(ctx context.Context, item *domain.QuoteItem) error {
	query := `
		INSERT INTO quote_items (
			id, tenant_id, quote_id, description, quantity,
			unit, unit_price, discount_pct, position
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9
		) RETURNING created_at, updated_at`

	err := r.pool.QueryRow(ctx, query,
		item.ID, item.TenantID, item.QuoteID, item.Description, item.Quantity,
		item.Unit, item.UnitPrice, item.DiscountPct, item.Position,
	).Scan(&item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		return fmt.Errorf("quote_item_repo: create: %w", err)
	}
	return nil
}

// ListByQuote returns all items for a given quote scoped to a tenant, ordered by position.
func (r *QuoteItemRepo) ListByQuote(ctx context.Context, quoteID uuid.UUID, tenantID uuid.UUID) ([]*domain.QuoteItem, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM quote_items WHERE quote_id = $1 AND tenant_id = $2 ORDER BY position ASC`,
		quoteItemColumns,
	)

	rows, err := r.pool.Query(ctx, query, quoteID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("quote_item_repo: list_by_quote query: %w", err)
	}
	defer rows.Close()

	var items []*domain.QuoteItem
	for rows.Next() {
		item, scanErr := scanQuoteItem(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("quote_item_repo: list_by_quote scan: %w", scanErr)
		}
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("quote_item_repo: list_by_quote rows: %w", err)
	}
	return items, nil
}

// Delete removes a single quote item by ID scoped to a tenant.
func (r *QuoteItemRepo) Delete(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error {
	tag, err := r.pool.Exec(ctx,
		`DELETE FROM quote_items WHERE id = $1 AND tenant_id = $2`,
		id, tenantID,
	)
	if err != nil {
		return fmt.Errorf("quote_item_repo: delete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

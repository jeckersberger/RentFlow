package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jeckersberger/EquipFlow/services/transport/internal/domain"
)

// itemColumns lists all columns of the transport_items table for consistent scanning.
const itemColumns = `id, order_id, equipment_id, quantity, notes`

// ItemRepo implements domain.TransportItemRepository using PostgreSQL.
type ItemRepo struct {
	pool *pgxpool.Pool
}

// NewItemRepo creates a new ItemRepo.
func NewItemRepo(pool *pgxpool.Pool) *ItemRepo {
	return &ItemRepo{pool: pool}
}

// scanItem scans a single transport_items row into a domain.TransportItem.
func scanItem(row pgx.Row) (*domain.TransportItem, error) {
	i := &domain.TransportItem{}
	var notes *string

	err := row.Scan(&i.ID, &i.OrderID, &i.EquipmentID, &i.Quantity, &notes)
	if err != nil {
		return nil, err
	}

	if notes != nil {
		i.Notes = *notes
	}

	return i, nil
}

// Create inserts a new transport item record.
func (r *ItemRepo) Create(ctx context.Context, item *domain.TransportItem) error {
	query := `
		INSERT INTO transport_items (id, order_id, equipment_id, quantity, notes)
		VALUES ($1, $2, $3, $4, $5)`

	_, err := r.pool.Exec(ctx, query,
		item.ID, item.OrderID, item.EquipmentID, item.Quantity, nilIfEmpty(item.Notes),
	)
	if err != nil {
		return fmt.Errorf("item_repo: create: %w", err)
	}
	return nil
}

// ListByOrder returns all transport items for a given order.
func (r *ItemRepo) ListByOrder(ctx context.Context, orderID uuid.UUID) ([]*domain.TransportItem, error) {
	query := fmt.Sprintf(`SELECT %s FROM transport_items WHERE order_id = $1 ORDER BY id ASC`, itemColumns)

	rows, err := r.pool.Query(ctx, query, orderID)
	if err != nil {
		return nil, fmt.Errorf("item_repo: list_by_order query: %w", err)
	}
	defer rows.Close()

	var items []*domain.TransportItem
	for rows.Next() {
		i, scanErr := scanItem(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("item_repo: list_by_order scan: %w", scanErr)
		}
		items = append(items, i)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("item_repo: list_by_order rows: %w", err)
	}
	return items, nil
}

// Ensure interface compliance at compile time.
var _ domain.TransportItemRepository = (*ItemRepo)(nil)

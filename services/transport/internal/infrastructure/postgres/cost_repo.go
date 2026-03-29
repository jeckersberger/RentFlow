package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jeckersberger/EquipFlow/services/transport/internal/domain"
)

// costColumns lists all columns of the transport_costs table for consistent scanning.
const costColumns = `id, tenant_id, order_id, cost_type, amount_cents, notes, created_at`

// CostRepo implements domain.TransportCostRepository using PostgreSQL.
type CostRepo struct {
	pool *pgxpool.Pool
}

// NewCostRepo creates a new CostRepo.
func NewCostRepo(pool *pgxpool.Pool) *CostRepo {
	return &CostRepo{pool: pool}
}

// scanCost scans a single transport_costs row into a domain.TransportCost.
func scanCost(row pgx.Row) (*domain.TransportCost, error) {
	c := &domain.TransportCost{}
	var notes *string

	err := row.Scan(&c.ID, &c.TenantID, &c.OrderID, &c.CostType, &c.AmountCents, &notes, &c.CreatedAt)
	if err != nil {
		return nil, err
	}

	if notes != nil {
		c.Notes = *notes
	}

	return c, nil
}

// Create inserts a new transport cost record.
func (r *CostRepo) Create(ctx context.Context, cost *domain.TransportCost) error {
	query := `
		INSERT INTO transport_costs (id, tenant_id, order_id, cost_type, amount_cents, notes)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at`

	err := r.pool.QueryRow(ctx, query,
		cost.ID, cost.TenantID, cost.OrderID,
		cost.CostType, cost.AmountCents, nilIfEmpty(cost.Notes),
	).Scan(&cost.CreatedAt)
	if err != nil {
		return fmt.Errorf("cost_repo: create: %w", err)
	}
	return nil
}

// ListByOrder returns all cost entries for a given transport order.
func (r *CostRepo) ListByOrder(ctx context.Context, orderID uuid.UUID) ([]*domain.TransportCost, error) {
	query := fmt.Sprintf(`SELECT %s FROM transport_costs WHERE order_id = $1 ORDER BY created_at DESC`, costColumns)

	rows, err := r.pool.Query(ctx, query, orderID)
	if err != nil {
		return nil, fmt.Errorf("cost_repo: list_by_order query: %w", err)
	}
	defer rows.Close()

	var items []*domain.TransportCost
	for rows.Next() {
		c, scanErr := scanCost(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("cost_repo: list_by_order scan: %w", scanErr)
		}
		items = append(items, c)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("cost_repo: list_by_order rows: %w", err)
	}
	return items, nil
}

// Ensure interface compliance at compile time.
var _ domain.TransportCostRepository = (*CostRepo)(nil)

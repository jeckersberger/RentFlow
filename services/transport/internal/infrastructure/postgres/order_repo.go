package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jeckersberger/EquipFlow/services/transport/internal/domain"

	apperrors "github.com/jeckersberger/EquipFlow/pkg/common/errors"
)

// orderColumns lists all columns of the transport_orders table for consistent scanning.
const orderColumns = `
	id, tenant_id, project_id, vehicle_id, driver_id, type, status,
	pickup_address, delivery_address, scheduled_at, completed_at,
	notes, created_by, created_at, updated_at`

// OrderRepo implements domain.TransportOrderRepository using PostgreSQL.
type OrderRepo struct {
	pool *pgxpool.Pool
}

// NewOrderRepo creates a new OrderRepo.
func NewOrderRepo(pool *pgxpool.Pool) *OrderRepo {
	return &OrderRepo{pool: pool}
}

// scanOrder scans a single transport_orders row into a domain.TransportOrder.
func scanOrder(row pgx.Row) (*domain.TransportOrder, error) {
	o := &domain.TransportOrder{}
	var (
		oType           *string
		status          *string
		pickupAddress   *string
		deliveryAddress *string
		notes           *string
	)

	err := row.Scan(
		&o.ID, &o.TenantID, &o.ProjectID, &o.VehicleID, &o.DriverID,
		&oType, &status,
		&pickupAddress, &deliveryAddress, &o.ScheduledAt, &o.CompletedAt,
		&notes, &o.CreatedBy, &o.CreatedAt, &o.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if oType != nil {
		o.Type = *oType
	}
	if status != nil {
		o.Status = *status
	}
	if pickupAddress != nil {
		o.PickupAddress = *pickupAddress
	}
	if deliveryAddress != nil {
		o.DeliveryAddress = *deliveryAddress
	}
	if notes != nil {
		o.Notes = *notes
	}

	return o, nil
}

// Create inserts a new transport order record.
func (r *OrderRepo) Create(ctx context.Context, order *domain.TransportOrder) error {
	query := `
		INSERT INTO transport_orders (
			id, tenant_id, project_id, vehicle_id, driver_id, type, status,
			pickup_address, delivery_address, scheduled_at,
			notes, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING created_at, updated_at`

	err := r.pool.QueryRow(ctx, query,
		order.ID, order.TenantID, order.ProjectID, order.VehicleID, order.DriverID,
		nilIfEmpty(order.Type), nilIfEmpty(order.Status),
		nilIfEmpty(order.PickupAddress), nilIfEmpty(order.DeliveryAddress),
		order.ScheduledAt,
		nilIfEmpty(order.Notes), order.CreatedBy,
	).Scan(&order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		return fmt.Errorf("order_repo: create: %w", err)
	}
	return nil
}

// List returns a filtered, paginated list of transport orders for a tenant plus total count.
func (r *OrderRepo) List(ctx context.Context, tenantID uuid.UUID, filter domain.TransportOrderFilter) ([]*domain.TransportOrder, int64, error) {
	// Build dynamic WHERE clause.
	conditions := []string{"tenant_id = $1"}
	args := []interface{}{tenantID}
	argIdx := 2

	if filter.ProjectID != nil {
		conditions = append(conditions, fmt.Sprintf("project_id = $%d", argIdx))
		args = append(args, *filter.ProjectID)
		argIdx++
	}
	if filter.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, *filter.Status)
		argIdx++
	}

	where := strings.Join(conditions, " AND ")

	// Count query.
	var total int64
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM transport_orders WHERE %s`, where)
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("order_repo: list count: %w", err)
	}

	// Data query with pagination.
	offset := (filter.Page - 1) * filter.PerPage
	dataQuery := fmt.Sprintf(
		`SELECT %s FROM transport_orders WHERE %s ORDER BY scheduled_at DESC NULLS LAST LIMIT $%d OFFSET $%d`,
		orderColumns, where, argIdx, argIdx+1,
	)
	args = append(args, filter.PerPage, offset)

	rows, err := r.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("order_repo: list query: %w", err)
	}
	defer rows.Close()

	var items []*domain.TransportOrder
	for rows.Next() {
		o, scanErr := scanOrder(rows)
		if scanErr != nil {
			return nil, 0, fmt.Errorf("order_repo: list scan: %w", scanErr)
		}
		items = append(items, o)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("order_repo: list rows: %w", err)
	}
	return items, total, nil
}

// GetByID retrieves a transport order by its primary key within a tenant scope.
func (r *OrderRepo) GetByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*domain.TransportOrder, error) {
	query := fmt.Sprintf(`SELECT %s FROM transport_orders WHERE id = $1 AND tenant_id = $2`, orderColumns)
	o, err := scanOrder(r.pool.QueryRow(ctx, query, id, tenantID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("order_repo: get_by_id: %w", err)
	}
	return o, nil
}

// Update updates an existing transport order record.
func (r *OrderRepo) Update(ctx context.Context, order *domain.TransportOrder) error {
	query := `
		UPDATE transport_orders SET
			project_id = $3,
			vehicle_id = $4,
			driver_id = $5,
			type = $6,
			status = $7,
			pickup_address = $8,
			delivery_address = $9,
			scheduled_at = $10,
			completed_at = $11,
			notes = $12,
			updated_at = NOW()
		WHERE id = $1 AND tenant_id = $2`

	tag, err := r.pool.Exec(ctx, query,
		order.ID, order.TenantID, order.ProjectID, order.VehicleID, order.DriverID,
		nilIfEmpty(order.Type), nilIfEmpty(order.Status),
		nilIfEmpty(order.PickupAddress), nilIfEmpty(order.DeliveryAddress),
		order.ScheduledAt, order.CompletedAt,
		nilIfEmpty(order.Notes),
	)
	if err != nil {
		return fmt.Errorf("order_repo: update: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

// Ensure interface compliance at compile time.
var _ domain.TransportOrderRepository = (*OrderRepo)(nil)

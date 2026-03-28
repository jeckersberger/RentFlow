package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jeckersberger/EquipFlow/services/warehouse/internal/domain"
)

// movementColumns lists all columns of the movements table.
const movementColumns = `id, tenant_id, equipment_id, from_location_id, to_location_id, quantity, reason, user_id, notes, created_at`

// MovementRepo implements domain.MovementRepository using PostgreSQL.
type MovementRepo struct {
	pool *pgxpool.Pool
}

// NewMovementRepo creates a new MovementRepo.
func NewMovementRepo(pool *pgxpool.Pool) *MovementRepo {
	return &MovementRepo{pool: pool}
}

// scanMovement scans a single movement row into a domain.Movement.
func scanMovement(row pgx.Row) (*domain.Movement, error) {
	m := &domain.Movement{}
	var reason, notes *string

	err := row.Scan(
		&m.ID, &m.TenantID, &m.EquipmentID,
		&m.FromLocationID, &m.ToLocationID,
		&m.Quantity, &reason, &m.UserID,
		&notes, &m.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	if reason != nil {
		m.Reason = *reason
	}
	if notes != nil {
		m.Notes = *notes
	}
	return m, nil
}

// Create inserts a new movement and scans back the generated fields.
func (r *MovementRepo) Create(ctx context.Context, movement *domain.Movement) error {
	query := `
		INSERT INTO movements (
			id, tenant_id, equipment_id, from_location_id, to_location_id,
			quantity, reason, user_id, notes
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		) RETURNING id, created_at`

	err := r.pool.QueryRow(ctx, query,
		movement.ID, movement.TenantID, movement.EquipmentID,
		movement.FromLocationID, movement.ToLocationID,
		movement.Quantity, nilIfEmpty(movement.Reason),
		movement.UserID, nilIfEmpty(movement.Notes),
	).Scan(&movement.ID, &movement.CreatedAt)
	if err != nil {
		return fmt.Errorf("movement_repo: create: %w", err)
	}
	return nil
}

// List returns a paginated list of movements for a tenant plus total count.
func (r *MovementRepo) List(ctx context.Context, tenantID uuid.UUID, page int, perPage int) ([]*domain.Movement, int64, error) {
	offset := (page - 1) * perPage

	var total int64
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM movements WHERE tenant_id = $1`, tenantID,
	).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("movement_repo: list count: %w", err)
	}

	query := fmt.Sprintf(
		`SELECT %s FROM movements WHERE tenant_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		movementColumns,
	)

	rows, err := r.pool.Query(ctx, query, tenantID, perPage, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("movement_repo: list query: %w", err)
	}
	defer rows.Close()

	var movements []*domain.Movement
	for rows.Next() {
		m, scanErr := scanMovement(rows)
		if scanErr != nil {
			return nil, 0, fmt.Errorf("movement_repo: list scan: %w", scanErr)
		}
		movements = append(movements, m)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("movement_repo: list rows: %w", err)
	}
	return movements, total, nil
}

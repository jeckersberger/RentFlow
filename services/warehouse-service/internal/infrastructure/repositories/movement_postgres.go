package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jeckersberger/rentflow/pkg/common/database"
	"github.com/jeckersberger/rentflow/services/warehouse-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/warehouse-service/internal/ports"
)

type MovementPostgres struct {
	db *database.PostgresPool
}

func NewMovementPostgres(db *database.PostgresPool) ports.MovementRepository {
	return &MovementPostgres{db: db}
}

func (r *MovementPostgres) Create(ctx context.Context, movement *domain.Movement) error {
	query := `
		INSERT INTO movements
		(id, tenant_id, equipment_id, from_location_id, to_location_id, movement_type, quantity, reason, user_id, project_id, timestamp, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err := r.db.Exec(ctx, query,
		movement.ID,
		movement.TenantID,
		movement.EquipmentID,
		movement.FromLocationID,
		movement.ToLocationID,
		string(movement.MovementType),
		movement.Quantity,
		movement.Reason,
		movement.UserID,
		movement.ProjectID,
		movement.Timestamp,
		movement.CreatedAt,
	)

	return err
}

func (r *MovementPostgres) GetByID(ctx context.Context, tenantID, id string) (*domain.Movement, error) {
	query := `
		SELECT id, tenant_id, equipment_id, from_location_id, to_location_id, movement_type, quantity, reason, user_id, project_id, timestamp, created_at
		FROM movements
		WHERE tenant_id = $1 AND id = $2
	`

	row := r.db.QueryRow(ctx, query, tenantID, id)
	return movementRowToMovement(row)
}

func (r *MovementPostgres) List(ctx context.Context, tenantID string, query *ports.MovementListQuery) ([]*domain.Movement, int, error) {
	where := "WHERE tenant_id = $1"
	args := []interface{}{tenantID}
	argIndex := 2

	if query.EquipmentID != nil {
		where += fmt.Sprintf(" AND equipment_id = $%d", argIndex)
		args = append(args, *query.EquipmentID)
		argIndex++
	}
	if query.FromLocation != nil {
		where += fmt.Sprintf(" AND from_location_id = $%d", argIndex)
		args = append(args, *query.FromLocation)
		argIndex++
	}
	if query.ToLocation != nil {
		where += fmt.Sprintf(" AND to_location_id = $%d", argIndex)
		args = append(args, *query.ToLocation)
		argIndex++
	}
	if query.MovementType != nil {
		where += fmt.Sprintf(" AND movement_type = $%d", argIndex)
		args = append(args, *query.MovementType)
		argIndex++
	}

	// Count query
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM movements %s", where)
	var total int
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// List query
	listQuery := fmt.Sprintf(`
		SELECT id, tenant_id, equipment_id, from_location_id, to_location_id, movement_type, quantity, reason, user_id, project_id, timestamp, created_at
		FROM movements
		%s
		ORDER BY timestamp DESC
		LIMIT $%d OFFSET $%d
	`, where, argIndex, argIndex+1)

	args = append(args, query.Limit, query.Offset)

	rows, err := r.db.Query(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	movements := make([]*domain.Movement, 0)
	for rows.Next() {
		movement, err := movementRowsToMovement(rows)
		if err != nil {
			return nil, 0, err
		}
		movements = append(movements, movement)
	}

	return movements, total, rows.Err()
}

func (r *MovementPostgres) GetByEquipmentID(ctx context.Context, tenantID, equipmentID string, limit, offset int) ([]*domain.Movement, int, error) {
	query := `
		SELECT id, tenant_id, equipment_id, from_location_id, to_location_id, movement_type, quantity, reason, user_id, project_id, timestamp, created_at
		FROM movements
		WHERE tenant_id = $1 AND equipment_id = $2
		ORDER BY timestamp DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.db.Query(ctx, query, tenantID, equipmentID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	movements := make([]*domain.Movement, 0)
	for rows.Next() {
		movement, err := movementRowsToMovement(rows)
		if err != nil {
			return nil, 0, err
		}
		movements = append(movements, movement)
	}

	// Get total count
	countQuery := "SELECT COUNT(*) FROM movements WHERE tenant_id = $1 AND equipment_id = $2"
	var total int
	if err := r.db.QueryRow(ctx, countQuery, tenantID, equipmentID).Scan(&total); err != nil {
		return nil, 0, err
	}

	return movements, total, rows.Err()
}

func movementRowToMovement(row *sql.Row) (*domain.Movement, error) {
	movement := &domain.Movement{}
	err := row.Scan(
		&movement.ID,
		&movement.TenantID,
		&movement.EquipmentID,
		&movement.FromLocationID,
		&movement.ToLocationID,
		(*string)(&movement.MovementType),
		&movement.Quantity,
		&movement.Reason,
		&movement.UserID,
		&movement.ProjectID,
		&movement.Timestamp,
		&movement.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("movement not found")
		}
		return nil, err
	}
	return movement, nil
}

func movementRowsToMovement(rows *sql.Rows) (*domain.Movement, error) {
	movement := &domain.Movement{}
	err := rows.Scan(
		&movement.ID,
		&movement.TenantID,
		&movement.EquipmentID,
		&movement.FromLocationID,
		&movement.ToLocationID,
		(*string)(&movement.MovementType),
		&movement.Quantity,
		&movement.Reason,
		&movement.UserID,
		&movement.ProjectID,
		&movement.Timestamp,
		&movement.CreatedAt,
	)
	return movement, err
}

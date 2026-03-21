package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jeckersberger/rentflow/pkg/common/database"
	"github.com/jeckersberger/rentflow/services/project-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/project-service/internal/ports"
)

type ReservationPostgres struct {
	db *database.PostgresPool
}

func NewReservationPostgres(db *database.PostgresPool) *ReservationPostgres {
	return &ReservationPostgres{db: db}
}

func (r *ReservationPostgres) Create(ctx context.Context, res *domain.Reservation) error {
	query := `
		INSERT INTO projects.reservations (
			id, tenant_id, project_id, equipment_id, start_date, end_date,
			status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.Exec(ctx, query,
		res.ID, res.TenantID, res.ProjectID, res.EquipmentID, res.StartDate,
		res.EndDate, string(res.Status), res.CreatedAt, res.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create reservation: %w", err)
	}

	return nil
}

func (r *ReservationPostgres) Update(ctx context.Context, res *domain.Reservation) error {
	query := `
		UPDATE projects.reservations SET
			status = $3, updated_at = $4
		WHERE id = $1 AND tenant_id = $2
	`

	result, err := r.db.Exec(ctx, query,
		res.ID, res.TenantID, string(res.Status), res.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update reservation: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *ReservationPostgres) GetByID(ctx context.Context, tenantID, reservationID string) (*domain.Reservation, error) {
	query := `
		SELECT id, tenant_id, project_id, equipment_id, start_date, end_date,
			   status, created_at, updated_at
		FROM projects.reservations
		WHERE id = $1 AND tenant_id = $2
	`

	res := &domain.Reservation{}
	err := r.db.QueryRow(ctx, query, reservationID, tenantID).Scan(
		&res.ID, &res.TenantID, &res.ProjectID, &res.EquipmentID, &res.StartDate,
		&res.EndDate, &res.Status, &res.CreatedAt, &res.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("failed to get reservation: %w", err)
	}

	return res, nil
}

func (r *ReservationPostgres) ListByProjectID(ctx context.Context, tenantID, projectID string) ([]*domain.Reservation, error) {
	query := `
		SELECT id, tenant_id, project_id, equipment_id, start_date, end_date,
			   status, created_at, updated_at
		FROM projects.reservations
		WHERE tenant_id = $1 AND project_id = $2
		ORDER BY start_date ASC
	`

	rows, err := r.db.Query(ctx, query, tenantID, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list reservations: %w", err)
	}
	defer rows.Close()

	reservations := make([]*domain.Reservation, 0)
	for rows.Next() {
		res := &domain.Reservation{}
		err := rows.Scan(
			&res.ID, &res.TenantID, &res.ProjectID, &res.EquipmentID, &res.StartDate,
			&res.EndDate, &res.Status, &res.CreatedAt, &res.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan reservation: %w", err)
		}
		reservations = append(reservations, res)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating reservations: %w", err)
	}

	return reservations, nil
}

func (r *ReservationPostgres) ListByEquipmentID(ctx context.Context, tenantID, equipmentID, startDate, endDate string) ([]*domain.Reservation, error) {
	query := `
		SELECT id, tenant_id, project_id, equipment_id, start_date, end_date,
			   status, created_at, updated_at
		FROM projects.reservations
		WHERE tenant_id = $1 AND equipment_id = $2 AND status != 'cancelled'
			AND (
				(start_date < $4 AND end_date > $3)
			)
		ORDER BY start_date ASC
	`

	rows, err := r.db.Query(ctx, query, tenantID, equipmentID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to check conflicts: %w", err)
	}
	defer rows.Close()

	reservations := make([]*domain.Reservation, 0)
	for rows.Next() {
		res := &domain.Reservation{}
		err := rows.Scan(
			&res.ID, &res.TenantID, &res.ProjectID, &res.EquipmentID, &res.StartDate,
			&res.EndDate, &res.Status, &res.CreatedAt, &res.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan reservation: %w", err)
		}
		reservations = append(reservations, res)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating reservations: %w", err)
	}

	return reservations, nil
}

func (r *ReservationPostgres) Delete(ctx context.Context, tenantID, reservationID string) error {
	query := `DELETE FROM projects.reservations WHERE id = $1 AND tenant_id = $2`

	result, err := r.db.Exec(ctx, query, reservationID, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete reservation: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

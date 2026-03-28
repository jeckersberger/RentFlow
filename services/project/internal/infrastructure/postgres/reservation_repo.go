package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jeckersberger/EquipFlow/services/project/internal/domain"

	apperrors "github.com/jeckersberger/EquipFlow/pkg/common/errors"
)

// reservationColumns lists all columns of the reservations table.
const reservationColumns = `
	id, tenant_id, project_id, equipment_id,
	quantity, start_date, end_date, status, created_at`

// ReservationRepo implements domain.ReservationRepository using PostgreSQL.
type ReservationRepo struct {
	pool *pgxpool.Pool
}

// NewReservationRepo creates a new ReservationRepo.
func NewReservationRepo(pool *pgxpool.Pool) *ReservationRepo {
	return &ReservationRepo{pool: pool}
}

// scanReservation scans a single reservation row into a domain.Reservation.
func scanReservation(row pgx.Row) (*domain.Reservation, error) {
	res := &domain.Reservation{}
	var status *string

	err := row.Scan(
		&res.ID, &res.TenantID, &res.ProjectID, &res.EquipmentID,
		&res.Quantity, &res.StartDate, &res.EndDate, &status, &res.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	res.Status = derefString(status)

	return res, nil
}

// Create inserts a new reservation and scans back the generated fields.
func (r *ReservationRepo) Create(ctx context.Context, reservation *domain.Reservation) error {
	query := `
		INSERT INTO reservations (
			id, tenant_id, project_id, equipment_id,
			quantity, start_date, end_date, status
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8
		) RETURNING created_at`

	err := r.pool.QueryRow(ctx, query,
		reservation.ID, reservation.TenantID, reservation.ProjectID, reservation.EquipmentID,
		reservation.Quantity, reservation.StartDate, reservation.EndDate,
		nilIfEmpty(reservation.Status),
	).Scan(&reservation.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return apperrors.ErrConflict
		}
		return fmt.Errorf("reservation_repo: create: %w", err)
	}
	return nil
}

// GetByID retrieves a reservation by primary key scoped to a tenant.
func (r *ReservationRepo) GetByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*domain.Reservation, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM reservations WHERE id = $1 AND tenant_id = $2`,
		reservationColumns,
	)
	res, err := scanReservation(r.pool.QueryRow(ctx, query, id, tenantID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("reservation_repo: get_by_id: %w", err)
	}
	return res, nil
}

// List returns all reservations for a tenant, ordered by start_date ascending.
func (r *ReservationRepo) List(ctx context.Context, tenantID uuid.UUID) ([]*domain.Reservation, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM reservations WHERE tenant_id = $1 ORDER BY start_date ASC`,
		reservationColumns,
	)

	rows, err := r.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("reservation_repo: list query: %w", err)
	}
	defer rows.Close()

	var items []*domain.Reservation
	for rows.Next() {
		res, scanErr := scanReservation(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("reservation_repo: list scan: %w", scanErr)
		}
		items = append(items, res)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("reservation_repo: list rows: %w", err)
	}
	return items, nil
}

// ListByEquipmentAndDateRange returns reservations for a given equipment item
// that overlap with the specified date range, ordered by start_date.
func (r *ReservationRepo) ListByEquipmentAndDateRange(ctx context.Context, equipmentID uuid.UUID, from time.Time, to time.Time) ([]*domain.Reservation, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM reservations
		 WHERE equipment_id = $1
		   AND start_date <= $3 AND end_date >= $2
		 ORDER BY start_date ASC`,
		reservationColumns,
	)

	rows, err := r.pool.Query(ctx, query, equipmentID, from, to)
	if err != nil {
		return nil, fmt.Errorf("reservation_repo: list_by_equipment_date_range query: %w", err)
	}
	defer rows.Close()

	var items []*domain.Reservation
	for rows.Next() {
		res, scanErr := scanReservation(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("reservation_repo: list_by_equipment_date_range scan: %w", scanErr)
		}
		items = append(items, res)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("reservation_repo: list_by_equipment_date_range rows: %w", err)
	}
	return items, nil
}

// UpdateStatus changes the status of a reservation.
func (r *ReservationRepo) UpdateStatus(ctx context.Context, id uuid.UUID, tenantID uuid.UUID, status string) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE reservations SET status = $3 WHERE id = $1 AND tenant_id = $2`,
		id, tenantID, status,
	)
	if err != nil {
		return fmt.Errorf("reservation_repo: update_status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

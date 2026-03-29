package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jeckersberger/EquipFlow/services/crew/internal/domain"
)

// availabilityColumns lists all columns of the crew_availability table for consistent scanning.
const availabilityColumns = `
	id, tenant_id, crew_member_id, date, status, note, created_at`

// AvailabilityRepo implements domain.AvailabilityRepository using PostgreSQL.
type AvailabilityRepo struct {
	pool *pgxpool.Pool
}

// NewAvailabilityRepo creates a new AvailabilityRepo.
func NewAvailabilityRepo(pool *pgxpool.Pool) *AvailabilityRepo {
	return &AvailabilityRepo{pool: pool}
}

// scanAvailability scans a single crew_availability row into a domain.CrewAvailability.
func scanAvailability(row pgx.Row) (*domain.CrewAvailability, error) {
	a := &domain.CrewAvailability{}
	var (
		date time.Time
		note *string
	)

	err := row.Scan(
		&a.ID, &a.TenantID, &a.CrewMemberID, &date, &a.Status, &note, &a.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	a.Date = date.Format("2006-01-02")
	if note != nil {
		a.Note = *note
	}

	return a, nil
}

// Upsert inserts or updates an availability record for a specific member and date.
func (r *AvailabilityRepo) Upsert(ctx context.Context, availability *domain.CrewAvailability) error {
	query := `
		INSERT INTO crew_availability (id, tenant_id, crew_member_id, date, status, note)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (tenant_id, crew_member_id, date) DO UPDATE SET
			status = EXCLUDED.status,
			note = EXCLUDED.note
		RETURNING id, created_at`

	err := r.pool.QueryRow(ctx, query,
		availability.ID, availability.TenantID, availability.CrewMemberID,
		availability.Date, availability.Status, nilIfEmpty(availability.Note),
	).Scan(&availability.ID, &availability.CreatedAt)
	if err != nil {
		return fmt.Errorf("availability_repo: upsert: %w", err)
	}
	return nil
}

// ListByMember returns all availability entries for a specific crew member within a date range.
func (r *AvailabilityRepo) ListByMember(ctx context.Context, crewMemberID uuid.UUID, tenantID uuid.UUID, from string, to string) ([]*domain.CrewAvailability, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM crew_availability
		 WHERE crew_member_id = $1 AND tenant_id = $2
		   AND date >= $3 AND date <= $4
		 ORDER BY date ASC`,
		availabilityColumns,
	)

	return r.queryAvailability(ctx, query, crewMemberID, tenantID, from, to)
}

// ListAll returns all availability entries for a tenant within a date range.
func (r *AvailabilityRepo) ListAll(ctx context.Context, tenantID uuid.UUID, from string, to string) ([]*domain.CrewAvailability, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM crew_availability
		 WHERE tenant_id = $1
		   AND date >= $2 AND date <= $3
		 ORDER BY date ASC, crew_member_id ASC`,
		availabilityColumns,
	)

	return r.queryAvailability(ctx, query, tenantID, from, to)
}

// queryAvailability is a helper that executes a query and scans rows into CrewAvailability slices.
func (r *AvailabilityRepo) queryAvailability(ctx context.Context, query string, args ...interface{}) ([]*domain.CrewAvailability, error) {
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("availability_repo: query: %w", err)
	}
	defer rows.Close()

	var items []*domain.CrewAvailability
	for rows.Next() {
		a := &domain.CrewAvailability{}
		var (
			date time.Time
			note *string
		)
		scanErr := rows.Scan(
			&a.ID, &a.TenantID, &a.CrewMemberID, &date, &a.Status, &note, &a.CreatedAt,
		)
		if scanErr != nil {
			return nil, fmt.Errorf("availability_repo: scan: %w", scanErr)
		}
		a.Date = date.Format("2006-01-02")
		if note != nil {
			a.Note = *note
		}
		items = append(items, a)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("availability_repo: rows: %w", err)
	}
	return items, nil
}

// Ensure interface compliance at compile time.
var _ domain.AvailabilityRepository = (*AvailabilityRepo)(nil)

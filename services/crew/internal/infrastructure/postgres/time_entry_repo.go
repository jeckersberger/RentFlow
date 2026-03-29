package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jeckersberger/EquipFlow/services/crew/internal/domain"
)

// timeEntryColumns lists all columns of the time_entries table for consistent scanning.
const timeEntryColumns = `
	id, tenant_id, crew_member_id, assignment_id, project_id,
	check_in_time, check_out_time, duration_minutes, break_minutes,
	notes, status, created_at`

// TimeEntryRepo implements domain.TimeEntryRepository using PostgreSQL.
type TimeEntryRepo struct {
	pool *pgxpool.Pool
}

// NewTimeEntryRepo creates a new TimeEntryRepo.
func NewTimeEntryRepo(pool *pgxpool.Pool) *TimeEntryRepo {
	return &TimeEntryRepo{pool: pool}
}

// scanTimeEntry scans a single time_entries row into a domain.TimeEntry.
func scanTimeEntry(row pgx.Row) (*domain.TimeEntry, error) {
	e := &domain.TimeEntry{}
	var (
		notes *string
	)

	err := row.Scan(
		&e.ID, &e.TenantID, &e.CrewMemberID, &e.AssignmentID, &e.ProjectID,
		&e.CheckInTime, &e.CheckOutTime, &e.DurationMinutes, &e.BreakMinutes,
		&notes, &e.Status, &e.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	if notes != nil {
		e.Notes = *notes
	}

	return e, nil
}

// Create inserts a new time entry record.
func (r *TimeEntryRepo) Create(ctx context.Context, entry *domain.TimeEntry) error {
	query := `
		INSERT INTO time_entries (
			id, tenant_id, crew_member_id, assignment_id, project_id,
			check_in_time, check_out_time, duration_minutes, break_minutes,
			notes, status
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING created_at`

	err := r.pool.QueryRow(ctx, query,
		entry.ID, entry.TenantID, entry.CrewMemberID, entry.AssignmentID, entry.ProjectID,
		entry.CheckInTime, entry.CheckOutTime, entry.DurationMinutes, entry.BreakMinutes,
		nilIfEmpty(entry.Notes), entry.Status,
	).Scan(&entry.CreatedAt)
	if err != nil {
		return fmt.Errorf("time_entry_repo: create: %w", err)
	}
	return nil
}

// Update modifies an existing time entry record (used for check-out).
func (r *TimeEntryRepo) Update(ctx context.Context, entry *domain.TimeEntry) error {
	query := `
		UPDATE time_entries SET
			check_out_time = $3, duration_minutes = $4, break_minutes = $5,
			notes = $6, status = $7
		WHERE id = $1 AND tenant_id = $2`

	tag, err := r.pool.Exec(ctx, query,
		entry.ID, entry.TenantID,
		entry.CheckOutTime, entry.DurationMinutes, entry.BreakMinutes,
		nilIfEmpty(entry.Notes), entry.Status,
	)
	if err != nil {
		return fmt.Errorf("time_entry_repo: update: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrTimeEntryNotFound
	}
	return nil
}

// GetActiveByMember returns the currently active (not checked-out) time entry for a member.
func (r *TimeEntryRepo) GetActiveByMember(ctx context.Context, crewMemberID uuid.UUID, tenantID uuid.UUID) (*domain.TimeEntry, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM time_entries WHERE crew_member_id = $1 AND tenant_id = $2 AND status = 'active' LIMIT 1`,
		timeEntryColumns,
	)

	e, err := scanTimeEntry(r.pool.QueryRow(ctx, query, crewMemberID, tenantID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // no active entry is not an error
		}
		return nil, fmt.Errorf("time_entry_repo: get_active_by_member: %w", err)
	}
	return e, nil
}

// ListByMember returns all time entries for a specific crew member within a date range.
func (r *TimeEntryRepo) ListByMember(ctx context.Context, crewMemberID uuid.UUID, tenantID uuid.UUID, from time.Time, to time.Time) ([]*domain.TimeEntry, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM time_entries
		 WHERE crew_member_id = $1 AND tenant_id = $2
		   AND check_in_time >= $3 AND check_in_time <= $4
		 ORDER BY check_in_time DESC`,
		timeEntryColumns,
	)

	return r.queryEntries(ctx, query, crewMemberID, tenantID, from, to)
}

// ListByProject returns all time entries for a specific project.
func (r *TimeEntryRepo) ListByProject(ctx context.Context, projectID uuid.UUID, tenantID uuid.UUID) ([]*domain.TimeEntry, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM time_entries
		 WHERE project_id = $1 AND tenant_id = $2
		 ORDER BY check_in_time DESC`,
		timeEntryColumns,
	)

	return r.queryEntries(ctx, query, projectID, tenantID)
}

// ListByTenant returns all time entries for a tenant within a date range.
func (r *TimeEntryRepo) ListByTenant(ctx context.Context, tenantID uuid.UUID, from time.Time, to time.Time) ([]*domain.TimeEntry, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM time_entries
		 WHERE tenant_id = $1
		   AND check_in_time >= $2 AND check_in_time <= $3
		 ORDER BY check_in_time DESC`,
		timeEntryColumns,
	)

	return r.queryEntries(ctx, query, tenantID, from, to)
}

// queryEntries is a helper that executes a query and scans rows into TimeEntry slices.
func (r *TimeEntryRepo) queryEntries(ctx context.Context, query string, args ...interface{}) ([]*domain.TimeEntry, error) {
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("time_entry_repo: query: %w", err)
	}
	defer rows.Close()

	var items []*domain.TimeEntry
	for rows.Next() {
		e := &domain.TimeEntry{}
		var notes *string
		scanErr := rows.Scan(
			&e.ID, &e.TenantID, &e.CrewMemberID, &e.AssignmentID, &e.ProjectID,
			&e.CheckInTime, &e.CheckOutTime, &e.DurationMinutes, &e.BreakMinutes,
			&notes, &e.Status, &e.CreatedAt,
		)
		if scanErr != nil {
			return nil, fmt.Errorf("time_entry_repo: scan: %w", scanErr)
		}
		if notes != nil {
			e.Notes = *notes
		}
		items = append(items, e)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("time_entry_repo: rows: %w", err)
	}
	return items, nil
}

// Ensure interface compliance at compile time.
var _ domain.TimeEntryRepository = (*TimeEntryRepo)(nil)

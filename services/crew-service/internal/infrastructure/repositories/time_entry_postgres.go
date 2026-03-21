package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/database"
	"github.com/jeckersberger/rentflow/services/crew-service/internal/domain"
)

type TimeEntryPostgres struct {
	db *database.PostgresPool
}

func NewTimeEntryPostgres(db *database.PostgresPool) *TimeEntryPostgres {
	return &TimeEntryPostgres{db: db}
}

func (r *TimeEntryPostgres) CreateTimeEntry(ctx context.Context, te *domain.TimeEntry) error {
	query := `
		INSERT INTO crew.time_entries (
			id, tenant_id, crew_member_id, project_id, date, start_time, end_time,
			break_minutes, total_hours, type, notes, status, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
		)
	`

	_, err := r.db.Exec(ctx, query,
		te.ID, te.TenantID, te.CrewMemberID, te.ProjectID, te.Date, te.StartTime, te.EndTime,
		te.BreakMinutes, te.TotalHours, string(te.Type), te.Notes, string(te.Status),
		te.CreatedAt, te.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create time entry: %w", err)
	}

	return nil
}

func (r *TimeEntryPostgres) UpdateTimeEntry(ctx context.Context, te *domain.TimeEntry) error {
	query := `
		UPDATE crew.time_entries SET
			break_minutes = $3, total_hours = $4, notes = $5, status = $6, updated_at = $7
		WHERE id = $1 AND tenant_id = $2
	`

	result, err := r.db.Exec(ctx, query,
		te.ID, te.TenantID, te.BreakMinutes, te.TotalHours, te.Notes, string(te.Status), te.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update time entry: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("time entry not found")
	}

	return nil
}

func (r *TimeEntryPostgres) GetTimeEntry(ctx context.Context, tenantID, timeEntryID string) (*domain.TimeEntry, error) {
	query := `
		SELECT id, tenant_id, crew_member_id, project_id, date, start_time, end_time,
		       break_minutes, total_hours, type, notes, status, created_at, updated_at
		FROM crew.time_entries
		WHERE id = $1 AND tenant_id = $2
	`

	var te domain.TimeEntry

	err := r.db.QueryRow(ctx, query, timeEntryID, tenantID).Scan(
		&te.ID, &te.TenantID, &te.CrewMemberID, &te.ProjectID, &te.Date, &te.StartTime, &te.EndTime,
		&te.BreakMinutes, &te.TotalHours, (*string)(&te.Type), &te.Notes, (*string)(&te.Status),
		&te.CreatedAt, &te.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("time entry not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get time entry: %w", err)
	}

	return &te, nil
}

func (r *TimeEntryPostgres) ListTimeEntries(ctx context.Context, tenantID string, limit, offset int) ([]*domain.TimeEntry, int64, error) {
	countQuery := `SELECT COUNT(*) FROM crew.time_entries WHERE tenant_id = $1`
	var total int64
	if err := r.db.QueryRow(ctx, countQuery, tenantID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count time entries: %w", err)
	}

	query := `
		SELECT id, tenant_id, crew_member_id, project_id, date, start_time, end_time,
		       break_minutes, total_hours, type, notes, status, created_at, updated_at
		FROM crew.time_entries
		WHERE tenant_id = $1
		ORDER BY date DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list time entries: %w", err)
	}
	defer rows.Close()

	var entries []*domain.TimeEntry
	for rows.Next() {
		var te domain.TimeEntry

		err := rows.Scan(
			&te.ID, &te.TenantID, &te.CrewMemberID, &te.ProjectID, &te.Date, &te.StartTime, &te.EndTime,
			&te.BreakMinutes, &te.TotalHours, (*string)(&te.Type), &te.Notes, (*string)(&te.Status),
			&te.CreatedAt, &te.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan time entry: %w", err)
		}

		entries = append(entries, &te)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows error: %w", err)
	}

	return entries, total, nil
}

func (r *TimeEntryPostgres) ListByCrewMemberAndPeriod(ctx context.Context, tenantID, crewMemberID string, from, to time.Time) ([]*domain.TimeEntry, int64, error) {
	query := `
		SELECT id, tenant_id, crew_member_id, project_id, date, start_time, end_time,
		       break_minutes, total_hours, type, notes, status, created_at, updated_at
		FROM crew.time_entries
		WHERE tenant_id = $1 AND crew_member_id = $2 AND date >= $3 AND date <= $4
		ORDER BY date DESC
	`

	rows, err := r.db.Query(ctx, query, tenantID, crewMemberID, from, to)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list time entries: %w", err)
	}
	defer rows.Close()

	var entries []*domain.TimeEntry
	for rows.Next() {
		var te domain.TimeEntry

		err := rows.Scan(
			&te.ID, &te.TenantID, &te.CrewMemberID, &te.ProjectID, &te.Date, &te.StartTime, &te.EndTime,
			&te.BreakMinutes, &te.TotalHours, (*string)(&te.Type), &te.Notes, (*string)(&te.Status),
			&te.CreatedAt, &te.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan time entry: %w", err)
		}

		entries = append(entries, &te)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows error: %w", err)
	}

	return entries, int64(len(entries)), nil
}

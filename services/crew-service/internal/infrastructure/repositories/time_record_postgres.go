package repositories

import (
	"context"
	"database/sql"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/crew-service/internal/domain"
)

// PostgresTimeRecordRepository implements the TimeRecordRepository interface
type PostgresTimeRecordRepository struct {
	db     *sql.DB
	logger logger.Logger
}

// NewPostgresTimeRecordRepository creates a new PostgreSQL time record repository
func NewPostgresTimeRecordRepository(db *sql.DB, log logger.Logger) *PostgresTimeRecordRepository {
	return &PostgresTimeRecordRepository{
		db:     db,
		logger: log,
	}
}

// FindByID retrieves a time record by ID
func (r *PostgresTimeRecordRepository) FindByID(ctx context.Context, id string) (*domain.TimeRecord, error) {
	query := `
		SELECT id, tenant_id, crew_member_id, assignment_id, date, start_time,
		       end_time, COALESCE(break_minutes, 0), COALESCE(overtime_minutes, 0), status, COALESCE(notes, ''), created_at, updated_at
		FROM time_records
		WHERE id = $1
	`

	var record domain.TimeRecord
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&record.ID, &record.TenantID, &record.CrewMemberID, &record.AssignmentID,
		&record.Date, &record.StartTime, &record.EndTime, &record.BreakMinutes,
		&record.OvertimeMinutes, &record.Status, &record.Notes,
		&record.CreatedAt, &record.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		r.logger.Error("failed to query time record", err)
		return nil, err
	}

	return &record, nil
}

// ListByCrewMember retrieves time records for a crew member
func (r *PostgresTimeRecordRepository) ListByCrewMember(ctx context.Context, crewMemberID string) ([]*domain.TimeRecord, error) {
	query := `
		SELECT id, tenant_id, crew_member_id, assignment_id, date, start_time,
		       end_time, COALESCE(break_minutes, 0), COALESCE(overtime_minutes, 0), status, COALESCE(notes, ''), created_at, updated_at
		FROM time_records
		WHERE crew_member_id = $1
		ORDER BY date DESC, start_time DESC
	`

	rows, err := r.db.QueryContext(ctx, query, crewMemberID)
	if err != nil {
		r.logger.Error("failed to query time records", err)
		return nil, err
	}
	defer rows.Close()

	var records []*domain.TimeRecord
	for rows.Next() {
		var record domain.TimeRecord
		if err := rows.Scan(
			&record.ID, &record.TenantID, &record.CrewMemberID, &record.AssignmentID,
			&record.Date, &record.StartTime, &record.EndTime, &record.BreakMinutes,
			&record.OvertimeMinutes, &record.Status, &record.Notes,
			&record.CreatedAt, &record.UpdatedAt,
		); err != nil {
			r.logger.Error("failed to scan time record", err)
			return nil, err
		}
		records = append(records, &record)
	}

	if err = rows.Err(); err != nil {
		r.logger.Error("error iterating time records", err)
		return nil, err
	}

	return records, nil
}

// ListByCrewMemberAndDate retrieves time records for a crew member on a specific date
func (r *PostgresTimeRecordRepository) ListByCrewMemberAndDate(ctx context.Context, crewMemberID string, date interface{}) ([]*domain.TimeRecord, error) {
	query := `
		SELECT id, tenant_id, crew_member_id, assignment_id, date, start_time,
		       end_time, COALESCE(break_minutes, 0), COALESCE(overtime_minutes, 0), status, COALESCE(notes, ''), created_at, updated_at
		FROM time_records
		WHERE crew_member_id = $1 AND date = $2
		ORDER BY start_time DESC
	`

	rows, err := r.db.QueryContext(ctx, query, crewMemberID, date)
	if err != nil {
		r.logger.Error("failed to query time records", err)
		return nil, err
	}
	defer rows.Close()

	var records []*domain.TimeRecord
	for rows.Next() {
		var record domain.TimeRecord
		if err := rows.Scan(
			&record.ID, &record.TenantID, &record.CrewMemberID, &record.AssignmentID,
			&record.Date, &record.StartTime, &record.EndTime, &record.BreakMinutes,
			&record.OvertimeMinutes, &record.Status, &record.Notes,
			&record.CreatedAt, &record.UpdatedAt,
		); err != nil {
			r.logger.Error("failed to scan time record", err)
			return nil, err
		}
		records = append(records, &record)
	}

	if err = rows.Err(); err != nil {
		r.logger.Error("error iterating time records", err)
		return nil, err
	}

	return records, nil
}

// ListByTenant retrieves all time records for a tenant
func (r *PostgresTimeRecordRepository) ListByTenant(ctx context.Context, tenantID string) ([]*domain.TimeRecord, error) {
	query := `
		SELECT id, tenant_id, crew_member_id, assignment_id, date, start_time,
		       end_time, COALESCE(break_minutes, 0), COALESCE(overtime_minutes, 0), status, COALESCE(notes, ''), created_at, updated_at
		FROM time_records
		WHERE tenant_id = $1
		ORDER BY date DESC, start_time DESC
		LIMIT 100
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		r.logger.Error("failed to query time records by tenant", err)
		return nil, err
	}
	defer rows.Close()

	var records []*domain.TimeRecord
	for rows.Next() {
		var record domain.TimeRecord
		if err := rows.Scan(
			&record.ID, &record.TenantID, &record.CrewMemberID, &record.AssignmentID,
			&record.Date, &record.StartTime, &record.EndTime, &record.BreakMinutes,
			&record.OvertimeMinutes, &record.Status, &record.Notes,
			&record.CreatedAt, &record.UpdatedAt,
		); err != nil {
			r.logger.Error("failed to scan time record", err)
			return nil, err
		}
		records = append(records, &record)
	}

	if err = rows.Err(); err != nil {
		r.logger.Error("error iterating time records", err)
		return nil, err
	}

	return records, nil
}

// Save persists a time record (creates or updates)
func (r *PostgresTimeRecordRepository) Save(ctx context.Context, record *domain.TimeRecord) error {
	query := `
		INSERT INTO time_records
		(id, tenant_id, crew_member_id, assignment_id, date, start_time,
		 end_time, break_minutes, overtime_minutes, status, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (id) DO UPDATE SET
		assignment_id = $4, end_time = $7, break_minutes = $8, overtime_minutes = $9,
		status = $10, notes = $11, updated_at = $13
	`

	_, err := r.db.ExecContext(ctx, query,
		record.ID, record.TenantID, record.CrewMemberID, record.AssignmentID,
		record.Date, record.StartTime, record.EndTime, record.BreakMinutes,
		record.OvertimeMinutes, string(record.Status), record.Notes,
		record.CreatedAt, record.UpdatedAt,
	)

	if err != nil {
		r.logger.Error("failed to save time record", err)
		return err
	}

	return nil
}

// Delete deletes a time record
func (r *PostgresTimeRecordRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM time_records WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Error("failed to delete time record", err)
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("failed to get rows affected", err)
		return err
	}

	if rows == 0 {
		return domain.ErrTimeRecordNotFound
	}

	return nil
}

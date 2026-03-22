package repositories

import (
	"context"
	"database/sql"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/crew-service/internal/domain"
)

// PostgresCrewAssignmentRepository implements the CrewAssignmentRepository interface
type PostgresCrewAssignmentRepository struct {
	db     *sql.DB
	logger logger.Logger
}

// NewPostgresCrewAssignmentRepository creates a new PostgreSQL assignment repository
func NewPostgresCrewAssignmentRepository(db *sql.DB, log logger.Logger) *PostgresCrewAssignmentRepository {
	return &PostgresCrewAssignmentRepository{
		db:     db,
		logger: log,
	}
}

// FindByID retrieves an assignment by ID
func (r *PostgresCrewAssignmentRepository) FindByID(ctx context.Context, id string) (*domain.CrewAssignment, error) {
	query := `
		SELECT id, tenant_id, crew_member_id, project_id, tour_id, role,
		       start_date, end_date, status, notes, created_at, updated_at
		FROM crew_assignments
		WHERE id = $1
	`

	var assignment domain.CrewAssignment
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&assignment.ID, &assignment.TenantID, &assignment.CrewMemberID,
		&assignment.ProjectID, &assignment.TourID, &assignment.Role,
		&assignment.StartDate, &assignment.EndDate, &assignment.Status,
		&assignment.Notes, &assignment.CreatedAt, &assignment.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		r.logger.Error("failed to query assignment", err)
		return nil, err
	}

	return &assignment, nil
}

// List retrieves assignments in a tenant with pagination
func (r *PostgresCrewAssignmentRepository) List(ctx context.Context, tenantID string, page, perPage int) ([]*domain.CrewAssignment, int, error) {
	offset := (page - 1) * perPage

	// Get total count
	countQuery := `SELECT COUNT(*) FROM crew_assignments WHERE tenant_id = $1`
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, tenantID).Scan(&total); err != nil {
		r.logger.Error("failed to count assignments", err)
		return nil, 0, err
	}

	// Get paginated results
	query := `
		SELECT id, tenant_id, crew_member_id, project_id, tour_id, role,
		       start_date, end_date, status, notes, created_at, updated_at
		FROM crew_assignments
		WHERE tenant_id = $1
		ORDER BY start_date DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID, perPage, offset)
	if err != nil {
		r.logger.Error("failed to query assignments", err)
		return nil, 0, err
	}
	defer rows.Close()

	var assignments []*domain.CrewAssignment
	for rows.Next() {
		var assignment domain.CrewAssignment
		if err := rows.Scan(
			&assignment.ID, &assignment.TenantID, &assignment.CrewMemberID,
			&assignment.ProjectID, &assignment.TourID, &assignment.Role,
			&assignment.StartDate, &assignment.EndDate, &assignment.Status,
			&assignment.Notes, &assignment.CreatedAt, &assignment.UpdatedAt,
		); err != nil {
			r.logger.Error("failed to scan assignment", err)
			return nil, 0, err
		}
		assignments = append(assignments, &assignment)
	}

	if err = rows.Err(); err != nil {
		r.logger.Error("error iterating assignments", err)
		return nil, 0, err
	}

	return assignments, total, nil
}

// ListByCrewMember retrieves assignments for a crew member
func (r *PostgresCrewAssignmentRepository) ListByCrewMember(ctx context.Context, crewMemberID string) ([]*domain.CrewAssignment, error) {
	query := `
		SELECT id, tenant_id, crew_member_id, project_id, tour_id, role,
		       start_date, end_date, status, notes, created_at, updated_at
		FROM crew_assignments
		WHERE crew_member_id = $1
		ORDER BY start_date DESC
	`

	rows, err := r.db.QueryContext(ctx, query, crewMemberID)
	if err != nil {
		r.logger.Error("failed to query assignments", err)
		return nil, err
	}
	defer rows.Close()

	var assignments []*domain.CrewAssignment
	for rows.Next() {
		var assignment domain.CrewAssignment
		if err := rows.Scan(
			&assignment.ID, &assignment.TenantID, &assignment.CrewMemberID,
			&assignment.ProjectID, &assignment.TourID, &assignment.Role,
			&assignment.StartDate, &assignment.EndDate, &assignment.Status,
			&assignment.Notes, &assignment.CreatedAt, &assignment.UpdatedAt,
		); err != nil {
			r.logger.Error("failed to scan assignment", err)
			return nil, err
		}
		assignments = append(assignments, &assignment)
	}

	if err = rows.Err(); err != nil {
		r.logger.Error("error iterating assignments", err)
		return nil, err
	}

	return assignments, nil
}

// ListByDateRange retrieves assignments within a date range
func (r *PostgresCrewAssignmentRepository) ListByDateRange(ctx context.Context, tenantID string, startDate, endDate interface{}) ([]*domain.CrewAssignment, error) {
	query := `
		SELECT id, tenant_id, crew_member_id, project_id, tour_id, role,
		       start_date, end_date, status, notes, created_at, updated_at
		FROM crew_assignments
		WHERE tenant_id = $1 AND start_date >= $2 AND end_date <= $3
		ORDER BY start_date DESC
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID, startDate, endDate)
	if err != nil {
		r.logger.Error("failed to query assignments", err)
		return nil, err
	}
	defer rows.Close()

	var assignments []*domain.CrewAssignment
	for rows.Next() {
		var assignment domain.CrewAssignment
		if err := rows.Scan(
			&assignment.ID, &assignment.TenantID, &assignment.CrewMemberID,
			&assignment.ProjectID, &assignment.TourID, &assignment.Role,
			&assignment.StartDate, &assignment.EndDate, &assignment.Status,
			&assignment.Notes, &assignment.CreatedAt, &assignment.UpdatedAt,
		); err != nil {
			r.logger.Error("failed to scan assignment", err)
			return nil, err
		}
		assignments = append(assignments, &assignment)
	}

	if err = rows.Err(); err != nil {
		r.logger.Error("error iterating assignments", err)
		return nil, err
	}

	return assignments, nil
}

// Save persists an assignment (creates or updates)
func (r *PostgresCrewAssignmentRepository) Save(ctx context.Context, assignment *domain.CrewAssignment) error {
	query := `
		INSERT INTO crew_assignments
		(id, tenant_id, crew_member_id, project_id, tour_id, role,
		 start_date, end_date, status, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (id) DO UPDATE SET
		crew_member_id = $3, project_id = $4, tour_id = $5, role = $6,
		start_date = $7, end_date = $8, status = $9, notes = $10, updated_at = $12
	`

	_, err := r.db.ExecContext(ctx, query,
		assignment.ID, assignment.TenantID, assignment.CrewMemberID,
		assignment.ProjectID, assignment.TourID, assignment.Role,
		assignment.StartDate, assignment.EndDate, string(assignment.Status),
		assignment.Notes, assignment.CreatedAt, assignment.UpdatedAt,
	)

	if err != nil {
		r.logger.Error("failed to save assignment", err)
		return err
	}

	return nil
}

// Delete deletes an assignment
func (r *PostgresCrewAssignmentRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM crew_assignments WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Error("failed to delete assignment", err)
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("failed to get rows affected", err)
		return err
	}

	if rows == 0 {
		return domain.ErrAssignmentNotFound
	}

	return nil
}

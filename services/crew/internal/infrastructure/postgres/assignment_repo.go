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

// assignmentColumns lists all columns of the crew_assignments table for consistent scanning.
const assignmentColumns = `
	id, crew_member_id, project_id, tenant_id, role,
	start_date, end_date, hours_planned, hours_actual,
	status, notes, created_at`

// AssignmentRepo implements domain.AssignmentRepository using PostgreSQL.
type AssignmentRepo struct {
	pool *pgxpool.Pool
}

// NewAssignmentRepo creates a new AssignmentRepo.
func NewAssignmentRepo(pool *pgxpool.Pool) *AssignmentRepo {
	return &AssignmentRepo{pool: pool}
}

// scanAssignment scans a single crew_assignments row into a domain.CrewAssignment.
func scanAssignment(row pgx.Row) (*domain.CrewAssignment, error) {
	a := &domain.CrewAssignment{}
	var (
		role      *string
		startDate *time.Time
		endDate   *time.Time
		notes     *string
	)

	err := row.Scan(
		&a.ID, &a.CrewMemberID, &a.ProjectID, &a.TenantID, &role,
		&startDate, &endDate, &a.HoursPlanned, &a.HoursActual,
		&a.Status, &notes, &a.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	if role != nil {
		a.Role = *role
	}
	if startDate != nil {
		a.StartDate = startDate.Format("2006-01-02")
	}
	if endDate != nil {
		a.EndDate = endDate.Format("2006-01-02")
	}
	if notes != nil {
		a.Notes = *notes
	}

	return a, nil
}

// Create inserts a new crew assignment record.
func (r *AssignmentRepo) Create(ctx context.Context, assignment *domain.CrewAssignment) error {
	query := `
		INSERT INTO crew_assignments (
			id, crew_member_id, project_id, tenant_id, role,
			start_date, end_date, hours_planned, hours_actual,
			status, notes
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING created_at`

	err := r.pool.QueryRow(ctx, query,
		assignment.ID, assignment.CrewMemberID, assignment.ProjectID, assignment.TenantID,
		nilIfEmpty(assignment.Role),
		nilIfEmpty(assignment.StartDate), nilIfEmpty(assignment.EndDate),
		assignment.HoursPlanned, assignment.HoursActual,
		assignment.Status, nilIfEmpty(assignment.Notes),
	).Scan(&assignment.CreatedAt)
	if err != nil {
		return fmt.Errorf("assignment_repo: create: %w", err)
	}
	return nil
}

// ListByMember returns all assignments for a specific crew member within a tenant.
func (r *AssignmentRepo) ListByMember(ctx context.Context, crewMemberID uuid.UUID, tenantID uuid.UUID) ([]*domain.CrewAssignment, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM crew_assignments WHERE crew_member_id = $1 AND tenant_id = $2 ORDER BY start_date DESC NULLS LAST`,
		assignmentColumns,
	)

	rows, err := r.pool.Query(ctx, query, crewMemberID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("assignment_repo: list_by_member query: %w", err)
	}
	defer rows.Close()

	var items []*domain.CrewAssignment
	for rows.Next() {
		a, scanErr := scanAssignment(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("assignment_repo: list_by_member scan: %w", scanErr)
		}
		items = append(items, a)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("assignment_repo: list_by_member rows: %w", err)
	}
	return items, nil
}

// ListAll returns all assignments for a tenant, optionally filtered by project_id.
func (r *AssignmentRepo) ListAll(ctx context.Context, tenantID uuid.UUID, projectID *uuid.UUID) ([]*domain.CrewAssignment, error) {
	var query string
	var args []interface{}

	if projectID != nil {
		query = fmt.Sprintf(
			`SELECT %s FROM crew_assignments WHERE tenant_id = $1 AND project_id = $2 ORDER BY start_date DESC NULLS LAST`,
			assignmentColumns,
		)
		args = []interface{}{tenantID, *projectID}
	} else {
		query = fmt.Sprintf(
			`SELECT %s FROM crew_assignments WHERE tenant_id = $1 ORDER BY start_date DESC NULLS LAST`,
			assignmentColumns,
		)
		args = []interface{}{tenantID}
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("assignment_repo: list_all query: %w", err)
	}
	defer rows.Close()

	var items []*domain.CrewAssignment
	for rows.Next() {
		a, scanErr := scanAssignment(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("assignment_repo: list_all scan: %w", scanErr)
		}
		items = append(items, a)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("assignment_repo: list_all rows: %w", err)
	}
	return items, nil
}

// Ensure interface compliance at compile time.
var _ domain.AssignmentRepository = (*AssignmentRepo)(nil)

package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jeckersberger/rentflow/pkg/common/database"
	"github.com/jeckersberger/rentflow/services/crew-service/internal/domain"
)

type AssignmentPostgres struct {
	db *database.PostgresPool
}

func NewAssignmentPostgres(db *database.PostgresPool) *AssignmentPostgres {
	return &AssignmentPostgres{db: db}
}

func (r *AssignmentPostgres) CreateAssignment(ctx context.Context, a *domain.Assignment) error {
	query := `
		INSERT INTO crew.assignments (
			id, tenant_id, crew_member_id, project_id, role, start_date, end_date, status, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		)
	`

	_, err := r.db.Exec(ctx, query,
		a.ID, a.TenantID, a.CrewMemberID, a.ProjectID, a.Role, a.StartDate, a.EndDate,
		string(a.Status), a.CreatedAt, a.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create assignment: %w", err)
	}

	return nil
}

func (r *AssignmentPostgres) GetAssignment(ctx context.Context, tenantID, assignmentID string) (*domain.Assignment, error) {
	query := `
		SELECT id, tenant_id, crew_member_id, project_id, role, start_date, end_date, status, created_at, updated_at
		FROM crew.assignments
		WHERE id = $1 AND tenant_id = $2
	`

	var a domain.Assignment

	err := r.db.QueryRow(ctx, query, assignmentID, tenantID).Scan(
		&a.ID, &a.TenantID, &a.CrewMemberID, &a.ProjectID, &a.Role, &a.StartDate, &a.EndDate,
		(*string)(&a.Status), &a.CreatedAt, &a.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("assignment not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get assignment: %w", err)
	}

	return &a, nil
}

func (r *AssignmentPostgres) ListAssignments(ctx context.Context, tenantID string, limit, offset int) ([]*domain.Assignment, int64, error) {
	countQuery := `SELECT COUNT(*) FROM crew.assignments WHERE tenant_id = $1`
	var total int64
	if err := r.db.QueryRow(ctx, countQuery, tenantID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count assignments: %w", err)
	}

	query := `
		SELECT id, tenant_id, crew_member_id, project_id, role, start_date, end_date, status, created_at, updated_at
		FROM crew.assignments
		WHERE tenant_id = $1
		ORDER BY start_date DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list assignments: %w", err)
	}
	defer rows.Close()

	var assignments []*domain.Assignment
	for rows.Next() {
		var a domain.Assignment

		err := rows.Scan(
			&a.ID, &a.TenantID, &a.CrewMemberID, &a.ProjectID, &a.Role, &a.StartDate, &a.EndDate,
			(*string)(&a.Status), &a.CreatedAt, &a.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan assignment: %w", err)
		}

		assignments = append(assignments, &a)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows error: %w", err)
	}

	return assignments, total, nil
}

func (r *AssignmentPostgres) GetByProject(ctx context.Context, tenantID, projectID string) ([]*domain.Assignment, error) {
	query := `
		SELECT id, tenant_id, crew_member_id, project_id, role, start_date, end_date, status, created_at, updated_at
		FROM crew.assignments
		WHERE tenant_id = $1 AND project_id = $2 AND status = 'active'
		ORDER BY start_date DESC
	`

	rows, err := r.db.Query(ctx, query, tenantID, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get assignments: %w", err)
	}
	defer rows.Close()

	var assignments []*domain.Assignment
	for rows.Next() {
		var a domain.Assignment

		err := rows.Scan(
			&a.ID, &a.TenantID, &a.CrewMemberID, &a.ProjectID, &a.Role, &a.StartDate, &a.EndDate,
			(*string)(&a.Status), &a.CreatedAt, &a.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan assignment: %w", err)
		}

		assignments = append(assignments, &a)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return assignments, nil
}

func (r *AssignmentPostgres) DeleteAssignment(ctx context.Context, tenantID, assignmentID string) error {
	query := `DELETE FROM crew.assignments WHERE id = $1 AND tenant_id = $2`

	result, err := r.db.Exec(ctx, query, assignmentID, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete assignment: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("assignment not found")
	}

	return nil
}

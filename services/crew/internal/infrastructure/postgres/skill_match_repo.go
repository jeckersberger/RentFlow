package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jeckersberger/EquipFlow/services/crew/internal/domain"
)

// SkillMatchRepo implements domain.SkillMatchRepository using PostgreSQL.
type SkillMatchRepo struct {
	pool *pgxpool.Pool
}

// NewSkillMatchRepo creates a new SkillMatchRepo.
func NewSkillMatchRepo(pool *pgxpool.Pool) *SkillMatchRepo {
	return &SkillMatchRepo{pool: pool}
}

// FindBySkill returns active crew members whose skills contain the given skill.
// An exact match (array element equals the skill, case-insensitive) gets
// ExactMatch=true; an ILIKE-based partial match gets ExactMatch=false.
func (r *SkillMatchRepo) FindBySkill(
	ctx context.Context,
	tenantID uuid.UUID,
	skill string,
) ([]domain.SkillMatchRow, error) {
	// Use a CTE that unnests skills for case-insensitive comparison.
	// exact_match: at least one element matches exactly (case-insensitive).
	// partial_match: at least one element contains the search term.
	query := `
		SELECT
			cm.id,
			cm.first_name,
			cm.last_name,
			EXISTS (
				SELECT 1 FROM unnest(cm.skills) AS s
				WHERE LOWER(s) = LOWER($2)
			) AS exact_match
		FROM crew_members cm
		WHERE cm.tenant_id = $1
		  AND cm.is_active = TRUE
		  AND (
		      EXISTS (SELECT 1 FROM unnest(cm.skills) AS s WHERE LOWER(s) = LOWER($2))
		      OR
		      EXISTS (SELECT 1 FROM unnest(cm.skills) AS s WHERE LOWER(s) LIKE '%' || LOWER($2) || '%')
		  )
		ORDER BY cm.last_name, cm.first_name`

	rows, err := r.pool.Query(ctx, query, tenantID, skill)
	if err != nil {
		return nil, fmt.Errorf("skill_match_repo: find_by_skill query: %w", err)
	}
	defer rows.Close()

	var results []domain.SkillMatchRow
	for rows.Next() {
		var row domain.SkillMatchRow
		if err := rows.Scan(&row.CrewMemberID, &row.FirstName, &row.LastName, &row.ExactMatch); err != nil {
			return nil, fmt.Errorf("skill_match_repo: find_by_skill scan: %w", err)
		}
		results = append(results, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("skill_match_repo: find_by_skill rows: %w", err)
	}

	return results, nil
}

// BlockedMemberIDs returns a set of crew member IDs that have an overlapping
// availability block (vacation, sick, etc.) within the given date range.
func (r *SkillMatchRepo) BlockedMemberIDs(
	ctx context.Context,
	tenantID uuid.UUID,
	from, to string,
) (map[uuid.UUID]bool, error) {
	query := `
		SELECT DISTINCT crew_member_id
		FROM crew_availability_blocks
		WHERE tenant_id = $1
		  AND start_date <= $3::DATE
		  AND end_date >= $2::DATE`

	rows, err := r.pool.Query(ctx, query, tenantID, from, to)
	if err != nil {
		return nil, fmt.Errorf("skill_match_repo: blocked_member_ids query: %w", err)
	}
	defer rows.Close()

	result := make(map[uuid.UUID]bool)
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("skill_match_repo: blocked_member_ids scan: %w", err)
		}
		result[id] = true
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("skill_match_repo: blocked_member_ids rows: %w", err)
	}

	return result, nil
}

// AssignedMemberIDs returns a set of crew member IDs that have a non-cancelled
// assignment overlapping the given date range. If excludeProjectID is non-nil,
// assignments for that project are ignored (the caller is planning for that project).
func (r *SkillMatchRepo) AssignedMemberIDs(
	ctx context.Context,
	tenantID uuid.UUID,
	from, to string,
	excludeProjectID *uuid.UUID,
) (map[uuid.UUID]bool, error) {
	var query string
	var args []interface{}

	if excludeProjectID != nil {
		query = `
			SELECT DISTINCT crew_member_id
			FROM crew_assignments
			WHERE tenant_id = $1
			  AND status NOT IN ('cancelled', 'completed')
			  AND start_date <= $3::DATE
			  AND end_date >= $2::DATE
			  AND project_id != $4`
		args = []interface{}{tenantID, from, to, *excludeProjectID}
	} else {
		query = `
			SELECT DISTINCT crew_member_id
			FROM crew_assignments
			WHERE tenant_id = $1
			  AND status NOT IN ('cancelled', 'completed')
			  AND start_date <= $3::DATE
			  AND end_date >= $2::DATE`
		args = []interface{}{tenantID, from, to}
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("skill_match_repo: assigned_member_ids query: %w", err)
	}
	defer rows.Close()

	result := make(map[uuid.UUID]bool)
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("skill_match_repo: assigned_member_ids scan: %w", err)
		}
		result[id] = true
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("skill_match_repo: assigned_member_ids rows: %w", err)
	}

	return result, nil
}

// Ensure interface compliance at compile time.
var _ domain.SkillMatchRepository = (*SkillMatchRepo)(nil)

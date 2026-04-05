package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jeckersberger/EquipFlow/services/crew/internal/domain"

	apperrors "github.com/jeckersberger/EquipFlow/pkg/common/errors"
)

// crewMemberColumns lists all columns of the crew_members table for consistent scanning.
const crewMemberColumns = `
	id, tenant_id, first_name, last_name, email, phone,
	role, skills, hourly_rate, is_active, notes, created_at, updated_at`

// CrewMemberRepo implements domain.CrewMemberRepository using PostgreSQL.
type CrewMemberRepo struct {
	pool *pgxpool.Pool
}

// NewCrewMemberRepo creates a new CrewMemberRepo.
func NewCrewMemberRepo(pool *pgxpool.Pool) *CrewMemberRepo {
	return &CrewMemberRepo{pool: pool}
}

// scanCrewMember scans a single crew_members row into a domain.CrewMember.
func scanCrewMember(row pgx.Row) (*domain.CrewMember, error) {
	m := &domain.CrewMember{}
	var (
		email  *string
		phone  *string
		notes  *string
		skills []string
	)

	err := row.Scan(
		&m.ID, &m.TenantID, &m.FirstName, &m.LastName, &email, &phone,
		&m.Role, &skills, &m.HourlyRate, &m.IsActive, &notes, &m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if email != nil {
		m.Email = *email
	}
	if phone != nil {
		m.Phone = *phone
	}
	if notes != nil {
		m.Notes = *notes
	}
	if skills != nil {
		m.Skills = skills
	} else {
		m.Skills = []string{}
	}

	return m, nil
}

// Create inserts a new crew member record.
func (r *CrewMemberRepo) Create(ctx context.Context, member *domain.CrewMember) error {
	query := `
		INSERT INTO crew_members (
			id, tenant_id, first_name, last_name, email, phone,
			role, skills, hourly_rate, is_active, notes
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING created_at, updated_at`

	err := r.pool.QueryRow(ctx, query,
		member.ID, member.TenantID, member.FirstName, member.LastName,
		nilIfEmpty(member.Email), nilIfEmpty(member.Phone),
		member.Role, member.Skills, member.HourlyRate, member.IsActive, nilIfEmpty(member.Notes),
	).Scan(&member.CreatedAt, &member.UpdatedAt)
	if err != nil {
		return fmt.Errorf("crew_member_repo: create: %w", err)
	}
	return nil
}

// GetByID retrieves a single crew member by primary key scoped to a tenant.
func (r *CrewMemberRepo) GetByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*domain.CrewMember, error) {
	query := fmt.Sprintf(`SELECT %s FROM crew_members WHERE id = $1 AND tenant_id = $2`, crewMemberColumns)
	m, err := scanCrewMember(r.pool.QueryRow(ctx, query, id, tenantID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("crew_member_repo: get_by_id: %w", err)
	}
	return m, nil
}

// List returns all crew members for a tenant, ordered by last_name.
func (r *CrewMemberRepo) List(ctx context.Context, tenantID uuid.UUID) ([]*domain.CrewMember, error) {
	query := fmt.Sprintf(`SELECT %s FROM crew_members WHERE tenant_id = $1 ORDER BY last_name, first_name`, crewMemberColumns)

	rows, err := r.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("crew_member_repo: list query: %w", err)
	}
	defer rows.Close()

	var items []*domain.CrewMember
	for rows.Next() {
		m, scanErr := scanCrewMember(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("crew_member_repo: list scan: %w", scanErr)
		}
		items = append(items, m)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("crew_member_repo: list rows: %w", err)
	}
	return items, nil
}

// Update modifies an existing crew member record.
func (r *CrewMemberRepo) Update(ctx context.Context, member *domain.CrewMember) error {
	query := `
		UPDATE crew_members SET
			first_name = $3, last_name = $4, email = $5, phone = $6,
			role = $7, skills = $8, hourly_rate = $9, is_active = $10, notes = $11,
			updated_at = NOW()
		WHERE id = $1 AND tenant_id = $2
		RETURNING updated_at`

	err := r.pool.QueryRow(ctx, query,
		member.ID, member.TenantID,
		member.FirstName, member.LastName,
		nilIfEmpty(member.Email), nilIfEmpty(member.Phone),
		member.Role, member.Skills, member.HourlyRate, member.IsActive, nilIfEmpty(member.Notes),
	).Scan(&member.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apperrors.ErrNotFound
		}
		return fmt.Errorf("crew_member_repo: update: %w", err)
	}
	return nil
}

// Delete removes a crew member by ID within a tenant scope.
func (r *CrewMemberRepo) Delete(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error {
	tag, err := r.pool.Exec(ctx,
		`DELETE FROM crew_members WHERE id = $1 AND tenant_id = $2`,
		id, tenantID,
	)
	if err != nil {
		return fmt.Errorf("crew_member_repo: delete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

// nilIfEmpty returns nil if the string is empty, otherwise a pointer to it.
func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// Ensure interface compliance at compile time.
var _ domain.CrewMemberRepository = (*CrewMemberRepo)(nil)

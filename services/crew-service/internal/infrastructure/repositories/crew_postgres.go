package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jeckersberger/rentflow/pkg/common/database"
	"github.com/jeckersberger/rentflow/services/crew-service/internal/domain"
	"github.com/lib/pq"
)

type CrewPostgres struct {
	db *database.PostgresPool
}

func NewCrewPostgres(db *database.PostgresPool) *CrewPostgres {
	return &CrewPostgres{db: db}
}

func (r *CrewPostgres) CreateCrewMember(ctx context.Context, cm *domain.CrewMember) error {
	query := `
		INSERT INTO crew.crew_members (
			id, tenant_id, user_id, first_name, last_name, email, phone, type, 
			skills, hourly_rate, daily_rate, status, availability_calendar, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
		)
	`

	skills := pq.StringArray(cm.Skills)

	_, err := r.db.Exec(ctx, query,
		cm.ID, cm.TenantID, cm.UserID, cm.FirstName, cm.LastName, cm.Email, cm.Phone,
		string(cm.Type), skills, cm.HourlyRate, cm.DailyRate, string(cm.Status),
		cm.AvailabilityCalendar, cm.CreatedAt, cm.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create crew member: %w", err)
	}

	return nil
}

func (r *CrewPostgres) UpdateCrewMember(ctx context.Context, cm *domain.CrewMember) error {
	query := `
		UPDATE crew.crew_members SET
			first_name = $3, last_name = $4, email = $5, phone = $6, type = $7,
			skills = $8, hourly_rate = $9, daily_rate = $10, status = $11,
			availability_calendar = $12, updated_at = $13
		WHERE id = $1 AND tenant_id = $2
	`

	skills := pq.StringArray(cm.Skills)

	result, err := r.db.Exec(ctx, query,
		cm.ID, cm.TenantID, cm.FirstName, cm.LastName, cm.Email, cm.Phone,
		string(cm.Type), skills, cm.HourlyRate, cm.DailyRate, string(cm.Status),
		cm.AvailabilityCalendar, cm.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update crew member: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("crew member not found")
	}

	return nil
}

func (r *CrewPostgres) GetCrewMember(ctx context.Context, tenantID, crewMemberID string) (*domain.CrewMember, error) {
	query := `
		SELECT id, tenant_id, user_id, first_name, last_name, email, phone, type,
		       skills, hourly_rate, daily_rate, status, availability_calendar, created_at, updated_at
		FROM crew.crew_members
		WHERE id = $1 AND tenant_id = $2
	`

	var cm domain.CrewMember
	var skills pq.StringArray

	err := r.db.QueryRow(ctx, query, crewMemberID, tenantID).Scan(
		&cm.ID, &cm.TenantID, &cm.UserID, &cm.FirstName, &cm.LastName, &cm.Email, &cm.Phone,
		(*string)(&cm.Type), &skills, &cm.HourlyRate, &cm.DailyRate, (*string)(&cm.Status),
		&cm.AvailabilityCalendar, &cm.CreatedAt, &cm.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("crew member not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get crew member: %w", err)
	}

	cm.Skills = fromStringArray(skills)

	return &cm, nil
}

func (r *CrewPostgres) ListCrewMembers(ctx context.Context, tenantID string, limit, offset int) ([]*domain.CrewMember, int64, error) {
	countQuery := `SELECT COUNT(*) FROM crew.crew_members WHERE tenant_id = $1`
	var total int64
	if err := r.db.QueryRow(ctx, countQuery, tenantID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count crew members: %w", err)
	}

	query := `
		SELECT id, tenant_id, user_id, first_name, last_name, email, phone, type,
		       skills, hourly_rate, daily_rate, status, availability_calendar, created_at, updated_at
		FROM crew.crew_members
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list crew members: %w", err)
	}
	defer rows.Close()

	var members []*domain.CrewMember
	for rows.Next() {
		var cm domain.CrewMember
		var skills pq.StringArray

		err := rows.Scan(
			&cm.ID, &cm.TenantID, &cm.UserID, &cm.FirstName, &cm.LastName, &cm.Email, &cm.Phone,
			(*string)(&cm.Type), &skills, &cm.HourlyRate, &cm.DailyRate, (*string)(&cm.Status),
			&cm.AvailabilityCalendar, &cm.CreatedAt, &cm.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan crew member: %w", err)
		}

		cm.Skills = fromStringArray(skills)
		members = append(members, &cm)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows error: %w", err)
	}

	return members, total, nil
}

func (r *CrewPostgres) DeleteCrewMember(ctx context.Context, tenantID, crewMemberID string) error {
	query := `DELETE FROM crew.crew_members WHERE id = $1 AND tenant_id = $2`

	result, err := r.db.Exec(ctx, query, crewMemberID, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete crew member: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("crew member not found")
	}

	return nil
}

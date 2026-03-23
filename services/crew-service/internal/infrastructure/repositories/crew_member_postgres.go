package repositories

import (
	"context"
	"database/sql"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/crew-service/internal/domain"
)

// PostgresCrewMemberRepository implements the CrewMemberRepository interface using PostgreSQL
type PostgresCrewMemberRepository struct {
	db     *sql.DB
	logger logger.Logger
}

// NewPostgresCrewMemberRepository creates a new PostgreSQL crew member repository
func NewPostgresCrewMemberRepository(db *sql.DB, log logger.Logger) *PostgresCrewMemberRepository {
	return &PostgresCrewMemberRepository{
		db:     db,
		logger: log,
	}
}

// FindByID retrieves a crew member by ID
func (r *PostgresCrewMemberRepository) FindByID(ctx context.Context, id string) (*domain.CrewMember, error) {
	query := `
		SELECT id, tenant_id, first_name, last_name, email, COALESCE(phone, ''), role, status,
		       hourly_rate, daily_rate, preferred_vehicle_id, COALESCE(emergency_contact, ''), COALESCE(notes, ''),
		       created_at, updated_at
		FROM crew_members
		WHERE id = $1
	`

	var member domain.CrewMember
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&member.ID, &member.TenantID, &member.FirstName, &member.LastName,
		&member.Email, &member.Phone, &member.Role, &member.Status,
		&member.HourlyRate, &member.DailyRate, &member.PreferredVehicleID,
		&member.EmergencyContact, &member.Notes,
		&member.CreatedAt, &member.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		r.logger.Error("failed to query crew member", err)
		return nil, err
	}

	return &member, nil
}

// FindByEmail retrieves a crew member by email within a tenant
func (r *PostgresCrewMemberRepository) FindByEmail(ctx context.Context, tenantID, email string) (*domain.CrewMember, error) {
	query := `
		SELECT id, tenant_id, first_name, last_name, email, COALESCE(phone, ''), role, status,
		       hourly_rate, daily_rate, preferred_vehicle_id, COALESCE(emergency_contact, ''), COALESCE(notes, ''),
		       created_at, updated_at
		FROM crew_members
		WHERE tenant_id = $1 AND email = $2
	`

	var member domain.CrewMember
	err := r.db.QueryRowContext(ctx, query, tenantID, email).Scan(
		&member.ID, &member.TenantID, &member.FirstName, &member.LastName,
		&member.Email, &member.Phone, &member.Role, &member.Status,
		&member.HourlyRate, &member.DailyRate, &member.PreferredVehicleID,
		&member.EmergencyContact, &member.Notes,
		&member.CreatedAt, &member.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		r.logger.Error("failed to query crew member by email", err)
		return nil, err
	}

	return &member, nil
}

// List retrieves crew members in a tenant with pagination
func (r *PostgresCrewMemberRepository) List(ctx context.Context, tenantID string, page, perPage int) ([]*domain.CrewMember, int, error) {
	offset := (page - 1) * perPage

	// Get total count
	countQuery := `SELECT COUNT(*) FROM crew_members WHERE tenant_id = $1`
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, tenantID).Scan(&total); err != nil {
		r.logger.Error("failed to count crew members", err)
		return nil, 0, err
	}

	// Get paginated results
	query := `
		SELECT id, tenant_id, first_name, last_name, email, COALESCE(phone, ''), role, status,
		       hourly_rate, daily_rate, preferred_vehicle_id, COALESCE(emergency_contact, ''), COALESCE(notes, ''),
		       created_at, updated_at
		FROM crew_members
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID, perPage, offset)
	if err != nil {
		r.logger.Error("failed to query crew members", err)
		return nil, 0, err
	}
	defer rows.Close()

	var members []*domain.CrewMember
	for rows.Next() {
		var member domain.CrewMember
		if err := rows.Scan(
			&member.ID, &member.TenantID, &member.FirstName, &member.LastName,
			&member.Email, &member.Phone, &member.Role, &member.Status,
			&member.HourlyRate, &member.DailyRate, &member.PreferredVehicleID,
			&member.EmergencyContact, &member.Notes,
			&member.CreatedAt, &member.UpdatedAt,
		); err != nil {
			r.logger.Error("failed to scan crew member", err)
			return nil, 0, err
		}
		members = append(members, &member)
	}

	if err = rows.Err(); err != nil {
		r.logger.Error("error iterating crew members", err)
		return nil, 0, err
	}

	return members, total, nil
}

// Save persists a crew member (creates or updates)
func (r *PostgresCrewMemberRepository) Save(ctx context.Context, member *domain.CrewMember) error {
	query := `
		INSERT INTO crew_members
		(id, tenant_id, first_name, last_name, email, phone, role, status,
		 hourly_rate, daily_rate, preferred_vehicle_id, emergency_contact, notes,
		 created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		ON CONFLICT (id) DO UPDATE SET
		first_name = $3, last_name = $4, email = $5, phone = $6, role = $7,
		status = $8, hourly_rate = $9, daily_rate = $10, preferred_vehicle_id = $11,
		emergency_contact = $12, notes = $13, updated_at = $15
	`

	_, err := r.db.ExecContext(ctx, query,
		member.ID, member.TenantID, member.FirstName, member.LastName,
		member.Email, member.Phone, string(member.Role), string(member.Status),
		member.HourlyRate, member.DailyRate, member.PreferredVehicleID,
		member.EmergencyContact, member.Notes,
		member.CreatedAt, member.UpdatedAt,
	)

	if err != nil {
		r.logger.Error("failed to save crew member", err)
		return err
	}

	return nil
}

// Delete deletes a crew member
func (r *PostgresCrewMemberRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM crew_members WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Error("failed to delete crew member", err)
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("failed to get rows affected", err)
		return err
	}

	if rows == 0 {
		return domain.ErrCrewMemberNotFound
	}

	return nil
}

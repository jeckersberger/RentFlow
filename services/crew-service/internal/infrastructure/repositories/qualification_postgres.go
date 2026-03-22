package repositories

import (
	"context"
	"database/sql"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/crew-service/internal/domain"
)

// PostgresQualificationRepository implements the QualificationRepository interface
type PostgresQualificationRepository struct {
	db     *sql.DB
	logger logger.Logger
}

// NewPostgresQualificationRepository creates a new PostgreSQL qualification repository
func NewPostgresQualificationRepository(db *sql.DB, log logger.Logger) *PostgresQualificationRepository {
	return &PostgresQualificationRepository{
		db:     db,
		logger: log,
	}
}

// FindByID retrieves a qualification by ID
func (r *PostgresQualificationRepository) FindByID(ctx context.Context, id string) (*domain.Qualification, error) {
	query := `
		SELECT id, tenant_id, crew_member_id, qualification_type, issued_at, expires_at,
		       certificate_number, issuing_authority, status, created_at, updated_at
		FROM qualifications
		WHERE id = $1
	`

	var qual domain.Qualification
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&qual.ID, &qual.TenantID, &qual.CrewMemberID, &qual.QualificationType,
		&qual.IssuedAt, &qual.ExpiresAt, &qual.CertificateNumber, &qual.IssuingAuthority,
		&qual.Status, &qual.CreatedAt, &qual.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		r.logger.Error("failed to query qualification", err)
		return nil, err
	}

	return &qual, nil
}

// ListByCrewMember retrieves qualifications for a crew member
func (r *PostgresQualificationRepository) ListByCrewMember(ctx context.Context, crewMemberID string) ([]*domain.Qualification, error) {
	query := `
		SELECT id, tenant_id, crew_member_id, qualification_type, issued_at, expires_at,
		       certificate_number, issuing_authority, status, created_at, updated_at
		FROM qualifications
		WHERE crew_member_id = $1
		ORDER BY issued_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, crewMemberID)
	if err != nil {
		r.logger.Error("failed to query qualifications", err)
		return nil, err
	}
	defer rows.Close()

	var quals []*domain.Qualification
	for rows.Next() {
		var qual domain.Qualification
		if err := rows.Scan(
			&qual.ID, &qual.TenantID, &qual.CrewMemberID, &qual.QualificationType,
			&qual.IssuedAt, &qual.ExpiresAt, &qual.CertificateNumber, &qual.IssuingAuthority,
			&qual.Status, &qual.CreatedAt, &qual.UpdatedAt,
		); err != nil {
			r.logger.Error("failed to scan qualification", err)
			return nil, err
		}
		quals = append(quals, &qual)
	}

	if err = rows.Err(); err != nil {
		r.logger.Error("error iterating qualifications", err)
		return nil, err
	}

	return quals, nil
}

// Save persists a qualification (creates or updates)
func (r *PostgresQualificationRepository) Save(ctx context.Context, qual *domain.Qualification) error {
	query := `
		INSERT INTO qualifications
		(id, tenant_id, crew_member_id, qualification_type, issued_at, expires_at,
		 certificate_number, issuing_authority, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (id) DO UPDATE SET
		qualification_type = $4, issued_at = $5, expires_at = $6,
		certificate_number = $7, issuing_authority = $8, status = $9, updated_at = $11
	`

	_, err := r.db.ExecContext(ctx, query,
		qual.ID, qual.TenantID, qual.CrewMemberID, string(qual.QualificationType),
		qual.IssuedAt, qual.ExpiresAt, qual.CertificateNumber, qual.IssuingAuthority,
		string(qual.Status), qual.CreatedAt, qual.UpdatedAt,
	)

	if err != nil {
		r.logger.Error("failed to save qualification", err)
		return err
	}

	return nil
}

// Delete deletes a qualification
func (r *PostgresQualificationRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM qualifications WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Error("failed to delete qualification", err)
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("failed to get rows affected", err)
		return err
	}

	if rows == 0 {
		return domain.ErrQualificationNotFound
	}

	return nil
}

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

// qualificationColumns lists all columns of the crew_qualifications table for consistent scanning.
const qualificationColumns = `
	id, crew_member_id, tenant_id, name, issued_at, expires_at,
	certificate_number, notes, created_at`

// QualificationRepo implements domain.QualificationRepository using PostgreSQL.
type QualificationRepo struct {
	pool *pgxpool.Pool
}

// NewQualificationRepo creates a new QualificationRepo.
func NewQualificationRepo(pool *pgxpool.Pool) *QualificationRepo {
	return &QualificationRepo{pool: pool}
}

// scanQualification scans a single crew_qualifications row into a domain.CrewQualification.
func scanQualification(row pgx.Row) (*domain.CrewQualification, error) {
	q := &domain.CrewQualification{}
	var (
		issuedAt          *time.Time
		expiresAt         *time.Time
		certificateNumber *string
		notes             *string
	)

	err := row.Scan(
		&q.ID, &q.CrewMemberID, &q.TenantID, &q.Name, &issuedAt, &expiresAt,
		&certificateNumber, &notes, &q.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	if issuedAt != nil {
		q.IssuedAt = issuedAt.Format("2006-01-02")
	}
	if expiresAt != nil {
		q.ExpiresAt = expiresAt.Format("2006-01-02")
	}
	if certificateNumber != nil {
		q.CertificateNumber = *certificateNumber
	}
	if notes != nil {
		q.Notes = *notes
	}

	return q, nil
}

// Create inserts a new crew qualification record.
func (r *QualificationRepo) Create(ctx context.Context, qualification *domain.CrewQualification) error {
	query := `
		INSERT INTO crew_qualifications (
			id, crew_member_id, tenant_id, name, issued_at, expires_at,
			certificate_number, notes
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING created_at`

	err := r.pool.QueryRow(ctx, query,
		qualification.ID, qualification.CrewMemberID, qualification.TenantID,
		qualification.Name,
		nilIfEmpty(qualification.IssuedAt), nilIfEmpty(qualification.ExpiresAt),
		nilIfEmpty(qualification.CertificateNumber), nilIfEmpty(qualification.Notes),
	).Scan(&qualification.CreatedAt)
	if err != nil {
		return fmt.Errorf("qualification_repo: create: %w", err)
	}
	return nil
}

// ListByMember returns all qualifications for a specific crew member within a tenant.
func (r *QualificationRepo) ListByMember(ctx context.Context, crewMemberID uuid.UUID, tenantID uuid.UUID) ([]*domain.CrewQualification, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM crew_qualifications WHERE crew_member_id = $1 AND tenant_id = $2 ORDER BY created_at DESC`,
		qualificationColumns,
	)

	rows, err := r.pool.Query(ctx, query, crewMemberID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("qualification_repo: list_by_member query: %w", err)
	}
	defer rows.Close()

	var items []*domain.CrewQualification
	for rows.Next() {
		q, scanErr := scanQualification(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("qualification_repo: list_by_member scan: %w", scanErr)
		}
		items = append(items, q)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("qualification_repo: list_by_member rows: %w", err)
	}
	return items, nil
}

// Ensure interface compliance at compile time.
var _ domain.QualificationRepository = (*QualificationRepo)(nil)

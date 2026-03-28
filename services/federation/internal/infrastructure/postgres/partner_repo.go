package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jeckersberger/EquipFlow/services/federation/internal/domain"

	apperrors "github.com/jeckersberger/EquipFlow/pkg/common/errors"
)

// partnerColumns lists all columns of the federation_partners table for consistent scanning.
const partnerColumns = `
	id, tenant_id, partner_name, partner_url, api_key_hash,
	status, notes, created_at, updated_at`

// PartnerRepo implements domain.FederationPartnerRepository using PostgreSQL.
type PartnerRepo struct {
	pool *pgxpool.Pool
}

// NewPartnerRepo creates a new PartnerRepo.
func NewPartnerRepo(pool *pgxpool.Pool) *PartnerRepo {
	return &PartnerRepo{pool: pool}
}

// scanPartner scans a single federation_partners row into a domain.FederationPartner.
func scanPartner(row pgx.Row) (*domain.FederationPartner, error) {
	p := &domain.FederationPartner{}
	var (
		partnerURL *string
		apiKeyHash *string
		status     *string
		notes      *string
	)

	err := row.Scan(
		&p.ID, &p.TenantID, &p.PartnerName, &partnerURL, &apiKeyHash,
		&status, &notes, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if partnerURL != nil {
		p.PartnerURL = *partnerURL
	}
	if apiKeyHash != nil {
		p.APIKeyHash = *apiKeyHash
	}
	if status != nil {
		p.Status = *status
	}
	if notes != nil {
		p.Notes = *notes
	}

	return p, nil
}

// Create inserts a new federation partner record.
func (r *PartnerRepo) Create(ctx context.Context, partner *domain.FederationPartner) error {
	query := `
		INSERT INTO federation_partners (
			id, tenant_id, partner_name, partner_url, api_key_hash,
			status, notes
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at`

	err := r.pool.QueryRow(ctx, query,
		partner.ID, partner.TenantID, partner.PartnerName,
		nilIfEmpty(partner.PartnerURL), nilIfEmpty(partner.APIKeyHash),
		partner.Status, nilIfEmpty(partner.Notes),
	).Scan(&partner.ID, &partner.CreatedAt, &partner.UpdatedAt)
	if err != nil {
		return fmt.Errorf("partner_repo: create: %w", err)
	}
	return nil
}

// GetByID retrieves a federation partner by its primary key within a tenant scope.
func (r *PartnerRepo) GetByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*domain.FederationPartner, error) {
	query := fmt.Sprintf(`SELECT %s FROM federation_partners WHERE id = $1 AND tenant_id = $2`, partnerColumns)
	p, err := scanPartner(r.pool.QueryRow(ctx, query, id, tenantID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("partner_repo: get_by_id: %w", err)
	}
	return p, nil
}

// List returns a filtered, paginated list of federation partners for a tenant plus total count.
func (r *PartnerRepo) List(ctx context.Context, tenantID uuid.UUID, filter domain.PartnerFilter) ([]*domain.FederationPartner, int64, error) {
	conditions := []string{"tenant_id = $1"}
	args := []interface{}{tenantID}
	argIdx := 2

	if filter.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, *filter.Status)
		argIdx++
	}

	where := strings.Join(conditions, " AND ")

	// Count query.
	var total int64
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM federation_partners WHERE %s`, where)
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("partner_repo: list count: %w", err)
	}

	// Data query with pagination.
	offset := (filter.Page - 1) * filter.PerPage
	dataQuery := fmt.Sprintf(
		`SELECT %s FROM federation_partners WHERE %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		partnerColumns, where, argIdx, argIdx+1,
	)
	args = append(args, filter.PerPage, offset)

	rows, err := r.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("partner_repo: list query: %w", err)
	}
	defer rows.Close()

	var items []*domain.FederationPartner
	for rows.Next() {
		p, scanErr := scanPartner(rows)
		if scanErr != nil {
			return nil, 0, fmt.Errorf("partner_repo: list scan: %w", scanErr)
		}
		items = append(items, p)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("partner_repo: list rows: %w", err)
	}
	return items, total, nil
}

// Update updates an existing federation partner record.
func (r *PartnerRepo) Update(ctx context.Context, partner *domain.FederationPartner) error {
	query := `
		UPDATE federation_partners SET
			partner_name = $3,
			partner_url = $4,
			api_key_hash = $5,
			status = $6,
			notes = $7,
			updated_at = NOW()
		WHERE id = $1 AND tenant_id = $2
		RETURNING updated_at`

	err := r.pool.QueryRow(ctx, query,
		partner.ID, partner.TenantID, partner.PartnerName,
		nilIfEmpty(partner.PartnerURL), nilIfEmpty(partner.APIKeyHash),
		partner.Status, nilIfEmpty(partner.Notes),
	).Scan(&partner.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apperrors.ErrNotFound
		}
		return fmt.Errorf("partner_repo: update: %w", err)
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
var _ domain.FederationPartnerRepository = (*PartnerRepo)(nil)

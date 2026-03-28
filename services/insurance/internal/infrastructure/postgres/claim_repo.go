package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	apperrors "github.com/jeckersberger/EquipFlow/pkg/common/errors"
	"github.com/jeckersberger/EquipFlow/services/insurance/internal/domain"
)

const claimColumns = `
	id, policy_id, tenant_id, equipment_id, claim_number, description,
	damage_amount, claim_amount, status, incident_date, filed_at,
	resolved_at, notes, created_by, created_at`

type ClaimRepo struct {
	pool *pgxpool.Pool
}

func NewClaimRepo(pool *pgxpool.Pool) *ClaimRepo {
	return &ClaimRepo{pool: pool}
}

func scanClaim(row pgx.Row) (*domain.InsuranceClaim, error) {
	c := &domain.InsuranceClaim{}
	var (
		equipmentID  *uuid.UUID
		claimNumber  *string
		incidentDate *string
		notes        *string
		createdBy    *uuid.UUID
	)

	err := row.Scan(
		&c.ID, &c.PolicyID, &c.TenantID, &equipmentID, &claimNumber, &c.Description,
		&c.DamageAmount, &c.ClaimAmount, &c.Status, &incidentDate, &c.FiledAt,
		&c.ResolvedAt, &notes, &createdBy, &c.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	c.EquipmentID = equipmentID
	c.IncidentDate = incidentDate
	if claimNumber != nil {
		c.ClaimNumber = *claimNumber
	}
	if notes != nil {
		c.Notes = *notes
	}
	c.CreatedBy = createdBy

	return c, nil
}

func (r *ClaimRepo) Create(ctx context.Context, claim *domain.InsuranceClaim) error {
	query := `
		INSERT INTO insurance_claims (
			id, policy_id, tenant_id, equipment_id, claim_number, description,
			damage_amount, claim_amount, status, incident_date, notes, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING filed_at, created_at`

	err := r.pool.QueryRow(ctx, query,
		claim.ID, claim.PolicyID, claim.TenantID, claim.EquipmentID,
		nilIfEmpty(claim.ClaimNumber), claim.Description,
		claim.DamageAmount, claim.ClaimAmount, claim.Status,
		nilIfEmpty(ptrToStr(claim.IncidentDate)),
		nilIfEmpty(claim.Notes), claim.CreatedBy,
	).Scan(&claim.FiledAt, &claim.CreatedAt)
	if err != nil {
		return fmt.Errorf("claim_repo: create: %w", err)
	}
	return nil
}

func (r *ClaimRepo) GetByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*domain.InsuranceClaim, error) {
	query := fmt.Sprintf(`SELECT %s FROM insurance_claims WHERE id = $1 AND tenant_id = $2`, claimColumns)
	claim, err := scanClaim(r.pool.QueryRow(ctx, query, id, tenantID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("claim_repo: get_by_id: %w", err)
	}
	return claim, nil
}

func (r *ClaimRepo) List(ctx context.Context, tenantID uuid.UUID, filter domain.ClaimFilter) ([]*domain.InsuranceClaim, int64, error) {
	conditions := []string{"tenant_id = $1"}
	args := []interface{}{tenantID}
	argIdx := 2

	if filter.PolicyID != nil {
		conditions = append(conditions, fmt.Sprintf("policy_id = $%d", argIdx))
		args = append(args, *filter.PolicyID)
		argIdx++
	}
	if filter.Status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, filter.Status)
		argIdx++
	}

	where := strings.Join(conditions, " AND ")

	var total int64
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM insurance_claims WHERE %s`, where)
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("claim_repo: list count: %w", err)
	}

	page := filter.Page
	if page < 1 {
		page = 1
	}
	perPage := filter.PerPage
	if perPage < 1 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	dataQuery := fmt.Sprintf(
		`SELECT %s FROM insurance_claims WHERE %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		claimColumns, where, argIdx, argIdx+1,
	)
	args = append(args, perPage, offset)

	rows, err := r.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("claim_repo: list query: %w", err)
	}
	defer rows.Close()

	var items []*domain.InsuranceClaim
	for rows.Next() {
		claim, scanErr := scanClaim(rows)
		if scanErr != nil {
			return nil, 0, fmt.Errorf("claim_repo: list scan: %w", scanErr)
		}
		items = append(items, claim)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("claim_repo: list rows: %w", err)
	}
	return items, total, nil
}

func (r *ClaimRepo) Update(ctx context.Context, claim *domain.InsuranceClaim) error {
	query := `
		UPDATE insurance_claims SET
			claim_number = $3, description = $4, damage_amount = $5,
			claim_amount = $6, status = $7, incident_date = $8,
			notes = $9, resolved_at = $10
		WHERE id = $1 AND tenant_id = $2
		RETURNING created_at`

	err := r.pool.QueryRow(ctx, query,
		claim.ID, claim.TenantID,
		nilIfEmpty(claim.ClaimNumber), claim.Description,
		claim.DamageAmount, claim.ClaimAmount, claim.Status,
		nilIfEmpty(ptrToStr(claim.IncidentDate)),
		nilIfEmpty(claim.Notes), claim.ResolvedAt,
	).Scan(&claim.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apperrors.ErrNotFound
		}
		return fmt.Errorf("claim_repo: update: %w", err)
	}
	return nil
}

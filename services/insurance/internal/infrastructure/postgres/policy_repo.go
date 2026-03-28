package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	apperrors "github.com/jeckersberger/EquipFlow/pkg/common/errors"
	"github.com/jeckersberger/EquipFlow/services/insurance/internal/domain"
)

const policyColumns = `
	id, tenant_id, name, provider, policy_number, coverage_type,
	coverage_amount, deductible, premium, start_date, end_date,
	is_active, notes, created_at, updated_at`

type PolicyRepo struct {
	pool *pgxpool.Pool
}

func NewPolicyRepo(pool *pgxpool.Pool) *PolicyRepo {
	return &PolicyRepo{pool: pool}
}

func scanPolicy(row pgx.Row) (*domain.InsurancePolicy, error) {
	p := &domain.InsurancePolicy{}
	var (
		provider     *string
		policyNumber *string
		startDate    *string
		endDate      *string
		notes        *string
	)

	err := row.Scan(
		&p.ID, &p.TenantID, &p.Name, &provider, &policyNumber, &p.CoverageType,
		&p.CoverageAmount, &p.Deductible, &p.Premium, &startDate, &endDate,
		&p.IsActive, &notes, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if provider != nil {
		p.Provider = *provider
	}
	if policyNumber != nil {
		p.PolicyNumber = *policyNumber
	}
	p.StartDate = startDate
	p.EndDate = endDate
	if notes != nil {
		p.Notes = *notes
	}

	return p, nil
}

func (r *PolicyRepo) Create(ctx context.Context, policy *domain.InsurancePolicy) error {
	query := `
		INSERT INTO insurance_policies (
			id, tenant_id, name, provider, policy_number, coverage_type,
			coverage_amount, deductible, premium, start_date, end_date,
			is_active, notes
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING created_at, updated_at`

	err := r.pool.QueryRow(ctx, query,
		policy.ID, policy.TenantID, policy.Name,
		nilIfEmpty(policy.Provider), nilIfEmpty(policy.PolicyNumber),
		policy.CoverageType, policy.CoverageAmount, policy.Deductible, policy.Premium,
		nilIfEmpty(ptrToStr(policy.StartDate)), nilIfEmpty(ptrToStr(policy.EndDate)),
		policy.IsActive, nilIfEmpty(policy.Notes),
	).Scan(&policy.CreatedAt, &policy.UpdatedAt)
	if err != nil {
		return fmt.Errorf("policy_repo: create: %w", err)
	}
	return nil
}

func (r *PolicyRepo) GetByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*domain.InsurancePolicy, error) {
	query := fmt.Sprintf(`SELECT %s FROM insurance_policies WHERE id = $1 AND tenant_id = $2`, policyColumns)
	policy, err := scanPolicy(r.pool.QueryRow(ctx, query, id, tenantID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("policy_repo: get_by_id: %w", err)
	}
	return policy, nil
}

func (r *PolicyRepo) List(ctx context.Context, tenantID uuid.UUID, filter domain.PolicyFilter) ([]*domain.InsurancePolicy, int64, error) {
	var total int64
	countQuery := `SELECT COUNT(*) FROM insurance_policies WHERE tenant_id = $1`
	err := r.pool.QueryRow(ctx, countQuery, tenantID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("policy_repo: list count: %w", err)
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
		`SELECT %s FROM insurance_policies WHERE tenant_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		policyColumns,
	)

	rows, err := r.pool.Query(ctx, dataQuery, tenantID, perPage, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("policy_repo: list query: %w", err)
	}
	defer rows.Close()

	var items []*domain.InsurancePolicy
	for rows.Next() {
		p, scanErr := scanPolicy(rows)
		if scanErr != nil {
			return nil, 0, fmt.Errorf("policy_repo: list scan: %w", scanErr)
		}
		items = append(items, p)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("policy_repo: list rows: %w", err)
	}
	return items, total, nil
}

func (r *PolicyRepo) Update(ctx context.Context, policy *domain.InsurancePolicy) error {
	query := `
		UPDATE insurance_policies SET
			name = $3, provider = $4, policy_number = $5, coverage_type = $6,
			coverage_amount = $7, deductible = $8, premium = $9,
			start_date = $10, end_date = $11, is_active = $12, notes = $13,
			updated_at = NOW()
		WHERE id = $1 AND tenant_id = $2
		RETURNING updated_at`

	err := r.pool.QueryRow(ctx, query,
		policy.ID, policy.TenantID,
		policy.Name, nilIfEmpty(policy.Provider), nilIfEmpty(policy.PolicyNumber),
		policy.CoverageType, policy.CoverageAmount, policy.Deductible, policy.Premium,
		nilIfEmpty(ptrToStr(policy.StartDate)), nilIfEmpty(ptrToStr(policy.EndDate)),
		policy.IsActive, nilIfEmpty(policy.Notes),
	).Scan(&policy.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apperrors.ErrNotFound
		}
		return fmt.Errorf("policy_repo: update: %w", err)
	}
	return nil
}

func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func ptrToStr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

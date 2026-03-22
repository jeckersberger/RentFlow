package database

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jeckersberger/rentflow/services/insurance-service/internal/domain"
)

// PolicyRepository implements the ports.PolicyRepository interface
type PolicyRepository struct {
	db *sql.DB
}

// NewPolicyRepository creates a new policy repository
func NewPolicyRepository(db *sql.DB) *PolicyRepository {
	return &PolicyRepository{db: db}
}

// Create inserts a new policy
func (pr *PolicyRepository) Create(ctx context.Context, policy *domain.Policy) error {
	query := `
		INSERT INTO policies 
		(id, tenant_id, policy_number, policy_type, provider, coverage_amount, deductible, 
		 premium_annual, premium_monthly, start_date, end_date, status, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`

	_, err := pr.db.ExecContext(ctx, query,
		policy.ID, policy.TenantID, policy.PolicyNumber, string(policy.PolicyType),
		policy.Provider, policy.CoverageAmount, policy.Deductible, policy.PremiumAnnual,
		policy.PremiumMonthly, policy.StartDate, policy.EndDate, string(policy.Status),
		policy.Notes, policy.CreatedAt, policy.UpdatedAt,
	)

	return err
}

// GetByID retrieves a policy by ID
func (pr *PolicyRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Policy, error) {
	query := `
		SELECT id, tenant_id, policy_number, policy_type, provider, coverage_amount, deductible,
		       premium_annual, premium_monthly, start_date, end_date, status, notes, created_at, updated_at
		FROM policies WHERE id = $1
	`

	var policy domain.Policy
	err := pr.db.QueryRowContext(ctx, query, id).Scan(
		&policy.ID, &policy.TenantID, &policy.PolicyNumber, (*string)(&policy.PolicyType),
		&policy.Provider, &policy.CoverageAmount, &policy.Deductible, &policy.PremiumAnnual,
		&policy.PremiumMonthly, &policy.StartDate, &policy.EndDate, (*string)(&policy.Status),
		&policy.Notes, &policy.CreatedAt, &policy.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &domain.ErrPolicyNotFound{ID: id.String()}
		}
		return nil, err
	}

	return &policy, nil
}

// GetByTenantID retrieves all policies for a tenant
func (pr *PolicyRepository) GetByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*domain.Policy, error) {
	query := `
		SELECT id, tenant_id, policy_number, policy_type, provider, coverage_amount, deductible,
		       premium_annual, premium_monthly, start_date, end_date, status, notes, created_at, updated_at
		FROM policies WHERE tenant_id = $1 ORDER BY created_at DESC
	`

	rows, err := pr.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var policies []*domain.Policy
	for rows.Next() {
		var policy domain.Policy
		if err := rows.Scan(
			&policy.ID, &policy.TenantID, &policy.PolicyNumber, (*string)(&policy.PolicyType),
			&policy.Provider, &policy.CoverageAmount, &policy.Deductible, &policy.PremiumAnnual,
			&policy.PremiumMonthly, &policy.StartDate, &policy.EndDate, (*string)(&policy.Status),
			&policy.Notes, &policy.CreatedAt, &policy.UpdatedAt,
		); err != nil {
			return nil, err
		}
		policies = append(policies, &policy)
	}

	return policies, rows.Err()
}

// GetActiveByTenantID retrieves all active policies for a tenant
func (pr *PolicyRepository) GetActiveByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*domain.Policy, error) {
	query := `
		SELECT id, tenant_id, policy_number, policy_type, provider, coverage_amount, deductible,
		       premium_annual, premium_monthly, start_date, end_date, status, notes, created_at, updated_at
		FROM policies 
		WHERE tenant_id = $1 AND status = 'active' AND end_date > CURRENT_DATE
		ORDER BY created_at DESC
	`

	rows, err := pr.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var policies []*domain.Policy
	for rows.Next() {
		var policy domain.Policy
		if err := rows.Scan(
			&policy.ID, &policy.TenantID, &policy.PolicyNumber, (*string)(&policy.PolicyType),
			&policy.Provider, &policy.CoverageAmount, &policy.Deductible, &policy.PremiumAnnual,
			&policy.PremiumMonthly, &policy.StartDate, &policy.EndDate, (*string)(&policy.Status),
			&policy.Notes, &policy.CreatedAt, &policy.UpdatedAt,
		); err != nil {
			return nil, err
		}
		policies = append(policies, &policy)
	}

	return policies, rows.Err()
}

// Update updates an existing policy
func (pr *PolicyRepository) Update(ctx context.Context, policy *domain.Policy) error {
	query := `
		UPDATE policies 
		SET provider = $1, coverage_amount = $2, deductible = $3, premium_annual = $4,
		    premium_monthly = $5, end_date = $6, status = $7, notes = $8, updated_at = $9
		WHERE id = $10
	`

	_, err := pr.db.ExecContext(ctx, query,
		policy.Provider, policy.CoverageAmount, policy.Deductible, policy.PremiumAnnual,
		policy.PremiumMonthly, policy.EndDate, string(policy.Status), policy.Notes, policy.UpdatedAt,
		policy.ID,
	)

	return err
}

// Delete deletes a policy
func (pr *PolicyRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := pr.db.ExecContext(ctx, "DELETE FROM policies WHERE id = $1", id)
	return err
}

// GetExpiringPolicies retrieves policies expiring within the specified days
func (pr *PolicyRepository) GetExpiringPolicies(ctx context.Context, daysUntilExpiry int) ([]*domain.Policy, error) {
	query := `
		SELECT id, tenant_id, policy_number, policy_type, provider, coverage_amount, deductible,
		       premium_annual, premium_monthly, start_date, end_date, status, notes, created_at, updated_at
		FROM policies 
		WHERE status = 'active' AND end_date <= CURRENT_DATE + INTERVAL '1 day' * $1
		ORDER BY end_date ASC
	`

	rows, err := pr.db.QueryContext(ctx, query, daysUntilExpiry)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var policies []*domain.Policy
	for rows.Next() {
		var policy domain.Policy
		if err := rows.Scan(
			&policy.ID, &policy.TenantID, &policy.PolicyNumber, (*string)(&policy.PolicyType),
			&policy.Provider, &policy.CoverageAmount, &policy.Deductible, &policy.PremiumAnnual,
			&policy.PremiumMonthly, &policy.StartDate, &policy.EndDate, (*string)(&policy.Status),
			&policy.Notes, &policy.CreatedAt, &policy.UpdatedAt,
		); err != nil {
			return nil, err
		}
		policies = append(policies, &policy)
	}

	return policies, rows.Err()
}

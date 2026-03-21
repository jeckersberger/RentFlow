package repositories

import (
	"context"
	"database/sql"

	"github.com/jeckersberger/rentflow/pkg/common/database"
	"github.com/jeckersberger/rentflow/services/insurance-service/internal/domain"
)

type PolicyPostgres struct {
	db *database.PostgresPool
}

func NewPolicyPostgres(db *database.PostgresPool) *PolicyPostgres {
	return &PolicyPostgres{db: db}
}

func (r *PolicyPostgres) Create(ctx context.Context, policy *domain.Policy) error {
	query := `
		INSERT INTO policies
		(id, tenant_id, policy_number, provider, type, coverage_amount, deductible, premium, start_date, end_date, status, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`
	_, err := r.db.Exec(ctx, query,
		policy.ID, policy.TenantID, policy.PolicyNumber, policy.Provider, policy.Type,
		policy.CoverageAmount, policy.Deductible, policy.Premium, policy.StartDate, policy.EndDate,
		policy.Status, policy.Notes, policy.CreatedAt, policy.UpdatedAt,
	)
	return err
}

func (r *PolicyPostgres) GetByID(ctx context.Context, tenantID, id string) (*domain.Policy, error) {
	query := `
		SELECT id, tenant_id, policy_number, provider, type, coverage_amount, deductible, premium, start_date, end_date, status, notes, created_at, updated_at
		FROM policies
		WHERE id = $1 AND tenant_id = $2
	`
	var p domain.Policy
	err := r.db.QueryRow(ctx, query, id, tenantID).Scan(
		&p.ID, &p.TenantID, &p.PolicyNumber, &p.Provider, &p.Type,
		&p.CoverageAmount, &p.Deductible, &p.Premium, &p.StartDate, &p.EndDate,
		&p.Status, &p.Notes, &p.CreatedAt, &p.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PolicyPostgres) ListByTenant(ctx context.Context, tenantID string) ([]*domain.Policy, error) {
	query := `
		SELECT id, tenant_id, policy_number, provider, type, coverage_amount, deductible, premium, start_date, end_date, status, notes, created_at, updated_at
		FROM policies
		WHERE tenant_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var policies []*domain.Policy
	for rows.Next() {
		var p domain.Policy
		err := rows.Scan(
			&p.ID, &p.TenantID, &p.PolicyNumber, &p.Provider, &p.Type,
			&p.CoverageAmount, &p.Deductible, &p.Premium, &p.StartDate, &p.EndDate,
			&p.Status, &p.Notes, &p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		policies = append(policies, &p)
	}
	return policies, rows.Err()
}

func (r *PolicyPostgres) ListActive(ctx context.Context, tenantID string) ([]*domain.Policy, error) {
	query := `
		SELECT id, tenant_id, policy_number, provider, type, coverage_amount, deductible, premium, start_date, end_date, status, notes, created_at, updated_at
		FROM policies
		WHERE tenant_id = $1 AND status = $2 AND end_date > NOW()
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, tenantID, domain.PolicyStatusActive)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var policies []*domain.Policy
	for rows.Next() {
		var p domain.Policy
		err := rows.Scan(
			&p.ID, &p.TenantID, &p.PolicyNumber, &p.Provider, &p.Type,
			&p.CoverageAmount, &p.Deductible, &p.Premium, &p.StartDate, &p.EndDate,
			&p.Status, &p.Notes, &p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		policies = append(policies, &p)
	}
	return policies, rows.Err()
}

func (r *PolicyPostgres) ListExpiring(ctx context.Context, tenantID string, days int) ([]*domain.Policy, error) {
	query := `
		SELECT id, tenant_id, policy_number, provider, type, coverage_amount, deductible, premium, start_date, end_date, status, notes, created_at, updated_at
		FROM policies
		WHERE tenant_id = $1 AND status = $2 AND end_date <= NOW() + INTERVAL '1 day' * $3 AND end_date > NOW()
		ORDER BY end_date ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID, domain.PolicyStatusActive, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var policies []*domain.Policy
	for rows.Next() {
		var p domain.Policy
		err := rows.Scan(
			&p.ID, &p.TenantID, &p.PolicyNumber, &p.Provider, &p.Type,
			&p.CoverageAmount, &p.Deductible, &p.Premium, &p.StartDate, &p.EndDate,
			&p.Status, &p.Notes, &p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		policies = append(policies, &p)
	}
	return policies, rows.Err()
}

func (r *PolicyPostgres) Update(ctx context.Context, policy *domain.Policy) error {
	query := `
		UPDATE policies
		SET provider = $1, coverage_amount = $2, deductible = $3, premium = $4, status = $5, notes = $6, updated_at = $7
		WHERE id = $8 AND tenant_id = $9
	`
	_, err := r.db.Exec(ctx, query,
		policy.Provider, policy.CoverageAmount, policy.Deductible, policy.Premium,
		policy.Status, policy.Notes, policy.UpdatedAt, policy.ID, policy.TenantID,
	)
	return err
}

func (r *PolicyPostgres) Delete(ctx context.Context, tenantID, id string) error {
	query := `DELETE FROM policies WHERE id = $1 AND tenant_id = $2`
	_, err := r.db.Exec(ctx, query, id, tenantID)
	return err
}

type ClaimPostgres struct {
	db *database.PostgresPool
}

func NewClaimPostgres(db *database.PostgresPool) *ClaimPostgres {
	return &ClaimPostgres{db: db}
}

func (r *ClaimPostgres) Create(ctx context.Context, claim *domain.Claim) error {
	query := `
		INSERT INTO claims
		(id, tenant_id, policy_id, equipment_id, incident_date, description, damage_amount, claim_amount, status, documents, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	docsJSON := "{}"
	if len(claim.Documents) > 0 {
		// Simple JSON array representation
		docsJSON = "[]"
	}

	_, err := r.db.Exec(ctx, query,
		claim.ID, claim.TenantID, claim.PolicyID, claim.EquipmentID, claim.IncidentDate,
		claim.Description, claim.DamageAmount, claim.ClaimAmount, claim.Status, docsJSON,
		claim.CreatedAt, claim.UpdatedAt,
	)
	return err
}

func (r *ClaimPostgres) GetByID(ctx context.Context, tenantID, id string) (*domain.Claim, error) {
	query := `
		SELECT id, tenant_id, policy_id, equipment_id, incident_date, description, damage_amount, claim_amount, status, documents, created_at, updated_at
		FROM claims
		WHERE id = $1 AND tenant_id = $2
	`
	var c domain.Claim
	var docsJSON string

	err := r.db.QueryRow(ctx, query, id, tenantID).Scan(
		&c.ID, &c.TenantID, &c.PolicyID, &c.EquipmentID, &c.IncidentDate,
		&c.Description, &c.DamageAmount, &c.ClaimAmount, &c.Status, &docsJSON,
		&c.CreatedAt, &c.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Parse documents JSON
	c.Documents = make([]string, 0)
	return &c, nil
}

func (r *ClaimPostgres) ListByTenant(ctx context.Context, tenantID string) ([]*domain.Claim, error) {
	query := `
		SELECT id, tenant_id, policy_id, equipment_id, incident_date, description, damage_amount, claim_amount, status, documents, created_at, updated_at
		FROM claims
		WHERE tenant_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var claims []*domain.Claim
	for rows.Next() {
		var c domain.Claim
		var docsJSON string

		err := rows.Scan(
			&c.ID, &c.TenantID, &c.PolicyID, &c.EquipmentID, &c.IncidentDate,
			&c.Description, &c.DamageAmount, &c.ClaimAmount, &c.Status, &docsJSON,
			&c.CreatedAt, &c.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		c.Documents = make([]string, 0)
		claims = append(claims, &c)
	}
	return claims, rows.Err()
}

func (r *ClaimPostgres) ListByEquipment(ctx context.Context, tenantID, equipmentID string) ([]*domain.Claim, error) {
	query := `
		SELECT id, tenant_id, policy_id, equipment_id, incident_date, description, damage_amount, claim_amount, status, documents, created_at, updated_at
		FROM claims
		WHERE tenant_id = $1 AND equipment_id = $2
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, tenantID, equipmentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var claims []*domain.Claim
	for rows.Next() {
		var c domain.Claim
		var docsJSON string

		err := rows.Scan(
			&c.ID, &c.TenantID, &c.PolicyID, &c.EquipmentID, &c.IncidentDate,
			&c.Description, &c.DamageAmount, &c.ClaimAmount, &c.Status, &docsJSON,
			&c.CreatedAt, &c.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		c.Documents = make([]string, 0)
		claims = append(claims, &c)
	}
	return claims, rows.Err()
}

func (r *ClaimPostgres) ListByPolicy(ctx context.Context, tenantID, policyID string) ([]*domain.Claim, error) {
	query := `
		SELECT id, tenant_id, policy_id, equipment_id, incident_date, description, damage_amount, claim_amount, status, documents, created_at, updated_at
		FROM claims
		WHERE tenant_id = $1 AND policy_id = $2
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, tenantID, policyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var claims []*domain.Claim
	for rows.Next() {
		var c domain.Claim
		var docsJSON string

		err := rows.Scan(
			&c.ID, &c.TenantID, &c.PolicyID, &c.EquipmentID, &c.IncidentDate,
			&c.Description, &c.DamageAmount, &c.ClaimAmount, &c.Status, &docsJSON,
			&c.CreatedAt, &c.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		c.Documents = make([]string, 0)
		claims = append(claims, &c)
	}
	return claims, rows.Err()
}

func (r *ClaimPostgres) Update(ctx context.Context, claim *domain.Claim) error {
	query := `
		UPDATE claims
		SET description = $1, damage_amount = $2, claim_amount = $3, status = $4, documents = $5, updated_at = $6
		WHERE id = $7 AND tenant_id = $8
	`
	docsJSON := "[]"
	_, err := r.db.Exec(ctx, query,
		claim.Description, claim.DamageAmount, claim.ClaimAmount, claim.Status, docsJSON,
		claim.UpdatedAt, claim.ID, claim.TenantID,
	)
	return err
}

func (r *ClaimPostgres) Delete(ctx context.Context, tenantID, id string) error {
	query := `DELETE FROM claims WHERE id = $1 AND tenant_id = $2`
	_, err := r.db.Exec(ctx, query, id, tenantID)
	return err
}

type RiskAssessmentPostgres struct {
	db *database.PostgresPool
}

func NewRiskAssessmentPostgres(db *database.PostgresPool) *RiskAssessmentPostgres {
	return &RiskAssessmentPostgres{db: db}
}

func (r *RiskAssessmentPostgres) Create(ctx context.Context, assessment *domain.RiskAssessment) error {
	query := `
		INSERT INTO risk_assessments
		(id, tenant_id, equipment_id, risk_score, factors, assessed_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.Exec(ctx, query,
		assessment.ID, assessment.TenantID, assessment.EquipmentID, assessment.RiskScore,
		assessment.Factors, assessment.AssessedAt,
	)
	return err
}

func (r *RiskAssessmentPostgres) GetByEquipment(ctx context.Context, tenantID, equipmentID string) (*domain.RiskAssessment, error) {
	query := `
		SELECT id, tenant_id, equipment_id, risk_score, factors, assessed_at
		FROM risk_assessments
		WHERE tenant_id = $1 AND equipment_id = $2
		ORDER BY assessed_at DESC
		LIMIT 1
	`
	var ra domain.RiskAssessment
	err := r.db.QueryRow(ctx, query, tenantID, equipmentID).Scan(
		&ra.ID, &ra.TenantID, &ra.EquipmentID, &ra.RiskScore, &ra.Factors, &ra.AssessedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &ra, nil
}

func (r *RiskAssessmentPostgres) ListByTenant(ctx context.Context, tenantID string) ([]*domain.RiskAssessment, error) {
	query := `
		SELECT id, tenant_id, equipment_id, risk_score, factors, assessed_at
		FROM risk_assessments
		WHERE tenant_id = $1
		ORDER BY assessed_at DESC
	`
	rows, err := r.db.Query(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var assessments []*domain.RiskAssessment
	for rows.Next() {
		var ra domain.RiskAssessment
		err := rows.Scan(
			&ra.ID, &ra.TenantID, &ra.EquipmentID, &ra.RiskScore, &ra.Factors, &ra.AssessedAt,
		)
		if err != nil {
			return nil, err
		}
		assessments = append(assessments, &ra)
	}
	return assessments, rows.Err()
}

func (r *RiskAssessmentPostgres) Update(ctx context.Context, assessment *domain.RiskAssessment) error {
	query := `
		UPDATE risk_assessments
		SET risk_score = $1, factors = $2, assessed_at = $3
		WHERE id = $4 AND tenant_id = $5
	`
	_, err := r.db.Exec(ctx, query,
		assessment.RiskScore, assessment.Factors, assessment.AssessedAt, assessment.ID, assessment.TenantID,
	)
	return err
}

func (r *RiskAssessmentPostgres) Delete(ctx context.Context, tenantID, equipmentID string) error {
	query := `DELETE FROM risk_assessments WHERE tenant_id = $1 AND equipment_id = $2`
	_, err := r.db.Exec(ctx, query, tenantID, equipmentID)
	return err
}

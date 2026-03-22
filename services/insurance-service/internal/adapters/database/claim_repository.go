package database

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jeckersberger/rentflow/services/insurance-service/internal/domain"
)

// ClaimRepository implements the ports.ClaimRepository interface
type ClaimRepository struct {
	db *sql.DB
}

// NewClaimRepository creates a new claim repository
func NewClaimRepository(db *sql.DB) *ClaimRepository {
	return &ClaimRepository{db: db}
}

// Create inserts a new claim
func (cr *ClaimRepository) Create(ctx context.Context, claim *domain.Claim) error {
	query := `
		INSERT INTO claims 
		(id, tenant_id, policy_id, claim_number, equipment_id, project_id, incident_date,
		 reported_date, description, damage_type, status, claimed_amount, approved_amount,
		 settled_amount, adjuster_notes, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
	`

	_, err := cr.db.ExecContext(ctx, query,
		claim.ID, claim.TenantID, claim.PolicyID, claim.ClaimNumber, claim.EquipmentID,
		claim.ProjectID, claim.IncidentDate, claim.ReportedDate, claim.Description,
		string(claim.DamageType), string(claim.Status), claim.ClaimedAmount, claim.ApprovedAmount,
		claim.SettledAmount, claim.AdjusterNotes, claim.CreatedBy, claim.CreatedAt, claim.UpdatedAt,
	)

	return err
}

// GetByID retrieves a claim by ID
func (cr *ClaimRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Claim, error) {
	query := `
		SELECT id, tenant_id, policy_id, claim_number, equipment_id, project_id, incident_date,
		       reported_date, description, damage_type, status, claimed_amount, approved_amount,
		       settled_amount, adjuster_notes, created_by, created_at, updated_at
		FROM claims WHERE id = $1
	`

	var claim domain.Claim
	err := cr.db.QueryRowContext(ctx, query, id).Scan(
		&claim.ID, &claim.TenantID, &claim.PolicyID, &claim.ClaimNumber, &claim.EquipmentID,
		&claim.ProjectID, &claim.IncidentDate, &claim.ReportedDate, &claim.Description,
		(*string)(&claim.DamageType), (*string)(&claim.Status), &claim.ClaimedAmount, &claim.ApprovedAmount,
		&claim.SettledAmount, &claim.AdjusterNotes, &claim.CreatedBy, &claim.CreatedAt, &claim.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &domain.ErrClaimNotFound{ID: id.String()}
		}
		return nil, err
	}

	return &claim, nil
}

// GetByTenantID retrieves all claims for a tenant
func (cr *ClaimRepository) GetByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*domain.Claim, error) {
	query := `
		SELECT id, tenant_id, policy_id, claim_number, equipment_id, project_id, incident_date,
		       reported_date, description, damage_type, status, claimed_amount, approved_amount,
		       settled_amount, adjuster_notes, created_by, created_at, updated_at
		FROM claims WHERE tenant_id = $1 ORDER BY created_at DESC
	`

	rows, err := cr.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var claims []*domain.Claim
	for rows.Next() {
		var claim domain.Claim
		if err := rows.Scan(
			&claim.ID, &claim.TenantID, &claim.PolicyID, &claim.ClaimNumber, &claim.EquipmentID,
			&claim.ProjectID, &claim.IncidentDate, &claim.ReportedDate, &claim.Description,
			(*string)(&claim.DamageType), (*string)(&claim.Status), &claim.ClaimedAmount, &claim.ApprovedAmount,
			&claim.SettledAmount, &claim.AdjusterNotes, &claim.CreatedBy, &claim.CreatedAt, &claim.UpdatedAt,
		); err != nil {
			return nil, err
		}
		claims = append(claims, &claim)
	}

	return claims, rows.Err()
}

// GetByPolicyID retrieves all claims for a policy
func (cr *ClaimRepository) GetByPolicyID(ctx context.Context, policyID uuid.UUID) ([]*domain.Claim, error) {
	query := `
		SELECT id, tenant_id, policy_id, claim_number, equipment_id, project_id, incident_date,
		       reported_date, description, damage_type, status, claimed_amount, approved_amount,
		       settled_amount, adjuster_notes, created_by, created_at, updated_at
		FROM claims WHERE policy_id = $1 ORDER BY created_at DESC
	`

	rows, err := cr.db.QueryContext(ctx, query, policyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var claims []*domain.Claim
	for rows.Next() {
		var claim domain.Claim
		if err := rows.Scan(
			&claim.ID, &claim.TenantID, &claim.PolicyID, &claim.ClaimNumber, &claim.EquipmentID,
			&claim.ProjectID, &claim.IncidentDate, &claim.ReportedDate, &claim.Description,
			(*string)(&claim.DamageType), (*string)(&claim.Status), &claim.ClaimedAmount, &claim.ApprovedAmount,
			&claim.SettledAmount, &claim.AdjusterNotes, &claim.CreatedBy, &claim.CreatedAt, &claim.UpdatedAt,
		); err != nil {
			return nil, err
		}
		claims = append(claims, &claim)
	}

	return claims, rows.Err()
}

// Update updates an existing claim
func (cr *ClaimRepository) Update(ctx context.Context, claim *domain.Claim) error {
	query := `
		UPDATE claims 
		SET description = $1, status = $2, claimed_amount = $3, approved_amount = $4,
		    settled_amount = $5, adjuster_notes = $6, updated_at = $7
		WHERE id = $8
	`

	_, err := cr.db.ExecContext(ctx, query,
		claim.Description, string(claim.Status), claim.ClaimedAmount, claim.ApprovedAmount,
		claim.SettledAmount, claim.AdjusterNotes, claim.UpdatedAt, claim.ID,
	)

	return err
}

// Delete deletes a claim
func (cr *ClaimRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := cr.db.ExecContext(ctx, "DELETE FROM claims WHERE id = $1", id)
	return err
}

// GetClaimsByStatus retrieves claims with a specific status for a tenant
func (cr *ClaimRepository) GetClaimsByStatus(ctx context.Context, tenantID uuid.UUID, status domain.ClaimStatus) ([]*domain.Claim, error) {
	query := `
		SELECT id, tenant_id, policy_id, claim_number, equipment_id, project_id, incident_date,
		       reported_date, description, damage_type, status, claimed_amount, approved_amount,
		       settled_amount, adjuster_notes, created_by, created_at, updated_at
		FROM claims WHERE tenant_id = $1 AND status = $2 ORDER BY created_at DESC
	`

	rows, err := cr.db.QueryContext(ctx, query, tenantID, string(status))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var claims []*domain.Claim
	for rows.Next() {
		var claim domain.Claim
		if err := rows.Scan(
			&claim.ID, &claim.TenantID, &claim.PolicyID, &claim.ClaimNumber, &claim.EquipmentID,
			&claim.ProjectID, &claim.IncidentDate, &claim.ReportedDate, &claim.Description,
			(*string)(&claim.DamageType), (*string)(&claim.Status), &claim.ClaimedAmount, &claim.ApprovedAmount,
			&claim.SettledAmount, &claim.AdjusterNotes, &claim.CreatedBy, &claim.CreatedAt, &claim.UpdatedAt,
		); err != nil {
			return nil, err
		}
		claims = append(claims, &claim)
	}

	return claims, rows.Err()
}

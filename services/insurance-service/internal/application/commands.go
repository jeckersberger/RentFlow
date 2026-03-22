package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jeckersberger/rentflow/services/insurance-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/insurance-service/internal/ports"
	"github.com/rs/zerolog"
)

// PolicyService handles policy-related business logic
type PolicyService struct {
	repo   ports.PolicyRepository
	logger zerolog.Logger
}

// NewPolicyService creates a new policy service
func NewPolicyService(repo ports.PolicyRepository, logger zerolog.Logger) *PolicyService {
	return &PolicyService{
		repo:   repo,
		logger: logger,
	}
}

// CreatePolicy creates a new policy
func (ps *PolicyService) CreatePolicy(ctx context.Context, req CreatePolicyRequest) (*PolicyResponse, error) {
	policyType := domain.PolicyType(req.PolicyType)
	policy := &domain.Policy{
		ID:             uuid.New(),
		TenantID:       req.TenantID,
		PolicyNumber:   req.PolicyNumber,
		PolicyType:     policyType,
		Provider:       req.Provider,
		CoverageAmount: req.CoverageAmount,
		Deductible:     req.Deductible,
		PremiumAnnual:  req.PremiumAnnual,
		PremiumMonthly: req.PremiumMonthly,
		StartDate:      req.StartDate,
		EndDate:        req.EndDate,
		Status:         domain.PolicyStatusActive,
		Notes:          req.Notes,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := ps.repo.Create(ctx, policy); err != nil {
		ps.logger.Error().Err(err).Msg("failed to create policy")
		return nil, err
	}

	return policyToResponse(policy), nil
}

// GetPolicy retrieves a policy by ID
func (ps *PolicyService) GetPolicy(ctx context.Context, id uuid.UUID) (*PolicyResponse, error) {
	policy, err := ps.repo.GetByID(ctx, id)
	if err != nil {
		ps.logger.Error().Err(err).Str("id", id.String()).Msg("failed to get policy")
		return nil, err
	}
	return policyToResponse(policy), nil
}

// GetTenantPolicies retrieves all policies for a tenant
func (ps *PolicyService) GetTenantPolicies(ctx context.Context, tenantID uuid.UUID) ([]*PolicyResponse, error) {
	policies, err := ps.repo.GetByTenantID(ctx, tenantID)
	if err != nil {
		ps.logger.Error().Err(err).Str("tenant_id", tenantID.String()).Msg("failed to get tenant policies")
		return nil, err
	}

	responses := make([]*PolicyResponse, len(policies))
	for i, p := range policies {
		responses[i] = policyToResponse(p)
	}
	return responses, nil
}

// GetActivePolicies retrieves all active policies for a tenant
func (ps *PolicyService) GetActivePolicies(ctx context.Context, tenantID uuid.UUID) ([]*PolicyResponse, error) {
	policies, err := ps.repo.GetActiveByTenantID(ctx, tenantID)
	if err != nil {
		ps.logger.Error().Err(err).Str("tenant_id", tenantID.String()).Msg("failed to get active policies")
		return nil, err
	}

	responses := make([]*PolicyResponse, len(policies))
	for i, p := range policies {
		responses[i] = policyToResponse(p)
	}
	return responses, nil
}

// UpdatePolicy updates a policy
func (ps *PolicyService) UpdatePolicy(ctx context.Context, id uuid.UUID, req UpdatePolicyRequest) (*PolicyResponse, error) {
	policy, err := ps.repo.GetByID(ctx, id)
	if err != nil {
		ps.logger.Error().Err(err).Str("id", id.String()).Msg("failed to get policy for update")
		return nil, err
	}

	if req.Provider != nil {
		policy.Provider = *req.Provider
	}
	if req.CoverageAmount != nil {
		policy.CoverageAmount = *req.CoverageAmount
	}
	if req.Deductible != nil {
		policy.Deductible = *req.Deductible
	}
	if req.PremiumAnnual != nil {
		policy.PremiumAnnual = *req.PremiumAnnual
	}
	if req.PremiumMonthly != nil {
		policy.PremiumMonthly = *req.PremiumMonthly
	}
	if req.EndDate != nil {
		policy.EndDate = *req.EndDate
	}
	if req.Status != nil {
		policy.Status = domain.PolicyStatus(*req.Status)
	}
	if req.Notes != nil {
		policy.Notes = req.Notes
	}
	policy.UpdatedAt = time.Now()

	if err := ps.repo.Update(ctx, policy); err != nil {
		ps.logger.Error().Err(err).Str("id", id.String()).Msg("failed to update policy")
		return nil, err
	}

	return policyToResponse(policy), nil
}

// CheckCoverage checks if a policy can cover a specific amount
func (ps *PolicyService) CheckCoverage(ctx context.Context, policyID uuid.UUID, amount float64) (bool, error) {
	policy, err := ps.repo.GetByID(ctx, policyID)
	if err != nil {
		return false, err
	}
	return policy.CanCover(amount), nil
}

// GetRenewalReminders retrieves policies expiring soon
func (ps *PolicyService) GetRenewalReminders(ctx context.Context, daysUntilExpiry int) ([]*PolicyResponse, error) {
	policies, err := ps.repo.GetExpiringPolicies(ctx, daysUntilExpiry)
	if err != nil {
		ps.logger.Error().Err(err).Int("days", daysUntilExpiry).Msg("failed to get expiring policies")
		return nil, err
	}

	responses := make([]*PolicyResponse, len(policies))
	for i, p := range policies {
		responses[i] = policyToResponse(p)
	}
	return responses, nil
}

// ClaimService handles claim-related business logic
type ClaimService struct {
	claimRepo     ports.ClaimRepository
	policyRepo    ports.PolicyRepository
	claimItemRepo ports.ClaimItemRepository
	logger        zerolog.Logger
}

// NewClaimService creates a new claim service
func NewClaimService(
	claimRepo ports.ClaimRepository,
	policyRepo ports.PolicyRepository,
	claimItemRepo ports.ClaimItemRepository,
	logger zerolog.Logger,
) *ClaimService {
	return &ClaimService{
		claimRepo:     claimRepo,
		policyRepo:    policyRepo,
		claimItemRepo: claimItemRepo,
		logger:        logger,
	}
}

// CreateClaim creates a new claim
func (cs *ClaimService) CreateClaim(ctx context.Context, req CreateClaimRequest) (*ClaimResponse, error) {
	// Verify policy exists and is active
	policy, err := cs.policyRepo.GetByID(ctx, req.PolicyID)
	if err != nil {
		cs.logger.Error().Err(err).Str("policy_id", req.PolicyID.String()).Msg("failed to get policy")
		return nil, err
	}

	if !policy.IsActive() {
		return nil, &domain.ErrPolicyInactive{ID: policy.ID.String()}
	}

	claim := &domain.Claim{
		ID:            uuid.New(),
		TenantID:      req.TenantID,
		PolicyID:      req.PolicyID,
		ClaimNumber:   fmt.Sprintf("CLM-%d-%s", time.Now().Unix(), uuid.New().String()[:8]),
		EquipmentID:   req.EquipmentID,
		ProjectID:     req.ProjectID,
		IncidentDate:  req.IncidentDate,
		ReportedDate:  time.Now(),
		Description:   req.Description,
		DamageType:    domain.DamageType(req.DamageType),
		Status:        domain.ClaimStatusReported,
		ClaimedAmount: req.ClaimedAmount,
		CreatedBy:     req.CreatedBy,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := cs.claimRepo.Create(ctx, claim); err != nil {
		cs.logger.Error().Err(err).Msg("failed to create claim")
		return nil, err
	}

	return claimToResponse(claim), nil
}

// GetClaim retrieves a claim by ID
func (cs *ClaimService) GetClaim(ctx context.Context, id uuid.UUID) (*ClaimResponse, error) {
	claim, err := cs.claimRepo.GetByID(ctx, id)
	if err != nil {
		cs.logger.Error().Err(err).Str("id", id.String()).Msg("failed to get claim")
		return nil, err
	}
	return claimToResponse(claim), nil
}

// GetTenantClaims retrieves all claims for a tenant
func (cs *ClaimService) GetTenantClaims(ctx context.Context, tenantID uuid.UUID) ([]*ClaimResponse, error) {
	claims, err := cs.claimRepo.GetByTenantID(ctx, tenantID)
	if err != nil {
		cs.logger.Error().Err(err).Str("tenant_id", tenantID.String()).Msg("failed to get tenant claims")
		return nil, err
	}

	responses := make([]*ClaimResponse, len(claims))
	for i, c := range claims {
		responses[i] = claimToResponse(c)
	}
	return responses, nil
}

// UpdateClaim updates a claim
func (cs *ClaimService) UpdateClaim(ctx context.Context, id uuid.UUID, req UpdateClaimRequest) (*ClaimResponse, error) {
	claim, err := cs.claimRepo.GetByID(ctx, id)
	if err != nil {
		cs.logger.Error().Err(err).Str("id", id.String()).Msg("failed to get claim for update")
		return nil, err
	}

	if req.Description != nil {
		claim.Description = *req.Description
	}
	if req.AdjusterNotes != nil {
		claim.AdjusterNotes = req.AdjusterNotes
	}
	if req.ApprovedAmount != nil {
		claim.ApprovedAmount = req.ApprovedAmount
	}
	if req.SettledAmount != nil {
		claim.SettledAmount = req.SettledAmount
	}
	claim.UpdatedAt = time.Now()

	if err := cs.claimRepo.Update(ctx, claim); err != nil {
		cs.logger.Error().Err(err).Str("id", id.String()).Msg("failed to update claim")
		return nil, err
	}

	return claimToResponse(claim), nil
}

// SubmitClaim transitions claim to submitted status
func (cs *ClaimService) SubmitClaim(ctx context.Context, id uuid.UUID) (*ClaimResponse, error) {
	claim, err := cs.claimRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := claim.TransitionTo(domain.ClaimStatusSubmitted); err != nil {
		cs.logger.Error().Err(err).Str("id", id.String()).Msg("invalid claim transition")
		return nil, err
	}

	if err := cs.claimRepo.Update(ctx, claim); err != nil {
		cs.logger.Error().Err(err).Str("id", id.String()).Msg("failed to submit claim")
		return nil, err
	}

	return claimToResponse(claim), nil
}

// ApproveClaim transitions claim to approved status with amount
func (cs *ClaimService) ApproveClaim(ctx context.Context, id uuid.UUID, approvedAmount float64, notes *string) (*ClaimResponse, error) {
	claim, err := cs.claimRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := claim.TransitionTo(domain.ClaimStatusApproved); err != nil {
		cs.logger.Error().Err(err).Str("id", id.String()).Msg("invalid claim transition to approved")
		return nil, err
	}

	claim.ApprovedAmount = &approvedAmount
	if notes != nil {
		claim.AdjusterNotes = notes
	}
	claim.UpdatedAt = time.Now()

	if err := cs.claimRepo.Update(ctx, claim); err != nil {
		cs.logger.Error().Err(err).Str("id", id.String()).Msg("failed to approve claim")
		return nil, err
	}

	return claimToResponse(claim), nil
}

// RejectClaim transitions claim to rejected status
func (cs *ClaimService) RejectClaim(ctx context.Context, id uuid.UUID, notes string) (*ClaimResponse, error) {
	claim, err := cs.claimRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := claim.TransitionTo(domain.ClaimStatusRejected); err != nil {
		cs.logger.Error().Err(err).Str("id", id.String()).Msg("invalid claim transition to rejected")
		return nil, err
	}

	claim.AdjusterNotes = &notes
	claim.UpdatedAt = time.Now()

	if err := cs.claimRepo.Update(ctx, claim); err != nil {
		cs.logger.Error().Err(err).Str("id", id.String()).Msg("failed to reject claim")
		return nil, err
	}

	return claimToResponse(claim), nil
}

// SettleClaim transitions claim to settled status with settlement amount
func (cs *ClaimService) SettleClaim(ctx context.Context, id uuid.UUID, settledAmount float64, notes *string) (*ClaimResponse, error) {
	claim, err := cs.claimRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := claim.TransitionTo(domain.ClaimStatusSettled); err != nil {
		cs.logger.Error().Err(err).Str("id", id.String()).Msg("invalid claim transition to settled")
		return nil, err
	}

	claim.SettledAmount = &settledAmount
	if notes != nil {
		claim.AdjusterNotes = notes
	}
	claim.UpdatedAt = time.Now()

	if err := cs.claimRepo.Update(ctx, claim); err != nil {
		cs.logger.Error().Err(err).Str("id", id.String()).Msg("failed to settle claim")
		return nil, err
	}

	return claimToResponse(claim), nil
}

// CreateClaimItem adds an item to a claim
func (cs *ClaimService) CreateClaimItem(ctx context.Context, claimID uuid.UUID, req CreateClaimItemRequest) (*ClaimItemResponse, error) {
	item := &domain.ClaimItem{
		ID:               uuid.New(),
		ClaimID:          claimID,
		EquipmentID:      req.EquipmentID,
		Description:      req.Description,
		ReplacementValue: req.ReplacementValue,
		RepairCost:       req.RepairCost,
		PhotoURLs:        req.PhotoURLs,
		CreatedAt:        time.Now(),
	}

	if err := cs.claimItemRepo.Create(ctx, item); err != nil {
		cs.logger.Error().Err(err).Str("claim_id", claimID.String()).Msg("failed to create claim item")
		return nil, err
	}

	return claimItemToResponse(item), nil
}

// GetClaimItems retrieves all items for a claim
func (cs *ClaimService) GetClaimItems(ctx context.Context, claimID uuid.UUID) ([]*ClaimItemResponse, error) {
	items, err := cs.claimItemRepo.GetByClaimID(ctx, claimID)
	if err != nil {
		cs.logger.Error().Err(err).Str("claim_id", claimID.String()).Msg("failed to get claim items")
		return nil, err
	}

	responses := make([]*ClaimItemResponse, len(items))
	for i, item := range items {
		responses[i] = claimItemToResponse(item)
	}
	return responses, nil
}

// Helper functions to convert domain models to responses

func policyToResponse(p *domain.Policy) *PolicyResponse {
	return &PolicyResponse{
		ID:             p.ID,
		TenantID:       p.TenantID,
		PolicyNumber:   p.PolicyNumber,
		PolicyType:     string(p.PolicyType),
		Provider:       p.Provider,
		CoverageAmount: p.CoverageAmount,
		Deductible:     p.Deductible,
		PremiumAnnual:  p.PremiumAnnual,
		PremiumMonthly: p.PremiumMonthly,
		StartDate:      p.StartDate,
		EndDate:        p.EndDate,
		Status:         string(p.Status),
		Notes:          p.Notes,
		CreatedAt:      p.CreatedAt,
		UpdatedAt:      p.UpdatedAt,
	}
}

func claimToResponse(c *domain.Claim) *ClaimResponse {
	return &ClaimResponse{
		ID:             c.ID,
		TenantID:       c.TenantID,
		PolicyID:       c.PolicyID,
		ClaimNumber:    c.ClaimNumber,
		EquipmentID:    c.EquipmentID,
		ProjectID:      c.ProjectID,
		IncidentDate:   c.IncidentDate,
		ReportedDate:   c.ReportedDate,
		Description:    c.Description,
		DamageType:     string(c.DamageType),
		Status:         string(c.Status),
		ClaimedAmount:  c.ClaimedAmount,
		ApprovedAmount: c.ApprovedAmount,
		SettledAmount:  c.SettledAmount,
		AdjusterNotes:  c.AdjusterNotes,
		CreatedBy:      c.CreatedBy,
		CreatedAt:      c.CreatedAt,
		UpdatedAt:      c.UpdatedAt,
	}
}

func claimItemToResponse(ci *domain.ClaimItem) *ClaimItemResponse {
	return &ClaimItemResponse{
		ID:               ci.ID,
		ClaimID:          ci.ClaimID,
		EquipmentID:      ci.EquipmentID,
		Description:      ci.Description,
		ReplacementValue: ci.ReplacementValue,
		RepairCost:       ci.RepairCost,
		PhotoURLs:        ci.PhotoURLs,
		CreatedAt:        ci.CreatedAt,
	}
}

package application

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/insurance-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/insurance-service/internal/ports"
)

type PolicyService struct {
	policyRepo ports.PolicyRepository
	logger     *logger.Logger
}

func NewPolicyService(policyRepo ports.PolicyRepository, log *logger.Logger) *PolicyService {
	return &PolicyService{
		policyRepo: policyRepo,
		logger:     log,
	}
}

func (s *PolicyService) CreatePolicy(ctx context.Context, tenantID string, req CreatePolicyRequest) (*PolicyResponse, error) {
	if tenantID == "" {
		return nil, domain.ErrTenantIDRequired
	}
	if req.PolicyNumber == "" || req.Provider == "" {
		return nil, domain.ErrInvalidInput
	}

	startDate, _ := time.Parse(time.RFC3339, req.StartDate)
	endDate, _ := time.Parse(time.RFC3339, req.EndDate)

	policy := &domain.Policy{
		ID:             fmt.Sprintf("pol_%d", time.Now().UnixNano()),
		TenantID:       tenantID,
		PolicyNumber:   req.PolicyNumber,
		Provider:       req.Provider,
		Type:           domain.PolicyType(req.Type),
		CoverageAmount: req.CoverageAmount,
		Deductible:     req.Deductible,
		Premium:        req.Premium,
		StartDate:      startDate,
		EndDate:        endDate,
		Status:         domain.PolicyStatusActive,
		Notes:          req.Notes,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := s.policyRepo.Create(ctx, policy); err != nil {
		s.logger.Error("Failed to create policy", err)
		return nil, err
	}

	return PolicyToDTO(policy), nil
}

func (s *PolicyService) GetPolicy(ctx context.Context, tenantID, id string) (*PolicyResponse, error) {
	if tenantID == "" || id == "" {
		return nil, domain.ErrInvalidInput
	}

	policy, err := s.policyRepo.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	if policy == nil {
		return nil, domain.ErrPolicyNotFound
	}

	return PolicyToDTO(policy), nil
}

func (s *PolicyService) ListPolicies(ctx context.Context, tenantID string) ([]*PolicyResponse, error) {
	if tenantID == "" {
		return nil, domain.ErrTenantIDRequired
	}

	policies, err := s.policyRepo.ListByTenant(ctx, tenantID)
	if err != nil {
		s.logger.Error("Failed to list policies", err)
		return nil, err
	}

	dtos := make([]*PolicyResponse, len(policies))
	for i, p := range policies {
		dtos[i] = PolicyToDTO(p)
	}

	return dtos, nil
}

func (s *PolicyService) ListActivePolicies(ctx context.Context, tenantID string) ([]*PolicyResponse, error) {
	if tenantID == "" {
		return nil, domain.ErrTenantIDRequired
	}

	policies, err := s.policyRepo.ListActive(ctx, tenantID)
	if err != nil {
		s.logger.Error("Failed to list active policies", err)
		return nil, err
	}

	dtos := make([]*PolicyResponse, len(policies))
	for i, p := range policies {
		dtos[i] = PolicyToDTO(p)
	}

	return dtos, nil
}

func (s *PolicyService) ListExpiringPolicies(ctx context.Context, tenantID string, days int) ([]*PolicyResponse, error) {
	if tenantID == "" {
		return nil, domain.ErrTenantIDRequired
	}

	policies, err := s.policyRepo.ListExpiring(ctx, tenantID, days)
	if err != nil {
		s.logger.Error("Failed to list expiring policies", err)
		return nil, err
	}

	dtos := make([]*PolicyResponse, len(policies))
	for i, p := range policies {
		dtos[i] = PolicyToDTO(p)
	}

	return dtos, nil
}

func (s *PolicyService) UpdatePolicy(ctx context.Context, tenantID, id string, req UpdatePolicyRequest) (*PolicyResponse, error) {
	if tenantID == "" || id == "" {
		return nil, domain.ErrInvalidInput
	}

	policy, err := s.policyRepo.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	if policy == nil {
		return nil, domain.ErrPolicyNotFound
	}

	policy.Provider = req.Provider
	policy.CoverageAmount = req.CoverageAmount
	policy.Deductible = req.Deductible
	policy.Premium = req.Premium
	policy.Notes = req.Notes
	policy.UpdatedAt = time.Now()

	if err := s.policyRepo.Update(ctx, policy); err != nil {
		s.logger.Error("Failed to update policy", err)
		return nil, err
	}

	return PolicyToDTO(policy), nil
}

func (s *PolicyService) DeletePolicy(ctx context.Context, tenantID, id string) error {
	if tenantID == "" || id == "" {
		return domain.ErrInvalidInput
	}

	return s.policyRepo.Delete(ctx, tenantID, id)
}

type ClaimService struct {
	claimRepo  ports.ClaimRepository
	policyRepo ports.PolicyRepository
	logger     *logger.Logger
}

func NewClaimService(claimRepo ports.ClaimRepository, policyRepo ports.PolicyRepository, log *logger.Logger) *ClaimService {
	return &ClaimService{
		claimRepo:  claimRepo,
		policyRepo: policyRepo,
		logger:     log,
	}
}

func (s *ClaimService) FileClaim(ctx context.Context, tenantID string, req CreateClaimRequest) (*ClaimResponse, error) {
	if tenantID == "" {
		return nil, domain.ErrTenantIDRequired
	}
	if req.PolicyID == "" || req.EquipmentID == "" {
		return nil, domain.ErrInvalidInput
	}

	policy, err := s.policyRepo.GetByID(ctx, tenantID, req.PolicyID)
	if err != nil {
		return nil, err
	}
	if policy == nil {
		return nil, domain.ErrPolicyNotFound
	}

	incidentDate, _ := time.Parse(time.RFC3339, req.IncidentDate)

	claimAmount := req.DamageAmount - policy.Deductible
	if claimAmount < 0 {
		claimAmount = 0
	}
	if claimAmount > policy.CoverageAmount {
		claimAmount = policy.CoverageAmount
	}

	claim := &domain.Claim{
		ID:           fmt.Sprintf("clm_%d", time.Now().UnixNano()),
		TenantID:     tenantID,
		PolicyID:     req.PolicyID,
		EquipmentID:  req.EquipmentID,
		IncidentDate: incidentDate,
		Description:  req.Description,
		DamageAmount: req.DamageAmount,
		ClaimAmount:  claimAmount,
		Status:       domain.ClaimStatusFiled,
		Documents:    req.Documents,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.claimRepo.Create(ctx, claim); err != nil {
		s.logger.Error("Failed to file claim", err)
		return nil, err
	}

	return ClaimToDTO(claim), nil
}

func (s *ClaimService) GetClaim(ctx context.Context, tenantID, id string) (*ClaimResponse, error) {
	if tenantID == "" || id == "" {
		return nil, domain.ErrInvalidInput
	}

	claim, err := s.claimRepo.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	if claim == nil {
		return nil, domain.ErrClaimNotFound
	}

	return ClaimToDTO(claim), nil
}

func (s *ClaimService) ListClaims(ctx context.Context, tenantID string) ([]*ClaimResponse, error) {
	if tenantID == "" {
		return nil, domain.ErrTenantIDRequired
	}

	claims, err := s.claimRepo.ListByTenant(ctx, tenantID)
	if err != nil {
		s.logger.Error("Failed to list claims", err)
		return nil, err
	}

	dtos := make([]*ClaimResponse, len(claims))
	for i, c := range claims {
		dtos[i] = ClaimToDTO(c)
	}

	return dtos, nil
}

func (s *ClaimService) ListClaimsByEquipment(ctx context.Context, tenantID, equipmentID string) ([]*ClaimResponse, error) {
	if tenantID == "" || equipmentID == "" {
		return nil, domain.ErrInvalidInput
	}

	claims, err := s.claimRepo.ListByEquipment(ctx, tenantID, equipmentID)
	if err != nil {
		s.logger.Error("Failed to list claims by equipment", err)
		return nil, err
	}

	dtos := make([]*ClaimResponse, len(claims))
	for i, c := range claims {
		dtos[i] = ClaimToDTO(c)
	}

	return dtos, nil
}

func (s *ClaimService) UpdateClaim(ctx context.Context, tenantID, id string, req UpdateClaimRequest) (*ClaimResponse, error) {
	if tenantID == "" || id == "" {
		return nil, domain.ErrInvalidInput
	}

	claim, err := s.claimRepo.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	if claim == nil {
		return nil, domain.ErrClaimNotFound
	}

	claim.Description = req.Description
	claim.DamageAmount = req.DamageAmount
	claim.Documents = req.Documents
	claim.UpdatedAt = time.Now()

	if err := s.claimRepo.Update(ctx, claim); err != nil {
		s.logger.Error("Failed to update claim", err)
		return nil, err
	}

	return ClaimToDTO(claim), nil
}

func (s *ClaimService) ApproveClaim(ctx context.Context, tenantID, id string, req ApproveClaimRequest) (*ClaimResponse, error) {
	if tenantID == "" || id == "" {
		return nil, domain.ErrInvalidInput
	}

	claim, err := s.claimRepo.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	if claim == nil {
		return nil, domain.ErrClaimNotFound
	}

	claim.Status = domain.ClaimStatusApproved
	claim.ClaimAmount = req.ApprovedAmount
	claim.UpdatedAt = time.Now()

	if err := s.claimRepo.Update(ctx, claim); err != nil {
		s.logger.Error("Failed to approve claim", err)
		return nil, err
	}

	return ClaimToDTO(claim), nil
}

func (s *ClaimService) RejectClaim(ctx context.Context, tenantID, id, reason string) (*ClaimResponse, error) {
	if tenantID == "" || id == "" {
		return nil, domain.ErrInvalidInput
	}

	claim, err := s.claimRepo.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	if claim == nil {
		return nil, domain.ErrClaimNotFound
	}

	claim.Status = domain.ClaimStatusRejected
	claim.ClaimAmount = 0
	claim.UpdatedAt = time.Now()

	if err := s.claimRepo.Update(ctx, claim); err != nil {
		s.logger.Error("Failed to reject claim", err)
		return nil, err
	}

	return ClaimToDTO(claim), nil
}

type RiskService struct {
	riskRepo ports.RiskAssessmentRepository
	logger   *logger.Logger
}

func NewRiskService(riskRepo ports.RiskAssessmentRepository, log *logger.Logger) *RiskService {
	return &RiskService{
		riskRepo: riskRepo,
		logger:   log,
	}
}

func (s *RiskService) AssessEquipment(ctx context.Context, tenantID string, req AssessRiskRequest) (*RiskAssessmentResponse, error) {
	if tenantID == "" || req.EquipmentID == "" {
		return nil, domain.ErrInvalidInput
	}

	riskScore := calculateRiskScore(req.Age, req.UsageHours)

	factors := map[string]interface{}{
		"age":           req.Age,
		"value":         req.Value,
		"usage_hours":   req.UsageHours,
		"depreciation":  float64(req.Age) * 0.15,
	}
	factorsJSON, _ := json.Marshal(factors)

	assessment := &domain.RiskAssessment{
		ID:          fmt.Sprintf("rsk_%d", time.Now().UnixNano()),
		TenantID:    tenantID,
		EquipmentID: req.EquipmentID,
		RiskScore:   riskScore,
		Factors:     factorsJSON,
		AssessedAt:  time.Now(),
	}

	if err := s.riskRepo.Create(ctx, assessment); err != nil {
		s.logger.Error("Failed to create risk assessment", err)
		return nil, err
	}

	return RiskAssessmentToDTO(assessment), nil
}

func (s *RiskService) GetRiskAssessment(ctx context.Context, tenantID, equipmentID string) (*RiskAssessmentResponse, error) {
	if tenantID == "" || equipmentID == "" {
		return nil, domain.ErrInvalidInput
	}

	assessment, err := s.riskRepo.GetByEquipment(ctx, tenantID, equipmentID)
	if err != nil {
		return nil, err
	}
	if assessment == nil {
		return nil, domain.ErrRiskAssessmentNotFound
	}

	return RiskAssessmentToDTO(assessment), nil
}

func (s *RiskService) ListAssessments(ctx context.Context, tenantID string) ([]*RiskAssessmentResponse, error) {
	if tenantID == "" {
		return nil, domain.ErrTenantIDRequired
	}

	assessments, err := s.riskRepo.ListByTenant(ctx, tenantID)
	if err != nil {
		s.logger.Error("Failed to list risk assessments", err)
		return nil, err
	}

	dtos := make([]*RiskAssessmentResponse, len(assessments))
	for i, a := range assessments {
		dtos[i] = RiskAssessmentToDTO(a)
	}

	return dtos, nil
}

func calculateRiskScore(age, usageHours int) float64 {
	// Simple risk calculation: higher age and usage = higher risk
	ageRisk := float64(age) * 0.05 // 5% per year
	usageRisk := float64(usageHours) / 10000.0 * 0.3 // scale to 0-0.3

	score := ageRisk + usageRisk
	if score > 1.0 {
		score = 1.0
	}
	return score
}

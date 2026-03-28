package application

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/services/insurance/internal/domain"
)

// ---------------------------------------------------------------------------
// Request DTOs
// ---------------------------------------------------------------------------

type CreatePolicyRequest struct {
	Name           string `json:"name"`
	Provider       string `json:"provider"`
	PolicyNumber   string `json:"policy_number"`
	CoverageType   string `json:"coverage_type"`
	CoverageAmount int64  `json:"coverage_amount"`
	Deductible     int64  `json:"deductible"`
	Premium        int64  `json:"premium"`
	StartDate      string `json:"start_date"`
	EndDate        string `json:"end_date"`
	IsActive       *bool  `json:"is_active"`
	Notes          string `json:"notes"`
}

type UpdatePolicyRequest struct {
	Name           *string `json:"name"`
	Provider       *string `json:"provider"`
	PolicyNumber   *string `json:"policy_number"`
	CoverageType   *string `json:"coverage_type"`
	CoverageAmount *int64  `json:"coverage_amount"`
	Deductible     *int64  `json:"deductible"`
	Premium        *int64  `json:"premium"`
	StartDate      *string `json:"start_date"`
	EndDate        *string `json:"end_date"`
	IsActive       *bool   `json:"is_active"`
	Notes          *string `json:"notes"`
}

// ---------------------------------------------------------------------------
// Service
// ---------------------------------------------------------------------------

type PolicyService struct {
	repo   domain.InsurancePolicyRepository
	logger zerolog.Logger
}

func NewPolicyService(repo domain.InsurancePolicyRepository, logger zerolog.Logger) *PolicyService {
	return &PolicyService{
		repo:   repo,
		logger: logger.With().Str("service", "policy").Logger(),
	}
}

func (s *PolicyService) Create(ctx context.Context, tenantID uuid.UUID, req CreatePolicyRequest) (*domain.InsurancePolicy, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}

	coverageType := "all_risk"
	if req.CoverageType != "" {
		coverageType = req.CoverageType
	}
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	policy := &domain.InsurancePolicy{
		ID:             uuid.New(),
		TenantID:       tenantID,
		Name:           req.Name,
		Provider:       req.Provider,
		PolicyNumber:   req.PolicyNumber,
		CoverageType:   coverageType,
		CoverageAmount: req.CoverageAmount,
		Deductible:     req.Deductible,
		Premium:        req.Premium,
		IsActive:       isActive,
		Notes:          req.Notes,
	}

	if req.StartDate != "" {
		policy.StartDate = &req.StartDate
	}
	if req.EndDate != "" {
		policy.EndDate = &req.EndDate
	}

	if err := s.repo.Create(ctx, policy); err != nil {
		return nil, fmt.Errorf("create policy: %w", err)
	}

	s.logger.Info().Str("policy_id", policy.ID.String()).Msg("insurance policy created")
	return policy, nil
}

func (s *PolicyService) GetByID(ctx context.Context, id, tenantID uuid.UUID) (*domain.InsurancePolicy, error) {
	return s.repo.GetByID(ctx, id, tenantID)
}

func (s *PolicyService) List(ctx context.Context, tenantID uuid.UUID, filter domain.PolicyFilter) ([]*domain.InsurancePolicy, int64, error) {
	return s.repo.List(ctx, tenantID, filter)
}

func (s *PolicyService) Update(ctx context.Context, id, tenantID uuid.UUID, req UpdatePolicyRequest) (*domain.InsurancePolicy, error) {
	existing, err := s.repo.GetByID(ctx, id, tenantID)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		existing.Name = *req.Name
	}
	if req.Provider != nil {
		existing.Provider = *req.Provider
	}
	if req.PolicyNumber != nil {
		existing.PolicyNumber = *req.PolicyNumber
	}
	if req.CoverageType != nil {
		existing.CoverageType = *req.CoverageType
	}
	if req.CoverageAmount != nil {
		existing.CoverageAmount = *req.CoverageAmount
	}
	if req.Deductible != nil {
		existing.Deductible = *req.Deductible
	}
	if req.Premium != nil {
		existing.Premium = *req.Premium
	}
	if req.StartDate != nil {
		existing.StartDate = req.StartDate
	}
	if req.EndDate != nil {
		existing.EndDate = req.EndDate
	}
	if req.IsActive != nil {
		existing.IsActive = *req.IsActive
	}
	if req.Notes != nil {
		existing.Notes = *req.Notes
	}

	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, err
	}

	return existing, nil
}

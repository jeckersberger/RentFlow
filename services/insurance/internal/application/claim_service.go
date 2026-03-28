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

type CreateClaimRequest struct {
	PolicyID     string `json:"policy_id"`
	EquipmentID  string `json:"equipment_id"`
	ClaimNumber  string `json:"claim_number"`
	Description  string `json:"description"`
	DamageAmount int64  `json:"damage_amount"`
	ClaimAmount  int64  `json:"claim_amount"`
	IncidentDate string `json:"incident_date"`
	Notes        string `json:"notes"`
}

type UpdateClaimRequest struct {
	ClaimNumber  *string `json:"claim_number"`
	Description  *string `json:"description"`
	DamageAmount *int64  `json:"damage_amount"`
	ClaimAmount  *int64  `json:"claim_amount"`
	Status       *string `json:"status"`
	IncidentDate *string `json:"incident_date"`
	Notes        *string `json:"notes"`
}

// ---------------------------------------------------------------------------
// Service
// ---------------------------------------------------------------------------

type ClaimService struct {
	repo   domain.InsuranceClaimRepository
	logger zerolog.Logger
}

func NewClaimService(repo domain.InsuranceClaimRepository, logger zerolog.Logger) *ClaimService {
	return &ClaimService{
		repo:   repo,
		logger: logger.With().Str("service", "claim").Logger(),
	}
}

func (s *ClaimService) Create(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID, req CreateClaimRequest) (*domain.InsuranceClaim, error) {
	if req.Description == "" {
		return nil, fmt.Errorf("description is required")
	}
	if req.PolicyID == "" {
		return nil, fmt.Errorf("policy_id is required")
	}

	policyID, err := uuid.Parse(req.PolicyID)
	if err != nil {
		return nil, fmt.Errorf("invalid policy_id: %w", err)
	}

	claim := &domain.InsuranceClaim{
		ID:           uuid.New(),
		PolicyID:     policyID,
		TenantID:     tenantID,
		ClaimNumber:  req.ClaimNumber,
		Description:  req.Description,
		DamageAmount: req.DamageAmount,
		ClaimAmount:  req.ClaimAmount,
		Status:       "submitted",
		Notes:        req.Notes,
		CreatedBy:    &userID,
	}

	if req.EquipmentID != "" {
		eqID, err := uuid.Parse(req.EquipmentID)
		if err != nil {
			return nil, fmt.Errorf("invalid equipment_id: %w", err)
		}
		claim.EquipmentID = &eqID
	}

	if req.IncidentDate != "" {
		claim.IncidentDate = &req.IncidentDate
	}

	if err := s.repo.Create(ctx, claim); err != nil {
		return nil, fmt.Errorf("create claim: %w", err)
	}

	s.logger.Info().Str("claim_id", claim.ID.String()).Msg("insurance claim created")
	return claim, nil
}

func (s *ClaimService) GetByID(ctx context.Context, id, tenantID uuid.UUID) (*domain.InsuranceClaim, error) {
	return s.repo.GetByID(ctx, id, tenantID)
}

func (s *ClaimService) List(ctx context.Context, tenantID uuid.UUID, filter domain.ClaimFilter) ([]*domain.InsuranceClaim, int64, error) {
	return s.repo.List(ctx, tenantID, filter)
}

func (s *ClaimService) Update(ctx context.Context, id, tenantID uuid.UUID, req UpdateClaimRequest) (*domain.InsuranceClaim, error) {
	existing, err := s.repo.GetByID(ctx, id, tenantID)
	if err != nil {
		return nil, err
	}

	if req.ClaimNumber != nil {
		existing.ClaimNumber = *req.ClaimNumber
	}
	if req.Description != nil {
		existing.Description = *req.Description
	}
	if req.DamageAmount != nil {
		existing.DamageAmount = *req.DamageAmount
	}
	if req.ClaimAmount != nil {
		existing.ClaimAmount = *req.ClaimAmount
	}
	if req.Status != nil {
		existing.Status = *req.Status
	}
	if req.IncidentDate != nil {
		existing.IncidentDate = req.IncidentDate
	}
	if req.Notes != nil {
		existing.Notes = *req.Notes
	}

	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, err
	}

	return existing, nil
}

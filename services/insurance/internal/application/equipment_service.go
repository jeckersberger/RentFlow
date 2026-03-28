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

type AddEquipmentRequest struct {
	EquipmentID  string `json:"equipment_id"`
	InsuredValue int64  `json:"insured_value"`
}

// ---------------------------------------------------------------------------
// Service
// ---------------------------------------------------------------------------

type EquipmentService struct {
	repo       domain.InsuredEquipmentRepository
	policyRepo domain.InsurancePolicyRepository
	logger     zerolog.Logger
}

func NewEquipmentService(repo domain.InsuredEquipmentRepository, policyRepo domain.InsurancePolicyRepository, logger zerolog.Logger) *EquipmentService {
	return &EquipmentService{
		repo:       repo,
		policyRepo: policyRepo,
		logger:     logger.With().Str("service", "equipment").Logger(),
	}
}

func (s *EquipmentService) Add(ctx context.Context, policyID uuid.UUID, tenantID uuid.UUID, req AddEquipmentRequest) (*domain.InsuredEquipment, error) {
	// Verify that the policy exists and belongs to the tenant.
	if _, err := s.policyRepo.GetByID(ctx, policyID, tenantID); err != nil {
		return nil, err
	}

	if req.EquipmentID == "" {
		return nil, fmt.Errorf("equipment_id is required")
	}
	equipmentID, err := uuid.Parse(req.EquipmentID)
	if err != nil {
		return nil, fmt.Errorf("invalid equipment_id: %w", err)
	}

	eq := &domain.InsuredEquipment{
		ID:           uuid.New(),
		PolicyID:     policyID,
		EquipmentID:  equipmentID,
		InsuredValue: req.InsuredValue,
	}

	if err := s.repo.Add(ctx, eq); err != nil {
		return nil, fmt.Errorf("add insured equipment: %w", err)
	}

	s.logger.Info().Str("equipment_id", eq.ID.String()).Msg("insured equipment added")
	return eq, nil
}

func (s *EquipmentService) ListByPolicy(ctx context.Context, policyID uuid.UUID, tenantID uuid.UUID, filter domain.EquipmentFilter) ([]*domain.InsuredEquipment, int64, error) {
	// Verify that the policy exists and belongs to the tenant.
	if _, err := s.policyRepo.GetByID(ctx, policyID, tenantID); err != nil {
		return nil, 0, err
	}

	return s.repo.ListByPolicy(ctx, policyID, filter)
}

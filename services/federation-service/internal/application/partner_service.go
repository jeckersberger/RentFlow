package application

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/federation-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/federation-service/internal/ports"
)

type PartnerService struct {
	partnerRepo ports.PartnerRepository
	log         logger.Logger
}

func NewPartnerService(pr ports.PartnerRepository, log logger.Logger) *PartnerService {
	return &PartnerService{
		partnerRepo: pr,
		log:         log,
	}
}

func (s *PartnerService) CreatePartner(ctx context.Context, tenantID uuid.UUID, req *CreatePartnerRequest) (*PartnerResponse, error) {
	partner := &domain.FederationPartner{
		ID:                uuid.New(),
		TenantID:          tenantID,
		PartnerName:       req.PartnerName,
		PartnerEndpoint:   req.PartnerEndpoint,
		Status:            domain.PartnerStatusPending,
		TrustLevel:        domain.TrustLevelBasic,
		SharedCategories: req.SharedCategories,
		DataPolicy:       req.DataPolicy,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := s.partnerRepo.Create(ctx, partner); err != nil {
		s.log.Error("Failed to create partner", err)
		return nil, err
	}

	return s.partnerToResponse(partner), nil
}

func (s *PartnerService) GetPartner(ctx context.Context, id uuid.UUID) (*PartnerResponse, error) {
	partner, err := s.partnerRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.partnerToResponse(partner), nil
}

func (s *PartnerService) ListPartners(ctx context.Context, tenantID uuid.UUID) ([]*PartnerResponse, error) {
	partners, err := s.partnerRepo.ListByTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	var responses []*PartnerResponse
	for _, p := range partners {
		responses = append(responses, s.partnerToResponse(p))
	}
	return responses, nil
}

func (s *PartnerService) ActivatePartner(ctx context.Context, id uuid.UUID) error {
	partner, err := s.partnerRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	partner.Status = domain.PartnerStatusActive
	partner.UpdatedAt = time.Now()
	return s.partnerRepo.Update(ctx, partner)
}

func (s *PartnerService) SuspendPartner(ctx context.Context, id uuid.UUID) error {
	partner, err := s.partnerRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	partner.Status = domain.PartnerStatusSuspended
	partner.UpdatedAt = time.Now()
	return s.partnerRepo.Update(ctx, partner)
}

func (s *PartnerService) UpdateTrustLevel(ctx context.Context, id uuid.UUID, level string) error {
	partner, err := s.partnerRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	partner.TrustLevel = domain.TrustLevel(level)
	partner.UpdatedAt = time.Now()
	return s.partnerRepo.Update(ctx, partner)
}

func (s *PartnerService) DeletePartner(ctx context.Context, id uuid.UUID) error {
	return s.partnerRepo.Delete(ctx, id)
}

func (s *PartnerService) partnerToResponse(p *domain.FederationPartner) *PartnerResponse {
	return &PartnerResponse{
		ID:              p.ID,
		TenantID:        p.TenantID,
		PartnerName:     p.PartnerName,
		PartnerEndpoint: p.PartnerEndpoint,
		Status:          string(p.Status),
		TrustLevel:      string(p.TrustLevel),
		CertFingerprint: p.CertFingerprint,
		CertExpiresAt:   p.CertExpiresAt,
		SharedCategories: p.SharedCategories,
		CreatedAt:       p.CreatedAt,
	}
}

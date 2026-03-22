package application

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/federation-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/federation-service/internal/ports"
)

type SharingService struct {
	requestRepo ports.SubRentalRequestRepository
	cacheRepo   ports.EquipmentCacheRepository
	partnerRepo ports.PartnerRepository
	log         logger.Logger
}

func NewSharingService(
	rr ports.SubRentalRequestRepository,
	cr ports.EquipmentCacheRepository,
	pr ports.PartnerRepository,
	log logger.Logger,
) *SharingService {
	return &SharingService{
		requestRepo: rr,
		cacheRepo:   cr,
		partnerRepo: pr,
		log:         log,
	}
}

func (s *SharingService) CreateRequest(ctx context.Context, tenantID uuid.UUID, req *CreateSubRentalRequestRequest) (*SubRentalRequestResponse, error) {
	srr := &domain.SubRentalRequest{
		ID:                   uuid.New(),
		TenantID:             tenantID,
		PartnerID:            req.PartnerID,
		Direction:            domain.RequestDirection(req.Direction),
		Status:               domain.RequestStatusPending,
		EquipmentCategory:    req.EquipmentCategory,
		EquipmentDescription: req.EquipmentDescription,
		Quantity:             req.Quantity,
		StartDate:            req.StartDate,
		EndDate:              req.EndDate,
		DailyRate:            req.DailyRate,
		TotalAmount:          req.DailyRate * float64(req.Quantity) * float64(req.EndDate.Sub(req.StartDate).Hours()/24),
		Notes:                req.Notes,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	if err := s.requestRepo.Create(ctx, srr); err != nil {
		s.log.Error("Failed to create sub-rental request", err)
		return nil, err
	}

	s.log.Info("Created sub-rental request", "request_id", srr.ID, "partner_id", srr.PartnerID)
	return s.requestToResponse(srr), nil
}

func (s *SharingService) GetRequest(ctx context.Context, id uuid.UUID) (*SubRentalRequestResponse, error) {
	srr, err := s.requestRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.requestToResponse(srr), nil
}

func (s *SharingService) ListRequests(ctx context.Context, tenantID uuid.UUID) ([]*SubRentalRequestResponse, error) {
	srrs, err := s.requestRepo.ListByTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	var responses []*SubRentalRequestResponse
	for _, srr := range srrs {
		responses = append(responses, s.requestToResponse(srr))
	}
	return responses, nil
}

func (s *SharingService) AcceptRequest(ctx context.Context, id uuid.UUID) error {
	srr, err := s.requestRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	srr.Status = domain.RequestStatusAccepted
	srr.UpdatedAt = time.Now()
	return s.requestRepo.Update(ctx, srr)
}

func (s *SharingService) RejectRequest(ctx context.Context, id uuid.UUID) error {
	srr, err := s.requestRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	srr.Status = domain.RequestStatusRejected
	srr.UpdatedAt = time.Now()
	return s.requestRepo.Update(ctx, srr)
}

func (s *SharingService) CompleteRequest(ctx context.Context, id uuid.UUID, handoverDocID uuid.UUID) error {
	srr, err := s.requestRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	srr.Status = domain.RequestStatusCompleted
	srr.HandoverDocumentID = &handoverDocID
	srr.UpdatedAt = time.Now()
	return s.requestRepo.Update(ctx, srr)
}

func (s *SharingService) SyncPartnerEquipment(ctx context.Context, partnerID uuid.UUID) error {
	s.log.Info("Syncing partner equipment cache", "partner_id", partnerID)

	// Clear existing cache
	if err := s.cacheRepo.DeleteByPartner(ctx, partnerID); err != nil {
		return err
	}

	// In a real implementation, would fetch from partner service
	return nil
}

func (s *SharingService) GetPartnerEquipment(ctx context.Context, partnerID uuid.UUID) ([]*EquipmentCacheResponse, error) {
	caches, err := s.cacheRepo.ListByPartner(ctx, partnerID)
	if err != nil {
		return nil, err
	}

	var responses []*EquipmentCacheResponse
	for _, c := range caches {
		responses = append(responses, s.cacheToResponse(c))
	}
	return responses, nil
}

func (s *SharingService) requestToResponse(srr *domain.SubRentalRequest) *SubRentalRequestResponse {
	return &SubRentalRequestResponse{
		ID:                   srr.ID,
		TenantID:             srr.TenantID,
		PartnerID:            srr.PartnerID,
		Direction:            string(srr.Direction),
		Status:               string(srr.Status),
		EquipmentCategory:    srr.EquipmentCategory,
		EquipmentDescription: srr.EquipmentDescription,
		Quantity:             srr.Quantity,
		StartDate:            srr.StartDate,
		EndDate:              srr.EndDate,
		DailyRate:            srr.DailyRate,
		TotalAmount:          srr.TotalAmount,
		HandoverDocumentID:   srr.HandoverDocumentID,
		InvoiceID:            srr.InvoiceID,
		CreatedAt:            srr.CreatedAt,
	}
}

func (s *SharingService) cacheToResponse(c *domain.PartnerEquipmentCache) *EquipmentCacheResponse {
	return &EquipmentCacheResponse{
		ID:               c.ID,
		PartnerID:        c.PartnerID,
		Category:         c.Category,
		ItemName:         c.ItemName,
		QuantityAvailable: c.QuantityAvailable,
		DailyRate:        c.DailyRate,
		LastSyncedAt:     c.LastSyncedAt,
	}
}

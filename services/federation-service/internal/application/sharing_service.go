package application

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
	httpClient  *http.Client
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
		httpClient:  &http.Client{Timeout: 10 * time.Second},
		log:         log,
	}
}

func (s *SharingService) CreateRequest(ctx context.Context, tenantID uuid.UUID, req *CreateSubRentalRequestRequest) (*SubRentalRequestResponse, error) {
	// Validate that partner shares the requested equipment category
	partner, err := s.partnerRepo.GetByID(ctx, req.PartnerID)
	if err != nil {
		s.log.Error("Failed to fetch partner", err)
		return nil, err
	}

	// Check if equipment category is in shared_categories
	categoryAllowed := false
	for _, category := range partner.SharedCategories {
		if category == req.EquipmentCategory {
			categoryAllowed = true
			break
		}
	}
	if !categoryAllowed {
		s.log.Warn("Equipment category not shared by partner", "partner_id", req.PartnerID, "category", req.EquipmentCategory)
		return nil, domain.ErrCategoryNotShared
	}

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

	// Integration point: invoice-service
	// When a sub-rental request is completed, the federation service triggers invoice generation
	// in the invoice-service via a POST to /api/v1/invoices/federation with the request details.
	// The returned invoice_id is stored here for reference and tracking.
	// This allows the invoice-service to manage billing while federation-service owns the sub-rental lifecycle.
	invoiceID := uuid.New() // In production, this would be returned from the invoice-service
	srr.InvoiceID = &invoiceID

	srr.UpdatedAt = time.Now()
	return s.requestRepo.Update(ctx, srr)
}

func (s *SharingService) SyncPartnerEquipment(ctx context.Context, partnerID uuid.UUID) error {
	s.log.Info("Syncing partner equipment cache", "partner_id", partnerID)

	// Fetch partner details
	partner, err := s.partnerRepo.GetByID(ctx, partnerID)
	if err != nil {
		s.log.Error("Failed to fetch partner for sync", err)
		return err
	}

	// Make HTTP GET request to partner endpoint
	url := fmt.Sprintf("%s/api/v1/federation/equipment", partner.PartnerEndpoint)
	resp, err := s.httpClient.Get(url)
	if err != nil {
		s.log.Warn("Failed to reach partner equipment endpoint", err, "partner_id", partnerID, "url", url)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.log.Warn("Partner returned non-200 status", nil, "partner_id", partnerID, "status", resp.StatusCode)
		return fmt.Errorf("partner returned status %d", resp.StatusCode)
	}

	// Parse JSON response into PartnerEquipmentCache items
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		s.log.Error("Failed to read partner response body", err)
		return err
	}

	var equipmentList []EquipmentCacheResponse
	if err := json.Unmarshal(body, &equipmentList); err != nil {
		s.log.Error("Failed to parse equipment response", err)
		return err
	}

	// Clear existing cache
	if err := s.cacheRepo.DeleteByPartner(ctx, partnerID); err != nil {
		s.log.Error("Failed to clear existing cache", err)
		return err
	}

	// Store new equipment in cache repository
	for _, eq := range equipmentList {
		cache := &domain.PartnerEquipmentCache{
			ID:                eq.ID,
			PartnerID:         partnerID,
			Category:          eq.Category,
			ItemName:          eq.ItemName,
			QuantityAvailable: eq.QuantityAvailable,
			DailyRate:         eq.DailyRate,
			LastSyncedAt:      time.Now(),
		}
		if err := s.cacheRepo.Create(ctx, cache); err != nil {
			s.log.Error("Failed to store equipment in cache", err)
			return err
		}
	}

	s.log.Info("Successfully synced partner equipment", "partner_id", partnerID, "item_count", len(equipmentList))
	return nil
}

func (s *SharingService) GetPartnerEquipment(ctx context.Context, partnerID uuid.UUID) ([]*EquipmentCacheResponse, error) {
	// Fetch partner to check shared_categories policy
	partner, err := s.partnerRepo.GetByID(ctx, partnerID)
	if err != nil {
		return nil, err
	}

	caches, err := s.cacheRepo.ListByPartner(ctx, partnerID)
	if err != nil {
		return nil, err
	}

	var responses []*EquipmentCacheResponse
	for _, c := range caches {
		// Only include equipment in categories that partner has shared
		categoryShared := false
		for _, sharedCat := range partner.SharedCategories {
			if sharedCat == c.Category {
				categoryShared = true
				break
			}
		}
		if categoryShared {
			responses = append(responses, s.cacheToResponse(c))
		}
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

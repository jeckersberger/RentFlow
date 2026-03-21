package application

import (
	"context"
	"fmt"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/federation-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/federation-service/internal/ports"
)

type SharingService struct {
	equipmentShareRepo ports.EquipmentShareRepository
	shareRequestRepo   ports.ShareRequestRepository
	peerRepo           ports.PeerRepository
	logger             logger.Logger
}

func NewSharingService(
	eqRepo ports.EquipmentShareRepository,
	reqRepo ports.ShareRequestRepository,
	peerRepo ports.PeerRepository,
	log logger.Logger,
) *SharingService {
	return &SharingService{
		equipmentShareRepo: eqRepo,
		shareRequestRepo:   reqRepo,
		peerRepo:           peerRepo,
		logger:             log,
	}
}

func (s *SharingService) ShareEquipment(ctx context.Context, cmd ShareEquipmentCommand) (*SharedEquipmentDTO, error) {
	if cmd.TenantID == "" || cmd.EquipmentID == "" || cmd.PeerID == "" {
		return nil, domain.ErrInvalidInput
	}

	peer, err := s.peerRepo.GetByID(ctx, cmd.TenantID, cmd.PeerID)
	if err != nil {
		return nil, err
	}
	if peer == nil {
		return nil, domain.ErrPeerNotFound
	}

	if peer.Status == domain.PeerStatusBlocked {
		return nil, domain.ErrPeerBlocked
	}

	share := &domain.SharedEquipment{
		ID:           fmt.Sprintf("share_%d", time.Now().UnixNano()),
		TenantID:     cmd.TenantID,
		EquipmentID:  cmd.EquipmentID,
		PeerID:       cmd.PeerID,
		Availability: cmd.Availability,
		PricePerDay:  cmd.PricePerDay,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.equipmentShareRepo.Create(ctx, share); err != nil {
		s.logger.Error("Failed to create equipment share", err)
		return nil, err
	}

	return SharedEquipmentToDTO(share), nil
}

func (s *SharingService) GetSharedEquipment(ctx context.Context, tenantID, id string) (*SharedEquipmentDTO, error) {
	if tenantID == "" || id == "" {
		return nil, domain.ErrInvalidInput
	}

	share, err := s.equipmentShareRepo.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	if share == nil {
		return nil, domain.ErrEquipmentNotFound
	}

	return SharedEquipmentToDTO(share), nil
}

func (s *SharingService) ListCatalog(ctx context.Context, tenantID string) ([]*SharedEquipmentDTO, error) {
	if tenantID == "" {
		return nil, domain.ErrTenantIDRequired
	}

	// In a real implementation, this would query across peers
	return make([]*SharedEquipmentDTO, 0), nil
}

func (s *SharingService) CreateShareRequest(ctx context.Context, cmd CreateShareRequestCommand) (*ShareRequestDTO, error) {
	if cmd.TenantID == "" || cmd.FromPeerID == "" || cmd.ToPeerID == "" || cmd.EquipmentID == "" {
		return nil, domain.ErrInvalidInput
	}

	req := &domain.ShareRequest{
		ID:          fmt.Sprintf("shareReq_%d", time.Now().UnixNano()),
		TenantID:    cmd.TenantID,
		FromPeerID:  cmd.FromPeerID,
		ToPeerID:    cmd.ToPeerID,
		EquipmentID: cmd.EquipmentID,
		StartDate:   cmd.StartDate,
		EndDate:     cmd.EndDate,
		Status:      domain.ShareRequestStatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.shareRequestRepo.Create(ctx, req); err != nil {
		s.logger.Error("Failed to create share request", err)
		return nil, err
	}

	return ShareRequestToDTO(req), nil
}

func (s *SharingService) GetShareRequest(ctx context.Context, tenantID, id string) (*ShareRequestDTO, error) {
	if tenantID == "" || id == "" {
		return nil, domain.ErrInvalidInput
	}

	req, err := s.shareRequestRepo.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	if req == nil {
		return nil, domain.ErrShareRequestNotFound
	}

	return ShareRequestToDTO(req), nil
}

func (s *SharingService) ListRequests(ctx context.Context, tenantID string) ([]*ShareRequestDTO, error) {
	if tenantID == "" {
		return nil, domain.ErrTenantIDRequired
	}

	reqs, err := s.shareRequestRepo.ListByTenant(ctx, tenantID)
	if err != nil {
		s.logger.Error("Failed to list share requests", err)
		return nil, err
	}

	dtos := make([]*ShareRequestDTO, len(reqs))
	for i, r := range reqs {
		dtos[i] = ShareRequestToDTO(r)
	}
	return dtos, nil
}

func (s *SharingService) ApproveRequest(ctx context.Context, tenantID, requestID string) (*ShareRequestDTO, error) {
	if tenantID == "" || requestID == "" {
		return nil, domain.ErrInvalidInput
	}

	req, err := s.shareRequestRepo.GetByID(ctx, tenantID, requestID)
	if err != nil {
		return nil, err
	}
	if req == nil {
		return nil, domain.ErrShareRequestNotFound
	}

	req.Status = domain.ShareRequestStatusApproved
	req.UpdatedAt = time.Now()

	if err := s.shareRequestRepo.Update(ctx, req); err != nil {
		s.logger.Error("Failed to approve share request", err)
		return nil, err
	}

	return ShareRequestToDTO(req), nil
}

func (s *SharingService) RejectRequest(ctx context.Context, tenantID, requestID string) (*ShareRequestDTO, error) {
	if tenantID == "" || requestID == "" {
		return nil, domain.ErrInvalidInput
	}

	req, err := s.shareRequestRepo.GetByID(ctx, tenantID, requestID)
	if err != nil {
		return nil, err
	}
	if req == nil {
		return nil, domain.ErrShareRequestNotFound
	}

	req.Status = domain.ShareRequestStatusRejected
	req.UpdatedAt = time.Now()

	if err := s.shareRequestRepo.Update(ctx, req); err != nil {
		s.logger.Error("Failed to reject share request", err)
		return nil, err
	}

	return ShareRequestToDTO(req), nil
}

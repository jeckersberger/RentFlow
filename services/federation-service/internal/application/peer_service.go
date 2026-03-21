package application

import (
	"context"
	"fmt"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/federation-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/federation-service/internal/ports"
)

type PeerService struct {
	repo   ports.PeerRepository
	logger logger.Logger
}

func NewPeerService(repo ports.PeerRepository, log logger.Logger) *PeerService {
	return &PeerService{
		repo:   repo,
		logger: log,
	}
}

func (s *PeerService) CreatePeer(ctx context.Context, cmd CreatePeerCommand) (*PeerDTO, error) {
	if cmd.TenantID == "" {
		return nil, domain.ErrTenantIDRequired
	}
	if cmd.Name == "" || cmd.URL == "" || cmd.PublicKey == "" {
		return nil, domain.ErrInvalidInput
	}

	existing, _ := s.repo.GetByURL(ctx, cmd.TenantID, cmd.URL)
	if existing != nil {
		return nil, fmt.Errorf("peer URL already exists")
	}

	peer := &domain.Peer{
		ID:        fmt.Sprintf("peer_%d", time.Now().UnixNano()),
		TenantID:  cmd.TenantID,
		Name:      cmd.Name,
		URL:       cmd.URL,
		PublicKey: cmd.PublicKey,
		Status:    domain.PeerStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.Create(ctx, peer); err != nil {
		s.logger.Error("Failed to create peer", err)
		return nil, err
	}

	return PeerToDTO(peer), nil
}

func (s *PeerService) GetPeer(ctx context.Context, tenantID, id string) (*PeerDTO, error) {
	if tenantID == "" || id == "" {
		return nil, domain.ErrInvalidInput
	}

	peer, err := s.repo.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	if peer == nil {
		return nil, domain.ErrPeerNotFound
	}

	return PeerToDTO(peer), nil
}

func (s *PeerService) ListPeers(ctx context.Context, tenantID string) ([]*PeerDTO, error) {
	if tenantID == "" {
		return nil, domain.ErrTenantIDRequired
	}

	peers, err := s.repo.ListByTenant(ctx, tenantID)
	if err != nil {
		s.logger.Error("Failed to list peers", err)
		return nil, err
	}

	dtos := make([]*PeerDTO, len(peers))
	for i, p := range peers {
		dtos[i] = PeerToDTO(p)
	}
	return dtos, nil
}

func (s *PeerService) UpdatePeer(ctx context.Context, cmd UpdatePeerCommand) (*PeerDTO, error) {
	if cmd.TenantID == "" || cmd.PeerID == "" {
		return nil, domain.ErrInvalidInput
	}

	peer, err := s.repo.GetByID(ctx, cmd.TenantID, cmd.PeerID)
	if err != nil {
		return nil, err
	}
	if peer == nil {
		return nil, domain.ErrPeerNotFound
	}

	if cmd.Name != "" {
		peer.Name = cmd.Name
	}
	if cmd.URL != "" {
		peer.URL = cmd.URL
	}
	if cmd.Status != "" {
		peer.Status = domain.PeerStatus(cmd.Status)
	}
	peer.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, peer); err != nil {
		s.logger.Error("Failed to update peer", err)
		return nil, err
	}

	return PeerToDTO(peer), nil
}

func (s *PeerService) DeletePeer(ctx context.Context, tenantID, id string) error {
	if tenantID == "" || id == "" {
		return domain.ErrInvalidInput
	}

	return s.repo.Delete(ctx, tenantID, id)
}

func (s *PeerService) RecordHandshake(ctx context.Context, tenantID, peerID string) error {
	if tenantID == "" || peerID == "" {
		return domain.ErrInvalidInput
	}

	peer, err := s.repo.GetByID(ctx, tenantID, peerID)
	if err != nil {
		return err
	}
	if peer == nil {
		return domain.ErrPeerNotFound
	}

	now := time.Now()
	peer.LastSeen = &now
	peer.Status = domain.PeerStatusActive

	return s.repo.Update(ctx, peer)
}

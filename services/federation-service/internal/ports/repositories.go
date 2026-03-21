package ports

import (
	"context"

	"github.com/jeckersberger/rentflow/services/federation-service/internal/domain"
)

type PeerRepository interface {
	Create(ctx context.Context, peer *domain.Peer) error
	GetByID(ctx context.Context, tenantID, id string) (*domain.Peer, error)
	ListByTenant(ctx context.Context, tenantID string) ([]*domain.Peer, error)
	Update(ctx context.Context, peer *domain.Peer) error
	Delete(ctx context.Context, tenantID, id string) error
	GetByURL(ctx context.Context, tenantID, url string) (*domain.Peer, error)
}

type EquipmentShareRepository interface {
	Create(ctx context.Context, share *domain.SharedEquipment) error
	GetByID(ctx context.Context, tenantID, id string) (*domain.SharedEquipment, error)
	ListByPeer(ctx context.Context, tenantID, peerID string) ([]*domain.SharedEquipment, error)
	ListByEquipment(ctx context.Context, tenantID, equipmentID string) ([]*domain.SharedEquipment, error)
	Update(ctx context.Context, share *domain.SharedEquipment) error
	Delete(ctx context.Context, tenantID, id string) error
}

type ShareRequestRepository interface {
	Create(ctx context.Context, req *domain.ShareRequest) error
	GetByID(ctx context.Context, tenantID, id string) (*domain.ShareRequest, error)
	ListByTenant(ctx context.Context, tenantID string) ([]*domain.ShareRequest, error)
	ListPending(ctx context.Context, tenantID, peerID string) ([]*domain.ShareRequest, error)
	Update(ctx context.Context, req *domain.ShareRequest) error
	Delete(ctx context.Context, tenantID, id string) error
}

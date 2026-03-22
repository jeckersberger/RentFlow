package repositories

import (
	"context"

	"github.com/jeckersberger/rentflow/pkg/common/database"
	"github.com/jeckersberger/rentflow/services/federation-service/internal/domain"
)

type PeerPostgres struct {
	db *database.PostgresPool
}

func NewPeerPostgres(db *database.PostgresPool) *PeerPostgres {
	return &PeerPostgres{db: db}
}

func (r *PeerPostgres) Create(ctx context.Context, peer *domain.Peer) error {
	query := `
		INSERT INTO federation_peers (id, tenant_id, name, url, public_key, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.Exec(ctx, query,
		peer.ID, peer.TenantID, peer.Name, peer.URL, peer.PublicKey, peer.Status, peer.CreatedAt, peer.UpdatedAt,
	)
	return err
}

func (r *PeerPostgres) GetByID(ctx context.Context, tenantID, id string) (*domain.Peer, error) {
	query := `
		SELECT id, tenant_id, name, url, public_key, status, last_seen, created_at, updated_at
		FROM federation_peers WHERE id = $1 AND tenant_id = $2
	`
	row := r.db.QueryRow(ctx, query, id, tenantID)

	peer := &domain.Peer{}
	err := row.Scan(&peer.ID, &peer.TenantID, &peer.Name, &peer.URL, &peer.PublicKey,
		&peer.Status, &peer.LastSeen, &peer.CreatedAt, &peer.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return peer, nil
}

func (r *PeerPostgres) ListByTenant(ctx context.Context, tenantID string) ([]*domain.Peer, error) {
	query := `
		SELECT id, tenant_id, name, url, public_key, status, last_seen, created_at, updated_at
		FROM federation_peers WHERE tenant_id = $1 ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var peers []*domain.Peer
	for rows.Next() {
		peer := &domain.Peer{}
		err := rows.Scan(&peer.ID, &peer.TenantID, &peer.Name, &peer.URL, &peer.PublicKey,
			&peer.Status, &peer.LastSeen, &peer.CreatedAt, &peer.UpdatedAt)
		if err != nil {
			return nil, err
		}
		peers = append(peers, peer)
	}

	return peers, rows.Err()
}

func (r *PeerPostgres) Update(ctx context.Context, peer *domain.Peer) error {
	query := `
		UPDATE federation_peers
		SET name = $1, url = $2, status = $3, last_seen = $4, updated_at = $5
		WHERE id = $6 AND tenant_id = $7
	`
	_, err := r.db.Exec(ctx, query,
		peer.Name, peer.URL, peer.Status, peer.LastSeen, peer.UpdatedAt, peer.ID, peer.TenantID,
	)
	return err
}

func (r *PeerPostgres) Delete(ctx context.Context, tenantID, id string) error {
	query := `DELETE FROM federation_peers WHERE id = $1 AND tenant_id = $2`
	_, err := r.db.Exec(ctx, query, id, tenantID)
	return err
}

func (r *PeerPostgres) GetByURL(ctx context.Context, tenantID, url string) (*domain.Peer, error) {
	query := `
		SELECT id, tenant_id, name, url, public_key, status, last_seen, created_at, updated_at
		FROM federation_peers WHERE url = $1 AND tenant_id = $2
	`
	row := r.db.QueryRow(ctx, query, url, tenantID)

	peer := &domain.Peer{}
	err := row.Scan(&peer.ID, &peer.TenantID, &peer.Name, &peer.URL, &peer.PublicKey,
		&peer.Status, &peer.LastSeen, &peer.CreatedAt, &peer.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return peer, nil
}

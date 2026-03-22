package repositories

import (
	"context"

	"github.com/jeckersberger/rentflow/pkg/common/database"
	"github.com/jeckersberger/rentflow/services/federation-service/internal/domain"
)

type ShareRequestPostgres struct {
	db *database.PostgresPool
}

func NewShareRequestPostgres(db *database.PostgresPool) *ShareRequestPostgres {
	return &ShareRequestPostgres{db: db}
}

func (r *ShareRequestPostgres) Create(ctx context.Context, req *domain.ShareRequest) error {
	query := `
		INSERT INTO federation_share_requests (id, tenant_id, from_peer_id, to_peer_id, equipment_id, start_date, end_date, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.db.Exec(ctx, query,
		req.ID, req.TenantID, req.FromPeerID, req.ToPeerID, req.EquipmentID, req.StartDate, req.EndDate, req.Status, req.CreatedAt, req.UpdatedAt,
	)
	return err
}

func (r *ShareRequestPostgres) GetByID(ctx context.Context, tenantID, id string) (*domain.ShareRequest, error) {
	query := `
		SELECT id, tenant_id, from_peer_id, to_peer_id, equipment_id, start_date, end_date, status, created_at, updated_at
		FROM federation_share_requests WHERE id = $1 AND tenant_id = $2
	`
	row := r.db.QueryRow(ctx, query, id, tenantID)

	req := &domain.ShareRequest{}
	err := row.Scan(&req.ID, &req.TenantID, &req.FromPeerID, &req.ToPeerID, &req.EquipmentID,
		&req.StartDate, &req.EndDate, &req.Status, &req.CreatedAt, &req.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return req, nil
}

func (r *ShareRequestPostgres) ListByTenant(ctx context.Context, tenantID string) ([]*domain.ShareRequest, error) {
	query := `
		SELECT id, tenant_id, from_peer_id, to_peer_id, equipment_id, start_date, end_date, status, created_at, updated_at
		FROM federation_share_requests WHERE tenant_id = $1 ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reqs []*domain.ShareRequest
	for rows.Next() {
		req := &domain.ShareRequest{}
		err := rows.Scan(&req.ID, &req.TenantID, &req.FromPeerID, &req.ToPeerID, &req.EquipmentID,
			&req.StartDate, &req.EndDate, &req.Status, &req.CreatedAt, &req.UpdatedAt)
		if err != nil {
			return nil, err
		}
		reqs = append(reqs, req)
	}

	return reqs, rows.Err()
}

func (r *ShareRequestPostgres) ListPending(ctx context.Context, tenantID, peerID string) ([]*domain.ShareRequest, error) {
	query := `
		SELECT id, tenant_id, from_peer_id, to_peer_id, equipment_id, start_date, end_date, status, created_at, updated_at
		FROM federation_share_requests
		WHERE tenant_id = $1 AND to_peer_id = $2 AND status = $3
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, tenantID, peerID, domain.ShareRequestStatusPending)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reqs []*domain.ShareRequest
	for rows.Next() {
		req := &domain.ShareRequest{}
		err := rows.Scan(&req.ID, &req.TenantID, &req.FromPeerID, &req.ToPeerID, &req.EquipmentID,
			&req.StartDate, &req.EndDate, &req.Status, &req.CreatedAt, &req.UpdatedAt)
		if err != nil {
			return nil, err
		}
		reqs = append(reqs, req)
	}

	return reqs, rows.Err()
}

func (r *ShareRequestPostgres) Update(ctx context.Context, req *domain.ShareRequest) error {
	query := `
		UPDATE federation_share_requests
		SET status = $1, updated_at = $2
		WHERE id = $3 AND tenant_id = $4
	`
	_, err := r.db.Exec(ctx, query, req.Status, req.UpdatedAt, req.ID, req.TenantID)
	return err
}

func (r *ShareRequestPostgres) Delete(ctx context.Context, tenantID, id string) error {
	query := `DELETE FROM federation_share_requests WHERE id = $1 AND tenant_id = $2`
	_, err := r.db.Exec(ctx, query, id, tenantID)
	return err
}

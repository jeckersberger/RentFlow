package repositories

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jeckersberger/rentflow/services/federation-service/internal/domain"
)

type EquipmentSharePostgres struct {
	db *pgxpool.Pool
}

func NewEquipmentSharePostgres(db *pgxpool.Pool) *EquipmentSharePostgres {
	return &EquipmentSharePostgres{db: db}
}

func (r *EquipmentSharePostgres) Create(ctx context.Context, share *domain.SharedEquipment) error {
	query := `
		INSERT INTO federation_equipment_shares (id, tenant_id, equipment_id, peer_id, availability, price_per_day, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.Exec(ctx, query,
		share.ID, share.TenantID, share.EquipmentID, share.PeerID, share.Availability, share.PricePerDay, share.CreatedAt, share.UpdatedAt,
	)
	return err
}

func (r *EquipmentSharePostgres) GetByID(ctx context.Context, tenantID, id string) (*domain.SharedEquipment, error) {
	query := `
		SELECT id, tenant_id, equipment_id, peer_id, availability, price_per_day, created_at, updated_at
		FROM federation_equipment_shares WHERE id = $1 AND tenant_id = $2
	`
	row := r.db.QueryRow(ctx, query, id, tenantID)

	share := &domain.SharedEquipment{}
	err := row.Scan(&share.ID, &share.TenantID, &share.EquipmentID, &share.PeerID,
		&share.Availability, &share.PricePerDay, &share.CreatedAt, &share.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return share, nil
}

func (r *EquipmentSharePostgres) ListByPeer(ctx context.Context, tenantID, peerID string) ([]*domain.SharedEquipment, error) {
	query := `
		SELECT id, tenant_id, equipment_id, peer_id, availability, price_per_day, created_at, updated_at
		FROM federation_equipment_shares WHERE tenant_id = $1 AND peer_id = $2 ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, tenantID, peerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var shares []*domain.SharedEquipment
	for rows.Next() {
		share := &domain.SharedEquipment{}
		err := rows.Scan(&share.ID, &share.TenantID, &share.EquipmentID, &share.PeerID,
			&share.Availability, &share.PricePerDay, &share.CreatedAt, &share.UpdatedAt)
		if err != nil {
			return nil, err
		}
		shares = append(shares, share)
	}

	return shares, rows.Err()
}

func (r *EquipmentSharePostgres) ListByEquipment(ctx context.Context, tenantID, equipmentID string) ([]*domain.SharedEquipment, error) {
	query := `
		SELECT id, tenant_id, equipment_id, peer_id, availability, price_per_day, created_at, updated_at
		FROM federation_equipment_shares WHERE tenant_id = $1 AND equipment_id = $2 ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, tenantID, equipmentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var shares []*domain.SharedEquipment
	for rows.Next() {
		share := &domain.SharedEquipment{}
		err := rows.Scan(&share.ID, &share.TenantID, &share.EquipmentID, &share.PeerID,
			&share.Availability, &share.PricePerDay, &share.CreatedAt, &share.UpdatedAt)
		if err != nil {
			return nil, err
		}
		shares = append(shares, share)
	}

	return shares, rows.Err()
}

func (r *EquipmentSharePostgres) Update(ctx context.Context, share *domain.SharedEquipment) error {
	query := `
		UPDATE federation_equipment_shares
		SET availability = $1, price_per_day = $2, updated_at = $3
		WHERE id = $4 AND tenant_id = $5
	`
	_, err := r.db.Exec(ctx, query, share.Availability, share.PricePerDay, share.UpdatedAt, share.ID, share.TenantID)
	return err
}

func (r *EquipmentSharePostgres) Delete(ctx context.Context, tenantID, id string) error {
	query := `DELETE FROM federation_equipment_shares WHERE id = $1 AND tenant_id = $2`
	_, err := r.db.Exec(ctx, query, id, tenantID)
	return err
}

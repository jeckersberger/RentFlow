package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jeckersberger/rentflow/pkg/common/database"
	"github.com/jeckersberger/rentflow/services/scanner-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/scanner-service/internal/ports"
)

type OfflineQueuePostgres struct {
	db *database.PostgresPool
}

func NewOfflineQueuePostgres(db *database.PostgresPool) ports.OfflineQueueRepository {
	return &OfflineQueuePostgres{db: db}
}

func (r *OfflineQueuePostgres) Create(ctx context.Context, item *domain.OfflineQueueItem) error {
	query := `
		INSERT INTO offline_queue
		(id, tenant_id, device_id, payload, created_at, sync_status)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.Exec(ctx, query,
		item.ID,
		item.TenantID,
		item.DeviceID,
		item.Payload,
		item.CreatedAt,
		item.SyncStatus,
	)

	return err
}

func (r *OfflineQueuePostgres) GetByID(ctx context.Context, id string) (*domain.OfflineQueueItem, error) {
	query := `
		SELECT id, tenant_id, device_id, payload, created_at, synced_at, sync_status
		FROM offline_queue
		WHERE id = $1
	`

	row := r.db.QueryRow(ctx, query, id)
	return queueRowToItem(row)
}

func (r *OfflineQueuePostgres) GetPending(ctx context.Context, tenantID string, limit int) ([]*domain.OfflineQueueItem, error) {
	query := `
		SELECT id, tenant_id, device_id, payload, created_at, synced_at, sync_status
		FROM offline_queue
		WHERE tenant_id = $1 AND sync_status IN ('pending', 'failed')
		ORDER BY created_at ASC
		LIMIT $2
	`

	rows, err := r.db.Query(ctx, query, tenantID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]*domain.OfflineQueueItem, 0)
	for rows.Next() {
		item, err := queueRowsToItem(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *OfflineQueuePostgres) Update(ctx context.Context, item *domain.OfflineQueueItem) error {
	query := `
		UPDATE offline_queue
		SET sync_status = $1, synced_at = $2
		WHERE id = $3
	`

	_, err := r.db.Exec(ctx, query,
		item.SyncStatus,
		item.SyncedAt,
		item.ID,
	)

	return err
}

func (r *OfflineQueuePostgres) DeleteByID(ctx context.Context, id string) error {
	query := `DELETE FROM offline_queue WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *OfflineQueuePostgres) GetCount(ctx context.Context, tenantID string, status string) (int, error) {
	query := `SELECT COUNT(*) FROM offline_queue WHERE tenant_id = $1 AND sync_status = $2`
	var count int
	err := r.db.QueryRow(ctx, query, tenantID, status).Scan(&count)
	return count, err
}

func queueRowToItem(row *sql.Row) (*domain.OfflineQueueItem, error) {
	item := &domain.OfflineQueueItem{}
	err := row.Scan(
		&item.ID,
		&item.TenantID,
		&item.DeviceID,
		&item.Payload,
		&item.CreatedAt,
		&item.SyncedAt,
		&item.SyncStatus,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("offline queue item not found")
		}
		return nil, err
	}
	return item, nil
}

func queueRowsToItem(rows *sql.Rows) (*domain.OfflineQueueItem, error) {
	item := &domain.OfflineQueueItem{}
	err := rows.Scan(
		&item.ID,
		&item.TenantID,
		&item.DeviceID,
		&item.Payload,
		&item.CreatedAt,
		&item.SyncedAt,
		&item.SyncStatus,
	)
	return item, err
}

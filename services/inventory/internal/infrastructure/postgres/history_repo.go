package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jeckersberger/EquipFlow/services/inventory/internal/domain"
)

// historyColumns lists all columns of the equipment_history table.
const historyColumns = `id, tenant_id, equipment_id, action, user_id, details, created_at`

// HistoryRepo implements domain.EquipmentHistoryRepository using PostgreSQL.
type HistoryRepo struct {
	pool *pgxpool.Pool
}

// NewHistoryRepo creates a new HistoryRepo.
func NewHistoryRepo(pool *pgxpool.Pool) *HistoryRepo {
	return &HistoryRepo{pool: pool}
}

// Record inserts a new equipment history entry.
func (r *HistoryRepo) Record(ctx context.Context, entry *domain.EquipmentHistory) error {
	query := `
		INSERT INTO equipment_history (
			id, tenant_id, equipment_id, action, user_id, details
		) VALUES (
			$1, $2, $3, $4, $5, $6
		) RETURNING created_at`

	err := r.pool.QueryRow(ctx, query,
		entry.ID, entry.TenantID, entry.EquipmentID,
		entry.Action, entry.UserID,
		nilIfEmptyJSON(entry.Details),
	).Scan(&entry.CreatedAt)
	if err != nil {
		return fmt.Errorf("history_repo: record: %w", err)
	}
	return nil
}

// ListByEquipment returns a paginated list of history entries for a specific equipment item.
func (r *HistoryRepo) ListByEquipment(ctx context.Context, equipmentID uuid.UUID, page int, perPage int) ([]*domain.EquipmentHistory, int64, error) {
	offset := (page - 1) * perPage

	var total int64
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM equipment_history WHERE equipment_id = $1`, equipmentID,
	).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("history_repo: list_by_equipment count: %w", err)
	}

	query := fmt.Sprintf(
		`SELECT %s FROM equipment_history WHERE equipment_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		historyColumns,
	)

	rows, err := r.pool.Query(ctx, query, equipmentID, perPage, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("history_repo: list_by_equipment query: %w", err)
	}
	defer rows.Close()

	var entries []*domain.EquipmentHistory
	for rows.Next() {
		h := &domain.EquipmentHistory{}
		var details []byte

		scanErr := rows.Scan(
			&h.ID, &h.TenantID, &h.EquipmentID,
			&h.Action, &h.UserID, &details,
			&h.CreatedAt,
		)
		if scanErr != nil {
			return nil, 0, fmt.Errorf("history_repo: list_by_equipment scan: %w", scanErr)
		}

		if details != nil {
			h.Details = json.RawMessage(details)
		}
		entries = append(entries, h)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("history_repo: list_by_equipment rows: %w", err)
	}
	return entries, total, nil
}

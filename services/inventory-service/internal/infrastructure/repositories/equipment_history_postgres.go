package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/jeckersberger/rentflow/pkg/common/database"
	"github.com/jeckersberger/rentflow/services/inventory-service/internal/domain"
)

type EquipmentHistoryPostgres struct {
	db *database.PostgresPool
}

func NewEquipmentHistoryPostgres(db *database.PostgresPool) *EquipmentHistoryPostgres {
	return &EquipmentHistoryPostgres{db: db}
}

func (r *EquipmentHistoryPostgres) Create(ctx context.Context, history *domain.EquipmentHistory) error {
	oldValueJSON, err := json.Marshal(history.OldValue)
	if err != nil {
		return fmt.Errorf("failed to marshal old_value: %w", err)
	}

	newValueJSON, err := json.Marshal(history.NewValue)
	if err != nil {
		return fmt.Errorf("failed to marshal new_value: %w", err)
	}

	query := `
		INSERT INTO inventory.equipment_history 
		(id, tenant_id, equipment_id, action, changed_by, old_value, new_value, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err = r.db.Exec(
		ctx,
		query,
		history.ID,
		history.TenantID,
		history.EquipmentID,
		history.Action,
		history.ChangedBy,
		string(oldValueJSON),
		string(newValueJSON),
		history.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create equipment history: %w", err)
	}

	return nil
}

func (r *EquipmentHistoryPostgres) GetByEquipmentID(
	ctx context.Context,
	tenantID, equipmentID string,
	limit, offset int,
) ([]*domain.EquipmentHistory, int64, error) {
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	// Get total count
	var total int64
	countQuery := `SELECT COUNT(*) FROM inventory.equipment_history WHERE tenant_id = $1 AND equipment_id = $2`
	err := r.db.QueryRow(ctx, countQuery, tenantID, equipmentID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count history: %w", err)
	}

	// Get paginated results
	query := `
		SELECT id, tenant_id, equipment_id, action, changed_by, old_value, new_value, created_at
		FROM inventory.equipment_history
		WHERE tenant_id = $1 AND equipment_id = $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.db.Query(ctx, query, tenantID, equipmentID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query equipment history: %w", err)
	}
	defer rows.Close()

	histories := make([]*domain.EquipmentHistory, 0)
	for rows.Next() {
		var h domain.EquipmentHistory
		var oldValueStr sql.NullString
		var newValueStr sql.NullString

		err := rows.Scan(
			&h.ID,
			&h.TenantID,
			&h.EquipmentID,
			&h.Action,
			&h.ChangedBy,
			&oldValueStr,
			&newValueStr,
			&h.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan history row: %w", err)
		}

		// Parse JSON
		if oldValueStr.Valid {
			err := json.Unmarshal([]byte(oldValueStr.String), &h.OldValue)
			if err != nil {
				return nil, 0, fmt.Errorf("failed to unmarshal old_value: %w", err)
			}
		}
		if newValueStr.Valid {
			err := json.Unmarshal([]byte(newValueStr.String), &h.NewValue)
			if err != nil {
				return nil, 0, fmt.Errorf("failed to unmarshal new_value: %w", err)
			}
		}

		histories = append(histories, &h)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating history rows: %w", err)
	}

	return histories, total, nil
}

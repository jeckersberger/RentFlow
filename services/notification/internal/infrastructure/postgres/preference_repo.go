package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jeckersberger/EquipFlow/services/notification/internal/domain"
)

const preferenceColumns = `id, tenant_id, user_id, channel, type, enabled, created_at, updated_at`

type PreferenceRepo struct {
	pool *pgxpool.Pool
}

func NewPreferenceRepo(pool *pgxpool.Pool) *PreferenceRepo {
	return &PreferenceRepo{pool: pool}
}

func (r *PreferenceRepo) List(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID) ([]*domain.NotificationPreference, error) {
	query := fmt.Sprintf(`SELECT %s FROM notification_preferences WHERE tenant_id = $1 AND user_id = $2 ORDER BY type, channel`, preferenceColumns)

	rows, err := r.pool.Query(ctx, query, tenantID, userID)
	if err != nil {
		return nil, fmt.Errorf("preference_repo: list: %w", err)
	}
	defer rows.Close()

	var items []*domain.NotificationPreference
	for rows.Next() {
		p := &domain.NotificationPreference{}
		if err := rows.Scan(
			&p.ID, &p.TenantID, &p.UserID, &p.Channel, &p.Type,
			&p.Enabled, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("preference_repo: list scan: %w", err)
		}
		items = append(items, p)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("preference_repo: list rows: %w", err)
	}
	return items, nil
}

func (r *PreferenceRepo) Upsert(ctx context.Context, pref *domain.NotificationPreference) error {
	query := `
		INSERT INTO notification_preferences (id, tenant_id, user_id, channel, type, enabled)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (tenant_id, user_id, channel, type)
		DO UPDATE SET enabled = EXCLUDED.enabled, updated_at = NOW()
		RETURNING id, created_at, updated_at`

	err := r.pool.QueryRow(ctx, query,
		pref.ID, pref.TenantID, pref.UserID, pref.Channel, pref.Type, pref.Enabled,
	).Scan(&pref.ID, &pref.CreatedAt, &pref.UpdatedAt)
	if err != nil {
		return fmt.Errorf("preference_repo: upsert: %w", err)
	}
	return nil
}

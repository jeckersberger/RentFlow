package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jeckersberger/EquipFlow/services/insurance/internal/domain"
)

const equipmentColumns = `id, policy_id, equipment_id, insured_value, added_at`

type EquipmentRepo struct {
	pool *pgxpool.Pool
}

func NewEquipmentRepo(pool *pgxpool.Pool) *EquipmentRepo {
	return &EquipmentRepo{pool: pool}
}

func (r *EquipmentRepo) Add(ctx context.Context, eq *domain.InsuredEquipment) error {
	query := `
		INSERT INTO insured_equipment (id, policy_id, equipment_id, insured_value)
		VALUES ($1, $2, $3, $4)
		RETURNING added_at`

	err := r.pool.QueryRow(ctx, query,
		eq.ID, eq.PolicyID, eq.EquipmentID, eq.InsuredValue,
	).Scan(&eq.AddedAt)
	if err != nil {
		return fmt.Errorf("equipment_repo: add: %w", err)
	}
	return nil
}

func (r *EquipmentRepo) ListByPolicy(ctx context.Context, policyID uuid.UUID, filter domain.EquipmentFilter) ([]*domain.InsuredEquipment, int64, error) {
	var total int64
	countQuery := `SELECT COUNT(*) FROM insured_equipment WHERE policy_id = $1`
	err := r.pool.QueryRow(ctx, countQuery, policyID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("equipment_repo: list count: %w", err)
	}

	page := filter.Page
	if page < 1 {
		page = 1
	}
	perPage := filter.PerPage
	if perPage < 1 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	dataQuery := fmt.Sprintf(
		`SELECT %s FROM insured_equipment WHERE policy_id = $1 ORDER BY added_at DESC LIMIT $2 OFFSET $3`,
		equipmentColumns,
	)

	rows, err := r.pool.Query(ctx, dataQuery, policyID, perPage, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("equipment_repo: list query: %w", err)
	}
	defer rows.Close()

	var items []*domain.InsuredEquipment
	for rows.Next() {
		eq := &domain.InsuredEquipment{}
		if scanErr := rows.Scan(&eq.ID, &eq.PolicyID, &eq.EquipmentID, &eq.InsuredValue, &eq.AddedAt); scanErr != nil {
			return nil, 0, fmt.Errorf("equipment_repo: list scan: %w", scanErr)
		}
		items = append(items, eq)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("equipment_repo: list rows: %w", err)
	}
	return items, total, nil
}

package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/jeckersberger/rentflow/pkg/common/database"
	"github.com/jeckersberger/rentflow/services/warehouse-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/warehouse-service/internal/ports"
)

type InventoryCheckPostgres struct {
	db *database.PostgresPool
}

func NewInventoryCheckPostgres(db *database.PostgresPool) ports.InventoryCheckRepository {
	return &InventoryCheckPostgres{db: db}
}

func (r *InventoryCheckPostgres) Create(ctx context.Context, check *domain.InventoryCheck) error {
	itemsJSON, err := json.Marshal(check.Items)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO inventory_checks
		(id, tenant_id, name, location_id, status, items, started_at, completed_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err = r.db.Exec(ctx, query,
		check.ID,
		check.TenantID,
		check.Name,
		check.ZoneID,
		string(check.Status),
		itemsJSON,
		check.StartedAt,
		check.CompletedAt,
		check.CreatedAt,
		check.UpdatedAt,
	)

	return err
}

func (r *InventoryCheckPostgres) GetByID(ctx context.Context, tenantID, id string) (*domain.InventoryCheck, error) {
	query := `
		SELECT id, tenant_id, name, location_id, status, items, started_at, completed_at, created_at, updated_at
		FROM inventory_checks
		WHERE tenant_id = $1 AND id = $2
	`

	row := r.db.QueryRow(ctx, query, tenantID, id)
	return inventoryCheckRowToCheck(row)
}

func (r *InventoryCheckPostgres) List(ctx context.Context, tenantID string, limit, offset int) ([]*domain.InventoryCheck, int, error) {
	query := `
		SELECT id, tenant_id, name, location_id, status, items, started_at, completed_at, created_at, updated_at
		FROM inventory_checks
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	checks := make([]*domain.InventoryCheck, 0)
	for rows.Next() {
		check, err := inventoryCheckRowsToCheck(rows)
		if err != nil {
			return nil, 0, err
		}
		checks = append(checks, check)
	}

	// Get total count
	countQuery := "SELECT COUNT(*) FROM inventory_checks WHERE tenant_id = $1"
	var total int
	if err := r.db.QueryRow(ctx, countQuery, tenantID).Scan(&total); err != nil {
		return nil, 0, err
	}

	return checks, total, rows.Err()
}

func (r *InventoryCheckPostgres) Update(ctx context.Context, check *domain.InventoryCheck) error {
	itemsJSON, err := json.Marshal(check.Items)
	if err != nil {
		return err
	}

	query := `
		UPDATE inventory_checks
		SET status = $1, items = $2, started_at = $3, completed_at = $4, updated_at = $5
		WHERE id = $6 AND tenant_id = $7
	`

	_, err = r.db.Exec(ctx, query,
		string(check.Status),
		itemsJSON,
		check.StartedAt,
		check.CompletedAt,
		check.UpdatedAt,
		check.ID,
		check.TenantID,
	)

	return err
}

func inventoryCheckRowToCheck(row *sql.Row) (*domain.InventoryCheck, error) {
	check := &domain.InventoryCheck{}
	var itemsJSON []byte

	err := row.Scan(
		&check.ID,
		&check.TenantID,
		&check.Name,
		&check.ZoneID,
		(*string)(&check.Status),
		&itemsJSON,
		&check.StartedAt,
		&check.CompletedAt,
		&check.CreatedAt,
		&check.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("inventory check not found")
		}
		return nil, err
	}

	if err := json.Unmarshal(itemsJSON, &check.Items); err != nil {
		return nil, err
	}

	return check, nil
}

func inventoryCheckRowsToCheck(rows *sql.Rows) (*domain.InventoryCheck, error) {
	check := &domain.InventoryCheck{}
	var itemsJSON []byte

	err := rows.Scan(
		&check.ID,
		&check.TenantID,
		&check.Name,
		&check.ZoneID,
		(*string)(&check.Status),
		&itemsJSON,
		&check.StartedAt,
		&check.CompletedAt,
		&check.CreatedAt,
		&check.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(itemsJSON, &check.Items); err != nil {
		return nil, err
	}

	return check, nil
}

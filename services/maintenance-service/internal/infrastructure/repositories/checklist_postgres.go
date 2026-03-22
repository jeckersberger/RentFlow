package repositories

import (
	"context"
	"encoding/json"

	"github.com/jeckersberger/rentflow/pkg/common/database"
	"github.com/jeckersberger/rentflow/services/maintenance-service/internal/domain"
)

type ChecklistPostgres struct {
	db *database.PostgresPool
}

func NewChecklistPostgres(db *database.PostgresPool) *ChecklistPostgres {
	return &ChecklistPostgres{db: db}
}

func (r *ChecklistPostgres) Create(ctx context.Context, checklist *domain.Checklist) error {
	itemsJSON, err := json.Marshal(checklist.Items)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO checklists (id, tenant_id, name, description, items, version, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err = r.db.Exec(ctx, query,
		checklist.ID, checklist.TenantID, checklist.Name, checklist.Description, string(itemsJSON), checklist.Version, checklist.CreatedAt,
	)
	return err
}

func (r *ChecklistPostgres) GetByID(ctx context.Context, tenantID, id string) (*domain.Checklist, error) {
	query := `
		SELECT id, tenant_id, name, description, items, version, created_at
		FROM checklists WHERE id = $1 AND tenant_id = $2
	`
	row := r.db.QueryRow(ctx, query, id, tenantID)

	checklist := &domain.Checklist{}
	var itemsJSON string
	err := row.Scan(&checklist.ID, &checklist.TenantID, &checklist.Name, &checklist.Description, &itemsJSON, &checklist.Version, &checklist.CreatedAt)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal([]byte(itemsJSON), &checklist.Items); err != nil {
		return nil, err
	}

	return checklist, nil
}

func (r *ChecklistPostgres) ListByTenant(ctx context.Context, tenantID string) ([]*domain.Checklist, error) {
	query := `
		SELECT id, tenant_id, name, description, items, version, created_at
		FROM checklists WHERE tenant_id = $1 ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var checklists []*domain.Checklist
	for rows.Next() {
		checklist := &domain.Checklist{}
		var itemsJSON string
		err := rows.Scan(&checklist.ID, &checklist.TenantID, &checklist.Name, &checklist.Description, &itemsJSON, &checklist.Version, &checklist.CreatedAt)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal([]byte(itemsJSON), &checklist.Items); err != nil {
			return nil, err
		}

		checklists = append(checklists, checklist)
	}

	return checklists, rows.Err()
}

func (r *ChecklistPostgres) Update(ctx context.Context, checklist *domain.Checklist) error {
	itemsJSON, err := json.Marshal(checklist.Items)
	if err != nil {
		return err
	}

	query := `
		UPDATE checklists
		SET name = $1, description = $2, items = $3, version = $4
		WHERE id = $5 AND tenant_id = $6
	`
	_, err = r.db.Exec(ctx, query,
		checklist.Name, checklist.Description, string(itemsJSON), checklist.Version, checklist.ID, checklist.TenantID,
	)
	return err
}

func (r *ChecklistPostgres) Delete(ctx context.Context, tenantID, id string) error {
	query := `DELETE FROM checklists WHERE id = $1 AND tenant_id = $2`
	_, err := r.db.Exec(ctx, query, id, tenantID)
	return err
}

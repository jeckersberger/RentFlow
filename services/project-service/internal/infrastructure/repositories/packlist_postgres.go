package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jeckersberger/rentflow/pkg/common/database"
	"github.com/jeckersberger/rentflow/services/project-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/project-service/internal/ports"
)

type PacklistPostgres struct {
	db *database.PostgresPool
}

func NewPacklistPostgres(db *database.PostgresPool) *PacklistPostgres {
	return &PacklistPostgres{db: db}
}

func (r *PacklistPostgres) Create(ctx context.Context, p *domain.Packlist) error {
	return r.db.WithTx(ctx, func(tx *sql.Tx) error {
		// Insert packlist
		query := `
			INSERT INTO projects.packlists (
				id, project_id, tenant_id, name, status, created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7)
		`

		_, err := tx.ExecContext(ctx, query,
			p.ID, p.ProjectID, p.TenantID, p.Name, string(p.Status), p.CreatedAt, p.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to create packlist: %w", err)
		}

		// Insert items
		for _, item := range p.Items {
			itemQuery := `
				INSERT INTO projects.packlist_items (
					id, packlist_id, equipment_id, equipment_name, quantity,
					quantity_packed, quantity_returned, notes, storage_location, status,
					created_at, updated_at
				) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
			`
			_, err = tx.ExecContext(ctx, itemQuery,
				item.ID, item.PacklistID, item.EquipmentID, item.EquipmentName,
				item.Quantity, item.QuantityPacked, item.QuantityReturned,
				item.Notes, item.StorageLocation, string(item.Status), item.CreatedAt, item.UpdatedAt,
			)
			if err != nil {
				return fmt.Errorf("failed to insert packlist item: %w", err)
			}
		}

		return nil
	})
}

func (r *PacklistPostgres) Update(ctx context.Context, p *domain.Packlist) error {
	return r.db.WithTx(ctx, func(tx *sql.Tx) error {
		// Update packlist
		query := `
			UPDATE projects.packlists SET
				name = $3, status = $4, updated_at = $5
			WHERE id = $1 AND tenant_id = $2
		`

		result, err := tx.ExecContext(ctx, query,
			p.ID, p.TenantID, p.Name, string(p.Status), p.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to update packlist: %w", err)
		}

		rows, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get rows affected: %w", err)
		}

		if rows == 0 {
			return sql.ErrNoRows
		}

		// Delete all existing items
		_, err = tx.ExecContext(ctx, `DELETE FROM projects.packlist_items WHERE packlist_id = $1`, p.ID)
		if err != nil {
			return fmt.Errorf("failed to delete items: %w", err)
		}

		// Insert updated items
		for _, item := range p.Items {
			itemQuery := `
				INSERT INTO projects.packlist_items (
					id, packlist_id, equipment_id, equipment_name, quantity,
					quantity_packed, quantity_returned, notes, storage_location, status,
					created_at, updated_at
				) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
			`
			_, err = tx.ExecContext(ctx, itemQuery,
				item.ID, item.PacklistID, item.EquipmentID, item.EquipmentName,
				item.Quantity, item.QuantityPacked, item.QuantityReturned,
				item.Notes, item.StorageLocation, string(item.Status), item.CreatedAt, item.UpdatedAt,
			)
			if err != nil {
				return fmt.Errorf("failed to insert packlist item: %w", err)
			}
		}

		return nil
	})
}

func (r *PacklistPostgres) GetByID(ctx context.Context, tenantID, packlistID string) (*domain.Packlist, error) {
	query := `
		SELECT id, project_id, tenant_id, name, status, created_at, updated_at
		FROM projects.packlists
		WHERE id = $1 AND tenant_id = $2
	`

	p := &domain.Packlist{}
	err := r.db.QueryRow(ctx, query, packlistID, tenantID).Scan(
		&p.ID, &p.ProjectID, &p.TenantID, &p.Name, &p.Status, &p.CreatedAt, &p.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("failed to get packlist: %w", err)
	}

	// Get items
	itemQuery := `
		SELECT id, packlist_id, equipment_id, equipment_name, quantity,
			   quantity_packed, quantity_returned, notes, storage_location, status,
			   created_at, updated_at
		FROM projects.packlist_items
		WHERE packlist_id = $1
		ORDER BY storage_location ASC, created_at ASC
	`

	rows, err := r.db.Query(ctx, itemQuery, packlistID)
	if err != nil {
		return nil, fmt.Errorf("failed to get packlist items: %w", err)
	}
	defer rows.Close()

	items := make([]domain.PacklistItem, 0)
	for rows.Next() {
		var item domain.PacklistItem
		err := rows.Scan(
			&item.ID, &item.PacklistID, &item.EquipmentID, &item.EquipmentName,
			&item.Quantity, &item.QuantityPacked, &item.QuantityReturned,
			&item.Notes, &item.StorageLocation, &item.Status, &item.CreatedAt, &item.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan item: %w", err)
		}
		items = append(items, item)
	}

	p.Items = items
	return p, nil
}

func (r *PacklistPostgres) ListByProjectID(ctx context.Context, tenantID, projectID string, limit, offset int) (*ports.PacklistListResult, error) {
	query := `
		SELECT id, project_id, tenant_id, name, status, created_at, updated_at
		FROM projects.packlists
		WHERE tenant_id = $1 AND project_id = $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`

	countQuery := `
		SELECT COUNT(*) FROM projects.packlists
		WHERE tenant_id = $1 AND project_id = $2
	`

	var count int64
	err := r.db.QueryRow(ctx, countQuery, tenantID, projectID).Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("failed to count packlists: %w", err)
	}

	rows, err := r.db.Query(ctx, query, tenantID, projectID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list packlists: %w", err)
	}
	defer rows.Close()

	packlists := make([]*domain.Packlist, 0)
	for rows.Next() {
		p := &domain.Packlist{}
		err := rows.Scan(
			&p.ID, &p.ProjectID, &p.TenantID, &p.Name, &p.Status, &p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan packlist: %w", err)
		}

		// Get items for this packlist
		itemQuery := `
			SELECT id, packlist_id, equipment_id, equipment_name, quantity,
				   quantity_packed, quantity_returned, notes, storage_location, status,
				   created_at, updated_at
			FROM projects.packlist_items
			WHERE packlist_id = $1
			ORDER BY storage_location ASC, created_at ASC
		`

		itemRows, err := r.db.Query(ctx, itemQuery, p.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get packlist items: %w", err)
		}
		defer itemRows.Close()

		items := make([]domain.PacklistItem, 0)
		for itemRows.Next() {
			var item domain.PacklistItem
			err := itemRows.Scan(
				&item.ID, &item.PacklistID, &item.EquipmentID, &item.EquipmentName,
				&item.Quantity, &item.QuantityPacked, &item.QuantityReturned,
				&item.Notes, &item.StorageLocation, &item.Status, &item.CreatedAt, &item.UpdatedAt,
			)
			if err != nil {
				return nil, fmt.Errorf("failed to scan item: %w", err)
			}
			items = append(items, item)
		}

		p.Items = items
		packlists = append(packlists, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating packlists: %w", err)
	}

	return &ports.PacklistListResult{
		Items:  packlists,
		Total:  count,
		Limit:  limit,
		Offset: offset,
	}, nil
}

func (r *PacklistPostgres) Delete(ctx context.Context, tenantID, packlistID string) error {
	return r.db.WithTx(ctx, func(tx *sql.Tx) error {
		// Delete items first
		_, err := tx.ExecContext(ctx, `DELETE FROM projects.packlist_items WHERE packlist_id = $1`, packlistID)
		if err != nil {
			return fmt.Errorf("failed to delete items: %w", err)
		}

		// Delete packlist
		result, err := tx.ExecContext(ctx, `DELETE FROM projects.packlists WHERE id = $1 AND tenant_id = $2`, packlistID, tenantID)
		if err != nil {
			return fmt.Errorf("failed to delete packlist: %w", err)
		}

		rows, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get rows affected: %w", err)
		}

		if rows == 0 {
			return sql.ErrNoRows
		}

		return nil
	})
}

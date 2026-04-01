package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jeckersberger/EquipFlow/services/project/internal/domain"

	apperrors "github.com/jeckersberger/EquipFlow/pkg/common/errors"
)

// packlistColumns lists all columns of the packlists table.
const packlistColumns = `id, project_id, tenant_id, name, status, created_by, created_at`

// packlistItemColumns lists all columns of the packlist_items table.
const packlistItemColumns = `
	id, packlist_id, equipment_id,
	quantity_planned, quantity_packed, quantity_returned,
	status, damaged,
	packed_by, packed_at, notes`

// PacklistRepo implements domain.PacklistRepository using PostgreSQL.
type PacklistRepo struct {
	pool *pgxpool.Pool
}

// NewPacklistRepo creates a new PacklistRepo.
func NewPacklistRepo(pool *pgxpool.Pool) *PacklistRepo {
	return &PacklistRepo{pool: pool}
}

// scanPacklist scans a single packlist row into a domain.Packlist.
func scanPacklist(row pgx.Row) (*domain.Packlist, error) {
	pl := &domain.Packlist{}
	var (
		status    *string
		createdBy *uuid.UUID
	)

	err := row.Scan(
		&pl.ID, &pl.ProjectID, &pl.TenantID, &pl.Name,
		&status, &createdBy, &pl.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	pl.Status = derefString(status)
	pl.CreatedBy = createdBy

	return pl, nil
}

// scanPacklistItem scans a single packlist_item row into a domain.PacklistItem.
func scanPacklistItem(row pgx.Row) (*domain.PacklistItem, error) {
	item := &domain.PacklistItem{}
	var (
		status   *string
		packedBy *uuid.UUID
		packedAt *time.Time
		notes    *string
	)

	err := row.Scan(
		&item.ID, &item.PacklistID, &item.EquipmentID,
		&item.QuantityPlanned, &item.QuantityPacked, &item.QuantityReturned,
		&status, &item.Damaged,
		&packedBy, &packedAt, &notes,
	)
	if err != nil {
		return nil, err
	}

	item.Status = derefString(status)
	if item.Status == "" {
		item.Status = domain.PacklistItemStatusPlanned
	}
	item.PackedBy = packedBy
	item.PackedAt = packedAt
	item.Notes = derefString(notes)

	return item, nil
}

// Create inserts a new packlist and scans back the generated fields.
func (r *PacklistRepo) Create(ctx context.Context, packlist *domain.Packlist) error {
	query := `
		INSERT INTO packlists (
			id, project_id, tenant_id, name, status, created_by
		) VALUES (
			$1, $2, $3, $4, $5, $6
		) RETURNING created_at`

	err := r.pool.QueryRow(ctx, query,
		packlist.ID, packlist.ProjectID, packlist.TenantID,
		packlist.Name, nilIfEmpty(packlist.Status), packlist.CreatedBy,
	).Scan(&packlist.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return apperrors.ErrConflict
		}
		return fmt.Errorf("packlist_repo: create: %w", err)
	}
	return nil
}

// GetByID retrieves a packlist by primary key scoped to a tenant.
func (r *PacklistRepo) GetByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*domain.Packlist, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM packlists WHERE id = $1 AND tenant_id = $2`,
		packlistColumns,
	)
	pl, err := scanPacklist(r.pool.QueryRow(ctx, query, id, tenantID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("packlist_repo: get_by_id: %w", err)
	}
	return pl, nil
}

// ListByProject returns all packlists for a given project, ordered by created_at.
func (r *PacklistRepo) ListByProject(ctx context.Context, projectID, tenantID uuid.UUID) ([]*domain.Packlist, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM packlists WHERE project_id = $1 AND tenant_id = $2 ORDER BY created_at ASC`,
		packlistColumns,
	)

	rows, err := r.pool.Query(ctx, query, projectID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("packlist_repo: list_by_project query: %w", err)
	}
	defer rows.Close()

	var items []*domain.Packlist
	for rows.Next() {
		pl, scanErr := scanPacklist(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("packlist_repo: list_by_project scan: %w", scanErr)
		}
		items = append(items, pl)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("packlist_repo: list_by_project rows: %w", err)
	}
	return items, nil
}

// List returns all packlists for a tenant, ordered by created_at descending.
func (r *PacklistRepo) List(ctx context.Context, tenantID uuid.UUID) ([]*domain.Packlist, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM packlists WHERE tenant_id = $1 ORDER BY created_at DESC`,
		packlistColumns,
	)

	rows, err := r.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("packlist_repo: list query: %w", err)
	}
	defer rows.Close()

	var items []*domain.Packlist
	for rows.Next() {
		pl, scanErr := scanPacklist(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("packlist_repo: list scan: %w", scanErr)
		}
		items = append(items, pl)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("packlist_repo: list rows: %w", err)
	}
	return items, nil
}

// UpdateStatus changes the status of a packlist.
func (r *PacklistRepo) UpdateStatus(ctx context.Context, id uuid.UUID, tenantID uuid.UUID, status string) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE packlists SET status = $3 WHERE id = $1 AND tenant_id = $2`,
		id, tenantID, status,
	)
	if err != nil {
		return fmt.Errorf("packlist_repo: update_status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

// AddItem inserts a new item into a packlist.
func (r *PacklistRepo) AddItem(ctx context.Context, item *domain.PacklistItem) error {
	query := `
		INSERT INTO packlist_items (
			id, packlist_id, equipment_id,
			quantity_planned, quantity_packed, quantity_returned,
			status, damaged,
			packed_by, packed_at, notes
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
		) RETURNING id`

	status := item.Status
	if status == "" {
		status = domain.PacklistItemStatusPlanned
	}

	err := r.pool.QueryRow(ctx, query,
		item.ID, item.PacklistID, item.EquipmentID,
		item.QuantityPlanned, item.QuantityPacked, item.QuantityReturned,
		status, item.Damaged,
		item.PackedBy, item.PackedAt, nilIfEmpty(item.Notes),
	).Scan(&item.ID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return apperrors.ErrConflict
		}
		return fmt.Errorf("packlist_repo: add_item: %w", err)
	}
	return nil
}

// GetItems returns all items for a given packlist, ordered by equipment_id.
// Tenant isolation: verifies the packlist belongs to the given tenant via JOIN.
func (r *PacklistRepo) GetItems(ctx context.Context, packlistID, tenantID uuid.UUID) ([]*domain.PacklistItem, error) {
	query := fmt.Sprintf(
		`SELECT pi.%s FROM packlist_items pi
		 JOIN packlists pl ON pl.id = pi.packlist_id
		 WHERE pi.packlist_id = $1 AND pl.tenant_id = $2
		 ORDER BY pi.equipment_id ASC`,
		"id, packlist_id, equipment_id, quantity_planned, quantity_packed, quantity_returned, status, damaged, packed_by, packed_at, notes",
	)

	rows, err := r.pool.Query(ctx, query, packlistID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("packlist_repo: get_items query: %w", err)
	}
	defer rows.Close()

	var items []*domain.PacklistItem
	for rows.Next() {
		item, scanErr := scanPacklistItem(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("packlist_repo: get_items scan: %w", scanErr)
		}
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("packlist_repo: get_items rows: %w", err)
	}
	return items, nil
}

// GetItemByID retrieves a single packlist item by ID, scoped to tenant.
func (r *PacklistRepo) GetItemByID(ctx context.Context, itemID, tenantID uuid.UUID) (*domain.PacklistItem, error) {
	query := `SELECT pi.id, pi.packlist_id, pi.equipment_id,
		pi.quantity_planned, pi.quantity_packed, pi.quantity_returned,
		pi.status, pi.damaged,
		pi.packed_by, pi.packed_at, pi.notes
		FROM packlist_items pi
		JOIN packlists pl ON pl.id = pi.packlist_id
		WHERE pi.id = $1 AND pl.tenant_id = $2`

	item, err := scanPacklistItem(r.pool.QueryRow(ctx, query, itemID, tenantID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("packlist_repo: get_item_by_id: %w", err)
	}
	return item, nil
}

// UpdateItemStatus changes the status and damaged flag of a single packlist item.
// Tenant isolation: only updates if the item's packlist belongs to the given tenant.
func (r *PacklistRepo) UpdateItemStatus(ctx context.Context, itemID, tenantID uuid.UUID, status string, damaged bool) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE packlist_items SET status = $2, damaged = $3
		 WHERE id = $1
		   AND packlist_id IN (SELECT id FROM packlists WHERE tenant_id = $4)`,
		itemID, status, damaged, tenantID,
	)
	if err != nil {
		return fmt.Errorf("packlist_repo: update_item_status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

// BulkUpdateItemStatus changes the status and damaged flag for multiple packlist items.
// Returns the number of rows affected. Tenant isolation via sub-select.
func (r *PacklistRepo) BulkUpdateItemStatus(ctx context.Context, itemIDs []uuid.UUID, tenantID uuid.UUID, status string, damaged bool) (int64, error) {
	if len(itemIDs) == 0 {
		return 0, nil
	}

	tag, err := r.pool.Exec(ctx,
		`UPDATE packlist_items SET status = $2, damaged = $3
		 WHERE id = ANY($1)
		   AND packlist_id IN (SELECT id FROM packlists WHERE tenant_id = $4)`,
		itemIDs, status, damaged, tenantID,
	)
	if err != nil {
		return 0, fmt.Errorf("packlist_repo: bulk_update_item_status: %w", err)
	}
	return tag.RowsAffected(), nil
}

// GetSummary returns aggregated status counts for all items in a packlist.
func (r *PacklistRepo) GetSummary(ctx context.Context, packlistID, tenantID uuid.UUID) (*domain.PacklistSummary, error) {
	query := `
		SELECT
			COUNT(*)::int AS total,
			COUNT(*) FILTER (WHERE pi.status = 'planned' OR pi.status IS NULL)::int AS planned,
			COUNT(*) FILTER (WHERE pi.status = 'packed')::int AS packed,
			COUNT(*) FILTER (WHERE pi.status = 'loaded')::int AS loaded,
			COUNT(*) FILTER (WHERE pi.status = 'on_site')::int AS on_site,
			COUNT(*) FILTER (WHERE pi.status = 'returned')::int AS returned,
			COUNT(*) FILTER (WHERE pi.damaged = true)::int AS damaged
		FROM packlist_items pi
		JOIN packlists pl ON pl.id = pi.packlist_id
		WHERE pi.packlist_id = $1 AND pl.tenant_id = $2`

	summary := &domain.PacklistSummary{}
	err := r.pool.QueryRow(ctx, query, packlistID, tenantID).Scan(
		&summary.Total,
		&summary.Planned,
		&summary.Packed,
		&summary.Loaded,
		&summary.OnSite,
		&summary.Returned,
		&summary.Damaged,
	)
	if err != nil {
		return nil, fmt.Errorf("packlist_repo: get_summary: %w", err)
	}
	return summary, nil
}

// UpdateItemPacked updates the packed quantity and packed-by info for a packlist item.
// Tenant isolation: only updates if the item's packlist belongs to the given tenant.
func (r *PacklistRepo) UpdateItemPacked(ctx context.Context, itemID uuid.UUID, quantityPacked int, packedBy *uuid.UUID, tenantID uuid.UUID) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE packlist_items
		 SET quantity_packed = $2, packed_by = $3, packed_at = NOW()
		 WHERE id = $1
		   AND packlist_id IN (SELECT id FROM packlists WHERE tenant_id = $4)`,
		itemID, quantityPacked, packedBy, tenantID,
	)
	if err != nil {
		return fmt.Errorf("packlist_repo: update_item_packed: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

// UpdateItemReturned updates the returned quantity for a packlist item.
// Tenant isolation: only updates if the item's packlist belongs to the given tenant.
func (r *PacklistRepo) UpdateItemReturned(ctx context.Context, itemID uuid.UUID, quantityReturned int, tenantID uuid.UUID) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE packlist_items SET quantity_returned = $2
		 WHERE id = $1
		   AND packlist_id IN (SELECT id FROM packlists WHERE tenant_id = $3)`,
		itemID, quantityReturned, tenantID,
	)
	if err != nil {
		return fmt.Errorf("packlist_repo: update_item_returned: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

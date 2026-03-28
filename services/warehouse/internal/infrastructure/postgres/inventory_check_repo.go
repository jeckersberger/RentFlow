package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jeckersberger/EquipFlow/services/warehouse/internal/domain"

	apperrors "github.com/jeckersberger/EquipFlow/pkg/common/errors"
)

// checkColumns lists all columns of the inventory_checks table.
const checkColumns = `id, tenant_id, zone_id, status, started_by, started_at, completed_at, expected_count, actual_count, discrepancy_count, notes`

// checkItemColumns lists all columns of the inventory_check_items table.
const checkItemColumns = `id, check_id, equipment_id, expected, found, scanned_at, notes`

// InventoryCheckRepo implements domain.InventoryCheckRepository using PostgreSQL.
type InventoryCheckRepo struct {
	pool *pgxpool.Pool
}

// NewInventoryCheckRepo creates a new InventoryCheckRepo.
func NewInventoryCheckRepo(pool *pgxpool.Pool) *InventoryCheckRepo {
	return &InventoryCheckRepo{pool: pool}
}

// scanCheck scans a single inventory_check row into a domain.InventoryCheck.
func scanCheck(row pgx.Row) (*domain.InventoryCheck, error) {
	c := &domain.InventoryCheck{}
	var notes *string

	err := row.Scan(
		&c.ID, &c.TenantID, &c.ZoneID, &c.Status,
		&c.StartedBy, &c.StartedAt, &c.CompletedAt,
		&c.ExpectedCount, &c.ActualCount, &c.DiscrepancyCount,
		&notes,
	)
	if err != nil {
		return nil, err
	}

	if notes != nil {
		c.Notes = *notes
	}
	return c, nil
}

// scanCheckItem scans a single inventory_check_item row into a domain.InventoryCheckItem.
func scanCheckItem(row pgx.Row) (*domain.InventoryCheckItem, error) {
	i := &domain.InventoryCheckItem{}
	var notes *string

	err := row.Scan(
		&i.ID, &i.CheckID, &i.EquipmentID,
		&i.Expected, &i.Found, &i.ScannedAt,
		&notes,
	)
	if err != nil {
		return nil, err
	}

	if notes != nil {
		i.Notes = *notes
	}
	return i, nil
}

// Create inserts a new inventory check and scans back the generated fields.
func (r *InventoryCheckRepo) Create(ctx context.Context, check *domain.InventoryCheck) error {
	query := `
		INSERT INTO inventory_checks (
			id, tenant_id, zone_id, status, started_by, started_at, notes
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		) RETURNING id, started_at`

	err := r.pool.QueryRow(ctx, query,
		check.ID, check.TenantID, check.ZoneID,
		check.Status, check.StartedBy, check.StartedAt,
		nilIfEmpty(check.Notes),
	).Scan(&check.ID, &check.StartedAt)
	if err != nil {
		return fmt.Errorf("inventory_check_repo: create: %w", err)
	}
	return nil
}

// GetByID retrieves an inventory check by primary key scoped to a tenant.
func (r *InventoryCheckRepo) GetByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*domain.InventoryCheck, error) {
	query := fmt.Sprintf(`SELECT %s FROM inventory_checks WHERE id = $1 AND tenant_id = $2`, checkColumns)
	c, err := scanCheck(r.pool.QueryRow(ctx, query, id, tenantID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("inventory_check_repo: get_by_id: %w", err)
	}
	return c, nil
}

// List returns a paginated list of inventory checks for a tenant plus total count.
func (r *InventoryCheckRepo) List(ctx context.Context, tenantID uuid.UUID, page int, perPage int) ([]*domain.InventoryCheck, int64, error) {
	offset := (page - 1) * perPage

	var total int64
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM inventory_checks WHERE tenant_id = $1`, tenantID,
	).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("inventory_check_repo: list count: %w", err)
	}

	query := fmt.Sprintf(
		`SELECT %s FROM inventory_checks WHERE tenant_id = $1 ORDER BY started_at DESC LIMIT $2 OFFSET $3`,
		checkColumns,
	)

	rows, err := r.pool.Query(ctx, query, tenantID, perPage, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("inventory_check_repo: list query: %w", err)
	}
	defer rows.Close()

	var checks []*domain.InventoryCheck
	for rows.Next() {
		c, scanErr := scanCheck(rows)
		if scanErr != nil {
			return nil, 0, fmt.Errorf("inventory_check_repo: list scan: %w", scanErr)
		}
		checks = append(checks, c)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("inventory_check_repo: list rows: %w", err)
	}
	return checks, total, nil
}

// Complete updates the inventory check to completed status with counts.
func (r *InventoryCheckRepo) Complete(ctx context.Context, check *domain.InventoryCheck) error {
	query := `
		UPDATE inventory_checks SET
			status = $3, completed_at = $4,
			expected_count = $5, actual_count = $6, discrepancy_count = $7
		WHERE id = $1 AND tenant_id = $2
		RETURNING started_at`

	err := r.pool.QueryRow(ctx, query,
		check.ID, check.TenantID,
		check.Status, check.CompletedAt,
		check.ExpectedCount, check.ActualCount, check.DiscrepancyCount,
	).Scan(&check.StartedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apperrors.ErrNotFound
		}
		return fmt.Errorf("inventory_check_repo: complete: %w", err)
	}
	return nil
}

// CreateItem inserts a new inventory check item.
func (r *InventoryCheckRepo) CreateItem(ctx context.Context, item *domain.InventoryCheckItem) error {
	query := `
		INSERT INTO inventory_check_items (
			id, check_id, equipment_id, expected, found, notes
		) VALUES (
			$1, $2, $3, $4, $5, $6
		) RETURNING id`

	err := r.pool.QueryRow(ctx, query,
		item.ID, item.CheckID, item.EquipmentID,
		item.Expected, item.Found, nilIfEmpty(item.Notes),
	).Scan(&item.ID)
	if err != nil {
		return fmt.Errorf("inventory_check_repo: create_item: %w", err)
	}
	return nil
}

// ScanItem marks an equipment item as found (scanned) in a check.
// If no matching item exists, it creates a new item with expected=false, found=true.
func (r *InventoryCheckRepo) ScanItem(ctx context.Context, checkID uuid.UUID, equipmentID uuid.UUID) error {
	now := time.Now()

	// Try to update an existing item first.
	tag, err := r.pool.Exec(ctx,
		`UPDATE inventory_check_items SET found = true, scanned_at = $3 WHERE check_id = $1 AND equipment_id = $2`,
		checkID, equipmentID, now,
	)
	if err != nil {
		return fmt.Errorf("inventory_check_repo: scan_item update: %w", err)
	}

	// If no existing item was updated, insert a new one (unexpected item found during scan).
	if tag.RowsAffected() == 0 {
		_, err = r.pool.Exec(ctx,
			`INSERT INTO inventory_check_items (id, check_id, equipment_id, expected, found, scanned_at) VALUES ($1, $2, $3, false, true, $4)`,
			uuid.New(), checkID, equipmentID, now,
		)
		if err != nil {
			return fmt.Errorf("inventory_check_repo: scan_item insert: %w", err)
		}
	}

	return nil
}

// GetItems returns all items for an inventory check.
func (r *InventoryCheckRepo) GetItems(ctx context.Context, checkID uuid.UUID) ([]*domain.InventoryCheckItem, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM inventory_check_items WHERE check_id = $1 ORDER BY equipment_id`,
		checkItemColumns,
	)

	rows, err := r.pool.Query(ctx, query, checkID)
	if err != nil {
		return nil, fmt.Errorf("inventory_check_repo: get_items query: %w", err)
	}
	defer rows.Close()

	var items []*domain.InventoryCheckItem
	for rows.Next() {
		i, scanErr := scanCheckItem(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("inventory_check_repo: get_items scan: %w", scanErr)
		}
		items = append(items, i)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("inventory_check_repo: get_items rows: %w", err)
	}
	return items, nil
}

// CountExpected returns the number of items expected (expected=true) in a check.
func (r *InventoryCheckRepo) CountExpected(ctx context.Context, checkID uuid.UUID) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM inventory_check_items WHERE check_id = $1 AND expected = true`,
		checkID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("inventory_check_repo: count_expected: %w", err)
	}
	return count, nil
}

// CountFound returns the number of items found (found=true) in a check.
func (r *InventoryCheckRepo) CountFound(ctx context.Context, checkID uuid.UUID) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM inventory_check_items WHERE check_id = $1 AND found = true`,
		checkID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("inventory_check_repo: count_found: %w", err)
	}
	return count, nil
}

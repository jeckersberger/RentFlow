package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	apperrors "github.com/jeckersberger/EquipFlow/pkg/common/errors"
	"github.com/jeckersberger/EquipFlow/services/inventory/internal/domain"
)

const priceRuleColumns = `
	id, tenant_id, equipment_id, category_id,
	base_price_day, base_price_week,
	tier2_from_days, tier2_price_day,
	tier3_from_days, tier3_price_day,
	qty_discount_threshold, qty_discount_pct,
	season_start, season_end, season_surcharge_pct,
	is_active, created_at, updated_at`

// PriceRuleRepo implements domain.PriceRuleRepository using PostgreSQL.
type PriceRuleRepo struct {
	pool *pgxpool.Pool
}

// NewPriceRuleRepo creates a new PriceRuleRepo.
func NewPriceRuleRepo(pool *pgxpool.Pool) *PriceRuleRepo {
	return &PriceRuleRepo{pool: pool}
}

func scanPriceRule(row pgx.Row) (*domain.PriceRule, error) {
	r := &domain.PriceRule{}
	err := row.Scan(
		&r.ID, &r.TenantID, &r.EquipmentID, &r.CategoryID,
		&r.BasePriceDay, &r.BasePriceWeek,
		&r.Tier2FromDays, &r.Tier2PriceDay,
		&r.Tier3FromDays, &r.Tier3PriceDay,
		&r.QtyDiscountThreshold, &r.QtyDiscountPct,
		&r.SeasonStart, &r.SeasonEnd, &r.SeasonSurchargePct,
		&r.IsActive, &r.CreatedAt, &r.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return r, nil
}

// Create inserts a new price rule.
func (repo *PriceRuleRepo) Create(ctx context.Context, rule *domain.PriceRule) error {
	query := `
		INSERT INTO price_rules (
			id, tenant_id, equipment_id, category_id,
			base_price_day, base_price_week,
			tier2_from_days, tier2_price_day,
			tier3_from_days, tier3_price_day,
			qty_discount_threshold, qty_discount_pct,
			season_start, season_end, season_surcharge_pct,
			is_active
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		RETURNING created_at, updated_at`

	err := repo.pool.QueryRow(ctx, query,
		rule.ID, rule.TenantID, rule.EquipmentID, rule.CategoryID,
		rule.BasePriceDay, rule.BasePriceWeek,
		rule.Tier2FromDays, rule.Tier2PriceDay,
		rule.Tier3FromDays, rule.Tier3PriceDay,
		rule.QtyDiscountThreshold, rule.QtyDiscountPct,
		rule.SeasonStart, rule.SeasonEnd, rule.SeasonSurchargePct,
		rule.IsActive,
	).Scan(&rule.CreatedAt, &rule.UpdatedAt)
	if err != nil {
		return fmt.Errorf("price_rule_repo: create: %w", err)
	}
	return nil
}

// GetByID retrieves a price rule by ID within a tenant scope.
func (repo *PriceRuleRepo) GetByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*domain.PriceRule, error) {
	query := fmt.Sprintf(`SELECT %s FROM price_rules WHERE id = $1 AND tenant_id = $2`, priceRuleColumns)
	r, err := scanPriceRule(repo.pool.QueryRow(ctx, query, id, tenantID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("price_rule_repo: get_by_id: %w", err)
	}
	return r, nil
}

// List returns a paginated list of price rules for a tenant.
func (repo *PriceRuleRepo) List(ctx context.Context, tenantID uuid.UUID, page int, perPage int) ([]*domain.PriceRule, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	var total int64
	err := repo.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM price_rules WHERE tenant_id = $1`,
		tenantID,
	).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("price_rule_repo: list count: %w", err)
	}

	query := fmt.Sprintf(
		`SELECT %s FROM price_rules WHERE tenant_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		priceRuleColumns,
	)
	rows, err := repo.pool.Query(ctx, query, tenantID, perPage, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("price_rule_repo: list query: %w", err)
	}
	defer rows.Close()

	var items []*domain.PriceRule
	for rows.Next() {
		r, scanErr := scanPriceRule(rows)
		if scanErr != nil {
			return nil, 0, fmt.Errorf("price_rule_repo: list scan: %w", scanErr)
		}
		items = append(items, r)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("price_rule_repo: list rows: %w", err)
	}
	return items, total, nil
}

// Update modifies an existing price rule.
func (repo *PriceRuleRepo) Update(ctx context.Context, rule *domain.PriceRule) error {
	query := `
		UPDATE price_rules SET
			equipment_id = $3, category_id = $4,
			base_price_day = $5, base_price_week = $6,
			tier2_from_days = $7, tier2_price_day = $8,
			tier3_from_days = $9, tier3_price_day = $10,
			qty_discount_threshold = $11, qty_discount_pct = $12,
			season_start = $13, season_end = $14, season_surcharge_pct = $15,
			is_active = $16,
			updated_at = NOW()
		WHERE id = $1 AND tenant_id = $2
		RETURNING updated_at`

	err := repo.pool.QueryRow(ctx, query,
		rule.ID, rule.TenantID,
		rule.EquipmentID, rule.CategoryID,
		rule.BasePriceDay, rule.BasePriceWeek,
		rule.Tier2FromDays, rule.Tier2PriceDay,
		rule.Tier3FromDays, rule.Tier3PriceDay,
		rule.QtyDiscountThreshold, rule.QtyDiscountPct,
		rule.SeasonStart, rule.SeasonEnd, rule.SeasonSurchargePct,
		rule.IsActive,
	).Scan(&rule.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apperrors.ErrNotFound
		}
		return fmt.Errorf("price_rule_repo: update: %w", err)
	}
	return nil
}

// Delete removes a price rule.
func (repo *PriceRuleRepo) Delete(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error {
	tag, err := repo.pool.Exec(ctx,
		`DELETE FROM price_rules WHERE id = $1 AND tenant_id = $2`,
		id, tenantID,
	)
	if err != nil {
		return fmt.Errorf("price_rule_repo: delete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

// FindByEquipment returns the active price rule for a specific equipment item.
func (repo *PriceRuleRepo) FindByEquipment(ctx context.Context, equipmentID uuid.UUID, tenantID uuid.UUID) (*domain.PriceRule, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM price_rules WHERE tenant_id = $1 AND equipment_id = $2 AND is_active = true LIMIT 1`,
		priceRuleColumns,
	)
	r, err := scanPriceRule(repo.pool.QueryRow(ctx, query, tenantID, equipmentID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("price_rule_repo: find_by_equipment: %w", err)
	}
	return r, nil
}

// FindByCategory returns the active price rule for a specific category.
func (repo *PriceRuleRepo) FindByCategory(ctx context.Context, categoryID uuid.UUID, tenantID uuid.UUID) (*domain.PriceRule, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM price_rules WHERE tenant_id = $1 AND category_id = $2 AND equipment_id IS NULL AND is_active = true LIMIT 1`,
		priceRuleColumns,
	)
	r, err := scanPriceRule(repo.pool.QueryRow(ctx, query, tenantID, categoryID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("price_rule_repo: find_by_category: %w", err)
	}
	return r, nil
}

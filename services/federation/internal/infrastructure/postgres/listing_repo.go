package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jeckersberger/EquipFlow/services/federation/internal/domain"

	apperrors "github.com/jeckersberger/EquipFlow/pkg/common/errors"
)

// listingColumns lists all columns of the shared_listings table for consistent scanning.
const listingColumns = `
	id, tenant_id, equipment_id, daily_rate, weekly_rate,
	available_from, available_until, is_active, notes,
	created_at, updated_at`

// ListingRepo implements domain.SharedListingRepository using PostgreSQL.
type ListingRepo struct {
	pool *pgxpool.Pool
}

// NewListingRepo creates a new ListingRepo.
func NewListingRepo(pool *pgxpool.Pool) *ListingRepo {
	return &ListingRepo{pool: pool}
}

// scanListing scans a single shared_listings row into a domain.SharedListing.
func scanListing(row pgx.Row) (*domain.SharedListing, error) {
	l := &domain.SharedListing{}
	var notes *string

	err := row.Scan(
		&l.ID, &l.TenantID, &l.EquipmentID, &l.DailyRate, &l.WeeklyRate,
		&l.AvailableFrom, &l.AvailableUntil, &l.IsActive, &notes,
		&l.CreatedAt, &l.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if notes != nil {
		l.Notes = *notes
	}

	return l, nil
}

// Create inserts a new shared listing record.
func (r *ListingRepo) Create(ctx context.Context, listing *domain.SharedListing) error {
	query := `
		INSERT INTO shared_listings (
			id, tenant_id, equipment_id, daily_rate, weekly_rate,
			available_from, available_until, is_active, notes
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at`

	err := r.pool.QueryRow(ctx, query,
		listing.ID, listing.TenantID, listing.EquipmentID,
		listing.DailyRate, listing.WeeklyRate,
		listing.AvailableFrom, listing.AvailableUntil,
		listing.IsActive, nilIfEmpty(listing.Notes),
	).Scan(&listing.ID, &listing.CreatedAt, &listing.UpdatedAt)
	if err != nil {
		return fmt.Errorf("listing_repo: create: %w", err)
	}
	return nil
}

// GetByID retrieves a shared listing by its primary key within a tenant scope.
func (r *ListingRepo) GetByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*domain.SharedListing, error) {
	query := fmt.Sprintf(`SELECT %s FROM shared_listings WHERE id = $1 AND tenant_id = $2`, listingColumns)
	l, err := scanListing(r.pool.QueryRow(ctx, query, id, tenantID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("listing_repo: get_by_id: %w", err)
	}
	return l, nil
}

// List returns a filtered, paginated list of shared listings for a tenant plus total count.
func (r *ListingRepo) List(ctx context.Context, tenantID uuid.UUID, filter domain.ListingFilter) ([]*domain.SharedListing, int64, error) {
	conditions := []string{"tenant_id = $1"}
	args := []interface{}{tenantID}
	argIdx := 2

	if filter.EquipmentID != nil {
		conditions = append(conditions, fmt.Sprintf("equipment_id = $%d", argIdx))
		args = append(args, *filter.EquipmentID)
		argIdx++
	}
	if filter.IsActive != nil {
		conditions = append(conditions, fmt.Sprintf("is_active = $%d", argIdx))
		args = append(args, *filter.IsActive)
		argIdx++
	}

	where := strings.Join(conditions, " AND ")

	// Count query.
	var total int64
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM shared_listings WHERE %s`, where)
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("listing_repo: list count: %w", err)
	}

	// Data query with pagination.
	offset := (filter.Page - 1) * filter.PerPage
	dataQuery := fmt.Sprintf(
		`SELECT %s FROM shared_listings WHERE %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		listingColumns, where, argIdx, argIdx+1,
	)
	args = append(args, filter.PerPage, offset)

	rows, err := r.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("listing_repo: list query: %w", err)
	}
	defer rows.Close()

	var items []*domain.SharedListing
	for rows.Next() {
		l, scanErr := scanListing(rows)
		if scanErr != nil {
			return nil, 0, fmt.Errorf("listing_repo: list scan: %w", scanErr)
		}
		items = append(items, l)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("listing_repo: list rows: %w", err)
	}
	return items, total, nil
}

// Update updates an existing shared listing record.
func (r *ListingRepo) Update(ctx context.Context, listing *domain.SharedListing) error {
	query := `
		UPDATE shared_listings SET
			equipment_id = $3,
			daily_rate = $4,
			weekly_rate = $5,
			available_from = $6,
			available_until = $7,
			is_active = $8,
			notes = $9,
			updated_at = NOW()
		WHERE id = $1 AND tenant_id = $2
		RETURNING updated_at`

	err := r.pool.QueryRow(ctx, query,
		listing.ID, listing.TenantID, listing.EquipmentID,
		listing.DailyRate, listing.WeeklyRate,
		listing.AvailableFrom, listing.AvailableUntil,
		listing.IsActive, nilIfEmpty(listing.Notes),
	).Scan(&listing.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apperrors.ErrNotFound
		}
		return fmt.Errorf("listing_repo: update: %w", err)
	}
	return nil
}

// Ensure interface compliance at compile time.
var _ domain.SharedListingRepository = (*ListingRepo)(nil)

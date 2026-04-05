package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jeckersberger/EquipFlow/services/crew/internal/domain"

	apperrors "github.com/jeckersberger/EquipFlow/pkg/common/errors"
)

// availabilityBlockColumns lists all columns of the crew_availability_blocks table for consistent scanning.
const availabilityBlockColumns = `
	id, tenant_id, crew_member_id, block_type, start_date, end_date,
	notes, created_at, updated_at`

// AvailabilityBlockRepo implements domain.AvailabilityBlockRepository using PostgreSQL.
type AvailabilityBlockRepo struct {
	pool *pgxpool.Pool
}

// NewAvailabilityBlockRepo creates a new AvailabilityBlockRepo.
func NewAvailabilityBlockRepo(pool *pgxpool.Pool) *AvailabilityBlockRepo {
	return &AvailabilityBlockRepo{pool: pool}
}

// scanAvailabilityBlock scans a single crew_availability_blocks row into a domain.AvailabilityBlock.
func scanAvailabilityBlock(row pgx.Row) (*domain.AvailabilityBlock, error) {
	b := &domain.AvailabilityBlock{}
	var (
		startDate time.Time
		endDate   time.Time
		notes     *string
	)

	err := row.Scan(
		&b.ID, &b.TenantID, &b.CrewMemberID, &b.BlockType,
		&startDate, &endDate, &notes, &b.CreatedAt, &b.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	b.StartDate = startDate.Format("2006-01-02")
	b.EndDate = endDate.Format("2006-01-02")
	if notes != nil {
		b.Notes = *notes
	}

	return b, nil
}

// Create inserts a new availability block record.
func (r *AvailabilityBlockRepo) Create(ctx context.Context, block *domain.AvailabilityBlock) error {
	query := `
		INSERT INTO crew_availability_blocks (
			id, tenant_id, crew_member_id, block_type, start_date, end_date, notes
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at, updated_at`

	err := r.pool.QueryRow(ctx, query,
		block.ID, block.TenantID, block.CrewMemberID,
		block.BlockType, block.StartDate, block.EndDate,
		nilIfEmpty(block.Notes),
	).Scan(&block.CreatedAt, &block.UpdatedAt)
	if err != nil {
		return fmt.Errorf("availability_block_repo: create: %w", err)
	}
	return nil
}

// GetByID retrieves a single availability block by primary key scoped to a tenant.
func (r *AvailabilityBlockRepo) GetByID(ctx context.Context, id, tenantID uuid.UUID) (*domain.AvailabilityBlock, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM crew_availability_blocks WHERE id = $1 AND tenant_id = $2`,
		availabilityBlockColumns,
	)

	b, err := scanAvailabilityBlock(r.pool.QueryRow(ctx, query, id, tenantID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("availability_block_repo: get_by_id: %w", err)
	}
	return b, nil
}

// ListByCrewMember returns all availability blocks for a specific crew member
// that overlap with the given date range.
func (r *AvailabilityBlockRepo) ListByCrewMember(
	ctx context.Context,
	crewMemberID, tenantID uuid.UUID,
	from, to string,
) ([]*domain.AvailabilityBlock, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM crew_availability_blocks
		 WHERE crew_member_id = $1 AND tenant_id = $2
		   AND start_date <= $4 AND end_date >= $3
		 ORDER BY start_date ASC`,
		availabilityBlockColumns,
	)

	return r.queryBlocks(ctx, query, crewMemberID, tenantID, from, to)
}

// ListByDateRange returns all availability blocks for a tenant that overlap
// with the given date range.
func (r *AvailabilityBlockRepo) ListByDateRange(
	ctx context.Context,
	tenantID uuid.UUID,
	from, to string,
) ([]*domain.AvailabilityBlock, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM crew_availability_blocks
		 WHERE tenant_id = $1
		   AND start_date <= $3 AND end_date >= $2
		 ORDER BY start_date ASC, crew_member_id ASC`,
		availabilityBlockColumns,
	)

	return r.queryBlocks(ctx, query, tenantID, from, to)
}

// Delete removes an availability block by ID within a tenant scope.
func (r *AvailabilityBlockRepo) Delete(ctx context.Context, id, tenantID uuid.UUID) error {
	tag, err := r.pool.Exec(ctx,
		`DELETE FROM crew_availability_blocks WHERE id = $1 AND tenant_id = $2`,
		id, tenantID,
	)
	if err != nil {
		return fmt.Errorf("availability_block_repo: delete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

// queryBlocks is a helper that executes a query and scans rows into AvailabilityBlock slices.
func (r *AvailabilityBlockRepo) queryBlocks(
	ctx context.Context,
	query string,
	args ...interface{},
) ([]*domain.AvailabilityBlock, error) {
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("availability_block_repo: query: %w", err)
	}
	defer rows.Close()

	var items []*domain.AvailabilityBlock
	for rows.Next() {
		b := &domain.AvailabilityBlock{}
		var (
			startDate time.Time
			endDate   time.Time
			notes     *string
		)
		scanErr := rows.Scan(
			&b.ID, &b.TenantID, &b.CrewMemberID, &b.BlockType,
			&startDate, &endDate, &notes, &b.CreatedAt, &b.UpdatedAt,
		)
		if scanErr != nil {
			return nil, fmt.Errorf("availability_block_repo: scan: %w", scanErr)
		}
		b.StartDate = startDate.Format("2006-01-02")
		b.EndDate = endDate.Format("2006-01-02")
		if notes != nil {
			b.Notes = *notes
		}
		items = append(items, b)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("availability_block_repo: rows: %w", err)
	}
	return items, nil
}

// Ensure interface compliance at compile time.
var _ domain.AvailabilityBlockRepository = (*AvailabilityBlockRepo)(nil)

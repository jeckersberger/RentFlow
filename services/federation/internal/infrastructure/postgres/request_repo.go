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

// requestColumns lists all columns of the federation_requests table for consistent scanning.
const requestColumns = `
	id, tenant_id, partner_id, listing_id, status,
	start_date, end_date, total_cost, notes,
	requested_by, created_at, updated_at`

// RequestRepo implements domain.FederationRequestRepository using PostgreSQL.
type RequestRepo struct {
	pool *pgxpool.Pool
}

// NewRequestRepo creates a new RequestRepo.
func NewRequestRepo(pool *pgxpool.Pool) *RequestRepo {
	return &RequestRepo{pool: pool}
}

// scanRequest scans a single federation_requests row into a domain.FederationRequest.
func scanRequest(row pgx.Row) (*domain.FederationRequest, error) {
	fr := &domain.FederationRequest{}
	var (
		status *string
		notes  *string
	)

	err := row.Scan(
		&fr.ID, &fr.TenantID, &fr.PartnerID, &fr.ListingID, &status,
		&fr.StartDate, &fr.EndDate, &fr.TotalCost, &notes,
		&fr.RequestedBy, &fr.CreatedAt, &fr.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if status != nil {
		fr.Status = *status
	}
	if notes != nil {
		fr.Notes = *notes
	}

	return fr, nil
}

// Create inserts a new federation request record.
func (r *RequestRepo) Create(ctx context.Context, req *domain.FederationRequest) error {
	query := `
		INSERT INTO federation_requests (
			id, tenant_id, partner_id, listing_id, status,
			start_date, end_date, total_cost, notes, requested_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at, updated_at`

	err := r.pool.QueryRow(ctx, query,
		req.ID, req.TenantID, req.PartnerID, req.ListingID,
		req.Status, req.StartDate, req.EndDate,
		req.TotalCost, nilIfEmpty(req.Notes), req.RequestedBy,
	).Scan(&req.ID, &req.CreatedAt, &req.UpdatedAt)
	if err != nil {
		return fmt.Errorf("request_repo: create: %w", err)
	}
	return nil
}

// GetByID retrieves a federation request by its primary key within a tenant scope.
func (r *RequestRepo) GetByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*domain.FederationRequest, error) {
	query := fmt.Sprintf(`SELECT %s FROM federation_requests WHERE id = $1 AND tenant_id = $2`, requestColumns)
	fr, err := scanRequest(r.pool.QueryRow(ctx, query, id, tenantID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("request_repo: get_by_id: %w", err)
	}
	return fr, nil
}

// List returns a filtered, paginated list of federation requests for a tenant plus total count.
func (r *RequestRepo) List(ctx context.Context, tenantID uuid.UUID, filter domain.RequestFilter) ([]*domain.FederationRequest, int64, error) {
	conditions := []string{"tenant_id = $1"}
	args := []interface{}{tenantID}
	argIdx := 2

	if filter.PartnerID != nil {
		conditions = append(conditions, fmt.Sprintf("partner_id = $%d", argIdx))
		args = append(args, *filter.PartnerID)
		argIdx++
	}
	if filter.ListingID != nil {
		conditions = append(conditions, fmt.Sprintf("listing_id = $%d", argIdx))
		args = append(args, *filter.ListingID)
		argIdx++
	}
	if filter.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, *filter.Status)
		argIdx++
	}

	where := strings.Join(conditions, " AND ")

	// Count query.
	var total int64
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM federation_requests WHERE %s`, where)
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("request_repo: list count: %w", err)
	}

	// Data query with pagination.
	offset := (filter.Page - 1) * filter.PerPage
	dataQuery := fmt.Sprintf(
		`SELECT %s FROM federation_requests WHERE %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		requestColumns, where, argIdx, argIdx+1,
	)
	args = append(args, filter.PerPage, offset)

	rows, err := r.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("request_repo: list query: %w", err)
	}
	defer rows.Close()

	var items []*domain.FederationRequest
	for rows.Next() {
		fr, scanErr := scanRequest(rows)
		if scanErr != nil {
			return nil, 0, fmt.Errorf("request_repo: list scan: %w", scanErr)
		}
		items = append(items, fr)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("request_repo: list rows: %w", err)
	}
	return items, total, nil
}

// Update updates an existing federation request record.
func (r *RequestRepo) Update(ctx context.Context, req *domain.FederationRequest) error {
	query := `
		UPDATE federation_requests SET
			partner_id = $3,
			listing_id = $4,
			status = $5,
			start_date = $6,
			end_date = $7,
			total_cost = $8,
			notes = $9,
			requested_by = $10,
			updated_at = NOW()
		WHERE id = $1 AND tenant_id = $2
		RETURNING updated_at`

	err := r.pool.QueryRow(ctx, query,
		req.ID, req.TenantID, req.PartnerID, req.ListingID,
		req.Status, req.StartDate, req.EndDate,
		req.TotalCost, nilIfEmpty(req.Notes), req.RequestedBy,
	).Scan(&req.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apperrors.ErrNotFound
		}
		return fmt.Errorf("request_repo: update: %w", err)
	}
	return nil
}

// Ensure interface compliance at compile time.
var _ domain.FederationRequestRepository = (*RequestRepo)(nil)

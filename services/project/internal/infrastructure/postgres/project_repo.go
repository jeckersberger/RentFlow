package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jeckersberger/EquipFlow/services/project/internal/domain"

	apperrors "github.com/jeckersberger/EquipFlow/pkg/common/errors"
)

// projectColumns lists all columns of the projects table for consistent scanning.
const projectColumns = `
	id, tenant_id, name, project_number, description, status,
	customer_id, contact_name, contact_email, contact_phone,
	venue_name, venue_address, venue_lat, venue_lng,
	start_date, end_date, setup_date, teardown_date,
	color, budget, currency, manager_id, notes,
	created_at, updated_at`

// ProjectRepo implements domain.ProjectRepository using PostgreSQL.
type ProjectRepo struct {
	pool *pgxpool.Pool
}

// NewProjectRepo creates a new ProjectRepo.
func NewProjectRepo(pool *pgxpool.Pool) *ProjectRepo {
	return &ProjectRepo{pool: pool}
}

// scanProject scans a single project row into a domain.Project, handling nullable columns.
func scanProject(row pgx.Row) (*domain.Project, error) {
	p := &domain.Project{}
	var (
		description  *string
		customerID   *uuid.UUID
		contactName  *string
		contactEmail *string
		contactPhone *string
		venueName    *string
		venueAddress *string
		venueLat     *float64
		venueLng     *float64
		startDate    *time.Time
		endDate      *time.Time
		setupDate    *time.Time
		teardownDate *time.Time
		color        *string
		budget       *int64
		currency     *string
		managerID    *uuid.UUID
		notes        *string
	)

	err := row.Scan(
		&p.ID, &p.TenantID, &p.Name, &p.ProjectNumber, &description, &p.Status,
		&customerID, &contactName, &contactEmail, &contactPhone,
		&venueName, &venueAddress, &venueLat, &venueLng,
		&startDate, &endDate, &setupDate, &teardownDate,
		&color, &budget, &currency, &managerID, &notes,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	p.Description = derefString(description)
	p.CustomerID = customerID
	p.ContactName = derefString(contactName)
	p.ContactEmail = derefString(contactEmail)
	p.ContactPhone = derefString(contactPhone)
	p.VenueName = derefString(venueName)
	p.VenueAddress = derefString(venueAddress)
	p.VenueLat = venueLat
	p.VenueLng = venueLng
	p.StartDate = startDate
	p.EndDate = endDate
	p.SetupDate = setupDate
	p.TeardownDate = teardownDate
	p.Color = derefString(color)
	p.Budget = derefInt64(budget)
	p.Currency = derefString(currency)
	p.ManagerID = managerID
	p.Notes = derefString(notes)

	return p, nil
}

// Create inserts a new project record and scans back the generated fields.
func (r *ProjectRepo) Create(ctx context.Context, project *domain.Project) error {
	query := `
		INSERT INTO projects (
			id, tenant_id, name, project_number, description, status,
			customer_id, contact_name, contact_email, contact_phone,
			venue_name, venue_address, venue_lat, venue_lng,
			start_date, end_date, setup_date, teardown_date,
			color, budget, currency, manager_id, notes
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9, $10,
			$11, $12, $13, $14,
			$15, $16, $17, $18,
			$19, $20, $21, $22, $23
		) RETURNING created_at, updated_at`

	err := r.pool.QueryRow(ctx, query,
		project.ID, project.TenantID, project.Name, project.ProjectNumber,
		nilIfEmpty(project.Description), project.Status,
		project.CustomerID, nilIfEmpty(project.ContactName),
		nilIfEmpty(project.ContactEmail), nilIfEmpty(project.ContactPhone),
		nilIfEmpty(project.VenueName), nilIfEmpty(project.VenueAddress),
		project.VenueLat, project.VenueLng,
		project.StartDate, project.EndDate, project.SetupDate, project.TeardownDate,
		nilIfEmpty(project.Color), nilIfZeroInt64(project.Budget),
		nilIfEmpty(project.Currency), project.ManagerID, nilIfEmpty(project.Notes),
	).Scan(&project.CreatedAt, &project.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return apperrors.ErrConflict
		}
		return fmt.Errorf("project_repo: create: %w", err)
	}
	return nil
}

// GetByID retrieves a single project record by primary key scoped to a tenant.
func (r *ProjectRepo) GetByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*domain.Project, error) {
	query := fmt.Sprintf(`SELECT %s FROM projects WHERE id = $1 AND tenant_id = $2`, projectColumns)
	p, err := scanProject(r.pool.QueryRow(ctx, query, id, tenantID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("project_repo: get_by_id: %w", err)
	}
	return p, nil
}

// List returns a filtered, paginated list of projects for a tenant plus total count.
func (r *ProjectRepo) List(ctx context.Context, tenantID uuid.UUID, filter domain.ProjectFilter) ([]*domain.Project, int64, error) {
	// Build dynamic WHERE clause.
	conditions := []string{"tenant_id = $1"}
	args := []interface{}{tenantID}
	argIdx := 2

	if filter.Status != nil && *filter.Status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, *filter.Status)
		argIdx++
	}
	if filter.Search != nil && *filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf(
			"(name ILIKE $%d OR description ILIKE $%d)",
			argIdx, argIdx,
		))
		args = append(args, "%"+*filter.Search+"%")
		argIdx++
	}
	if filter.CustomerID != nil {
		conditions = append(conditions, fmt.Sprintf("customer_id = $%d", argIdx))
		args = append(args, *filter.CustomerID)
		argIdx++
	}
	if filter.StartAfter != nil {
		conditions = append(conditions, fmt.Sprintf("start_date >= $%d", argIdx))
		args = append(args, *filter.StartAfter)
		argIdx++
	}
	if filter.EndBefore != nil {
		conditions = append(conditions, fmt.Sprintf("end_date <= $%d", argIdx))
		args = append(args, *filter.EndBefore)
		argIdx++
	}

	where := strings.Join(conditions, " AND ")

	// Count query.
	var total int64
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM projects WHERE %s`, where)
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("project_repo: list count: %w", err)
	}

	// Data query with pagination.
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
		`SELECT %s FROM projects WHERE %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		projectColumns, where, argIdx, argIdx+1,
	)
	args = append(args, perPage, offset)

	rows, err := r.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("project_repo: list query: %w", err)
	}
	defer rows.Close()

	var items []*domain.Project
	for rows.Next() {
		p, scanErr := scanProject(rows)
		if scanErr != nil {
			return nil, 0, fmt.Errorf("project_repo: list scan: %w", scanErr)
		}
		items = append(items, p)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("project_repo: list rows: %w", err)
	}
	return items, total, nil
}

// Update modifies an existing project record.
func (r *ProjectRepo) Update(ctx context.Context, project *domain.Project) error {
	query := `
		UPDATE projects SET
			name = $3, project_number = $4, description = $5, status = $6,
			customer_id = $7, contact_name = $8, contact_email = $9, contact_phone = $10,
			venue_name = $11, venue_address = $12, venue_lat = $13, venue_lng = $14,
			start_date = $15, end_date = $16, setup_date = $17, teardown_date = $18,
			color = $19, budget = $20, currency = $21, manager_id = $22, notes = $23,
			updated_at = NOW()
		WHERE id = $1 AND tenant_id = $2
		RETURNING updated_at`

	err := r.pool.QueryRow(ctx, query,
		project.ID, project.TenantID,
		project.Name, project.ProjectNumber, nilIfEmpty(project.Description), project.Status,
		project.CustomerID, nilIfEmpty(project.ContactName),
		nilIfEmpty(project.ContactEmail), nilIfEmpty(project.ContactPhone),
		nilIfEmpty(project.VenueName), nilIfEmpty(project.VenueAddress),
		project.VenueLat, project.VenueLng,
		project.StartDate, project.EndDate, project.SetupDate, project.TeardownDate,
		nilIfEmpty(project.Color), nilIfZeroInt64(project.Budget),
		nilIfEmpty(project.Currency), project.ManagerID, nilIfEmpty(project.Notes),
	).Scan(&project.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apperrors.ErrNotFound
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return apperrors.ErrConflict
		}
		return fmt.Errorf("project_repo: update: %w", err)
	}
	return nil
}

// Delete removes a project by primary key scoped to a tenant.
func (r *ProjectRepo) Delete(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error {
	tag, err := r.pool.Exec(ctx,
		`DELETE FROM projects WHERE id = $1 AND tenant_id = $2`,
		id, tenantID,
	)
	if err != nil {
		return fmt.Errorf("project_repo: delete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

// UpdateStatus changes the status of a project.
func (r *ProjectRepo) UpdateStatus(ctx context.Context, id uuid.UUID, tenantID uuid.UUID, status string) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE projects SET status = $3, updated_at = NOW() WHERE id = $1 AND tenant_id = $2`,
		id, tenantID, status,
	)
	if err != nil {
		return fmt.Errorf("project_repo: update_status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

// Search performs an ILIKE search on project name and description with pagination.
func (r *ProjectRepo) Search(ctx context.Context, tenantID uuid.UUID, query string, page int, perPage int) ([]*domain.Project, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	offset := (page - 1) * perPage
	pattern := "%" + query + "%"

	searchCondition := `tenant_id = $1 AND (name ILIKE $2 OR description ILIKE $2)`

	// Count query.
	var total int64
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM projects WHERE %s`, searchCondition)
	err := r.pool.QueryRow(ctx, countQuery, tenantID, pattern).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("project_repo: search count: %w", err)
	}

	// Data query.
	dataQuery := fmt.Sprintf(
		`SELECT %s FROM projects WHERE %s ORDER BY created_at DESC LIMIT $3 OFFSET $4`,
		projectColumns, searchCondition,
	)

	rows, err := r.pool.Query(ctx, dataQuery, tenantID, pattern, perPage, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("project_repo: search query: %w", err)
	}
	defer rows.Close()

	var items []*domain.Project
	for rows.Next() {
		p, scanErr := scanProject(rows)
		if scanErr != nil {
			return nil, 0, fmt.Errorf("project_repo: search scan: %w", scanErr)
		}
		items = append(items, p)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("project_repo: search rows: %w", err)
	}
	return items, total, nil
}

// GetByDateRange returns projects whose date range overlaps the given range.
func (r *ProjectRepo) GetByDateRange(ctx context.Context, tenantID uuid.UUID, from time.Time, to time.Time) ([]*domain.Project, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM projects
		 WHERE tenant_id = $1
		   AND start_date IS NOT NULL AND end_date IS NOT NULL
		   AND start_date <= $3 AND end_date >= $2
		 ORDER BY start_date ASC`,
		projectColumns,
	)

	rows, err := r.pool.Query(ctx, query, tenantID, from, to)
	if err != nil {
		return nil, fmt.Errorf("project_repo: get_by_date_range query: %w", err)
	}
	defer rows.Close()

	var items []*domain.Project
	for rows.Next() {
		p, scanErr := scanProject(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("project_repo: get_by_date_range scan: %w", scanErr)
		}
		items = append(items, p)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("project_repo: get_by_date_range rows: %w", err)
	}
	return items, nil
}

// nilIfEmpty returns nil if the string is empty, otherwise a pointer to it.
func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// nilIfZeroInt64 returns nil if the value is 0, otherwise a pointer to it.
func nilIfZeroInt64(v int64) *int64 {
	if v == 0 {
		return nil
	}
	return &v
}

// derefString safely dereferences a *string, returning "" if nil.
func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// derefInt64 safely dereferences a *int64, returning 0 if nil.
func derefInt64(v *int64) int64 {
	if v == nil {
		return 0
	}
	return *v
}

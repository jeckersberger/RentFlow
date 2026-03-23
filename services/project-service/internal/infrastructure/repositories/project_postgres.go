package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jeckersberger/rentflow/pkg/common/database"
	"github.com/jeckersberger/rentflow/services/project-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/project-service/internal/ports"
	"github.com/lib/pq"
)

type ProjectPostgres struct {
	db *database.PostgresPool
}

func NewProjectPostgres(db *database.PostgresPool) *ProjectPostgres {
	return &ProjectPostgres{db: db}
}

// projectSelectColumns uses COALESCE for all nullable string/numeric columns to avoid
// NULL-to-Go-string scan errors with database/sql.
const projectSelectColumns = `id, tenant_id, name,
			   COALESCE(description, '') as description,
			   client_name,
			   COALESCE(client_email, '') as client_email,
			   COALESCE(client_phone, '') as client_phone,
			   COALESCE(client_street, '') as client_street,
			   COALESCE(client_city, '') as client_city,
			   COALESCE(client_state, '') as client_state,
			   COALESCE(client_postal_code, '') as client_postal_code,
			   COALESCE(client_country, '') as client_country,
			   COALESCE(client_coordinates, '') as client_coordinates,
			   COALESCE(venue_street, '') as venue_street,
			   COALESCE(venue_city, '') as venue_city,
			   COALESCE(venue_state, '') as venue_state,
			   COALESCE(venue_postal_code, '') as venue_postal_code,
			   COALESCE(venue_country, '') as venue_country,
			   COALESCE(venue_coordinates, '') as venue_coordinates,
			   status, start_date, end_date,
			   setup_date, teardown_date,
			   COALESCE(project_manager, '') as project_manager,
			   COALESCE(budget, 0) as budget,
			   COALESCE(currency, 'USD') as currency,
			   COALESCE(notes, '') as notes,
			   tags, created_at, updated_at,
			   COALESCE(created_by_user_id, '') as created_by_user_id`

func (r *ProjectPostgres) scanProject(scanner interface{ Scan(...interface{}) error }) (*domain.Project, error) {
	p := &domain.Project{}
	err := scanner.Scan(
		&p.ID, &p.TenantID, &p.Name, &p.Description, &p.ClientName, &p.ClientEmail,
		&p.ClientPhone, &p.ClientAddress.Street, &p.ClientAddress.City, &p.ClientAddress.State,
		&p.ClientAddress.PostalCode, &p.ClientAddress.Country, &p.ClientAddress.Coordinates,
		&p.VenueAddress.Street, &p.VenueAddress.City, &p.VenueAddress.State,
		&p.VenueAddress.PostalCode, &p.VenueAddress.Country, &p.VenueAddress.Coordinates,
		&p.Status, &p.StartDate, &p.EndDate, &p.SetupDate, &p.TeardownDate,
		&p.ProjectManager, &p.Budget, &p.Currency, &p.Notes, pq.Array(&p.Tags),
		&p.CreatedAt, &p.UpdatedAt, &p.CreatedByUserID,
	)
	return p, err
}

func (r *ProjectPostgres) Create(ctx context.Context, p *domain.Project) error {
	query := `
		INSERT INTO projects.projects (
			id, tenant_id, name, description, client_name, client_email,
			client_phone, client_street, client_city, client_state,
			client_postal_code, client_country, client_coordinates,
			venue_street, venue_city, venue_state, venue_postal_code,
			venue_country, venue_coordinates, status, start_date, end_date,
			setup_date, teardown_date, project_manager, budget, currency,
			notes, tags, created_at, updated_at, created_by_user_id
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14,
			$15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26,
			$27, $28, $29, $30, $31, $32
		)
	`

	_, err := r.db.Exec(ctx, query,
		p.ID, p.TenantID, p.Name, p.Description, p.ClientName, p.ClientEmail,
		p.ClientPhone, p.ClientAddress.Street, p.ClientAddress.City, p.ClientAddress.State,
		p.ClientAddress.PostalCode, p.ClientAddress.Country, p.ClientAddress.Coordinates,
		p.VenueAddress.Street, p.VenueAddress.City, p.VenueAddress.State,
		p.VenueAddress.PostalCode, p.VenueAddress.Country, p.VenueAddress.Coordinates,
		string(p.Status), p.StartDate, p.EndDate, p.SetupDate, p.TeardownDate,
		p.ProjectManager, p.Budget, p.Currency, p.Notes, pq.Array(p.Tags),
		p.CreatedAt, p.UpdatedAt, p.CreatedByUserID,
	)

	if err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}

	return nil
}

func (r *ProjectPostgres) Update(ctx context.Context, p *domain.Project) error {
	query := `
		UPDATE projects.projects SET
			name = $3, description = $4, client_name = $5, client_email = $6,
			client_phone = $7, client_street = $8, client_city = $9,
			client_state = $10, client_postal_code = $11, client_country = $12,
			client_coordinates = $13, venue_street = $14, venue_city = $15,
			venue_state = $16, venue_postal_code = $17, venue_country = $18,
			venue_coordinates = $19, status = $20, start_date = $21, end_date = $22,
			setup_date = $23, teardown_date = $24, project_manager = $25,
			budget = $26, currency = $27, notes = $28, tags = $29, updated_at = $30
		WHERE id = $1 AND tenant_id = $2
	`

	result, err := r.db.Exec(ctx, query,
		p.ID, p.TenantID, p.Name, p.Description, p.ClientName, p.ClientEmail,
		p.ClientPhone, p.ClientAddress.Street, p.ClientAddress.City, p.ClientAddress.State,
		p.ClientAddress.PostalCode, p.ClientAddress.Country, p.ClientAddress.Coordinates,
		p.VenueAddress.Street, p.VenueAddress.City, p.VenueAddress.State,
		p.VenueAddress.PostalCode, p.VenueAddress.Country, p.VenueAddress.Coordinates,
		string(p.Status), p.StartDate, p.EndDate, p.SetupDate, p.TeardownDate,
		p.ProjectManager, p.Budget, p.Currency, p.Notes, pq.Array(p.Tags), p.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update project: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *ProjectPostgres) GetByID(ctx context.Context, tenantID, projectID string) (*domain.Project, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM projects.projects
		WHERE id = $1 AND tenant_id = $2
	`, projectSelectColumns)

	row := r.db.QueryRow(ctx, query, projectID, tenantID)
	p, err := r.scanProject(row)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	return p, nil
}

func (r *ProjectPostgres) List(ctx context.Context, tenantID string, limit, offset int) (*ports.ProjectListResult, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM projects.projects
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, projectSelectColumns)

	countQuery := `SELECT COUNT(*) FROM projects.projects WHERE tenant_id = $1`

	var count int64
	err := r.db.QueryRow(ctx, countQuery, tenantID).Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("failed to count projects: %w", err)
	}

	rows, err := r.db.Query(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}
	defer rows.Close()

	projects := make([]*domain.Project, 0)
	for rows.Next() {
		p, err := r.scanProject(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan project: %w", err)
		}
		projects = append(projects, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating projects: %w", err)
	}

	return &ports.ProjectListResult{
		Items:  projects,
		Total:  count,
		Limit:  limit,
		Offset: offset,
	}, nil
}

func (r *ProjectPostgres) Search(ctx context.Context, tenantID, term string, limit, offset int) (*ports.ProjectListResult, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM projects.projects
		WHERE tenant_id = $1 AND (
			name ILIKE $2 OR
			description ILIKE $2 OR
			client_name ILIKE $2
		)
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`, projectSelectColumns)

	countQuery := `
		SELECT COUNT(*) FROM projects.projects
		WHERE tenant_id = $1 AND (
			name ILIKE $2 OR
			description ILIKE $2 OR
			client_name ILIKE $2
		)
	`

	searchTerm := "%" + term + "%"

	var count int64
	err := r.db.QueryRow(ctx, countQuery, tenantID, searchTerm).Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("failed to count projects: %w", err)
	}

	rows, err := r.db.Query(ctx, query, tenantID, searchTerm, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search projects: %w", err)
	}
	defer rows.Close()

	projects := make([]*domain.Project, 0)
	for rows.Next() {
		p, err := r.scanProject(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan project: %w", err)
		}
		projects = append(projects, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating projects: %w", err)
	}

	return &ports.ProjectListResult{
		Items:  projects,
		Total:  count,
		Limit:  limit,
		Offset: offset,
	}, nil
}

func (r *ProjectPostgres) ListByDateRange(ctx context.Context, tenantID, startDate, endDate string) ([]*domain.Project, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM projects.projects
		WHERE tenant_id = $1 AND start_date <= $3 AND end_date >= $2
		ORDER BY start_date ASC
	`, projectSelectColumns)

	rows, err := r.db.Query(ctx, query, tenantID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects by date range: %w", err)
	}
	defer rows.Close()

	projects := make([]*domain.Project, 0)
	for rows.Next() {
		p, err := r.scanProject(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan project: %w", err)
		}
		projects = append(projects, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating projects: %w", err)
	}

	return projects, nil
}

func (r *ProjectPostgres) Delete(ctx context.Context, tenantID, projectID string) error {
	query := `DELETE FROM projects.projects WHERE id = $1 AND tenant_id = $2`

	result, err := r.db.Exec(ctx, query, projectID, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

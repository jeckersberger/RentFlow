package repositories

import (
	"context"
	"database/sql"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/auth-service/internal/domain"
)

// PostgresTenantRepository implements the TenantRepository interface using PostgreSQL
type PostgresTenantRepository struct {
	db     *sql.DB
	logger logger.Logger
}

// NewPostgresTenantRepository creates a new PostgreSQL tenant repository
func NewPostgresTenantRepository(db *sql.DB, log logger.Logger) *PostgresTenantRepository {
	return &PostgresTenantRepository{
		db:     db,
		logger: log,
	}
}

// FindByID retrieves a tenant by ID
func (r *PostgresTenantRepository) FindByID(ctx context.Context, id string) (*domain.Tenant, error) {
	query := `
		SELECT id, name, slug, address_street, address_city, address_zip, address_country,
		       logo, default_language, currency, tax_rate, invoice_prefix, status, created_at, updated_at
		FROM auth.tenants
		WHERE id = $1
	`

	row := r.db.QueryRowContext(ctx, query, id)
	return r.scanTenantRow(row)
}

// FindBySlug retrieves a tenant by slug
func (r *PostgresTenantRepository) FindBySlug(ctx context.Context, slug string) (*domain.Tenant, error) {
	query := `
		SELECT id, name, slug, address_street, address_city, address_zip, address_country,
		       logo, default_language, currency, tax_rate, invoice_prefix, status, created_at, updated_at
		FROM auth.tenants
		WHERE slug = $1
	`

	row := r.db.QueryRowContext(ctx, query, slug)
	return r.scanTenantRow(row)
}

// List retrieves tenants with pagination
func (r *PostgresTenantRepository) List(ctx context.Context, page, perPage int) ([]*domain.Tenant, int, error) {
	// Calculate offset
	offset := (page - 1) * perPage

	// Count total
	countQuery := `SELECT COUNT(*) FROM auth.tenants`
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery).Scan(&total); err != nil {
		r.logger.Error("failed to count tenants", err)
		return nil, 0, err
	}

	// Query tenants
	query := `
		SELECT id, name, slug, address_street, address_city, address_zip, address_country,
		       logo, default_language, currency, tax_rate, invoice_prefix, status, created_at, updated_at
		FROM auth.tenants
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, perPage, offset)
	if err != nil {
		r.logger.Error("failed to query tenants", err)
		return nil, 0, err
	}
	defer rows.Close()

	var tenants []*domain.Tenant
	for rows.Next() {
		tenant, err := r.scanTenantFromRow(rows)
		if err != nil {
			r.logger.Error("failed to scan tenant", err)
			continue
		}
		tenants = append(tenants, tenant)
	}

	if err = rows.Err(); err != nil {
		r.logger.Error("error iterating tenants", err)
		return nil, 0, err
	}

	return tenants, total, nil
}

// Save persists a tenant (creates or updates)
func (r *PostgresTenantRepository) Save(ctx context.Context, tenant *domain.Tenant) error {
	// Check if tenant exists
	exists := false
	err := r.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM auth.tenants WHERE id = $1)", tenant.ID).Scan(&exists)
	if err != nil {
		r.logger.Error("failed to check tenant existence", err, "id", tenant.ID)
		return err
	}

	now := time.Now()

	if exists {
		// Update
		query := `
			UPDATE auth.tenants
			SET name = $2, slug = $3, address_street = $4, address_city = $5, address_zip = $6, address_country = $7,
			    logo = $8, default_language = $9, currency = $10, tax_rate = $11, invoice_prefix = $12, status = $13, updated_at = $14
			WHERE id = $1
		`
		_, err := r.db.ExecContext(ctx, query,
			tenant.ID,
			tenant.Name,
			tenant.Slug,
			tenant.Address.Street,
			tenant.Address.City,
			tenant.Address.ZIP,
			tenant.Address.Country,
			tenant.Logo,
			tenant.Settings.DefaultLanguage,
			tenant.Settings.Currency,
			tenant.Settings.TaxRate,
			tenant.Settings.InvoicePrefix,
			string(tenant.Status),
			now,
		)
		if err != nil {
			r.logger.Error("failed to update tenant", err, "id", tenant.ID)
			return err
		}
	} else {
		// Insert
		query := `
			INSERT INTO auth.tenants
			(id, name, slug, address_street, address_city, address_zip, address_country, logo, default_language, currency, tax_rate, invoice_prefix, status, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		`
		_, err := r.db.ExecContext(ctx, query,
			tenant.ID,
			tenant.Name,
			tenant.Slug,
			tenant.Address.Street,
			tenant.Address.City,
			tenant.Address.ZIP,
			tenant.Address.Country,
			tenant.Logo,
			tenant.Settings.DefaultLanguage,
			tenant.Settings.Currency,
			tenant.Settings.TaxRate,
			tenant.Settings.InvoicePrefix,
			string(tenant.Status),
			now,
			now,
		)
		if err != nil {
			r.logger.Error("failed to insert tenant", err, "id", tenant.ID, "name", tenant.Name)
			return err
		}
	}

	return nil
}

// Delete deletes a tenant
func (r *PostgresTenantRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM auth.tenants WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Error("failed to delete tenant", err, "id", id)
		return err
	}
	return nil
}

// Helper functions

func (r *PostgresTenantRepository) scanTenantRow(row *sql.Row) (*domain.Tenant, error) {
	var id, name, slug, addressStreet, addressCity, addressZip, addressCountry, logo, status string
	var defaultLanguage, currency, invoicePrefix string
	var taxRate float64
	var createdAt, updatedAt time.Time

	err := row.Scan(&id, &name, &slug, &addressStreet, &addressCity, &addressZip, &addressCountry,
		&logo, &defaultLanguage, &currency, &taxRate, &invoicePrefix, &status, &createdAt, &updatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	tenant := &domain.Tenant{
		AggregateRoot: domain.AggregateRoot{
			ID:      id,
			Type:    "Tenant",
			Changes: []interface{}{},
		},
		Name: name,
		Slug: slug,
		Address: domain.Address{
			Street:  addressStreet,
			City:    addressCity,
			ZIP:     addressZip,
			Country: addressCountry,
		},
		Logo: logo,
		Settings: domain.TenantSettings{
			DefaultLanguage: defaultLanguage,
			Currency:        currency,
			TaxRate:         taxRate,
			InvoicePrefix:   invoicePrefix,
		},
		Status:    domain.TenantStatus(status),
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}

	return tenant, nil
}

func (r *PostgresTenantRepository) scanTenantFromRow(rows *sql.Rows) (*domain.Tenant, error) {
	var id, name, slug, addressStreet, addressCity, addressZip, addressCountry, logo, status string
	var defaultLanguage, currency, invoicePrefix string
	var taxRate float64
	var createdAt, updatedAt time.Time

	err := rows.Scan(&id, &name, &slug, &addressStreet, &addressCity, &addressZip, &addressCountry,
		&logo, &defaultLanguage, &currency, &taxRate, &invoicePrefix, &status, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}

	tenant := &domain.Tenant{
		AggregateRoot: domain.AggregateRoot{
			ID:      id,
			Type:    "Tenant",
			Changes: []interface{}{},
		},
		Name: name,
		Slug: slug,
		Address: domain.Address{
			Street:  addressStreet,
			City:    addressCity,
			ZIP:     addressZip,
			Country: addressCountry,
		},
		Logo: logo,
		Settings: domain.TenantSettings{
			DefaultLanguage: defaultLanguage,
			Currency:        currency,
			TaxRate:         taxRate,
			InvoicePrefix:   invoicePrefix,
		},
		Status:    domain.TenantStatus(status),
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}

	return tenant, nil
}

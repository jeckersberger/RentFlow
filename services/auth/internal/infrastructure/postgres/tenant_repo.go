package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jeckersberger/EquipFlow/services/auth/internal/domain"

	apperrors "github.com/jeckersberger/EquipFlow/pkg/common/errors"
)

// TenantRepo implements domain.TenantRepository using PostgreSQL.
type TenantRepo struct {
	pool *pgxpool.Pool
}

// NewTenantRepo creates a new TenantRepo.
func NewTenantRepo(pool *pgxpool.Pool) *TenantRepo {
	return &TenantRepo{pool: pool}
}

// Create inserts a new tenant and scans back the generated fields.
func (r *TenantRepo) Create(ctx context.Context, tenant *domain.Tenant) error {
	query := `
		INSERT INTO tenants (
			id, name, slug, email, phone,
			address_street, address_city, address_zip, address_country,
			tax_number, vat_id, iban, bic, bank_name,
			logo_url, currency, locale, timezone, industry,
			invoice_prefix, invoice_counter, settings, industry_profile, is_active
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9,
			$10, $11, $12, $13, $14,
			$15, $16, $17, $18, $19,
			$20, $21, $22, $23, $24
		) RETURNING id, created_at, updated_at`

	err := r.pool.QueryRow(ctx, query,
		tenant.ID, tenant.Name, tenant.Slug, tenant.Email, tenant.Phone,
		tenant.AddressStreet, tenant.AddressCity, tenant.AddressZip, tenant.AddressCountry,
		tenant.TaxNumber, tenant.VatID, tenant.IBAN, tenant.BIC, tenant.BankName,
		tenant.LogoURL, tenant.Currency, tenant.Locale, tenant.Timezone, tenant.Industry,
		tenant.InvoicePrefix, tenant.InvoiceCounter, tenant.Settings, tenant.IndustryProfile, tenant.IsActive,
	).Scan(&tenant.ID, &tenant.CreatedAt, &tenant.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrSlugTaken
		}
		return fmt.Errorf("tenant_repo: create: %w", err)
	}
	return nil
}

// GetByID retrieves a tenant by its primary key.
func (r *TenantRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
	query := `
		SELECT id, name, slug, email, phone,
			address_street, address_city, address_zip, address_country,
			tax_number, vat_id, iban, bic, bank_name,
			logo_url, currency, locale, timezone, industry,
			invoice_prefix, invoice_counter, settings, industry_profile,
			is_active, created_at, updated_at
		FROM tenants WHERE id = $1`

	t := &domain.Tenant{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&t.ID, &t.Name, &t.Slug, &t.Email, &t.Phone,
		&t.AddressStreet, &t.AddressCity, &t.AddressZip, &t.AddressCountry,
		&t.TaxNumber, &t.VatID, &t.IBAN, &t.BIC, &t.BankName,
		&t.LogoURL, &t.Currency, &t.Locale, &t.Timezone, &t.Industry,
		&t.InvoicePrefix, &t.InvoiceCounter, &t.Settings, &t.IndustryProfile,
		&t.IsActive, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("tenant_repo: get_by_id: %w", err)
	}
	return t, nil
}

// GetBySlug retrieves a tenant by its unique slug.
func (r *TenantRepo) GetBySlug(ctx context.Context, slug string) (*domain.Tenant, error) {
	query := `
		SELECT id, name, slug, email, phone,
			address_street, address_city, address_zip, address_country,
			tax_number, vat_id, iban, bic, bank_name,
			logo_url, currency, locale, timezone, industry,
			invoice_prefix, invoice_counter, settings, industry_profile,
			is_active, created_at, updated_at
		FROM tenants WHERE slug = $1`

	t := &domain.Tenant{}
	err := r.pool.QueryRow(ctx, query, slug).Scan(
		&t.ID, &t.Name, &t.Slug, &t.Email, &t.Phone,
		&t.AddressStreet, &t.AddressCity, &t.AddressZip, &t.AddressCountry,
		&t.TaxNumber, &t.VatID, &t.IBAN, &t.BIC, &t.BankName,
		&t.LogoURL, &t.Currency, &t.Locale, &t.Timezone, &t.Industry,
		&t.InvoicePrefix, &t.InvoiceCounter, &t.Settings, &t.IndustryProfile,
		&t.IsActive, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("tenant_repo: get_by_slug: %w", err)
	}
	return t, nil
}

// Update modifies an existing tenant.
func (r *TenantRepo) Update(ctx context.Context, tenant *domain.Tenant) error {
	query := `
		UPDATE tenants SET
			name = $2, slug = $3, email = $4, phone = $5,
			address_street = $6, address_city = $7, address_zip = $8, address_country = $9,
			tax_number = $10, vat_id = $11, iban = $12, bic = $13, bank_name = $14,
			logo_url = $15, currency = $16, locale = $17, timezone = $18, industry = $19,
			invoice_prefix = $20, invoice_counter = $21, settings = $22, industry_profile = $23,
			is_active = $24, updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at`

	err := r.pool.QueryRow(ctx, query,
		tenant.ID, tenant.Name, tenant.Slug, tenant.Email, tenant.Phone,
		tenant.AddressStreet, tenant.AddressCity, tenant.AddressZip, tenant.AddressCountry,
		tenant.TaxNumber, tenant.VatID, tenant.IBAN, tenant.BIC, tenant.BankName,
		tenant.LogoURL, tenant.Currency, tenant.Locale, tenant.Timezone, tenant.Industry,
		tenant.InvoicePrefix, tenant.InvoiceCounter, tenant.Settings, tenant.IndustryProfile,
		tenant.IsActive,
	).Scan(&tenant.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrSlugTaken
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return apperrors.ErrNotFound
		}
		return fmt.Errorf("tenant_repo: update: %w", err)
	}
	return nil
}

// SlugExists checks whether a tenant with the given slug already exists.
func (r *TenantRepo) SlugExists(ctx context.Context, slug string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM tenants WHERE slug = $1)`, slug,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("tenant_repo: slug_exists: %w", err)
	}
	return exists, nil
}

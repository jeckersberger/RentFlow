package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jeckersberger/EquipFlow/services/customer/internal/domain"

	apperrors "github.com/jeckersberger/EquipFlow/pkg/common/errors"
)

// customerColumns lists all columns of the customers table for consistent scanning.
const customerColumns = `
	id, tenant_id, company_name, customer_number, email, phone, website,
	billing_address_street, billing_address_city, billing_address_zip,
	billing_address_country, tax_id, notes, is_active,
	created_at, updated_at`

// CustomerRepo implements domain.CustomerRepository using PostgreSQL.
type CustomerRepo struct {
	pool *pgxpool.Pool
}

// NewCustomerRepo creates a new CustomerRepo.
func NewCustomerRepo(pool *pgxpool.Pool) *CustomerRepo {
	return &CustomerRepo{pool: pool}
}

// scanCustomer scans a single customer row into a domain.Customer, handling nullable columns.
func scanCustomer(row pgx.Row) (*domain.Customer, error) {
	c := &domain.Customer{}
	var (
		customerNumber        *string
		email                 *string
		phone                 *string
		website               *string
		billingAddressStreet  *string
		billingAddressCity    *string
		billingAddressZip     *string
		billingAddressCountry *string
		taxID                 *string
		notes                 *string
	)

	err := row.Scan(
		&c.ID, &c.TenantID, &c.CompanyName, &customerNumber, &email, &phone, &website,
		&billingAddressStreet, &billingAddressCity, &billingAddressZip,
		&billingAddressCountry, &taxID, &notes, &c.IsActive,
		&c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	c.CustomerNumber = derefString(customerNumber)
	c.Email = derefString(email)
	c.Phone = derefString(phone)
	c.Website = derefString(website)
	c.BillingAddressStreet = derefString(billingAddressStreet)
	c.BillingAddressCity = derefString(billingAddressCity)
	c.BillingAddressZip = derefString(billingAddressZip)
	c.BillingAddressCountry = derefString(billingAddressCountry)
	c.TaxID = derefString(taxID)
	c.Notes = derefString(notes)

	return c, nil
}

// Create inserts a new customer record and scans back the generated fields.
func (r *CustomerRepo) Create(ctx context.Context, customer *domain.Customer) error {
	query := `
		INSERT INTO customers (
			id, tenant_id, company_name, customer_number, email, phone, website,
			billing_address_street, billing_address_city, billing_address_zip,
			billing_address_country, tax_id, notes, is_active
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7,
			$8, $9, $10, $11, $12, $13, $14
		) RETURNING created_at, updated_at`

	err := r.pool.QueryRow(ctx, query,
		customer.ID, customer.TenantID, customer.CompanyName,
		nilIfEmpty(customer.CustomerNumber), nilIfEmpty(customer.Email),
		nilIfEmpty(customer.Phone), nilIfEmpty(customer.Website),
		nilIfEmpty(customer.BillingAddressStreet), nilIfEmpty(customer.BillingAddressCity),
		nilIfEmpty(customer.BillingAddressZip), nilIfEmpty(customer.BillingAddressCountry),
		nilIfEmpty(customer.TaxID), nilIfEmpty(customer.Notes), customer.IsActive,
	).Scan(&customer.CreatedAt, &customer.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return apperrors.ErrConflict
		}
		return fmt.Errorf("customer_repo: create: %w", err)
	}
	return nil
}

// GetByID retrieves a single customer record by primary key scoped to a tenant.
func (r *CustomerRepo) GetByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*domain.Customer, error) {
	query := fmt.Sprintf(`SELECT %s FROM customers WHERE id = $1 AND tenant_id = $2`, customerColumns)
	c, err := scanCustomer(r.pool.QueryRow(ctx, query, id, tenantID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("customer_repo: get_by_id: %w", err)
	}
	return c, nil
}

// List returns a filtered, paginated list of customers for a tenant plus total count.
func (r *CustomerRepo) List(ctx context.Context, tenantID uuid.UUID, filter domain.CustomerFilter) ([]*domain.Customer, int64, error) {
	// Build dynamic WHERE clause.
	conditions := []string{"tenant_id = $1"}
	args := []interface{}{tenantID}
	argIdx := 2

	if filter.Active != nil {
		conditions = append(conditions, fmt.Sprintf("is_active = $%d", argIdx))
		args = append(args, *filter.Active)
		argIdx++
	}
	if filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf(
			"(company_name ILIKE $%d OR contact_first_name ILIKE $%d OR contact_last_name ILIKE $%d OR contact_email ILIKE $%d)",
			argIdx, argIdx, argIdx, argIdx,
		))
		args = append(args, "%"+filter.Search+"%")
		argIdx++
	}

	where := strings.Join(conditions, " AND ")

	// Count query.
	var total int64
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM customers WHERE %s`, where)
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("customer_repo: list count: %w", err)
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
		`SELECT %s FROM customers WHERE %s ORDER BY company_name ASC LIMIT $%d OFFSET $%d`,
		customerColumns, where, argIdx, argIdx+1,
	)
	args = append(args, perPage, offset)

	rows, err := r.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("customer_repo: list query: %w", err)
	}
	defer rows.Close()

	var items []*domain.Customer
	for rows.Next() {
		c, scanErr := scanCustomer(rows)
		if scanErr != nil {
			return nil, 0, fmt.Errorf("customer_repo: list scan: %w", scanErr)
		}
		items = append(items, c)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("customer_repo: list rows: %w", err)
	}
	return items, total, nil
}

// Update modifies an existing customer record.
func (r *CustomerRepo) Update(ctx context.Context, customer *domain.Customer) error {
	query := `
		UPDATE customers SET
			company_name = $3, customer_number = $4, email = $5, phone = $6, website = $7,
			billing_address_street = $8, billing_address_city = $9, billing_address_zip = $10,
			billing_address_country = $11, tax_id = $12, notes = $13,
			updated_at = NOW()
		WHERE id = $1 AND tenant_id = $2
		RETURNING updated_at`

	err := r.pool.QueryRow(ctx, query,
		customer.ID, customer.TenantID,
		customer.CompanyName, nilIfEmpty(customer.CustomerNumber),
		nilIfEmpty(customer.Email), nilIfEmpty(customer.Phone), nilIfEmpty(customer.Website),
		nilIfEmpty(customer.BillingAddressStreet), nilIfEmpty(customer.BillingAddressCity),
		nilIfEmpty(customer.BillingAddressZip), nilIfEmpty(customer.BillingAddressCountry),
		nilIfEmpty(customer.TaxID), nilIfEmpty(customer.Notes),
	).Scan(&customer.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apperrors.ErrNotFound
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return apperrors.ErrConflict
		}
		return fmt.Errorf("customer_repo: update: %w", err)
	}
	return nil
}

// Deactivate sets is_active=false for a customer (soft delete).
func (r *CustomerRepo) Delete(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE customers SET is_active = FALSE, updated_at = NOW() WHERE id = $1 AND tenant_id = $2 AND is_active = TRUE`,
		id, tenantID,
	)
	if err != nil {
		return fmt.Errorf("customer_repo: delete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

// Search performs an ILIKE search on customer company_name, customer_number, and email with pagination.
func (r *CustomerRepo) Search(ctx context.Context, tenantID uuid.UUID, query string, page int, perPage int) ([]*domain.Customer, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	offset := (page - 1) * perPage
	pattern := "%" + query + "%"

	searchCondition := `tenant_id = $1 AND (company_name ILIKE $2 OR customer_number ILIKE $2 OR email ILIKE $2)`

	// Count query.
	var total int64
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM customers WHERE %s`, searchCondition)
	err := r.pool.QueryRow(ctx, countQuery, tenantID, pattern).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("customer_repo: search count: %w", err)
	}

	// Data query.
	dataQuery := fmt.Sprintf(
		`SELECT %s FROM customers WHERE %s ORDER BY company_name ASC LIMIT $3 OFFSET $4`,
		customerColumns, searchCondition,
	)

	rows, err := r.pool.Query(ctx, dataQuery, tenantID, pattern, perPage, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("customer_repo: search query: %w", err)
	}
	defer rows.Close()

	var items []*domain.Customer
	for rows.Next() {
		c, scanErr := scanCustomer(rows)
		if scanErr != nil {
			return nil, 0, fmt.Errorf("customer_repo: search scan: %w", scanErr)
		}
		items = append(items, c)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("customer_repo: search rows: %w", err)
	}
	return items, total, nil
}

// nilIfEmpty returns nil if the string is empty, otherwise a pointer to it.
func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// derefString safely dereferences a *string, returning "" if nil.
func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jeckersberger/rentflow/pkg/common/database"
	"github.com/jeckersberger/rentflow/services/project-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/project-service/internal/ports"
)

type CustomerPostgres struct {
	db *database.PostgresPool
}

func NewCustomerPostgres(db *database.PostgresPool) *CustomerPostgres {
	return &CustomerPostgres{db: db}
}

func (r *CustomerPostgres) Create(ctx context.Context, customer *domain.Customer) error {
	query := `
		INSERT INTO projects.customers (
			id, tenant_id, name, email, phone, address_street, address_city,
			address_postcode, address_country, tax_id, notes, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		)
	`

	_, err := r.db.Exec(ctx, query,
		customer.ID, customer.TenantID, customer.Name, customer.Email, customer.Phone,
		customer.AddressStreet, customer.AddressCity, customer.AddressPostcode,
		customer.AddressCountry, customer.TaxID, customer.Notes,
		customer.CreatedAt, customer.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create customer: %w", err)
	}

	return nil
}

func (r *CustomerPostgres) GetByID(ctx context.Context, tenantID, customerID string) (*domain.Customer, error) {
	query := `
		SELECT id, tenant_id, name, email, phone, address_street, address_city,
		       address_postcode, address_country, tax_id, notes, created_at, updated_at
		FROM projects.customers
		WHERE id = $1 AND tenant_id = $2
	`

	c := &domain.Customer{}
	err := r.db.QueryRow(ctx, query, customerID, tenantID).Scan(
		&c.ID, &c.TenantID, &c.Name, &c.Email, &c.Phone,
		&c.AddressStreet, &c.AddressCity, &c.AddressPostcode,
		&c.AddressCountry, &c.TaxID, &c.Notes, &c.CreatedAt, &c.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}

	return c, nil
}

func (r *CustomerPostgres) List(ctx context.Context, tenantID string, limit, offset int) (*ports.CustomerListResult, error) {
	query := `
		SELECT id, tenant_id, name, email, phone, address_street, address_city,
		       address_postcode, address_country, tax_id, notes, created_at, updated_at
		FROM projects.customers
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	countQuery := `SELECT COUNT(*) FROM projects.customers WHERE tenant_id = $1`

	var count int64
	err := r.db.QueryRow(ctx, countQuery, tenantID).Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("failed to count customers: %w", err)
	}

	rows, err := r.db.Query(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list customers: %w", err)
	}
	defer rows.Close()

	customers := make([]*domain.Customer, 0)
	for rows.Next() {
		c := &domain.Customer{}
		err := rows.Scan(
			&c.ID, &c.TenantID, &c.Name, &c.Email, &c.Phone,
			&c.AddressStreet, &c.AddressCity, &c.AddressPostcode,
			&c.AddressCountry, &c.TaxID, &c.Notes, &c.CreatedAt, &c.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan customer: %w", err)
		}
		customers = append(customers, c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating customers: %w", err)
	}

	return &ports.CustomerListResult{
		Items:  customers,
		Total:  count,
		Limit:  limit,
		Offset: offset,
	}, nil
}

func (r *CustomerPostgres) Update(ctx context.Context, customer *domain.Customer) error {
	query := `
		UPDATE projects.customers SET
			name = $3, email = $4, phone = $5, address_street = $6,
			address_city = $7, address_postcode = $8, address_country = $9,
			tax_id = $10, notes = $11, updated_at = $12
		WHERE id = $1 AND tenant_id = $2
	`

	result, err := r.db.Exec(ctx, query,
		customer.ID, customer.TenantID, customer.Name, customer.Email, customer.Phone,
		customer.AddressStreet, customer.AddressCity, customer.AddressPostcode,
		customer.AddressCountry, customer.TaxID, customer.Notes, customer.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update customer: %w", err)
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

func (r *CustomerPostgres) Delete(ctx context.Context, tenantID, customerID string) error {
	query := `DELETE FROM projects.customers WHERE id = $1 AND tenant_id = $2`

	result, err := r.db.Exec(ctx, query, customerID, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete customer: %w", err)
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

func (r *CustomerPostgres) Search(ctx context.Context, tenantID, term string, limit, offset int) (*ports.CustomerListResult, error) {
	query := `
		SELECT id, tenant_id, name, email, phone, address_street, address_city,
		       address_postcode, address_country, tax_id, notes, created_at, updated_at
		FROM projects.customers
		WHERE tenant_id = $1 AND (name ILIKE $2 OR email ILIKE $2 OR phone ILIKE $2)
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`

	countQuery := `
		SELECT COUNT(*) FROM projects.customers
		WHERE tenant_id = $1 AND (name ILIKE $2 OR email ILIKE $2 OR phone ILIKE $2)
	`

	searchTerm := "%" + term + "%"

	var count int64
	err := r.db.QueryRow(ctx, countQuery, tenantID, searchTerm).Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("failed to count customers: %w", err)
	}

	rows, err := r.db.Query(ctx, query, tenantID, searchTerm, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search customers: %w", err)
	}
	defer rows.Close()

	customers := make([]*domain.Customer, 0)
	for rows.Next() {
		c := &domain.Customer{}
		err := rows.Scan(
			&c.ID, &c.TenantID, &c.Name, &c.Email, &c.Phone,
			&c.AddressStreet, &c.AddressCity, &c.AddressPostcode,
			&c.AddressCountry, &c.TaxID, &c.Notes, &c.CreatedAt, &c.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan customer: %w", err)
		}
		customers = append(customers, c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating customers: %w", err)
	}

	return &ports.CustomerListResult{
		Items:  customers,
		Total:  count,
		Limit:  limit,
		Offset: offset,
	}, nil
}

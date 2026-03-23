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

type ContactPostgres struct {
	db *database.PostgresPool
}

func NewContactPostgres(db *database.PostgresPool) *ContactPostgres {
	return &ContactPostgres{db: db}
}

func (r *ContactPostgres) Create(ctx context.Context, contact *domain.Contact) error {
	query := `
		INSERT INTO projects.contacts (
			id, tenant_id, type, company_name, first_name, last_name, email, phone,
			mobile, website, street, house_number, zip, city, country, vat_id,
			notes, tags, created_at, updated_at, created_by
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15,
			$16, $17, $18, $19, $20, $21
		)
	`

	_, err := r.db.Exec(ctx, query,
		contact.ID, contact.TenantID, contact.Type, contact.CompanyName,
		contact.FirstName, contact.LastName, contact.Email, contact.Phone,
		contact.Mobile, contact.Website, contact.Street, contact.HouseNumber,
		contact.Zip, contact.City, contact.Country, contact.VatID,
		contact.Notes, pq.Array(contact.Tags), contact.CreatedAt, contact.UpdatedAt,
		nullString(contact.CreatedBy),
	)

	if err != nil {
		return fmt.Errorf("failed to create contact: %w", err)
	}

	return nil
}

func (r *ContactPostgres) GetByID(ctx context.Context, tenantID, contactID string) (*domain.Contact, error) {
	query := `
		SELECT id, tenant_id, type, company_name, first_name, last_name, email, phone,
		       mobile, website, street, house_number, zip, city, country, vat_id,
		       notes, tags, created_at, updated_at, created_by
		FROM projects.contacts
		WHERE id = $1 AND tenant_id = $2
	`

	c := &domain.Contact{}
	var tags pq.StringArray
	var createdBy sql.NullString
	err := r.db.QueryRow(ctx, query, contactID, tenantID).Scan(
		&c.ID, &c.TenantID, &c.Type, &c.CompanyName, &c.FirstName, &c.LastName,
		&c.Email, &c.Phone, &c.Mobile, &c.Website, &c.Street, &c.HouseNumber,
		&c.Zip, &c.City, &c.Country, &c.VatID, &c.Notes, &tags,
		&c.CreatedAt, &c.UpdatedAt, &createdBy,
	)

	if err != nil {
		return nil, err
	}

	c.Tags = []string(tags)
	if createdBy.Valid {
		c.CreatedBy = createdBy.String
	}

	return c, nil
}

func (r *ContactPostgres) List(ctx context.Context, tenantID string, limit, offset int) (*ports.ContactListResult, error) {
	countQuery := `SELECT COUNT(*) FROM projects.contacts WHERE tenant_id = $1`
	var total int64
	if err := r.db.QueryRow(ctx, countQuery, tenantID).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count contacts: %w", err)
	}

	query := `
		SELECT id, tenant_id, type, company_name, first_name, last_name, email, phone,
		       mobile, website, street, house_number, zip, city, country, vat_id,
		       notes, tags, created_at, updated_at, created_by
		FROM projects.contacts
		WHERE tenant_id = $1
		ORDER BY updated_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list contacts: %w", err)
	}
	defer rows.Close()

	var contacts []*domain.Contact
	for rows.Next() {
		c := &domain.Contact{}
		var tags pq.StringArray
		var createdBy sql.NullString
		if err := rows.Scan(
			&c.ID, &c.TenantID, &c.Type, &c.CompanyName, &c.FirstName, &c.LastName,
			&c.Email, &c.Phone, &c.Mobile, &c.Website, &c.Street, &c.HouseNumber,
			&c.Zip, &c.City, &c.Country, &c.VatID, &c.Notes, &tags,
			&c.CreatedAt, &c.UpdatedAt, &createdBy,
		); err != nil {
			return nil, fmt.Errorf("failed to scan contact: %w", err)
		}
		c.Tags = []string(tags)
		if createdBy.Valid {
			c.CreatedBy = createdBy.String
		}
		contacts = append(contacts, c)
	}

	return &ports.ContactListResult{
		Items:  contacts,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}, nil
}

func (r *ContactPostgres) Update(ctx context.Context, contact *domain.Contact) error {
	query := `
		UPDATE projects.contacts SET
			type = $3, company_name = $4, first_name = $5, last_name = $6,
			email = $7, phone = $8, mobile = $9, website = $10,
			street = $11, house_number = $12, zip = $13, city = $14,
			country = $15, vat_id = $16, notes = $17, tags = $18,
			updated_at = $19
		WHERE id = $1 AND tenant_id = $2
	`

	result, err := r.db.Exec(ctx, query,
		contact.ID, contact.TenantID, contact.Type, contact.CompanyName,
		contact.FirstName, contact.LastName, contact.Email, contact.Phone,
		contact.Mobile, contact.Website, contact.Street, contact.HouseNumber,
		contact.Zip, contact.City, contact.Country, contact.VatID,
		contact.Notes, pq.Array(contact.Tags), contact.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update contact: %w", err)
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

func (r *ContactPostgres) Delete(ctx context.Context, tenantID, contactID string) error {
	query := `DELETE FROM projects.contacts WHERE id = $1 AND tenant_id = $2`
	result, err := r.db.Exec(ctx, query, contactID, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete contact: %w", err)
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

func (r *ContactPostgres) Search(ctx context.Context, tenantID, term string, limit, offset int) (*ports.ContactListResult, error) {
	searchPattern := "%" + term + "%"

	countQuery := `
		SELECT COUNT(*) FROM projects.contacts
		WHERE tenant_id = $1 AND (
			company_name ILIKE $2 OR first_name ILIKE $2 OR last_name ILIKE $2 OR
			email ILIKE $2 OR city ILIKE $2
		)
	`
	var total int64
	if err := r.db.QueryRow(ctx, countQuery, tenantID, searchPattern).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count search results: %w", err)
	}

	query := `
		SELECT id, tenant_id, type, company_name, first_name, last_name, email, phone,
		       mobile, website, street, house_number, zip, city, country, vat_id,
		       notes, tags, created_at, updated_at, created_by
		FROM projects.contacts
		WHERE tenant_id = $1 AND (
			company_name ILIKE $2 OR first_name ILIKE $2 OR last_name ILIKE $2 OR
			email ILIKE $2 OR city ILIKE $2
		)
		ORDER BY updated_at DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.db.Query(ctx, query, tenantID, searchPattern, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search contacts: %w", err)
	}
	defer rows.Close()

	var contacts []*domain.Contact
	for rows.Next() {
		c := &domain.Contact{}
		var tags pq.StringArray
		var createdBy sql.NullString
		if err := rows.Scan(
			&c.ID, &c.TenantID, &c.Type, &c.CompanyName, &c.FirstName, &c.LastName,
			&c.Email, &c.Phone, &c.Mobile, &c.Website, &c.Street, &c.HouseNumber,
			&c.Zip, &c.City, &c.Country, &c.VatID, &c.Notes, &tags,
			&c.CreatedAt, &c.UpdatedAt, &createdBy,
		); err != nil {
			return nil, fmt.Errorf("failed to scan contact: %w", err)
		}
		c.Tags = []string(tags)
		if createdBy.Valid {
			c.CreatedBy = createdBy.String
		}
		contacts = append(contacts, c)
	}

	return &ports.ContactListResult{
		Items:  contacts,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}, nil
}

func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

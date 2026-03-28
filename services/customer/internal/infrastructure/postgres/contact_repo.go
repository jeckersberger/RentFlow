package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jeckersberger/EquipFlow/services/customer/internal/domain"

	apperrors "github.com/jeckersberger/EquipFlow/pkg/common/errors"
)

const contactColumns = `
	id, tenant_id, customer_id, first_name, last_name, email, phone, mobile,
	position, is_primary, notes, created_at, updated_at`

type ContactRepo struct {
	pool *pgxpool.Pool
}

func NewContactRepo(pool *pgxpool.Pool) *ContactRepo {
	return &ContactRepo{pool: pool}
}

func scanContact(row pgx.Row) (*domain.Contact, error) {
	c := &domain.Contact{}
	var (
		email    *string
		phone    *string
		mobile   *string
		position *string
		notes    *string
	)

	err := row.Scan(
		&c.ID, &c.TenantID, &c.CustomerID, &c.FirstName, &c.LastName,
		&email, &phone, &mobile, &position, &c.IsPrimary, &notes,
		&c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	c.Email = derefString(email)
	c.Phone = derefString(phone)
	c.Mobile = derefString(mobile)
	c.Position = derefString(position)
	c.Notes = derefString(notes)

	return c, nil
}

func (r *ContactRepo) Create(ctx context.Context, contact *domain.Contact) error {
	query := `
		INSERT INTO contacts (
			id, tenant_id, customer_id, first_name, last_name, email, phone, mobile,
			position, is_primary, notes
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING created_at, updated_at`

	err := r.pool.QueryRow(ctx, query,
		contact.ID, contact.TenantID, contact.CustomerID,
		contact.FirstName, contact.LastName,
		nilIfEmpty(contact.Email), nilIfEmpty(contact.Phone), nilIfEmpty(contact.Mobile),
		nilIfEmpty(contact.Position), contact.IsPrimary, nilIfEmpty(contact.Notes),
	).Scan(&contact.CreatedAt, &contact.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return apperrors.Wrap(apperrors.ErrBadRequest, "Kunde existiert nicht")
		}
		return fmt.Errorf("contact_repo: create: %w", err)
	}
	return nil
}

func (r *ContactRepo) GetByID(ctx context.Context, id, tenantID uuid.UUID) (*domain.Contact, error) {
	query := fmt.Sprintf(`SELECT %s FROM contacts WHERE id = $1 AND tenant_id = $2`, contactColumns)
	c, err := scanContact(r.pool.QueryRow(ctx, query, id, tenantID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("contact_repo: get_by_id: %w", err)
	}
	return c, nil
}

func (r *ContactRepo) ListByCustomer(ctx context.Context, customerID uuid.UUID, tenantID uuid.UUID) ([]*domain.Contact, error) {
	query := fmt.Sprintf(`SELECT %s FROM contacts WHERE customer_id = $1 AND tenant_id = $2 ORDER BY is_primary DESC, last_name ASC`, contactColumns)
	rows, err := r.pool.Query(ctx, query, customerID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("contact_repo: list_by_customer query: %w", err)
	}
	defer rows.Close()

	var items []*domain.Contact
	for rows.Next() {
		c, scanErr := scanContact(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("contact_repo: list_by_customer scan: %w", scanErr)
		}
		items = append(items, c)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("contact_repo: list_by_customer rows: %w", err)
	}
	return items, nil
}

func (r *ContactRepo) Update(ctx context.Context, contact *domain.Contact) error {
	query := `
		UPDATE contacts SET
			first_name = $3, last_name = $4, email = $5, phone = $6, mobile = $7,
			position = $8, is_primary = $9, notes = $10,
			updated_at = NOW()
		WHERE id = $1 AND tenant_id = $2
		RETURNING updated_at`

	err := r.pool.QueryRow(ctx, query,
		contact.ID, contact.TenantID,
		contact.FirstName, contact.LastName,
		nilIfEmpty(contact.Email), nilIfEmpty(contact.Phone), nilIfEmpty(contact.Mobile),
		nilIfEmpty(contact.Position), contact.IsPrimary, nilIfEmpty(contact.Notes),
	).Scan(&contact.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apperrors.ErrNotFound
		}
		return fmt.Errorf("contact_repo: update: %w", err)
	}
	return nil
}

func (r *ContactRepo) Delete(ctx context.Context, id, tenantID uuid.UUID) error {
	tag, err := r.pool.Exec(ctx,
		`DELETE FROM contacts WHERE id = $1 AND tenant_id = $2`,
		id, tenantID,
	)
	if err != nil {
		return fmt.Errorf("contact_repo: delete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

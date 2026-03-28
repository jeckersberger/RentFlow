package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jeckersberger/EquipFlow/services/customer/internal/domain"
)

type NoteRepo struct {
	pool *pgxpool.Pool
}

func NewNoteRepo(pool *pgxpool.Pool) *NoteRepo {
	return &NoteRepo{pool: pool}
}

func scanNote(row pgx.Row) (*domain.ContactNote, error) {
	n := &domain.ContactNote{}
	var createdBy *uuid.UUID

	err := row.Scan(&n.ID, &n.TenantID, &n.ContactID, &n.Content, &createdBy, &n.CreatedAt)
	if err != nil {
		return nil, err
	}

	n.CreatedBy = createdBy
	return n, nil
}

func (r *NoteRepo) Create(ctx context.Context, note *domain.ContactNote) error {
	query := `
		INSERT INTO contact_notes (id, tenant_id, contact_id, content, created_by)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at`

	err := r.pool.QueryRow(ctx, query,
		note.ID, note.TenantID, note.ContactID, note.Content, note.CreatedBy,
	).Scan(&note.CreatedAt)
	if err != nil {
		return fmt.Errorf("note_repo: create: %w", err)
	}
	return nil
}

func (r *NoteRepo) ListByContact(ctx context.Context, contactID uuid.UUID, tenantID uuid.UUID) ([]*domain.ContactNote, error) {
	query := `SELECT id, tenant_id, contact_id, content, created_by, created_at
		FROM contact_notes WHERE contact_id = $1 AND tenant_id = $2 ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, query, contactID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("note_repo: list_by_contact query: %w", err)
	}
	defer rows.Close()

	var items []*domain.ContactNote
	for rows.Next() {
		n, scanErr := scanNote(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("note_repo: list_by_contact scan: %w", scanErr)
		}
		items = append(items, n)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("note_repo: list_by_contact rows: %w", err)
	}
	return items, nil
}

package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jeckersberger/rentflow/pkg/common/database"
	"github.com/jeckersberger/rentflow/services/invoice-service/internal/domain"
)

type DunningPostgres struct {
	db *database.PostgresPool
}

func NewDunningPostgres(db *database.PostgresPool) *DunningPostgres {
	return &DunningPostgres{db: db}
}

func (r *DunningPostgres) Create(ctx context.Context, dunning *domain.DunningEntry) error {
	query := `
		INSERT INTO invoice.dunning (
			id, tenant_id, invoice_id, invoice_number, level, sent_at, due_date, fee, notes, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := r.db.Exec(ctx, query,
		dunning.ID, dunning.TenantID, dunning.InvoiceID, dunning.InvoiceNumber,
		dunning.Level, dunning.SentAt, dunning.DueDate, dunning.Fee,
		dunning.Notes, dunning.CreatedAt, dunning.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create dunning entry: %w", err)
	}

	return nil
}

func (r *DunningPostgres) GetByID(ctx context.Context, tenantID, dunningID string) (*domain.DunningEntry, error) {
	query := `
		SELECT id, tenant_id, invoice_id, invoice_number, level, sent_at, due_date, fee, notes, created_at, updated_at
		FROM invoice.dunning
		WHERE id = $1 AND tenant_id = $2
	`

	row := r.db.QueryRow(ctx, query, dunningID, tenantID)

	dunning := &domain.DunningEntry{}
	err := row.Scan(
		&dunning.ID, &dunning.TenantID, &dunning.InvoiceID, &dunning.InvoiceNumber,
		&dunning.Level, &dunning.SentAt, &dunning.DueDate, &dunning.Fee,
		&dunning.Notes, &dunning.CreatedAt, &dunning.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.ErrDunningNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get dunning entry: %w", err)
	}

	return dunning, nil
}

func (r *DunningPostgres) List(ctx context.Context, tenantID, invoiceID string) ([]*domain.DunningEntry, error) {
	query := `
		SELECT id, tenant_id, invoice_id, invoice_number, level, sent_at, due_date, fee, notes, created_at, updated_at
		FROM invoice.dunning
		WHERE tenant_id = $1 AND invoice_id = $2
		ORDER BY level DESC, created_at DESC
	`

	rows, err := r.db.Query(ctx, query, tenantID, invoiceID)
	if err != nil {
		return nil, fmt.Errorf("failed to list dunning entries: %w", err)
	}
	defer rows.Close()

	var entries []*domain.DunningEntry
	for rows.Next() {
		dunning := &domain.DunningEntry{}
		err := rows.Scan(
			&dunning.ID, &dunning.TenantID, &dunning.InvoiceID, &dunning.InvoiceNumber,
			&dunning.Level, &dunning.SentAt, &dunning.DueDate, &dunning.Fee,
			&dunning.Notes, &dunning.CreatedAt, &dunning.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan dunning entry: %w", err)
		}
		entries = append(entries, dunning)
	}

	return entries, nil
}

func (r *DunningPostgres) Update(ctx context.Context, dunning *domain.DunningEntry) error {
	query := `
		UPDATE invoice.dunning SET
			sent_at = $3, due_date = $4, fee = $5, notes = $6, updated_at = $7
		WHERE id = $1 AND tenant_id = $2
	`

	_, err := r.db.Exec(ctx, query,
		dunning.ID, dunning.TenantID, dunning.SentAt, dunning.DueDate,
		dunning.Fee, dunning.Notes, dunning.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update dunning entry: %w", err)
	}

	return nil
}

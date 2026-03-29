package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	apperrors "github.com/jeckersberger/EquipFlow/pkg/common/errors"
	"github.com/jeckersberger/EquipFlow/services/invoice/internal/domain"
)

// DunningRepo implements domain.DunningRepository using PostgreSQL.
type DunningRepo struct {
	pool *pgxpool.Pool
}

// NewDunningRepo creates a new DunningRepo.
func NewDunningRepo(pool *pgxpool.Pool) *DunningRepo {
	return &DunningRepo{pool: pool}
}

// Create inserts a new dunning entry.
func (r *DunningRepo) Create(ctx context.Context, entry *domain.DunningEntry) error {
	query := `
		INSERT INTO dunning_entries (id, tenant_id, invoice_id, level, fee_cents, sent_at, status, notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING created_at`

	err := r.pool.QueryRow(ctx, query,
		entry.ID, entry.TenantID, entry.InvoiceID,
		entry.Level, entry.FeeCents, entry.SentAt,
		entry.Status, nilIfEmpty(entry.Notes),
	).Scan(&entry.CreatedAt)
	if err != nil {
		return fmt.Errorf("dunning_repo: create: %w", err)
	}
	return nil
}

// ListByInvoice returns all dunning entries for a given invoice.
func (r *DunningRepo) ListByInvoice(ctx context.Context, invoiceID uuid.UUID, tenantID uuid.UUID) ([]*domain.DunningEntry, error) {
	query := `
		SELECT id, tenant_id, invoice_id, level, fee_cents, sent_at, status, notes, created_at
		FROM dunning_entries
		WHERE invoice_id = $1 AND tenant_id = $2
		ORDER BY created_at ASC`

	rows, err := r.pool.Query(ctx, query, invoiceID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("dunning_repo: list_by_invoice: %w", err)
	}
	defer rows.Close()

	var entries []*domain.DunningEntry
	for rows.Next() {
		e := &domain.DunningEntry{}
		var notes *string
		err := rows.Scan(
			&e.ID, &e.TenantID, &e.InvoiceID,
			&e.Level, &e.FeeCents, &e.SentAt,
			&e.Status, &notes, &e.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("dunning_repo: list_by_invoice scan: %w", err)
		}
		e.Notes = derefString(notes)
		entries = append(entries, e)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("dunning_repo: list_by_invoice rows: %w", err)
	}
	return entries, nil
}

// ListOverdue returns sent invoices that are past their due date.
func (r *DunningRepo) ListOverdue(ctx context.Context, tenantID uuid.UUID) ([]*domain.Invoice, error) {
	query := fmt.Sprintf(`
		SELECT %s FROM invoices
		WHERE tenant_id = $1
		  AND status IN ('sent', 'partial_paid')
		  AND due_date IS NOT NULL
		  AND due_date < $2
		ORDER BY due_date ASC`, invoiceColumns)

	now := time.Now().Format("2006-01-02")
	rows, err := r.pool.Query(ctx, query, tenantID, now)
	if err != nil {
		return nil, fmt.Errorf("dunning_repo: list_overdue: %w", err)
	}
	defer rows.Close()

	var invoices []*domain.Invoice
	for rows.Next() {
		inv, scanErr := scanInvoice(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("dunning_repo: list_overdue scan: %w", scanErr)
		}
		invoices = append(invoices, inv)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("dunning_repo: list_overdue rows: %w", err)
	}
	return invoices, nil
}

// GetConfig retrieves the dunning configuration for a tenant.
func (r *DunningRepo) GetConfig(ctx context.Context, tenantID uuid.UUID) (*domain.DunningConfig, error) {
	query := `
		SELECT id, tenant_id, reminder_days, dunning1_days, dunning2_days,
		       reminder_fee, dunning1_fee, dunning2_fee, auto_send,
		       created_at, updated_at
		FROM dunning_config
		WHERE tenant_id = $1`

	cfg := &domain.DunningConfig{}
	err := r.pool.QueryRow(ctx, query, tenantID).Scan(
		&cfg.ID, &cfg.TenantID,
		&cfg.ReminderDays, &cfg.Dunning1Days, &cfg.Dunning2Days,
		&cfg.ReminderFee, &cfg.Dunning1Fee, &cfg.Dunning2Fee,
		&cfg.AutoSend,
		&cfg.CreatedAt, &cfg.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("dunning_repo: get_config: %w", err)
	}
	return cfg, nil
}

// UpsertConfig creates or updates the dunning configuration for a tenant.
func (r *DunningRepo) UpsertConfig(ctx context.Context, config *domain.DunningConfig) error {
	query := `
		INSERT INTO dunning_config (
			id, tenant_id, reminder_days, dunning1_days, dunning2_days,
			reminder_fee, dunning1_fee, dunning2_fee, auto_send
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (tenant_id) DO UPDATE SET
			reminder_days = EXCLUDED.reminder_days,
			dunning1_days = EXCLUDED.dunning1_days,
			dunning2_days = EXCLUDED.dunning2_days,
			reminder_fee  = EXCLUDED.reminder_fee,
			dunning1_fee  = EXCLUDED.dunning1_fee,
			dunning2_fee  = EXCLUDED.dunning2_fee,
			auto_send     = EXCLUDED.auto_send,
			updated_at    = NOW()
		RETURNING id, created_at, updated_at`

	err := r.pool.QueryRow(ctx, query,
		config.ID, config.TenantID,
		config.ReminderDays, config.Dunning1Days, config.Dunning2Days,
		config.ReminderFee, config.Dunning1Fee, config.Dunning2Fee,
		config.AutoSend,
	).Scan(&config.ID, &config.CreatedAt, &config.UpdatedAt)
	if err != nil {
		return fmt.Errorf("dunning_repo: upsert_config: %w", err)
	}
	return nil
}

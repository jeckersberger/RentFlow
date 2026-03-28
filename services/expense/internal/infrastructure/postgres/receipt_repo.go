package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jeckersberger/EquipFlow/services/expense/internal/domain"
)

type ReceiptRepo struct {
	pool *pgxpool.Pool
}

func NewReceiptRepo(pool *pgxpool.Pool) *ReceiptRepo {
	return &ReceiptRepo{pool: pool}
}

func (r *ReceiptRepo) Create(ctx context.Context, receipt *domain.ExpenseReceipt) error {
	query := `
		INSERT INTO expense_receipts (id, expense_id, tenant_id, file_name, file_path, file_size, mime_type, uploaded_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING created_at`

	err := r.pool.QueryRow(ctx, query,
		receipt.ID, receipt.ExpenseID, receipt.TenantID,
		receipt.FileName, receipt.FilePath, receipt.FileSize,
		nilIfEmpty(receipt.MimeType), nilUUID(receipt.UploadedBy),
	).Scan(&receipt.CreatedAt)
	if err != nil {
		return fmt.Errorf("receipt_repo: create: %w", err)
	}
	return nil
}

func (r *ReceiptRepo) ListByExpense(ctx context.Context, expenseID uuid.UUID, tenantID uuid.UUID) ([]*domain.ExpenseReceipt, error) {
	query := `SELECT id, expense_id, tenant_id, file_name, file_path, file_size, mime_type, uploaded_by, created_at
		FROM expense_receipts WHERE expense_id = $1 AND tenant_id = $2 ORDER BY created_at ASC`

	rows, err := r.pool.Query(ctx, query, expenseID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("receipt_repo: list: %w", err)
	}
	defer rows.Close()

	var items []*domain.ExpenseReceipt
	for rows.Next() {
		rec := &domain.ExpenseReceipt{}
		var mimeType *string
		var uploadedBy *uuid.UUID
		if err := rows.Scan(
			&rec.ID, &rec.ExpenseID, &rec.TenantID,
			&rec.FileName, &rec.FilePath, &rec.FileSize,
			&mimeType, &uploadedBy, &rec.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("receipt_repo: list scan: %w", err)
		}
		rec.MimeType = derefString(mimeType)
		if uploadedBy != nil {
			rec.UploadedBy = *uploadedBy
		}
		items = append(items, rec)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("receipt_repo: list rows: %w", err)
	}
	return items, nil
}

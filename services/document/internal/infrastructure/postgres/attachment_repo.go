package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jeckersberger/EquipFlow/services/document/internal/domain"

	apperrors "github.com/jeckersberger/EquipFlow/pkg/common/errors"
)

// attachmentColumns lists all columns of the attachments table for consistent scanning.
const attachmentColumns = `
	id, tenant_id, reference_id, reference_type, file_name, file_path,
	file_size, mime_type, uploaded_by, created_at`

// AttachmentRepo implements domain.AttachmentRepository using PostgreSQL.
type AttachmentRepo struct {
	pool *pgxpool.Pool
}

// NewAttachmentRepo creates a new AttachmentRepo.
func NewAttachmentRepo(pool *pgxpool.Pool) *AttachmentRepo {
	return &AttachmentRepo{pool: pool}
}

// scanAttachment scans a single attachments row into a domain.Attachment.
func scanAttachment(row pgx.Row) (*domain.Attachment, error) {
	a := &domain.Attachment{}
	var mimeType *string

	err := row.Scan(
		&a.ID, &a.TenantID, &a.ReferenceID, &a.ReferenceType,
		&a.FileName, &a.FilePath, &a.FileSize, &mimeType,
		&a.UploadedBy, &a.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	if mimeType != nil {
		a.MimeType = *mimeType
	}

	return a, nil
}

// Create inserts a new attachment.
func (r *AttachmentRepo) Create(ctx context.Context, att *domain.Attachment) error {
	query := `
		INSERT INTO attachments (
			id, tenant_id, reference_id, reference_type, file_name, file_path,
			file_size, mime_type, uploaded_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING created_at`

	err := r.pool.QueryRow(ctx, query,
		att.ID, att.TenantID, att.ReferenceID, att.ReferenceType,
		att.FileName, att.FilePath, att.FileSize,
		nilIfEmpty(att.MimeType), att.UploadedBy,
	).Scan(&att.CreatedAt)
	if err != nil {
		return fmt.Errorf("attachment_repo: create: %w", err)
	}
	return nil
}

// GetByID retrieves an attachment by its primary key within a tenant scope.
func (r *AttachmentRepo) GetByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*domain.Attachment, error) {
	query := fmt.Sprintf(`SELECT %s FROM attachments WHERE id = $1 AND tenant_id = $2`, attachmentColumns)
	a, err := scanAttachment(r.pool.QueryRow(ctx, query, id, tenantID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("attachment_repo: get_by_id: %w", err)
	}
	return a, nil
}

// ListByReference returns all attachments for a given reference within a tenant scope.
func (r *AttachmentRepo) ListByReference(
	ctx context.Context,
	tenantID uuid.UUID,
	referenceID uuid.UUID,
	referenceType string,
) ([]*domain.Attachment, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM attachments WHERE tenant_id = $1 AND reference_id = $2 AND reference_type = $3 ORDER BY created_at DESC`,
		attachmentColumns,
	)

	rows, err := r.pool.Query(ctx, query, tenantID, referenceID, referenceType)
	if err != nil {
		return nil, fmt.Errorf("attachment_repo: list_by_reference query: %w", err)
	}
	defer rows.Close()

	var items []*domain.Attachment
	for rows.Next() {
		a, scanErr := scanAttachment(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("attachment_repo: list_by_reference scan: %w", scanErr)
		}
		items = append(items, a)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("attachment_repo: list_by_reference rows: %w", err)
	}
	return items, nil
}

// Delete removes an attachment by its primary key within a tenant scope.
func (r *AttachmentRepo) Delete(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error {
	tag, err := r.pool.Exec(ctx,
		`DELETE FROM attachments WHERE id = $1 AND tenant_id = $2`,
		id, tenantID,
	)
	if err != nil {
		return fmt.Errorf("attachment_repo: delete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

// Ensure interface compliance at compile time.
var _ domain.AttachmentRepository = (*AttachmentRepo)(nil)

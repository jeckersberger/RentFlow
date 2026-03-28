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

// documentColumns lists all columns of the documents table for consistent scanning.
const documentColumns = `
	id, tenant_id, template_id, type, title, reference_id, reference_type,
	content, file_path, file_size, mime_type, status, created_by,
	created_at, updated_at`

// DocumentRepo implements domain.DocumentRepository using PostgreSQL.
type DocumentRepo struct {
	pool *pgxpool.Pool
}

// NewDocumentRepo creates a new DocumentRepo.
func NewDocumentRepo(pool *pgxpool.Pool) *DocumentRepo {
	return &DocumentRepo{pool: pool}
}

// scanDocument scans a single documents row into a domain.Document.
func scanDocument(row pgx.Row) (*domain.Document, error) {
	d := &domain.Document{}
	var (
		referenceType *string
		content       *string
		filePath      *string
		mimeType      *string
		status        *string
	)

	err := row.Scan(
		&d.ID, &d.TenantID, &d.TemplateID, &d.Type, &d.Title,
		&d.ReferenceID, &referenceType,
		&content, &filePath, &d.FileSize, &mimeType, &status, &d.CreatedBy,
		&d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if referenceType != nil {
		d.ReferenceType = *referenceType
	}
	if content != nil {
		d.Content = *content
	}
	if filePath != nil {
		d.FilePath = *filePath
	}
	if mimeType != nil {
		d.MimeType = *mimeType
	}
	if status != nil {
		d.Status = *status
	}

	return d, nil
}

// Create inserts a new document.
func (r *DocumentRepo) Create(ctx context.Context, doc *domain.Document) error {
	query := `
		INSERT INTO documents (
			id, tenant_id, template_id, type, title, reference_id, reference_type,
			content, file_path, file_size, mime_type, status, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING created_at, updated_at`

	err := r.pool.QueryRow(ctx, query,
		doc.ID, doc.TenantID, doc.TemplateID, doc.Type, doc.Title,
		doc.ReferenceID, nilIfEmpty(doc.ReferenceType),
		nilIfEmpty(doc.Content), nilIfEmpty(doc.FilePath),
		doc.FileSize, nilIfEmpty(doc.MimeType), nilIfEmpty(doc.Status),
		doc.CreatedBy,
	).Scan(&doc.CreatedAt, &doc.UpdatedAt)
	if err != nil {
		return fmt.Errorf("document_repo: create: %w", err)
	}
	return nil
}

// GetByID retrieves a document by its primary key within a tenant scope.
func (r *DocumentRepo) GetByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*domain.Document, error) {
	query := fmt.Sprintf(`SELECT %s FROM documents WHERE id = $1 AND tenant_id = $2`, documentColumns)
	d, err := scanDocument(r.pool.QueryRow(ctx, query, id, tenantID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("document_repo: get_by_id: %w", err)
	}
	return d, nil
}

// List returns all documents for a tenant.
func (r *DocumentRepo) List(ctx context.Context, tenantID uuid.UUID) ([]*domain.Document, error) {
	query := fmt.Sprintf(`SELECT %s FROM documents WHERE tenant_id = $1 ORDER BY created_at DESC`, documentColumns)

	rows, err := r.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("document_repo: list query: %w", err)
	}
	defer rows.Close()

	var items []*domain.Document
	for rows.Next() {
		d, scanErr := scanDocument(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("document_repo: list scan: %w", scanErr)
		}
		items = append(items, d)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("document_repo: list rows: %w", err)
	}
	return items, nil
}

// nilIfEmpty returns nil if the string is empty, otherwise a pointer to it.
func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// Ensure interface compliance at compile time.
var _ domain.DocumentRepository = (*DocumentRepo)(nil)

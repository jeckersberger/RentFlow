package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jeckersberger/rentflow/pkg/common/database"
	"github.com/jeckersberger/rentflow/services/document-service/internal/domain"
)

type DocumentPostgres struct {
	db *database.PostgresPool
}

func NewDocumentPostgres(db *database.PostgresPool) *DocumentPostgres {
	return &DocumentPostgres{db: db}
}

func (r *DocumentPostgres) Create(ctx context.Context, doc *domain.Document) error {
	query := `
		INSERT INTO documents.documents (
			id, tenant_id, name, type, entity_type, entity_id, file_ref, mime_type,
			size, checksum, created_by, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		)
	`

	_, err := r.db.Exec(ctx, query,
		doc.ID, doc.TenantID, doc.Name, string(doc.Type), doc.EntityType, doc.EntityID,
		doc.FileRef, doc.MimeType, doc.Size, doc.Checksum, doc.CreatedBy, doc.CreatedAt, doc.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create document: %w", err)
	}

	return nil
}

func (r *DocumentPostgres) GetByID(ctx context.Context, tenantID, documentID string) (*domain.Document, error) {
	query := `
		SELECT id, tenant_id, name, type, entity_type, entity_id, file_ref, mime_type,
		       size, checksum, created_by, created_at, updated_at
		FROM documents.documents
		WHERE id = $1 AND tenant_id = $2
	`

	var doc domain.Document
	err := r.db.QueryRow(ctx, query, documentID, tenantID).Scan(
		&doc.ID, &doc.TenantID, &doc.Name, (*string)(&doc.Type), &doc.EntityType, &doc.EntityID,
		&doc.FileRef, &doc.MimeType, &doc.Size, &doc.Checksum, &doc.CreatedBy, &doc.CreatedAt, &doc.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("document not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get document: %w", err)
	}

	return &doc, nil
}

func (r *DocumentPostgres) List(ctx context.Context, tenantID string, limit, offset int) ([]*domain.Document, int64, error) {
	countQuery := `SELECT COUNT(*) FROM documents.documents WHERE tenant_id = $1`
	var total int64
	if err := r.db.QueryRow(ctx, countQuery, tenantID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count documents: %w", err)
	}

	query := `
		SELECT id, tenant_id, name, type, entity_type, entity_id, file_ref, mime_type,
		       size, checksum, created_by, created_at, updated_at
		FROM documents.documents
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list documents: %w", err)
	}
	defer rows.Close()

	var docs []*domain.Document
	for rows.Next() {
		var doc domain.Document
		err := rows.Scan(
			&doc.ID, &doc.TenantID, &doc.Name, (*string)(&doc.Type), &doc.EntityType, &doc.EntityID,
			&doc.FileRef, &doc.MimeType, &doc.Size, &doc.Checksum, &doc.CreatedBy, &doc.CreatedAt, &doc.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan document: %w", err)
		}
		docs = append(docs, &doc)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows error: %w", err)
	}

	return docs, total, nil
}

func (r *DocumentPostgres) ListByEntity(ctx context.Context, tenantID, entityType, entityID string, limit, offset int) ([]*domain.Document, int64, error) {
	countQuery := `SELECT COUNT(*) FROM documents.documents WHERE tenant_id = $1 AND entity_type = $2 AND entity_id = $3`
	var total int64
	if err := r.db.QueryRow(ctx, countQuery, tenantID, entityType, entityID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count documents: %w", err)
	}

	query := `
		SELECT id, tenant_id, name, type, entity_type, entity_id, file_ref, mime_type,
		       size, checksum, created_by, created_at, updated_at
		FROM documents.documents
		WHERE tenant_id = $1 AND entity_type = $2 AND entity_id = $3
		ORDER BY created_at DESC
		LIMIT $4 OFFSET $5
	`

	rows, err := r.db.Query(ctx, query, tenantID, entityType, entityID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list documents: %w", err)
	}
	defer rows.Close()

	var docs []*domain.Document
	for rows.Next() {
		var doc domain.Document
		err := rows.Scan(
			&doc.ID, &doc.TenantID, &doc.Name, (*string)(&doc.Type), &doc.EntityType, &doc.EntityID,
			&doc.FileRef, &doc.MimeType, &doc.Size, &doc.Checksum, &doc.CreatedBy, &doc.CreatedAt, &doc.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan document: %w", err)
		}
		docs = append(docs, &doc)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows error: %w", err)
	}

	return docs, total, nil
}

func (r *DocumentPostgres) Delete(ctx context.Context, tenantID, documentID string) error {
	query := `DELETE FROM documents.documents WHERE id = $1 AND tenant_id = $2`

	result, err := r.db.Exec(ctx, query, documentID, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("document not found")
	}

	return nil
}

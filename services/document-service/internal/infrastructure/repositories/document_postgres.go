package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

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
	metadataJSON := []byte("{}")
	if doc.Metadata != nil {
		data, err := json.Marshal(doc.Metadata)
		if err != nil {
			return err
		}
		metadataJSON = data
	}

	query := `
		INSERT INTO documents
		(id, tenant_id, document_type, reference_id, document_number, title, status,
		 current_version, template_id, metadata, checksum_sha256, previous_checksum,
		 created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`
	_, err := r.db.Exec(ctx, query,
		doc.ID, doc.TenantID, string(doc.DocumentType), doc.ReferenceID,
		doc.DocumentNumber, doc.Title, string(doc.Status), doc.CurrentVersion,
		doc.TemplateID, metadataJSON, doc.ChecksumSHA256, doc.PreviousChecksum,
		doc.CreatedBy, doc.CreatedAt, doc.UpdatedAt,
	)
	return err
}

func (r *DocumentPostgres) GetByID(ctx context.Context, tenantID, id string) (*domain.Document, error) {
	query := `
		SELECT id, tenant_id, document_type, reference_id, document_number, title, status,
		       current_version, template_id, metadata, checksum_sha256, previous_checksum,
		       created_by, created_at, updated_at
		FROM documents
		WHERE id = $1 AND tenant_id = $2
	`
	var doc domain.Document
	var metadataJSON []byte

	err := r.db.QueryRow(ctx, query, id, tenantID).Scan(
		&doc.ID, &doc.TenantID, &doc.DocumentType, &doc.ReferenceID,
		&doc.DocumentNumber, &doc.Title, &doc.Status, &doc.CurrentVersion,
		&doc.TemplateID, &metadataJSON, &doc.ChecksumSHA256, &doc.PreviousChecksum,
		&doc.CreatedBy, &doc.CreatedAt, &doc.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	doc.Metadata = make(map[string]interface{})
	if err := json.Unmarshal(metadataJSON, &doc.Metadata); err != nil {
		return nil, err
	}

	return &doc, nil
}

func (r *DocumentPostgres) ListByTenant(ctx context.Context, tenantID string) ([]*domain.Document, error) {
	query := `
		SELECT id, tenant_id, document_type, reference_id, document_number, title, status,
		       current_version, template_id, metadata, checksum_sha256, previous_checksum,
		       created_by, created_at, updated_at
		FROM documents
		WHERE tenant_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var docs []*domain.Document
	for rows.Next() {
		var doc domain.Document
		var metadataJSON []byte

		if err := rows.Scan(
			&doc.ID, &doc.TenantID, &doc.DocumentType, &doc.ReferenceID,
			&doc.DocumentNumber, &doc.Title, &doc.Status, &doc.CurrentVersion,
			&doc.TemplateID, &metadataJSON, &doc.ChecksumSHA256, &doc.PreviousChecksum,
			&doc.CreatedBy, &doc.CreatedAt, &doc.UpdatedAt,
		); err != nil {
			return nil, err
		}

		doc.Metadata = make(map[string]interface{})
		if err := json.Unmarshal(metadataJSON, &doc.Metadata); err != nil {
			return nil, err
		}

		docs = append(docs, &doc)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return docs, nil
}

func (r *DocumentPostgres) ListByType(ctx context.Context, tenantID string, docType domain.DocumentType) ([]*domain.Document, error) {
	query := `
		SELECT id, tenant_id, document_type, reference_id, document_number, title, status,
		       current_version, template_id, metadata, checksum_sha256, previous_checksum,
		       created_by, created_at, updated_at
		FROM documents
		WHERE tenant_id = $1 AND document_type = $2
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, tenantID, string(docType))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var docs []*domain.Document
	for rows.Next() {
		var doc domain.Document
		var metadataJSON []byte

		if err := rows.Scan(
			&doc.ID, &doc.TenantID, &doc.DocumentType, &doc.ReferenceID,
			&doc.DocumentNumber, &doc.Title, &doc.Status, &doc.CurrentVersion,
			&doc.TemplateID, &metadataJSON, &doc.ChecksumSHA256, &doc.PreviousChecksum,
			&doc.CreatedBy, &doc.CreatedAt, &doc.UpdatedAt,
		); err != nil {
			return nil, err
		}

		doc.Metadata = make(map[string]interface{})
		if err := json.Unmarshal(metadataJSON, &doc.Metadata); err != nil {
			return nil, err
		}

		docs = append(docs, &doc)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return docs, nil
}

func (r *DocumentPostgres) Update(ctx context.Context, doc *domain.Document) error {
	metadataJSON := []byte("{}")
	if doc.Metadata != nil {
		data, err := json.Marshal(doc.Metadata)
		if err != nil {
			return err
		}
		metadataJSON = data
	}

	query := `
		UPDATE documents
		SET status = $1, current_version = $2, checksum_sha256 = $3,
		    previous_checksum = $4, metadata = $5, updated_at = $6
		WHERE id = $7 AND tenant_id = $8
	`
	_, err := r.db.Exec(ctx, query,
		string(doc.Status), doc.CurrentVersion, doc.ChecksumSHA256,
		doc.PreviousChecksum, metadataJSON, doc.UpdatedAt,
		doc.ID, doc.TenantID,
	)
	return err
}

func (r *DocumentPostgres) Archive(ctx context.Context, tenantID, docID string) error {
	query := `
		UPDATE documents
		SET status = $1, updated_at = $2
		WHERE id = $3 AND tenant_id = $4
	`
	_, err := r.db.Exec(ctx, query,
		string(domain.DocumentStatusArchived), time.Now(), docID, tenantID,
	)
	return err
}

func (r *DocumentPostgres) ListAllForChecksumChain(ctx context.Context, tenantID string) ([]*domain.Document, error) {
	query := `
		SELECT id, tenant_id, document_type, reference_id, document_number, title, status,
		       current_version, template_id, metadata, checksum_sha256, previous_checksum,
		       created_by, created_at, updated_at
		FROM documents
		WHERE tenant_id = $1
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var docs []*domain.Document
	for rows.Next() {
		var doc domain.Document
		var metadataJSON []byte

		if err := rows.Scan(
			&doc.ID, &doc.TenantID, &doc.DocumentType, &doc.ReferenceID,
			&doc.DocumentNumber, &doc.Title, &doc.Status, &doc.CurrentVersion,
			&doc.TemplateID, &metadataJSON, &doc.ChecksumSHA256, &doc.PreviousChecksum,
			&doc.CreatedBy, &doc.CreatedAt, &doc.UpdatedAt,
		); err != nil {
			return nil, err
		}

		doc.Metadata = make(map[string]interface{})
		if err := json.Unmarshal(metadataJSON, &doc.Metadata); err != nil {
			return nil, err
		}

		docs = append(docs, &doc)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return docs, nil
}

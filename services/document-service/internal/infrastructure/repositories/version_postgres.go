package repositories

import (
	"context"
	"database/sql"

	"github.com/jeckersberger/rentflow/pkg/common/database"
	"github.com/jeckersberger/rentflow/services/document-service/internal/domain"
)

type DocumentVersionPostgres struct {
	db *database.PostgresPool
}

func NewDocumentVersionPostgres(db *database.PostgresPool) *DocumentVersionPostgres {
	return &DocumentVersionPostgres{db: db}
}

func (r *DocumentVersionPostgres) Create(ctx context.Context, version *domain.DocumentVersion) error {
	query := `
		INSERT INTO document_versions
		(id, document_id, version_number, file_path, file_size, mime_type,
		 checksum_sha256, changes_description, created_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.db.Exec(ctx, query,
		version.ID, version.DocumentID, version.VersionNumber, version.FilePath,
		version.FileSize, version.MimeType, version.ChecksumSHA256,
		version.ChangesDescription, version.CreatedBy, version.CreatedAt,
	)
	return err
}

func (r *DocumentVersionPostgres) GetByID(ctx context.Context, versionID string) (*domain.DocumentVersion, error) {
	query := `
		SELECT id, document_id, version_number, file_path, file_size, mime_type,
		       checksum_sha256, changes_description, created_by, created_at
		FROM document_versions
		WHERE id = $1
	`
	var version domain.DocumentVersion
	err := r.db.QueryRow(ctx, query, versionID).Scan(
		&version.ID, &version.DocumentID, &version.VersionNumber, &version.FilePath,
		&version.FileSize, &version.MimeType, &version.ChecksumSHA256,
		&version.ChangesDescription, &version.CreatedBy, &version.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &version, nil
}

func (r *DocumentVersionPostgres) ListByDocument(ctx context.Context, docID string) ([]*domain.DocumentVersion, error) {
	query := `
		SELECT id, document_id, version_number, file_path, file_size, mime_type,
		       checksum_sha256, changes_description, created_by, created_at
		FROM document_versions
		WHERE document_id = $1
		ORDER BY version_number ASC
	`
	rows, err := r.db.Query(ctx, query, docID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []*domain.DocumentVersion
	for rows.Next() {
		var version domain.DocumentVersion
		if err := rows.Scan(
			&version.ID, &version.DocumentID, &version.VersionNumber, &version.FilePath,
			&version.FileSize, &version.MimeType, &version.ChecksumSHA256,
			&version.ChangesDescription, &version.CreatedBy, &version.CreatedAt,
		); err != nil {
			return nil, err
		}
		versions = append(versions, &version)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return versions, nil
}

func (r *DocumentVersionPostgres) GetByVersion(ctx context.Context, docID string, versionNumber int) (*domain.DocumentVersion, error) {
	query := `
		SELECT id, document_id, version_number, file_path, file_size, mime_type,
		       checksum_sha256, changes_description, created_by, created_at
		FROM document_versions
		WHERE document_id = $1 AND version_number = $2
	`
	var version domain.DocumentVersion
	err := r.db.QueryRow(ctx, query, docID, versionNumber).Scan(
		&version.ID, &version.DocumentID, &version.VersionNumber, &version.FilePath,
		&version.FileSize, &version.MimeType, &version.ChecksumSHA256,
		&version.ChangesDescription, &version.CreatedBy, &version.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &version, nil
}

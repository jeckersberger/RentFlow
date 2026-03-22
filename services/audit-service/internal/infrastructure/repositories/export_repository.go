package repositories

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/audit-service/internal/domain"
)

type PostgresExportRepository struct {
	db  *sql.DB
	log logger.Logger
}

func NewPostgresExportRepository(db *sql.DB, log logger.Logger) *PostgresExportRepository {
	return &PostgresExportRepository{db: db, log: log}
}

func (r *PostgresExportRepository) Create(ctx context.Context, export *domain.AuditExport) (*domain.AuditExport, error) {
	query := `INSERT INTO audit_exports (id, tenant_id, export_type, date_from, date_to, status, requested_by, created_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
	RETURNING id, created_at`

	err := r.db.QueryRowContext(ctx, query,
		export.ID, export.TenantID, export.ExportType, export.DateFrom, export.DateTo, export.Status, export.RequestedBy).
		Scan(&export.ID, &export.CreatedAt)

	if err != nil {
		r.log.Error("Failed to create export", err)
		return nil, domain.ErrDatabaseError
	}

	return export, nil
}

func (r *PostgresExportRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.AuditExport, error) {
	query := `SELECT id, tenant_id, export_type, date_from, date_to, status, file_path, file_size_bytes, checksum, requested_by, completed_at, created_at
	FROM audit_exports WHERE id = $1`

	export := &domain.AuditExport{}

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&export.ID, &export.TenantID, &export.ExportType, &export.DateFrom, &export.DateTo, &export.Status,
		&export.FilePath, &export.FileSizeBytes, &export.Checksum, &export.RequestedBy, &export.CompletedAt, &export.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, domain.ErrExportNotFound
	}
	if err != nil {
		return nil, domain.ErrDatabaseError
	}

	return export, nil
}

func (r *PostgresExportRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*domain.AuditExport, error) {
	query := `SELECT id, tenant_id, export_type, date_from, date_to, status, file_path, file_size_bytes, checksum, requested_by, completed_at, created_at
	FROM audit_exports WHERE tenant_id = $1 ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, domain.ErrDatabaseError
	}
	defer rows.Close()

	var exports []*domain.AuditExport

	for rows.Next() {
		export := &domain.AuditExport{}

		err := rows.Scan(&export.ID, &export.TenantID, &export.ExportType, &export.DateFrom, &export.DateTo, &export.Status,
			&export.FilePath, &export.FileSizeBytes, &export.Checksum, &export.RequestedBy, &export.CompletedAt, &export.CreatedAt)
		if err != nil {
			continue
		}

		exports = append(exports, export)
	}

	return exports, nil
}

func (r *PostgresExportRepository) Update(ctx context.Context, export *domain.AuditExport) (*domain.AuditExport, error) {
	query := `UPDATE audit_exports SET status = $1, file_size_bytes = $2, checksum = $3, completed_at = $4
	WHERE id = $5
	RETURNING id`

	err := r.db.QueryRowContext(ctx, query, export.Status, export.FileSizeBytes, export.Checksum, export.CompletedAt, export.ID).
		Scan(&export.ID)

	if err != nil {
		r.log.Error("Failed to update export", err)
		return nil, domain.ErrDatabaseError
	}

	return export, nil
}

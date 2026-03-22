package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jeckersberger/rentflow/pkg/common/database"
	"github.com/jeckersberger/rentflow/services/scanner-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/scanner-service/internal/ports"
)

type ScanSessionPostgres struct {
	db *database.PostgresPool
}

func NewScanSessionPostgres(db *database.PostgresPool) ports.ScanSessionRepository {
	return &ScanSessionPostgres{db: db}
}

func (r *ScanSessionPostgres) Create(ctx context.Context, session *domain.ScanSession) error {
	query := `
		INSERT INTO scan_sessions
		(id, tenant_id, user_id, context, project_id, started_at, ended_at,
		 device_type, device_id, total_scans, successful_scans, failed_scans, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`

	_, err := r.db.Exec(ctx, query,
		session.ID,
		session.TenantID,
		session.UserID,
		string(session.Context),
		session.ProjectID,
		session.StartedAt,
		session.EndedAt,
		string(session.DeviceType),
		session.DeviceID,
		session.TotalScans,
		session.SuccessfulScans,
		session.FailedScans,
		session.CreatedAt,
		session.UpdatedAt,
	)

	return err
}

func (r *ScanSessionPostgres) GetByID(ctx context.Context, tenantID, id string) (*domain.ScanSession, error) {
	query := `
		SELECT id, tenant_id, user_id, context, project_id, started_at, ended_at,
		       device_type, device_id, total_scans, successful_scans, failed_scans, created_at, updated_at
		FROM scan_sessions
		WHERE tenant_id = $1 AND id = $2
	`

	row := r.db.QueryRow(ctx, query, tenantID, id)
	return sessionRowToSession(row)
}

func (r *ScanSessionPostgres) Update(ctx context.Context, session *domain.ScanSession) error {
	query := `
		UPDATE scan_sessions
		SET user_id = $1, context = $2, project_id = $3, started_at = $4, ended_at = $5,
		    device_type = $6, device_id = $7, total_scans = $8, successful_scans = $9,
		    failed_scans = $10, updated_at = $11
		WHERE id = $12 AND tenant_id = $13
	`

	_, err := r.db.Exec(ctx, query,
		session.UserID,
		string(session.Context),
		session.ProjectID,
		session.StartedAt,
		session.EndedAt,
		string(session.DeviceType),
		session.DeviceID,
		session.TotalScans,
		session.SuccessfulScans,
		session.FailedScans,
		session.UpdatedAt,
		session.ID,
		session.TenantID,
	)

	return err
}

func (r *ScanSessionPostgres) List(ctx context.Context, tenantID string, limit, offset int) ([]*domain.ScanSession, int, error) {
	query := `
		SELECT id, tenant_id, user_id, context, project_id, started_at, ended_at,
		       device_type, device_id, total_scans, successful_scans, failed_scans, created_at, updated_at
		FROM scan_sessions
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	sessions := make([]*domain.ScanSession, 0)
	for rows.Next() {
		session, err := sessionRowsToSession(rows)
		if err != nil {
			return nil, 0, err
		}
		sessions = append(sessions, session)
	}

	// Get total count
	countQuery := "SELECT COUNT(*) FROM scan_sessions WHERE tenant_id = $1"
	var total int
	if err := r.db.QueryRow(ctx, countQuery, tenantID).Scan(&total); err != nil {
		return nil, 0, err
	}

	return sessions, total, rows.Err()
}

func (r *ScanSessionPostgres) GetByUserAndContext(ctx context.Context, tenantID, userID string, context domain.ScanContext) (*domain.ScanSession, error) {
	query := `
		SELECT id, tenant_id, user_id, context, project_id, started_at, ended_at,
		       device_type, device_id, total_scans, successful_scans, failed_scans, created_at, updated_at
		FROM scan_sessions
		WHERE tenant_id = $1 AND user_id = $2 AND context = $3 AND ended_at IS NULL
		ORDER BY created_at DESC
		LIMIT 1
	`

	row := r.db.QueryRow(ctx, query, tenantID, userID, string(context))
	return sessionRowToSession(row)
}

func sessionRowToSession(row *sql.Row) (*domain.ScanSession, error) {
	session := &domain.ScanSession{}
	err := row.Scan(
		&session.ID,
		&session.TenantID,
		&session.UserID,
		(*string)(&session.Context),
		&session.ProjectID,
		&session.StartedAt,
		&session.EndedAt,
		(*string)(&session.DeviceType),
		&session.DeviceID,
		&session.TotalScans,
		&session.SuccessfulScans,
		&session.FailedScans,
		&session.CreatedAt,
		&session.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("scan session not found")
		}
		return nil, err
	}
	return session, nil
}

func sessionRowsToSession(rows *sql.Rows) (*domain.ScanSession, error) {
	session := &domain.ScanSession{}
	err := rows.Scan(
		&session.ID,
		&session.TenantID,
		&session.UserID,
		(*string)(&session.Context),
		&session.ProjectID,
		&session.StartedAt,
		&session.EndedAt,
		(*string)(&session.DeviceType),
		&session.DeviceID,
		&session.TotalScans,
		&session.SuccessfulScans,
		&session.FailedScans,
		&session.CreatedAt,
		&session.UpdatedAt,
	)
	return session, err
}

package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jeckersberger/EquipFlow/services/scanner/internal/domain"

	apperrors "github.com/jeckersberger/EquipFlow/pkg/common/errors"
)

// sessionColumns lists all columns of the scan_sessions table for consistent scanning.
const sessionColumns = `
	id, tenant_id, session_type, project_id, started_by,
	started_at, ended_at, items_count, signature_data, status, created_at`

// SessionRepo implements domain.ScanSessionRepository using PostgreSQL.
type SessionRepo struct {
	pool *pgxpool.Pool
}

// NewSessionRepo creates a new SessionRepo.
func NewSessionRepo(pool *pgxpool.Pool) *SessionRepo {
	return &SessionRepo{pool: pool}
}

// scanSession scans a single scan_sessions row into a domain.ScanSession.
func scanSession(row pgx.Row) (*domain.ScanSession, error) {
	s := &domain.ScanSession{}
	var (
		signatureData *string
	)

	err := row.Scan(
		&s.ID, &s.TenantID, &s.SessionType, &s.ProjectID, &s.StartedBy,
		&s.StartedAt, &s.EndedAt, &s.ItemsCount, &signatureData, &s.Status, &s.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	if signatureData != nil {
		s.SignatureData = *signatureData
	}

	return s, nil
}

// Create inserts a new scan session record.
func (r *SessionRepo) Create(ctx context.Context, session *domain.ScanSession) error {
	query := `
		INSERT INTO scan_sessions (
			id, tenant_id, session_type, project_id, started_by,
			started_at, items_count, status
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at`

	err := r.pool.QueryRow(ctx, query,
		session.ID, session.TenantID, session.SessionType, session.ProjectID,
		session.StartedBy, session.StartedAt, session.ItemsCount, session.Status,
	).Scan(&session.ID, &session.CreatedAt)
	if err != nil {
		return fmt.Errorf("session_repo: create: %w", err)
	}
	return nil
}

// GetByID retrieves a scan session by its primary key within a tenant scope.
func (r *SessionRepo) GetByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*domain.ScanSession, error) {
	query := fmt.Sprintf(`SELECT %s FROM scan_sessions WHERE id = $1 AND tenant_id = $2`, sessionColumns)
	s, err := scanSession(r.pool.QueryRow(ctx, query, id, tenantID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("session_repo: get_by_id: %w", err)
	}
	return s, nil
}

// End marks a scan session as completed by setting ended_at and status.
func (r *SessionRepo) End(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE scan_sessions SET ended_at = NOW(), status = $3 WHERE id = $1 AND tenant_id = $2 AND status = 'active'`,
		id, tenantID, domain.SessionStatusCompleted,
	)
	if err != nil {
		return fmt.Errorf("session_repo: end: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

// UpdateSignature saves a base64 signature for a session.
func (r *SessionRepo) UpdateSignature(ctx context.Context, id uuid.UUID, tenantID uuid.UUID, signatureData string) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE scan_sessions SET signature_data = $3 WHERE id = $1 AND tenant_id = $2`,
		id, tenantID, signatureData,
	)
	if err != nil {
		return fmt.Errorf("session_repo: update_signature: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

// IncrementItems increments the items_count for a session.
func (r *SessionRepo) IncrementItems(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE scan_sessions SET items_count = items_count + 1 WHERE id = $1 AND tenant_id = $2`,
		id, tenantID,
	)
	if err != nil {
		return fmt.Errorf("session_repo: increment_items: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

// GetSessionEvents returns all scan events linked to a session.
func (r *SessionRepo) GetSessionEvents(ctx context.Context, sessionID uuid.UUID, tenantID uuid.UUID) ([]*domain.ScanEvent, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM scan_events WHERE session_id = $1 AND tenant_id = $2 ORDER BY timestamp ASC`,
		scanEventColumns,
	)

	rows, err := r.pool.Query(ctx, query, sessionID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("session_repo: get_session_events query: %w", err)
	}
	defer rows.Close()

	var items []*domain.ScanEvent
	for rows.Next() {
		e, scanErr := scanScanEvent(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("session_repo: get_session_events scan: %w", scanErr)
		}
		items = append(items, e)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("session_repo: get_session_events rows: %w", err)
	}
	return items, nil
}

// Ensure interface compliance at compile time.
var _ domain.ScanSessionRepository = (*SessionRepo)(nil)

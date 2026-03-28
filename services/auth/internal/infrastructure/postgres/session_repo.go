package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jeckersberger/EquipFlow/services/auth/internal/domain"

	apperrors "github.com/jeckersberger/EquipFlow/pkg/common/errors"
)

// SessionRepo implements domain.SessionRepository using PostgreSQL.
type SessionRepo struct {
	pool *pgxpool.Pool
}

// NewSessionRepo creates a new SessionRepo.
func NewSessionRepo(pool *pgxpool.Pool) *SessionRepo {
	return &SessionRepo{pool: pool}
}

// Create inserts a new session and scans back the generated fields.
func (r *SessionRepo) Create(ctx context.Context, session *domain.Session) error {
	query := `
		INSERT INTO sessions (
			id, user_id, token_hash, ip_address, user_agent,
			last_activity, expires_at, is_active
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at`

	err := r.pool.QueryRow(ctx, query,
		session.ID, session.UserID, session.TokenHash, session.IPAddress, session.UserAgent,
		session.LastActivity, session.ExpiresAt, session.IsActive,
	).Scan(&session.ID, &session.CreatedAt)
	if err != nil {
		return fmt.Errorf("session_repo: create: %w", err)
	}
	return nil
}

// GetByTokenHash retrieves an active, non-expired session by its token hash.
func (r *SessionRepo) GetByTokenHash(ctx context.Context, tokenHash string) (*domain.Session, error) {
	query := `
		SELECT id, user_id, token_hash, ip_address, user_agent,
			last_activity, expires_at, is_active, created_at
		FROM sessions
		WHERE token_hash = $1 AND is_active = true AND expires_at > NOW()`

	s := &domain.Session{}
	err := r.pool.QueryRow(ctx, query, tokenHash).Scan(
		&s.ID, &s.UserID, &s.TokenHash, &s.IPAddress, &s.UserAgent,
		&s.LastActivity, &s.ExpiresAt, &s.IsActive, &s.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("session_repo: get_by_token_hash: %w", err)
	}
	return s, nil
}

// Deactivate marks a single session as inactive.
func (r *SessionRepo) Deactivate(ctx context.Context, sessionID uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE sessions SET is_active = false WHERE id = $1`,
		sessionID,
	)
	if err != nil {
		return fmt.Errorf("session_repo: deactivate: %w", err)
	}
	return nil
}

// DeactivateAllForUser marks all sessions for a user as inactive.
func (r *SessionRepo) DeactivateAllForUser(ctx context.Context, userID uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE sessions SET is_active = false WHERE user_id = $1`,
		userID,
	)
	if err != nil {
		return fmt.Errorf("session_repo: deactivate_all_for_user: %w", err)
	}
	return nil
}

// UpdateLastActivity refreshes the last_activity timestamp for a session.
func (r *SessionRepo) UpdateLastActivity(ctx context.Context, sessionID uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE sessions SET last_activity = NOW() WHERE id = $1`,
		sessionID,
	)
	if err != nil {
		return fmt.Errorf("session_repo: update_last_activity: %w", err)
	}
	return nil
}

// UpdateTokenHash replaces the token hash on an existing session (for token rotation).
func (r *SessionRepo) UpdateTokenHash(ctx context.Context, sessionID uuid.UUID, newTokenHash string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE sessions SET token_hash = $2, last_activity = NOW() WHERE id = $1`,
		sessionID, newTokenHash,
	)
	if err != nil {
		return fmt.Errorf("session_repo: update_token_hash: %w", err)
	}
	return nil
}

// CleanupExpired removes expired or deactivated sessions and returns
// the number of rows deleted.
func (r *SessionRepo) CleanupExpired(ctx context.Context) (int64, error) {
	tag, err := r.pool.Exec(ctx,
		`DELETE FROM sessions WHERE expires_at < NOW() OR is_active = false`,
	)
	if err != nil {
		return 0, fmt.Errorf("session_repo: cleanup_expired: %w", err)
	}
	return tag.RowsAffected(), nil
}

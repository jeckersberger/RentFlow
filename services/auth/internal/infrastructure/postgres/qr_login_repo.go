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

// QRLoginRepo implements domain.QRLoginRepository using PostgreSQL.
type QRLoginRepo struct {
	pool *pgxpool.Pool
}

// NewQRLoginRepo creates a new QRLoginRepo.
func NewQRLoginRepo(pool *pgxpool.Pool) *QRLoginRepo {
	return &QRLoginRepo{pool: pool}
}

// Create inserts a new QR login token and scans back the generated fields.
func (r *QRLoginRepo) Create(ctx context.Context, token *domain.QRLoginToken) error {
	query := `
		INSERT INTO qr_login_tokens (
			id, tenant_id, user_id, token, expires_at, used
		) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at`

	err := r.pool.QueryRow(ctx, query,
		token.ID, token.TenantID, token.UserID, token.Token,
		token.ExpiresAt, token.Used,
	).Scan(&token.ID, &token.CreatedAt)
	if err != nil {
		return fmt.Errorf("qr_login_repo: create: %w", err)
	}
	return nil
}

// GetByToken retrieves a non-used, non-expired QR login token by its value.
func (r *QRLoginRepo) GetByToken(ctx context.Context, token string) (*domain.QRLoginToken, error) {
	query := `
		SELECT id, tenant_id, user_id, token, expires_at, used, created_at
		FROM qr_login_tokens
		WHERE token = $1 AND used = false AND expires_at > NOW()`

	t := &domain.QRLoginToken{}
	err := r.pool.QueryRow(ctx, query, token).Scan(
		&t.ID, &t.TenantID, &t.UserID, &t.Token,
		&t.ExpiresAt, &t.Used, &t.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("qr_login_repo: get_by_token: %w", err)
	}
	return t, nil
}

// MarkUsed sets the used flag to true on a QR login token.
func (r *QRLoginRepo) MarkUsed(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE qr_login_tokens SET used = true WHERE id = $1`,
		id,
	)
	if err != nil {
		return fmt.Errorf("qr_login_repo: mark_used: %w", err)
	}
	return nil
}

// ClaimToken atomically marks a QR token as used and returns it.
// Returns ErrNotFound if token doesn't exist, is expired, or already used.
func (r *QRLoginRepo) ClaimToken(ctx context.Context, token string) (*domain.QRLoginToken, error) {
	query := `
		UPDATE qr_login_tokens
		SET used = true
		WHERE token = $1 AND used = false AND expires_at > NOW()
		RETURNING id, tenant_id, user_id, token, expires_at, used, created_at`

	t := &domain.QRLoginToken{}
	err := r.pool.QueryRow(ctx, query, token).Scan(
		&t.ID, &t.TenantID, &t.UserID, &t.Token,
		&t.ExpiresAt, &t.Used, &t.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("qr_login_repo: claim_token: %w", err)
	}
	return t, nil
}

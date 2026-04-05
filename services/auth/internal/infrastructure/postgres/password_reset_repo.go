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

// PasswordResetRepo implements domain.PasswordResetRepository using PostgreSQL.
type PasswordResetRepo struct {
	pool *pgxpool.Pool
}

// NewPasswordResetRepo creates a new PasswordResetRepo.
func NewPasswordResetRepo(pool *pgxpool.Pool) *PasswordResetRepo {
	return &PasswordResetRepo{pool: pool}
}

// Create inserts a new password reset token and scans back the generated fields.
func (r *PasswordResetRepo) Create(ctx context.Context, token *domain.PasswordResetToken) error {
	query := `
		INSERT INTO password_reset_tokens (
			id, tenant_id, user_id, token_hash, expires_at, used
		) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at`

	err := r.pool.QueryRow(ctx, query,
		token.ID, token.TenantID, token.UserID, token.TokenHash,
		token.ExpiresAt, token.Used,
	).Scan(&token.ID, &token.CreatedAt)
	if err != nil {
		return fmt.Errorf("password_reset_repo: create: %w", err)
	}
	return nil
}

// GetByTokenHash retrieves a non-used, non-expired password reset token by its hash.
func (r *PasswordResetRepo) GetByTokenHash(ctx context.Context, tokenHash string) (*domain.PasswordResetToken, error) {
	query := `
		SELECT id, tenant_id, user_id, token_hash, expires_at, used, created_at
		FROM password_reset_tokens
		WHERE token_hash = $1 AND used = false AND expires_at > NOW()`

	t := &domain.PasswordResetToken{}
	err := r.pool.QueryRow(ctx, query, tokenHash).Scan(
		&t.ID, &t.TenantID, &t.UserID, &t.TokenHash,
		&t.ExpiresAt, &t.Used, &t.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("password_reset_repo: get_by_token_hash: %w", err)
	}
	return t, nil
}

// MarkUsed sets the used flag to true on a password reset token.
func (r *PasswordResetRepo) MarkUsed(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE password_reset_tokens SET used = true WHERE id = $1`,
		id,
	)
	if err != nil {
		return fmt.Errorf("password_reset_repo: mark_used: %w", err)
	}
	return nil
}

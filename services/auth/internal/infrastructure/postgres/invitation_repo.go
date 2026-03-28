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

// InvitationRepo implements domain.InvitationRepository using PostgreSQL.
type InvitationRepo struct {
	pool *pgxpool.Pool
}

// NewInvitationRepo creates a new InvitationRepo.
func NewInvitationRepo(pool *pgxpool.Pool) *InvitationRepo {
	return &InvitationRepo{pool: pool}
}

// Create inserts a new invitation and scans back the generated fields.
func (r *InvitationRepo) Create(ctx context.Context, inv *domain.Invitation) error {
	query := `
		INSERT INTO invitations (
			id, tenant_id, email, role, token,
			invited_by, expires_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at`

	err := r.pool.QueryRow(ctx, query,
		inv.ID, inv.TenantID, inv.Email, inv.Role, inv.Token,
		inv.InvitedBy, inv.ExpiresAt,
	).Scan(&inv.ID, &inv.CreatedAt)
	if err != nil {
		return fmt.Errorf("invitation_repo: create: %w", err)
	}
	return nil
}

// GetByToken retrieves a pending (not accepted, not expired) invitation by its token.
func (r *InvitationRepo) GetByToken(ctx context.Context, token string) (*domain.Invitation, error) {
	query := `
		SELECT id, tenant_id, email, role, token,
			invited_by, accepted_at, expires_at, created_at
		FROM invitations
		WHERE token = $1 AND accepted_at IS NULL AND expires_at > NOW()`

	inv := &domain.Invitation{}
	err := r.pool.QueryRow(ctx, query, token).Scan(
		&inv.ID, &inv.TenantID, &inv.Email, &inv.Role, &inv.Token,
		&inv.InvitedBy, &inv.AcceptedAt, &inv.ExpiresAt, &inv.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("invitation_repo: get_by_token: %w", err)
	}
	return inv, nil
}

// MarkAccepted sets the accepted_at timestamp on an invitation.
func (r *InvitationRepo) MarkAccepted(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE invitations SET accepted_at = NOW() WHERE id = $1`,
		id,
	)
	if err != nil {
		return fmt.Errorf("invitation_repo: mark_accepted: %w", err)
	}
	return nil
}

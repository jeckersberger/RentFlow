package repositories

import (
	"context"
	"database/sql"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/auth-service/internal/ports"
)

type PostgresInvitationRepository struct {
	db     *sql.DB
	logger logger.Logger
}

func NewPostgresInvitationRepository(db *sql.DB, log logger.Logger) *PostgresInvitationRepository {
	return &PostgresInvitationRepository{db: db, logger: log}
}

func (r *PostgresInvitationRepository) Save(ctx context.Context, inv *ports.Invitation) error {
	query := `
		INSERT INTO auth.invitations (id, tenant_id, email, role, token, status, invited_by, expires_at, claimed_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (id) DO UPDATE SET
			status = EXCLUDED.status,
			claimed_at = EXCLUDED.claimed_at
	`
	_, err := r.db.ExecContext(ctx, query,
		inv.ID, inv.TenantID, inv.Email, inv.Role, inv.Token, inv.Status,
		inv.InvitedBy, inv.ExpiresAt, inv.ClaimedAt, inv.CreatedAt,
	)
	return err
}

func (r *PostgresInvitationRepository) FindByToken(ctx context.Context, token string) (*ports.Invitation, error) {
	query := `
		SELECT id, tenant_id, email, role, token, status, invited_by, expires_at, claimed_at, created_at
		FROM auth.invitations WHERE token = $1
	`
	inv := &ports.Invitation{}
	err := r.db.QueryRowContext(ctx, query, token).Scan(
		&inv.ID, &inv.TenantID, &inv.Email, &inv.Role, &inv.Token, &inv.Status,
		&inv.InvitedBy, &inv.ExpiresAt, &inv.ClaimedAt, &inv.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return inv, err
}

func (r *PostgresInvitationRepository) FindByEmail(ctx context.Context, tenantID, email string) (*ports.Invitation, error) {
	query := `
		SELECT id, tenant_id, email, role, token, status, invited_by, expires_at, claimed_at, created_at
		FROM auth.invitations WHERE tenant_id = $1 AND email = $2 AND status = 'pending'
		ORDER BY created_at DESC LIMIT 1
	`
	inv := &ports.Invitation{}
	err := r.db.QueryRowContext(ctx, query, tenantID, email).Scan(
		&inv.ID, &inv.TenantID, &inv.Email, &inv.Role, &inv.Token, &inv.Status,
		&inv.InvitedBy, &inv.ExpiresAt, &inv.ClaimedAt, &inv.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return inv, err
}

func (r *PostgresInvitationRepository) ListByTenant(ctx context.Context, tenantID string) ([]*ports.Invitation, error) {
	query := `
		SELECT id, tenant_id, email, role, token, status, invited_by, expires_at, claimed_at, created_at
		FROM auth.invitations WHERE tenant_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invitations []*ports.Invitation
	for rows.Next() {
		inv := &ports.Invitation{}
		if err := rows.Scan(
			&inv.ID, &inv.TenantID, &inv.Email, &inv.Role, &inv.Token, &inv.Status,
			&inv.InvitedBy, &inv.ExpiresAt, &inv.ClaimedAt, &inv.CreatedAt,
		); err != nil {
			return nil, err
		}
		invitations = append(invitations, inv)
	}
	return invitations, nil
}

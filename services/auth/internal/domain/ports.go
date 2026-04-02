package domain

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// TenantRepository defines persistence operations for tenants.
type TenantRepository interface {
	Create(ctx context.Context, tenant *Tenant) error
	GetByID(ctx context.Context, id uuid.UUID) (*Tenant, error)
	GetBySlug(ctx context.Context, slug string) (*Tenant, error)
	Update(ctx context.Context, tenant *Tenant) error
	SlugExists(ctx context.Context, slug string) (bool, error)
}

// UserRepository defines persistence operations for users.
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByIDAndTenant(ctx context.Context, id, tenantID uuid.UUID) (*User, error)
	GetByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*User, error)
	List(ctx context.Context, tenantID uuid.UUID, page, perPage int) ([]*User, int64, error)
	Update(ctx context.Context, user *User) error
	UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error
	IncrementFailedLogins(ctx context.Context, userID uuid.UUID) (int, error)
	ResetFailedLogins(ctx context.Context, userID uuid.UUID) error
	LockUntil(ctx context.Context, userID uuid.UUID, lockedUntil time.Time) error
	UpdateLastLogin(ctx context.Context, userID uuid.UUID) error
	Deactivate(ctx context.Context, userID uuid.UUID) error
	CountByTenant(ctx context.Context, tenantID uuid.UUID) (int64, error)
}

// SessionRepository defines persistence operations for sessions.
type SessionRepository interface {
	Create(ctx context.Context, session *Session) error
	GetByTokenHash(ctx context.Context, tokenHash string) (*Session, error)
	Deactivate(ctx context.Context, sessionID uuid.UUID) error
	DeactivateAllForUser(ctx context.Context, userID uuid.UUID) error
	UpdateLastActivity(ctx context.Context, sessionID uuid.UUID) error
	UpdateTokenHash(ctx context.Context, sessionID uuid.UUID, newTokenHash string) error
	CleanupExpired(ctx context.Context) (int64, error)
}

// TenantConfigRepository defines persistence operations for tenant config.
type TenantConfigRepository interface {
	Get(ctx context.Context, tenantID uuid.UUID, key string) (*TenantConfig, error)
	GetAll(ctx context.Context, tenantID uuid.UUID) ([]*TenantConfig, error)
	Set(ctx context.Context, tenantID uuid.UUID, key string, value json.RawMessage) error
	Delete(ctx context.Context, tenantID uuid.UUID, key string) error
}

// SetupRepository defines persistence operations for system setup state.
type SetupRepository interface {
	GetState(ctx context.Context) (*SetupState, error)
	CreateState(ctx context.Context, state *SetupState) error
	MarkComplete(ctx context.Context, state *SetupState) error
}

// InvitationRepository defines persistence operations for invitations.
type InvitationRepository interface {
	Create(ctx context.Context, inv *Invitation) error
	GetByToken(ctx context.Context, token string) (*Invitation, error)
	MarkAccepted(ctx context.Context, id uuid.UUID) error
}

// QRLoginRepository defines persistence operations for QR login tokens.
type QRLoginRepository interface {
	Create(ctx context.Context, token *QRLoginToken) error
	GetByToken(ctx context.Context, token string) (*QRLoginToken, error)
	MarkUsed(ctx context.Context, id uuid.UUID) error
}

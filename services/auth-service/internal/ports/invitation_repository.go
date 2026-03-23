package ports

import (
	"context"
	"time"
)

// Invitation represents a pending user invitation
type Invitation struct {
	ID        string
	TenantID  string
	Email     string
	Role      string
	Token     string
	Status    string // pending, accepted, expired
	InvitedBy string
	ExpiresAt time.Time
	ClaimedAt *time.Time
	CreatedAt time.Time
}

// InvitationRepository defines the interface for invitation data access
type InvitationRepository interface {
	Save(ctx context.Context, inv *Invitation) error
	FindByToken(ctx context.Context, token string) (*Invitation, error)
	FindByEmail(ctx context.Context, tenantID, email string) (*Invitation, error)
	ListByTenant(ctx context.Context, tenantID string) ([]*Invitation, error)
}

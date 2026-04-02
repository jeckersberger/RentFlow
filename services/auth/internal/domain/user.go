package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

// Role constants define the available user roles.
const (
	RoleAdmin   = "admin"
	RoleManager = "manager"
	RoleUser    = "user"
	RoleViewer  = "viewer"
)

// User represents an authenticated user within a tenant.
type User struct {
	ID           uuid.UUID  `json:"id"`
	TenantID     uuid.UUID  `json:"tenant_id"`
	Email        string     `json:"email"`
	Username     string     `json:"username,omitempty"`
	PasswordHash string     `json:"-"`
	FirstName    string     `json:"first_name,omitempty"`
	LastName     string     `json:"last_name,omitempty"`
	Role         string     `json:"role"`
	Phone        string     `json:"phone,omitempty"`
	AvatarURL    string     `json:"avatar_url,omitempty"`
	IsActive     bool       `json:"is_active"`
	FailedLogins int        `json:"failed_logins"`
	LockedUntil  *time.Time `json:"locked_until,omitempty"`
	LastLogin    *time.Time `json:"last_login,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// IsLocked returns true if the user account is currently locked.
func (u *User) IsLocked() bool {
	if u.LockedUntil == nil {
		return false
	}
	return u.LockedUntil.After(time.Now())
}

// FullName returns the user's full name (first + last).
func (u *User) FullName() string {
	parts := []string{}
	if u.FirstName != "" {
		parts = append(parts, u.FirstName)
	}
	if u.LastName != "" {
		parts = append(parts, u.LastName)
	}
	return strings.Join(parts, " ")
}

// Session represents an active user session.
type Session struct {
	ID           uuid.UUID `json:"id"`
	UserID       uuid.UUID `json:"user_id"`
	TenantID     uuid.UUID `json:"tenant_id"`
	TokenHash    string    `json:"-"`
	IPAddress    string    `json:"ip_address,omitempty"`
	UserAgent    string    `json:"user_agent,omitempty"`
	LastActivity time.Time `json:"last_activity"`
	ExpiresAt    time.Time `json:"expires_at"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
}

// Invitation represents a pending invitation to join a tenant.
type Invitation struct {
	ID        uuid.UUID  `json:"id"`
	TenantID  uuid.UUID  `json:"tenant_id"`
	Email     string     `json:"email"`
	Role      string     `json:"role"`
	Token     string     `json:"-"`
	InvitedBy uuid.UUID  `json:"invited_by"`
	AcceptedAt *time.Time `json:"accepted_at,omitempty"`
	ExpiresAt time.Time  `json:"expires_at"`
	CreatedAt time.Time  `json:"created_at"`
}

// PasswordReset represents a password reset request.
type PasswordReset struct {
	ID        uuid.UUID  `json:"id"`
	UserID    uuid.UUID  `json:"user_id"`
	TokenHash string     `json:"-"`
	ExpiresAt time.Time  `json:"expires_at"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// QRLoginToken represents a one-time token for scanner-app QR-based login.
type QRLoginToken struct {
	ID        uuid.UUID `json:"id"`
	TenantID  uuid.UUID `json:"tenant_id"`
	UserID    uuid.UUID `json:"user_id"`
	Token     string    `json:"-"`
	ExpiresAt time.Time `json:"expires_at"`
	Used      bool      `json:"used"`
	CreatedAt time.Time `json:"created_at"`
}

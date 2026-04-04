package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// User.IsLocked
// ---------------------------------------------------------------------------

func TestUserIsLocked(t *testing.T) {
	now := time.Now()
	past := now.Add(-1 * time.Hour)
	future := now.Add(1 * time.Hour)

	tests := []struct {
		name        string
		lockedUntil *time.Time
		want        bool
	}{
		{name: "nil locked_until means not locked", lockedUntil: nil, want: false},
		{name: "past locked_until means not locked", lockedUntil: &past, want: false},
		{name: "future locked_until means locked", lockedUntil: &future, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &User{LockedUntil: tt.lockedUntil}
			if got := u.IsLocked(); got != tt.want {
				t.Errorf("IsLocked() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// User.FullName
// ---------------------------------------------------------------------------

func TestUserFullName(t *testing.T) {
	tests := []struct {
		name      string
		firstName string
		lastName  string
		want      string
	}{
		{name: "both names", firstName: "Janis", lastName: "Eckersberger", want: "Janis Eckersberger"},
		{name: "first only", firstName: "Janis", lastName: "", want: "Janis"},
		{name: "last only", firstName: "", lastName: "Eckersberger", want: "Eckersberger"},
		{name: "neither", firstName: "", lastName: "", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &User{FirstName: tt.firstName, LastName: tt.lastName}
			if got := u.FullName(); got != tt.want {
				t.Errorf("FullName() = %q, want %q", got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Role constants
// ---------------------------------------------------------------------------

func TestRoleConstants(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{"RoleAdmin", RoleAdmin, "admin"},
		{"RoleManager", RoleManager, "manager"},
		{"RoleUser", RoleUser, "user"},
		{"RoleViewer", RoleViewer, "viewer"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("%s = %q, want %q", tt.name, tt.got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// QRLoginToken — basic field validation
// ---------------------------------------------------------------------------

func TestQRLoginTokenFields(t *testing.T) {
	tenantID := uuid.New()
	userID := uuid.New()

	token := &QRLoginToken{
		ID:        uuid.New(),
		TenantID:  tenantID,
		UserID:    userID,
		Token:     "abc123",
		ExpiresAt: time.Now().Add(5 * time.Minute),
		Used:      false,
	}

	if token.TenantID != tenantID {
		t.Errorf("TenantID = %v, want %v", token.TenantID, tenantID)
	}
	if token.UserID != userID {
		t.Errorf("UserID = %v, want %v", token.UserID, userID)
	}
	if token.Used {
		t.Error("new QR token should not be used")
	}
	if token.ExpiresAt.Before(time.Now()) {
		t.Error("new QR token should not be expired")
	}
}

// ---------------------------------------------------------------------------
// Session fields
// ---------------------------------------------------------------------------

func TestSessionIsActive(t *testing.T) {
	session := &Session{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		TenantID:  uuid.New(),
		TokenHash: "somehash",
		IsActive:  true,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	if !session.IsActive {
		t.Error("session should be active")
	}
	if session.ExpiresAt.Before(time.Now()) {
		t.Error("session should not be expired")
	}
}

// ---------------------------------------------------------------------------
// Error constants — verify they are non-nil and have proper codes
// ---------------------------------------------------------------------------

func TestDomainErrors(t *testing.T) {
	errors := []error{
		ErrInvalidCredentials,
		ErrAccountLocked,
		ErrAccountDisabled,
		ErrInvalidToken,
		ErrEmailTaken,
		ErrWeakPassword,
		ErrSessionExpired,
		ErrUserNotFound,
		ErrQRTokenInvalid,
		ErrQRTokenUsed,
	}

	for _, err := range errors {
		if err == nil {
			t.Error("domain error should not be nil")
		}
		if err.Error() == "" {
			t.Error("domain error message should not be empty")
		}
	}
}

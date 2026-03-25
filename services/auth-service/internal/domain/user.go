package domain

import (
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/events"
)

// AggregateRoot is the base for event sourcing
type AggregateRoot struct {
	ID      string
	Type    string
	Version int64
	Changes []interface{}
}

// UserStatus represents the status of a user
type UserStatus string

const (
	UserStatusActive   UserStatus = "active"
	UserStatusInactive UserStatus = "inactive"
	UserStatusLocked   UserStatus = "locked"
	UserStatusDeleted  UserStatus = "deleted"
)

// Role represents a user role
type Role string

const (
	RoleSuperadmin Role = "superadmin"
	RoleAdmin      Role = "admin"
	RoleManager    Role = "manager"
	RoleWarehouse  Role = "warehouse"
	RoleDriver     Role = "driver"
	RoleCrew       Role = "crew"
	RoleFreelancer Role = "freelancer"
	RoleReadOnly   Role = "readonly"
)

// User represents a user in the auth domain (Aggregate Root for event sourcing)
type User struct {
	AggregateRoot
	Email        string
	PasswordHash string
	FirstName    string
	LastName     string
	TenantID     string
	Roles        []string
	Status       UserStatus
	CreatedAt    time.Time
	UpdatedAt    time.Time
	LastLoginAt  *time.Time
	FailedLogins int
	LockedAt     *time.Time
}

// NewUser creates a new user aggregate
func NewUser(id, email, passwordHash, firstName, lastName, tenantID string) *User {
	return &User{
		AggregateRoot: AggregateRoot{
			ID:      id,
			Type:    "User",
			Version: 0,
		},
		Email:        email,
		PasswordHash: passwordHash,
		FirstName:    firstName,
		LastName:     lastName,
		TenantID:     tenantID,
		Roles:        []string{},
		Status:       UserStatusActive,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		FailedLogins: 0,
	}
}

// Register registers a new user
func (u *User) Register(email, passwordHash, firstName, lastName, tenantID string) error {
	u.Email = email
	u.PasswordHash = passwordHash
	u.FirstName = firstName
	u.LastName = lastName
	u.TenantID = tenantID
	u.Status = UserStatusActive
	u.CreatedAt = time.Now()
	u.UpdatedAt = time.Now()

	data, _ := NewEventData("UserRegistered", map[string]interface{}{
		"id":       u.ID,
		"email":    u.Email,
		"tenantID": u.TenantID,
	}, nil)
	u.Apply(*data)

	return nil
}

// RecordLogin records a successful login
func (u *User) RecordLogin() error {
	now := time.Now()
	u.LastLoginAt = &now
	u.FailedLogins = 0
	u.UpdatedAt = now

	data, _ := NewEventData("UserLoggedIn", map[string]interface{}{
		"id":          u.ID,
		"email":       u.Email,
		"lastLoginAt": u.LastLoginAt,
		"tenantID":    u.TenantID,
	}, nil)
	u.Apply(*data)

	return nil
}

// RecordFailedLogin records a failed login attempt
func (u *User) RecordFailedLogin() error {
	u.FailedLogins++
	u.UpdatedAt = time.Now()

	// Lock after 5 failed attempts
	if u.FailedLogins >= 5 {
		u.Status = UserStatusLocked
		now := time.Now()
		u.LockedAt = &now
		data, _ := NewEventData("UserLocked", map[string]interface{}{
			"id":           u.ID,
			"email":        u.Email,
			"tenantID":     u.TenantID,
			"reason":       "too_many_failed_logins",
			"failedLogins": u.FailedLogins,
		}, nil)
		u.Apply(*data)
	} else {
		data, _ := NewEventData("LoginFailed", map[string]interface{}{
			"id":           u.ID,
			"email":        u.Email,
			"tenantID":     u.TenantID,
			"failedLogins": u.FailedLogins,
		}, nil)
		u.Apply(*data)
	}

	return nil
}

// ChangePassword changes the user's password
func (u *User) ChangePassword(newPasswordHash string) error {
	u.PasswordHash = newPasswordHash
	u.UpdatedAt = time.Now()

	data, _ := NewEventData("UserPasswordChanged", map[string]interface{}{
		"id":       u.ID,
		"email":    u.Email,
		"tenantID": u.TenantID,
	}, nil)
	u.Apply(*data)

	return nil
}

// AssignRole assigns a role to the user
func (u *User) AssignRole(role string) error {
	// Check if role already exists
	for _, r := range u.Roles {
		if r == role {
			return nil // Already has this role
		}
	}

	u.Roles = append(u.Roles, role)
	u.UpdatedAt = time.Now()

	data, _ := NewEventData("UserRoleChanged", map[string]interface{}{
		"id":       u.ID,
		"email":    u.Email,
		"tenantID": u.TenantID,
		"roles":    u.Roles,
	}, nil)
	u.Apply(*data)

	return nil
}

// RemoveRole removes a role from the user
func (u *User) RemoveRole(role string) error {
	newRoles := []string{}
	for _, r := range u.Roles {
		if r != role {
			newRoles = append(newRoles, r)
		}
	}

	u.Roles = newRoles
	u.UpdatedAt = time.Now()

	data, _ := NewEventData("UserRoleChanged", map[string]interface{}{
		"id":       u.ID,
		"email":    u.Email,
		"tenantID": u.TenantID,
		"roles":    u.Roles,
	}, nil)
	u.Apply(*data)

	return nil
}

// Deactivate deactivates the user
func (u *User) Deactivate() error {
	u.Status = UserStatusInactive
	u.UpdatedAt = time.Now()

	data, _ := NewEventData("UserDeactivated", map[string]interface{}{
		"id":       u.ID,
		"email":    u.Email,
		"tenantID": u.TenantID,
	}, nil)
	u.Apply(*data)

	return nil
}

// Lock locks the user account
func (u *User) Lock() error {
	u.Status = UserStatusLocked
	u.UpdatedAt = time.Now()

	data, _ := NewEventData("UserLocked", map[string]interface{}{
		"id":       u.ID,
		"email":    u.Email,
		"tenantID": u.TenantID,
	}, nil)
	u.Apply(*data)

	return nil
}

// Unlock unlocks the user account
func (u *User) Unlock() error {
	u.Status = UserStatusActive
	u.FailedLogins = 0
	u.UpdatedAt = time.Now()

	data, _ := NewEventData("UserUnlocked", map[string]interface{}{
		"id":       u.ID,
		"email":    u.Email,
		"tenantID": u.TenantID,
	}, nil)
	u.Apply(*data)

	return nil
}

// UpdateProfile updates user profile information
func (u *User) UpdateProfile(firstName, lastName string) error {
	u.FirstName = firstName
	u.LastName = lastName
	u.UpdatedAt = time.Now()

	data, _ := NewEventData("UserProfileUpdated", map[string]interface{}{
		"id":        u.ID,
		"email":     u.Email,
		"tenantID":  u.TenantID,
		"firstName": u.FirstName,
		"lastName":  u.LastName,
	}, nil)
	u.Apply(*data)

	return nil
}

// NewEventData creates event data from arbitrary data
func NewEventData(eventType string, data interface{}, metadata interface{}) (*events.EventData, error) {
	return events.NewEventData(eventType, data, metadata)
}

// IsLocked returns true if the user account is currently locked
func (u *User) IsLocked() bool {
	return u.Status == UserStatusLocked
}

// ShouldAutoUnlock checks if the user should be automatically unlocked (15 minutes have passed)
func (u *User) ShouldAutoUnlock() bool {
	if u.Status != UserStatusLocked || u.LockedAt == nil {
		return false
	}
	// Auto-unlock after 15 minutes
	return time.Since(*u.LockedAt) > 15*time.Minute
}

// Apply adds an event to the uncommitted changes
func (u *User) Apply(event events.EventData) {
	u.Changes = append(u.Changes, event)
}

// RolesToString converts a slice of roles to a comma-separated string for storage
func RolesToString(roles []string) string {
	result := ""
	for i, role := range roles {
		if i > 0 {
			result += ","
		}
		result += role
	}
	return result
}

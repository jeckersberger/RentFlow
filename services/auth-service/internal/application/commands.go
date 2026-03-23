package application

// RegisterUserCommand represents a user registration command
type RegisterUserCommand struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	TenantID  string `json:"tenant_id"`
}

// LoginCommand represents a login command
type LoginCommand struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// ChangePasswordCommand represents a password change command
type ChangePasswordCommand struct {
	UserID      string `json:"user_id"`
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

// UpdateProfileCommand represents a profile update command
type UpdateProfileCommand struct {
	UserID    string `json:"user_id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// AssignRoleCommand represents a role assignment command
type AssignRoleCommand struct {
	UserID   string `json:"user_id"`
	Role     string `json:"role"`
	TenantID string `json:"tenant_id"`
}

// RemoveRoleCommand represents a role removal command
type RemoveRoleCommand struct {
	UserID   string `json:"user_id"`
	Role     string `json:"role"`
	TenantID string `json:"tenant_id"`
}

// DeactivateUserCommand represents a user deactivation command
type DeactivateUserCommand struct {
	UserID   string `json:"user_id"`
	TenantID string `json:"tenant_id"`
}

// UnlockUserCommand represents a user unlock command
type UnlockUserCommand struct {
	UserID   string `json:"user_id"`
	TenantID string `json:"tenant_id"`
}

// CreateTenantCommand represents a tenant creation command
type CreateTenantCommand struct {
	Name            string  `json:"name"`
	Slug            string  `json:"slug"`
	DefaultLanguage string  `json:"default_language"`
	Currency        string  `json:"currency"`
	TaxRate         float64 `json:"tax_rate"`
	InvoicePrefix   string  `json:"invoice_prefix"`
}

// InviteUserCommand represents an employee invitation command
type InviteUserCommand struct {
	TenantID  string `json:"tenant_id"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	InvitedBy string `json:"invited_by"`
}

// AcceptInvitationCommand represents accepting an invitation
type AcceptInvitationCommand struct {
	Token     string `json:"token"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// UpdateTenantCommand represents a tenant update command
type UpdateTenantCommand struct {
	TenantID        string  `json:"tenant_id"`
	Name            string  `json:"name"`
	Slug            string  `json:"slug"`
	DefaultLanguage string  `json:"default_language"`
	Currency        string  `json:"currency"`
	TaxRate         float64 `json:"tax_rate"`
	InvoicePrefix   string  `json:"invoice_prefix"`
}

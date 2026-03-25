package application

// UserDTO represents a user in responses
type UserDTO struct {
	ID          string   `json:"id"`
	Email       string   `json:"email"`
	FirstName   string   `json:"first_name"`
	LastName    string   `json:"last_name"`
	Roles       []string `json:"roles"`
	TenantID    string   `json:"tenant_id"`
	Status      string   `json:"status"`
	LastLoginAt *string  `json:"last_login_at"`
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
}

// TokenPair represents access and refresh tokens
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// TenantDTO represents a tenant in responses
type TenantDTO struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	Slug            string  `json:"slug"`
	Status          string  `json:"status"`
	DefaultLanguage string  `json:"default_language"`
	Currency        string  `json:"currency"`
	TaxRate         float64 `json:"tax_rate"`
	InvoicePrefix   string  `json:"invoice_prefix"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
}

// PaginatedResult represents a paginated response
type PaginatedResult struct {
	Data       interface{} `json:"data"`
	Page       int         `json:"page"`
	PerPage    int         `json:"per_page"`
	Total      int         `json:"total"`
	TotalPages int         `json:"total_pages"`
}

// InvitationDTO represents an invitation in responses
type InvitationDTO struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	Status    string `json:"status"`
	InvitedBy string `json:"invited_by"`
	ExpiresAt string `json:"expires_at"`
	CreatedAt string `json:"created_at"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

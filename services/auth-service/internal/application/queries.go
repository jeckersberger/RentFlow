package application

// GetUserByIDQuery represents a query to get a user by ID
type GetUserByIDQuery struct {
	UserID   string
	TenantID string
}

// GetUserByEmailQuery represents a query to get a user by email
type GetUserByEmailQuery struct {
	Email    string
	TenantID string
}

// ListUsersQuery represents a query to list users
type ListUsersQuery struct {
	TenantID string
	Page     int
	PerPage  int
}

// GetTenantByIDQuery represents a query to get a tenant by ID
type GetTenantByIDQuery struct {
	TenantID string
}

// GetTenantBySlugQuery represents a query to get a tenant by slug
type GetTenantBySlugQuery struct {
	Slug string
}

// ListTenantsQuery represents a query to list tenants
type ListTenantsQuery struct {
	Page    int
	PerPage int
}

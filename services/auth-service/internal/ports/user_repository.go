package ports

import (
	"context"

	"github.com/jeckersberger/rentflow/services/auth-service/internal/domain"
)

// UserRepository defines the interface for user data access
// This is a port that adapters will implement
type UserRepository interface {
	// FindByID retrieves a user by ID
	FindByID(ctx context.Context, id string) (*domain.User, error)

	// FindByEmail retrieves a user by email within a tenant
	FindByEmail(ctx context.Context, tenantID, email string) (*domain.User, error)

	// FindByUsername retrieves a user by username
	FindByUsername(ctx context.Context, tenantID, username string) (*domain.User, error)

	// FindByLogin retrieves a user by username or email (auto-detect)
	FindByLogin(ctx context.Context, login string) (*domain.User, error)

	// List retrieves users in a tenant with pagination
	List(ctx context.Context, tenantID string, page, perPage int) ([]*domain.User, int, error)

	// Save persists a user (creates or updates)
	Save(ctx context.Context, user *domain.User) error

	// Delete deletes a user
	Delete(ctx context.Context, id string) error
}

// TenantRepository defines the interface for tenant data access
type TenantRepository interface {
	// FindByID retrieves a tenant by ID
	FindByID(ctx context.Context, id string) (*domain.Tenant, error)

	// FindBySlug retrieves a tenant by slug
	FindBySlug(ctx context.Context, slug string) (*domain.Tenant, error)

	// List retrieves tenants with pagination
	List(ctx context.Context, page, perPage int) ([]*domain.Tenant, int, error)

	// Save persists a tenant (creates or updates)
	Save(ctx context.Context, tenant *domain.Tenant) error

	// Delete deletes a tenant
	Delete(ctx context.Context, id string) error
}

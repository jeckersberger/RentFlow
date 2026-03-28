package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/services/auth/internal/domain"
)

// CreateUserRequest holds the data for creating a new user.
type CreateUserRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Role      string `json:"role"`
	Phone     string `json:"phone,omitempty"`
}

// UpdateUserRequest holds optional fields for patching a user profile.
type UpdateUserRequest struct {
	Email     *string `json:"email,omitempty"`
	FirstName *string `json:"first_name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
	Phone     *string `json:"phone,omitempty"`
	AvatarURL *string `json:"avatar_url,omitempty"`
}

// UpdateRoleRequest holds the new role for a user.
type UpdateRoleRequest struct {
	Role string `json:"role"`
}

// UserService handles admin user management operations.
type UserService struct {
	userRepo domain.UserRepository
	logger   zerolog.Logger
}

// NewUserService creates a new UserService.
func NewUserService(userRepo domain.UserRepository, logger zerolog.Logger) *UserService {
	return &UserService{
		userRepo: userRepo,
		logger:   logger,
	}
}

// List returns a paginated list of users for a tenant.
func (s *UserService) List(ctx context.Context, tenantID uuid.UUID, page, perPage int) ([]*domain.User, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	users, total, err := s.userRepo.List(ctx, tenantID, page, perPage)
	if err != nil {
		s.logger.Error().Err(err).Str("tenant_id", tenantID.String()).Msg("failed to list users")
		return nil, 0, fmt.Errorf("listing users failed: %w", err)
	}
	return users, total, nil
}

// GetByID returns a single user by ID, scoped to a tenant.
func (s *UserService) GetByID(ctx context.Context, userID, tenantID uuid.UUID) (*domain.User, error) {
	user, err := s.userRepo.GetByIDAndTenant(ctx, userID, tenantID)
	if err != nil {
		s.logger.Error().Err(err).Str("user_id", userID.String()).Msg("failed to get user")
		return nil, domain.ErrUserNotFound
	}
	return user, nil
}

// Create validates the request, hashes the password, and persists a new user.
func (s *UserService) Create(ctx context.Context, tenantID uuid.UUID, req CreateUserRequest) (*domain.User, error) {
	if req.Email == "" {
		return nil, domain.ErrInvalidEmail
	}
	if err := validatePassword(req.Password); err != nil {
		return nil, err
	}
	if !isValidRole(req.Role) {
		return nil, fmt.Errorf("invalid role: %s", req.Role)
	}

	passwordHash, err := HashPassword(req.Password)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to hash password")
		return nil, fmt.Errorf("password hashing failed: %w", err)
	}

	now := time.Now()
	user := &domain.User{
		ID:           uuid.New(),
		TenantID:     tenantID,
		Email:        req.Email,
		PasswordHash: passwordHash,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Role:         req.Role,
		Phone:        req.Phone,
		IsActive:     true,
		FailedLogins: 0,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		s.logger.Error().Err(err).Str("email", req.Email).Msg("failed to create user")
		return nil, fmt.Errorf("user creation failed: %w", err)
	}

	s.logger.Info().
		Str("user_id", user.ID.String()).
		Str("email", user.Email).
		Str("role", user.Role).
		Msg("user created")

	return user, nil
}

// Update patches non-nil fields on an existing user, scoped to a tenant.
func (s *UserService) Update(ctx context.Context, userID, tenantID uuid.UUID, req UpdateUserRequest) (*domain.User, error) {
	user, err := s.userRepo.GetByIDAndTenant(ctx, userID, tenantID)
	if err != nil {
		s.logger.Error().Err(err).Str("user_id", userID.String()).Msg("user not found for update")
		return nil, domain.ErrUserNotFound
	}

	if req.Email != nil {
		user.Email = *req.Email
	}
	if req.FirstName != nil {
		user.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		user.LastName = *req.LastName
	}
	if req.Phone != nil {
		user.Phone = *req.Phone
	}
	if req.AvatarURL != nil {
		user.AvatarURL = *req.AvatarURL
	}
	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(ctx, user); err != nil {
		s.logger.Error().Err(err).Str("user_id", userID.String()).Msg("failed to update user")
		return nil, fmt.Errorf("user update failed: %w", err)
	}

	s.logger.Info().Str("user_id", userID.String()).Msg("user updated")
	return user, nil
}

// UpdateRole changes the role for a user after validating it, scoped to a tenant.
func (s *UserService) UpdateRole(ctx context.Context, userID, tenantID uuid.UUID, req UpdateRoleRequest) error {
	if !isValidRole(req.Role) {
		return fmt.Errorf("invalid role: %s; must be one of admin, manager, user, viewer", req.Role)
	}

	user, err := s.userRepo.GetByIDAndTenant(ctx, userID, tenantID)
	if err != nil {
		s.logger.Error().Err(err).Str("user_id", userID.String()).Msg("user not found for role update")
		return domain.ErrUserNotFound
	}

	user.Role = req.Role
	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(ctx, user); err != nil {
		s.logger.Error().Err(err).Str("user_id", userID.String()).Msg("failed to update role")
		return fmt.Errorf("role update failed: %w", err)
	}

	s.logger.Info().
		Str("user_id", userID.String()).
		Str("new_role", req.Role).
		Msg("user role updated")

	return nil
}

// Deactivate marks a user as inactive, scoped to a tenant.
func (s *UserService) Deactivate(ctx context.Context, userID, tenantID uuid.UUID) error {
	user, err := s.userRepo.GetByIDAndTenant(ctx, userID, tenantID)
	if err != nil {
		s.logger.Error().Err(err).Str("user_id", userID.String()).Msg("user not found for deactivation")
		return domain.ErrUserNotFound
	}

	user.IsActive = false
	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(ctx, user); err != nil {
		s.logger.Error().Err(err).Str("user_id", userID.String()).Msg("failed to deactivate user")
		return fmt.Errorf("user deactivation failed: %w", err)
	}

	s.logger.Info().Str("user_id", userID.String()).Msg("user deactivated")
	return nil
}

// isValidRole checks whether the given role is allowed.
func isValidRole(role string) bool {
	switch role {
	case "admin", "manager", "user", "viewer":
		return true
	}
	return false
}

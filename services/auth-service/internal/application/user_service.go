package application

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/auth-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/auth-service/internal/ports"
)

// UserService handles user-related business logic
type UserService struct {
	userRepo       ports.UserRepository
	tenantRepo     ports.TenantRepository
	passwordMgr    *PasswordManager
	tokenMgr       *TokenManager
	logger         logger.Logger
}

// NewUserService creates a new user service
func NewUserService(
	userRepo ports.UserRepository,
	tenantRepo ports.TenantRepository,
	tokenSecret string,
	log logger.Logger,
) *UserService {
	return &UserService{
		userRepo:    userRepo,
		tenantRepo:  tenantRepo,
		passwordMgr: NewPasswordManager(),
		tokenMgr:    NewTokenManager(tokenSecret),
		logger:      log,
	}
}

// Register registers a new user
func (s *UserService) Register(ctx context.Context, cmd RegisterUserCommand) (*UserDTO, error) {
	// Validate tenant exists
	tenant, err := s.tenantRepo.FindByID(ctx, cmd.TenantID)
	if err != nil {
		s.logger.Error("failed to find tenant", err, "tenantID", cmd.TenantID)
		return nil, domain.ErrTenantNotFound
	}
	if tenant == nil {
		return nil, domain.ErrTenantNotFound
	}

	// Check if email already exists
	existingUser, _ := s.userRepo.FindByEmail(ctx, cmd.TenantID, cmd.Email)
	if existingUser != nil {
		return nil, domain.ErrEmailExists
	}

	// Validate password
	if err := s.passwordMgr.ValidatePassword(cmd.Password); err != nil {
		return nil, err
	}

	// Hash password
	passwordHash, err := s.passwordMgr.HashPassword(cmd.Password)
	if err != nil {
		s.logger.Error("failed to hash password", err)
		return nil, err
	}

	// Generate user ID
	userID := generateID()

	// Create user aggregate
	user := domain.NewUser(userID, cmd.Email, passwordHash, cmd.FirstName, cmd.LastName, cmd.TenantID)
	user.AssignRole("readonly") // Default role

	// Save user
	if err := s.userRepo.Save(ctx, user); err != nil {
		s.logger.Error("failed to save user", err, "email", cmd.Email)
		return nil, err
	}

	s.logger.Info("user registered", "id", userID, "email", cmd.Email, "tenantID", cmd.TenantID)

	return s.toUserDTO(user), nil
}

// Login authenticates a user and returns tokens
func (s *UserService) Login(ctx context.Context, cmd LoginCommand) (*TokenPair, error) {
	// Note: tenantID should come from context or be determined from email
	// For now, we'll search all tenants
	user, err := s.userRepo.FindByEmail(ctx, "", cmd.Email)
	if err != nil {
		s.logger.Error("failed to find user", err, "email", cmd.Email)
		return nil, domain.ErrInvalidCredentials
	}
	if user == nil {
		return nil, domain.ErrInvalidCredentials
	}

	// Check if user is locked
	if user.Status == domain.UserStatusLocked {
		return nil, domain.ErrUserLocked
	}

	// Verify password
	if !s.passwordMgr.VerifyPassword(cmd.Password, user.PasswordHash) {
		// Record failed login
		user.RecordFailedLogin()
		s.userRepo.Save(ctx, user)
		s.logger.Warn("failed login attempt", "email", cmd.Email, "failedAttempts", user.FailedLogins)
		return nil, domain.ErrInvalidCredentials
	}

	// Check if user is inactive
	if user.Status == domain.UserStatusInactive {
		return nil, domain.ErrInvalidCredentials
	}

	// Record successful login
	user.RecordLogin()
	if err := s.userRepo.Save(ctx, user); err != nil {
		s.logger.Error("failed to update login time", err, "id", user.ID)
	}

	// Generate tokens
	jti := generateID()
	accessToken, err := s.tokenMgr.CreateAccessToken(
		user.ID,
		user.TenantID,
		user.Email,
		fmt.Sprintf("%s %s", user.FirstName, user.LastName),
		user.Roles,
		jti,
		"", // ipAddress - could come from request
		"", // userAgentHash - could come from request
		jti, // sessionID
	)
	if err != nil {
		s.logger.Error("failed to create access token", err)
		return nil, err
	}

	refreshToken, err := s.tokenMgr.CreateRefreshToken(user.ID, user.TenantID, user.Email, jti)
	if err != nil {
		s.logger.Error("failed to create refresh token", err)
		return nil, err
	}

	s.logger.Info("user logged in", "id", user.ID, "email", user.Email)

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.tokenMgr.GetAccessTokenExpiry().Seconds()),
		TokenType:    "Bearer",
	}, nil
}

// RefreshToken refreshes the access token
func (s *UserService) RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error) {
	// Verify refresh token
	claims, err := s.tokenMgr.VerifyRefreshToken(refreshToken)
	if err != nil {
		s.logger.Warn("failed to verify refresh token", "error", err.Error())
		return nil, fmt.Errorf("invalid refresh token")
	}

	// Get user
	user, err := s.userRepo.FindByID(ctx, claims.Subject)
	if err != nil {
		s.logger.Error("failed to find user", err, "id", claims.Subject)
		return nil, domain.ErrUserNotFound
	}
	if user == nil {
		return nil, domain.ErrUserNotFound
	}

	// Check if user is still active
	if user.Status != domain.UserStatusActive {
		return nil, fmt.Errorf("user is not active")
	}

	// Generate new tokens
	jti := generateID()
	newAccessToken, err := s.tokenMgr.CreateAccessToken(
		user.ID,
		user.TenantID,
		user.Email,
		fmt.Sprintf("%s %s", user.FirstName, user.LastName),
		user.Roles,
		jti,
		"",
		"",
		jti,
	)
	if err != nil {
		s.logger.Error("failed to create access token", err)
		return nil, err
	}

	newRefreshToken, err := s.tokenMgr.CreateRefreshToken(user.ID, user.TenantID, user.Email, jti)
	if err != nil {
		s.logger.Error("failed to create refresh token", err)
		return nil, err
	}

	s.logger.Info("token refreshed", "id", user.ID)

	return &TokenPair{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    int64(s.tokenMgr.GetAccessTokenExpiry().Seconds()),
		TokenType:    "Bearer",
	}, nil
}

// ChangePassword changes user password
func (s *UserService) ChangePassword(ctx context.Context, cmd ChangePasswordCommand) error {
	// Get user
	user, err := s.userRepo.FindByID(ctx, cmd.UserID)
	if err != nil {
		return err
	}
	if user == nil {
		return domain.ErrUserNotFound
	}

	// Verify old password
	if !s.passwordMgr.VerifyPassword(cmd.OldPassword, user.PasswordHash) {
		return domain.ErrInvalidCredentials
	}

	// Validate new password
	if err := s.passwordMgr.ValidatePassword(cmd.NewPassword); err != nil {
		return err
	}

	// Hash new password
	newPasswordHash, err := s.passwordMgr.HashPassword(cmd.NewPassword)
	if err != nil {
		return err
	}

	// Update password
	user.ChangePassword(newPasswordHash)
	if err := s.userRepo.Save(ctx, user); err != nil {
		s.logger.Error("failed to save user", err, "id", cmd.UserID)
		return err
	}

	s.logger.Info("password changed", "id", cmd.UserID)
	return nil
}

// GetUser gets a user by ID
func (s *UserService) GetUser(ctx context.Context, id string) (*UserDTO, error) {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, domain.ErrUserNotFound
	}
	return s.toUserDTO(user), nil
}

// ListUsers lists users in a tenant
func (s *UserService) ListUsers(ctx context.Context, query ListUsersQuery) (*PaginatedResult, error) {
	users, total, err := s.userRepo.List(ctx, query.TenantID, query.Page, query.PerPage)
	if err != nil {
		s.logger.Error("failed to list users", err)
		return nil, err
	}

	userDTOs := make([]UserDTO, len(users))
	for i, user := range users {
		userDTOs[i] = *s.toUserDTO(user)
	}

	totalPages := (total + query.PerPage - 1) / query.PerPage

	return &PaginatedResult{
		Data:       userDTOs,
		Page:       query.Page,
		PerPage:    query.PerPage,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

// AssignRole assigns a role to a user
func (s *UserService) AssignRole(ctx context.Context, cmd AssignRoleCommand) error {
	user, err := s.userRepo.FindByID(ctx, cmd.UserID)
	if err != nil {
		return err
	}
	if user == nil {
		return domain.ErrUserNotFound
	}

	user.AssignRole(cmd.Role)
	if err := s.userRepo.Save(ctx, user); err != nil {
		s.logger.Error("failed to save user", err, "id", cmd.UserID)
		return err
	}

	s.logger.Info("role assigned", "id", cmd.UserID, "role", cmd.Role)
	return nil
}

// RemoveRole removes a role from a user
func (s *UserService) RemoveRole(ctx context.Context, userID, role string) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return domain.ErrUserNotFound
	}

	user.RemoveRole(role)
	if err := s.userRepo.Save(ctx, user); err != nil {
		s.logger.Error("failed to save user", err, "id", userID)
		return err
	}

	s.logger.Info("role removed", "id", userID, "role", role)
	return nil
}

// DeactivateUser deactivates a user
func (s *UserService) DeactivateUser(ctx context.Context, cmd DeactivateUserCommand) error {
	user, err := s.userRepo.FindByID(ctx, cmd.UserID)
	if err != nil {
		return err
	}
	if user == nil {
		return domain.ErrUserNotFound
	}

	user.Deactivate()
	if err := s.userRepo.Save(ctx, user); err != nil {
		s.logger.Error("failed to save user", err, "id", cmd.UserID)
		return err
	}

	s.logger.Info("user deactivated", "id", cmd.UserID)
	return nil
}

// UnlockUser unlocks a user account
func (s *UserService) UnlockUser(ctx context.Context, cmd UnlockUserCommand) error {
	user, err := s.userRepo.FindByID(ctx, cmd.UserID)
	if err != nil {
		return err
	}
	if user == nil {
		return domain.ErrUserNotFound
	}

	user.Unlock()
	if err := s.userRepo.Save(ctx, user); err != nil {
		s.logger.Error("failed to save user", err, "id", cmd.UserID)
		return err
	}

	s.logger.Info("user unlocked", "id", cmd.UserID)
	return nil
}

// UpdateProfile updates user profile
func (s *UserService) UpdateProfile(ctx context.Context, cmd UpdateProfileCommand) error {
	user, err := s.userRepo.FindByID(ctx, cmd.UserID)
	if err != nil {
		return err
	}
	if user == nil {
		return domain.ErrUserNotFound
	}

	user.UpdateProfile(cmd.FirstName, cmd.LastName)
	if err := s.userRepo.Save(ctx, user); err != nil {
		s.logger.Error("failed to save user", err, "id", cmd.UserID)
		return err
	}

	s.logger.Info("profile updated", "id", cmd.UserID)
	return nil
}

// Helper functions

func (s *UserService) toUserDTO(user *domain.User) *UserDTO {
	return &UserDTO{
		ID:        user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Roles:     user.Roles,
		TenantID:  user.TenantID,
		Status:    string(user.Status),
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
		UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
	}
}

// generateID generates a random ID
func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

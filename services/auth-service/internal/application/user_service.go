package application

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/auth-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/auth-service/internal/ports"
)

// UserService handles user-related business logic
type UserService struct {
	userRepo            ports.UserRepository
	tenantRepo          ports.TenantRepository
	invitationRepo      ports.InvitationRepository
	passwordMgr         *PasswordManager
	tokenMgr            *TokenManager
	logger              logger.Logger
	db                  *sql.DB
	notificationURL     string
}

// NewUserService creates a new user service
func NewUserService(
	userRepo ports.UserRepository,
	tenantRepo ports.TenantRepository,
	tokenMgr *TokenManager,
	log logger.Logger,
) *UserService {
	return &UserService{
		userRepo:    userRepo,
		tenantRepo:  tenantRepo,
		passwordMgr: NewPasswordManager(),
		tokenMgr:    tokenMgr,
		logger:      log,
	}
}

// SetInvitationRepo sets the invitation repository (optional dependency)
func (s *UserService) SetInvitationRepo(repo ports.InvitationRepository) {
	s.invitationRepo = repo
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

	// Check if user is locked and handle auto-unlock
	if user.IsLocked() {
		if user.ShouldAutoUnlock() {
			// Auto-unlock the account
			user.Unlock()
			if err := s.userRepo.Save(ctx, user); err != nil {
				s.logger.Error("failed to auto-unlock user", err, "id", user.ID)
			}
			s.logger.Info("user auto-unlocked", "id", user.ID, "email", user.Email)
		} else {
			return nil, domain.ErrUserLocked
		}
	}

	// Check if user is inactive or deleted
	if user.Status == domain.UserStatusInactive || user.Status == domain.UserStatusDeleted {
		return nil, domain.ErrInvalidCredentials
	}

	// Verify password
	if !s.passwordMgr.VerifyPassword(cmd.Password, user.PasswordHash) {
		// Record failed login
		user.RecordFailedLogin()
		s.userRepo.Save(ctx, user)
		s.logger.Warn("failed login attempt", "email", cmd.Email, "failedAttempts", user.FailedLogins)
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
		"",  // ipAddress - could come from request
		"",  // userAgentHash - could come from request
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

// SetDB sets the database connection for password reset queries
func (s *UserService) SetDB(db *sql.DB) {
	s.db = db
}

// SetNotificationURL sets the notification-service base URL for sending emails
func (s *UserService) SetNotificationURL(url string) {
	s.notificationURL = url
}

// ForgotPassword generates a password reset token for the given email
func (s *UserService) ForgotPassword(ctx context.Context, cmd ForgotPasswordCommand) (string, error) {
	// Find user by email (search across all tenants)
	user, err := s.userRepo.FindByEmail(ctx, "", cmd.Email)
	if err != nil || user == nil {
		// Don't reveal whether the email exists
		s.logger.Info("forgot password request for unknown email", "email", cmd.Email)
		return "", nil
	}

	if s.db == nil {
		return "", fmt.Errorf("database not configured for password resets")
	}

	// Generate reset token (32 bytes = 64 hex chars)
	tokenBytes := make([]byte, 32)
	rand.Read(tokenBytes)
	token := hex.EncodeToString(tokenBytes)

	// Save to DB
	id := generateID()
	expiresAt := time.Now().Add(1 * time.Hour)
	_, err = s.db.ExecContext(ctx,
		`INSERT INTO auth.password_resets (id, user_id, token, expires_at, created_at)
		 VALUES ($1, $2, $3, $4, $5)`,
		id, user.ID, token, expiresAt, time.Now(),
	)
	if err != nil {
		s.logger.Error("failed to save password reset token", err, "userID", user.ID)
		return "", err
	}

	// Send password reset email asynchronously (fire-and-forget)
	if s.notificationURL != "" {
		sendEmailAsync(s.notificationURL, "/api/v1/notifications/send-email/password-reset", map[string]string{
			"to":   user.Email,
			"link": "http://localhost:3000/reset-password/" + token,
		}, s.logger)
	}

	s.logger.Info("password reset token generated",
		"userID", user.ID,
		"email", user.Email,
		"tokenPrefix", token[:8]+"...",
		"expiresAt", expiresAt.Format(time.RFC3339),
	)

	return token, nil
}

// ResetPassword validates a reset token and sets a new password
func (s *UserService) ResetPassword(ctx context.Context, cmd ResetPasswordCommand) error {
	if s.db == nil {
		return fmt.Errorf("database not configured for password resets")
	}

	// Look up the token
	var id, userID string
	var expiresAt time.Time
	var usedAt sql.NullTime

	err := s.db.QueryRowContext(ctx,
		`SELECT id, user_id, expires_at, used_at
		 FROM auth.password_resets
		 WHERE token = $1`,
		cmd.Token,
	).Scan(&id, &userID, &expiresAt, &usedAt)

	if err == sql.ErrNoRows {
		return fmt.Errorf("invalid or expired reset token")
	}
	if err != nil {
		s.logger.Error("failed to look up reset token", err)
		return fmt.Errorf("invalid or expired reset token")
	}

	// Check if already used
	if usedAt.Valid {
		return fmt.Errorf("reset token has already been used")
	}

	// Check expiry
	if time.Now().After(expiresAt) {
		return fmt.Errorf("reset token has expired")
	}

	// Validate new password
	if err := s.passwordMgr.ValidatePassword(cmd.NewPassword); err != nil {
		return err
	}

	// Hash new password
	newHash, err := s.passwordMgr.HashPassword(cmd.NewPassword)
	if err != nil {
		return err
	}

	// Get user and update password
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil || user == nil {
		return domain.ErrUserNotFound
	}

	user.ChangePassword(newHash)
	if err := s.userRepo.Save(ctx, user); err != nil {
		s.logger.Error("failed to save user after password reset", err, "userID", userID)
		return err
	}

	// Mark token as used
	now := time.Now()
	_, err = s.db.ExecContext(ctx,
		`UPDATE auth.password_resets SET used_at = $1 WHERE id = $2`,
		now, id,
	)
	if err != nil {
		s.logger.Error("failed to mark reset token as used", err, "id", id)
		// Don't fail the reset — password was already changed
	}

	s.logger.Info("password reset completed", "userID", userID)
	return nil
}

// GetUserByEmail finds a user by email across all tenants and returns a DTO
func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*UserDTO, error) {
	user, err := s.userRepo.FindByEmail(ctx, "", email)
	if err != nil || user == nil {
		return nil, domain.ErrUserNotFound
	}
	return s.toUserDTO(user), nil
}

// QRLogin validates a QR token and returns tokens + user (Scanner App contract)
func (s *UserService) QRLogin(ctx context.Context, qrToken string) (*TokenPair, *UserDTO, error) {
	if qrToken == "" {
		return nil, nil, fmt.Errorf("qr_token is required")
	}

	if s.db == nil {
		return nil, nil, fmt.Errorf("database not configured for QR login")
	}

	// Look up the QR token
	var userID string
	var expiresAt time.Time
	var usedAt sql.NullTime

	err := s.db.QueryRowContext(ctx,
		`SELECT user_id, expires_at, used_at
		 FROM auth.qr_login_tokens
		 WHERE token = $1`,
		qrToken,
	).Scan(&userID, &expiresAt, &usedAt)

	if err == sql.ErrNoRows {
		return nil, nil, domain.ErrInvalidCredentials
	}
	if err != nil {
		s.logger.Error("failed to look up QR token", err)
		return nil, nil, domain.ErrInvalidCredentials
	}

	// Check if already used
	if usedAt.Valid {
		return nil, nil, fmt.Errorf("QR token already used")
	}

	// Check expiry
	if time.Now().After(expiresAt) {
		return nil, nil, fmt.Errorf("QR token expired")
	}

	// Mark token as used
	now := time.Now()
	_, _ = s.db.ExecContext(ctx,
		`UPDATE auth.qr_login_tokens SET used_at = $1 WHERE token = $2`,
		now, qrToken,
	)

	// Get user
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil || user == nil {
		return nil, nil, domain.ErrUserNotFound
	}

	if user.Status != domain.UserStatusActive {
		return nil, nil, domain.ErrInvalidCredentials
	}

	// Record login
	user.RecordLogin()
	_ = s.userRepo.Save(ctx, user)

	// Generate tokens
	jti := generateID()
	accessToken, err := s.tokenMgr.CreateAccessToken(
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
		return nil, nil, err
	}

	refreshToken, err := s.tokenMgr.CreateRefreshToken(user.ID, user.TenantID, user.Email, jti)
	if err != nil {
		return nil, nil, err
	}

	s.logger.Info("QR login successful", "id", user.ID, "email", user.Email)

	tokens := &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.tokenMgr.GetAccessTokenExpiry().Seconds()),
		TokenType:    "Bearer",
	}

	return tokens, s.toUserDTO(user), nil
}

// GenerateQRToken creates a one-time QR login token for the Scanner App
func (s *UserService) GenerateQRToken(ctx context.Context, userID, tenantID string) (string, time.Time, error) {
	if s.db == nil {
		return "", time.Time{}, fmt.Errorf("database not configured for QR tokens")
	}

	// Generate 32-byte hex token
	tokenBytes := make([]byte, 32)
	rand.Read(tokenBytes)
	token := hex.EncodeToString(tokenBytes)

	id := generateID()
	expiresAt := time.Now().Add(5 * time.Minute)

	_, err := s.db.ExecContext(ctx,
		`INSERT INTO auth.qr_login_tokens (id, token, user_id, tenant_id, expires_at, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		id, token, userID, tenantID, expiresAt, time.Now(),
	)
	if err != nil {
		s.logger.Error("failed to create QR login token", err, "userID", userID)
		return "", time.Time{}, err
	}

	s.logger.Info("QR login token generated", "userID", userID, "expiresAt", expiresAt.Format(time.RFC3339))
	return token, expiresAt, nil
}

// QRTokenStatus checks if a QR token has been redeemed (for browser polling)
func (s *UserService) QRTokenStatus(ctx context.Context, qrToken string) (string, bool, error) {
	if s.db == nil {
		return "", false, fmt.Errorf("database not configured for QR tokens")
	}

	var expiresAt time.Time
	var usedAt sql.NullTime
	var status string

	err := s.db.QueryRowContext(ctx,
		`SELECT expires_at, used_at
		 FROM auth.qr_login_tokens
		 WHERE token = $1`,
		qrToken,
	).Scan(&expiresAt, &usedAt)

	if err == sql.ErrNoRows {
		return "not_found", false, nil
	}
	if err != nil {
		s.logger.Error("failed to look up QR token status", err)
		return "error", false, err
	}

	if usedAt.Valid {
		status = "redeemed"
		return status, true, nil
	}

	if time.Now().After(expiresAt) {
		status = "expired"
		return status, false, nil
	}

	status = "pending"
	return status, false, nil
}

// Helper functions

func (s *UserService) toUserDTO(user *domain.User) *UserDTO {
	dto := &UserDTO{
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
	if user.LastLoginAt != nil {
		formatted := user.LastLoginAt.Format(time.RFC3339)
		dto.LastLoginAt = &formatted
	}
	return dto
}

// ActivateUser reactivates a deactivated user
func (s *UserService) ActivateUser(ctx context.Context, cmd DeactivateUserCommand) error {
	user, err := s.userRepo.FindByID(ctx, cmd.UserID)
	if err != nil {
		return err
	}
	if user == nil {
		return domain.ErrUserNotFound
	}

	user.Status = domain.UserStatusActive
	user.UpdatedAt = time.Now()
	if err := s.userRepo.Save(ctx, user); err != nil {
		s.logger.Error("failed to activate user", err, "id", cmd.UserID)
		return err
	}

	s.logger.Info("user activated", "id", cmd.UserID)
	return nil
}

// SoftDeleteUser marks a user as deleted
func (s *UserService) SoftDeleteUser(ctx context.Context, cmd DeactivateUserCommand) error {
	user, err := s.userRepo.FindByID(ctx, cmd.UserID)
	if err != nil {
		return err
	}
	if user == nil {
		return domain.ErrUserNotFound
	}

	user.Status = domain.UserStatusDeleted
	user.UpdatedAt = time.Now()
	if err := s.userRepo.Save(ctx, user); err != nil {
		s.logger.Error("failed to soft-delete user", err, "id", cmd.UserID)
		return err
	}

	s.logger.Info("user soft-deleted", "id", cmd.UserID)
	return nil
}

// InviteUser creates an invitation for a new employee
func (s *UserService) InviteUser(ctx context.Context, cmd InviteUserCommand) (*InvitationDTO, error) {
	if s.invitationRepo == nil {
		return nil, fmt.Errorf("invitation system not configured")
	}

	// Check if user already exists
	existing, _ := s.userRepo.FindByEmail(ctx, cmd.TenantID, cmd.Email)
	if existing != nil {
		return nil, fmt.Errorf("user with this email already exists")
	}

	// Check for existing pending invitation
	existingInv, _ := s.invitationRepo.FindByEmail(ctx, cmd.TenantID, cmd.Email)
	if existingInv != nil {
		return nil, fmt.Errorf("invitation already pending for this email")
	}

	// Generate invitation token
	tokenBytes := make([]byte, 32)
	rand.Read(tokenBytes)
	token := hex.EncodeToString(tokenBytes)

	role := cmd.Role
	if role == "" {
		role = "readonly"
	}

	inv := &ports.Invitation{
		ID:        generateID(),
		TenantID:  cmd.TenantID,
		Email:     cmd.Email,
		Role:      role,
		Token:     token,
		Status:    "pending",
		InvitedBy: cmd.InvitedBy,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour), // 7 days
		CreatedAt: time.Now(),
	}

	if err := s.invitationRepo.Save(ctx, inv); err != nil {
		s.logger.Error("failed to save invitation", err, "email", cmd.Email)
		return nil, err
	}

	// Send invitation email asynchronously (fire-and-forget)
	if s.notificationURL != "" {
		sendEmailAsync(s.notificationURL, "/api/v1/notifications/send-email/invitation", map[string]string{
			"to":           inv.Email,
			"inviter_name": "Admin",
			"link":         "http://localhost:3000/register?token=" + inv.Token,
		}, s.logger)
	}

	s.logger.Info("user invited", "email", cmd.Email, "role", role, "tenantID", cmd.TenantID)

	return &InvitationDTO{
		ID:        inv.ID,
		Email:     inv.Email,
		Role:      inv.Role,
		Status:    inv.Status,
		InvitedBy: inv.InvitedBy,
		ExpiresAt: inv.ExpiresAt.Format(time.RFC3339),
		CreatedAt: inv.CreatedAt.Format(time.RFC3339),
	}, nil
}

// AcceptInvitation allows a user to accept an invitation and set their password
func (s *UserService) AcceptInvitation(ctx context.Context, cmd AcceptInvitationCommand) (*UserDTO, error) {
	if s.invitationRepo == nil {
		return nil, fmt.Errorf("invitation system not configured")
	}

	inv, err := s.invitationRepo.FindByToken(ctx, cmd.Token)
	if err != nil || inv == nil {
		return nil, fmt.Errorf("invalid or expired invitation")
	}

	if inv.Status != "pending" {
		return nil, fmt.Errorf("invitation already used")
	}

	if time.Now().After(inv.ExpiresAt) {
		return nil, fmt.Errorf("invitation expired")
	}

	// Register the user
	userDTO, err := s.Register(ctx, RegisterUserCommand{
		Email:     inv.Email,
		Password:  cmd.Password,
		FirstName: cmd.FirstName,
		LastName:  cmd.LastName,
		TenantID:  inv.TenantID,
	})
	if err != nil {
		return nil, err
	}

	// Assign the invited role
	if inv.Role != "readonly" {
		s.AssignRole(ctx, AssignRoleCommand{
			UserID:   userDTO.ID,
			Role:     inv.Role,
			TenantID: inv.TenantID,
		})
	}

	// Mark invitation as accepted
	now := time.Now()
	inv.Status = "accepted"
	inv.ClaimedAt = &now
	s.invitationRepo.Save(ctx, inv)

	s.logger.Info("invitation accepted", "email", inv.Email, "userID", userDTO.ID)

	return userDTO, nil
}

// ListInvitations lists all invitations for a tenant
func (s *UserService) ListInvitations(ctx context.Context, tenantID string) ([]InvitationDTO, error) {
	if s.invitationRepo == nil {
		return []InvitationDTO{}, nil
	}

	invitations, err := s.invitationRepo.ListByTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	dtos := make([]InvitationDTO, len(invitations))
	for i, inv := range invitations {
		dtos[i] = InvitationDTO{
			ID:        inv.ID,
			Email:     inv.Email,
			Role:      inv.Role,
			Status:    inv.Status,
			InvitedBy: inv.InvitedBy,
			ExpiresAt: inv.ExpiresAt.Format(time.RFC3339),
			CreatedAt: inv.CreatedAt.Format(time.RFC3339),
		}
	}
	return dtos, nil
}

// generateID generates a random ID
func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

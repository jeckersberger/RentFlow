package application

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"golang.org/x/crypto/argon2"

	"github.com/jeckersberger/EquipFlow/pkg/common/middleware"
	"github.com/jeckersberger/EquipFlow/services/auth/internal/domain"
)

// Argon2id parameters.
const (
	argon2Time    = 1
	argon2Memory  = 64 * 1024
	argon2Threads = 4
	argon2KeyLen  = 32
	argon2SaltLen = 16
)

// TokenPair holds access and refresh tokens returned after authentication.
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
}

// LoginRequest represents a login attempt.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RefreshRequest carries a refresh token for token rotation.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// ChangePasswordRequest carries old and new passwords.
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

// AuthService orchestrates authentication and token lifecycle.
type AuthService struct {
	userRepo      domain.UserRepository
	sessionRepo   domain.SessionRepository
	qrLoginRepo   domain.QRLoginRepository
	resetRepo     domain.PasswordResetRepository
	privateKey    *rsa.PrivateKey
	publicKey     *rsa.PublicKey
	accessExpiry  time.Duration
	refreshExpiry time.Duration
	logger        zerolog.Logger
}

// NewAuthService creates a new AuthService with all dependencies.
func NewAuthService(
	userRepo domain.UserRepository,
	sessionRepo domain.SessionRepository,
	qrLoginRepo domain.QRLoginRepository,
	resetRepo domain.PasswordResetRepository,
	privateKey *rsa.PrivateKey,
	publicKey *rsa.PublicKey,
	accessExpiry time.Duration,
	refreshExpiry time.Duration,
	logger zerolog.Logger,
) *AuthService {
	if accessExpiry == 0 {
		accessExpiry = 1 * time.Hour
	}
	if refreshExpiry == 0 {
		refreshExpiry = 30 * 24 * time.Hour
	}
	return &AuthService{
		userRepo:      userRepo,
		sessionRepo:   sessionRepo,
		qrLoginRepo:   qrLoginRepo,
		resetRepo:     resetRepo,
		privateKey:    privateKey,
		publicKey:     publicKey,
		accessExpiry:  accessExpiry,
		refreshExpiry: refreshExpiry,
		logger:        logger,
	}
}

// Login authenticates a user and returns a token pair plus the user.
func (s *AuthService) Login(
	ctx context.Context,
	tenantID uuid.UUID,
	req LoginRequest,
	ipAddress, userAgent string,
) (*TokenPair, *domain.User, error) {
	user, err := s.userRepo.GetByEmail(ctx, tenantID, req.Email)
	if err != nil {
		s.logger.Warn().Str("email", req.Email).Msg("user not found during login")
		return nil, nil, domain.ErrInvalidCredentials
	}

	if !user.IsActive {
		s.logger.Warn().Str("user_id", user.ID.String()).Msg("login attempt on disabled account")
		return nil, nil, domain.ErrAccountDisabled
	}

	if user.IsLocked() {
		s.logger.Warn().Str("user_id", user.ID.String()).Msg("login attempt on locked account")
		return nil, nil, domain.ErrAccountLocked
	}

	if !VerifyPassword(user.PasswordHash, req.Password) {
		// Use the DB-returned count (atomic) instead of stale in-memory value
		dbCount, err := s.userRepo.IncrementFailedLogins(ctx, user.ID)
		if err != nil {
			s.logger.Error().Err(err).Msg("failed to increment failed logins")
			dbCount = user.FailedLogins + 1 // fallback to in-memory estimate
		}
		s.applyBruteForceLock(ctx, user.ID, dbCount)
		return nil, nil, domain.ErrInvalidCredentials
	}

	if err := s.userRepo.ResetFailedLogins(ctx, user.ID); err != nil {
		s.logger.Error().Err(err).Msg("failed to reset failed logins")
	}
	if err := s.userRepo.UpdateLastLogin(ctx, user.ID); err != nil {
		s.logger.Error().Err(err).Msg("failed to update last login")
	}

	accessToken, expiresAt, err := s.generateAccessToken(user)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to generate access token")
		return nil, nil, fmt.Errorf("token generation failed: %w", err)
	}

	refreshToken, err := generateRefreshToken()
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to generate refresh token")
		return nil, nil, fmt.Errorf("refresh token generation failed: %w", err)
	}

	tokenHash := hashToken(refreshToken)
	session := &domain.Session{
		ID:        uuid.New(),
		UserID:    user.ID,
		TenantID:  tenantID,
		TokenHash: tokenHash,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		IsActive:  true,
		ExpiresAt: time.Now().Add(s.refreshExpiry),
		CreatedAt: time.Now(),
	}
	if err := s.sessionRepo.Create(ctx, session); err != nil {
		s.logger.Error().Err(err).Msg("failed to create session")
		return nil, nil, fmt.Errorf("session creation failed: %w", err)
	}

	s.logger.Info().Str("user_id", user.ID.String()).Msg("user logged in")

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
	}, user, nil
}

// Refresh rotates the refresh token and issues a new access token.
func (s *AuthService) Refresh(ctx context.Context, req RefreshRequest) (*TokenPair, error) {
	tokenHash := hashToken(req.RefreshToken)

	session, err := s.sessionRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil || session == nil || !session.IsActive || time.Now().After(session.ExpiresAt) {
		return nil, domain.ErrSessionExpired
	}

	user, err := s.userRepo.GetByID(ctx, session.UserID)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to get user for token refresh")
		return nil, domain.ErrSessionExpired
	}

	accessToken, expiresAt, err := s.generateAccessToken(user)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to generate access token on refresh")
		return nil, fmt.Errorf("token generation failed: %w", err)
	}

	newRefreshToken, err := generateRefreshToken()
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to generate new refresh token")
		return nil, fmt.Errorf("refresh token generation failed: %w", err)
	}

	newTokenHash := hashToken(newRefreshToken)
	if err := s.sessionRepo.UpdateTokenHash(ctx, session.ID, newTokenHash); err != nil {
		s.logger.Error().Err(err).Msg("failed to update session token hash")
		return nil, fmt.Errorf("session update failed: %w", err)
	}

	s.logger.Info().Str("user_id", session.UserID.String()).Msg("token refreshed")

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresAt:    expiresAt,
	}, nil
}

// Logout deactivates the session associated with the given refresh token.
func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	tokenHash := hashToken(refreshToken)

	session, err := s.sessionRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil || session == nil {
		return nil // idempotent — already gone
	}

	if err := s.sessionRepo.Deactivate(ctx, session.ID); err != nil {
		s.logger.Error().Err(err).Msg("failed to deactivate session")
		return fmt.Errorf("logout failed: %w", err)
	}

	s.logger.Info().Str("session_id", session.ID.String()).Msg("session deactivated")
	return nil
}

// ChangePassword verifies the old password, hashes the new one, and invalidates all sessions.
func (s *AuthService) ChangePassword(ctx context.Context, userID uuid.UUID, req ChangePasswordRequest) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.logger.Error().Err(err).Msg("user not found for password change")
		return domain.ErrUserNotFound
	}

	if !VerifyPassword(user.PasswordHash, req.OldPassword) {
		return domain.ErrInvalidCredentials
	}

	if err := validatePassword(req.NewPassword); err != nil {
		return err
	}

	newHash, err := HashPassword(req.NewPassword)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to hash new password")
		return fmt.Errorf("password hashing failed: %w", err)
	}

	if err := s.userRepo.UpdatePassword(ctx, userID, newHash); err != nil {
		s.logger.Error().Err(err).Msg("failed to update password")
		return fmt.Errorf("password update failed: %w", err)
	}

	if err := s.sessionRepo.DeactivateAllForUser(ctx, userID); err != nil {
		s.logger.Error().Err(err).Msg("failed to deactivate sessions after password change")
	}

	s.logger.Info().Str("user_id", userID.String()).Msg("password changed, all sessions invalidated")
	return nil
}

// ValidateToken parses and validates a JWT, returning the embedded claims.
func (s *AuthService) ValidateToken(ctx context.Context, tokenString string) (*middleware.Claims, error) {
	claims := &middleware.Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.publicKey, nil
	})
	if err != nil || !token.Valid {
		return nil, domain.ErrInvalidToken
	}
	return claims, nil
}

// GenerateQRToken creates a one-time QR login token for the given user.
// The token is valid for 5 minutes and can be exchanged once for JWTs.
func (s *AuthService) GenerateQRToken(ctx context.Context, tenantID, userID uuid.UUID) (*domain.QRLoginToken, error) {
	// Verify the target user exists and is active.
	user, err := s.userRepo.GetByIDAndTenant(ctx, userID, tenantID)
	if err != nil {
		s.logger.Error().Err(err).Str("user_id", userID.String()).Msg("user not found for QR token generation")
		return nil, domain.ErrUserNotFound
	}
	if !user.IsActive {
		return nil, domain.ErrAccountDisabled
	}

	// Generate a random 32-byte hex token (64 chars).
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		s.logger.Error().Err(err).Msg("failed to generate QR token")
		return nil, fmt.Errorf("QR token generation failed: %w", err)
	}
	token := hex.EncodeToString(tokenBytes)

	qrToken := &domain.QRLoginToken{
		ID:        uuid.New(),
		TenantID:  tenantID,
		UserID:    userID,
		Token:     token,
		ExpiresAt: time.Now().Add(5 * time.Minute),
		Used:      false,
	}

	if err := s.qrLoginRepo.Create(ctx, qrToken); err != nil {
		s.logger.Error().Err(err).Msg("failed to store QR login token")
		return nil, fmt.Errorf("QR token storage failed: %w", err)
	}

	s.logger.Info().
		Str("user_id", userID.String()).
		Str("token_id", qrToken.ID.String()).
		Msg("QR login token generated")

	return qrToken, nil
}

// ValidateQRToken validates a one-time QR token, marks it as used,
// and returns a JWT token pair plus the associated user.
func (s *AuthService) ValidateQRToken(
	ctx context.Context,
	token string,
	ipAddress, userAgent string,
) (*TokenPair, *domain.User, error) {
	// Atomic: mark used + return in single UPDATE ... RETURNING (prevents TOCTOU race)
	qrToken, err := s.qrLoginRepo.ClaimToken(ctx, token)
	if err != nil {
		s.logger.Warn().Str("token", token[:min(8, len(token))]).Msg("QR token invalid, expired, or already used")
		return nil, nil, domain.ErrQRTokenInvalid
	}

	// Load the user.
	user, err := s.userRepo.GetByIDAndTenant(ctx, qrToken.UserID, qrToken.TenantID)
	if err != nil {
		s.logger.Error().Err(err).Msg("user not found for QR login")
		return nil, nil, domain.ErrUserNotFound
	}

	if !user.IsActive {
		return nil, nil, domain.ErrAccountDisabled
	}

	if user.IsLocked() {
		return nil, nil, domain.ErrAccountLocked
	}

	// Update last login timestamp.
	if err := s.userRepo.UpdateLastLogin(ctx, user.ID); err != nil {
		s.logger.Error().Err(err).Msg("failed to update last login for QR login")
	}

	// Generate JWT access token.
	accessToken, expiresAt, err := s.generateAccessToken(user)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to generate access token for QR login")
		return nil, nil, fmt.Errorf("token generation failed: %w", err)
	}

	// Generate refresh token.
	refreshToken, err := generateRefreshToken()
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to generate refresh token for QR login")
		return nil, nil, fmt.Errorf("refresh token generation failed: %w", err)
	}

	// Create session.
	tokenHash := hashToken(refreshToken)
	session := &domain.Session{
		ID:        uuid.New(),
		UserID:    user.ID,
		TenantID:  qrToken.TenantID,
		TokenHash: tokenHash,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		IsActive:  true,
		ExpiresAt: time.Now().Add(s.refreshExpiry),
		CreatedAt: time.Now(),
	}
	if err := s.sessionRepo.Create(ctx, session); err != nil {
		s.logger.Error().Err(err).Msg("failed to create session for QR login")
		return nil, nil, fmt.Errorf("session creation failed: %w", err)
	}

	s.logger.Info().
		Str("user_id", user.ID.String()).
		Str("token_id", qrToken.ID.String()).
		Msg("user logged in via QR token")

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
	}, user, nil
}

// ForgotPassword generates a reset token, stores the hash, and logs the plain token.
// Always returns nil to prevent email enumeration.
func (s *AuthService) ForgotPassword(ctx context.Context, email string, tenantID uuid.UUID) error {
	user, err := s.userRepo.GetByEmail(ctx, tenantID, email)
	if err != nil {
		// User not found — return nil to prevent email enumeration.
		s.logger.Debug().Str("email", email).Msg("forgot-password: user not found, returning success anyway")
		return nil
	}

	if !user.IsActive {
		s.logger.Debug().Str("user_id", user.ID.String()).Msg("forgot-password: account disabled, returning success anyway")
		return nil
	}

	// Generate a random 32-byte token.
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		s.logger.Error().Err(err).Msg("forgot-password: failed to generate token")
		return nil
	}
	plainToken := hex.EncodeToString(tokenBytes)

	// Hash token with SHA256 for storage.
	tokenHash := hashToken(plainToken)

	resetToken := &domain.PasswordResetToken{
		ID:        uuid.New(),
		TenantID:  tenantID,
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(1 * time.Hour),
		Used:      false,
	}

	if err := s.resetRepo.Create(ctx, resetToken); err != nil {
		s.logger.Error().Err(err).Msg("forgot-password: failed to store reset token")
		return nil
	}

	// In production this would be sent via the email/notification service.
	s.logger.Info().
		Str("user_id", user.ID.String()).
		Str("email", user.Email).
		Str("reset_token", plainToken).
		Msg("password reset token generated (send via email in production)")

	return nil
}

// ResetPassword validates the plain token and updates the user's password.
func (s *AuthService) ResetPassword(ctx context.Context, plainToken, newPassword string) error {
	// Hash the provided token to look up in the database.
	tokenHash := hashToken(plainToken)

	resetToken, err := s.resetRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		s.logger.Warn().Msg("reset-password: invalid or expired token")
		return domain.ErrResetTokenInvalid
	}

	// Defensive checks (repo already filters, but be explicit).
	if resetToken.Used || time.Now().After(resetToken.ExpiresAt) {
		return domain.ErrResetTokenInvalid
	}

	// Validate password strength.
	if err := validatePassword(newPassword); err != nil {
		return err
	}

	// Hash the new password with Argon2id.
	newHash, err := HashPassword(newPassword)
	if err != nil {
		s.logger.Error().Err(err).Msg("reset-password: failed to hash new password")
		return fmt.Errorf("password hashing failed: %w", err)
	}

	// Update the user's password.
	if err := s.userRepo.UpdatePassword(ctx, resetToken.UserID, newHash); err != nil {
		s.logger.Error().Err(err).Msg("reset-password: failed to update password")
		return fmt.Errorf("password update failed: %w", err)
	}

	// Mark the reset token as used.
	if err := s.resetRepo.MarkUsed(ctx, resetToken.ID); err != nil {
		s.logger.Error().Err(err).Msg("reset-password: failed to mark token as used")
	}

	// Invalidate all active sessions for this user.
	if err := s.sessionRepo.DeactivateAllForUser(ctx, resetToken.UserID); err != nil {
		s.logger.Error().Err(err).Msg("reset-password: failed to invalidate sessions")
	}

	s.logger.Info().
		Str("user_id", resetToken.UserID.String()).
		Msg("password reset completed, all sessions invalidated")

	return nil
}

// --- private helpers ---

func (s *AuthService) generateAccessToken(user *domain.User) (string, int64, error) {
	now := time.Now()
	expiresAt := now.Add(s.accessExpiry)

	claims := &middleware.Claims{
		UserID:   user.ID,
		TenantID: user.TenantID,
		Email:    user.Email,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "cratedesk-auth",
			Subject:   user.ID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signed, err := token.SignedString(s.privateKey)
	if err != nil {
		return "", 0, err
	}
	return signed, expiresAt.Unix(), nil
}

func generateRefreshToken() (string, error) {
	b := make([]byte, 64)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

// HashPassword hashes a password using Argon2id and returns the PHC-format string.
func HashPassword(password string) (string, error) {
	salt := make([]byte, argon2SaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	hash := argon2.IDKey([]byte(password), salt, argon2Time, argon2Memory, argon2Threads, argon2KeyLen)

	saltB64 := base64.RawStdEncoding.EncodeToString(salt)
	hashB64 := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		argon2Memory, argon2Time, argon2Threads, saltB64, hashB64,
	), nil
}

// VerifyPassword verifies a password against an Argon2id PHC-format hash.
func VerifyPassword(phcHash, password string) bool {
	parts := strings.Split(phcHash, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return false
	}

	var m, t, p uint32
	_, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &m, &t, &p)
	if err != nil {
		return false
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false
	}

	expectedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false
	}

	computed := argon2.IDKey([]byte(password), salt, t, m, uint8(p), uint32(len(expectedHash)))

	if len(computed) != len(expectedHash) {
		return false
	}
	// Constant-time comparison to prevent timing attacks.
	var diff byte
	for i := range computed {
		diff |= computed[i] ^ expectedHash[i]
	}
	return diff == 0
}

func validatePassword(password string) error {
	if len(password) < 8 {
		return domain.ErrWeakPassword
	}
	return nil
}

func (s *AuthService) applyBruteForceLock(ctx context.Context, userID uuid.UUID, count int) {
	var lockDuration time.Duration
	switch {
	case count >= 15:
		lockDuration = 30 * time.Minute
	case count >= 10:
		lockDuration = 5 * time.Minute
	case count >= 5:
		lockDuration = 1 * time.Minute
	default:
		return
	}

	lockUntil := time.Now().Add(lockDuration)
	if err := s.userRepo.LockUntil(ctx, userID, lockUntil); err != nil {
		s.logger.Error().Err(err).
			Str("user_id", userID.String()).
			Int("failed_count", count).
			Msg("failed to apply brute-force lock")
	} else {
		s.logger.Warn().
			Str("user_id", userID.String()).
			Int("failed_count", count).
			Dur("lock_duration", lockDuration).
			Msg("brute-force lock applied")
	}
}

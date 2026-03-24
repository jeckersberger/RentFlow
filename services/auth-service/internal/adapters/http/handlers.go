package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/pkg/common/middleware"
	"github.com/jeckersberger/rentflow/services/auth-service/internal/application"
	"github.com/jeckersberger/rentflow/services/auth-service/internal/domain"
)

// CurrentVersion is the hardcoded current version of RentFlow
const CurrentVersion = "1.0.0"

// versionCache caches the GitHub release check result for 6 hours
var (
	versionCacheMu    sync.Mutex
	cachedVersionInfo *VersionInfo
	versionCacheTime  time.Time
	versionCacheTTL   = 6 * time.Hour
)

// VersionInfo represents the system version response
type VersionInfo struct {
	CurrentVersion  string  `json:"current_version"`
	LatestVersion   *string `json:"latest_version"`
	UpdateAvailable bool    `json:"update_available"`
	ReleaseURL      string  `json:"release_url,omitempty"`
	ReleaseNotes    string  `json:"release_notes,omitempty"`
	CheckedAt       string  `json:"checked_at"`
}

// gitHubRelease represents the relevant fields from the GitHub releases API
type gitHubRelease struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
	Body    string `json:"body"`
}

// Handlers holds references to all service handlers
type Handlers struct {
	userService   *application.UserService
	tenantService *application.TenantService
	setupService  *application.SetupService
	sessionMgr    *application.SessionManager
	logger        logger.Logger
}

// NewHandlers creates a new handlers instance
func NewHandlers(
	userService *application.UserService,
	tenantService *application.TenantService,
	setupService *application.SetupService,
	log logger.Logger,
) *Handlers {
	return &Handlers{
		userService:   userService,
		tenantService: tenantService,
		setupService:  setupService,
		logger:        log,
	}
}

// Register handles user registration
func (h *Handlers) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	var cmd application.RegisterUserCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON")
		return
	}

	user, err := h.userService.Register(r.Context(), cmd)
	if err != nil {
		switch err {
		case domain.ErrEmailExists:
			writeError(w, http.StatusConflict, "EMAIL_EXISTS", "Email already exists")
		case domain.ErrTenantNotFound:
			writeError(w, http.StatusNotFound, "TENANT_NOT_FOUND", "Tenant not found")
		default:
			h.logger.Error("register error", err)
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		}
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"data":    user,
		"message": "User registered successfully",
	})
}

// Login handles user login
func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	var cmd application.LoginCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON")
		return
	}

	tokens, err := h.userService.Login(r.Context(), cmd)
	if err != nil {
		// Fehlgeschlagenen Login in Redis tracken
		ip := extractIPFromRequest(r)
		if h.sessionMgr != nil {
			count, _ := h.sessionMgr.RecordFailedLogin(r.Context(), ip)
			h.logger.Warn("failed login attempt", "email", cmd.Email, "ip", ip, "failCount", count)
		}

		switch err {
		case domain.ErrInvalidCredentials:
			writeError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Invalid email or password")
		case domain.ErrUserLocked:
			writeError(w, http.StatusForbidden, "USER_LOCKED", "User account is locked")
		default:
			h.logger.Error("login error", err)
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		}
		return
	}

	// Erfolgreicher Login: Brute-Force-Counter zuruecksetzen
	ip := extractIPFromRequest(r)
	if h.sessionMgr != nil {
		_ = h.sessionMgr.ResetFailedLogins(r.Context(), ip)
	}

	// Set refresh token as HTTP-only cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    tokens.RefreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   7 * 24 * 60 * 60, // 7 days
	})

	// Fetch user object for the Scanner App contract
	user, _ := h.userService.GetUserByEmail(r.Context(), cmd.Email)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data": map[string]interface{}{
			"access_token":  tokens.AccessToken,
			"refresh_token": tokens.RefreshToken,
			"expires_in":    tokens.ExpiresIn,
			"token_type":    tokens.TokenType,
			"user":          user,
		},
		"message": "Login successful",
	})
}

// Refresh handles token refresh
func (h *Handlers) Refresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	// Get refresh token from cookie or body
	var refreshToken string
	cookie, err := r.Cookie("refresh_token")
	if err == nil {
		refreshToken = cookie.Value
	} else {
		var req map[string]string
		json.NewDecoder(r.Body).Decode(&req)
		refreshToken = req["refresh_token"]
	}

	if refreshToken == "" {
		writeError(w, http.StatusBadRequest, "MISSING_REFRESH_TOKEN", "Refresh token is required")
		return
	}

	tokens, err := h.userService.RefreshToken(r.Context(), refreshToken)
	if err != nil {
		h.logger.Warn("refresh token error", "error", err.Error())
		writeError(w, http.StatusUnauthorized, "INVALID_REFRESH_TOKEN", "Invalid or expired refresh token")
		return
	}

	// Update refresh token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    tokens.RefreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   7 * 24 * 60 * 60,
	})

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data": map[string]interface{}{
			"access_token":  tokens.AccessToken,
			"refresh_token": tokens.RefreshToken,
			"expires_in":    tokens.ExpiresIn,
			"token_type":    tokens.TokenType,
		},
		"message": "Token refreshed successfully",
	})
}

// Logout handles user logout
func (h *Handlers) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	// Clear refresh token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})

	w.WriteHeader(http.StatusNoContent)
}

// GetMe handles getting current user
func (h *Handlers) GetMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	// Extract user ID from JWT claims
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Not authenticated")
		return
	}

	user, err := h.userService.GetUser(r.Context(), userID)
	if err != nil {
		h.logger.Error("get user error", err)
		writeError(w, http.StatusNotFound, "USER_NOT_FOUND", "User not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data":    user,
		"message": "User retrieved successfully",
	})
}

// ChangePassword handles password change
func (h *Handlers) ChangePassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Not authenticated")
		return
	}

	var cmd application.ChangePasswordCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON")
		return
	}

	cmd.UserID = userID
	if err := h.userService.ChangePassword(r.Context(), cmd); err != nil {
		switch err {
		case domain.ErrInvalidCredentials:
			writeError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Current password is incorrect")
		default:
			h.logger.Error("change password error", err)
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		}
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Password changed successfully",
	})
}

// ListUsers handles listing users (admin only)
func (h *Handlers) ListUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	tenantID := middleware.GetTenantIDFromClaims(r.Context())
	if tenantID == "" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Not authenticated")
		return
	}

	// Check admin role
	if !middleware.HasRole(r.Context(), "admin") {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "Admin role required")
		return
	}

	page := 1
	perPage := 20

	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	if pp := r.URL.Query().Get("per_page"); pp != "" {
		if parsed, err := strconv.Atoi(pp); err == nil && parsed > 0 {
			perPage = parsed
		}
	}

	query := application.ListUsersQuery{
		TenantID: tenantID,
		Page:     page,
		PerPage:  perPage,
	}

	result, err := h.userService.ListUsers(r.Context(), query)
	if err != nil {
		h.logger.Error("list users error", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data":    result,
		"message": "Users retrieved successfully",
	})
}

// GetUser handles getting a specific user
func (h *Handlers) GetUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	userID := extractUserIDFromPath(r.URL.Path)
	if userID == "" {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "User ID is required")
		return
	}

	user, err := h.userService.GetUser(r.Context(), userID)
	if err != nil {
		h.logger.Error("get user error", err)
		writeError(w, http.StatusNotFound, "USER_NOT_FOUND", "User not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data":    user,
		"message": "User retrieved successfully",
	})
}

// UpdateProfile handles profile update
func (h *Handlers) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Not authenticated")
		return
	}

	var cmd application.UpdateProfileCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON")
		return
	}

	cmd.UserID = userID
	if err := h.userService.UpdateProfile(r.Context(), cmd); err != nil {
		h.logger.Error("update profile error", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	user, _ := h.userService.GetUser(r.Context(), userID)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data":    user,
		"message": "Profile updated successfully",
	})
}

// AssignRole handles role assignment (admin only)
func (h *Handlers) AssignRole(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	if !middleware.HasRole(r.Context(), "admin") {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "Admin role required")
		return
	}

	userID := extractUserIDFromPath(r.URL.Path)
	if userID == "" {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "User ID is required")
		return
	}

	var req map[string]string
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON")
		return
	}

	cmd := application.AssignRoleCommand{
		UserID:   userID,
		Role:     req["role"],
		TenantID: middleware.GetTenantIDFromClaims(r.Context()),
	}

	if err := h.userService.AssignRole(r.Context(), cmd); err != nil {
		h.logger.Error("assign role error", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Role assigned successfully",
	})
}

// DeleteUser handles user deletion (admin only)
func (h *Handlers) DeleteUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	if !middleware.HasRole(r.Context(), "admin") {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "Admin role required")
		return
	}

	userID := extractUserIDFromPath(r.URL.Path)
	if userID == "" {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "User ID is required")
		return
	}

	tenantID := middleware.GetTenantIDFromClaims(r.Context())
	if err := h.userService.DeactivateUser(r.Context(), application.DeactivateUserCommand{
		UserID:   userID,
		TenantID: tenantID,
	}); err != nil {
		h.logger.Error("delete user error", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// CreateTenant handles tenant creation
func (h *Handlers) CreateTenant(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	var cmd application.CreateTenantCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON")
		return
	}

	tenant, err := h.tenantService.CreateTenant(r.Context(), cmd)
	if err != nil {
		h.logger.Error("create tenant error", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"data":    tenant,
		"message": "Tenant created successfully",
	})
}

// GetTenant handles getting a tenant
func (h *Handlers) GetTenant(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	tenantID := extractTenantIDFromPath(r.URL.Path)
	if tenantID == "" {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Tenant ID is required")
		return
	}

	tenant, err := h.tenantService.GetTenant(r.Context(), application.GetTenantByIDQuery{
		TenantID: tenantID,
	})
	if err != nil {
		h.logger.Error("get tenant error", err)
		writeError(w, http.StatusNotFound, "TENANT_NOT_FOUND", "Tenant not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data":    tenant,
		"message": "Tenant retrieved successfully",
	})
}

// UpdateTenant handles tenant update
func (h *Handlers) UpdateTenant(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	tenantID := extractTenantIDFromPath(r.URL.Path)
	if tenantID == "" {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Tenant ID is required")
		return
	}

	var cmd application.UpdateTenantCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON")
		return
	}

	cmd.TenantID = tenantID
	tenant, err := h.tenantService.UpdateTenant(r.Context(), cmd)
	if err != nil {
		h.logger.Error("update tenant error", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data":    tenant,
		"message": "Tenant updated successfully",
	})
}

// InviteUser handles employee invitation
func (h *Handlers) InviteUser(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantIDFromClaims(r.Context())
	userID := middleware.GetUserID(r.Context())
	if tenantID == "" || userID == "" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Not authenticated")
		return
	}

	var req struct {
		Email string `json:"email"`
		Role  string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON")
		return
	}

	if req.Email == "" {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Email is required")
		return
	}

	cmd := application.InviteUserCommand{
		TenantID:  tenantID,
		Email:     req.Email,
		Role:      req.Role,
		InvitedBy: userID,
	}

	result, err := h.userService.InviteUser(r.Context(), cmd)
	if err != nil {
		h.logger.Error("failed to invite user", err, "email", req.Email)
		writeError(w, http.StatusBadRequest, "INVITATION_ERROR", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, result)
}

// AcceptInvitation handles accepting an invitation (no auth required)
func (h *Handlers) AcceptInvitation(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")
	if token == "" {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Token is required")
		return
	}

	var req struct {
		Password  string `json:"password"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON")
		return
	}

	cmd := application.AcceptInvitationCommand{
		Token:     token,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}

	user, err := h.userService.AcceptInvitation(r.Context(), cmd)
	if err != nil {
		h.logger.Error("failed to accept invitation", err)
		writeError(w, http.StatusBadRequest, "INVITATION_ERROR", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, user)
}

// ListInvitations lists all invitations for the tenant
func (h *Handlers) ListInvitations(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantIDFromClaims(r.Context())
	if tenantID == "" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Not authenticated")
		return
	}

	invitations, err := h.userService.ListInvitations(r.Context(), tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list invitations")
		return
	}

	writeJSON(w, http.StatusOK, invitations)
}

// ForgotPassword handles forgot password requests
func (h *Handlers) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	var cmd application.ForgotPasswordCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON")
		return
	}

	// Always return success to avoid email enumeration
	_, err := h.userService.ForgotPassword(r.Context(), cmd)
	if err != nil {
		h.logger.Error("forgot password error", err)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": "If an account with that email exists, a password reset link has been generated. Check the server logs.",
	})
}

// ResetPassword handles password reset with token
func (h *Handlers) ResetPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	var cmd application.ResetPasswordCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON")
		return
	}

	if cmd.Token == "" || cmd.NewPassword == "" {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Token and new password are required")
		return
	}

	if err := h.userService.ResetPassword(r.Context(), cmd); err != nil {
		h.logger.Warn("reset password error", "error", err.Error())
		writeError(w, http.StatusBadRequest, "RESET_ERROR", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Password has been reset successfully",
	})
}

// QRLogin handles QR code quick login for the Scanner App
func (h *Handlers) QRLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	var req struct {
		QRToken string `json:"qr_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON")
		return
	}

	if req.QRToken == "" {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "qr_token is required")
		return
	}

	tokens, user, err := h.userService.QRLogin(r.Context(), req.QRToken)
	if err != nil {
		h.logger.Warn("QR login failed", "error", err.Error())
		writeError(w, http.StatusUnauthorized, "INVALID_QR_TOKEN", "Invalid or expired QR token")
		return
	}

	// Set refresh token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    tokens.RefreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   7 * 24 * 60 * 60,
	})

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data": map[string]interface{}{
			"access_token":  tokens.AccessToken,
			"refresh_token": tokens.RefreshToken,
			"expires_in":    tokens.ExpiresIn,
			"token_type":    tokens.TokenType,
			"user":          user,
		},
		"message": "QR login successful",
	})
}

// GenerateQRToken generates a one-time QR login token (authenticated endpoint)
func (h *Handlers) GenerateQRToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	userID := middleware.GetUserID(r.Context())
	tenantID := middleware.GetTenantIDFromClaims(r.Context())
	if userID == "" || tenantID == "" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Not authenticated")
		return
	}

	token, expiresAt, err := h.userService.GenerateQRToken(r.Context(), userID, tenantID)
	if err != nil {
		h.logger.Error("failed to generate QR token", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"qr_token":   token,
		"expires_at": expiresAt.Format(time.RFC3339),
	})
}

// GetSystemVersion returns current and latest version info
func (h *Handlers) GetSystemVersion(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	info := getVersionInfo(h.logger)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data":    info,
		"message": "Version info retrieved successfully",
	})
}

// getVersionInfo returns cached version info or fetches from GitHub
func getVersionInfo(log logger.Logger) *VersionInfo {
	versionCacheMu.Lock()
	defer versionCacheMu.Unlock()

	if cachedVersionInfo != nil && time.Since(versionCacheTime) < versionCacheTTL {
		return cachedVersionInfo
	}

	info := &VersionInfo{
		CurrentVersion:  CurrentVersion,
		UpdateAvailable: false,
		CheckedAt:       time.Now().UTC().Format(time.RFC3339),
	}

	// Try to fetch latest release from GitHub
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get("https://api.github.com/repos/jeckersberger/Lagerverwaltung-und-Rechnungsbearbeitungssoftware/releases/latest")
	if err != nil {
		log.Warn("failed to check GitHub for updates", "error", err.Error())
		cachedVersionInfo = info
		versionCacheTime = time.Now()
		return info
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Warn("GitHub API returned non-200", "status", resp.StatusCode)
		cachedVersionInfo = info
		versionCacheTime = time.Now()
		return info
	}

	var release gitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		log.Warn("failed to decode GitHub release", "error", err.Error())
		cachedVersionInfo = info
		versionCacheTime = time.Now()
		return info
	}

	// Strip "v" prefix from tag if present
	latestVersion := strings.TrimPrefix(release.TagName, "v")
	info.LatestVersion = &latestVersion
	info.ReleaseURL = release.HTMLURL
	info.ReleaseNotes = release.Body

	// Simple version comparison: if they differ, an update is available
	if latestVersion != CurrentVersion {
		info.UpdateAvailable = true
	}

	if info.ReleaseURL == "" && latestVersion != "" {
		info.ReleaseURL = fmt.Sprintf("https://github.com/jeckersberger/rentflow/releases/tag/v%s", latestVersion)
	}

	cachedVersionInfo = info
	versionCacheTime = time.Now()
	return info
}

// Helper functions

func writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, statusCode int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"code":    code,
		"message": message,
	})
}

func extractUserIDFromPath(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) > 4 && parts[4] != "" && parts[4] != "roles" {
		return parts[4]
	}
	return ""
}

func extractTenantIDFromPath(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) > 4 {
		return parts[4]
	}
	return ""
}

// readBody reads the entire request body as a string
func readBody(r *http.Request) (string, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// extractIPFromRequest extrahiert die Client-IP aus dem Request
func extractIPFromRequest(r *http.Request) string {
	if cfip := r.Header.Get("CF-Connecting-IP"); cfip != "" {
		return cfip
	}
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	parts := strings.Split(r.RemoteAddr, ":")
	if len(parts) > 0 {
		return parts[0]
	}
	return r.RemoteAddr
}

// GetSetupStatus returns the current setup status
func (h *Handlers) GetSetupStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	status, err := h.setupService.GetStatus(r.Context())
	if err != nil {
		h.logger.Error("failed to get setup status", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data":    status,
		"message": "Setup status retrieved successfully",
	})
}

// CompleteSetup completes the one-time setup wizard
func (h *Handlers) CompleteSetup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	var req application.SetupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON")
		return
	}

	// Validate required fields
	if req.CompanyName == "" || req.CompanySlug == "" || req.AdminEmail == "" || req.AdminPassword == "" || req.SetupToken == "" {
		writeError(w, http.StatusBadRequest, "MISSING_REQUIRED_FIELDS", "Missing required fields")
		return
	}

	if err := h.setupService.CompleteSetup(r.Context(), req); err != nil {
		switch err {
		case application.ErrSetupAlreadyCompleted:
			writeError(w, http.StatusForbidden, "SETUP_ALREADY_COMPLETED", "Setup is already completed")
		case application.ErrInvalidSetupToken:
			writeError(w, http.StatusUnauthorized, "INVALID_SETUP_TOKEN", "Invalid setup token")
		case domain.ErrWeakPassword:
			writeError(w, http.StatusBadRequest, "WEAK_PASSWORD", "Password does not meet requirements")
		default:
			h.logger.Error("setup error", err)
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		}
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Setup completed successfully",
	})
}

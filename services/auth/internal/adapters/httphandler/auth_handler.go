package httphandler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/pkg/common/errors"
	"github.com/jeckersberger/EquipFlow/pkg/common/middleware"
	"github.com/jeckersberger/EquipFlow/pkg/common/response"
	"github.com/jeckersberger/EquipFlow/services/auth/internal/application"
	"github.com/jeckersberger/EquipFlow/services/auth/internal/domain"
)

// UserResponse is the JSON-safe public representation of a user (no password hash).
type UserResponse struct {
	ID        uuid.UUID  `json:"id"`
	TenantID  uuid.UUID  `json:"tenant_id"`
	Email     string     `json:"email"`
	Username  string     `json:"username,omitempty"`
	FirstName string     `json:"first_name"`
	LastName  string     `json:"last_name"`
	Role      string     `json:"role"`
	Phone     string     `json:"phone,omitempty"`
	AvatarURL string     `json:"avatar_url,omitempty"`
	IsActive  bool       `json:"is_active"`
	LastLogin *time.Time `json:"last_login,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// toUserResponse converts a domain.User to the public UserResponse.
func toUserResponse(u *domain.User) UserResponse {
	return UserResponse{
		ID:        u.ID,
		TenantID:  u.TenantID,
		Email:     u.Email,
		Username:  u.Username,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Role:      u.Role,
		Phone:     u.Phone,
		AvatarURL: u.AvatarURL,
		IsActive:  u.IsActive,
		LastLogin: u.LastLogin,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

// AuthHandler handles authentication HTTP endpoints.
type AuthHandler struct {
	authService *application.AuthService
	userService *application.UserService
	tenantRepo  domain.TenantRepository
	logger      zerolog.Logger
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(
	authService *application.AuthService,
	userService *application.UserService,
	tenantRepo domain.TenantRepository,
	logger zerolog.Logger,
) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		userService: userService,
		tenantRepo:  tenantRepo,
		logger:      logger.With().Str("handler", "auth").Logger(),
	}
}

// loginBody is the JSON request body for the login endpoint.
type loginBody struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	TenantSlug string `json:"tenant_slug,omitempty"`
}

// Login authenticates a user and returns a JWT token pair with user info.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var body loginBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiger Request-Body"))
		return
	}

	if body.Email == "" || body.Password == "" {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Email und Passwort sind erforderlich"))
		return
	}

	tenantID, err := h.resolveTenant(r, body.TenantSlug)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	req := application.LoginRequest{
		Email:    body.Email,
		Password: body.Password,
	}

	tokens, user, err := h.authService.Login(r.Context(), tenantID, req, r.RemoteAddr, r.UserAgent())
	if err != nil {
		h.logger.Warn().Err(err).Str("email", body.Email).Msg("login failed")
		errors.HandleError(w, err)
		return
	}

	response.Success(w, map[string]interface{}{
		"tokens": tokens,
		"user":   toUserResponse(user),
	})
}

// refreshBody is the JSON request body for the refresh endpoint.
type refreshBody struct {
	RefreshToken string `json:"refresh_token"`
}

// Refresh exchanges a refresh token for a new token pair.
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var body refreshBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiger Request-Body"))
		return
	}

	if body.RefreshToken == "" {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "refresh_token ist erforderlich"))
		return
	}

	tokens, err := h.authService.Refresh(r.Context(), application.RefreshRequest{
		RefreshToken: body.RefreshToken,
	})
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, tokens)
}

// logoutBody is the JSON request body for the logout endpoint.
type logoutBody struct {
	RefreshToken string `json:"refresh_token"`
}

// Logout revokes the given refresh token.
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var body logoutBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiger Request-Body"))
		return
	}

	if err := h.authService.Logout(r.Context(), body.RefreshToken); err != nil {
		errors.HandleError(w, err)
		return
	}

	response.NoContent(w)
}

// Me returns the currently authenticated user's profile.
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	user, err := h.userService.GetByID(r.Context(), claims.UserID, claims.TenantID)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, toUserResponse(user))
}

// UpdateProfile updates the authenticated user's profile fields.
func (h *AuthHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	var req application.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiger Request-Body"))
		return
	}

	user, err := h.userService.Update(r.Context(), claims.UserID, claims.TenantID, req)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, toUserResponse(user))
}

// ChangePassword changes the authenticated user's password.
func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	var req application.ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiger Request-Body"))
		return
	}

	if req.OldPassword == "" || req.NewPassword == "" {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Altes und neues Passwort sind erforderlich"))
		return
	}

	if err := h.authService.ChangePassword(r.Context(), claims.UserID, req); err != nil {
		errors.HandleError(w, err)
		return
	}

	response.NoContent(w)
}

// ForgotPassword initiates a password reset (not yet implemented).
func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, _ *http.Request) {
	response.Error(w, http.StatusNotImplemented, "NOT_IMPLEMENTED",
		"Passwort-Zuruecksetzung per E-Mail noch nicht verfuegbar")
}

// ResetPassword completes a password reset (not yet implemented).
func (h *AuthHandler) ResetPassword(w http.ResponseWriter, _ *http.Request) {
	response.Error(w, http.StatusNotImplemented, "NOT_IMPLEMENTED",
		"Passwort-Zuruecksetzung noch nicht verfuegbar")
}

// resolveTenant looks up the tenant by slug. If the slug is empty, falls back
// to the X-Tenant-Slug header. Returns an error if no tenant can be resolved.
func (h *AuthHandler) resolveTenant(r *http.Request, slug string) (uuid.UUID, error) {
	ctx := r.Context()

	if slug == "" {
		slug = r.Header.Get("X-Tenant-Slug")
	}

	if slug == "" {
		return uuid.Nil, errors.Wrap(errors.ErrBadRequest, "tenant_slug ist erforderlich")
	}

	tenant, err := h.tenantRepo.GetBySlug(ctx, slug)
	if err != nil {
		return uuid.Nil, errors.Wrap(errors.ErrNotFound, "Tenant nicht gefunden")
	}

	if !tenant.IsActive {
		return uuid.Nil, errors.Wrap(errors.ErrForbidden, "Tenant ist deaktiviert")
	}

	return tenant.ID, nil
}

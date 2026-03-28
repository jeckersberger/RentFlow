package httphandler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/pkg/common/errors"
	"github.com/jeckersberger/EquipFlow/pkg/common/middleware"
	"github.com/jeckersberger/EquipFlow/pkg/common/pagination"
	"github.com/jeckersberger/EquipFlow/pkg/common/response"
	"github.com/jeckersberger/EquipFlow/services/auth/internal/application"
	"github.com/jeckersberger/EquipFlow/services/auth/internal/domain"
)

// UserHandler handles admin user management HTTP endpoints.
type UserHandler struct {
	userService *application.UserService
	logger      zerolog.Logger
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(userService *application.UserService, logger zerolog.Logger) *UserHandler {
	return &UserHandler{
		userService: userService,
		logger:      logger.With().Str("handler", "user").Logger(),
	}
}

// List returns a paginated list of users for the current tenant.
func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	if err := requireRole(claims, domain.RoleAdmin, domain.RoleManager); err != nil {
		errors.HandleError(w, err)
		return
	}

	p := pagination.Parse(r)

	users, total, err := h.userService.List(r.Context(), claims.TenantID, p.Page, p.PerPage)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	out := make([]UserResponse, 0, len(users))
	for _, u := range users {
		out = append(out, toUserResponse(u))
	}

	meta := response.Meta{
		Page:    p.Page,
		PerPage: p.PerPage,
		Total:   total,
	}

	response.SuccessWithMeta(w, out, meta)
}

// Get returns a single user by ID.
func (h *UserHandler) Get(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	userID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	user, err := h.userService.GetByID(r.Context(), userID, claims.TenantID)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, toUserResponse(user))
}

// Create creates a new user in the current tenant. Requires admin role.
func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	if err := requireRole(claims, domain.RoleAdmin); err != nil {
		errors.HandleError(w, err)
		return
	}

	var req application.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiger Request-Body"))
		return
	}

	if req.Email == "" || req.Password == "" || req.FirstName == "" || req.LastName == "" {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest,
			"Email, Passwort, Vorname und Nachname sind erforderlich"))
		return
	}

	user, err := h.userService.Create(r.Context(), claims.TenantID, req)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Created(w, toUserResponse(user))
}

// Update updates an existing user's profile. Requires admin role.
func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	if err := requireRole(claims, domain.RoleAdmin); err != nil {
		errors.HandleError(w, err)
		return
	}

	userID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	var req application.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiger Request-Body"))
		return
	}

	user, err := h.userService.Update(r.Context(), userID, claims.TenantID, req)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, toUserResponse(user))
}

// UpdateRole changes a user's role. Requires admin.
func (h *UserHandler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	if err := requireRole(claims, domain.RoleAdmin); err != nil {
		errors.HandleError(w, err)
		return
	}

	userID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	var req application.UpdateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiger Request-Body"))
		return
	}

	if req.Role == "" {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Rolle ist erforderlich"))
		return
	}

	if err := h.userService.UpdateRole(r.Context(), userID, claims.TenantID, req); err != nil {
		errors.HandleError(w, err)
		return
	}

	response.NoContent(w)
}

// Delete soft-deletes (deactivates) a user. Requires admin.
func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	if err := requireRole(claims, domain.RoleAdmin); err != nil {
		errors.HandleError(w, err)
		return
	}

	userID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	if err := h.userService.Deactivate(r.Context(), userID, claims.TenantID); err != nil {
		errors.HandleError(w, err)
		return
	}

	response.NoContent(w)
}

// Invite sends an invitation to a new user (not yet implemented).
func (h *UserHandler) Invite(w http.ResponseWriter, _ *http.Request) {
	response.Error(w, http.StatusNotImplemented, "NOT_IMPLEMENTED",
		"Einladungsfunktion noch nicht verfuegbar")
}

// requireRole checks that the caller's role is one of the allowed roles.
func requireRole(claims *middleware.Claims, allowed ...string) error {
	for _, role := range allowed {
		if claims.Role == role {
			return nil
		}
	}
	return domain.ErrInsufficientRole
}

// parseUUID parses a string into a uuid.UUID, returning an AppError on failure.
func parseUUID(raw string) (uuid.UUID, error) {
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, errors.Wrap(errors.ErrBadRequest, "Ungueltige UUID: "+raw)
	}
	return id, nil
}

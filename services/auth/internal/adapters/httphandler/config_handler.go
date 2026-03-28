package httphandler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/pkg/common/errors"
	"github.com/jeckersberger/EquipFlow/pkg/common/middleware"
	"github.com/jeckersberger/EquipFlow/pkg/common/response"
	"github.com/jeckersberger/EquipFlow/services/auth/internal/application"
	"github.com/jeckersberger/EquipFlow/services/auth/internal/domain"
)

// ConfigHandler handles per-tenant configuration HTTP endpoints.
type ConfigHandler struct {
	configService *application.ConfigService
	logger        zerolog.Logger
}

// NewConfigHandler creates a new ConfigHandler.
func NewConfigHandler(configService *application.ConfigService, logger zerolog.Logger) *ConfigHandler {
	return &ConfigHandler{
		configService: configService,
		logger:        logger.With().Str("handler", "config").Logger(),
	}
}

// GetAll returns all configuration entries for the current tenant.
func (h *ConfigHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	entries, err := h.configService.GetAll(r.Context(), claims.TenantID)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, entries)
}

// Get returns a single configuration entry by key.
func (h *ConfigHandler) Get(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	key := chi.URLParam(r, "key")
	if key == "" {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Config-Key ist erforderlich"))
		return
	}

	entry, err := h.configService.Get(r.Context(), claims.TenantID, key)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, entry)
}

// setConfigBody is the JSON request body for setting a config value.
type setConfigBody struct {
	Value json.RawMessage `json:"value"`
}

// Set creates or updates a configuration entry. Requires admin role.
func (h *ConfigHandler) Set(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	if err := requireRole(claims, domain.RoleAdmin); err != nil {
		errors.HandleError(w, err)
		return
	}

	key := chi.URLParam(r, "key")
	if key == "" {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Config-Key ist erforderlich"))
		return
	}

	var body setConfigBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiger Request-Body"))
		return
	}

	if body.Value == nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "value ist erforderlich"))
		return
	}

	if err := h.configService.Set(r.Context(), claims.TenantID, key, body.Value); err != nil {
		errors.HandleError(w, err)
		return
	}

	response.NoContent(w)
}

// Delete removes a configuration entry. Requires admin role.
func (h *ConfigHandler) Delete(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	if err := requireRole(claims, domain.RoleAdmin); err != nil {
		errors.HandleError(w, err)
		return
	}

	key := chi.URLParam(r, "key")
	if key == "" {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Config-Key ist erforderlich"))
		return
	}

	if err := h.configService.Delete(r.Context(), claims.TenantID, key); err != nil {
		errors.HandleError(w, err)
		return
	}

	response.NoContent(w)
}

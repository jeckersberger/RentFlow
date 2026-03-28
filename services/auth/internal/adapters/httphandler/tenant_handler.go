package httphandler

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/pkg/common/errors"
	"github.com/jeckersberger/EquipFlow/pkg/common/middleware"
	"github.com/jeckersberger/EquipFlow/pkg/common/response"
	"github.com/jeckersberger/EquipFlow/services/auth/internal/application"
	"github.com/jeckersberger/EquipFlow/services/auth/internal/domain"
)

// TenantHandler handles tenant management HTTP endpoints.
type TenantHandler struct {
	tenantService *application.TenantService
	logger        zerolog.Logger
}

// NewTenantHandler creates a new TenantHandler.
func NewTenantHandler(tenantService *application.TenantService, logger zerolog.Logger) *TenantHandler {
	return &TenantHandler{
		tenantService: tenantService,
		logger:        logger.With().Str("handler", "tenant").Logger(),
	}
}

// GetCurrent returns the tenant for the authenticated user.
func (h *TenantHandler) GetCurrent(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	tenant, err := h.tenantService.GetCurrent(r.Context(), claims.TenantID)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, tenant)
}

// Update updates tenant information. Requires admin role.
func (h *TenantHandler) Update(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	if err := requireRole(claims, domain.RoleAdmin); err != nil {
		errors.HandleError(w, err)
		return
	}

	var req application.UpdateTenantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiger Request-Body"))
		return
	}

	tenant, err := h.tenantService.Update(r.Context(), claims.TenantID, req)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, tenant)
}

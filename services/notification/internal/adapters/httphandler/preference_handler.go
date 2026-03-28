package httphandler

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/pkg/common/errors"
	"github.com/jeckersberger/EquipFlow/pkg/common/middleware"
	"github.com/jeckersberger/EquipFlow/pkg/common/response"
	"github.com/jeckersberger/EquipFlow/services/notification/internal/application"
)

type PreferenceHandler struct {
	service *application.PreferenceService
	logger  zerolog.Logger
}

func NewPreferenceHandler(service *application.PreferenceService, logger zerolog.Logger) *PreferenceHandler {
	return &PreferenceHandler{
		service: service,
		logger:  logger.With().Str("handler", "preference").Logger(),
	}
}

func (h *PreferenceHandler) List(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	items, err := h.service.List(r.Context(), claims.TenantID, claims.UserID)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, items)
}

func (h *PreferenceHandler) Update(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	var req application.UpdatePreferenceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiger Request-Body"))
		return
	}

	updated, err := h.service.Update(r.Context(), claims.TenantID, claims.UserID, req)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, updated)
}

package httphandler

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/pkg/common/errors"
	"github.com/jeckersberger/EquipFlow/pkg/common/middleware"
	"github.com/jeckersberger/EquipFlow/pkg/common/response"
	"github.com/jeckersberger/EquipFlow/services/inventory/internal/application"
)

// AvailabilityHandler handles equipment availability HTTP endpoints.
type AvailabilityHandler struct {
	availabilityService *application.AvailabilityService
	logger              zerolog.Logger
}

// NewAvailabilityHandler creates a new AvailabilityHandler.
func NewAvailabilityHandler(
	availabilityService *application.AvailabilityService,
	logger zerolog.Logger,
) *AvailabilityHandler {
	return &AvailabilityHandler{
		availabilityService: availabilityService,
		logger:              logger.With().Str("handler", "availability").Logger(),
	}
}

// CheckSingle handles POST /api/v1/equipment/availability-check
func (h *AvailabilityHandler) CheckSingle(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	var req application.AvailabilityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiger Request-Body"))
		return
	}

	result, err := h.availabilityService.CheckSingle(r.Context(), claims.TenantID, req)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, result)
}

// CheckBatch handles POST /api/v1/equipment/batch-availability
func (h *AvailabilityHandler) CheckBatch(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	var req application.BatchAvailabilityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiger Request-Body"))
		return
	}

	results, err := h.availabilityService.CheckBatch(r.Context(), claims.TenantID, req)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, results)
}

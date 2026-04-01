package httphandler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
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

// ListAvailability handles GET /api/v1/equipment/availability
// Query params: from (YYYY-MM-DD, required), to (YYYY-MM-DD, required), category_id (optional UUID)
func (h *AvailabilityHandler) ListAvailability(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	q := r.URL.Query()

	from, err := parseDate(q.Get("from"))
	if err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Parameter 'from' ist erforderlich (Format: YYYY-MM-DD)"))
		return
	}
	to, err := parseDate(q.Get("to"))
	if err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Parameter 'to' ist erforderlich (Format: YYYY-MM-DD)"))
		return
	}

	if to.Before(from) {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "'to' darf nicht vor 'from' liegen"))
		return
	}

	var categoryID *uuid.UUID
	if v := q.Get("category_id"); v != "" {
		id, parseErr := uuid.Parse(v)
		if parseErr != nil {
			errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltige category_id"))
			return
		}
		categoryID = &id
	}

	results, err := h.availabilityService.GetTypeAvailability(r.Context(), claims.TenantID, from, to, categoryID)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, results)
}

// GetItemAvailability handles GET /api/v1/equipment/{id}/availability
// Query params: from (YYYY-MM-DD, required), to (YYYY-MM-DD, required)
func (h *AvailabilityHandler) GetItemAvailability(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	q := r.URL.Query()

	from, err := parseDate(q.Get("from"))
	if err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Parameter 'from' ist erforderlich (Format: YYYY-MM-DD)"))
		return
	}
	to, err := parseDate(q.Get("to"))
	if err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Parameter 'to' ist erforderlich (Format: YYYY-MM-DD)"))
		return
	}

	if to.Before(from) {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "'to' darf nicht vor 'from' liegen"))
		return
	}

	bookings, err := h.availabilityService.GetEquipmentBookings(r.Context(), claims.TenantID, id, from, to)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, bookings)
}

// parseDate parses a YYYY-MM-DD string into a time.Time.
func parseDate(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, fmt.Errorf("date is required")
	}
	return time.Parse("2006-01-02", s)
}

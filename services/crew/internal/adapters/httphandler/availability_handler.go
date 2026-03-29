package httphandler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/pkg/common/errors"
	"github.com/jeckersberger/EquipFlow/pkg/common/middleware"
	"github.com/jeckersberger/EquipFlow/pkg/common/response"
	"github.com/jeckersberger/EquipFlow/services/crew/internal/application"
)

// AvailabilityHandler handles crew availability HTTP endpoints.
type AvailabilityHandler struct {
	availabilitySvc *application.AvailabilityService
	logger          zerolog.Logger
}

// NewAvailabilityHandler creates a new AvailabilityHandler.
func NewAvailabilityHandler(
	availabilitySvc *application.AvailabilityService,
	logger zerolog.Logger,
) *AvailabilityHandler {
	return &AvailabilityHandler{
		availabilitySvc: availabilitySvc,
		logger:          logger.With().Str("handler", "availability").Logger(),
	}
}

// ---------------------------------------------------------------------------
// Endpoints
// ---------------------------------------------------------------------------

// Set handles POST /api/v1/crew/availability — set availability for a date.
func (h *AvailabilityHandler) Set(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	var req application.SetAvailabilityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiger Request-Body"))
		return
	}

	availability, err := h.availabilitySvc.SetAvailability(r.Context(), claims.TenantID, req)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Created(w, availability)
}

// ListAll handles GET /api/v1/crew/availability?from=&to= — availability matrix for all members.
func (h *AvailabilityHandler) ListAll(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	q := r.URL.Query()
	from, to := defaultAvailabilityRange(q.Get("from"), q.Get("to"))

	entries, err := h.availabilitySvc.ListAll(r.Context(), claims.TenantID, from, to)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, entries)
}

// ListForMember handles GET /api/v1/crew/{id}/availability?from=&to= — availability for one member.
func (h *AvailabilityHandler) ListForMember(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	memberID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	q := r.URL.Query()
	from, to := defaultAvailabilityRange(q.Get("from"), q.Get("to"))

	entries, err := h.availabilitySvc.ListByMember(r.Context(), claims.TenantID, memberID, from, to)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, entries)
}

// defaultAvailabilityRange parses from/to query params or defaults to 30-day window.
func defaultAvailabilityRange(fromStr, toStr string) (string, string) {
	now := time.Now()
	from := now.Format("2006-01-02")
	to := now.AddDate(0, 0, 30).Format("2006-01-02")

	if fromStr != "" {
		if _, err := time.Parse("2006-01-02", fromStr); err == nil {
			from = fromStr
		}
	}
	if toStr != "" {
		if _, err := time.Parse("2006-01-02", toStr); err == nil {
			to = toStr
		}
	}

	return from, to
}

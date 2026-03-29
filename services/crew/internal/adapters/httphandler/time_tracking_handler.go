package httphandler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/pkg/common/errors"
	"github.com/jeckersberger/EquipFlow/pkg/common/middleware"
	"github.com/jeckersberger/EquipFlow/pkg/common/response"
	"github.com/jeckersberger/EquipFlow/services/crew/internal/application"
)

// TimeTrackingHandler handles time tracking HTTP endpoints.
type TimeTrackingHandler struct {
	timeTrackingSvc *application.TimeTrackingService
	logger          zerolog.Logger
}

// NewTimeTrackingHandler creates a new TimeTrackingHandler.
func NewTimeTrackingHandler(
	timeTrackingSvc *application.TimeTrackingService,
	logger zerolog.Logger,
) *TimeTrackingHandler {
	return &TimeTrackingHandler{
		timeTrackingSvc: timeTrackingSvc,
		logger:          logger.With().Str("handler", "time_tracking").Logger(),
	}
}

// ---------------------------------------------------------------------------
// Endpoints
// ---------------------------------------------------------------------------

// CheckIn handles POST /api/v1/crew/time-tracking/check-in
func (h *TimeTrackingHandler) CheckIn(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	var req application.CheckInRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiger Request-Body"))
		return
	}

	entry, err := h.timeTrackingSvc.CheckIn(r.Context(), claims.TenantID, req)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Created(w, entry)
}

// CheckOut handles POST /api/v1/crew/time-tracking/check-out
func (h *TimeTrackingHandler) CheckOut(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	var req application.CheckOutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiger Request-Body"))
		return
	}

	entry, err := h.timeTrackingSvc.CheckOut(r.Context(), claims.TenantID, req)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, entry)
}

// GetActive handles GET /api/v1/crew/time-tracking/active/{memberId}
func (h *TimeTrackingHandler) GetActive(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	memberID, err := parseUUID(chi.URLParam(r, "memberId"))
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	entry, err := h.timeTrackingSvc.GetActiveEntry(r.Context(), claims.TenantID, memberID)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	if entry == nil {
		response.Success(w, nil)
		return
	}

	response.Success(w, entry)
}

// ListEntries handles GET /api/v1/crew/time-tracking/entries?member_id=&from=&to=
func (h *TimeTrackingHandler) ListEntries(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	q := r.URL.Query()

	// Parse optional member_id filter.
	var memberID *uuid.UUID
	if v := q.Get("member_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltige member_id"))
			return
		}
		memberID = &id
	}

	// Parse optional project_id filter.
	var projectID *uuid.UUID
	if v := q.Get("project_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltige project_id"))
			return
		}
		projectID = &id
	}

	// If project_id is set, use ListByProject.
	if projectID != nil {
		entries, err := h.timeTrackingSvc.ListByProject(r.Context(), claims.TenantID, *projectID)
		if err != nil {
			errors.HandleError(w, err)
			return
		}
		response.Success(w, entries)
		return
	}

	// Parse date range (default: last 30 days).
	from, to := defaultDateRange(q.Get("from"), q.Get("to"))

	if memberID != nil {
		entries, err := h.timeTrackingSvc.ListByMember(r.Context(), claims.TenantID, *memberID, from, to)
		if err != nil {
			errors.HandleError(w, err)
			return
		}
		response.Success(w, entries)
		return
	}

	// No member_id: return all entries for tenant in range.
	entries, err := h.timeTrackingSvc.ListByMember(r.Context(), claims.TenantID, uuid.Nil, from, to)
	if err != nil {
		errors.HandleError(w, err)
		return
	}
	response.Success(w, entries)
}

// Export handles GET /api/v1/crew/time-tracking/export?from=&to=
func (h *TimeTrackingHandler) Export(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	q := r.URL.Query()
	from, to := defaultDateRange(q.Get("from"), q.Get("to"))

	csvBytes, err := h.timeTrackingSvc.ExportCSV(r.Context(), claims.TenantID, from, to)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=time_entries.csv")
	w.WriteHeader(http.StatusOK)
	w.Write(csvBytes)
}

// defaultDateRange parses from/to query params or defaults to 30-day window.
func defaultDateRange(fromStr, toStr string) (time.Time, time.Time) {
	to := time.Now()
	from := to.AddDate(0, 0, -30)

	if fromStr != "" {
		if t, err := time.Parse("2006-01-02", fromStr); err == nil {
			from = t
		}
	}
	if toStr != "" {
		if t, err := time.Parse("2006-01-02", toStr); err == nil {
			// End of day.
			to = t.Add(24*time.Hour - time.Second)
		}
	}

	return from, to
}

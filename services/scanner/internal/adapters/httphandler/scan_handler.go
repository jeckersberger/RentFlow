package httphandler

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/pkg/common/errors"
	"github.com/jeckersberger/EquipFlow/pkg/common/middleware"
	"github.com/jeckersberger/EquipFlow/pkg/common/pagination"
	"github.com/jeckersberger/EquipFlow/pkg/common/response"
	"github.com/jeckersberger/EquipFlow/services/scanner/internal/application"
	"github.com/jeckersberger/EquipFlow/services/scanner/internal/domain"
)

// ScanHandler handles scan-related HTTP endpoints.
type ScanHandler struct {
	scanService *application.ScanService
	logger      zerolog.Logger
}

// NewScanHandler creates a new ScanHandler.
func NewScanHandler(
	scanService *application.ScanService,
	logger zerolog.Logger,
) *ScanHandler {
	return &ScanHandler{
		scanService: scanService,
		logger:      logger.With().Str("handler", "scan").Logger(),
	}
}

// ---------------------------------------------------------------------------
// Endpoints
// ---------------------------------------------------------------------------

// Scan handles POST /api/v1/scanner/scan — Barcode/RFID lookup.
func (h *ScanHandler) Scan(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	var req application.ScanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiger Request-Body"))
		return
	}

	event, err := h.scanService.Scan(r.Context(), claims.TenantID, claims.UserID, req)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Created(w, event)
}

// Checkout handles POST /api/v1/scanner/checkout — Equipment auf Projekt buchen (bulk).
func (h *ScanHandler) Checkout(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	var req application.CheckoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiger Request-Body"))
		return
	}

	events, err := h.scanService.Checkout(r.Context(), claims.TenantID, claims.UserID, req)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Created(w, events)
}

// Checkin handles POST /api/v1/scanner/checkin — Equipment zurueckgeben (bulk, mit Rating).
func (h *ScanHandler) Checkin(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	var req application.CheckinRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiger Request-Body"))
		return
	}

	events, err := h.scanService.Checkin(r.Context(), claims.TenantID, claims.UserID, req)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Created(w, events)
}

// BulkSync handles POST /api/v1/scanner/bulk — Offline-Queue sync (idempotent).
func (h *ScanHandler) BulkSync(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	var req application.BulkSyncRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiger Request-Body"))
		return
	}

	result, err := h.scanService.BulkSync(r.Context(), claims.TenantID, claims.UserID, req)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, result)
}

// AdhocBooking handles POST /api/v1/scanner/adhoc-booking — Spontanbuchung.
func (h *ScanHandler) AdhocBooking(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	var req application.AdhocBookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiger Request-Body"))
		return
	}

	event, err := h.scanService.AdhocBooking(r.Context(), claims.TenantID, claims.UserID, req)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Created(w, event)
}

// ListEvents handles GET /api/v1/scanner/events — Scan-Historie.
func (h *ScanHandler) ListEvents(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	p := pagination.Parse(r)
	q := r.URL.Query()

	filter := domain.ScanEventFilter{
		Page:    p.Page,
		PerPage: p.PerPage,
	}

	if v := q.Get("equipment_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltige equipment_id"))
			return
		}
		filter.EquipmentID = &id
	}
	if v := q.Get("project_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltige project_id"))
			return
		}
		filter.ProjectID = &id
	}
	if v := q.Get("action"); v != "" {
		filter.Action = &v
	}
	if v := q.Get("device_id"); v != "" {
		filter.DeviceID = &v
	}

	items, total, err := h.scanService.ListEvents(r.Context(), claims.TenantID, filter)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.SuccessWithMeta(w, items, response.Meta{
		Page:    p.Page,
		PerPage: p.PerPage,
		Total:   total,
	})
}

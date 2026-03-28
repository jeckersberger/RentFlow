package httphandler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/pkg/common/errors"
	"github.com/jeckersberger/EquipFlow/pkg/common/middleware"
	"github.com/jeckersberger/EquipFlow/pkg/common/response"
	"github.com/jeckersberger/EquipFlow/services/scanner/internal/application"
)

// DeviceHandler handles scanner device HTTP endpoints.
type DeviceHandler struct {
	deviceService *application.DeviceService
	logger        zerolog.Logger
}

// NewDeviceHandler creates a new DeviceHandler.
func NewDeviceHandler(
	deviceService *application.DeviceService,
	logger zerolog.Logger,
) *DeviceHandler {
	return &DeviceHandler{
		deviceService: deviceService,
		logger:        logger.With().Str("handler", "device").Logger(),
	}
}

// ---------------------------------------------------------------------------
// Endpoints
// ---------------------------------------------------------------------------

// Register handles POST /api/v1/scanner/devices/register — Geraet registrieren.
func (h *DeviceHandler) Register(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	var req application.RegisterDeviceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiger Request-Body"))
		return
	}

	device, err := h.deviceService.Register(r.Context(), claims.TenantID, req)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Created(w, device)
}

// List handles GET /api/v1/scanner/devices — Alle Geraete.
func (h *DeviceHandler) List(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	devices, err := h.deviceService.List(r.Context(), claims.TenantID)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, devices)
}

// CheckRing handles GET /api/v1/scanner/devices/ring — Ring-Abfrage (Polling).
func (h *DeviceHandler) CheckRing(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	deviceID := r.URL.Query().Get("device_id")
	if deviceID == "" {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "device_id ist erforderlich"))
		return
	}

	ringRequested, err := h.deviceService.CheckRing(r.Context(), claims.TenantID, deviceID)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, map[string]bool{"ring_requested": ringRequested})
}

// AckRing handles POST /api/v1/scanner/devices/ring-ack — Ring bestaetigt.
func (h *DeviceHandler) AckRing(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	var req application.RingAckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiger Request-Body"))
		return
	}

	if err := h.deviceService.AckRing(r.Context(), claims.TenantID, req); err != nil {
		errors.HandleError(w, err)
		return
	}

	response.NoContent(w)
}

// TriggerRing handles POST /api/v1/scanner/devices/{id}/ring — Ring ausloesen (Webapp).
func (h *DeviceHandler) TriggerRing(w http.ResponseWriter, r *http.Request) {
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

	if err := h.deviceService.TriggerRing(r.Context(), id, claims.TenantID); err != nil {
		errors.HandleError(w, err)
		return
	}

	response.NoContent(w)
}

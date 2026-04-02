package httphandler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/pkg/common/errors"
	"github.com/jeckersberger/EquipFlow/pkg/common/middleware"
	"github.com/jeckersberger/EquipFlow/pkg/common/response"
	"github.com/jeckersberger/EquipFlow/services/transport/internal/application"
)

// VehicleHandler handles vehicle-related HTTP endpoints.
type VehicleHandler struct {
	vehicleService *application.VehicleService
	logger         zerolog.Logger
}

// NewVehicleHandler creates a new VehicleHandler.
func NewVehicleHandler(
	vehicleService *application.VehicleService,
	logger zerolog.Logger,
) *VehicleHandler {
	return &VehicleHandler{
		vehicleService: vehicleService,
		logger:         logger.With().Str("handler", "vehicle").Logger(),
	}
}

// ---------------------------------------------------------------------------
// Endpoints
// ---------------------------------------------------------------------------

// Create handles POST /api/v1/vehicles — Fahrzeug erstellen.
func (h *VehicleHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	var req application.CreateVehicleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiger Request-Body"))
		return
	}

	vehicle, err := h.vehicleService.Create(r.Context(), claims.TenantID, req)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Created(w, vehicle)
}

// List handles GET /api/v1/vehicles — Alle Fahrzeuge.
func (h *VehicleHandler) List(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	vehicles, err := h.vehicleService.List(r.Context(), claims.TenantID)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, vehicles)
}

// GetByID handles GET /api/v1/vehicles/{id} — Fahrzeug nach ID.
func (h *VehicleHandler) GetByID(w http.ResponseWriter, r *http.Request) {
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

	vehicle, err := h.vehicleService.GetByID(r.Context(), id, claims.TenantID)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, vehicle)
}

// Update handles PUT /api/v1/vehicles/{id} — Fahrzeug aktualisieren.
func (h *VehicleHandler) Update(w http.ResponseWriter, r *http.Request) {
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

	var req application.UpdateVehicleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiger Request-Body"))
		return
	}

	vehicle, err := h.vehicleService.Update(r.Context(), id, claims.TenantID, req)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, vehicle)
}

// Delete handles DELETE /api/v1/vehicles/{id} — Fahrzeug loeschen.
func (h *VehicleHandler) Delete(w http.ResponseWriter, r *http.Request) {
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

	if err := h.vehicleService.Delete(r.Context(), id, claims.TenantID); err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, map[string]string{"message": "Fahrzeug geloescht"})
}

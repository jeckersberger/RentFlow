package httphandler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/pkg/common/errors"
	"github.com/jeckersberger/EquipFlow/pkg/common/middleware"
	"github.com/jeckersberger/EquipFlow/pkg/common/pagination"
	"github.com/jeckersberger/EquipFlow/pkg/common/response"
	"github.com/jeckersberger/EquipFlow/services/inventory/internal/application"
)

// FlightcaseHandler handles flightcase HTTP endpoints.
type FlightcaseHandler struct {
	flightcaseService *application.FlightcaseService
	logger            zerolog.Logger
}

// NewFlightcaseHandler creates a new FlightcaseHandler.
func NewFlightcaseHandler(
	flightcaseService *application.FlightcaseService,
	logger zerolog.Logger,
) *FlightcaseHandler {
	return &FlightcaseHandler{
		flightcaseService: flightcaseService,
		logger:            logger.With().Str("handler", "flightcase").Logger(),
	}
}

// List returns a paginated list of flightcases for the current tenant.
func (h *FlightcaseHandler) List(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	p := pagination.Parse(r)

	flightcases, total, err := h.flightcaseService.List(r.Context(), claims.TenantID, p.Page, p.PerPage)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.SuccessWithMeta(w, flightcases, response.Meta{
		Page:    p.Page,
		PerPage: p.PerPage,
		Total:   total,
	})
}

// Get returns a single flightcase by its ID.
func (h *FlightcaseHandler) Get(w http.ResponseWriter, r *http.Request) {
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

	flightcase, err := h.flightcaseService.GetByID(r.Context(), id, claims.TenantID)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, flightcase)
}

// Create creates a new flightcase.
func (h *FlightcaseHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	var req application.CreateFlightcaseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiger Request-Body"))
		return
	}

	created, err := h.flightcaseService.Create(r.Context(), claims.TenantID, req)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Created(w, created)
}

// Update updates an existing flightcase.
func (h *FlightcaseHandler) Update(w http.ResponseWriter, r *http.Request) {
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

	var req application.UpdateFlightcaseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiger Request-Body"))
		return
	}

	updated, err := h.flightcaseService.Update(r.Context(), id, claims.TenantID, req)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, updated)
}

// Delete removes a flightcase by its ID.
func (h *FlightcaseHandler) Delete(w http.ResponseWriter, r *http.Request) {
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

	if err := h.flightcaseService.Delete(r.Context(), id, claims.TenantID); err != nil {
		errors.HandleError(w, err)
		return
	}

	response.NoContent(w)
}

// AddItem adds an equipment item to a flightcase.
func (h *FlightcaseHandler) AddItem(w http.ResponseWriter, r *http.Request) {
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

	var req application.AddFlightcaseItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiger Request-Body"))
		return
	}

	created, err := h.flightcaseService.AddItem(r.Context(), claims.TenantID, id, req)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Created(w, created)
}

// RemoveItem removes an equipment item from a flightcase.
func (h *FlightcaseHandler) RemoveItem(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	// Validate flightcase ID (ensures it belongs to the route).
	if _, err := parseUUID(chi.URLParam(r, "id")); err != nil {
		errors.HandleError(w, err)
		return
	}

	itemID, err := parseUUID(chi.URLParam(r, "itemId"))
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	if err := h.flightcaseService.RemoveItem(r.Context(), claims.TenantID, itemID); err != nil {
		errors.HandleError(w, err)
		return
	}

	response.NoContent(w)
}

// GetItems returns all items in a flightcase.
func (h *FlightcaseHandler) GetItems(w http.ResponseWriter, r *http.Request) {
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

	items, err := h.flightcaseService.GetItems(r.Context(), claims.TenantID, id)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, items)
}

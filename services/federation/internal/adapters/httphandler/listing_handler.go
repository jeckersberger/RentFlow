package httphandler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/pkg/common/errors"
	"github.com/jeckersberger/EquipFlow/pkg/common/middleware"
	"github.com/jeckersberger/EquipFlow/pkg/common/pagination"
	"github.com/jeckersberger/EquipFlow/pkg/common/response"
	"github.com/jeckersberger/EquipFlow/services/federation/internal/application"
	"github.com/jeckersberger/EquipFlow/services/federation/internal/domain"
)

// ListingHandler handles shared listing HTTP endpoints.
type ListingHandler struct {
	listingService *application.ListingService
	logger         zerolog.Logger
}

// NewListingHandler creates a new ListingHandler.
func NewListingHandler(
	listingService *application.ListingService,
	logger zerolog.Logger,
) *ListingHandler {
	return &ListingHandler{
		listingService: listingService,
		logger:         logger.With().Str("handler", "listing").Logger(),
	}
}

// ---------------------------------------------------------------------------
// Endpoints
// ---------------------------------------------------------------------------

// Create handles POST /api/v1/shared-listings — Listing erstellen.
func (h *ListingHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	var req application.CreateListingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiger Request-Body"))
		return
	}

	listing, err := h.listingService.Create(r.Context(), claims.TenantID, req)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Created(w, listing)
}

// List handles GET /api/v1/shared-listings — Alle Listings auflisten.
func (h *ListingHandler) List(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	p := pagination.Parse(r)
	q := r.URL.Query()

	filter := domain.ListingFilter{
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
	if v := q.Get("is_active"); v != "" {
		active := v == "true"
		filter.IsActive = &active
	}

	items, total, err := h.listingService.List(r.Context(), claims.TenantID, filter)
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

// Get handles GET /api/v1/shared-listings/{id} — Listing abrufen.
func (h *ListingHandler) Get(w http.ResponseWriter, r *http.Request) {
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

	listing, err := h.listingService.GetByID(r.Context(), id, claims.TenantID)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, listing)
}

// Update handles PUT /api/v1/shared-listings/{id} — Listing aktualisieren.
func (h *ListingHandler) Update(w http.ResponseWriter, r *http.Request) {
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

	var req application.UpdateListingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiger Request-Body"))
		return
	}

	listing, err := h.listingService.Update(r.Context(), id, claims.TenantID, req)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, listing)
}

// SearchPartnerEquipment handles POST /api/v1/federation/search — search partner equipment.
func (h *ListingHandler) SearchPartnerEquipment(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	var req application.SearchPartnerEquipmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiger Request-Body"))
		return
	}

	results, err := h.listingService.SearchPartnerEquipment(r.Context(), claims.TenantID, req)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, results)
}

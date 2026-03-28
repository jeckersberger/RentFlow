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
	"github.com/jeckersberger/EquipFlow/services/inventory/internal/application"
	"github.com/jeckersberger/EquipFlow/services/inventory/internal/domain"
)

// EquipmentHandler handles equipment HTTP endpoints.
type EquipmentHandler struct {
	equipmentService *application.EquipmentService
	logger           zerolog.Logger
}

// NewEquipmentHandler creates a new EquipmentHandler.
func NewEquipmentHandler(
	equipmentService *application.EquipmentService,
	logger zerolog.Logger,
) *EquipmentHandler {
	return &EquipmentHandler{
		equipmentService: equipmentService,
		logger:           logger.With().Str("handler", "equipment").Logger(),
	}
}

// ---------------------------------------------------------------------------
// Endpoints
// ---------------------------------------------------------------------------

// List returns a paginated, filtered list of equipment for the current tenant.
func (h *EquipmentHandler) List(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	p := pagination.Parse(r)
	q := r.URL.Query()

	filter := domain.EquipmentFilter{
		Page:    p.Page,
		PerPage: p.PerPage,
	}

	if v := q.Get("category_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltige category_id"))
			return
		}
		filter.CategoryID = &id
	}
	if v := q.Get("status"); v != "" {
		filter.Status = &v
	}
	if v := q.Get("condition"); v != "" {
		filter.Condition = &v
	}
	if v := q.Get("search"); v != "" {
		filter.Search = &v
	}

	items, total, err := h.equipmentService.List(r.Context(), claims.TenantID, filter)
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

// Search performs a full-text search across equipment for the current tenant.
func (h *EquipmentHandler) Search(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	query := r.URL.Query().Get("q")
	if query == "" {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Suchbegriff (q) ist erforderlich"))
		return
	}

	p := pagination.Parse(r)

	items, total, err := h.equipmentService.Search(r.Context(), claims.TenantID, query, p.Page, p.PerPage)
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

// LookupBarcode looks up a single equipment item by its barcode.
func (h *EquipmentHandler) LookupBarcode(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	barcode := chi.URLParam(r, "barcode")
	if barcode == "" {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Barcode ist erforderlich"))
		return
	}

	item, err := h.equipmentService.GetByBarcode(r.Context(), claims.TenantID, barcode)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, item)
}

// LookupRFID looks up a single equipment item by its RFID tag.
func (h *EquipmentHandler) LookupRFID(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	tag := chi.URLParam(r, "tag")
	if tag == "" {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "RFID-Tag ist erforderlich"))
		return
	}

	item, err := h.equipmentService.GetByRFID(r.Context(), claims.TenantID, tag)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, item)
}

// Get returns a single equipment item by its ID.
func (h *EquipmentHandler) Get(w http.ResponseWriter, r *http.Request) {
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

	item, err := h.equipmentService.GetByID(r.Context(), id, claims.TenantID)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, item)
}

// Create creates a new equipment item.
func (h *EquipmentHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	var req application.CreateEquipmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiger Request-Body"))
		return
	}

	created, err := h.equipmentService.Create(r.Context(), claims.TenantID, claims.UserID, req)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Created(w, created)
}

// Update updates an existing equipment item.
func (h *EquipmentHandler) Update(w http.ResponseWriter, r *http.Request) {
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

	var req application.UpdateEquipmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiger Request-Body"))
		return
	}

	updated, err := h.equipmentService.Update(r.Context(), id, claims.TenantID, claims.UserID, req)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, updated)
}

// Delete soft-deletes (deactivates) an equipment item.
func (h *EquipmentHandler) Delete(w http.ResponseWriter, r *http.Request) {
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

	if err := h.equipmentService.Deactivate(r.Context(), id, claims.TenantID, claims.UserID); err != nil {
		errors.HandleError(w, err)
		return
	}

	response.NoContent(w)
}

// UpdateStatus changes the status of an equipment item.
func (h *EquipmentHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
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

	var req application.UpdateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiger Request-Body"))
		return
	}

	if req.Status == "" {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Status ist erforderlich"))
		return
	}

	if err := h.equipmentService.UpdateStatus(r.Context(), id, claims.TenantID, claims.UserID, req); err != nil {
		errors.HandleError(w, err)
		return
	}

	response.NoContent(w)
}

// UpdateCondition changes the condition of an equipment item.
func (h *EquipmentHandler) UpdateCondition(w http.ResponseWriter, r *http.Request) {
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

	var req application.UpdateConditionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiger Request-Body"))
		return
	}

	if req.Condition == "" {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Zustand ist erforderlich"))
		return
	}

	if err := h.equipmentService.UpdateCondition(r.Context(), id, claims.TenantID, claims.UserID, req); err != nil {
		errors.HandleError(w, err)
		return
	}

	response.NoContent(w)
}

// AssignRFID assigns an RFID tag to an equipment item.
func (h *EquipmentHandler) AssignRFID(w http.ResponseWriter, r *http.Request) {
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

	var req application.AssignRFIDRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiger Request-Body"))
		return
	}

	if req.RFIDTag == "" {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "RFID-Tag ist erforderlich"))
		return
	}

	if err := h.equipmentService.AssignRFID(r.Context(), id, claims.TenantID, claims.UserID, req); err != nil {
		errors.HandleError(w, err)
		return
	}

	response.NoContent(w)
}

// GetHistory returns the audit history for an equipment item.
func (h *EquipmentHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
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

	p := pagination.Parse(r)

	history, total, err := h.equipmentService.GetHistory(r.Context(), id, p.Page, p.PerPage)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.SuccessWithMeta(w, history, response.Meta{
		Page:    p.Page,
		PerPage: p.PerPage,
		Total:   total,
	})
}

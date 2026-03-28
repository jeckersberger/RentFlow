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
	"github.com/jeckersberger/EquipFlow/services/transport/internal/application"
	"github.com/jeckersberger/EquipFlow/services/transport/internal/domain"
)

// OrderHandler handles transport order HTTP endpoints.
type OrderHandler struct {
	orderService *application.OrderService
	logger       zerolog.Logger
}

// NewOrderHandler creates a new OrderHandler.
func NewOrderHandler(
	orderService *application.OrderService,
	logger zerolog.Logger,
) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
		logger:       logger.With().Str("handler", "order").Logger(),
	}
}

// ---------------------------------------------------------------------------
// Endpoints
// ---------------------------------------------------------------------------

// Create handles POST /api/v1/transport-orders — Transportauftrag erstellen.
func (h *OrderHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	var req application.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiger Request-Body"))
		return
	}

	order, err := h.orderService.Create(r.Context(), claims.TenantID, claims.UserID, req)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Created(w, order)
}

// List handles GET /api/v1/transport-orders — Transportauftraege auflisten.
func (h *OrderHandler) List(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	p := pagination.Parse(r)
	q := r.URL.Query()

	filter := domain.TransportOrderFilter{
		Page:    p.Page,
		PerPage: p.PerPage,
	}

	if v := q.Get("project_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltige project_id"))
			return
		}
		filter.ProjectID = &id
	}
	if v := q.Get("status"); v != "" {
		filter.Status = &v
	}

	items, total, err := h.orderService.List(r.Context(), claims.TenantID, filter)
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

// GetByID handles GET /api/v1/transport-orders/{id} — Transportauftrag nach ID.
func (h *OrderHandler) GetByID(w http.ResponseWriter, r *http.Request) {
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

	order, err := h.orderService.GetByID(r.Context(), id, claims.TenantID)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, order)
}

// Update handles PUT /api/v1/transport-orders/{id} — Transportauftrag aktualisieren.
func (h *OrderHandler) Update(w http.ResponseWriter, r *http.Request) {
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

	var req application.UpdateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiger Request-Body"))
		return
	}

	order, err := h.orderService.Update(r.Context(), id, claims.TenantID, req)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, order)
}

// Complete handles PATCH /api/v1/transport-orders/{id}/complete — Transportauftrag abschliessen.
func (h *OrderHandler) Complete(w http.ResponseWriter, r *http.Request) {
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

	order, err := h.orderService.Complete(r.Context(), id, claims.TenantID)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, order)
}

// AddItem handles POST /api/v1/transport-orders/{id}/items — Item hinzufuegen.
func (h *OrderHandler) AddItem(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	orderID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	var req application.AddItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiger Request-Body"))
		return
	}

	item, err := h.orderService.AddItem(r.Context(), orderID, claims.TenantID, req)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Created(w, item)
}

// ListItems handles GET /api/v1/transport-orders/{id}/items — Items auflisten.
func (h *OrderHandler) ListItems(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	orderID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	items, err := h.orderService.ListItems(r.Context(), orderID, claims.TenantID)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, items)
}

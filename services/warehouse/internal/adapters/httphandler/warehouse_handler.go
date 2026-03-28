package httphandler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/pkg/common/errors"
	"github.com/jeckersberger/EquipFlow/pkg/common/middleware"
	"github.com/jeckersberger/EquipFlow/pkg/common/response"
	"github.com/jeckersberger/EquipFlow/services/warehouse/internal/application"
)

// WarehouseHandler handles warehouse HTTP endpoints.
type WarehouseHandler struct {
	warehouseService *application.WarehouseService
	logger           zerolog.Logger
}

// NewWarehouseHandler creates a new WarehouseHandler.
func NewWarehouseHandler(
	warehouseService *application.WarehouseService,
	logger zerolog.Logger,
) *WarehouseHandler {
	return &WarehouseHandler{
		warehouseService: warehouseService,
		logger:           logger.With().Str("handler", "warehouse").Logger(),
	}
}

// List returns all warehouses for the current tenant.
func (h *WarehouseHandler) List(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	warehouses, err := h.warehouseService.List(r.Context(), claims.TenantID)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, warehouses)
}

// Get returns a single warehouse by its ID.
func (h *WarehouseHandler) Get(w http.ResponseWriter, r *http.Request) {
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

	warehouse, err := h.warehouseService.GetByID(r.Context(), id, claims.TenantID)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, warehouse)
}

// Create creates a new warehouse.
func (h *WarehouseHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	var req application.CreateWarehouseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiger Request-Body"))
		return
	}

	created, err := h.warehouseService.Create(r.Context(), claims.TenantID, req)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Created(w, created)
}

package httphandler

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/pkg/common/errors"
	"github.com/jeckersberger/EquipFlow/pkg/common/middleware"
	"github.com/jeckersberger/EquipFlow/pkg/common/response"
	"github.com/jeckersberger/EquipFlow/services/warehouse/internal/application"
)

// RackHandler handles rack HTTP endpoints.
type RackHandler struct {
	rackService *application.RackService
	logger      zerolog.Logger
}

// NewRackHandler creates a new RackHandler.
func NewRackHandler(
	rackService *application.RackService,
	logger zerolog.Logger,
) *RackHandler {
	return &RackHandler{
		rackService: rackService,
		logger:      logger.With().Str("handler", "rack").Logger(),
	}
}

// List returns all racks for the current tenant.
func (h *RackHandler) List(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	racks, err := h.rackService.List(r.Context(), claims.TenantID)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, racks)
}

// Create creates a new rack.
func (h *RackHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	var req application.CreateRackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiger Request-Body"))
		return
	}

	created, err := h.rackService.Create(r.Context(), claims.TenantID, req)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Created(w, created)
}

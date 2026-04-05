package httphandler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/pkg/common/errors"
	"github.com/jeckersberger/EquipFlow/pkg/common/middleware"
	"github.com/jeckersberger/EquipFlow/pkg/common/response"
	"github.com/jeckersberger/EquipFlow/services/crew/internal/application"
)

// AvailabilityBlockHandler handles crew availability block HTTP endpoints.
type AvailabilityBlockHandler struct {
	blockSvc *application.AvailabilityBlockService
	logger   zerolog.Logger
}

// NewAvailabilityBlockHandler creates a new AvailabilityBlockHandler.
func NewAvailabilityBlockHandler(
	blockSvc *application.AvailabilityBlockService,
	logger zerolog.Logger,
) *AvailabilityBlockHandler {
	return &AvailabilityBlockHandler{
		blockSvc: blockSvc,
		logger:   logger.With().Str("handler", "availability_block").Logger(),
	}
}

// ---------------------------------------------------------------------------
// Endpoints
// ---------------------------------------------------------------------------

// Create handles POST /api/v1/crew/{id}/availability-blocks — create a new block.
func (h *AvailabilityBlockHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	memberID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	var req application.CreateAvailabilityBlockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiger Request-Body"))
		return
	}

	block, err := h.blockSvc.Create(r.Context(), claims.TenantID, memberID, req)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Created(w, block)
}

// ListForMember handles GET /api/v1/crew/{id}/availability-blocks?from=&to=
func (h *AvailabilityBlockHandler) ListForMember(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	memberID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	q := r.URL.Query()
	from, to := defaultAvailabilityRange(q.Get("from"), q.Get("to"))

	blocks, err := h.blockSvc.ListByMember(r.Context(), claims.TenantID, memberID, from, to)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, blocks)
}

// Delete handles DELETE /api/v1/crew/{id}/availability-blocks/{blockId}
func (h *AvailabilityBlockHandler) Delete(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	// Parse member ID (for route consistency, not used in delete logic).
	if _, err := parseUUID(chi.URLParam(r, "id")); err != nil {
		errors.HandleError(w, err)
		return
	}

	blockID, err := parseUUID(chi.URLParam(r, "blockId"))
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	if err := h.blockSvc.Delete(r.Context(), claims.TenantID, blockID); err != nil {
		errors.HandleError(w, err)
		return
	}

	response.NoContent(w)
}

// ListAll handles GET /api/v1/crew/availability-blocks?from=&to= — all crew blocks.
func (h *AvailabilityBlockHandler) ListAll(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	q := r.URL.Query()
	from, to := defaultAvailabilityRange(q.Get("from"), q.Get("to"))

	blocks, err := h.blockSvc.ListAll(r.Context(), claims.TenantID, from, to)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, blocks)
}

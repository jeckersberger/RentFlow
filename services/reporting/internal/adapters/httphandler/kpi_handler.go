package httphandler

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/pkg/common/errors"
	"github.com/jeckersberger/EquipFlow/pkg/common/middleware"
	"github.com/jeckersberger/EquipFlow/pkg/common/response"
	"github.com/jeckersberger/EquipFlow/services/reporting/internal/application"
)

// KPIHandler handles KPI snapshot HTTP endpoints.
type KPIHandler struct {
	service *application.KPIService
	logger  zerolog.Logger
}

// NewKPIHandler creates a new KPIHandler.
func NewKPIHandler(service *application.KPIService, logger zerolog.Logger) *KPIHandler {
	return &KPIHandler{
		service: service,
		logger:  logger.With().Str("handler", "kpi").Logger(),
	}
}

// GetLatest handles GET /api/v1/reporting/kpis — latest KPI snapshot.
func (h *KPIHandler) GetLatest(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	snap, err := h.service.GetLatest(r.Context(), claims.TenantID)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, snap)
}

// GetHistory handles GET /api/v1/reporting/kpis/history?from=&to= — KPI history for charts.
func (h *KPIHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")

	items, err := h.service.GetHistory(r.Context(), claims.TenantID, from, to)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, items)
}

// CreateSnapshot handles POST /api/v1/reporting/kpis/snapshot — create today's snapshot.
func (h *KPIHandler) CreateSnapshot(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	var req application.CreateKPISnapshotRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiger Request-Body"))
		return
	}

	snap, err := h.service.CreateSnapshot(r.Context(), claims.TenantID, req)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Created(w, snap)
}

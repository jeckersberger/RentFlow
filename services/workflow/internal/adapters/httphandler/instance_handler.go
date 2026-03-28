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
	"github.com/jeckersberger/EquipFlow/services/workflow/internal/application"
	"github.com/jeckersberger/EquipFlow/services/workflow/internal/domain"
)

// InstanceHandler handles workflow instance HTTP endpoints.
type InstanceHandler struct {
	instService *application.InstanceService
	logger      zerolog.Logger
}

// NewInstanceHandler creates a new InstanceHandler.
func NewInstanceHandler(
	instService *application.InstanceService,
	logger zerolog.Logger,
) *InstanceHandler {
	return &InstanceHandler{
		instService: instService,
		logger:      logger.With().Str("handler", "instance").Logger(),
	}
}

// ---------------------------------------------------------------------------
// Endpoints
// ---------------------------------------------------------------------------

// Create handles POST /api/v1/workflow-instances — Neue Instanz starten.
func (h *InstanceHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	var req application.CreateInstanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiger Request-Body"))
		return
	}

	inst, err := h.instService.Create(r.Context(), claims.TenantID, claims.UserID, req)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Created(w, inst)
}

// List handles GET /api/v1/workflow-instances — Alle Instanzen auflisten.
func (h *InstanceHandler) List(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	p := pagination.Parse(r)
	q := r.URL.Query()

	filter := domain.InstanceFilter{
		Page:    p.Page,
		PerPage: p.PerPage,
	}

	if v := q.Get("reference_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltige reference_id"))
			return
		}
		filter.ReferenceID = &id
	}
	if v := q.Get("status"); v != "" {
		filter.Status = &v
	}

	items, total, err := h.instService.List(r.Context(), claims.TenantID, filter)
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

// GetByID handles GET /api/v1/workflow-instances/{id} — Instanz abrufen.
func (h *InstanceHandler) GetByID(w http.ResponseWriter, r *http.Request) {
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

	inst, err := h.instService.GetByID(r.Context(), id, claims.TenantID)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, inst)
}

// PerformAction handles POST /api/v1/workflow-instances/{id}/action — Aktion ausfuehren.
func (h *InstanceHandler) PerformAction(w http.ResponseWriter, r *http.Request) {
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

	var req application.PerformActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiger Request-Body"))
		return
	}

	inst, err := h.instService.PerformAction(r.Context(), id, claims.TenantID, claims.UserID, req)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, inst)
}

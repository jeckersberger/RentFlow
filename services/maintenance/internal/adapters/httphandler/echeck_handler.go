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
	"github.com/jeckersberger/EquipFlow/services/maintenance/internal/application"
	"github.com/jeckersberger/EquipFlow/services/maintenance/internal/domain"
)

// ECheckHandler handles E-Check related HTTP endpoints.
type ECheckHandler struct {
	service *application.ECheckService
	logger  zerolog.Logger
}

// NewECheckHandler creates a new ECheckHandler.
func NewECheckHandler(service *application.ECheckService, logger zerolog.Logger) *ECheckHandler {
	return &ECheckHandler{
		service: service,
		logger:  logger.With().Str("handler", "echeck").Logger(),
	}
}

// Create handles POST /api/v1/maintenance/echeck — E-Check-Ergebnis erfassen.
func (h *ECheckHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	var req application.CreateECheckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiger Request-Body"))
		return
	}

	created, err := h.service.Create(r.Context(), claims.TenantID, req)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Created(w, created)
}

// List handles GET /api/v1/maintenance/echeck — Alle E-Check-Eintraege auflisten.
func (h *ECheckHandler) List(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	p := pagination.Parse(r)

	filter := domain.ECheckFilter{
		Page:    p.Page,
		PerPage: p.PerPage,
	}

	eqIDStr := r.URL.Query().Get("equipment_id")
	if eqIDStr != "" {
		parsed, err := uuid.Parse(eqIDStr)
		if err != nil {
			errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltige equipment_id"))
			return
		}
		filter.EquipmentID = &parsed
	}

	items, total, err := h.service.List(r.Context(), claims.TenantID, filter)
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

// ListOverdue handles GET /api/v1/maintenance/echeck/overdue — Ueberfaellige E-Checks.
func (h *ECheckHandler) ListOverdue(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	items, err := h.service.ListOverdue(r.Context(), claims.TenantID)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, items)
}

// ListByEquipment handles GET /api/v1/maintenance/echeck/{equipmentId} — E-Check-Historie pro Geraet.
func (h *ECheckHandler) ListByEquipment(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	equipmentID, err := parseUUID(chi.URLParam(r, "equipmentId"))
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	items, err := h.service.ListByEquipment(r.Context(), equipmentID, claims.TenantID)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, items)
}

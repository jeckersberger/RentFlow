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
	"github.com/jeckersberger/EquipFlow/services/expense/internal/application"
	"github.com/jeckersberger/EquipFlow/services/expense/internal/domain"
)

// RecurringHandler handles recurring expense HTTP endpoints.
type RecurringHandler struct {
	service *application.RecurringService
	logger  zerolog.Logger
}

// NewRecurringHandler creates a new RecurringHandler.
func NewRecurringHandler(service *application.RecurringService, logger zerolog.Logger) *RecurringHandler {
	return &RecurringHandler{
		service: service,
		logger:  logger.With().Str("handler", "recurring").Logger(),
	}
}

// List handles GET /api/v1/expenses/recurring — list recurring expenses.
func (h *RecurringHandler) List(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	p := pagination.Parse(r)
	filter := domain.RecurringExpenseFilter{
		Page:    p.Page,
		PerPage: p.PerPage,
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

// Create handles POST /api/v1/expenses/recurring — create recurring expense.
func (h *RecurringHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	var req application.CreateRecurringRequest
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

// Update handles PUT /api/v1/expenses/recurring/{id} — update recurring expense.
func (h *RecurringHandler) Update(w http.ResponseWriter, r *http.Request) {
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

	var req application.UpdateRecurringRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiger Request-Body"))
		return
	}

	updated, err := h.service.Update(r.Context(), id, claims.TenantID, req)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, updated)
}

// Delete handles DELETE /api/v1/expenses/recurring/{id} — delete recurring expense.
func (h *RecurringHandler) Delete(w http.ResponseWriter, r *http.Request) {
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

	if err := h.service.Delete(r.Context(), id, claims.TenantID); err != nil {
		errors.HandleError(w, err)
		return
	}

	response.NoContent(w)
}

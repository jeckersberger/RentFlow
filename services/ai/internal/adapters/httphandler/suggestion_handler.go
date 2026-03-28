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
	"github.com/jeckersberger/EquipFlow/services/ai/internal/application"
	"github.com/jeckersberger/EquipFlow/services/ai/internal/domain"
)

type SuggestionHandler struct {
	service *application.SuggestionService
	logger  zerolog.Logger
}

func NewSuggestionHandler(service *application.SuggestionService, logger zerolog.Logger) *SuggestionHandler {
	return &SuggestionHandler{
		service: service,
		logger:  logger.With().Str("handler", "suggestion").Logger(),
	}
}

func (h *SuggestionHandler) List(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	p := pagination.Parse(r)
	sugType := r.URL.Query().Get("type")
	sugStatus := r.URL.Query().Get("status")

	filter := domain.SuggestionFilter{
		Page:    p.Page,
		PerPage: p.PerPage,
		Type:    sugType,
		Status:  sugStatus,
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

func (h *SuggestionHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	var req application.CreateSuggestionRequest
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

func (h *SuggestionHandler) Accept(w http.ResponseWriter, r *http.Request) {
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

	updated, err := h.service.Accept(r.Context(), id, claims.TenantID, claims.UserID)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, updated)
}

func (h *SuggestionHandler) Dismiss(w http.ResponseWriter, r *http.Request) {
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

	updated, err := h.service.Dismiss(r.Context(), id, claims.TenantID)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, updated)
}

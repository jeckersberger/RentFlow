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
	"github.com/jeckersberger/EquipFlow/services/audit/internal/application"
	"github.com/jeckersberger/EquipFlow/services/audit/internal/domain"
)

type AuditLogHandler struct {
	service *application.AuditLogService
	logger  zerolog.Logger
}

func NewAuditLogHandler(service *application.AuditLogService, logger zerolog.Logger) *AuditLogHandler {
	return &AuditLogHandler{
		service: service,
		logger:  logger.With().Str("handler", "audit_log").Logger(),
	}
}

func (h *AuditLogHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	var req application.CreateAuditLogRequest
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

func (h *AuditLogHandler) List(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	p := pagination.Parse(r)
	entityType := r.URL.Query().Get("entity_type")

	filter := domain.AuditLogFilter{
		Page:       p.Page,
		PerPage:    p.PerPage,
		EntityType: entityType,
	}

	if raw := r.URL.Query().Get("entity_id"); raw != "" {
		id, err := parseUUID(raw)
		if err != nil {
			errors.HandleError(w, err)
			return
		}
		filter.EntityID = &id
	}

	if raw := r.URL.Query().Get("user_id"); raw != "" {
		id, err := parseUUID(raw)
		if err != nil {
			errors.HandleError(w, err)
			return
		}
		filter.UserID = &id
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

func (h *AuditLogHandler) Get(w http.ResponseWriter, r *http.Request) {
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

	entry, err := h.service.GetByID(r.Context(), id, claims.TenantID)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, entry)
}

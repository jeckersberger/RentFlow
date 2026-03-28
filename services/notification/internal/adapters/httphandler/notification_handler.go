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
	"github.com/jeckersberger/EquipFlow/services/notification/internal/application"
	"github.com/jeckersberger/EquipFlow/services/notification/internal/domain"
)

type NotificationHandler struct {
	service *application.NotificationService
	logger  zerolog.Logger
}

func NewNotificationHandler(service *application.NotificationService, logger zerolog.Logger) *NotificationHandler {
	return &NotificationHandler{
		service: service,
		logger:  logger.With().Str("handler", "notification").Logger(),
	}
}

func (h *NotificationHandler) List(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	p := pagination.Parse(r)

	filter := domain.NotificationFilter{
		Page:    p.Page,
		PerPage: p.PerPage,
	}

	isReadParam := r.URL.Query().Get("is_read")
	if isReadParam == "true" {
		v := true
		filter.IsRead = &v
	} else if isReadParam == "false" {
		v := false
		filter.IsRead = &v
	}

	items, total, err := h.service.List(r.Context(), claims.TenantID, claims.UserID, filter)
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

func (h *NotificationHandler) UnreadCount(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	count, err := h.service.UnreadCount(r.Context(), claims.TenantID, claims.UserID)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, domain.UnreadCount{Count: count})
}

func (h *NotificationHandler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
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

	if err := h.service.MarkAsRead(r.Context(), id, claims.TenantID); err != nil {
		errors.HandleError(w, err)
		return
	}

	response.NoContent(w)
}

func (h *NotificationHandler) MarkAllRead(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	count, err := h.service.MarkAllRead(r.Context(), claims.TenantID, claims.UserID)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, map[string]int64{"updated": count})
}

func (h *NotificationHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	var req application.CreateNotificationRequest
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

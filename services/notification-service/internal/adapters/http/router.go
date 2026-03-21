package http

import (
	"encoding/json"
	"net/http"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/notification-service/internal/application"
	"github.com/jeckersberger/rentflow/services/notification-service/internal/domain"
)

func NewRouter(svc *application.NotificationService, log logger.Logger) *http.ServeMux {
	router := http.NewServeMux()
	h := &Handler{svc: svc, logger: log}

	router.HandleFunc("GET /health", healthHandler)
	router.HandleFunc("GET /ready", readyHandler)

	router.HandleFunc("GET /api/v1/notifications", h.ListNotifications)
	router.HandleFunc("GET /api/v1/notifications/unread-count", h.GetUnreadCount)
	router.HandleFunc("PUT /api/v1/notifications/{id}/read", h.MarkAsRead)
	router.HandleFunc("POST /api/v1/notifications/mark-all-read", h.MarkAllAsRead)
	router.HandleFunc("POST /api/v1/notifications/send", h.SendNotification)
	router.HandleFunc("GET /api/v1/notifications/preferences", h.ListPreferences)
	router.HandleFunc("PUT /api/v1/notifications/preferences", h.UpdatePreferences)

	return router
}

type Handler struct {
	svc    *application.NotificationService
	logger logger.Logger
}

func (h *Handler) ListNotifications(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	userID := r.Header.Get("X-User-ID")
	if tenantID == "" || userID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant and user IDs required")
		return
	}

	notifs, err := h.svc.ListNotifications(r.Context(), tenantID, userID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{"data": notifs})
}

func (h *Handler) GetUnreadCount(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	userID := r.Header.Get("X-User-ID")
	if tenantID == "" || userID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant and user IDs required")
		return
	}

	count, err := h.svc.GetUnreadCount(r.Context(), tenantID, userID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]int{"unread_count": count})
}

func (h *Handler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	if err := h.svc.MarkAsRead(r.Context(), tenantID, id); err != nil {
		h.handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) MarkAllAsRead(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	userID := r.Header.Get("X-User-ID")
	if tenantID == "" || userID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant and user IDs required")
		return
	}

	if err := h.svc.MarkAllAsRead(r.Context(), tenantID, userID); err != nil {
		h.handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) SendNotification(w http.ResponseWriter, r *http.Request) {
	var cmd application.SendNotificationCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cmd.TenantID = r.Header.Get("X-Tenant-ID")
	if cmd.TenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	notif, err := h.svc.SendNotification(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusCreated, notif)
}

func (h *Handler) ListPreferences(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"data":[]}`))
}

func (h *Handler) UpdatePreferences(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

func (h *Handler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) respondError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func (h *Handler) handleError(w http.ResponseWriter, err error) {
	if err == domain.ErrNotificationNotFound {
		h.respondError(w, http.StatusNotFound, err.Error())
		return
	}
	h.logger.Error("Error", err)
	h.respondError(w, http.StatusInternalServerError, "internal server error")
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy","service":"notification-service"}`))
}

func readyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ready","service":"notification-service"}`))
}

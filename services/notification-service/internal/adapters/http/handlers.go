package http

import (
	"encoding/json"
	nethttp "net/http"

	"github.com/google/uuid"
	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/notification-service/internal/application"
	"github.com/jeckersberger/rentflow/services/notification-service/internal/domain"
)

type NotificationHandler struct {
	service          *application.NotificationService
	channelService   *application.ChannelService
	preferenceService *application.PreferenceService
	digestService    *application.DigestService
	log              logger.Logger
}

func NewNotificationHandler(
	service *application.NotificationService,
	channelService *application.ChannelService,
	preferenceService *application.PreferenceService,
	digestService *application.DigestService,
	log logger.Logger,
) *NotificationHandler {
	return &NotificationHandler{
		service:          service,
		channelService:   channelService,
		preferenceService: preferenceService,
		digestService:    digestService,
		log:              log,
	}
}

func (h *NotificationHandler) SendNotification(w nethttp.ResponseWriter, r *nethttp.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, nethttp.StatusBadRequest, "X-Tenant-ID header required")
		return
	}

	var req application.SendNotificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, nethttp.StatusBadRequest, "Invalid request body")
		return
	}

	tenantUUID, _ := uuid.Parse(tenantID)
	cmd := domain.SendNotificationCmd{
		TenantID:  tenantUUID,
		UserID:    req.UserID,
		EventType: req.EventType,
		Title:     req.Title,
		Body:      req.Body,
		Data:      req.Data,
	}

	notification, err := h.service.SendNotification(r.Context(), cmd)
	if err != nil {
		h.respondError(w, nethttp.StatusInternalServerError, "Failed to send notification")
		return
	}

	h.respondJSON(w, nethttp.StatusCreated, notification)
}

func (h *NotificationHandler) BroadcastNotification(w nethttp.ResponseWriter, r *nethttp.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, nethttp.StatusBadRequest, "X-Tenant-ID header required")
		return
	}

	var req application.BroadcastRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, nethttp.StatusBadRequest, "Invalid request body")
		return
	}

	tenantUUID, _ := uuid.Parse(tenantID)
	cmd := domain.BroadcastNotificationCmd{
		TenantID:  tenantUUID,
		UserIDs:   req.UserIDs,
		EventType: req.EventType,
		Title:     req.Title,
		Body:      req.Body,
		Data:      req.Data,
	}

	count, err := h.service.BroadcastNotification(r.Context(), cmd)
	if err != nil {
		h.respondError(w, nethttp.StatusInternalServerError, "Failed to broadcast notification")
		return
	}

	h.respondJSON(w, nethttp.StatusOK, map[string]interface{}{"sent_count": count})
}

func (h *NotificationHandler) ListNotifications(w nethttp.ResponseWriter, r *nethttp.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	userID := r.Header.Get("X-User-ID")

	tenantUUID, _ := uuid.Parse(tenantID)
	userUUID, _ := uuid.Parse(userID)

	notifications, err := h.service.GetNotifications(r.Context(), tenantUUID, userUUID)
	if err != nil {
		h.respondError(w, nethttp.StatusInternalServerError, "Failed to fetch notifications")
		return
	}

	h.respondJSON(w, nethttp.StatusOK, notifications)
}

func (h *NotificationHandler) GetUnreadCount(w nethttp.ResponseWriter, r *nethttp.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	userID := r.Header.Get("X-User-ID")

	tenantUUID, _ := uuid.Parse(tenantID)
	userUUID, _ := uuid.Parse(userID)

	count, err := h.service.GetUnreadCount(r.Context(), tenantUUID, userUUID)
	if err != nil {
		h.respondError(w, nethttp.StatusInternalServerError, "Failed to get unread count")
		return
	}

	h.respondJSON(w, nethttp.StatusOK, map[string]interface{}{"unread_count": count})
}

func (h *NotificationHandler) MarkAsRead(w nethttp.ResponseWriter, r *nethttp.Request) {
	idStr := r.PathValue("id")
	notificationID, err := uuid.Parse(idStr)
	if err != nil {
		h.respondError(w, nethttp.StatusBadRequest, "Invalid notification ID")
		return
	}

	err = h.service.MarkAsRead(r.Context(), notificationID)
	if err != nil {
		h.respondError(w, nethttp.StatusInternalServerError, "Failed to mark as read")
		return
	}

	h.respondJSON(w, nethttp.StatusOK, map[string]string{"status": "marked as read"})
}

func (h *NotificationHandler) MarkAllAsRead(w nethttp.ResponseWriter, r *nethttp.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	userID := r.Header.Get("X-User-ID")

	tenantUUID, _ := uuid.Parse(tenantID)
	userUUID, _ := uuid.Parse(userID)

	err := h.service.MarkAllAsRead(r.Context(), tenantUUID, userUUID)
	if err != nil {
		h.respondError(w, nethttp.StatusInternalServerError, "Failed to mark all as read")
		return
	}

	h.respondJSON(w, nethttp.StatusOK, map[string]string{"status": "all marked as read"})
}

func (h *NotificationHandler) UpdatePreference(w nethttp.ResponseWriter, r *nethttp.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	userID := r.Header.Get("X-User-ID")

	var req application.PreferenceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, nethttp.StatusBadRequest, "Invalid request body")
		return
	}

	tenantUUID, _ := uuid.Parse(tenantID)
	userUUID, _ := uuid.Parse(userID)

	cmd := domain.UpdatePreferenceCmd{
		TenantID:        tenantUUID,
		UserID:          userUUID,
		EventType:       req.EventType,
		Channels:        req.Channels,
		IsEnabled:       req.IsEnabled,
		QuietHoursStart: req.QuietHoursStart,
		QuietHoursEnd:   req.QuietHoursEnd,
		DigestMode:      req.DigestMode,
	}

	pref, err := h.preferenceService.UpdatePreference(r.Context(), cmd)
	if err != nil {
		h.respondError(w, nethttp.StatusInternalServerError, "Failed to update preference")
		return
	}

	h.respondJSON(w, nethttp.StatusOK, pref)
}

func (h *NotificationHandler) GetPreferences(w nethttp.ResponseWriter, r *nethttp.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	userID := r.Header.Get("X-User-ID")

	tenantUUID, _ := uuid.Parse(tenantID)
	userUUID, _ := uuid.Parse(userID)

	prefs, err := h.preferenceService.GetPreferences(r.Context(), tenantUUID, userUUID)
	if err != nil {
		h.respondError(w, nethttp.StatusInternalServerError, "Failed to fetch preferences")
		return
	}

	h.respondJSON(w, nethttp.StatusOK, prefs)
}

func (h *NotificationHandler) RegisterChannel(w nethttp.ResponseWriter, r *nethttp.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	userID := r.Header.Get("X-User-ID")

	var req application.RegisterChannelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, nethttp.StatusBadRequest, "Invalid request body")
		return
	}

	tenantUUID, _ := uuid.Parse(tenantID)
	userUUID, _ := uuid.Parse(userID)

	cmd := domain.RegisterChannelCmd{
		TenantID: tenantUUID,
		UserID:   userUUID,
		Type:     req.Type,
		Config:   req.Config,
	}

	channel, err := h.channelService.RegisterChannel(r.Context(), cmd)
	if err != nil {
		h.respondError(w, nethttp.StatusInternalServerError, "Failed to register channel")
		return
	}

	h.respondJSON(w, nethttp.StatusCreated, channel)
}

func (h *NotificationHandler) ListChannels(w nethttp.ResponseWriter, r *nethttp.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	userID := r.Header.Get("X-User-ID")

	tenantUUID, _ := uuid.Parse(tenantID)
	userUUID, _ := uuid.Parse(userID)

	channels, err := h.channelService.GetChannels(r.Context(), tenantUUID, userUUID)
	if err != nil {
		h.respondError(w, nethttp.StatusInternalServerError, "Failed to fetch channels")
		return
	}

	h.respondJSON(w, nethttp.StatusOK, channels)
}

func (h *NotificationHandler) DeleteChannel(w nethttp.ResponseWriter, r *nethttp.Request) {
	idStr := r.PathValue("id")
	channelID, err := uuid.Parse(idStr)
	if err != nil {
		h.respondError(w, nethttp.StatusBadRequest, "Invalid channel ID")
		return
	}

	err = h.channelService.DeleteChannel(r.Context(), channelID)
	if err != nil {
		h.respondError(w, nethttp.StatusInternalServerError, "Failed to delete channel")
		return
	}

	w.WriteHeader(nethttp.StatusNoContent)
}

func (h *NotificationHandler) GetDashboard(w nethttp.ResponseWriter, r *nethttp.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")

	tenantUUID, _ := uuid.Parse(tenantID)

	stats, err := h.service.GetDashboardStats(r.Context(), tenantUUID)
	if err != nil {
		h.respondError(w, nethttp.StatusInternalServerError, "Failed to get dashboard stats")
		return
	}

	h.respondJSON(w, nethttp.StatusOK, stats)
}

func (h *NotificationHandler) respondJSON(w nethttp.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func (h *NotificationHandler) respondError(w nethttp.ResponseWriter, statusCode int, message string) {
	h.respondJSON(w, statusCode, map[string]string{"error": message})
}

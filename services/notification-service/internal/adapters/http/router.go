package http

import (
	nethttp "net/http"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/notification-service/internal/application"
)

func SetupRoutes(
	router *nethttp.ServeMux,
	notificationService *application.NotificationService,
	channelService *application.ChannelService,
	preferenceService *application.PreferenceService,
	digestService *application.DigestService,
	log logger.Logger,
) {
	handler := NewNotificationHandler(notificationService, channelService, preferenceService, digestService, log)

	router.HandleFunc("POST /api/v1/notifications/send", handler.SendNotification)
	router.HandleFunc("POST /api/v1/notifications/broadcast", handler.BroadcastNotification)
	router.HandleFunc("GET /api/v1/notifications", handler.ListNotifications)
	router.HandleFunc("GET /api/v1/notifications/unread-count", handler.GetUnreadCount)
	router.HandleFunc("PUT /api/v1/notifications/{id}/read", handler.MarkAsRead)
	router.HandleFunc("PUT /api/v1/notifications/read-all", handler.MarkAllAsRead)
	router.HandleFunc("GET /api/v1/notifications/preferences", handler.GetPreferences)
	router.HandleFunc("PUT /api/v1/notifications/preferences", handler.UpdatePreference)
	router.HandleFunc("POST /api/v1/notifications/channels", handler.RegisterChannel)
	router.HandleFunc("GET /api/v1/notifications/channels", handler.ListChannels)
	router.HandleFunc("DELETE /api/v1/notifications/channels/{id}", handler.DeleteChannel)
	router.HandleFunc("GET /api/v1/notifications/dashboard", handler.GetDashboard)
}

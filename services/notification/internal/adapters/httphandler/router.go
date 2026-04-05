package httphandler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jeckersberger/EquipFlow/pkg/common/middleware"
)

func NewRouter(
	notificationHandler *NotificationHandler,
	preferenceHandler *PreferenceHandler,
	emailHandler *EmailHandler,
	wsHandler *WSHandler,
	healthHandler http.HandlerFunc,
	livenessHandler http.HandlerFunc,
	jwtMiddleware func(http.Handler) http.Handler,
	corsMiddleware func(http.Handler) http.Handler,
	recoveryMiddleware func(http.Handler) http.Handler,
	requestIDMiddleware func(http.Handler) http.Handler,
) *chi.Mux {
	r := chi.NewRouter()

	r.Use(recoveryMiddleware)
	r.Use(middleware.MaxBodySize(1 << 20)) // 1 MB body limit
	r.Use(requestIDMiddleware)
	r.Use(corsMiddleware)

	r.Get("/health", healthHandler)
	r.Get("/livez", livenessHandler)

	r.Group(func(r chi.Router) {
		r.Use(jwtMiddleware)

		r.Route("/api/v1/notifications", func(r chi.Router) {
			// Read: all authenticated users
			r.Get("/", notificationHandler.List)
			r.Get("/unread-count", notificationHandler.UnreadCount)
			r.Get("/ws", wsHandler.ServeWS)

			// Write: admin, manager, or user
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole("admin", "manager", "user"))
				r.Post("/", notificationHandler.Create)
				r.Patch("/{id}/read", notificationHandler.MarkAsRead)
				r.Post("/mark-all-read", notificationHandler.MarkAllRead)
			})
		})

		r.Route("/api/v1/notification-preferences", func(r chi.Router) {
			// Read: all authenticated users
			r.Get("/", preferenceHandler.List)

			// Write: admin, manager, or user
			r.With(middleware.RequireRole("admin", "manager", "user")).Put("/", preferenceHandler.Update)
		})

		r.Route("/api/v1/email", func(r chi.Router) {
			// Read: all authenticated users
			r.Get("/status", emailHandler.Status)

			// Write: admin, manager, or user
			r.With(middleware.RequireRole("admin", "manager", "user")).Post("/send", emailHandler.Send)
		})
	})

	return r
}

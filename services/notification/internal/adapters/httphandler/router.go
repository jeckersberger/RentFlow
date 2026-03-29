package httphandler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewRouter(
	notificationHandler *NotificationHandler,
	preferenceHandler *PreferenceHandler,
	emailHandler *EmailHandler,
	healthHandler http.HandlerFunc,
	livenessHandler http.HandlerFunc,
	jwtMiddleware func(http.Handler) http.Handler,
	corsMiddleware func(http.Handler) http.Handler,
	recoveryMiddleware func(http.Handler) http.Handler,
	requestIDMiddleware func(http.Handler) http.Handler,
) *chi.Mux {
	r := chi.NewRouter()

	r.Use(recoveryMiddleware)
	r.Use(requestIDMiddleware)
	r.Use(corsMiddleware)

	r.Get("/health", healthHandler)
	r.Get("/livez", livenessHandler)

	r.Group(func(r chi.Router) {
		r.Use(jwtMiddleware)

		r.Route("/api/v1/notifications", func(r chi.Router) {
			r.Get("/", notificationHandler.List)
			r.Post("/", notificationHandler.Create)
			r.Get("/unread-count", notificationHandler.UnreadCount)
			r.Patch("/{id}/read", notificationHandler.MarkAsRead)
			r.Post("/mark-all-read", notificationHandler.MarkAllRead)
		})

		r.Route("/api/v1/notification-preferences", func(r chi.Router) {
			r.Get("/", preferenceHandler.List)
			r.Put("/", preferenceHandler.Update)
		})

		r.Route("/api/v1/email", func(r chi.Router) {
			r.Post("/send", emailHandler.Send)
			r.Get("/status", emailHandler.Status)
		})
	})

	return r
}

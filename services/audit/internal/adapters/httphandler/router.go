package httphandler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewRouter(
	auditLogHandler *AuditLogHandler,
	auditPolicyHandler *AuditPolicyHandler,
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

		r.Route("/api/v1/audit-logs", func(r chi.Router) {
			r.Post("/", auditLogHandler.Create)
			r.Get("/", auditLogHandler.List)
			r.Post("/append", auditLogHandler.Append)
			r.Get("/verify", auditLogHandler.Verify)
			r.Get("/export", auditLogHandler.Export)
			r.Get("/{id}", auditLogHandler.Get)
		})

		r.Route("/api/v1/audit-policies", func(r chi.Router) {
			r.Post("/", auditPolicyHandler.Create)
			r.Get("/", auditPolicyHandler.List)
			r.Put("/{id}", auditPolicyHandler.Update)
		})
	})

	return r
}

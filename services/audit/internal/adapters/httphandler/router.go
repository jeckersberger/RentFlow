package httphandler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jeckersberger/EquipFlow/pkg/common/middleware"
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
	r.Use(middleware.MaxBodySize(1 << 20)) // 1 MB body limit
	r.Use(requestIDMiddleware)
	r.Use(corsMiddleware)

	r.Get("/health", healthHandler)
	r.Get("/livez", livenessHandler)

	r.Group(func(r chi.Router) {
		r.Use(jwtMiddleware)

		// Audit: read-only for all authenticated users (no creates from API)
		r.Route("/api/v1/audit-logs", func(r chi.Router) {
			r.Get("/", auditLogHandler.List)
			r.Get("/verify", auditLogHandler.Verify)
			r.Get("/export", auditLogHandler.Export)
			r.Get("/{id}", auditLogHandler.Get)
		})

		r.Route("/api/v1/audit-policies", func(r chi.Router) {
			r.Get("/", auditPolicyHandler.List)
		})
	})

	return r
}

package httphandler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jeckersberger/EquipFlow/pkg/common/middleware"
)

func NewRouter(
	policyHandler *PolicyHandler,
	equipmentHandler *EquipmentHandler,
	claimHandler *ClaimHandler,
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

		r.Route("/api/v1/insurance-policies", func(r chi.Router) {
			// Read: all authenticated users
			r.Get("/", policyHandler.List)
			r.Get("/{id}", policyHandler.Get)
			r.Get("/{id}/equipment", equipmentHandler.List)

			// Write: manager or admin only
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole("admin", "manager"))
				r.Post("/", policyHandler.Create)
				r.Put("/{id}", policyHandler.Update)
				r.Post("/{id}/equipment", equipmentHandler.Add)
			})
		})

		r.Route("/api/v1/insurance-claims", func(r chi.Router) {
			// Read: all authenticated users
			r.Get("/", claimHandler.List)
			r.Get("/{id}", claimHandler.Get)

			// Write: manager or admin only
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole("admin", "manager"))
				r.Post("/", claimHandler.Create)
				r.Put("/{id}", claimHandler.Update)
				r.Patch("/{id}/status", claimHandler.UpdateStatus)
			})
		})
	})

	return r
}

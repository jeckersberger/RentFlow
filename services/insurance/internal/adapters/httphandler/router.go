package httphandler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
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
	r.Use(requestIDMiddleware)
	r.Use(corsMiddleware)

	r.Get("/health", healthHandler)
	r.Get("/livez", livenessHandler)

	r.Group(func(r chi.Router) {
		r.Use(jwtMiddleware)

		r.Route("/api/v1/insurance-policies", func(r chi.Router) {
			r.Get("/", policyHandler.List)
			r.Post("/", policyHandler.Create)
			r.Get("/{id}", policyHandler.Get)
			r.Put("/{id}", policyHandler.Update)
			r.Post("/{id}/equipment", equipmentHandler.Add)
			r.Get("/{id}/equipment", equipmentHandler.List)
		})

		r.Route("/api/v1/insurance-claims", func(r chi.Router) {
			r.Get("/", claimHandler.List)
			r.Post("/", claimHandler.Create)
			r.Get("/{id}", claimHandler.Get)
			r.Put("/{id}", claimHandler.Update)
		})
	})

	return r
}

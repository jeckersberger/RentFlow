package httphandler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jeckersberger/EquipFlow/pkg/common/middleware"
)

func NewRouter(
	predictionHandler *PredictionHandler,
	suggestionHandler *SuggestionHandler,
	trainingDataHandler *TrainingDataHandler,
	aiProviderHandler *AIProviderHandler,
	packlistHandler *PacklistHandler,
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

		r.Route("/api/v1/ai-predictions", func(r chi.Router) {
			// Read: all authenticated users
			r.Get("/", predictionHandler.List)
			r.Get("/{id}", predictionHandler.Get)

			// Write: admin, manager, or user (no deletes in AI)
			r.With(middleware.RequireRole("admin", "manager", "user")).Post("/", predictionHandler.Create)
		})

		r.Route("/api/v1/ai-suggestions", func(r chi.Router) {
			// Read: all authenticated users
			r.Get("/", suggestionHandler.List)

			// Write: admin, manager, or user (no deletes in AI)
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole("admin", "manager", "user"))
				r.Post("/", suggestionHandler.Create)
				r.Patch("/{id}/accept", suggestionHandler.Accept)
				r.Patch("/{id}/dismiss", suggestionHandler.Dismiss)
			})
		})

		r.Route("/api/v1/ai-training-data", func(r chi.Router) {
			// Read: all authenticated users
			r.Get("/", trainingDataHandler.List)

			// Write: admin, manager, or user (no deletes in AI)
			r.With(middleware.RequireRole("admin", "manager", "user")).Post("/", trainingDataHandler.Create)
		})

		r.Route("/api/v1/ai", func(r chi.Router) {
			// Read: all authenticated users
			r.Get("/status", aiProviderHandler.Status)
			r.Get("/readyz", packlistHandler.Readyz)
			r.Get("/meta", packlistHandler.Meta)

			// Write: admin, manager, or user (no deletes in AI)
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole("admin", "manager", "user"))
				r.Post("/complete", aiProviderHandler.Complete)
				r.Post("/smart-asset", aiProviderHandler.SmartAsset)
				r.Post("/packlist/suggest", packlistHandler.Suggest)
			})
		})
	})

	return r
}

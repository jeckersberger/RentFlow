package httphandler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewRouter(
	predictionHandler *PredictionHandler,
	suggestionHandler *SuggestionHandler,
	trainingDataHandler *TrainingDataHandler,
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

		r.Route("/api/v1/ai-predictions", func(r chi.Router) {
			r.Get("/", predictionHandler.List)
			r.Post("/", predictionHandler.Create)
			r.Get("/{id}", predictionHandler.Get)
		})

		r.Route("/api/v1/ai-suggestions", func(r chi.Router) {
			r.Get("/", suggestionHandler.List)
			r.Post("/", suggestionHandler.Create)
			r.Patch("/{id}/accept", suggestionHandler.Accept)
			r.Patch("/{id}/dismiss", suggestionHandler.Dismiss)
		})

		r.Route("/api/v1/ai-training-data", func(r chi.Router) {
			r.Get("/", trainingDataHandler.List)
			r.Post("/", trainingDataHandler.Create)
		})
	})

	return r
}

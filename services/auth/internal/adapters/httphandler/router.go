package httphandler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// NewRouter creates and configures the chi router with all auth-service routes.
func NewRouter(
	authHandler *AuthHandler,
	userHandler *UserHandler,
	tenantHandler *TenantHandler,
	setupHandler *SetupHandler,
	configHandler *ConfigHandler,
	healthHandler http.HandlerFunc,
	livenessHandler http.HandlerFunc,
	jwtMiddleware func(http.Handler) http.Handler,
	corsMiddleware func(http.Handler) http.Handler,
	rateLimitMiddleware func(http.Handler) http.Handler,
	recoveryMiddleware func(http.Handler) http.Handler,
	requestIDMiddleware func(http.Handler) http.Handler,
) *chi.Mux {
	r := chi.NewRouter()

	// Global middleware (order matters: outermost first).
	r.Use(recoveryMiddleware)
	r.Use(requestIDMiddleware)
	r.Use(corsMiddleware)

	// Health endpoints (no auth required).
	r.Get("/health", healthHandler)
	r.Get("/livez", livenessHandler)

	// Setup endpoints (no auth required — used for initial bootstrap).
	r.Route("/api/v1/setup", func(r chi.Router) {
		r.Post("/init", setupHandler.Init)
		r.Post("/complete", setupHandler.Complete)
	})

	// Auth endpoints (no JWT required, but rate-limited).
	r.Route("/api/v1/auth", func(r chi.Router) {
		r.Use(rateLimitMiddleware)
		r.Post("/login", authHandler.Login)
		r.Post("/refresh", authHandler.Refresh)
		r.Post("/forgot-password", authHandler.ForgotPassword)
		r.Post("/reset-password", authHandler.ResetPassword)
	})

	// Protected routes (JWT required).
	r.Group(func(r chi.Router) {
		r.Use(jwtMiddleware)

		// Auth — authenticated actions.
		r.Post("/api/v1/auth/logout", authHandler.Logout)
		r.Get("/api/v1/auth/me", authHandler.Me)
		r.Put("/api/v1/auth/profile", authHandler.UpdateProfile)
		r.Put("/api/v1/auth/password", authHandler.ChangePassword)

		// Users — admin management.
		r.Route("/api/v1/users", func(r chi.Router) {
			r.Get("/", userHandler.List)
			r.Post("/", userHandler.Create)
			r.Get("/{id}", userHandler.Get)
			r.Put("/{id}", userHandler.Update)
			r.Put("/{id}/role", userHandler.UpdateRole)
			r.Delete("/{id}", userHandler.Delete)
			r.Post("/invite", userHandler.Invite)
		})

		// Tenant — current tenant info.
		r.Get("/api/v1/tenants/current", tenantHandler.GetCurrent)
		r.Put("/api/v1/tenants/current", tenantHandler.Update)

		// Config — per-tenant configuration.
		r.Route("/api/v1/config", func(r chi.Router) {
			r.Get("/", configHandler.GetAll)
			r.Get("/{key}", configHandler.Get)
			r.Put("/{key}", configHandler.Set)
			r.Delete("/{key}", configHandler.Delete)
		})
	})

	return r
}

package http

import (
	"net/http"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/pkg/common/middleware"
	"github.com/jeckersberger/rentflow/services/auth-service/internal/application"
)

// SetupRoutes sets up all HTTP routes
func SetupRoutes(
	mux *http.ServeMux,
	userService *application.UserService,
	tenantService *application.TenantService,
	jwtSecret string,
	log logger.Logger,
) {
	handlers := NewHandlers(userService, tenantService, log)

	// Auth routes (no authentication required)
	mux.HandleFunc("POST /api/v1/auth/register", handlers.Register)
	mux.HandleFunc("POST /api/v1/auth/login", handlers.Login)
	mux.HandleFunc("POST /api/v1/auth/refresh", handlers.Refresh)

	// Authenticated routes
	authMiddleware := middleware.JWTAuthMiddleware(jwtSecret)

	// Logout (authenticated)
	mux.HandleFunc("POST /api/v1/auth/logout", authMiddleware(http.HandlerFunc(handlers.Logout)).ServeHTTP)

	// User profile (authenticated)
	mux.HandleFunc("GET /api/v1/auth/me", authMiddleware(http.HandlerFunc(handlers.GetMe)).ServeHTTP)
	mux.HandleFunc("PUT /api/v1/auth/password", authMiddleware(http.HandlerFunc(handlers.ChangePassword)).ServeHTTP)
	mux.HandleFunc("PUT /api/v1/auth/profile", authMiddleware(http.HandlerFunc(handlers.UpdateProfile)).ServeHTTP)

	// User management (admin only)
	mux.HandleFunc("GET /api/v1/users", authMiddleware(http.HandlerFunc(handlers.ListUsers)).ServeHTTP)
	mux.HandleFunc("GET /api/v1/users/{id}", authMiddleware(http.HandlerFunc(handlers.GetUser)).ServeHTTP)
	mux.HandleFunc("PUT /api/v1/users/{id}/roles", authMiddleware(http.HandlerFunc(handlers.AssignRole)).ServeHTTP)
	mux.HandleFunc("DELETE /api/v1/users/{id}", authMiddleware(http.HandlerFunc(handlers.DeleteUser)).ServeHTTP)

	// Tenant management
	mux.HandleFunc("POST /api/v1/tenants", handlers.CreateTenant)
	mux.HandleFunc("GET /api/v1/tenants/{id}", handlers.GetTenant)
	mux.HandleFunc("PUT /api/v1/tenants/{id}", handlers.UpdateTenant)
}

package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/config"
	"github.com/jeckersberger/rentflow/pkg/common/health"
	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/pkg/common/middleware"
)

// Server represents an HTTP server with common functionality
type Server struct {
	router      *http.ServeMux
	config      *config.Config
	logger      logger.Logger
	middlewares []func(http.Handler) http.Handler
	server      *http.Server
	checker     *health.HealthChecker
}

// NewServer creates a new HTTP server with common configuration
func NewServer(cfg *config.Config, log logger.Logger) *Server {
	if cfg == nil {
		panic("config cannot be nil")
	}
	if log == nil {
		panic("logger cannot be nil")
	}

	return &Server{
		router:      http.NewServeMux(),
		config:      cfg,
		logger:      log,
		middlewares: []func(http.Handler) http.Handler{},
		checker:     health.NewHealthChecker(),
	}
}

// Use adds a middleware to the server
// Middlewares are applied in the order they are added
func (s *Server) Use(mw func(http.Handler) http.Handler) {
	s.middlewares = append(s.middlewares, mw)
}

// Handle registers a handler for the given pattern
func (s *Server) Handle(pattern string, handler http.Handler) {
	// Wrap handler with middlewares
	wrappedHandler := handler
	for i := len(s.middlewares) - 1; i >= 0; i-- {
		wrappedHandler = s.middlewares[i](wrappedHandler)
	}

	s.router.Handle(pattern, wrappedHandler)
}

// HandleFunc registers a handler function for the given pattern
func (s *Server) HandleFunc(pattern string, handler http.HandlerFunc) {
	s.Handle(pattern, handler)
}

// RegisterHealthChecks registers health checks
func (s *Server) RegisterHealthChecks(checks ...health.HealthCheck) {
	for _, check := range checks {
		s.checker.Register(check)
	}
}

// setupHealthRoutes registers the health check routes
func (s *Server) setupHealthRoutes() {
	// Health check endpoint
	s.router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		response := s.checker.Check()
		w.Header().Set("Content-Type", "application/json")

		statusCode := http.StatusOK
		if response.Status != health.StatusHealthy {
			statusCode = http.StatusServiceUnavailable
		}

		w.WriteHeader(statusCode)
		_ = json.NewEncoder(w).Encode(response)
	})

	// Ready check endpoint (similar to health but can be more strict)
	s.router.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		response := s.checker.Check()
		w.Header().Set("Content-Type", "application/json")

		statusCode := http.StatusOK
		if response.Status != health.StatusHealthy {
			statusCode = http.StatusServiceUnavailable
		}

		w.WriteHeader(statusCode)
		_ = json.NewEncoder(w).Encode(response)
	})

	// Liveness probe
	s.router.HandleFunc("/live", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status": "alive",
		})
	})
}

// Router returns the underlying HTTP router
// This allows services to register additional routes directly
func (s *Server) Router() *http.ServeMux {
	return s.router
}

// Start starts the HTTP server with graceful shutdown support
func (s *Server) Start(ctx context.Context) error {
	// Setup health routes
	s.setupHealthRoutes()

	// Create HTTP server
	s.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", s.config.ServicePort),
		Handler:      s.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Log startup
	s.logger.Info("Starting HTTP server",
		"service", s.config.ServiceName,
		"port", s.config.ServicePort,
		"environment", s.config.Environment,
	)

	// Start server in a goroutine
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("Server error", err)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Block until signal received
	<-sigChan

	// Graceful shutdown
	s.logger.Info("Shutting down server")

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second) //nolint:contextcheck // intentional new context for graceful shutdown
	defer cancel()

	if err := s.server.Shutdown(shutdownCtx); err != nil { //nolint:contextcheck // using shutdown context from above
		s.logger.Error("Server shutdown error", err)
		return err
	}

	s.logger.Info("Server stopped")
	return nil
}

// StartWithContext starts the server and returns a channel that signals when shutdown is complete
func (s *Server) StartWithContext(ctx context.Context) <-chan error {
	errChan := make(chan error, 1)

	// Setup health routes
	s.setupHealthRoutes()

	// Create HTTP server
	s.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", s.config.ServicePort),
		Handler:      s.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Log startup
	s.logger.Info("Starting HTTP server",
		"service", s.config.ServiceName,
		"port", s.config.ServicePort,
		"environment", s.config.Environment,
	)

	// Start server in a goroutine
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	// Monitor context for cancellation
	go func() {
		<-ctx.Done()
		_ = s.Shutdown() //nolint:contextcheck // intentional: Shutdown creates its own timeout context
		errChan <- ctx.Err()
	}()

	return errChan
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown() error {
	if s.server == nil {
		return nil
	}

	s.logger.Info("Shutting down server")

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.server.Shutdown(shutdownCtx); err != nil {
		s.logger.Error("Server shutdown error", err)
		return err
	}

	s.logger.Info("Server stopped")
	return nil
}

// GetAddr returns the server address
func (s *Server) GetAddr() string {
	return s.server.Addr
}

// NotFoundHandler returns a 404 handler
func (s *Server) NotFoundHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code":    "NOT_FOUND",
			"message": "Resource not found",
			"path":    r.RequestURI,
		})
	}
}

// MethodNotAllowedHandler returns a 405 handler
func (s *Server) MethodNotAllowedHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code":    "METHOD_NOT_ALLOWED",
			"message": "Method not allowed",
			"method":  r.Method,
			"path":    r.RequestURI,
		})
	}
}

// ServerConfig represents server configuration
type ServerConfig struct {
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	MaxTimeout   time.Duration
}

// DefaultServerConfig returns default server configuration
func DefaultServerConfig() ServerConfig {
	return ServerConfig{
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
		MaxTimeout:   30 * time.Second,
	}
}

// WithMiddlewares is a helper to add multiple middlewares at once
func (s *Server) WithMiddlewares(mws ...func(http.Handler) http.Handler) *Server {
	for _, mw := range mws {
		s.Use(mw)
	}
	return s
}

// WithRequestLogging adds request logging middleware
func (s *Server) WithRequestLogging() *Server {
	s.Use(middleware.RequestLogging(s.logger))
	return s
}

// WithPanicRecovery adds panic recovery middleware
func (s *Server) WithPanicRecovery() *Server {
	s.Use(middleware.PanicRecovery(s.logger))
	return s
}

// WithCORS adds CORS middleware
func (s *Server) WithCORS() *Server {
	s.Use(middleware.CORSMiddleware(s.config))
	return s
}

// WithJWTAuth adds JWT authentication middleware
func (s *Server) WithJWTAuth() *Server {
	s.Use(middleware.JWTAuthMiddleware(s.config.JWTSecret))
	return s
}

// WithTenantMiddleware adds tenant middleware
func (s *Server) WithTenantMiddleware() *Server {
	s.Use(middleware.TenantMiddleware())
	return s
}

// WithRateLimit adds rate limiting middleware
func (s *Server) WithRateLimit(requestsPerMinute int) *Server {
	s.Use(middleware.RateLimitMiddleware(requestsPerMinute))
	return s
}

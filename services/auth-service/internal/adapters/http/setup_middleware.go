package http

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/auth-service/internal/application"
)

// SetupGuardMiddleware ensures setup is completed before allowing service access
// Allows: /api/v1/setup/*, /health, /ready, /api/v1/auth/.well-known/jwks
// Blocks: All other routes with 503 if setup is not completed
func SetupGuardMiddleware(setupService *application.SetupService, log logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Allowed paths that bypass setup check
			allowedPaths := []string{
				"/api/v1/setup/",
				"/health",
				"/ready",
				"/api/v1/auth/.well-known/jwks",
			}

			isAllowed := false
			for _, path := range allowedPaths {
				if strings.HasPrefix(r.URL.Path, path) {
					isAllowed = true
					break
				}
			}

			if isAllowed {
				next.ServeHTTP(w, r)
				return
			}

			// Check if setup is required
			setupRequired, err := setupService.IsSetupRequired(r.Context())
			if err != nil {
				log.Error("failed to check setup status", err)
				writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to check setup status")
				return
			}

			if setupRequired {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusServiceUnavailable)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"code":    "SERVICE_UNAVAILABLE",
					"message": "Setup required. Please complete the setup wizard at GET /api/v1/setup/status and POST /api/v1/setup/complete",
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

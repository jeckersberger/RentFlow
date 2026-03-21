package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
)

// PanicRecovery creates a middleware that recovers from panics
// It logs the panic and returns a 500 Internal Server Error
func PanicRecovery(log logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					// Log the panic
					log.Error("Panic recovered",
						"error", err,
						"path", r.RequestURI,
						"method", r.Method,
						"stack_trace", string(debug.Stack()),
					)

					// Write error response
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)

					errorResponse := map[string]interface{}{
						"code":    "INTERNAL_SERVER_ERROR",
						"message": "An internal error occurred",
					}

					json.NewEncoder(w).Encode(errorResponse)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// RecoveryWithCustomHandler creates a middleware that recovers from panics
// and calls a custom handler function
func RecoveryWithCustomHandler(log logger.Logger, handler func(http.ResponseWriter, *http.Request, interface{})) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					// Log the panic
					log.Error("Panic recovered",
						"error", err,
						"path", r.RequestURI,
						"method", r.Method,
						"stack_trace", string(debug.Stack()),
					)

					// Call custom handler
					handler(w, r, err)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// PanicDetails holds information about a panic
type PanicDetails struct {
	Error      string `json:"error"`
	StackTrace string `json:"stack_trace,omitempty"`
	Path       string `json:"path"`
	Method     string `json:"method"`
	Timestamp  string `json:"timestamp"`
}

// RecoveryWithDetails creates a middleware that recovers from panics
// and returns detailed error information (useful for development/staging)
func RecoveryWithDetails(log logger.Logger, includeStackTrace bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					errorMsg := fmt.Sprintf("%v", err)
					stackTrace := string(debug.Stack())

					// Log the panic
					log.Error("Panic recovered",
						"error", errorMsg,
						"path", r.RequestURI,
						"method", r.Method,
						"stack_trace", stackTrace,
					)

					// Write error response
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)

					details := PanicDetails{
						Error:     errorMsg,
						Path:      r.RequestURI,
						Method:    r.Method,
						Timestamp: fmt.Sprintf("%d", time.Now().Unix()),
					}

					if includeStackTrace {
						details.StackTrace = stackTrace
					}

					json.NewEncoder(w).Encode(details)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}


package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/pkg/common/response"
)

// Recovery returns a middleware that recovers from panics, logs the stack
// trace, and returns an HTTP 500 response.
func Recovery(logger zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					stack := debug.Stack()

					logger.Error().
						Interface("panic", rec).
						Str("method", r.Method).
						Str("path", r.URL.Path).
						Bytes("stack", stack).
						Msg("panic recovered")

					response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred")
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

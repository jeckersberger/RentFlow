package middleware

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rentflow/pkg/common/logger"
)

// ResponseWriter wraps http.ResponseWriter to capture status code and written bytes
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    int64
}

// WriteHeader captures the status code
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Write captures the number of bytes written
func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.written += int64(n)
	return n, err
}

// RequestID type for context
type requestIDKey string

const (
	requestIDContextKey requestIDKey = "request_id"
)

// RequestLogging creates a middleware that logs HTTP requests
// Logs method, path, status code, duration, and request ID
func RequestLogging(log logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Generate request ID
			requestID := generateRequestID()

			// Create a response writer wrapper to capture status code
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			// Record start time
			start := time.Now()

			// Call the next handler
			next.ServeHTTP(wrapped, r)

			// Calculate duration
			duration := time.Since(start)

			// Log the request
			log.Info("HTTP Request",
				"request_id", requestID,
				"method", r.Method,
				"path", r.RequestURI,
				"status", wrapped.statusCode,
				"duration_ms", duration.Milliseconds(),
				"bytes_written", wrapped.written,
			)
		})
	}
}

// RequestWithID creates a middleware that adds a request ID to the context
func RequestWithID() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := generateRequestID()

			// Add request ID to response header
			w.Header().Set("X-Request-ID", requestID)

			// Add request ID to context
			ctx := context.WithValue(r.Context(), requestIDContextKey, requestID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetRequestID extracts the request ID from the context
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDContextKey).(string); ok {
		return id
	}
	return ""
}

// RequestBodyLogging creates a middleware that logs request bodies (for debugging)
// WARNING: This should only be used in development, not production
// It can expose sensitive information
func RequestBodyLogging(log logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only log for certain methods
			if r.Method == "GET" || r.Method == "HEAD" || r.Method == "OPTIONS" {
				next.ServeHTTP(w, r)
				return
			}

			// Read body
			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				log.Error("Failed to read request body", err)
				next.ServeHTTP(w, r)
				return
			}

			// Log body (truncate if too large)
			bodyStr := string(bodyBytes)
			if len(bodyStr) > 1000 {
				bodyStr = bodyStr[:1000] + "..."
			}
			log.Debug("Request body", "body", bodyStr)

			// Restore body
			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

			next.ServeHTTP(w, r)
		})
	}
}

// ResponseBodyLogging creates a middleware that logs response bodies (for debugging)
// WARNING: This should only be used in development, not production
// It can expose sensitive information
func ResponseBodyLogging(log logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Create a response writer wrapper that captures body
			buf := &bytes.Buffer{}
			wrapped := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			// Create a multi-writer to write to both the response and buffer
			mw := io.MultiWriter(w, buf)

			// Replace the response writer temporarily
			originalResponseWriter := wrapped.ResponseWriter
			wrapped.ResponseWriter = writerFunc(func(p []byte) (int, error) {
				return mw.Write(p)
			})

			next.ServeHTTP(wrapped, r)

			// Log body (truncate if too large)
			bodyStr := buf.String()
			if len(bodyStr) > 1000 {
				bodyStr = bodyStr[:1000] + "..."
			}
			log.Debug("Response body", "body", bodyStr)
		})
	}
}

// writerFunc adapts a function to the http.ResponseWriter interface
type writerFunc func([]byte) (int, error)

func (f writerFunc) Write(p []byte) (int, error) {
	return f(p)
}

// generateRequestID generates a unique request ID
// In production, you might use a library like google/uuid
func generateRequestID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// Add missing import
import "context"

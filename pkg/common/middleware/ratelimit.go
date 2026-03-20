package middleware

import (
	"net/http"
	"sync"
	"time"
)

// RateLimiter is a simple in-memory rate limiter
// For production use, consider using Redis for distributed rate limiting
type RateLimiter struct {
	requestsPerMinute int
	window            time.Duration
	clients           map[string]*clientLimit
	mu                sync.RWMutex
	cleanupTicker     *time.Ticker
}

// clientLimit tracks requests for a single client
type clientLimit struct {
	requests  int
	firstSeen time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(requestsPerMinute int) *RateLimiter {
	rl := &RateLimiter{
		requestsPerMinute: requestsPerMinute,
		window:            time.Minute,
		clients:           make(map[string]*clientLimit),
	}

	// Start a cleanup goroutine to remove old entries
	rl.cleanupTicker = time.NewTicker(5 * time.Minute)
	go rl.cleanup()

	return rl
}

// Allow checks if the request is allowed based on the rate limit
// Returns true if the request is allowed, false if it exceeds the limit
func (rl *RateLimiter) Allow(clientID string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	limit, exists := rl.clients[clientID]

	// If client doesn't exist or the window has expired, create new entry
	if !exists || now.Sub(limit.firstSeen) > rl.window {
		rl.clients[clientID] = &clientLimit{
			requests:  1,
			firstSeen: now,
		}
		return true
	}

	// Check if limit is exceeded
	if limit.requests >= rl.requestsPerMinute {
		return false
	}

	// Increment request count
	limit.requests++
	return true
}

// GetRemainingRequests returns the number of remaining requests for a client
func (rl *RateLimiter) GetRemainingRequests(clientID string) int {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	limit, exists := rl.clients[clientID]
	if !exists {
		return rl.requestsPerMinute
	}

	remaining := rl.requestsPerMinute - limit.requests
	if remaining < 0 {
		return 0
	}

	return remaining
}

// Reset resets the rate limit for a specific client
func (rl *RateLimiter) Reset(clientID string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	delete(rl.clients, clientID)
}

// cleanup removes expired entries
func (rl *RateLimiter) cleanup() {
	for range rl.cleanupTicker.C {
		rl.mu.Lock()

		now := time.Now()
		for clientID, limit := range rl.clients {
			if now.Sub(limit.firstSeen) > rl.window {
				delete(rl.clients, clientID)
			}
		}

		rl.mu.Unlock()
	}
}

// Stop stops the cleanup goroutine
func (rl *RateLimiter) Stop() {
	if rl.cleanupTicker != nil {
		rl.cleanupTicker.Stop()
	}
}

// RateLimitMiddleware creates a middleware that applies rate limiting
// It uses the client's IP address as the identifier
func RateLimitMiddleware(requestsPerMinute int) func(http.Handler) http.Handler {
	limiter := NewRateLimiter(requestsPerMinute)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get client identifier (IP address)
			clientID := getClientIP(r)

			// Check rate limit
			if !limiter.Allow(clientID) {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Retry-After", "60")
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"code":"RATE_LIMIT_EXCEEDED","message":"Too many requests, please try again later"}`))
				return
			}

			// Add rate limit headers
			remaining := limiter.GetRemainingRequests(clientID)
			w.Header().Set("X-RateLimit-Limit", "60")
			w.Header().Set("X-RateLimit-Remaining", string(rune(remaining)))
			w.Header().Set("X-RateLimit-Reset", string(rune(time.Now().Add(time.Minute).Unix())))

			next.ServeHTTP(w, r)
		})
	}
}

// RateLimitByUserMiddleware creates a middleware that applies rate limiting per user
// It uses the user ID from the JWT claims
func RateLimitByUserMiddleware(requestsPerMinute int) func(http.Handler) http.Handler {
	limiter := NewRateLimiter(requestsPerMinute)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get user ID from claims
			userID := GetUserID(r.Context())
			if userID == "" {
				userID = getClientIP(r)
			}

			// Check rate limit
			if !limiter.Allow(userID) {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Retry-After", "60")
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"code":"RATE_LIMIT_EXCEEDED","message":"Too many requests, please try again later"}`))
				return
			}

			// Add rate limit headers
			remaining := limiter.GetRemainingRequests(userID)
			w.Header().Set("X-RateLimit-Limit", "60")
			w.Header().Set("X-RateLimit-Remaining", string(rune(remaining)))
			w.Header().Set("X-RateLimit-Reset", string(rune(time.Now().Add(time.Minute).Unix())))

			next.ServeHTTP(w, r)
		})
	}
}

// getClientIP extracts the client IP from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (for proxies)
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		return forwarded
	}

	// Check X-Real-IP header
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}

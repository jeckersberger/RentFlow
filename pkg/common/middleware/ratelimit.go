package middleware

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/jeckersberger/EquipFlow/pkg/common/response"
)

// RateLimitConfig configures the Redis-based rate limiter.
type RateLimitConfig struct {
	// MaxRequests is the maximum number of requests allowed per window.
	MaxRequests int
	// Window is the duration of the rate limit window.
	Window time.Duration
	// Client is the Redis client used for storing counters.
	Client *redis.Client
}

// RateLimit returns a middleware that enforces rate limiting per IP + endpoint
// using a Redis sliding window counter.
func RateLimit(cfg RateLimitConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			ip := extractIP(r)
			key := fmt.Sprintf("ratelimit:%s:%s", ip, r.URL.Path)

			allowed, err := checkRateLimit(ctx, cfg.Client, key, cfg.MaxRequests, cfg.Window)
			if err != nil {
				// If Redis is down, allow the request (fail open).
				next.ServeHTTP(w, r)
				return
			}

			if !allowed {
				response.Error(w, http.StatusTooManyRequests, "RATE_LIMITED", "Too many requests, please try again later")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func checkRateLimit(ctx context.Context, client *redis.Client, key string, max int, window time.Duration) (bool, error) {
	pipe := client.Pipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, window)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, err
	}

	count := incr.Val()
	return count <= int64(max), nil
}

func extractIP(r *http.Request) string {
	// Prefer Cloudflare's trusted header (set by the edge, not spoofable).
	if cfIP := r.Header.Get("CF-Connecting-IP"); cfIP != "" {
		return cfIP
	}

	// Fall back to the remote address (set by Traefik / the TCP connection).
	// We do NOT trust X-Forwarded-For or X-Real-IP from untrusted clients
	// because they can be spoofed to bypass rate limiting.
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

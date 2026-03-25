package http

import (
	"encoding/json"
	nethttp "net/http"
	"strings"
	"sync"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/auth-service/internal/application"
)

// RateLimiter implementiert ein einfaches Token-Bucket Rate-Limiting pro IP
type RateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	rate     int           // Max Anfragen
	window   time.Duration // Zeitfenster
	logger   logger.Logger
}

type visitor struct {
	count    int
	windowStart time.Time
}

// NewRateLimiter erstellt einen neuen Rate-Limiter
func NewRateLimiter(rate int, window time.Duration, log logger.Logger) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rate,
		window:   window,
		logger:   log,
	}

	// Cleanup goroutine - entfernt abgelaufene Eintraege alle 5 Minuten
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			rl.cleanup()
		}
	}()

	return rl
}

func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for ip, v := range rl.visitors {
		if now.Sub(v.windowStart) > rl.window {
			delete(rl.visitors, ip)
		}
	}
}

// Allow prueft ob eine IP eine Anfrage machen darf
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	v, exists := rl.visitors[ip]

	if !exists || now.Sub(v.windowStart) > rl.window {
		rl.visitors[ip] = &visitor{count: 1, windowStart: now}
		return true
	}

	v.count++
	return v.count <= rl.rate
}

// RateLimitMiddleware gibt 429 zurueck wenn zu viele Anfragen kommen
func RateLimitMiddleware(rl *RateLimiter) func(nethttp.Handler) nethttp.Handler {
	return func(next nethttp.Handler) nethttp.Handler {
		return nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
			ip := extractIP(r)
			if !rl.Allow(ip) {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Retry-After", "60")
				w.WriteHeader(nethttp.StatusTooManyRequests)
				json.NewEncoder(w).Encode(map[string]string{
					"error": "Zu viele Anfragen. Bitte versuche es spaeter erneut.",
				})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// BruteForceMiddleware prueft ob eine IP wegen zu vieler Fehlversuche gesperrt ist
func BruteForceMiddleware(sm *application.SessionManager, log logger.Logger) func(nethttp.Handler) nethttp.Handler {
	return func(next nethttp.Handler) nethttp.Handler {
		return nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
			ip := extractIP(r)

			blocked, err := sm.IsIPBlocked(r.Context(), ip)
			if err != nil {
				log.Error("brute-force check fehlgeschlagen", err)
				// Bei Fehler blockieren (fail-closed) — Redis nicht erreichbar = kein Login
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(nethttp.StatusServiceUnavailable)
				json.NewEncoder(w).Encode(map[string]string{
					"error": "Dienst voruebergehend nicht verfuegbar. Bitte versuche es spaeter erneut.",
				})
				return
			}

			if blocked {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(nethttp.StatusTooManyRequests)
				json.NewEncoder(w).Encode(map[string]string{
					"error": "Zu viele fehlgeschlagene Anmeldeversuche. Bitte warte 15 Minuten.",
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// extractIP extrahiert die Client-IP aus dem Request
func extractIP(r *nethttp.Request) string {
	// X-Forwarded-For Header (hinter Reverse Proxy / Cloudflare Tunnel)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}

	// X-Real-IP Header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// CF-Connecting-IP (Cloudflare)
	if cfip := r.Header.Get("CF-Connecting-IP"); cfip != "" {
		return cfip
	}

	// Fallback: RemoteAddr
	parts := strings.Split(r.RemoteAddr, ":")
	if len(parts) > 0 {
		return parts[0]
	}

	return r.RemoteAddr
}

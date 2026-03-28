package health

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// Checker verifies service dependencies are reachable.
type Checker interface {
	Check(ctx context.Context) Status
}

// Status represents the overall health state.
type Status struct {
	Status string `json:"status"`
	DB     string `json:"db"`
	Redis  string `json:"redis"`
}

// Handler returns an http.HandlerFunc that checks DB and Redis health.
func Handler(pool *pgxpool.Pool, redisClient *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		status := Status{
			Status: "ok",
			DB:     "connected",
			Redis:  "connected",
		}
		httpStatus := http.StatusOK

		// Check PostgreSQL
		if pool != nil {
			if err := pool.Ping(ctx); err != nil {
				status.DB = "disconnected"
				status.Status = "degraded"
				httpStatus = http.StatusServiceUnavailable
			}
		} else {
			status.DB = "not_configured"
		}

		// Check Redis
		if redisClient != nil {
			if err := redisClient.Ping(ctx).Err(); err != nil {
				status.Redis = "disconnected"
				status.Status = "degraded"
				httpStatus = http.StatusServiceUnavailable
			}
		} else {
			status.Redis = "not_configured"
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(httpStatus)
		_ = json.NewEncoder(w).Encode(status)
	}
}

// LivenessHandler returns a simple 200 OK for Kubernetes liveness probes.
func LivenessHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "alive"})
	}
}

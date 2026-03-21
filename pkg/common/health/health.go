package health

import (
	"sync"
	"time"
)

type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusUnhealthy Status = "unhealthy"
)

type HealthCheck interface {
	Check() (Status, string)
	Name() string
}

type HealthChecker struct {
	checks map[string]HealthCheck
	mu     sync.RWMutex
}

type HealthResponse struct {
	Status    Status                 `json:"status"`
	Checks    map[string]CheckResult `json:"checks"`
	Timestamp time.Time              `json:"timestamp"`
}

type CheckResult struct {
	Status  Status `json:"status"`
	Message string `json:"message"`
}

// NewHealthChecker creates a new health checker
func NewHealthChecker() *HealthChecker {
	return &HealthChecker{
		checks: make(map[string]HealthCheck),
	}
}

// Register adds a health check
func (h *HealthChecker) Register(check HealthCheck) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.checks[check.Name()] = check
}

// Check runs all registered health checks
func (h *HealthChecker) Check() HealthResponse {
	h.mu.RLock()
	defer h.mu.RUnlock()

	response := HealthResponse{
		Status:    StatusHealthy,
		Checks:    make(map[string]CheckResult),
		Timestamp: time.Now(),
	}

	for name, check := range h.checks {
		status, message := check.Check()
		response.Checks[name] = CheckResult{
			Status:  status,
			Message: message,
		}
		if status == StatusUnhealthy {
			response.Status = StatusUnhealthy
		}
	}

	return response
}

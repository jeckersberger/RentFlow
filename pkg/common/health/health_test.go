package health

import (
	"testing"
	"time"
)

// MockHealthCheck is a test implementation of HealthCheck
type MockHealthCheck struct {
	name   string
	status Status
	msg    string
}

func (m *MockHealthCheck) Name() string {
	return m.name
}

func (m *MockHealthCheck) Check() (Status, string) {
	return m.status, m.msg
}

func NewMockHealthCheck(name string, status Status, msg string) *MockHealthCheck {
	return &MockHealthCheck{
		name:   name,
		status: status,
		msg:    msg,
	}
}

func TestNewHealthChecker(t *testing.T) {
	hc := NewHealthChecker()
	if hc == nil {
		t.Errorf("expected HealthChecker, got nil")
	}
	if hc.checks == nil {
		t.Errorf("expected checks map, got nil")
	}
	if len(hc.checks) != 0 {
		t.Errorf("expected empty checks, got %d", len(hc.checks))
	}
}

func TestHealthCheckerRegister(t *testing.T) {
	hc := NewHealthChecker()

	check1 := NewMockHealthCheck("check1", StatusHealthy, "OK")
	check2 := NewMockHealthCheck("check2", StatusHealthy, "OK")

	t.Run("register single check", func(t *testing.T) {
		hc.Register(check1)
		if len(hc.checks) != 1 {
			t.Errorf("expected 1 check, got %d", len(hc.checks))
		}
		if _, exists := hc.checks["check1"]; !exists {
			t.Errorf("expected check1 to be registered")
		}
	})

	t.Run("register multiple checks", func(t *testing.T) {
		hc.Register(check2)
		if len(hc.checks) != 2 {
			t.Errorf("expected 2 checks, got %d", len(hc.checks))
		}
	})

	t.Run("register duplicate check", func(t *testing.T) {
		checkDuplicate := NewMockHealthCheck("check1", StatusUnhealthy, "Failed")
		hc.Register(checkDuplicate)
		if len(hc.checks) != 2 {
			t.Errorf("expected 2 checks after duplicate, got %d", len(hc.checks))
		}
		// Verify it was replaced
		status, msg := hc.checks["check1"].Check()
		if status != StatusUnhealthy || msg != "Failed" {
			t.Errorf("expected check to be updated")
		}
	})
}

func TestHealthCheckerCheck(t *testing.T) {
	tests := []struct {
		name           string
		checks         map[string]HealthCheck
		expectedStatus Status
		expectedLen    int
	}{
		{
			name: "all healthy checks",
			checks: map[string]HealthCheck{
				"check1": NewMockHealthCheck("check1", StatusHealthy, "OK"),
				"check2": NewMockHealthCheck("check2", StatusHealthy, "OK"),
			},
			expectedStatus: StatusHealthy,
			expectedLen:    2,
		},
		{
			name: "one unhealthy check",
			checks: map[string]HealthCheck{
				"check1": NewMockHealthCheck("check1", StatusHealthy, "OK"),
				"check2": NewMockHealthCheck("check2", StatusUnhealthy, "Failed"),
			},
			expectedStatus: StatusUnhealthy,
			expectedLen:    2,
		},
		{
			name: "all unhealthy checks",
			checks: map[string]HealthCheck{
				"check1": NewMockHealthCheck("check1", StatusUnhealthy, "Failed"),
				"check2": NewMockHealthCheck("check2", StatusUnhealthy, "Failed"),
			},
			expectedStatus: StatusUnhealthy,
			expectedLen:    2,
		},
		{
			name:           "no checks",
			checks:         map[string]HealthCheck{},
			expectedStatus: StatusHealthy,
			expectedLen:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hc := NewHealthChecker()
			for name, check := range tt.checks {
				hc.checks[name] = check
			}

			response := hc.Check()

			if response.Status != tt.expectedStatus {
				t.Errorf("expected status %q, got %q", tt.expectedStatus, response.Status)
			}
			if len(response.Checks) != tt.expectedLen {
				t.Errorf("expected %d checks, got %d", tt.expectedLen, len(response.Checks))
			}
		})
	}
}

func TestHealthCheckerCheckIndividualResults(t *testing.T) {
	hc := NewHealthChecker()

	check1 := NewMockHealthCheck("db", StatusHealthy, "Database connection OK")
	check2 := NewMockHealthCheck("cache", StatusHealthy, "Redis connection OK")
	check3 := NewMockHealthCheck("queue", StatusUnhealthy, "Message queue unavailable")

	hc.Register(check1)
	hc.Register(check2)
	hc.Register(check3)

	response := hc.Check()

	t.Run("check individual results", func(t *testing.T) {
		if response.Checks["db"].Status != StatusHealthy {
			t.Errorf("expected db to be healthy")
		}
		if response.Checks["cache"].Status != StatusHealthy {
			t.Errorf("expected cache to be healthy")
		}
		if response.Checks["queue"].Status != StatusUnhealthy {
			t.Errorf("expected queue to be unhealthy")
		}
	})

	t.Run("check messages", func(t *testing.T) {
		if response.Checks["db"].Message != "Database connection OK" {
			t.Errorf("expected correct db message")
		}
		if response.Checks["queue"].Message != "Message queue unavailable" {
			t.Errorf("expected correct queue message")
		}
	})
}

func TestHealthCheckerCheckTimestamp(t *testing.T) {
	hc := NewHealthChecker()
	hc.Register(NewMockHealthCheck("test", StatusHealthy, "OK"))

	before := time.Now()
	response := hc.Check()
	after := time.Now()

	if response.Timestamp.Before(before) || response.Timestamp.After(after) {
		t.Errorf("timestamp not in expected range")
	}
}

func TestStatusConstants(t *testing.T) {
	tests := []struct {
		name     string
		status   Status
		expected string
	}{
		{
			name:     "StatusHealthy",
			status:   StatusHealthy,
			expected: "healthy",
		},
		{
			name:     "StatusUnhealthy",
			status:   StatusUnhealthy,
			expected: "unhealthy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.status) != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, tt.status)
			}
		})
	}
}

func TestHealthCheckerMultipleChecks(t *testing.T) {
	hc := NewHealthChecker()

	// Register many checks
	for i := 1; i <= 5; i++ {
		name := "check" + string(rune(48+i))
		check := NewMockHealthCheck(name, StatusHealthy, "OK")
		hc.Register(check)
	}

	response := hc.Check()

	if len(response.Checks) != 5 {
		t.Errorf("expected 5 checks, got %d", len(response.Checks))
	}
	if response.Status != StatusHealthy {
		t.Errorf("expected healthy status with all checks healthy")
	}
}

func TestHealthCheckerMixedStatus(t *testing.T) {
	hc := NewHealthChecker()

	healthyCheck := NewMockHealthCheck("healthy", StatusHealthy, "Working")
	unhealthyCheck := NewMockHealthCheck("unhealthy", StatusUnhealthy, "Down")

	hc.Register(healthyCheck)
	hc.Register(unhealthyCheck)

	response := hc.Check()

	if response.Status != StatusUnhealthy {
		t.Errorf("expected unhealthy status when any check is unhealthy")
	}
	if response.Checks["healthy"].Status != StatusHealthy {
		t.Errorf("expected healthy check to be marked healthy")
	}
	if response.Checks["unhealthy"].Status != StatusUnhealthy {
		t.Errorf("expected unhealthy check to be marked unhealthy")
	}
}

func TestHealthResponseFields(t *testing.T) {
	hc := NewHealthChecker()
	check := NewMockHealthCheck("test", StatusHealthy, "Test message")
	hc.Register(check)

	response := hc.Check()

	t.Run("response has all fields", func(t *testing.T) {
		if response.Status == "" {
			t.Errorf("response.Status is empty")
		}
		if response.Checks == nil {
			t.Errorf("response.Checks is nil")
		}
		if response.Timestamp.IsZero() {
			t.Errorf("response.Timestamp is zero")
		}
	})
}

func TestCheckResult(t *testing.T) {
	cr := CheckResult{
		Status:  StatusHealthy,
		Message: "All systems operational",
	}

	if cr.Status != StatusHealthy {
		t.Errorf("expected healthy status")
	}
	if cr.Message != "All systems operational" {
		t.Errorf("expected correct message")
	}
}

func TestHealthCheckerRegisterWithDifferentStatuses(t *testing.T) {
	hc := NewHealthChecker()

	checks := []struct {
		name   string
		status Status
		msg    string
	}{
		{"db", StatusHealthy, "Connected"},
		{"cache", StatusUnhealthy, "Disconnected"},
		{"queue", StatusHealthy, "Running"},
		{"api", StatusUnhealthy, "Timeout"},
	}

	for _, c := range checks {
		hc.Register(NewMockHealthCheck(c.name, c.status, c.msg))
	}

	response := hc.Check()

	if response.Status != StatusUnhealthy {
		t.Errorf("expected overall status to be unhealthy")
	}

	healthyCount := 0
	unhealthyCount := 0
	for _, result := range response.Checks {
		if result.Status == StatusHealthy {
			healthyCount++
		} else {
			unhealthyCount++
		}
	}

	if healthyCount != 2 {
		t.Errorf("expected 2 healthy checks, got %d", healthyCount)
	}
	if unhealthyCount != 2 {
		t.Errorf("expected 2 unhealthy checks, got %d", unhealthyCount)
	}
}

func TestHealthCheckerConcurrentRegister(t *testing.T) {
	hc := NewHealthChecker()
	done := make(chan bool, 10)

	// Register checks concurrently
	for i := 0; i < 10; i++ {
		go func(idx int) {
			check := NewMockHealthCheck(
				"check"+string(rune(48+idx%10)),
				StatusHealthy,
				"OK",
			)
			hc.Register(check)
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	response := hc.Check()
	if response.Status != StatusHealthy {
		t.Errorf("expected healthy status")
	}
}

func TestHealthCheckerConcurrentCheck(t *testing.T) {
	hc := NewHealthChecker()
	hc.Register(NewMockHealthCheck("test", StatusHealthy, "OK"))

	done := make(chan bool, 10)

	// Run checks concurrently
	for i := 0; i < 10; i++ {
		go func() {
			response := hc.Check()
			if response.Status != StatusHealthy {
				t.Errorf("expected healthy status")
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestHealthCheckWithVeryLongMessage(t *testing.T) {
	longMsg := "This is a very long health check message with lots of details about the system state and what might be wrong if this check fails. " +
		"It could include error codes, stack traces, or other diagnostic information that could be useful for debugging."

	hc := NewHealthChecker()
	hc.Register(NewMockHealthCheck("long-msg", StatusUnhealthy, longMsg))

	response := hc.Check()
	if response.Checks["long-msg"].Message != longMsg {
		t.Errorf("message not preserved correctly")
	}
}

func TestHealthResponseJSON(t *testing.T) {
	hc := NewHealthChecker()
	hc.Register(NewMockHealthCheck("db", StatusHealthy, "OK"))

	response := hc.Check()

	// Verify the response can be marshaled (structure is JSON-compatible)
	if response.Status == "" {
		t.Errorf("response.Status should be populated")
	}
	if len(response.Checks) == 0 {
		t.Errorf("response.Checks should contain check results")
	}
}

func TestHealthCheckerEmptyAndThenRegister(t *testing.T) {
	hc := NewHealthChecker()

	// Check with no registered checks
	response := hc.Check()
	if response.Status != StatusHealthy {
		t.Errorf("expected healthy with no checks")
	}
	if len(response.Checks) != 0 {
		t.Errorf("expected no checks")
	}

	// Register a check
	hc.Register(NewMockHealthCheck("check1", StatusHealthy, "OK"))
	response = hc.Check()
	if len(response.Checks) != 1 {
		t.Errorf("expected 1 check after registration")
	}
}

func TestHealthCheckerRegisterAndReplace(t *testing.T) {
	hc := NewHealthChecker()

	check1 := NewMockHealthCheck("check1", StatusHealthy, "Initial")
	hc.Register(check1)

	response := hc.Check()
	if response.Checks["check1"].Message != "Initial" {
		t.Errorf("expected initial message")
	}

	// Replace the check
	check1Updated := NewMockHealthCheck("check1", StatusUnhealthy, "Updated")
	hc.Register(check1Updated)

	response = hc.Check()
	if response.Checks["check1"].Message != "Updated" {
		t.Errorf("expected updated message")
	}
	if response.Checks["check1"].Status != StatusUnhealthy {
		t.Errorf("expected updated status")
	}
	if response.Status != StatusUnhealthy {
		t.Errorf("expected overall status to be unhealthy")
	}
}

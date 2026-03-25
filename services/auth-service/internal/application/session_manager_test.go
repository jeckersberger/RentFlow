package application

import (
	"context"
	"testing"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/cache"
	"github.com/jeckersberger/rentflow/pkg/common/logger"
)

// ---------------------------------------------------------------------------
// Helper: create a SessionManager backed by InMemoryCache
// ---------------------------------------------------------------------------

func newTestSessionManager() *SessionManager {
	c := cache.NewInMemoryCache()
	log := logger.New("error", "test") // quiet logger for tests
	return NewSessionManager(c, log)
}

// ---------------------------------------------------------------------------
// Session CRUD Tests
// ---------------------------------------------------------------------------

func TestCreateSession_Success(t *testing.T) {
	sm := newTestSessionManager()
	ctx := context.Background()

	session, err := sm.CreateSession(ctx, "user-1", "tenant-1", []string{"admin"}, "127.0.0.1", "TestAgent")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if session == nil {
		t.Fatal("expected session, got nil")
	}
	if session.ID == "" {
		t.Error("expected non-empty session ID")
	}
	if session.UserID != "user-1" {
		t.Errorf("expected UserID 'user-1', got '%s'", session.UserID)
	}
	if session.TenantID != "tenant-1" {
		t.Errorf("expected TenantID 'tenant-1', got '%s'", session.TenantID)
	}
	if len(session.Roles) != 1 || session.Roles[0] != "admin" {
		t.Errorf("expected roles [admin], got %v", session.Roles)
	}
	if session.IPAddress != "127.0.0.1" {
		t.Errorf("expected IP '127.0.0.1', got '%s'", session.IPAddress)
	}
	if session.Rotations != 0 {
		t.Errorf("expected 0 rotations, got %d", session.Rotations)
	}
}

func TestCreateSession_MissingUserID(t *testing.T) {
	sm := newTestSessionManager()
	ctx := context.Background()

	_, err := sm.CreateSession(ctx, "", "tenant-1", []string{"admin"}, "127.0.0.1", "TestAgent")
	if err == nil {
		t.Fatal("expected error for empty userID")
	}
}

func TestCreateSession_MissingTenantID(t *testing.T) {
	sm := newTestSessionManager()
	ctx := context.Background()

	_, err := sm.CreateSession(ctx, "user-1", "", []string{"admin"}, "127.0.0.1", "TestAgent")
	if err == nil {
		t.Fatal("expected error for empty tenantID")
	}
}

func TestGetSession_Success(t *testing.T) {
	sm := newTestSessionManager()
	ctx := context.Background()

	created, _ := sm.CreateSession(ctx, "user-1", "tenant-1", []string{"admin"}, "127.0.0.1", "TestAgent")

	fetched, err := sm.GetSession(ctx, created.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if fetched.UserID != "user-1" {
		t.Errorf("expected UserID 'user-1', got '%s'", fetched.UserID)
	}
	if fetched.ID != created.ID {
		t.Errorf("session IDs don't match: '%s' vs '%s'", created.ID, fetched.ID)
	}
}

func TestGetSession_NotFound(t *testing.T) {
	sm := newTestSessionManager()
	ctx := context.Background()

	_, err := sm.GetSession(ctx, "nonexistent-session")
	if err == nil {
		t.Fatal("expected error for nonexistent session")
	}
}

func TestGetSession_EmptyID(t *testing.T) {
	sm := newTestSessionManager()
	ctx := context.Background()

	_, err := sm.GetSession(ctx, "")
	if err == nil {
		t.Fatal("expected error for empty session ID")
	}
}

// ---------------------------------------------------------------------------
// Session Refresh / Rotation Tests
// ---------------------------------------------------------------------------

func TestRefreshSession_IncrementsRotation(t *testing.T) {
	sm := newTestSessionManager()
	ctx := context.Background()

	session, _ := sm.CreateSession(ctx, "user-1", "tenant-1", []string{"admin"}, "127.0.0.1", "TestAgent")

	err := sm.RefreshSession(ctx, session.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	updated, _ := sm.GetSession(ctx, session.ID)
	if updated.Rotations != 1 {
		t.Errorf("expected 1 rotation, got %d", updated.Rotations)
	}
}

func TestRefreshSession_MaxRotationsReached(t *testing.T) {
	sm := newTestSessionManager()
	ctx := context.Background()

	session, _ := sm.CreateSession(ctx, "user-1", "tenant-1", []string{"admin"}, "127.0.0.1", "TestAgent")

	// Rotate maxRotations times
	for i := 0; i < maxRotations; i++ {
		err := sm.RefreshSession(ctx, session.ID)
		if err != nil {
			t.Fatalf("rotation %d failed: %v", i+1, err)
		}
	}

	// Next rotation should fail
	err := sm.RefreshSession(ctx, session.ID)
	if err == nil {
		t.Fatal("expected error when max rotations exceeded")
	}
}

// ---------------------------------------------------------------------------
// Session Invalidation Tests
// ---------------------------------------------------------------------------

func TestInvalidateSession(t *testing.T) {
	sm := newTestSessionManager()
	ctx := context.Background()

	session, _ := sm.CreateSession(ctx, "user-1", "tenant-1", []string{"admin"}, "127.0.0.1", "TestAgent")

	err := sm.InvalidateSession(ctx, session.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Session should no longer be retrievable
	_, err = sm.GetSession(ctx, session.ID)
	if err == nil {
		t.Fatal("expected error after invalidation")
	}
}

func TestInvalidateAllUserSessions(t *testing.T) {
	sm := newTestSessionManager()
	ctx := context.Background()

	s1, _ := sm.CreateSession(ctx, "user-1", "tenant-1", []string{"admin"}, "127.0.0.1", "Agent1")
	s2, _ := sm.CreateSession(ctx, "user-1", "tenant-1", []string{"admin"}, "192.168.0.1", "Agent2")

	err := sm.InvalidateAllUserSessions(ctx, "user-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Both sessions should be gone
	_, err1 := sm.GetSession(ctx, s1.ID)
	_, err2 := sm.GetSession(ctx, s2.ID)
	if err1 == nil || err2 == nil {
		t.Fatal("expected both sessions to be invalidated")
	}
}

func TestInvalidateAllUserSessions_NoSessions(t *testing.T) {
	sm := newTestSessionManager()
	ctx := context.Background()

	// Should not error even if user has no sessions
	err := sm.InvalidateAllUserSessions(ctx, "user-no-sessions")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// Brute-Force / Failed Login Tests
// ---------------------------------------------------------------------------

func TestRecordFailedLogin_IncrementsCounter(t *testing.T) {
	sm := newTestSessionManager()
	ctx := context.Background()

	count, err := sm.RecordFailedLogin(ctx, "10.0.0.1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if count != 1 {
		t.Errorf("expected count 1, got %d", count)
	}

	count, err = sm.RecordFailedLogin(ctx, "10.0.0.1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if count != 2 {
		t.Errorf("expected count 2, got %d", count)
	}
}

func TestRecordFailedLogin_EmptyIP(t *testing.T) {
	sm := newTestSessionManager()
	ctx := context.Background()

	_, err := sm.RecordFailedLogin(ctx, "")
	if err == nil {
		t.Fatal("expected error for empty IP")
	}
}

func TestGetFailedLoginCount_NoEntries(t *testing.T) {
	sm := newTestSessionManager()
	ctx := context.Background()

	count, err := sm.GetFailedLoginCount(ctx, "10.0.0.2")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0, got %d", count)
	}
}

func TestIsIPBlocked_BelowThreshold(t *testing.T) {
	sm := newTestSessionManager()
	ctx := context.Background()

	// Record fewer than maxFailedLogins attempts
	for i := 0; i < maxFailedLogins-1; i++ {
		sm.RecordFailedLogin(ctx, "10.0.0.3")
	}

	blocked, err := sm.IsIPBlocked(ctx, "10.0.0.3")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if blocked {
		t.Error("expected IP not to be blocked yet")
	}
}

func TestIsIPBlocked_AtThreshold(t *testing.T) {
	sm := newTestSessionManager()
	ctx := context.Background()

	// Record exactly maxFailedLogins attempts
	for i := 0; i < maxFailedLogins; i++ {
		sm.RecordFailedLogin(ctx, "10.0.0.4")
	}

	blocked, err := sm.IsIPBlocked(ctx, "10.0.0.4")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !blocked {
		t.Error("expected IP to be blocked at threshold")
	}
}

func TestIsIPBlocked_AboveThreshold(t *testing.T) {
	sm := newTestSessionManager()
	ctx := context.Background()

	for i := 0; i < maxFailedLogins+5; i++ {
		sm.RecordFailedLogin(ctx, "10.0.0.5")
	}

	blocked, err := sm.IsIPBlocked(ctx, "10.0.0.5")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !blocked {
		t.Error("expected IP to be blocked above threshold")
	}
}

func TestResetFailedLogins(t *testing.T) {
	sm := newTestSessionManager()
	ctx := context.Background()

	// Build up failed logins
	for i := 0; i < maxFailedLogins; i++ {
		sm.RecordFailedLogin(ctx, "10.0.0.6")
	}

	// Verify blocked
	blocked, _ := sm.IsIPBlocked(ctx, "10.0.0.6")
	if !blocked {
		t.Fatal("expected IP to be blocked before reset")
	}

	// Reset
	err := sm.ResetFailedLogins(ctx, "10.0.0.6")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify unblocked
	blocked, _ = sm.IsIPBlocked(ctx, "10.0.0.6")
	if blocked {
		t.Error("expected IP to be unblocked after reset")
	}

	count, _ := sm.GetFailedLoginCount(ctx, "10.0.0.6")
	if count != 0 {
		t.Errorf("expected 0 failed logins after reset, got %d", count)
	}
}

func TestIsIPBlocked_DifferentIPsIndependent(t *testing.T) {
	sm := newTestSessionManager()
	ctx := context.Background()

	// Block one IP
	for i := 0; i < maxFailedLogins; i++ {
		sm.RecordFailedLogin(ctx, "10.0.0.7")
	}

	// Other IP should not be blocked
	blocked, _ := sm.IsIPBlocked(ctx, "10.0.0.8")
	if blocked {
		t.Error("expected different IP to not be blocked")
	}
}

// ---------------------------------------------------------------------------
// bruteforceTTLForCount Tests
// ---------------------------------------------------------------------------

func TestBruteforceTTLForCount(t *testing.T) {
	tests := []struct {
		count    int
		expected time.Duration
	}{
		{0, 1 * time.Minute},
		{1, 1 * time.Minute},
		{5, 1 * time.Minute},
		{9, 1 * time.Minute},
		{10, 5 * time.Minute},
		{14, 5 * time.Minute},
		{15, 15 * time.Minute},
		{19, 15 * time.Minute},
		{20, 30 * time.Minute},
		{24, 30 * time.Minute},
		{25, 1 * time.Hour},
		{50, 1 * time.Hour},
		{100, 1 * time.Hour},
	}

	for _, tt := range tests {
		result := bruteforceTTLForCount(tt.count)
		if result != tt.expected {
			t.Errorf("bruteforceTTLForCount(%d) = %v, expected %v", tt.count, result, tt.expected)
		}
	}
}

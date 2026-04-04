package domain

import (
	"testing"
)

// ---------------------------------------------------------------------------
// ValidAction
// ---------------------------------------------------------------------------

func TestValidAction(t *testing.T) {
	valid := []string{ActionScan, ActionCheckout, ActionCheckin, ActionInventoryScan, ActionAdhocBooking}
	for _, a := range valid {
		t.Run("valid_"+a, func(t *testing.T) {
			if !ValidAction(a) {
				t.Errorf("ValidAction(%q) = false, want true", a)
			}
		})
	}

	invalid := []string{"", "unknown", "SCAN", "Checkout", "delete", "move"}
	for _, a := range invalid {
		t.Run("invalid_"+a, func(t *testing.T) {
			if ValidAction(a) {
				t.Errorf("ValidAction(%q) = true, want false", a)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Action constants
// ---------------------------------------------------------------------------

func TestActionConstants(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{"ActionScan", ActionScan, "scan"},
		{"ActionCheckout", ActionCheckout, "checkout"},
		{"ActionCheckin", ActionCheckin, "checkin"},
		{"ActionInventoryScan", ActionInventoryScan, "inventory_scan"},
		{"ActionAdhocBooking", ActionAdhocBooking, "adhoc_booking"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("%s = %q, want %q", tt.name, tt.got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// ValidSessionType
// ---------------------------------------------------------------------------

func TestValidSessionType(t *testing.T) {
	valid := []string{SessionTypeCheckout, SessionTypeCheckin, SessionTypeInventory}
	for _, s := range valid {
		t.Run("valid_"+s, func(t *testing.T) {
			if !ValidSessionType(s) {
				t.Errorf("ValidSessionType(%q) = false, want true", s)
			}
		})
	}

	invalid := []string{"", "scan", "unknown", "CHECKOUT"}
	for _, s := range invalid {
		t.Run("invalid_"+s, func(t *testing.T) {
			if ValidSessionType(s) {
				t.Errorf("ValidSessionType(%q) = true, want false", s)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Session status constants
// ---------------------------------------------------------------------------

func TestSessionStatusConstants(t *testing.T) {
	if SessionStatusActive != "active" {
		t.Errorf("SessionStatusActive = %q, want %q", SessionStatusActive, "active")
	}
	if SessionStatusCompleted != "completed" {
		t.Errorf("SessionStatusCompleted = %q, want %q", SessionStatusCompleted, "completed")
	}
}

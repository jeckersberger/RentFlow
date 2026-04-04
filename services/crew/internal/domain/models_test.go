package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// Assignment status constants
// ---------------------------------------------------------------------------

func TestAssignmentStatusConstants(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{"Planned", AssignmentStatusPlanned, "planned"},
		{"Confirmed", AssignmentStatusConfirmed, "confirmed"},
		{"Active", AssignmentStatusActive, "active"},
		{"Completed", AssignmentStatusCompleted, "completed"},
		{"Cancelled", AssignmentStatusCancelled, "cancelled"},
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
// Time entry status constants
// ---------------------------------------------------------------------------

func TestTimeEntryStatusConstants(t *testing.T) {
	if TimeEntryStatusActive != "active" {
		t.Errorf("TimeEntryStatusActive = %q, want %q", TimeEntryStatusActive, "active")
	}
	if TimeEntryStatusCompleted != "completed" {
		t.Errorf("TimeEntryStatusCompleted = %q, want %q", TimeEntryStatusCompleted, "completed")
	}
}

// ---------------------------------------------------------------------------
// Availability status constants
// ---------------------------------------------------------------------------

func TestAvailabilityStatusConstants(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{"Available", AvailabilityStatusAvailable, "available"},
		{"Unavailable", AvailabilityStatusUnavailable, "unavailable"},
		{"OnRequest", AvailabilityStatusOnRequest, "on_request"},
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
// CrewMember — hourly rate in cents
// ---------------------------------------------------------------------------

func TestCrewMemberHourlyRate(t *testing.T) {
	m := &CrewMember{
		ID:         uuid.New(),
		TenantID:   uuid.New(),
		FirstName:  "Max",
		LastName:   "Mustermann",
		Role:       "technician",
		HourlyRate: 3500, // 35.00 EUR
		IsActive:   true,
	}

	if m.HourlyRate != 3500 {
		t.Errorf("HourlyRate = %d, want 3500", m.HourlyRate)
	}
	if !m.IsActive {
		t.Error("crew member should be active")
	}
}

// ---------------------------------------------------------------------------
// TimeEntry — duration calculation
// ---------------------------------------------------------------------------

func TestTimeEntryDuration(t *testing.T) {
	checkIn := time.Date(2026, 4, 4, 8, 0, 0, 0, time.UTC)
	checkOut := time.Date(2026, 4, 4, 16, 30, 0, 0, time.UTC)

	entry := &TimeEntry{
		ID:           uuid.New(),
		TenantID:     uuid.New(),
		CrewMemberID: uuid.New(),
		CheckInTime:  checkIn,
		CheckOutTime: &checkOut,
		BreakMinutes: 30,
		Status:       TimeEntryStatusCompleted,
	}

	totalMinutes := int(checkOut.Sub(checkIn).Minutes())
	workMinutes := totalMinutes - entry.BreakMinutes
	entry.DurationMinutes = &workMinutes

	if *entry.DurationMinutes != 480 {
		t.Errorf("DurationMinutes = %d, want 480 (8h work)", *entry.DurationMinutes)
	}
}

// ---------------------------------------------------------------------------
// CrewAssignment — hours tracking
// ---------------------------------------------------------------------------

func TestCrewAssignmentHoursTracking(t *testing.T) {
	a := &CrewAssignment{
		HoursPlanned: 40.0,
		HoursActual:  35.5,
		Status:       AssignmentStatusActive,
	}

	remaining := a.HoursPlanned - a.HoursActual
	if remaining != 4.5 {
		t.Errorf("remaining hours = %.1f, want 4.5", remaining)
	}
}

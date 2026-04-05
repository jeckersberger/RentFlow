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
// Availability block type constants
// ---------------------------------------------------------------------------

func TestBlockTypeConstants(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{"Vacation", BlockTypeVacation, "vacation"},
		{"Sick", BlockTypeSick, "sick"},
		{"Training", BlockTypeTraining, "training"},
		{"Blocked", BlockTypeBlocked, "blocked"},
		{"Other", BlockTypeOther, "other"},
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
// AvailabilityBlock — date range
// ---------------------------------------------------------------------------

func TestAvailabilityBlockDateRange(t *testing.T) {
	block := &AvailabilityBlock{
		ID:           uuid.New(),
		TenantID:     uuid.New(),
		CrewMemberID: uuid.New(),
		BlockType:    BlockTypeVacation,
		StartDate:    "2026-04-10",
		EndDate:      "2026-04-17",
		Notes:        "Easter vacation",
	}

	if block.BlockType != "vacation" {
		t.Errorf("BlockType = %q, want %q", block.BlockType, "vacation")
	}
	if block.StartDate >= block.EndDate {
		t.Error("start_date should be before end_date")
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

// ---------------------------------------------------------------------------
// CrewMember — skills field
// ---------------------------------------------------------------------------

func TestCrewMemberSkills(t *testing.T) {
	m := &CrewMember{
		ID:       uuid.New(),
		TenantID: uuid.New(),
		FirstName: "Lisa",
		LastName:  "Schmidt",
		Role:      "technician",
		Skills:    []string{"Tontechniker", "Rigger", "Lichtdesigner"},
		IsActive:  true,
	}

	if len(m.Skills) != 3 {
		t.Errorf("Skills length = %d, want 3", len(m.Skills))
	}
	if m.Skills[0] != "Tontechniker" {
		t.Errorf("Skills[0] = %q, want %q", m.Skills[0], "Tontechniker")
	}
}

func TestCrewMemberSkillsEmpty(t *testing.T) {
	m := &CrewMember{
		ID:       uuid.New(),
		TenantID: uuid.New(),
		FirstName: "Tom",
		LastName:  "Weber",
		Skills:    []string{},
	}

	if m.Skills == nil {
		t.Error("Skills should be empty slice, not nil")
	}
	if len(m.Skills) != 0 {
		t.Errorf("Skills length = %d, want 0", len(m.Skills))
	}
}

// ---------------------------------------------------------------------------
// SkillMatchResponse — structure
// ---------------------------------------------------------------------------

func TestSkillMatchResponseStructure(t *testing.T) {
	memberID := uuid.New()

	resp := &SkillMatchResponse{
		Matches: []SkillMatchResult{
			{
				Skill:    "Tontechniker",
				Required: 2,
				Matched: []SkillMatchCandidate{
					{CrewMemberID: memberID, Name: "Max Mueller", Score: 1.0, Available: true},
				},
			},
		},
		UnmatchedSkills: []string{"Pyrotechniker"},
	}

	if len(resp.Matches) != 1 {
		t.Fatalf("Matches length = %d, want 1", len(resp.Matches))
	}
	if resp.Matches[0].Skill != "Tontechniker" {
		t.Errorf("Matches[0].Skill = %q, want %q", resp.Matches[0].Skill, "Tontechniker")
	}
	if resp.Matches[0].Required != 2 {
		t.Errorf("Matches[0].Required = %d, want 2", resp.Matches[0].Required)
	}
	if len(resp.Matches[0].Matched) != 1 {
		t.Fatalf("Matches[0].Matched length = %d, want 1", len(resp.Matches[0].Matched))
	}
	if resp.Matches[0].Matched[0].Score != 1.0 {
		t.Errorf("Score = %f, want 1.0", resp.Matches[0].Matched[0].Score)
	}
	if !resp.Matches[0].Matched[0].Available {
		t.Error("Candidate should be available")
	}
	if len(resp.UnmatchedSkills) != 1 {
		t.Fatalf("UnmatchedSkills length = %d, want 1", len(resp.UnmatchedSkills))
	}
	if resp.UnmatchedSkills[0] != "Pyrotechniker" {
		t.Errorf("UnmatchedSkills[0] = %q, want %q", resp.UnmatchedSkills[0], "Pyrotechniker")
	}
}

// ---------------------------------------------------------------------------
// SkillMatchCandidate — scoring
// ---------------------------------------------------------------------------

func TestSkillMatchCandidateScoring(t *testing.T) {
	tests := []struct {
		name      string
		score     float64
		available bool
	}{
		{"ExactAvailable", 1.0, true},
		{"PartialAvailable", 0.8, true},
		{"ExactUnavailable", 0.0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := SkillMatchCandidate{
				CrewMemberID: uuid.New(),
				Name:         "Test Person",
				Score:        tt.score,
				Available:    tt.available,
			}
			if c.Score != tt.score {
				t.Errorf("Score = %f, want %f", c.Score, tt.score)
			}
			if c.Available != tt.available {
				t.Errorf("Available = %v, want %v", c.Available, tt.available)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Skill-matching error constants
// ---------------------------------------------------------------------------

func TestSkillMatchErrorConstants(t *testing.T) {
	if ErrNoSkillsRequested.Error() != "required_skills must not be empty" {
		t.Errorf("ErrNoSkillsRequested = %q", ErrNoSkillsRequested.Error())
	}
	if ErrInvalidSkillCount.Error() != "each required skill must have count >= 1" {
		t.Errorf("ErrInvalidSkillCount = %q", ErrInvalidSkillCount.Error())
	}
	if ErrDateFromRequired.Error() != "date_from is required (YYYY-MM-DD)" {
		t.Errorf("ErrDateFromRequired = %q", ErrDateFromRequired.Error())
	}
	if ErrDateToRequired.Error() != "date_to is required (YYYY-MM-DD)" {
		t.Errorf("ErrDateToRequired = %q", ErrDateToRequired.Error())
	}
}

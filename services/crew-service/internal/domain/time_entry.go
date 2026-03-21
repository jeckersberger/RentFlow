package domain

import (
	"fmt"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/events"
)

type TimeEntryType string
type TimeEntryStatus string

const (
	TimeTypWork    TimeEntryType = "work"
	TimeTypeTravel TimeEntryType = "travel"
	TimeTypeSetup  TimeEntryType = "setup"
	TimeTypeTeardown TimeEntryType = "teardown"
)

const (
	TimeStatusPending  TimeEntryStatus = "pending"
	TimeStatusApproved TimeEntryStatus = "approved"
	TimeStatusRejected TimeEntryStatus = "rejected"
)

type TimeEntry struct {
	events.AggregateRoot
	TenantID      string
	CrewMemberID  string
	ProjectID     string
	Date          time.Time
	StartTime     time.Time
	EndTime       time.Time
	BreakMinutes  int
	TotalHours    float64
	Type          TimeEntryType
	Notes         string
	Status        TimeEntryStatus
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func NewTimeEntry(id, tenantID, crewMemberID, projectID string, date time.Time, startTime, endTime time.Time, entryType TimeEntryType) *TimeEntry {
	now := time.Now()
	totalHours := calculateHours(startTime, endTime, 0)

	return &TimeEntry{
		AggregateRoot: *events.NewAggregateRoot(id, "time_entry"),
		TenantID:      tenantID,
		CrewMemberID:  crewMemberID,
		ProjectID:     projectID,
		Date:          date,
		StartTime:     startTime,
		EndTime:       endTime,
		BreakMinutes:  0,
		TotalHours:    totalHours,
		Type:          entryType,
		Status:        TimeStatusPending,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

func (t *TimeEntry) SetBreak(minutes int) error {
	if minutes < 0 {
		return fmt.Errorf("break minutes cannot be negative")
	}
	t.BreakMinutes = minutes
	t.TotalHours = calculateHours(t.StartTime, t.EndTime, minutes)
	t.UpdatedAt = time.Now()
	return nil
}

func (t *TimeEntry) Approve() error {
	if t.Status != TimeStatusPending {
		return fmt.Errorf("can only approve pending entries")
	}
	t.Status = TimeStatusApproved
	t.UpdatedAt = time.Now()
	return nil
}

func (t *TimeEntry) Reject() error {
	if t.Status != TimeStatusPending {
		return fmt.Errorf("can only reject pending entries")
	}
	t.Status = TimeStatusRejected
	t.UpdatedAt = time.Now()
	return nil
}

func (t *TimeEntry) Validate() error {
	if t.ID == "" {
		return fmt.Errorf("time entry ID cannot be empty")
	}
	if t.TenantID == "" {
		return fmt.Errorf("tenant ID cannot be empty")
	}
	if t.CrewMemberID == "" {
		return fmt.Errorf("crew member ID cannot be empty")
	}
	if t.StartTime.After(t.EndTime) {
		return fmt.Errorf("start time must be before end time")
	}
	return nil
}

func calculateHours(startTime, endTime time.Time, breakMinutes int) float64 {
	duration := endTime.Sub(startTime)
	hours := duration.Hours()
	breakHours := float64(breakMinutes) / 60.0
	return hours - breakHours
}

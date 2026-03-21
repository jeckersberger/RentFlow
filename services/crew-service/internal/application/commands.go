package application

import "time"

type CreateCrewMemberCommand struct {
	TenantID string
	UserID   string
	FirstName string
	LastName string
	Email    string
	Phone    string
	Type     string
}

type UpdateCrewMemberCommand struct {
	ID       string
	TenantID string
	FirstName string
	LastName string
	Email    string
	Phone    string
	HourlyRate float64
	DailyRate float64
}

type CreateTimeEntryCommand struct {
	TenantID     string
	CrewMemberID string
	ProjectID    string
	Date         time.Time
	StartTime    time.Time
	EndTime      time.Time
	Type         string
	Notes        string
}

type UpdateTimeEntryCommand struct {
	ID           string
	TenantID     string
	BreakMinutes int
	Notes        string
}

type CreateAssignmentCommand struct {
	TenantID     string
	CrewMemberID string
	ProjectID    string
	Role         string
	StartDate    time.Time
	EndDate      time.Time
}

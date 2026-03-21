package application

import (
	"time"

	"github.com/jeckersberger/rentflow/services/crew-service/internal/domain"
)

type CrewMemberDTO struct {
	ID                   string    `json:"id"`
	TenantID             string    `json:"tenant_id"`
	UserID               string    `json:"user_id"`
	FirstName            string    `json:"first_name"`
	LastName             string    `json:"last_name"`
	Email                string    `json:"email"`
	Phone                string    `json:"phone"`
	Type                 string    `json:"type"`
	Skills               []string  `json:"skills"`
	HourlyRate           float64   `json:"hourly_rate"`
	DailyRate            float64   `json:"daily_rate"`
	Status               string    `json:"status"`
	AvailabilityCalendar string    `json:"availability_calendar"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

type TimeEntryDTO struct {
	ID           string    `json:"id"`
	TenantID     string    `json:"tenant_id"`
	CrewMemberID string    `json:"crew_member_id"`
	ProjectID    string    `json:"project_id"`
	Date         time.Time `json:"date"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
	BreakMinutes int       `json:"break_minutes"`
	TotalHours   float64   `json:"total_hours"`
	Type         string    `json:"type"`
	Notes        string    `json:"notes"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type AssignmentDTO struct {
	ID           string    `json:"id"`
	TenantID     string    `json:"tenant_id"`
	CrewMemberID string    `json:"crew_member_id"`
	ProjectID    string    `json:"project_id"`
	Role         string    `json:"role"`
	StartDate    time.Time `json:"start_date"`
	EndDate      time.Time `json:"end_date"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type TimeEntrySummary struct {
	CrewMemberID  string  `json:"crew_member_id"`
	TotalHours    float64 `json:"total_hours"`
	Entries       int     `json:"entries"`
	ApprovedHours float64 `json:"approved_hours"`
}

func CrewMemberToDTO(cm *domain.CrewMember) *CrewMemberDTO {
	return &CrewMemberDTO{
		ID:                   cm.ID,
		TenantID:             cm.TenantID,
		UserID:               cm.UserID,
		FirstName:            cm.FirstName,
		LastName:             cm.LastName,
		Email:                cm.Email,
		Phone:                cm.Phone,
		Type:                 string(cm.Type),
		Skills:               cm.Skills,
		HourlyRate:           cm.HourlyRate,
		DailyRate:            cm.DailyRate,
		Status:               string(cm.Status),
		AvailabilityCalendar: cm.AvailabilityCalendar,
		CreatedAt:            cm.CreatedAt,
		UpdatedAt:            cm.UpdatedAt,
	}
}

func TimeEntryToDTO(te *domain.TimeEntry) *TimeEntryDTO {
	return &TimeEntryDTO{
		ID:           te.ID,
		TenantID:     te.TenantID,
		CrewMemberID: te.CrewMemberID,
		ProjectID:    te.ProjectID,
		Date:         te.Date,
		StartTime:    te.StartTime,
		EndTime:      te.EndTime,
		BreakMinutes: te.BreakMinutes,
		TotalHours:   te.TotalHours,
		Type:         string(te.Type),
		Notes:        te.Notes,
		Status:       string(te.Status),
		CreatedAt:    te.CreatedAt,
		UpdatedAt:    te.UpdatedAt,
	}
}

func AssignmentToDTO(a *domain.Assignment) *AssignmentDTO {
	return &AssignmentDTO{
		ID:           a.ID,
		TenantID:     a.TenantID,
		CrewMemberID: a.CrewMemberID,
		ProjectID:    a.ProjectID,
		Role:         a.Role,
		StartDate:    a.StartDate,
		EndDate:      a.EndDate,
		Status:       string(a.Status),
		CreatedAt:    a.CreatedAt,
		UpdatedAt:    a.UpdatedAt,
	}
}

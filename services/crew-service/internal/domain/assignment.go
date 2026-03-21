package domain

import (
	"fmt"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/events"
)

type AssignmentStatus string

const (
	AssignmentStatusActive   AssignmentStatus = "active"
	AssignmentStatusCompleted AssignmentStatus = "completed"
	AssignmentStatusCancelled AssignmentStatus = "cancelled"
)

type Assignment struct {
	events.AggregateRoot
	TenantID     string
	CrewMemberID string
	ProjectID    string
	Role         string
	StartDate    time.Time
	EndDate      time.Time
	Status       AssignmentStatus
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func NewAssignment(id, tenantID, crewMemberID, projectID, role string, startDate, endDate time.Time) *Assignment {
	now := time.Now()
	return &Assignment{
		AggregateRoot: *events.NewAggregateRoot(id, "assignment"),
		TenantID:      tenantID,
		CrewMemberID:  crewMemberID,
		ProjectID:     projectID,
		Role:          role,
		StartDate:     startDate,
		EndDate:       endDate,
		Status:        AssignmentStatusActive,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

func (a *Assignment) Complete() error {
	if a.Status != AssignmentStatusActive {
		return fmt.Errorf("can only complete active assignments")
	}
	a.Status = AssignmentStatusCompleted
	a.UpdatedAt = time.Now()
	return nil
}

func (a *Assignment) Cancel() error {
	if a.Status == AssignmentStatusCompleted {
		return fmt.Errorf("cannot cancel completed assignments")
	}
	a.Status = AssignmentStatusCancelled
	a.UpdatedAt = time.Now()
	return nil
}

func (a *Assignment) IsActive(date time.Time) bool {
	return a.Status == AssignmentStatusActive &&
		!date.Before(a.StartDate) &&
		!date.After(a.EndDate)
}

func (a *Assignment) Validate() error {
	if a.ID == "" {
		return fmt.Errorf("assignment ID cannot be empty")
	}
	if a.TenantID == "" {
		return fmt.Errorf("tenant ID cannot be empty")
	}
	if a.CrewMemberID == "" {
		return fmt.Errorf("crew member ID cannot be empty")
	}
	if a.StartDate.After(a.EndDate) {
		return fmt.Errorf("start date must be before end date")
	}
	return nil
}

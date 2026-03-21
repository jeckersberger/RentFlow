package domain

import (
	"fmt"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/events"
)

type CrewMemberType string
type CrewMemberStatus string

const (
	CrewTypeEmployee   CrewMemberType = "employee"
	CrewTypeFreelancer CrewMemberType = "freelancer"
)

const (
	CrewStatusActive   CrewMemberStatus = "active"
	CrewStatusInactive CrewMemberStatus = "inactive"
)

type CrewMember struct {
	events.AggregateRoot
	TenantID             string
	UserID               string
	FirstName            string
	LastName             string
	Email                string
	Phone                string
	Type                 CrewMemberType
	Skills               []string
	HourlyRate           float64
	DailyRate            float64
	Status               CrewMemberStatus
	AvailabilityCalendar string
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

func NewCrewMember(id, tenantID, userID, firstName, lastName, email, phone string, crewType CrewMemberType) *CrewMember {
	now := time.Now()
	return &CrewMember{
		AggregateRoot:        *events.NewAggregateRoot(id, "crew_member"),
		TenantID:             tenantID,
		UserID:               userID,
		FirstName:            firstName,
		LastName:             lastName,
		Email:                email,
		Phone:                phone,
		Type:                 crewType,
		Skills:               []string{},
		Status:               CrewStatusActive,
		AvailabilityCalendar: "",
		CreatedAt:            now,
		UpdatedAt:            now,
	}
}

func (c *CrewMember) AddSkill(skill string) error {
	if skill == "" {
		return fmt.Errorf("skill cannot be empty")
	}
	for _, s := range c.Skills {
		if s == skill {
			return fmt.Errorf("skill already exists")
		}
	}
	c.Skills = append(c.Skills, skill)
	c.UpdatedAt = time.Now()
	return nil
}

func (c *CrewMember) SetRates(hourly, daily float64) error {
	if hourly < 0 || daily < 0 {
		return fmt.Errorf("rates cannot be negative")
	}
	c.HourlyRate = hourly
	c.DailyRate = daily
	c.UpdatedAt = time.Now()
	return nil
}

func (c *CrewMember) SetAvailability(calendar string) error {
	c.AvailabilityCalendar = calendar
	c.UpdatedAt = time.Now()
	return nil
}

func (c *CrewMember) SetStatus(status CrewMemberStatus) error {
	c.Status = status
	c.UpdatedAt = time.Now()
	return nil
}

func (c *CrewMember) Validate() error {
	if c.ID == "" {
		return fmt.Errorf("crew member ID cannot be empty")
	}
	if c.TenantID == "" {
		return fmt.Errorf("tenant ID cannot be empty")
	}
	if c.FirstName == "" || c.LastName == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if c.Email == "" {
		return fmt.Errorf("email cannot be empty")
	}
	return nil
}

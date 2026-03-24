package domain

import (
	"fmt"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/events"
)

type Project struct {
	events.AggregateRoot
	TenantID        string
	Name            string
	Description     string
	ClientName      string
	ClientEmail     string
	ClientPhone     string
	ClientAddress   Address
	VenueAddress    Address
	Status          ProjectStatus
	StartDate       time.Time
	EndDate         time.Time
	SetupDate       *time.Time
	TeardownDate    *time.Time
	ProjectManager  string // user ID
	Budget          float64
	Currency        string
	Notes           string
	Tags            []string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	CreatedByUserID string
}

type ProjectStatus string

const (
	ProjectDraft      ProjectStatus = "draft"
	ProjectQuoted     ProjectStatus = "quoted"
	ProjectConfirmed  ProjectStatus = "confirmed"
	ProjectInProgress ProjectStatus = "in_progress"
	ProjectCompleted  ProjectStatus = "completed"
	ProjectCancelled  ProjectStatus = "cancelled"
	ProjectInvoiced   ProjectStatus = "invoiced"
)

func NewProject(id, tenantID, name, clientName, userID string) *Project {
	now := time.Now()
	return &Project{
		AggregateRoot:   *events.NewAggregateRoot(id, "project"),
		TenantID:        tenantID,
		Name:            name,
		ClientName:      clientName,
		Status:          ProjectDraft,
		Tags:            []string{},
		Currency:        "USD",
		CreatedAt:       now,
		UpdatedAt:       now,
		CreatedByUserID: userID,
	}
}

func (p *Project) UpdateBasicInfo(name, description, clientName, clientEmail, clientPhone string) error {
	if name == "" {
		return fmt.Errorf("project name cannot be empty")
	}
	if clientName == "" {
		return fmt.Errorf("client name cannot be empty")
	}
	p.Name = name
	p.Description = description
	p.ClientName = clientName
	p.ClientEmail = clientEmail
	p.ClientPhone = clientPhone
	p.UpdatedAt = time.Now()
	return nil
}

func (p *Project) SetAddresses(clientAddr, venueAddr Address) error {
	if clientAddr.City == "" || venueAddr.City == "" {
		return fmt.Errorf("both client and venue addresses are required")
	}
	p.ClientAddress = clientAddr
	p.VenueAddress = venueAddr
	p.UpdatedAt = time.Now()
	return nil
}

func (p *Project) SetDates(startDate, endDate time.Time) error {
	if endDate.Before(startDate) {
		return fmt.Errorf("end date cannot be before start date")
	}
	p.StartDate = startDate
	p.EndDate = endDate
	p.UpdatedAt = time.Now()
	return nil
}

func (p *Project) SetupAndTeardownDates(setupDate, teardownDate *time.Time) error {
	if setupDate != nil && teardownDate != nil {
		if teardownDate.Before(*setupDate) {
			return fmt.Errorf("teardown date cannot be before setup date")
		}
	}
	p.SetupDate = setupDate
	p.TeardownDate = teardownDate
	p.UpdatedAt = time.Now()
	return nil
}

func (p *Project) ChangeStatus(newStatus ProjectStatus) error {
	if newStatus == p.Status {
		return fmt.Errorf("project already in status %s", newStatus)
	}

	// Validate state transitions
	switch p.Status {
	case ProjectDraft:
		if newStatus != ProjectQuoted && newStatus != ProjectCancelled {
			return fmt.Errorf("cannot transition from %s to %s", p.Status, newStatus)
		}
	case ProjectQuoted:
		if newStatus != ProjectConfirmed && newStatus != ProjectDraft && newStatus != ProjectCancelled {
			return fmt.Errorf("cannot transition from %s to %s", p.Status, newStatus)
		}
	case ProjectConfirmed:
		if newStatus != ProjectInProgress && newStatus != ProjectCancelled {
			return fmt.Errorf("cannot transition from %s to %s", p.Status, newStatus)
		}
	case ProjectInProgress:
		if newStatus != ProjectCompleted && newStatus != ProjectCancelled {
			return fmt.Errorf("cannot transition from %s to %s", p.Status, newStatus)
		}
	case ProjectCompleted:
		if newStatus != ProjectInvoiced {
			return fmt.Errorf("cannot transition from %s to %s", p.Status, newStatus)
		}
	case ProjectCancelled, ProjectInvoiced:
		return fmt.Errorf("cannot transition from %s status", p.Status)
	}

	p.Status = newStatus
	p.UpdatedAt = time.Now()
	return nil
}

func (p *Project) SetBudget(budget float64, currency string) error {
	if budget < 0 {
		return fmt.Errorf("budget cannot be negative")
	}
	if currency == "" {
		return fmt.Errorf("currency cannot be empty")
	}
	p.Budget = budget
	p.Currency = currency
	p.UpdatedAt = time.Now()
	return nil
}

func (p *Project) SetProjectManager(userID string) error {
	if userID == "" {
		return fmt.Errorf("project manager ID cannot be empty")
	}
	p.ProjectManager = userID
	p.UpdatedAt = time.Now()
	return nil
}

func (p *Project) AddTag(tag string) error {
	if tag == "" {
		return fmt.Errorf("tag cannot be empty")
	}
	for _, existing := range p.Tags {
		if existing == tag {
			return fmt.Errorf("tag already exists")
		}
	}
	p.Tags = append(p.Tags, tag)
	p.UpdatedAt = time.Now()
	return nil
}

func (p *Project) RemoveTag(tag string) error {
	for i, existing := range p.Tags {
		if existing == tag {
			p.Tags = append(p.Tags[:i], p.Tags[i+1:]...)
			p.UpdatedAt = time.Now()
			return nil
		}
	}
	return fmt.Errorf("tag not found")
}

func (p *Project) Validate() error {
	if p.ID == "" {
		return fmt.Errorf("project ID cannot be empty")
	}
	if p.TenantID == "" {
		return fmt.Errorf("tenant ID cannot be empty")
	}
	if p.Name == "" {
		return fmt.Errorf("project name cannot be empty")
	}
	// client_name is optional — projects can be created without a customer
	return nil
}

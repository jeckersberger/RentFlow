package domain

import (
	"fmt"
	"time"
)

type ContactType string

const (
	ContactTypeCompany ContactType = "company"
	ContactTypePerson  ContactType = "person"
)

type Contact struct {
	ID          string
	TenantID    string
	Type        ContactType
	CompanyName string
	FirstName   string
	LastName    string
	Email       string
	Phone       string
	Mobile      string
	Website     string
	Street      string
	HouseNumber string
	Zip         string
	City        string
	Country     string
	VatID       string
	Notes       string
	Tags        []string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	CreatedBy   string
}

func NewContact(id, tenantID string, contactType ContactType) *Contact {
	now := time.Now()
	return &Contact{
		ID:        id,
		TenantID:  tenantID,
		Type:      contactType,
		Country:   "Deutschland",
		Tags:      []string{},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (c *Contact) DisplayName() string {
	if c.Type == ContactTypeCompany {
		return c.CompanyName
	}
	name := c.FirstName
	if c.LastName != "" {
		if name != "" {
			name += " "
		}
		name += c.LastName
	}
	return name
}

func (c *Contact) Validate() error {
	if c.ID == "" {
		return fmt.Errorf("contact ID cannot be empty")
	}
	if c.TenantID == "" {
		return fmt.Errorf("tenant ID cannot be empty")
	}
	if c.Type != ContactTypeCompany && c.Type != ContactTypePerson {
		return fmt.Errorf("invalid contact type: %s", c.Type)
	}
	if c.Type == ContactTypeCompany && c.CompanyName == "" {
		return fmt.Errorf("company name is required for company contacts")
	}
	if c.Type == ContactTypePerson && c.LastName == "" {
		return fmt.Errorf("last name is required for person contacts")
	}
	return nil
}

func (c *Contact) Update() {
	c.UpdatedAt = time.Now()
}

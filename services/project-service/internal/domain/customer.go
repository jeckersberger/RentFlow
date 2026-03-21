package domain

import (
	"fmt"
	"time"
)

type Customer struct {
	ID          string
	TenantID    string
	Name        string
	Email       string
	Phone       string
	AddressStreet string
	AddressCity string
	AddressPostcode string
	AddressCountry string
	TaxID       string
	Notes       string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewCustomer(id, tenantID, name string) *Customer {
	now := time.Now()
	return &Customer{
		ID:             id,
		TenantID:       tenantID,
		Name:           name,
		AddressCountry: "DE",
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

func (c *Customer) UpdateBasicInfo(name, email, phone string) error {
	if name == "" {
		return fmt.Errorf("customer name cannot be empty")
	}
	c.Name = name
	c.Email = email
	c.Phone = phone
	c.UpdatedAt = time.Now()
	return nil
}

func (c *Customer) SetAddress(street, city, postcode, country string) error {
	if city == "" {
		return fmt.Errorf("city is required")
	}
	c.AddressStreet = street
	c.AddressCity = city
	c.AddressPostcode = postcode
	if country == "" {
		country = "DE"
	}
	c.AddressCountry = country
	c.UpdatedAt = time.Now()
	return nil
}

func (c *Customer) SetTaxID(taxID string) {
	c.TaxID = taxID
	c.UpdatedAt = time.Now()
}

func (c *Customer) SetNotes(notes string) {
	c.Notes = notes
	c.UpdatedAt = time.Now()
}

func (c *Customer) Validate() error {
	if c.ID == "" {
		return fmt.Errorf("customer ID cannot be empty")
	}
	if c.TenantID == "" {
		return fmt.Errorf("tenant ID cannot be empty")
	}
	if c.Name == "" {
		return fmt.Errorf("customer name cannot be empty")
	}
	return nil
}

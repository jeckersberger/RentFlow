package domain

import (
	"fmt"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/events"
)

type Flightcase struct {
	events.AggregateRoot
	TenantID        string
	Name            string
	Description     string
	Barcode         string
	Contents        []FlightcaseItem
	Weight          float64
	LocationID      string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	CreatedByUserID string
}

type FlightcaseItem struct {
	EquipmentID string
	Quantity    int
	AddedAt     time.Time
}

func NewFlightcase(id, tenantID, name, barcode, userID string) *Flightcase {
	now := time.Now()
	return &Flightcase{
		AggregateRoot:   *events.NewAggregateRoot(id, "flightcase"),
		TenantID:        tenantID,
		Name:            name,
		Barcode:         barcode,
		Contents:        []FlightcaseItem{},
		CreatedAt:       now,
		UpdatedAt:       now,
		CreatedByUserID: userID,
	}
}

func (f *Flightcase) AddItem(equipmentID string, quantity int) error {
	if equipmentID == "" {
		return fmt.Errorf("equipment ID cannot be empty")
	}
	if quantity <= 0 {
		return fmt.Errorf("quantity must be positive")
	}

	// Check if item already exists
	for i, item := range f.Contents {
		if item.EquipmentID == equipmentID {
			f.Contents[i].Quantity += quantity
			f.UpdatedAt = time.Now()
			return nil
		}
	}

	// Add new item
	f.Contents = append(f.Contents, FlightcaseItem{
		EquipmentID: equipmentID,
		Quantity:    quantity,
		AddedAt:     time.Now(),
	})
	f.UpdatedAt = time.Now()
	return nil
}

func (f *Flightcase) RemoveItem(equipmentID string, quantity int) error {
	if equipmentID == "" {
		return fmt.Errorf("equipment ID cannot be empty")
	}
	if quantity <= 0 {
		return fmt.Errorf("quantity must be positive")
	}

	for i, item := range f.Contents {
		if item.EquipmentID == equipmentID {
			if item.Quantity < quantity {
				return fmt.Errorf("not enough items to remove")
			}
			if item.Quantity == quantity {
				// Remove the entire item
				f.Contents = append(f.Contents[:i], f.Contents[i+1:]...)
			} else {
				f.Contents[i].Quantity -= quantity
			}
			f.UpdatedAt = time.Now()
			return nil
		}
	}

	return fmt.Errorf("equipment not found in flightcase")
}

func (f *Flightcase) SetLocation(locationID string) error {
	if locationID == "" {
		return fmt.Errorf("location ID cannot be empty")
	}
	f.LocationID = locationID
	f.UpdatedAt = time.Now()
	return nil
}

func (f *Flightcase) UpdateWeight(weight float64) error {
	if weight < 0 {
		return fmt.Errorf("weight cannot be negative")
	}
	f.Weight = weight
	f.UpdatedAt = time.Now()
	return nil
}

func (f *Flightcase) Validate() error {
	if f.ID == "" {
		return fmt.Errorf("flightcase ID cannot be empty")
	}
	if f.TenantID == "" {
		return fmt.Errorf("tenant ID cannot be empty")
	}
	if f.Name == "" {
		return fmt.Errorf("flightcase name cannot be empty")
	}
	if f.Barcode == "" {
		return fmt.Errorf("barcode cannot be empty")
	}
	return nil
}

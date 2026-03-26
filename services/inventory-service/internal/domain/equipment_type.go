package domain

import (
	"fmt"
	"time"
)

type EquipmentType struct {
	ID               string
	TenantID         string
	Name             string
	Description      string
	CategoryID       string
	Manufacturer     string
	Model            string
	SKUPrefix        string
	RentalPriceDay   float64
	RentalPriceWeek  float64
	ReplacementValue float64
	Weight           float64
	Dimensions       Dimensions
	ImageURL         string
	Tags             []string
	CustomFields     map[string]string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func NewEquipmentType(id, tenantID, name, categoryID string) *EquipmentType {
	now := time.Now()
	return &EquipmentType{
		ID:           id,
		TenantID:     tenantID,
		Name:         name,
		CategoryID:   categoryID,
		Tags:         []string{},
		CustomFields: make(map[string]string),
		Dimensions:   Dimensions{Unit: "cm"},
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

func (et *EquipmentType) Validate() error {
	if et.ID == "" {
		return fmt.Errorf("equipment type ID cannot be empty")
	}
	if et.TenantID == "" {
		return fmt.Errorf("tenant ID cannot be empty")
	}
	if et.Name == "" {
		return fmt.Errorf("equipment type name cannot be empty")
	}
	// CategoryID is optional — equipment types can exist without a category
	return nil
}

package domain

import (
	"fmt"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/events"
)

type Equipment struct {
	events.AggregateRoot
	TenantID        string
	Name            string
	Description     string
	CategoryID      string
	SKU             string
	SerialNumber    string
	Barcode         string
	Status          EquipmentStatus
	Condition       EquipmentCondition
	PurchaseDate    *time.Time
	PurchasePrice   float64
	RentalPriceDay  float64
	RentalPriceWeek float64
	Weight          float64
	Dimensions      Dimensions
	LocationID      string
	ImageRefs       []string
	Tags            []string
	CustomFields    map[string]string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	CreatedByUserID string
	RfidTag         string
}

type EquipmentStatus string

const (
	StatusAvailable     EquipmentStatus = "available"
	StatusReserved      EquipmentStatus = "reserved"
	StatusCheckedOut    EquipmentStatus = "checked_out"
	StatusInMaintenance EquipmentStatus = "in_maintenance"
	StatusRetired       EquipmentStatus = "retired"
	StatusDamaged       EquipmentStatus = "damaged"
)

type EquipmentCondition string

const (
	ConditionNew       EquipmentCondition = "new"
	ConditionExcellent EquipmentCondition = "excellent"
	ConditionGood      EquipmentCondition = "good"
	ConditionFair      EquipmentCondition = "fair"
	ConditionPoor      EquipmentCondition = "poor"
	ConditionDamaged   EquipmentCondition = "damaged"
	ConditionDefective EquipmentCondition = "defective"
)

func (c EquipmentCondition) IsValid() bool {
	switch c {
	case ConditionNew, ConditionExcellent, ConditionGood, ConditionFair, ConditionPoor, ConditionDamaged, ConditionDefective:
		return true
	}
	return false
}

type Dimensions struct {
	Length float64 `json:"length"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
	Unit   string  `json:"unit"`
}

func NewEquipment(id, tenantID, name, categoryID, sku, barcode, userID string) *Equipment {
	now := time.Now()
	return &Equipment{
		AggregateRoot:   *events.NewAggregateRoot(id, "equipment"),
		TenantID:        tenantID,
		Name:            name,
		CategoryID:      categoryID,
		SKU:             sku,
		Barcode:         barcode,
		Status:          StatusAvailable,
		Condition:       ConditionGood,
		Tags:            []string{},
		CustomFields:    make(map[string]string),
		ImageRefs:       []string{},
		CreatedAt:       now,
		UpdatedAt:       now,
		CreatedByUserID: userID,
	}
}

func (e *Equipment) UpdateBasicInfo(name, description string) error {
	if name == "" {
		return fmt.Errorf("equipment name cannot be empty")
	}
	e.Name = name
	e.Description = description
	e.UpdatedAt = time.Now()
	return nil
}

func (e *Equipment) SetLocation(locationID string) error {
	if locationID == "" {
		return fmt.Errorf("location ID cannot be empty")
	}
	e.LocationID = locationID
	e.UpdatedAt = time.Now()
	return nil
}

func (e *Equipment) ChangeStatus(newStatus EquipmentStatus) error {
	if newStatus == e.Status {
		return fmt.Errorf("equipment already in status %s", newStatus)
	}

	// Validate state transitions
	switch e.Status {
	case StatusAvailable:
		if newStatus != StatusReserved && newStatus != StatusCheckedOut &&
			newStatus != StatusInMaintenance && newStatus != StatusDamaged &&
			newStatus != StatusRetired {
			return fmt.Errorf("cannot transition from %s to %s", e.Status, newStatus)
		}
	case StatusReserved:
		if newStatus != StatusCheckedOut && newStatus != StatusAvailable &&
			newStatus != StatusRetired {
			return fmt.Errorf("cannot transition from %s to %s", e.Status, newStatus)
		}
	case StatusCheckedOut:
		if newStatus != StatusAvailable && newStatus != StatusDamaged &&
			newStatus != StatusInMaintenance && newStatus != StatusRetired {
			return fmt.Errorf("cannot transition from %s to %s", e.Status, newStatus)
		}
	case StatusInMaintenance:
		if newStatus != StatusAvailable && newStatus != StatusDamaged &&
			newStatus != StatusRetired {
			return fmt.Errorf("cannot transition from %s to %s", e.Status, newStatus)
		}
	case StatusDamaged:
		if newStatus != StatusInMaintenance && newStatus != StatusAvailable &&
			newStatus != StatusRetired {
			return fmt.Errorf("cannot transition from %s to %s", e.Status, newStatus)
		}
	case StatusRetired:
		return fmt.Errorf("cannot transition from retired status")
	}

	e.Status = newStatus
	e.UpdatedAt = time.Now()
	return nil
}

func (e *Equipment) UpdateCondition(condition EquipmentCondition) error {
	if !condition.IsValid() {
		return fmt.Errorf("invalid condition: %s", condition)
	}
	if condition == e.Condition {
		return fmt.Errorf("condition already set to %s", condition)
	}
	e.Condition = condition
	e.UpdatedAt = time.Now()
	return nil
}

func (e *Equipment) AddImage(ref string) error {
	if ref == "" {
		return fmt.Errorf("image reference cannot be empty")
	}
	for _, existing := range e.ImageRefs {
		if existing == ref {
			return fmt.Errorf("image already added")
		}
	}
	e.ImageRefs = append(e.ImageRefs, ref)
	e.UpdatedAt = time.Now()
	return nil
}

func (e *Equipment) AddTag(tag string) error {
	if tag == "" {
		return fmt.Errorf("tag cannot be empty")
	}
	for _, existing := range e.Tags {
		if existing == tag {
			return fmt.Errorf("tag already exists")
		}
	}
	e.Tags = append(e.Tags, tag)
	e.UpdatedAt = time.Now()
	return nil
}

func (e *Equipment) SetCustomField(key, value string) error {
	if key == "" {
		return fmt.Errorf("field key cannot be empty")
	}
	e.CustomFields[key] = value
	e.UpdatedAt = time.Now()
	return nil
}

func (e *Equipment) Validate() error {
	if e.ID == "" {
		return fmt.Errorf("equipment ID cannot be empty")
	}
	if e.TenantID == "" {
		return fmt.Errorf("tenant ID cannot be empty")
	}
	if e.Name == "" {
		return fmt.Errorf("equipment name cannot be empty")
	}
	if e.Barcode == "" {
		return fmt.Errorf("barcode cannot be empty")
	}
	// CategoryID is optional — equipment can exist without a category
	return nil
}

package domain

import (
	"fmt"
	"time"
)

type InventoryCheckStatus string

const (
	InventoryCheckPlanned     InventoryCheckStatus = "planned"
	InventoryCheckInProgress  InventoryCheckStatus = "in_progress"
	InventoryCheckCompleted   InventoryCheckStatus = "completed"
)

type InventoryCheck struct {
	ID          string
	TenantID    string
	Name        string
	LocationID  *string
	Status      InventoryCheckStatus
	Items       []InventoryCheckItem
	StartedAt   *time.Time
	CompletedAt *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type InventoryCheckItem struct {
	EquipmentID   string
	ExpectedCount int
	ActualCount   int
	Status        string
	Notes         string
	ScannedAt     *time.Time
}

func NewInventoryCheck(id, tenantID, name string) *InventoryCheck {
	return &InventoryCheck{
		ID:        id,
		TenantID:  tenantID,
		Name:      name,
		Status:    InventoryCheckPlanned,
		Items:     make([]InventoryCheckItem, 0),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (ic *InventoryCheck) Validate() error {
	if ic.TenantID == "" {
		return fmt.Errorf("tenant ID is required")
	}
	if ic.Name == "" {
		return fmt.Errorf("name is required")
	}
	return nil
}

func (ic *InventoryCheck) Start() error {
	if ic.Status != InventoryCheckPlanned {
		return fmt.Errorf("can only start a planned inventory check")
	}
	now := time.Now()
	ic.StartedAt = &now
	ic.Status = InventoryCheckInProgress
	ic.UpdatedAt = now
	return nil
}

func (ic *InventoryCheck) Complete() error {
	if ic.Status != InventoryCheckInProgress {
		return fmt.Errorf("can only complete an in-progress inventory check")
	}
	now := time.Now()
	ic.CompletedAt = &now
	ic.Status = InventoryCheckCompleted
	ic.UpdatedAt = now
	return nil
}

func (ic *InventoryCheck) AddItem(equipmentID string, expectedCount int) {
	item := InventoryCheckItem{
		EquipmentID:   equipmentID,
		ExpectedCount: expectedCount,
		ActualCount:   0,
		Status:        "pending",
	}
	ic.Items = append(ic.Items, item)
}

func (ic *InventoryCheck) ScanItem(equipmentID string) (*InventoryCheckItem, error) {
	for i := range ic.Items {
		if ic.Items[i].EquipmentID == equipmentID {
			ic.Items[i].ActualCount++
			now := time.Now()
			ic.Items[i].ScannedAt = &now

			if ic.Items[i].ActualCount > ic.Items[i].ExpectedCount {
				ic.Items[i].Status = "surplus"
			} else if ic.Items[i].ActualCount == ic.Items[i].ExpectedCount {
				ic.Items[i].Status = "found"
			}

			ic.UpdatedAt = time.Now()
			return &ic.Items[i], nil
		}
	}

	return nil, fmt.Errorf("equipment not in inventory check")
}

func (ic *InventoryCheck) GetDiscrepancies() []InventoryCheckItem {
	discrepancies := make([]InventoryCheckItem, 0)
	for _, item := range ic.Items {
		if item.ActualCount != item.ExpectedCount {
			discrepancies = append(discrepancies, item)
		}
	}
	return discrepancies
}

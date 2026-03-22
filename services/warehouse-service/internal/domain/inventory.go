package domain

import (
	"fmt"
	"time"
)

type InventoryCheckType string

const (
	InventoryCheckTypeFull      InventoryCheckType = "full"
	InventoryCheckTypeCycle     InventoryCheckType = "cycle"
	InventoryCheckTypeSpotCheck InventoryCheckType = "spot_check"
)

type InventoryCheckStatus string

const (
	InventoryCheckStatusPlanned    InventoryCheckStatus = "planned"
	InventoryCheckStatusInProgress InventoryCheckStatus = "in_progress"
	InventoryCheckStatusCompleted  InventoryCheckStatus = "completed"
)

type InventoryItemStatus string

const (
	InventoryItemStatusPending   InventoryItemStatus = "pending"
	InventoryItemStatusFound     InventoryItemStatus = "found"
	InventoryItemStatusSurplus   InventoryItemStatus = "surplus"
	InventoryItemStatusMissing   InventoryItemStatus = "missing"
	InventoryItemStatusDiscarded InventoryItemStatus = "discarded"
)

// InventoryCheck represents an inventory count operation
type InventoryCheck struct {
	ID          string
	TenantID    string
	WarehouseID *string
	ZoneID      *string
	CheckType   InventoryCheckType
	Name        string
	Description string
	Status      InventoryCheckStatus
	Items       []InventoryCheckItem
	StartedAt   *time.Time
	CompletedAt *time.Time
	CompletedBy *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// InventoryCheckItem represents an item scanned during inventory check
type InventoryCheckItem struct {
	ID          string
	TenantID    string
	CheckID     string
	EquipmentID string
	LocationID  *string
	ExpectedCount int
	ActualCount   int
	Variance      int
	Status        InventoryItemStatus
	ScannedAt     *time.Time
	Notes         string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// NewInventoryCheck creates a new inventory check
func NewInventoryCheck(id, tenantID, name string, checkType InventoryCheckType) *InventoryCheck {
	return &InventoryCheck{
		ID:        id,
		TenantID:  tenantID,
		Name:      name,
		CheckType: checkType,
		Status:    InventoryCheckStatusPlanned,
		Items:     make([]InventoryCheckItem, 0),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// Validate validates the inventory check
func (ic *InventoryCheck) Validate() error {
	if ic.TenantID == "" {
		return fmt.Errorf("tenant ID is required")
	}
	if ic.Name == "" {
		return fmt.Errorf("name is required")
	}
	if ic.CheckType == "" {
		return fmt.Errorf("check type is required")
	}
	return nil
}

// Start starts the inventory check
func (ic *InventoryCheck) Start() error {
	if ic.Status != InventoryCheckStatusPlanned {
		return fmt.Errorf("can only start a planned inventory check")
	}
	now := time.Now()
	ic.StartedAt = &now
	ic.Status = InventoryCheckStatusInProgress
	ic.UpdatedAt = now
	return nil
}

// Complete completes the inventory check
func (ic *InventoryCheck) Complete(completedBy string) error {
	if ic.Status != InventoryCheckStatusInProgress {
		return fmt.Errorf("can only complete an in-progress inventory check")
	}
	now := time.Now()
	ic.CompletedAt = &now
	ic.CompletedBy = &completedBy
	ic.Status = InventoryCheckStatusCompleted
	ic.UpdatedAt = now
	return nil
}

// AddItem adds an item to the inventory check
func (ic *InventoryCheck) AddItem(id, equipmentID string, expectedCount int, locationID *string) *InventoryCheckItem {
	item := InventoryCheckItem{
		ID:            id,
		TenantID:      ic.TenantID,
		CheckID:       ic.ID,
		EquipmentID:   equipmentID,
		LocationID:    locationID,
		ExpectedCount: expectedCount,
		ActualCount:   0,
		Status:        InventoryItemStatusPending,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	ic.Items = append(ic.Items, item)
	return &item
}

// ScanItem scans an equipment item during the inventory check
func (ic *InventoryCheck) ScanItem(equipmentID string, locationID *string) error {
	if ic.Status != InventoryCheckStatusInProgress {
		return fmt.Errorf("can only scan items in an in-progress inventory check")
	}

	for i := range ic.Items {
		if ic.Items[i].EquipmentID == equipmentID {
			ic.Items[i].ActualCount++
			now := time.Now()
			ic.Items[i].ScannedAt = &now
			ic.Items[i].Variance = ic.Items[i].ActualCount - ic.Items[i].ExpectedCount

			// Determine status based on variance
			if ic.Items[i].Variance > 0 {
				ic.Items[i].Status = InventoryItemStatusSurplus
			} else if ic.Items[i].Variance == 0 && ic.Items[i].ActualCount > 0 {
				ic.Items[i].Status = InventoryItemStatusFound
			}

			ic.Items[i].UpdatedAt = now
			ic.UpdatedAt = now
			return nil
		}
	}

	return fmt.Errorf("equipment %s not in inventory check", equipmentID)
}

// GetDiscrepancies returns items with variance
func (ic *InventoryCheck) GetDiscrepancies() []InventoryCheckItem {
	discrepancies := make([]InventoryCheckItem, 0)
	for _, item := range ic.Items {
		if item.Variance != 0 {
			discrepancies = append(discrepancies, item)
		}
	}
	return discrepancies
}

// GetMissingItems returns items that were not found
func (ic *InventoryCheck) GetMissingItems() []InventoryCheckItem {
	missing := make([]InventoryCheckItem, 0)
	for _, item := range ic.Items {
		if item.ActualCount == 0 && item.ExpectedCount > 0 {
			item.Status = InventoryItemStatusMissing
			missing = append(missing, item)
		}
	}
	return missing
}

// GetSummary returns a summary of the inventory check
func (ic *InventoryCheck) GetSummary() map[string]interface{} {
	totalExpected := 0
	totalActual := 0
	variances := 0
	missing := 0
	surplus := 0

	for _, item := range ic.Items {
		totalExpected += item.ExpectedCount
		totalActual += item.ActualCount
		if item.Variance != 0 {
			variances++
		}
		if item.Variance < 0 {
			missing++
		} else if item.Variance > 0 {
			surplus++
		}
	}

	accuracy := 0.0
	if totalExpected > 0 {
		accuracy = float64(totalActual) / float64(totalExpected) * 100
	}

	return map[string]interface{}{
		"total_expected": totalExpected,
		"total_actual":   totalActual,
		"variance_count": variances,
		"missing_count":  missing,
		"surplus_count":  surplus,
		"accuracy":       accuracy,
	}
}

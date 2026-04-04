package domain

import (
	"testing"

	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// Check status constants
// ---------------------------------------------------------------------------

func TestCheckStatusConstants(t *testing.T) {
	if CheckStatusInProgress != "in_progress" {
		t.Errorf("CheckStatusInProgress = %q, want %q", CheckStatusInProgress, "in_progress")
	}
	if CheckStatusCompleted != "completed" {
		t.Errorf("CheckStatusCompleted = %q, want %q", CheckStatusCompleted, "completed")
	}
}

// ---------------------------------------------------------------------------
// Warehouse — basic fields
// ---------------------------------------------------------------------------

func TestWarehouseFields(t *testing.T) {
	w := &Warehouse{
		ID:       uuid.New(),
		TenantID: uuid.New(),
		Name:     "Hauptlager",
		Code:     "HL-01",
		IsActive: true,
	}

	if w.Name != "Hauptlager" {
		t.Errorf("Name = %q, want %q", w.Name, "Hauptlager")
	}
	if w.Code != "HL-01" {
		t.Errorf("Code = %q, want %q", w.Code, "HL-01")
	}
	if !w.IsActive {
		t.Error("warehouse should be active")
	}
}

// ---------------------------------------------------------------------------
// Zone — climate control
// ---------------------------------------------------------------------------

func TestZoneClimateControl(t *testing.T) {
	z := &Zone{
		ID:                uuid.New(),
		WarehouseID:       uuid.New(),
		TenantID:          uuid.New(),
		Name:              "Klimalager",
		ClimateControlled: true,
	}

	if !z.ClimateControlled {
		t.Error("zone should be climate controlled")
	}
}

// ---------------------------------------------------------------------------
// Rack — capacity
// ---------------------------------------------------------------------------

func TestRackCapacity(t *testing.T) {
	r := &Rack{
		Levels:       4,
		BaysPerLevel: 6,
	}

	totalSlots := r.Levels * r.BaysPerLevel
	if totalSlots != 24 {
		t.Errorf("total slots = %d, want 24", totalSlots)
	}
}

// ---------------------------------------------------------------------------
// InventoryCheck — discrepancy tracking
// ---------------------------------------------------------------------------

func TestInventoryCheckDiscrepancy(t *testing.T) {
	check := &InventoryCheck{
		Status:        CheckStatusInProgress,
		ExpectedCount: 100,
		ActualCount:   97,
	}

	check.DiscrepancyCount = check.ExpectedCount - check.ActualCount
	if check.DiscrepancyCount != 3 {
		t.Errorf("DiscrepancyCount = %d, want 3", check.DiscrepancyCount)
	}
}

// ---------------------------------------------------------------------------
// Movement — direction
// ---------------------------------------------------------------------------

func TestMovementDirection(t *testing.T) {
	fromID := uuid.New()
	toID := uuid.New()

	m := &Movement{
		EquipmentID:    uuid.New(),
		FromLocationID: &fromID,
		ToLocationID:   &toID,
		Quantity:       5,
		Reason:         "project delivery",
	}

	if m.FromLocationID == nil {
		t.Error("FromLocationID should not be nil")
	}
	if m.ToLocationID == nil {
		t.Error("ToLocationID should not be nil")
	}
	if m.Quantity != 5 {
		t.Errorf("Quantity = %d, want 5", m.Quantity)
	}
}

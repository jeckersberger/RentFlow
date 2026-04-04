package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestMaintenanceScheduleNextDue(t *testing.T) {
	now := time.Now()
	nextDue := now.Add(30 * 24 * time.Hour)

	s := &MaintenanceSchedule{
		ID:           uuid.New(),
		TenantID:     uuid.New(),
		EquipmentID:  uuid.New(),
		Name:         "Jaehrliche Inspektion",
		IntervalDays: 365,
		NextDueAt:    &nextDue,
		IsActive:     true,
	}

	if s.IntervalDays != 365 {
		t.Errorf("IntervalDays = %d, want 365", s.IntervalDays)
	}
	if s.NextDueAt == nil || s.NextDueAt.Before(now) {
		t.Error("NextDueAt should be in the future")
	}
}

func TestMaintenanceTaskCostInCents(t *testing.T) {
	task := &MaintenanceTask{
		ID:       uuid.New(),
		Title:    "Kabelcheck",
		Status:   "pending",
		Priority: "high",
		Cost:     15000, // 150.00 EUR
	}

	if task.Cost != 15000 {
		t.Errorf("Cost = %d, want 15000", task.Cost)
	}
}

func TestMaintenanceTaskStatuses(t *testing.T) {
	validStatuses := []string{"pending", "in_progress", "completed"}
	for _, s := range validStatuses {
		task := &MaintenanceTask{Status: s}
		if task.Status == "" {
			t.Errorf("status %q should not be empty", s)
		}
	}
}

func TestMaintenanceTaskPriorities(t *testing.T) {
	validPriorities := []string{"low", "normal", "high", "critical"}
	for _, p := range validPriorities {
		task := &MaintenanceTask{Priority: p}
		if task.Priority == "" {
			t.Errorf("priority %q should not be empty", p)
		}
	}
}

func TestECheckRecordMeasurements(t *testing.T) {
	resistance := 0.15
	insulation := 2.5
	leakage := 0.3

	check := &ECheckRecord{
		ID:                            uuid.New(),
		EquipmentID:                   uuid.New(),
		CheckDate:                     "2026-04-01",
		Result:                        "passed",
		ProtectionConductorResistance: &resistance,
		InsulationResistance:          &insulation,
		LeakageCurrent:                &leakage,
	}

	if *check.ProtectionConductorResistance != 0.15 {
		t.Errorf("Resistance = %f, want 0.15", *check.ProtectionConductorResistance)
	}
	if *check.InsulationResistance < 1.0 {
		t.Error("InsulationResistance should be >= 1.0 MOhm for pass")
	}
}

func TestDomainErrors(t *testing.T) {
	errors := []error{
		ErrScheduleNotFound, ErrTaskNotFound, ErrNameRequired,
		ErrTitleRequired, ErrEquipmentRequired, ErrTaskAlreadyCompleted,
		ErrCheckDateRequired, ErrECheckNotFound,
	}
	for _, err := range errors {
		if err == nil || err.Error() == "" {
			t.Error("domain error should not be nil or empty")
		}
	}
}

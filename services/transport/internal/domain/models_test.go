package domain

import (
	"testing"
)

func TestValidOrderType(t *testing.T) {
	valid := []string{OrderTypeDelivery, OrderTypePickup, OrderTypeTransfer}
	for _, v := range valid {
		if !ValidOrderType(v) {
			t.Errorf("ValidOrderType(%q) = false, want true", v)
		}
	}
	invalid := []string{"", "unknown", "DELIVERY", "move"}
	for _, v := range invalid {
		if ValidOrderType(v) {
			t.Errorf("ValidOrderType(%q) = true, want false", v)
		}
	}
}

func TestValidStatus(t *testing.T) {
	valid := []string{StatusPlanned, StatusInTransit, StatusCompleted, StatusCancelled}
	for _, v := range valid {
		if !ValidStatus(v) {
			t.Errorf("ValidStatus(%q) = false, want true", v)
		}
	}
	invalid := []string{"", "active", "IN_TRANSIT", "done"}
	for _, v := range invalid {
		if ValidStatus(v) {
			t.Errorf("ValidStatus(%q) = true, want false", v)
		}
	}
}

func TestVehicleTypeConstants(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{"Van", VehicleTypeVan, "van"},
		{"Truck", VehicleTypeTruck, "truck"},
		{"Trailer", VehicleTypeTrailer, "trailer"},
		{"Car", VehicleTypeCar, "car"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("%s = %q, want %q", tt.name, tt.got, tt.want)
			}
		})
	}
}

func TestCapacityCheckResult(t *testing.T) {
	result := &CapacityCheckResult{
		PayloadKg:  3500,
		VolumeM3:   20,
		UsedKg:     2800,
		UsedM3:     15,
		FitsWeight: true,
		FitsVolume: true,
	}

	remainingKg := result.PayloadKg - result.UsedKg
	if remainingKg != 700 {
		t.Errorf("remaining kg = %d, want 700", remainingKg)
	}

	if !result.FitsWeight || !result.FitsVolume {
		t.Error("should fit both weight and volume")
	}

	overloaded := &CapacityCheckResult{
		PayloadKg:  1000,
		UsedKg:     1500,
		FitsWeight: false,
	}
	if overloaded.FitsWeight {
		t.Error("should not fit when overloaded")
	}
}

func TestTransportCostAmountCents(t *testing.T) {
	cost := &TransportCost{
		CostType:    "fuel",
		AmountCents: 8500, // 85.00 EUR
	}
	if cost.AmountCents != 8500 {
		t.Errorf("AmountCents = %d, want 8500", cost.AmountCents)
	}
}

func TestDomainErrors(t *testing.T) {
	errors := []error{
		ErrVehicleNotFound, ErrOrderNotFound, ErrInvalidOrderType,
		ErrInvalidStatus, ErrMissingVehicleName, ErrOrderAlreadyCompleted,
		ErrNoVehicleAssigned, ErrCostTypeRequired,
	}
	for _, err := range errors {
		if err == nil || err.Error() == "" {
			t.Error("domain error should not be nil or empty")
		}
	}
}

package domain

import (
	"testing"
)

// ---------------------------------------------------------------------------
// ValidateStatus — Equipment status validation
// ---------------------------------------------------------------------------

func TestEquipmentValidateStatus(t *testing.T) {
	e := &Equipment{}

	validStatuses := []string{
		StatusAvailable,
		StatusReserved,
		StatusCheckedOut,
		StatusInMaintenance,
		StatusDamaged,
		StatusRetired,
	}

	for _, s := range validStatuses {
		t.Run("valid_"+s, func(t *testing.T) {
			if !e.ValidateStatus(s) {
				t.Errorf("ValidateStatus(%q) = false, want true", s)
			}
		})
	}

	invalidStatuses := []string{"", "unknown", "AVAILABLE", "Available", "active", "deleted"}
	for _, s := range invalidStatuses {
		t.Run("invalid_"+s, func(t *testing.T) {
			if e.ValidateStatus(s) {
				t.Errorf("ValidateStatus(%q) = true, want false", s)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// ValidateCondition — Equipment condition validation
// ---------------------------------------------------------------------------

func TestEquipmentValidateCondition(t *testing.T) {
	e := &Equipment{}

	validConditions := []string{
		ConditionOperational,
		ConditionGood,
		ConditionFair,
		ConditionDamaged,
		ConditionDecommissioned,
	}

	for _, c := range validConditions {
		t.Run("valid_"+c, func(t *testing.T) {
			if !e.ValidateCondition(c) {
				t.Errorf("ValidateCondition(%q) = false, want true", c)
			}
		})
	}

	invalidConditions := []string{"", "unknown", "GOOD", "Good", "excellent", "broken"}
	for _, c := range invalidConditions {
		t.Run("invalid_"+c, func(t *testing.T) {
			if e.ValidateCondition(c) {
				t.Errorf("ValidateCondition(%q) = true, want false", c)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// IsAvailable — combines status and active flag
// ---------------------------------------------------------------------------

func TestEquipmentIsAvailable(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		isActive bool
		want     bool
	}{
		{name: "available and active", status: StatusAvailable, isActive: true, want: true},
		{name: "available but inactive", status: StatusAvailable, isActive: false, want: false},
		{name: "reserved and active", status: StatusReserved, isActive: true, want: false},
		{name: "checked_out and active", status: StatusCheckedOut, isActive: true, want: false},
		{name: "in_maintenance and active", status: StatusInMaintenance, isActive: true, want: false},
		{name: "damaged and active", status: StatusDamaged, isActive: true, want: false},
		{name: "retired and active", status: StatusRetired, isActive: true, want: false},
		{name: "retired and inactive", status: StatusRetired, isActive: false, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Equipment{Status: tt.status, IsActive: tt.isActive}
			got := e.IsAvailable()
			if got != tt.want {
				t.Errorf("Equipment{Status: %q, IsActive: %v}.IsAvailable() = %v, want %v",
					tt.status, tt.isActive, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Status constants — Verify string values
// ---------------------------------------------------------------------------

func TestEquipmentStatusConstants(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{"StatusAvailable", StatusAvailable, "available"},
		{"StatusReserved", StatusReserved, "reserved"},
		{"StatusCheckedOut", StatusCheckedOut, "checked_out"},
		{"StatusInMaintenance", StatusInMaintenance, "in_maintenance"},
		{"StatusDamaged", StatusDamaged, "damaged"},
		{"StatusRetired", StatusRetired, "retired"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("%s = %q, want %q", tt.name, tt.got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Condition constants — Verify string values
// ---------------------------------------------------------------------------

func TestEquipmentConditionConstants(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{"ConditionOperational", ConditionOperational, "operational"},
		{"ConditionGood", ConditionGood, "good"},
		{"ConditionFair", ConditionFair, "fair"},
		{"ConditionDamaged", ConditionDamaged, "damaged"},
		{"ConditionDecommissioned", ConditionDecommissioned, "decommissioned"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("%s = %q, want %q", tt.name, tt.got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// History Action constants
// ---------------------------------------------------------------------------

func TestHistoryActionConstants(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{"ActionCreated", ActionCreated, "created"},
		{"ActionUpdated", ActionUpdated, "updated"},
		{"ActionStatusChanged", ActionStatusChanged, "status_changed"},
		{"ActionConditionChanged", ActionConditionChanged, "condition_changed"},
		{"ActionRFIDAssigned", ActionRFIDAssigned, "rfid_assigned"},
		{"ActionCheckedOut", ActionCheckedOut, "checked_out"},
		{"ActionCheckedIn", ActionCheckedIn, "checked_in"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("%s = %q, want %q", tt.name, tt.got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Equipment pricing (all in cents)
// ---------------------------------------------------------------------------

func TestEquipmentPricingInCents(t *testing.T) {
	e := Equipment{
		RentalPriceDay:   5000,  // 50.00 EUR
		RentalPriceWeek:  25000, // 250.00 EUR
		ReplacementValue: 150000, // 1500.00 EUR
		PurchasePrice:    120000, // 1200.00 EUR
	}

	if e.RentalPriceDay != 5000 {
		t.Errorf("RentalPriceDay = %d, want 5000", e.RentalPriceDay)
	}
	if e.RentalPriceWeek != 25000 {
		t.Errorf("RentalPriceWeek = %d, want 25000", e.RentalPriceWeek)
	}

	// Weekly price should be less than 7 * daily price (discount expected)
	weeklyIfDaily := e.RentalPriceDay * 7
	if e.RentalPriceWeek >= weeklyIfDaily {
		t.Errorf("Weekly price %d should be less than 7 * daily price %d for this fixture",
			e.RentalPriceWeek, weeklyIfDaily)
	}
}

// ---------------------------------------------------------------------------
// Equipment quantity tracking
// ---------------------------------------------------------------------------

func TestEquipmentQuantityTracking(t *testing.T) {
	e := Equipment{
		QuantityTotal:     10,
		QuantityAvailable: 7,
	}

	inUse := e.QuantityTotal - e.QuantityAvailable
	if inUse != 3 {
		t.Errorf("in-use quantity = %d, want 3", inUse)
	}
}

// ---------------------------------------------------------------------------
// PriceRule tier calculation
// ---------------------------------------------------------------------------

func TestPriceRuleTierCalculation(t *testing.T) {
	tier2From := 4
	tier2Price := int64(4000)
	tier3From := 8
	tier3Price := int64(3000)

	rule := PriceRule{
		BasePriceDay:  5000,
		Tier2FromDays: &tier2From,
		Tier2PriceDay: &tier2Price,
		Tier3FromDays: &tier3From,
		Tier3PriceDay: &tier3Price,
		IsActive:      true,
	}

	tests := []struct {
		name     string
		days     int
		wantRate int64
	}{
		{name: "1 day base rate", days: 1, wantRate: 5000},
		{name: "3 days base rate", days: 3, wantRate: 5000},
		{name: "4 days tier 2", days: 4, wantRate: 4000},
		{name: "7 days tier 2", days: 7, wantRate: 4000},
		{name: "8 days tier 3", days: 8, wantRate: 3000},
		{name: "30 days tier 3", days: 30, wantRate: 3000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rate := rule.BasePriceDay
			if rule.Tier2FromDays != nil && tt.days >= *rule.Tier2FromDays {
				rate = *rule.Tier2PriceDay
			}
			if rule.Tier3FromDays != nil && tt.days >= *rule.Tier3FromDays {
				rate = *rule.Tier3PriceDay
			}

			if rate != tt.wantRate {
				t.Errorf("rate for %d days = %d, want %d", tt.days, rate, tt.wantRate)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// PriceRule quantity discount
// ---------------------------------------------------------------------------

func TestPriceRuleQuantityDiscount(t *testing.T) {
	threshold := 5
	rule := PriceRule{
		BasePriceDay:         5000,
		QtyDiscountThreshold: &threshold,
		QtyDiscountPct:       10,
	}

	tests := []struct {
		name         string
		qty          int
		wantDiscount bool
	}{
		{name: "below threshold", qty: 3, wantDiscount: false},
		{name: "at threshold", qty: 5, wantDiscount: true},
		{name: "above threshold", qty: 10, wantDiscount: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasDiscount := rule.QtyDiscountThreshold != nil && tt.qty >= *rule.QtyDiscountThreshold
			if hasDiscount != tt.wantDiscount {
				t.Errorf("qty=%d hasDiscount=%v, want %v", tt.qty, hasDiscount, tt.wantDiscount)
			}

			if hasDiscount {
				baseTotal := rule.BasePriceDay * int64(tt.qty)
				discounted := baseTotal - (baseTotal * int64(rule.QtyDiscountPct) / 100)
				if discounted >= baseTotal {
					t.Errorf("discounted total %d should be less than base total %d", discounted, baseTotal)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TypeAvailabilitySummary percentage
// ---------------------------------------------------------------------------

func TestTypeAvailabilitySummaryPercentage(t *testing.T) {
	tests := []struct {
		name    string
		total   int
		avail   int
		wantPct int
	}{
		{name: "all available", total: 10, avail: 10, wantPct: 100},
		{name: "none available", total: 10, avail: 0, wantPct: 0},
		{name: "half available", total: 10, avail: 5, wantPct: 50},
		{name: "one of three", total: 3, avail: 1, wantPct: 33},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pct := 0
			if tt.total > 0 {
				pct = tt.avail * 100 / tt.total
			}
			if pct != tt.wantPct {
				t.Errorf("availability pct = %d, want %d", pct, tt.wantPct)
			}
		})
	}
}

package domain

import (
	"testing"
)

func TestValidateClaimTransition(t *testing.T) {
	// Happy path — full workflow
	workflow := []struct{ from, to string }{
		{"reported", "assessed"},
		{"assessed", "documented"},
		{"documented", "repair_approved"},
		{"repair_approved", "claim_submitted"},
		{"claim_submitted", "claim_approved"},
		{"claim_approved", "closed"},
	}
	for _, step := range workflow {
		t.Run(step.from+"->"+step.to, func(t *testing.T) {
			if !ValidateClaimTransition(step.from, step.to) {
				t.Errorf("transition %s -> %s should be valid", step.from, step.to)
			}
		})
	}

	// Invalid transitions
	invalid := []struct{ from, to string }{
		{"reported", "closed"},
		{"reported", "documented"},
		{"assessed", "closed"},
		{"closed", "reported"},
		{"unknown", "reported"},
	}
	for _, step := range invalid {
		t.Run("invalid_"+step.from+"->"+step.to, func(t *testing.T) {
			if ValidateClaimTransition(step.from, step.to) {
				t.Errorf("transition %s -> %s should be invalid", step.from, step.to)
			}
		})
	}

	// Legacy "submitted" status
	t.Run("legacy_submitted_to_reported", func(t *testing.T) {
		if !ValidateClaimTransition("submitted", "reported") {
			t.Error("submitted -> reported should be valid (legacy)")
		}
	})
	t.Run("legacy_submitted_to_assessed", func(t *testing.T) {
		if !ValidateClaimTransition("submitted", "assessed") {
			t.Error("submitted -> assessed should be valid (legacy)")
		}
	})
	t.Run("legacy_submitted_to_closed", func(t *testing.T) {
		if ValidateClaimTransition("submitted", "closed") {
			t.Error("submitted -> closed should be invalid")
		}
	})
}

func TestValidClaimStatuses(t *testing.T) {
	expected := []string{
		"reported", "assessed", "documented", "repair_approved",
		"claim_submitted", "claim_approved", "closed", "submitted",
	}
	for _, s := range expected {
		if !ValidClaimStatuses[s] {
			t.Errorf("ValidClaimStatuses[%q] = false, want true", s)
		}
	}

	invalid := []string{"", "unknown", "REPORTED", "denied"}
	for _, s := range invalid {
		if ValidClaimStatuses[s] {
			t.Errorf("ValidClaimStatuses[%q] = true, want false", s)
		}
	}
}

func TestInsurancePolicyAmountsInCents(t *testing.T) {
	p := &InsurancePolicy{
		Name:           "Veranstaltungshaftpflicht",
		CoverageAmount: 500000000, // 5 Mio EUR
		Deductible:     50000,     // 500 EUR
		Premium:        120000,    // 1200 EUR/Jahr
		IsActive:       true,
	}

	if p.CoverageAmount != 500000000 {
		t.Errorf("CoverageAmount = %d, want 500000000", p.CoverageAmount)
	}
	if p.Premium > p.CoverageAmount {
		t.Error("premium should be less than coverage")
	}
}

func TestInsuranceClaimAmounts(t *testing.T) {
	c := &InsuranceClaim{
		DamageAmount: 250000, // 2500 EUR
		ClaimAmount:  200000, // 2000 EUR (minus Selbstbeteiligung)
		Status:       "reported",
	}

	if c.ClaimAmount > c.DamageAmount {
		t.Error("claim amount should not exceed damage amount")
	}
}

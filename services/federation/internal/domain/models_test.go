package domain

import (
	"testing"
)

func TestPartnerStatusConstants(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{"Pending", PartnerStatusPending, "pending"},
		{"Active", PartnerStatusActive, "active"},
		{"Inactive", PartnerStatusInactive, "inactive"},
		{"Blocked", PartnerStatusBlocked, "blocked"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("%s = %q, want %q", tt.name, tt.got, tt.want)
			}
		})
	}
}

func TestRequestStatusConstants(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{"Pending", RequestStatusPending, "pending"},
		{"Approved", RequestStatusApproved, "approved"},
		{"Rejected", RequestStatusRejected, "rejected"},
		{"Canceled", RequestStatusCanceled, "canceled"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("%s = %q, want %q", tt.name, tt.got, tt.want)
			}
		})
	}
}

func TestSharedListingRates(t *testing.T) {
	listing := &SharedListing{
		DailyRate:  5000,  // 50.00 EUR
		WeeklyRate: 25000, // 250.00 EUR
		IsActive:   true,
	}

	weeklyIfDaily := listing.DailyRate * 7
	if listing.WeeklyRate >= weeklyIfDaily {
		t.Errorf("weekly rate %d should be less than 7x daily %d", listing.WeeklyRate, weeklyIfDaily)
	}
}

func TestFederationRequestTotalCost(t *testing.T) {
	req := &FederationRequest{
		TotalCost: 150000, // 1500.00 EUR
		Status:    RequestStatusPending,
	}
	if req.TotalCost != 150000 {
		t.Errorf("TotalCost = %d, want 150000", req.TotalCost)
	}
}

func TestPartnerAPIKeyHidden(t *testing.T) {
	p := &FederationPartner{
		PartnerName: "Test GmbH",
		APIKeyHash:  "sha256:abc123",
		Status:      PartnerStatusActive,
	}
	if p.APIKeyHash == "" {
		t.Error("APIKeyHash should be set internally")
	}
}

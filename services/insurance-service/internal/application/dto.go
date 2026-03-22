package application

import (
	"time"

	"github.com/google/uuid"
)

// CreatePolicyRequest is the DTO for creating a policy
type CreatePolicyRequest struct {
	TenantID       uuid.UUID `json:"tenant_id"`
	PolicyNumber   string    `json:"policy_number"`
	PolicyType     string    `json:"policy_type"`
	Provider       string    `json:"provider"`
	CoverageAmount float64   `json:"coverage_amount"`
	Deductible     float64   `json:"deductible"`
	PremiumAnnual  float64   `json:"premium_annual"`
	PremiumMonthly float64   `json:"premium_monthly"`
	StartDate      time.Time `json:"start_date"`
	EndDate        time.Time `json:"end_date"`
	Notes          *string   `json:"notes,omitempty"`
}

// UpdatePolicyRequest is the DTO for updating a policy
type UpdatePolicyRequest struct {
	Provider       *string    `json:"provider,omitempty"`
	CoverageAmount *float64   `json:"coverage_amount,omitempty"`
	Deductible     *float64   `json:"deductible,omitempty"`
	PremiumAnnual  *float64   `json:"premium_annual,omitempty"`
	PremiumMonthly *float64   `json:"premium_monthly,omitempty"`
	EndDate        *time.Time `json:"end_date,omitempty"`
	Status         *string    `json:"status,omitempty"`
	Notes          *string    `json:"notes,omitempty"`
}

// PolicyResponse is the DTO for returning a policy
type PolicyResponse struct {
	ID             uuid.UUID `json:"id"`
	TenantID       uuid.UUID `json:"tenant_id"`
	PolicyNumber   string    `json:"policy_number"`
	PolicyType     string    `json:"policy_type"`
	Provider       string    `json:"provider"`
	CoverageAmount float64   `json:"coverage_amount"`
	Deductible     float64   `json:"deductible"`
	PremiumAnnual  float64   `json:"premium_annual"`
	PremiumMonthly float64   `json:"premium_monthly"`
	StartDate      time.Time `json:"start_date"`
	EndDate        time.Time `json:"end_date"`
	Status         string    `json:"status"`
	Notes          *string   `json:"notes,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// CreateClaimRequest is the DTO for creating a claim
type CreateClaimRequest struct {
	TenantID    uuid.UUID `json:"tenant_id"`
	PolicyID    uuid.UUID `json:"policy_id"`
	EquipmentID *uuid.UUID `json:"equipment_id,omitempty"`
	ProjectID   *uuid.UUID `json:"project_id,omitempty"`
	IncidentDate time.Time `json:"incident_date"`
	Description string    `json:"description"`
	DamageType  string    `json:"damage_type"`
	ClaimedAmount float64 `json:"claimed_amount"`
	CreatedBy   uuid.UUID `json:"created_by"`
}

// UpdateClaimRequest is the DTO for updating a claim
type UpdateClaimRequest struct {
	Description   *string   `json:"description,omitempty"`
	AdjusterNotes *string   `json:"adjuster_notes,omitempty"`
	ApprovedAmount *float64 `json:"approved_amount,omitempty"`
	SettledAmount  *float64 `json:"settled_amount,omitempty"`
}

// ClaimResponse is the DTO for returning a claim
type ClaimResponse struct {
	ID             uuid.UUID  `json:"id"`
	TenantID       uuid.UUID  `json:"tenant_id"`
	PolicyID       uuid.UUID  `json:"policy_id"`
	ClaimNumber    string     `json:"claim_number"`
	EquipmentID    *uuid.UUID `json:"equipment_id,omitempty"`
	ProjectID      *uuid.UUID `json:"project_id,omitempty"`
	IncidentDate   time.Time  `json:"incident_date"`
	ReportedDate   time.Time  `json:"reported_date"`
	Description    string     `json:"description"`
	DamageType     string     `json:"damage_type"`
	Status         string     `json:"status"`
	ClaimedAmount  float64    `json:"claimed_amount"`
	ApprovedAmount *float64   `json:"approved_amount,omitempty"`
	SettledAmount  *float64   `json:"settled_amount,omitempty"`
	AdjusterNotes  *string    `json:"adjuster_notes,omitempty"`
	CreatedBy      uuid.UUID  `json:"created_by"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// CreateClaimItemRequest is the DTO for creating a claim item
type CreateClaimItemRequest struct {
	EquipmentID      uuid.UUID  `json:"equipment_id"`
	Description      string     `json:"description"`
	ReplacementValue float64    `json:"replacement_value"`
	RepairCost       *float64   `json:"repair_cost,omitempty"`
	PhotoURLs        []string   `json:"photo_urls,omitempty"`
}

// ClaimItemResponse is the DTO for returning a claim item
type ClaimItemResponse struct {
	ID               uuid.UUID `json:"id"`
	ClaimID          uuid.UUID `json:"claim_id"`
	EquipmentID      uuid.UUID `json:"equipment_id"`
	Description      string    `json:"description"`
	ReplacementValue float64   `json:"replacement_value"`
	RepairCost       *float64  `json:"repair_cost,omitempty"`
	PhotoURLs        []string  `json:"photo_urls,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
}

// TransitionClaimRequest is the DTO for transitioning a claim status
type TransitionClaimRequest struct {
	Status string `json:"status"`
}

// ApproveClaimRequest is the DTO for approving a claim
type ApproveClaimRequest struct {
	ApprovedAmount float64 `json:"approved_amount"`
	AdjusterNotes  *string `json:"adjuster_notes,omitempty"`
}

// RejectClaimRequest is the DTO for rejecting a claim
type RejectClaimRequest struct {
	AdjusterNotes string `json:"adjuster_notes"`
}

// SettleClaimRequest is the DTO for settling a claim
type SettleClaimRequest struct {
	SettledAmount float64 `json:"settled_amount"`
	AdjusterNotes *string `json:"adjuster_notes,omitempty"`
}

// DashboardStatsResponse is the DTO for dashboard statistics
type DashboardStatsResponse struct {
	YearlyStats []YearlyClaimStats `json:"yearly_stats"`
	PremiumDevelopment []PremiumDevelopment `json:"premium_development"`
}

// YearlyClaimStats represents claims per year
type YearlyClaimStats struct {
	Year       int   `json:"year"`
	ClaimCount int   `json:"claim_count"`
	TotalClaimed float64 `json:"total_claimed"`
	TotalApproved float64 `json:"total_approved"`
}

// PremiumDevelopment represents premium information over time
type PremiumDevelopment struct {
	Year int     `json:"year"`
	AnnualPremium float64 `json:"annual_premium"`
	ActivePolicies int `json:"active_policies"`
}

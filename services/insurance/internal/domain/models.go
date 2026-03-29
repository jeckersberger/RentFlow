package domain

import (
	"time"

	"github.com/google/uuid"
)

type InsurancePolicy struct {
	ID             uuid.UUID  `json:"id"`
	TenantID       uuid.UUID  `json:"tenant_id"`
	Name           string     `json:"name"`
	Provider       string     `json:"provider"`
	PolicyNumber   string     `json:"policy_number"`
	CoverageType   string     `json:"coverage_type"`
	CoverageAmount int64      `json:"coverage_amount"`
	Deductible     int64      `json:"deductible"`
	Premium        int64      `json:"premium"`
	StartDate      *string    `json:"start_date,omitempty"`
	EndDate        *string    `json:"end_date,omitempty"`
	IsActive       bool       `json:"is_active"`
	Notes          string     `json:"notes"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type InsuredEquipment struct {
	ID           uuid.UUID `json:"id"`
	PolicyID     uuid.UUID `json:"policy_id"`
	EquipmentID  uuid.UUID `json:"equipment_id"`
	InsuredValue int64     `json:"insured_value"`
	AddedAt      time.Time `json:"added_at"`
}

type InsuranceClaim struct {
	ID           uuid.UUID  `json:"id"`
	PolicyID     uuid.UUID  `json:"policy_id"`
	TenantID     uuid.UUID  `json:"tenant_id"`
	EquipmentID  *uuid.UUID `json:"equipment_id,omitempty"`
	ClaimNumber  string     `json:"claim_number"`
	Description  string     `json:"description"`
	DamageAmount int64      `json:"damage_amount"`
	ClaimAmount  int64      `json:"claim_amount"`
	Status       string     `json:"status"`
	IncidentDate *string    `json:"incident_date,omitempty"`
	FiledAt      time.Time  `json:"filed_at"`
	ResolvedAt   *time.Time `json:"resolved_at,omitempty"`
	Notes        string     `json:"notes"`
	CreatedBy    *uuid.UUID `json:"created_by,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

// ---------------------------------------------------------------------------
// Claim status workflow
// ---------------------------------------------------------------------------

// Valid claim statuses and their allowed transitions.
// reported -> assessed -> documented -> repair_approved -> claim_submitted -> claim_approved -> closed
var validClaimTransitions = map[string][]string{
	"reported":        {"assessed"},
	"assessed":        {"documented"},
	"documented":      {"repair_approved"},
	"repair_approved": {"claim_submitted"},
	"claim_submitted": {"claim_approved"},
	"claim_approved":  {"closed"},
}

// ValidClaimStatuses lists all valid claim statuses.
var ValidClaimStatuses = map[string]bool{
	"reported":        true,
	"assessed":        true,
	"documented":      true,
	"repair_approved": true,
	"claim_submitted": true,
	"claim_approved":  true,
	"closed":          true,
	// Legacy status from initial implementation.
	"submitted":       true,
}

// ValidateClaimTransition checks if a status transition is allowed.
func ValidateClaimTransition(from, to string) bool {
	// Allow transition from legacy "submitted" status to "reported" or "assessed".
	if from == "submitted" {
		return to == "reported" || to == "assessed"
	}

	allowed, ok := validClaimTransitions[from]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == to {
			return true
		}
	}
	return false
}

type PolicyFilter struct {
	Page    int
	PerPage int
}

type EquipmentFilter struct {
	Page    int
	PerPage int
}

type ClaimFilter struct {
	Page     int
	PerPage  int
	PolicyID *uuid.UUID
	Status   string
}

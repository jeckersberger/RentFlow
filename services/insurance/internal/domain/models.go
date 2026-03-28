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

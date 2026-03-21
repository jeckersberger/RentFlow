package domain

import (
	"encoding/json"
	"errors"
	"time"
)

var (
	ErrPolicyNotFound         = errors.New("policy not found")
	ErrClaimNotFound          = errors.New("claim not found")
	ErrRiskAssessmentNotFound = errors.New("risk assessment not found")
	ErrInvalidInput           = errors.New("invalid input")
	ErrTenantIDRequired       = errors.New("tenant ID required")
)

type PolicyType string

const (
	PolicyTypeEquipment PolicyType = "equipment"
	PolicyTypeLiability PolicyType = "liability"
	PolicyTypeTransport PolicyType = "transport"
	PolicyTypeEvent     PolicyType = "event"
)

type PolicyStatus string

const (
	PolicyStatusActive    PolicyStatus = "active"
	PolicyStatusExpired   PolicyStatus = "expired"
	PolicyStatusCancelled PolicyStatus = "cancelled"
)

type ClaimStatus string

const (
	ClaimStatusFiled    ClaimStatus = "filed"
	ClaimStatusReview   ClaimStatus = "under_review"
	ClaimStatusApproved ClaimStatus = "approved"
	ClaimStatusRejected ClaimStatus = "rejected"
	ClaimStatusPaid     ClaimStatus = "paid"
)

type Policy struct {
	ID             string       `json:"id"`
	TenantID       string       `json:"tenant_id"`
	PolicyNumber   string       `json:"policy_number"`
	Provider       string       `json:"provider"`
	Type           PolicyType   `json:"type"`
	CoverageAmount float64      `json:"coverage_amount"`
	Deductible     float64      `json:"deductible"`
	Premium        float64      `json:"premium"`
	StartDate      time.Time    `json:"start_date"`
	EndDate        time.Time    `json:"end_date"`
	Status         PolicyStatus `json:"status"`
	Notes          string       `json:"notes"`
	CreatedAt      time.Time    `json:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at"`
}

type Claim struct {
	ID           string      `json:"id"`
	TenantID     string      `json:"tenant_id"`
	PolicyID     string      `json:"policy_id"`
	EquipmentID  string      `json:"equipment_id"`
	IncidentDate time.Time   `json:"incident_date"`
	Description  string      `json:"description"`
	DamageAmount float64     `json:"damage_amount"`
	ClaimAmount  float64     `json:"claim_amount"`
	Status       ClaimStatus `json:"status"`
	Documents    []string    `json:"documents"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
}

type RiskAssessment struct {
	ID          string          `json:"id"`
	TenantID    string          `json:"tenant_id"`
	EquipmentID string          `json:"equipment_id"`
	RiskScore   float64         `json:"risk_score"`
	Factors     json.RawMessage `json:"factors"`
	AssessedAt  time.Time       `json:"assessed_at"`
}

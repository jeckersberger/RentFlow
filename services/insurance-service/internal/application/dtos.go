package application

import (
	"encoding/json"
	"time"

	"github.com/jeckersberger/rentflow/services/insurance-service/internal/domain"
)

type CreatePolicyRequest struct {
	PolicyNumber   string  `json:"policy_number"`
	Provider       string  `json:"provider"`
	Type           string  `json:"type"`
	CoverageAmount float64 `json:"coverage_amount"`
	Deductible     float64 `json:"deductible"`
	Premium        float64 `json:"premium"`
	StartDate      string  `json:"start_date"`
	EndDate        string  `json:"end_date"`
	Notes          string  `json:"notes"`
}

type UpdatePolicyRequest struct {
	Provider       string  `json:"provider"`
	CoverageAmount float64 `json:"coverage_amount"`
	Deductible     float64 `json:"deductible"`
	Premium        float64 `json:"premium"`
	Notes          string  `json:"notes"`
}

type PolicyResponse struct {
	ID             string    `json:"id"`
	PolicyNumber   string    `json:"policy_number"`
	Provider       string    `json:"provider"`
	Type           string    `json:"type"`
	CoverageAmount float64   `json:"coverage_amount"`
	Deductible     float64   `json:"deductible"`
	Premium        float64   `json:"premium"`
	StartDate      time.Time `json:"start_date"`
	EndDate        time.Time `json:"end_date"`
	Status         string    `json:"status"`
	Notes          string    `json:"notes"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type CreateClaimRequest struct {
	PolicyID     string    `json:"policy_id"`
	EquipmentID  string    `json:"equipment_id"`
	IncidentDate string    `json:"incident_date"`
	Description  string    `json:"description"`
	DamageAmount float64   `json:"damage_amount"`
	Documents    []string  `json:"documents"`
}

type UpdateClaimRequest struct {
	Description  string    `json:"description"`
	DamageAmount float64   `json:"damage_amount"`
	Documents    []string  `json:"documents"`
}

type ClaimResponse struct {
	ID           string    `json:"id"`
	PolicyID     string    `json:"policy_id"`
	EquipmentID  string    `json:"equipment_id"`
	IncidentDate time.Time `json:"incident_date"`
	Description  string    `json:"description"`
	DamageAmount float64   `json:"damage_amount"`
	ClaimAmount  float64   `json:"claim_amount"`
	Status       string    `json:"status"`
	Documents    []string  `json:"documents"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type ApproveClaimRequest struct {
	ApprovedAmount float64 `json:"approved_amount"`
	Notes          string  `json:"notes"`
}

type RejectClaimRequest struct {
	Reason string `json:"reason"`
}

type AssessRiskRequest struct {
	EquipmentID string  `json:"equipment_id"`
	Age         int     `json:"age"`
	Value       float64 `json:"value"`
	UsageHours  int     `json:"usage_hours"`
}

type RiskAssessmentResponse struct {
	ID          string          `json:"id"`
	EquipmentID string          `json:"equipment_id"`
	RiskScore   float64         `json:"risk_score"`
	Category    string          `json:"category"`
	Factors     json.RawMessage `json:"factors"`
	AssessedAt  time.Time       `json:"assessed_at"`
}

func PolicyToDTO(p *domain.Policy) *PolicyResponse {
	return &PolicyResponse{
		ID:             p.ID,
		PolicyNumber:   p.PolicyNumber,
		Provider:       p.Provider,
		Type:           string(p.Type),
		CoverageAmount: p.CoverageAmount,
		Deductible:     p.Deductible,
		Premium:        p.Premium,
		StartDate:      p.StartDate,
		EndDate:        p.EndDate,
		Status:         string(p.Status),
		Notes:          p.Notes,
		CreatedAt:      p.CreatedAt,
		UpdatedAt:      p.UpdatedAt,
	}
}

func ClaimToDTO(c *domain.Claim) *ClaimResponse {
	return &ClaimResponse{
		ID:           c.ID,
		PolicyID:     c.PolicyID,
		EquipmentID:  c.EquipmentID,
		IncidentDate: c.IncidentDate,
		Description:  c.Description,
		DamageAmount: c.DamageAmount,
		ClaimAmount:  c.ClaimAmount,
		Status:       string(c.Status),
		Documents:    c.Documents,
		CreatedAt:    c.CreatedAt,
		UpdatedAt:    c.UpdatedAt,
	}
}

func RiskAssessmentToDTO(r *domain.RiskAssessment) *RiskAssessmentResponse {
	score := r.RiskScore
	category := "Low"
	if score > 0.6 {
		category = "High"
	} else if score > 0.3 {
		category = "Medium"
	}

	return &RiskAssessmentResponse{
		ID:          r.ID,
		EquipmentID: r.EquipmentID,
		RiskScore:   r.RiskScore,
		Category:    category,
		Factors:     r.Factors,
		AssessedAt:  r.AssessedAt,
	}
}

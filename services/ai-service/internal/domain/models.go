package domain

import (
	"errors"
	"time"
)

var (
	ErrAIRequestNotFound         = errors.New("ai request not found")
	ErrAnonymizationRuleNotFound = errors.New("anonymization rule not found")
	ErrInvalidInput              = errors.New("invalid input")
	ErrTenantIDRequired          = errors.New("tenant ID required")
	ErrProviderUnavailable       = errors.New("provider unavailable")
)

type AIStatus string

const (
	AIStatusPending   AIStatus = "pending"
	AIStatusCompleted AIStatus = "completed"
	AIStatusFailed    AIStatus = "failed"
)

type AIRequest struct {
	ID               string    `json:"id"`
	TenantID         string    `json:"tenant_id"`
	Provider         string    `json:"provider"`
	Model            string    `json:"model"`
	Prompt           string    `json:"prompt"`
	AnonymizedPrompt string    `json:"anonymized_prompt"`
	Response         string    `json:"response"`
	TokensUsed       int       `json:"tokens_used"`
	CostEstimate     float64   `json:"cost_estimate"`
	LatencyMs        int64     `json:"latency_ms"`
	Status           AIStatus  `json:"status"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type AnonymizationRuleType string

const (
	RuleTypeName    AnonymizationRuleType = "name"
	RuleTypeEmail   AnonymizationRuleType = "email"
	RuleTypePhone   AnonymizationRuleType = "phone"
	RuleTypeAddress AnonymizationRuleType = "address"
	RuleTypeCompany AnonymizationRuleType = "company"
)

type AnonymizationRule struct {
	ID          string                `json:"id"`
	TenantID    string                `json:"tenant_id"`
	Pattern     string                `json:"pattern"`
	Replacement string                `json:"replacement"`
	Type        AnonymizationRuleType `json:"type"`
	CreatedAt   time.Time             `json:"created_at"`
}

type AnonymizationMapping struct {
	ID         string    `json:"id"`
	TenantID   string    `json:"tenant_id"`
	Original   string    `json:"original"`
	Anonymized string    `json:"anonymized"`
	CreatedAt  time.Time `json:"created_at"`
}

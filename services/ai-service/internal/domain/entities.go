package domain

import (
	"time"
)

// AIRequest represents an AI processing request
type AIRequest struct {
	ID               string
	TenantID         string
	ProviderID       string
	RequestType      RequestType
	InputText        string
	AnonymizedInput  *string
	OutputText       *string
	ModelUsed        *string
	TokensUsed       *int
	LatencyMs        *int
	Status           RequestStatus
	ErrorMessage     *string
	CreatedAt        time.Time
	CompletedAt      *time.Time
}

// RequestType represents types of AI requests
type RequestType string

const (
	RequestTypePriceOptimization     RequestType = "price_optimization"
	RequestTypeDemandForecast        RequestType = "demand_forecast"
	RequestTypeSmartAssetCreator     RequestType = "smart_asset_creator"
	RequestTypePredictiveMaintenance RequestType = "predictive_maintenance"
	RequestTypeGeneral               RequestType = "general"
)

// RequestStatus represents request status
type RequestStatus string

const (
	RequestStatusPending    RequestStatus = "pending"
	RequestStatusProcessing RequestStatus = "processing"
	RequestStatusCompleted  RequestStatus = "completed"
	RequestStatusFailed     RequestStatus = "failed"
)

// AIFeedback represents user feedback on an AI response
type AIFeedback struct {
	ID        string
	TenantID  string
	RequestID string
	Rating    int // 1-5
	Comment   *string
	IsCorrect *bool
	CreatedAt time.Time
}

// FewShotExample represents a few-shot learning example
type FewShotExample struct {
	ID            string
	TenantID      string
	RequestType   RequestType
	InputExample  string
	OutputExample string
	IsActive      bool
	UsageCount    int
	AvgRating     *float64
	CreatedAt     time.Time
}

// AIProvider represents an AI provider configuration
type AIProvider struct {
	ID          string
	TenantID    string
	Name        ProviderType
	APIEndpoint *string
	ModelName   string
	IsActive    bool
	Priority    int
	Config      map[string]interface{}
	CreatedAt   time.Time
}

// ProviderType represents types of AI providers
type ProviderType string

const (
	ProviderTypeClaude   ProviderType = "claude"
	ProviderTypeGPT4o    ProviderType = "gpt4o"
	ProviderTypeGemini   ProviderType = "gemini"
	ProviderTypeMistral  ProviderType = "mistral"
	ProviderTypeOllama   ProviderType = "ollama"
)

// AnonymizationPattern represents PII patterns to anonymize
type AnonymizationPattern struct {
	PatternType PatternType
	Pattern     string
	Replacement string
}

// PatternType represents types of PII patterns
type PatternType string

const (
	PatternTypeName       PatternType = "name"
	PatternTypeEmail      PatternType = "email"
	PatternTypeIBAN       PatternType = "iban"
	PatternTypePhone      PatternType = "phone"
	PatternTypeAddress    PatternType = "address"
	PatternTypeTaxID      PatternType = "tax_id"
	PatternTypeCustomerID PatternType = "customer_id"
)

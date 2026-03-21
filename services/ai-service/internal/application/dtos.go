package application

import (
	"time"

	"github.com/jeckersberger/rentflow/services/ai-service/internal/domain"
)

type PredictRequest struct {
	Prompt   string `json:"prompt"`
	Provider string `json:"provider,omitempty"`
	Model    string `json:"model,omitempty"`
}

type PredictResponse struct {
	ID           string    `json:"id"`
	Response     string    `json:"response"`
	TokensUsed   int       `json:"tokens_used"`
	CostEstimate float64   `json:"cost_estimate"`
	LatencyMs    int64     `json:"latency_ms"`
	CreatedAt    time.Time `json:"created_at"`
}

type ClassifyRequest struct {
	Text       string   `json:"text"`
	Categories []string `json:"categories"`
	Provider   string   `json:"provider,omitempty"`
	Model      string   `json:"model,omitempty"`
}

type ClassifyResponse struct {
	ID             string             `json:"id"`
	Classification map[string]float64 `json:"classification"`
	TopCategory    string             `json:"top_category"`
	Confidence     float64            `json:"confidence"`
	LatencyMs      int64              `json:"latency_ms"`
	CreatedAt      time.Time          `json:"created_at"`
}

type AnonymizeRequest struct {
	Text string `json:"text"`
}

type AnonymizeResponse struct {
	ID             string `json:"id"`
	AnonymizedText string `json:"anonymized_text"`
	MappingID      string `json:"mapping_id"`
}

type DeanonymizeRequest struct {
	Text      string `json:"text"`
	MappingID string `json:"mapping_id"`
}

type DeanonymizeResponse struct {
	ID               string `json:"id"`
	DeanonymizedText string `json:"deanonymized_text"`
}

type SuggestRequest struct {
	Context  string `json:"context"`
	Category string `json:"category"`
	Provider string `json:"provider,omitempty"`
	Model    string `json:"model,omitempty"`
}

type SuggestResponse struct {
	ID          string    `json:"id"`
	Suggestions []string  `json:"suggestions"`
	LatencyMs   int64     `json:"latency_ms"`
	CreatedAt   time.Time `json:"created_at"`
}

type ProviderInfo struct {
	Name    string `json:"name"`
	Model   string `json:"model"`
	Enabled bool   `json:"enabled"`
	Primary bool   `json:"primary"`
}

type UsageStats struct {
	TotalRequests    int64   `json:"total_requests"`
	CompletedCount   int64   `json:"completed_count"`
	FailedCount      int64   `json:"failed_count"`
	TotalTokens      int64   `json:"total_tokens"`
	TotalCost        float64 `json:"total_cost"`
	AverageLatencyMs float64 `json:"average_latency_ms"`
}

type AnonymizationRuleRequest struct {
	Pattern     string `json:"pattern"`
	Replacement string `json:"replacement"`
	Type        string `json:"type"`
}

type AnonymizationRuleResponse struct {
	ID          string    `json:"id"`
	Pattern     string    `json:"pattern"`
	Replacement string    `json:"replacement"`
	Type        string    `json:"type"`
	CreatedAt   time.Time `json:"created_at"`
}

func AIRequestToDTO(req *domain.AIRequest) *PredictResponse {
	return &PredictResponse{
		ID:           req.ID,
		Response:     req.Response,
		TokensUsed:   req.TokensUsed,
		CostEstimate: req.CostEstimate,
		LatencyMs:    req.LatencyMs,
		CreatedAt:    req.CreatedAt,
	}
}

func AnonymizationRuleToDTO(rule *domain.AnonymizationRule) *AnonymizationRuleResponse {
	return &AnonymizationRuleResponse{
		ID:          rule.ID,
		Pattern:     rule.Pattern,
		Replacement: rule.Replacement,
		Type:        string(rule.Type),
		CreatedAt:   rule.CreatedAt,
	}
}

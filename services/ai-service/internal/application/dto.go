package application

import (
	"time"

	"github.com/jeckersberger/rentflow/services/ai-service/internal/domain"
)

// AIRequestDTO represents an AI request in responses
type AIRequestDTO struct {
	ID              string `json:"id"`
	TenantID        string `json:"tenant_id"`
	ProviderID      string `json:"provider_id"`
	RequestType     string `json:"request_type"`
	InputText       string `json:"input_text"`
	AnonymizedInput *string `json:"anonymized_input,omitempty"`
	OutputText      *string `json:"output_text,omitempty"`
	ModelUsed       *string `json:"model_used,omitempty"`
	TokensUsed      *int `json:"tokens_used,omitempty"`
	LatencyMs       *int `json:"latency_ms,omitempty"`
	Status          string `json:"status"`
	ErrorMessage    *string `json:"error_message,omitempty"`
	CreatedAt       string `json:"created_at"`
	CompletedAt     *string `json:"completed_at,omitempty"`
}

// AIFeedbackDTO represents feedback in responses
type AIFeedbackDTO struct {
	ID        string `json:"id"`
	TenantID  string `json:"tenant_id"`
	RequestID string `json:"request_id"`
	Rating    int `json:"rating"`
	Comment   *string `json:"comment,omitempty"`
	IsCorrect *bool `json:"is_correct,omitempty"`
	CreatedAt string `json:"created_at"`
}

// FewShotExampleDTO represents a few-shot example in responses
type FewShotExampleDTO struct {
	ID            string `json:"id"`
	TenantID      string `json:"tenant_id"`
	RequestType   string `json:"request_type"`
	InputExample  string `json:"input_example"`
	OutputExample string `json:"output_example"`
	IsActive      bool `json:"is_active"`
	UsageCount    int `json:"usage_count"`
	AvgRating     *float64 `json:"avg_rating,omitempty"`
	CreatedAt     string `json:"created_at"`
}

// AIProviderDTO represents an AI provider in responses
type AIProviderDTO struct {
	ID          string `json:"id"`
	TenantID    string `json:"tenant_id"`
	Name        string `json:"name"`
	APIEndpoint *string `json:"api_endpoint,omitempty"`
	ModelName   string `json:"model_name"`
	IsActive    bool `json:"is_active"`
	Priority    int `json:"priority"`
	CreatedAt   string `json:"created_at"`
}

// DashboardDTO represents AI service statistics
type DashboardDTO struct {
	TotalRequests      int64 `json:"total_requests"`
	RequestsThisDay    int64 `json:"requests_this_day"`
	RequestsThisWeek   int64 `json:"requests_this_week"`
	AverageRating      *float64 `json:"average_rating,omitempty"`
	TopProvider        *string `json:"top_provider,omitempty"`
	AverageLatencyMs   *float64 `json:"average_latency_ms,omitempty"`
	SuccessRate        float64 `json:"success_rate"`
}

// ToAIRequestDTO converts domain model to DTO
func ToAIRequestDTO(r *domain.AIRequest) *AIRequestDTO {
	dto := &AIRequestDTO{
		ID:              r.ID,
		TenantID:        r.TenantID,
		ProviderID:      r.ProviderID,
		RequestType:     string(r.RequestType),
		InputText:       r.InputText,
		AnonymizedInput: r.AnonymizedInput,
		OutputText:      r.OutputText,
		ModelUsed:       r.ModelUsed,
		TokensUsed:      r.TokensUsed,
		LatencyMs:       r.LatencyMs,
		Status:          string(r.Status),
		ErrorMessage:    r.ErrorMessage,
		CreatedAt:       r.CreatedAt.Format(time.RFC3339),
	}
	if r.CompletedAt != nil {
		completedAt := r.CompletedAt.Format(time.RFC3339)
		dto.CompletedAt = &completedAt
	}
	return dto
}

// ToAIFeedbackDTO converts domain model to DTO
func ToAIFeedbackDTO(f *domain.AIFeedback) *AIFeedbackDTO {
	return &AIFeedbackDTO{
		ID:        f.ID,
		TenantID:  f.TenantID,
		RequestID: f.RequestID,
		Rating:    f.Rating,
		Comment:   f.Comment,
		IsCorrect: f.IsCorrect,
		CreatedAt: f.CreatedAt.Format(time.RFC3339),
	}
}

// ToFewShotExampleDTO converts domain model to DTO
func ToFewShotExampleDTO(e *domain.FewShotExample) *FewShotExampleDTO {
	return &FewShotExampleDTO{
		ID:            e.ID,
		TenantID:      e.TenantID,
		RequestType:   string(e.RequestType),
		InputExample:  e.InputExample,
		OutputExample: e.OutputExample,
		IsActive:      e.IsActive,
		UsageCount:    e.UsageCount,
		AvgRating:     e.AvgRating,
		CreatedAt:     e.CreatedAt.Format(time.RFC3339),
	}
}

// ToAIProviderDTO converts domain model to DTO
func ToAIProviderDTO(p *domain.AIProvider) *AIProviderDTO {
	return &AIProviderDTO{
		ID:          p.ID,
		TenantID:    p.TenantID,
		Name:        string(p.Name),
		APIEndpoint: p.APIEndpoint,
		ModelName:   p.ModelName,
		IsActive:    p.IsActive,
		Priority:    p.Priority,
		CreatedAt:   p.CreatedAt.Format(time.RFC3339),
	}
}

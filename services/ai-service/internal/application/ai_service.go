package application

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/ai-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/ai-service/internal/ports"
)

type AIService struct {
	aiRepo   ports.AIRequestRepository
	ruleRepo ports.AnonymizationRuleRepository
	mapRepo  ports.AnonymizationMappingRepository
	logger   logger.Logger
}

func NewAIService(
	aiRepo ports.AIRequestRepository,
	ruleRepo ports.AnonymizationRuleRepository,
	mapRepo ports.AnonymizationMappingRepository,
	log logger.Logger,
) *AIService {
	return &AIService{
		aiRepo:   aiRepo,
		ruleRepo: ruleRepo,
		mapRepo:  mapRepo,
		logger:   log,
	}
}

func (s *AIService) Predict(ctx context.Context, tenantID string, req PredictRequest) (*PredictResponse, error) {
	if tenantID == "" {
		return nil, domain.ErrTenantIDRequired
	}
	if req.Prompt == "" {
		return nil, domain.ErrInvalidInput
	}

	if req.Provider == "" {
		req.Provider = "claude"
	}
	if req.Model == "" {
		req.Model = "claude-3-sonnet"
	}

	start := time.Now()

	// Anonymize prompt before sending
	anonPrompt := req.Prompt
	rules, err := s.ruleRepo.ListByTenant(ctx, tenantID)
	if err != nil {
		s.logger.Error("Failed to fetch anonymization rules", err)
	}

	if len(rules) > 0 {
		anonPrompt, _, _ = s.anonymizeText(ctx, tenantID, req.Prompt, rules)
	}

	// Simulate AI response (in production, call actual AI provider)
	response := "This is a simulated response to: " + anonPrompt
	tokensUsed := len(strings.Fields(req.Prompt)) + len(strings.Fields(response))
	costEstimate := float64(tokensUsed) * 0.0001

	latency := time.Since(start).Milliseconds()

	aiReq := &domain.AIRequest{
		ID:               fmt.Sprintf("air_%d", time.Now().UnixNano()),
		TenantID:         tenantID,
		Provider:         req.Provider,
		Model:            req.Model,
		Prompt:           req.Prompt,
		AnonymizedPrompt: anonPrompt,
		Response:         response,
		TokensUsed:       tokensUsed,
		CostEstimate:     costEstimate,
		LatencyMs:        latency,
		Status:           domain.AIStatusCompleted,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := s.aiRepo.Create(ctx, aiReq); err != nil {
		s.logger.Error("Failed to create AI request", err)
		return nil, err
	}

	return &PredictResponse{
		ID:           aiReq.ID,
		Response:     response,
		TokensUsed:   tokensUsed,
		CostEstimate: costEstimate,
		LatencyMs:    latency,
		CreatedAt:    aiReq.CreatedAt,
	}, nil
}

func (s *AIService) Classify(ctx context.Context, tenantID string, req ClassifyRequest) (*ClassifyResponse, error) {
	if tenantID == "" {
		return nil, domain.ErrTenantIDRequired
	}
	if req.Text == "" || len(req.Categories) == 0 {
		return nil, domain.ErrInvalidInput
	}

	if req.Provider == "" {
		req.Provider = "claude"
	}
	if req.Model == "" {
		req.Model = "claude-3-sonnet"
	}

	start := time.Now()

	// Simulate classification (in production, call actual AI provider)
	classification := make(map[string]float64)
	baseScore := 1.0 / float64(len(req.Categories))
	for _, cat := range req.Categories {
		classification[cat] = baseScore
	}

	topCategory := req.Categories[0]
	confidence := baseScore
	latency := time.Since(start).Milliseconds()

	aiReq := &domain.AIRequest{
		ID:        fmt.Sprintf("air_%d", time.Now().UnixNano()),
		TenantID:  tenantID,
		Provider:  req.Provider,
		Model:     req.Model,
		Prompt:    req.Text,
		Response:  fmt.Sprintf("Classification: %s (%.2f%%)", topCategory, confidence*100),
		Status:    domain.AIStatusCompleted,
		LatencyMs: latency,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.aiRepo.Create(ctx, aiReq); err != nil {
		s.logger.Error("Failed to create classification request", err)
		return nil, err
	}

	return &ClassifyResponse{
		ID:             aiReq.ID,
		Classification: classification,
		TopCategory:    topCategory,
		Confidence:     confidence,
		LatencyMs:      latency,
		CreatedAt:      aiReq.CreatedAt,
	}, nil
}

func (s *AIService) Anonymize(ctx context.Context, tenantID, text string) (*AnonymizeResponse, error) {
	if tenantID == "" {
		return nil, domain.ErrTenantIDRequired
	}
	if text == "" {
		return nil, domain.ErrInvalidInput
	}

	rules, err := s.ruleRepo.ListByTenant(ctx, tenantID)
	if err != nil {
		s.logger.Error("Failed to fetch anonymization rules", err)
		return nil, err
	}

	anonText, mappingID, _ := s.anonymizeText(ctx, tenantID, text, rules)

	return &AnonymizeResponse{
		ID:             fmt.Sprintf("air_%d", time.Now().UnixNano()),
		AnonymizedText: anonText,
		MappingID:      mappingID,
	}, nil
}

func (s *AIService) Deanonymize(ctx context.Context, tenantID, text, mappingID string) (*DeanonymizeResponse, error) {
	if tenantID == "" {
		return nil, domain.ErrTenantIDRequired
	}
	if text == "" || mappingID == "" {
		return nil, domain.ErrInvalidInput
	}

	// In production, would fetch mapping and deanonymize
	// For now, return same text
	return &DeanonymizeResponse{
		ID:               fmt.Sprintf("air_%d", time.Now().UnixNano()),
		DeanonymizedText: text,
	}, nil
}

func (s *AIService) Suggest(ctx context.Context, tenantID string, req SuggestRequest) (*SuggestResponse, error) {
	if tenantID == "" {
		return nil, domain.ErrTenantIDRequired
	}
	if req.Context == "" {
		return nil, domain.ErrInvalidInput
	}

	if req.Provider == "" {
		req.Provider = "claude"
	}
	if req.Model == "" {
		req.Model = "claude-3-sonnet"
	}

	start := time.Now()

	// Simulate suggestions
	suggestions := []string{
		"Suggestion 1 for " + req.Category,
		"Suggestion 2 for " + req.Category,
		"Suggestion 3 for " + req.Category,
	}

	latency := time.Since(start).Milliseconds()

	aiReq := &domain.AIRequest{
		ID:        fmt.Sprintf("air_%d", time.Now().UnixNano()),
		TenantID:  tenantID,
		Provider:  req.Provider,
		Model:     req.Model,
		Prompt:    req.Context,
		Response:  strings.Join(suggestions, "; "),
		Status:    domain.AIStatusCompleted,
		LatencyMs: latency,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.aiRepo.Create(ctx, aiReq); err != nil {
		s.logger.Error("Failed to create suggestion request", err)
		return nil, err
	}

	return &SuggestResponse{
		ID:          aiReq.ID,
		Suggestions: suggestions,
		LatencyMs:   latency,
		CreatedAt:   aiReq.CreatedAt,
	}, nil
}

func (s *AIService) GetProviders(ctx context.Context) ([]ProviderInfo, error) {
	// Return configured providers
	return []ProviderInfo{
		{Name: "claude", Model: "claude-3-sonnet", Enabled: true, Primary: true},
		{Name: "openai", Model: "gpt-4", Enabled: true, Primary: false},
		{Name: "gemini", Model: "gemini-pro", Enabled: true, Primary: false},
		{Name: "ollama", Model: "llama2", Enabled: false, Primary: false},
	}, nil
}

func (s *AIService) GetUsageStats(ctx context.Context, tenantID string) (*UsageStats, error) {
	if tenantID == "" {
		return nil, domain.ErrTenantIDRequired
	}

	requests, err := s.aiRepo.ListByTenant(ctx, tenantID)
	if err != nil {
		s.logger.Error("Failed to fetch usage stats", err)
		return nil, err
	}

	stats := &UsageStats{}
	for _, req := range requests {
		stats.TotalRequests++
		if req.Status == domain.AIStatusCompleted {
			stats.CompletedCount++
		} else if req.Status == domain.AIStatusFailed {
			stats.FailedCount++
		}
		stats.TotalTokens += int64(req.TokensUsed)
		stats.TotalCost += req.CostEstimate
	}

	if stats.CompletedCount > 0 {
		stats.AverageLatencyMs = 0
		for _, req := range requests {
			if req.Status == domain.AIStatusCompleted {
				stats.AverageLatencyMs += float64(req.LatencyMs)
			}
		}
		stats.AverageLatencyMs /= float64(stats.CompletedCount)
	}

	return stats, nil
}

func (s *AIService) anonymizeText(ctx context.Context, tenantID, text string, rules []*domain.AnonymizationRule) (string, string, error) {
	anonText := text
	mappingID := fmt.Sprintf("map_%d", time.Now().UnixNano())

	for _, rule := range rules {
		re, err := regexp.Compile(rule.Pattern)
		if err != nil {
			s.logger.Error("Invalid regex pattern", err)
			continue
		}

		anonText = re.ReplaceAllString(anonText, rule.Replacement)
	}

	mapping := &domain.AnonymizationMapping{
		ID:         mappingID,
		TenantID:   tenantID,
		Original:   text,
		Anonymized: anonText,
		CreatedAt:  time.Now(),
	}

	if err := s.mapRepo.Create(ctx, mapping); err != nil {
		s.logger.Error("Failed to save anonymization mapping", err)
	}

	return anonText, mappingID, nil
}

func (s *AIService) CreateAnonymizationRule(ctx context.Context, tenantID string, req AnonymizationRuleRequest) (*AnonymizationRuleResponse, error) {
	if tenantID == "" {
		return nil, domain.ErrTenantIDRequired
	}
	if req.Pattern == "" || req.Replacement == "" {
		return nil, domain.ErrInvalidInput
	}

	rule := &domain.AnonymizationRule{
		ID:          fmt.Sprintf("rule_%d", time.Now().UnixNano()),
		TenantID:    tenantID,
		Pattern:     req.Pattern,
		Replacement: req.Replacement,
		Type:        domain.AnonymizationRuleType(req.Type),
		CreatedAt:   time.Now(),
	}

	if err := s.ruleRepo.Create(ctx, rule); err != nil {
		s.logger.Error("Failed to create anonymization rule", err)
		return nil, err
	}

	return AnonymizationRuleToDTO(rule), nil
}

func (s *AIService) GetAnonymizationRules(ctx context.Context, tenantID string) ([]*AnonymizationRuleResponse, error) {
	if tenantID == "" {
		return nil, domain.ErrTenantIDRequired
	}

	rules, err := s.ruleRepo.ListByTenant(ctx, tenantID)
	if err != nil {
		s.logger.Error("Failed to fetch anonymization rules", err)
		return nil, err
	}

	dtos := make([]*AnonymizationRuleResponse, len(rules))
	for i, rule := range rules {
		dtos[i] = AnonymizationRuleToDTO(rule)
	}

	return dtos, nil
}

func (s *AIService) DeleteAnonymizationRule(ctx context.Context, tenantID, ruleID string) error {
	if tenantID == "" || ruleID == "" {
		return domain.ErrInvalidInput
	}

	return s.ruleRepo.Delete(ctx, tenantID, ruleID)
}

package application

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/ai-service/internal/domain"
)

// ProviderRequest represents a request to an AI provider
type ProviderRequest struct {
	SystemPrompt     string
	UserPrompt       string
	MaxTokens        int
	Temperature      float64
	FewShotExamples  []FewShotPair
}

// FewShotPair represents a few-shot example pair
type FewShotPair struct {
	Input  string
	Output string
}

// ProviderResponse represents a response from an AI provider
type ProviderResponse struct {
	Text       string
	TokensUsed int
	Model      string
	LatencyMs  int
}

// AIProvider defines the interface for AI providers
type AIProvider interface {
	Name() string
	Complete(ctx context.Context, req ProviderRequest) (*ProviderResponse, error)
	IsAvailable() bool
}

// ProviderService manages AI provider operations
type ProviderService struct {
	providers map[string]AIProvider
	logger    logger.Logger
}

// NewProviderService creates a new provider service
func NewProviderService(log logger.Logger) *ProviderService {
	return &ProviderService{
		providers: make(map[string]AIProvider),
		logger:    log,
	}
}

// RegisterProvider registers an AI provider
func (s *ProviderService) RegisterProvider(providerType domain.ProviderType, config map[string]interface{}, apiKey string) error {
	var provider AIProvider
	finalAPIKey := apiKey

	switch providerType {
	case domain.ProviderTypeClaude:
		if finalAPIKey == "" {
			finalAPIKey = os.Getenv("ANTHROPIC_API_KEY")
		}
		if finalAPIKey == "" {
			return fmt.Errorf("API key not configured for provider Claude (set ANTHROPIC_API_KEY)")
		}
		provider = NewClaudeProvider(finalAPIKey, config, s.logger)
	case domain.ProviderTypeGPT4o:
		if finalAPIKey == "" {
			finalAPIKey = os.Getenv("OPENAI_API_KEY")
		}
		if finalAPIKey == "" {
			return fmt.Errorf("API key not configured for provider GPT-4o (set OPENAI_API_KEY)")
		}
		provider = NewGPT4oProvider(finalAPIKey, config, s.logger)
	case domain.ProviderTypeGemini:
		if finalAPIKey == "" {
			finalAPIKey = os.Getenv("GOOGLE_AI_KEY")
		}
		if finalAPIKey == "" {
			return fmt.Errorf("API key not configured for provider Gemini (set GOOGLE_AI_KEY)")
		}
		provider = NewGeminiProvider(finalAPIKey, config, s.logger)
	case domain.ProviderTypeMistral:
		if finalAPIKey == "" {
			finalAPIKey = os.Getenv("MISTRAL_API_KEY")
		}
		if finalAPIKey == "" {
			return fmt.Errorf("API key not configured for provider Mistral (set MISTRAL_API_KEY)")
		}
		provider = NewMistralProvider(finalAPIKey, config, s.logger)
	case domain.ProviderTypeOllama:
		provider = NewOllamaProvider(config, s.logger)
	default:
		return errors.New("unsupported provider type")
	}

	if !provider.IsAvailable() {
		return domain.ErrProviderNotAvailable
	}

	s.providers[string(providerType)] = provider
	return nil
}

// GetProvider gets a registered provider by type
func (s *ProviderService) GetProvider(providerType string) (AIProvider, error) {
	provider, exists := s.providers[providerType]
	if !exists {
		return nil, domain.ErrAIProviderNotFound
	}
	if !provider.IsAvailable() {
		return nil, domain.ErrProviderNotAvailable
	}
	return provider, nil
}

// Complete sends a request to the specified provider
func (s *ProviderService) Complete(ctx context.Context, providerType string, req ProviderRequest) (*ProviderResponse, error) {
	provider, err := s.GetProvider(providerType)
	if err != nil {
		return nil, err
	}

	start := time.Now()
	response, err := provider.Complete(ctx, req)
	if err != nil {
		return nil, err
	}

	if response != nil {
		response.LatencyMs = int(time.Since(start).Milliseconds())
	}

	return response, nil
}

// ClaudeProvider implements AIProvider for Claude
type ClaudeProvider struct {
	apiKey string
	config map[string]interface{}
	logger logger.Logger
	model  string
}

// NewClaudeProvider creates a new Claude provider
func NewClaudeProvider(apiKey string, config map[string]interface{}, log logger.Logger) *ClaudeProvider {
	model := "claude-sonnet-4-20250514"
	if m, ok := config["model"].(string); ok {
		model = m
	}
	return &ClaudeProvider{
		apiKey: apiKey,
		config: config,
		logger: log,
		model:  model,
	}
}

// Name returns the provider name
func (p *ClaudeProvider) Name() string { return "claude" }

// IsAvailable checks if provider is available
func (p *ClaudeProvider) IsAvailable() bool { return p.apiKey != "" }

// Complete sends request to Claude API
func (p *ClaudeProvider) Complete(ctx context.Context, req ProviderRequest) (*ProviderResponse, error) {
	const apiURL = "https://api.anthropic.com/v1/messages"

	// Build request body
	messages := []map[string]interface{}{
		{
			"role":    "user",
			"content": req.UserPrompt,
		},
	}

	requestBody := map[string]interface{}{
		"model":       p.model,
		"max_tokens":  req.MaxTokens,
		"temperature": req.Temperature,
		"system":      req.SystemPrompt,
		"messages":    messages,
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("x-api-key", p.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")
	httpReq.Header.Set("content-type", "application/json")

	// Execute request with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to call Claude API: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for API errors
	if resp.StatusCode != http.StatusOK {
		var errResp map[string]interface{}
		json.Unmarshal(respBody, &errResp)
		return nil, fmt.Errorf("Claude API error (status %d): %v", resp.StatusCode, errResp)
	}

	// Parse response
	var apiResp struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
		Usage struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	}

	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse Claude response: %w", err)
	}

	if len(apiResp.Content) == 0 {
		return nil, errors.New("no content in Claude response")
	}

	return &ProviderResponse{
		Text:       apiResp.Content[0].Text,
		TokensUsed: apiResp.Usage.InputTokens + apiResp.Usage.OutputTokens,
		Model:      p.model,
	}, nil
}

// GPT4oProvider implements AIProvider for GPT-4o
type GPT4oProvider struct {
	apiKey string
	config map[string]interface{}
	logger logger.Logger
	model  string
}

// NewGPT4oProvider creates a new GPT-4o provider
func NewGPT4oProvider(apiKey string, config map[string]interface{}, log logger.Logger) *GPT4oProvider {
	model := "gpt-4o"
	if m, ok := config["model"].(string); ok {
		model = m
	}
	return &GPT4oProvider{
		apiKey: apiKey,
		config: config,
		logger: log,
		model:  model,
	}
}

// Name returns the provider name
func (p *GPT4oProvider) Name() string { return "gpt4o" }

// IsAvailable checks if provider is available
func (p *GPT4oProvider) IsAvailable() bool { return p.apiKey != "" }

// Complete sends request to OpenAI API
func (p *GPT4oProvider) Complete(ctx context.Context, req ProviderRequest) (*ProviderResponse, error) {
	const apiURL = "https://api.openai.com/v1/chat/completions"

	// Build messages
	messages := []map[string]interface{}{
		{
			"role":    "system",
			"content": req.SystemPrompt,
		},
		{
			"role":    "user",
			"content": req.UserPrompt,
		},
	}

	requestBody := map[string]interface{}{
		"model":       p.model,
		"messages":    messages,
		"max_tokens":  req.MaxTokens,
		"temperature": req.Temperature,
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	// Execute request with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to call OpenAI API: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for API errors
	if resp.StatusCode != http.StatusOK {
		var errResp map[string]interface{}
		json.Unmarshal(respBody, &errResp)
		return nil, fmt.Errorf("OpenAI API error (status %d): %v", resp.StatusCode, errResp)
	}

	// Parse response
	var apiResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			TotalTokens int `json:"total_tokens"`
		} `json:"usage"`
	}

	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse OpenAI response: %w", err)
	}

	if len(apiResp.Choices) == 0 {
		return nil, errors.New("no choices in OpenAI response")
	}

	return &ProviderResponse{
		Text:       apiResp.Choices[0].Message.Content,
		TokensUsed: apiResp.Usage.TotalTokens,
		Model:      p.model,
	}, nil
}

// GeminiProvider implements AIProvider for Gemini
type GeminiProvider struct {
	apiKey string
	config map[string]interface{}
	logger logger.Logger
	model  string
}

// NewGeminiProvider creates a new Gemini provider
func NewGeminiProvider(apiKey string, config map[string]interface{}, log logger.Logger) *GeminiProvider {
	model := "gemini-pro"
	if m, ok := config["model"].(string); ok {
		model = m
	}
	return &GeminiProvider{
		apiKey: apiKey,
		config: config,
		logger: log,
		model:  model,
	}
}

// Name returns the provider name
func (p *GeminiProvider) Name() string { return "gemini" }

// IsAvailable checks if provider is available
func (p *GeminiProvider) IsAvailable() bool { return p.apiKey != "" }

// Complete sends request to Google Gemini API
func (p *GeminiProvider) Complete(ctx context.Context, req ProviderRequest) (*ProviderResponse, error) {
	apiURL := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", p.model, p.apiKey)

	// Build request body
	requestBody := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"role": "user",
				"parts": []map[string]interface{}{
					{
						"text": req.UserPrompt,
					},
				},
			},
		},
		"generationConfig": map[string]interface{}{
			"maxOutputTokens": req.MaxTokens,
			"temperature":     req.Temperature,
		},
	}

	if req.SystemPrompt != "" {
		requestBody["systemInstruction"] = map[string]interface{}{
			"parts": []map[string]interface{}{
				{
					"text": req.SystemPrompt,
				},
			},
		}
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")

	// Execute request with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to call Gemini API: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for API errors
	if resp.StatusCode != http.StatusOK {
		var errResp map[string]interface{}
		json.Unmarshal(respBody, &errResp)
		return nil, fmt.Errorf("Gemini API error (status %d): %v", resp.StatusCode, errResp)
	}

	// Parse response
	var apiResp struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
		UsageMetadata struct {
			PromptTokenCount     int `json:"promptTokenCount"`
			CandidatesTokenCount int `json:"candidatesTokenCount"`
		} `json:"usageMetadata"`
	}

	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse Gemini response: %w", err)
	}

	if len(apiResp.Candidates) == 0 || len(apiResp.Candidates[0].Content.Parts) == 0 {
		return nil, errors.New("no content in Gemini response")
	}

	return &ProviderResponse{
		Text:       apiResp.Candidates[0].Content.Parts[0].Text,
		TokensUsed: apiResp.UsageMetadata.PromptTokenCount + apiResp.UsageMetadata.CandidatesTokenCount,
		Model:      p.model,
	}, nil
}

// MistralProvider implements AIProvider for Mistral
type MistralProvider struct {
	apiKey string
	config map[string]interface{}
	logger logger.Logger
	model  string
}

// NewMistralProvider creates a new Mistral provider
func NewMistralProvider(apiKey string, config map[string]interface{}, log logger.Logger) *MistralProvider {
	model := "mistral-large-latest"
	if m, ok := config["model"].(string); ok {
		model = m
	}
	return &MistralProvider{
		apiKey: apiKey,
		config: config,
		logger: log,
		model:  model,
	}
}

// Name returns the provider name
func (p *MistralProvider) Name() string { return "mistral" }

// IsAvailable checks if provider is available
func (p *MistralProvider) IsAvailable() bool { return p.apiKey != "" }

// Complete sends request to Mistral API
func (p *MistralProvider) Complete(ctx context.Context, req ProviderRequest) (*ProviderResponse, error) {
	const apiURL = "https://api.mistral.ai/v1/chat/completions"

	// Build messages
	messages := []map[string]interface{}{
		{
			"role":    "system",
			"content": req.SystemPrompt,
		},
		{
			"role":    "user",
			"content": req.UserPrompt,
		},
	}

	requestBody := map[string]interface{}{
		"model":       p.model,
		"messages":    messages,
		"max_tokens":  req.MaxTokens,
		"temperature": req.Temperature,
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	// Execute request with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to call Mistral API: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for API errors
	if resp.StatusCode != http.StatusOK {
		var errResp map[string]interface{}
		json.Unmarshal(respBody, &errResp)
		return nil, fmt.Errorf("Mistral API error (status %d): %v", resp.StatusCode, errResp)
	}

	// Parse response
	var apiResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
		} `json:"usage"`
	}

	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse Mistral response: %w", err)
	}

	if len(apiResp.Choices) == 0 {
		return nil, errors.New("no choices in Mistral response")
	}

	return &ProviderResponse{
		Text:       apiResp.Choices[0].Message.Content,
		TokensUsed: apiResp.Usage.PromptTokens + apiResp.Usage.CompletionTokens,
		Model:      p.model,
	}, nil
}

// OllamaProvider implements AIProvider for Ollama
type OllamaProvider struct {
	endpoint string
	config   map[string]interface{}
	logger   logger.Logger
	model    string
}

// NewOllamaProvider creates a new Ollama provider
func NewOllamaProvider(config map[string]interface{}, log logger.Logger) *OllamaProvider {
	endpoint := os.Getenv("OLLAMA_URL")
	if endpoint == "" {
		endpoint = "http://localhost:11434"
	}
	if e, ok := config["endpoint"].(string); ok {
		endpoint = e
	}

	model := "mistral"
	if m, ok := config["model"].(string); ok {
		model = m
	}

	return &OllamaProvider{
		endpoint: endpoint,
		config:   config,
		logger:   log,
		model:    model,
	}
}

// Name returns the provider name
func (p *OllamaProvider) Name() string { return "ollama" }

// IsAvailable checks if provider is available
func (p *OllamaProvider) IsAvailable() bool { return p.endpoint != "" }

// Complete sends request to Ollama API
func (p *OllamaProvider) Complete(ctx context.Context, req ProviderRequest) (*ProviderResponse, error) {
	apiURL := fmt.Sprintf("%s/api/generate", p.endpoint)

	// Combine system and user prompts
	prompt := req.UserPrompt
	if req.SystemPrompt != "" {
		prompt = req.SystemPrompt + "\n\n" + req.UserPrompt
	}

	requestBody := map[string]interface{}{
		"model":  p.model,
		"prompt": prompt,
		"stream": false,
	}

	// Add optional parameters if set
	if req.Temperature > 0 {
		requestBody["temperature"] = req.Temperature
	}
	if req.MaxTokens > 0 {
		requestBody["num_predict"] = req.MaxTokens
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")

	// Execute request with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to call Ollama API: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for API errors
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Ollama API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	// Parse response
	var apiResp struct {
		Response string `json:"response"`
		TotalDuration    int64 `json:"total_duration"`
		LoadDuration     int64 `json:"load_duration"`
		PromptEvalCount  int   `json:"prompt_eval_count"`
		EvalCount        int   `json:"eval_count"`
	}

	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse Ollama response: %w", err)
	}

	return &ProviderResponse{
		Text:       apiResp.Response,
		TokensUsed: apiResp.PromptEvalCount + apiResp.EvalCount,
		Model:      p.model,
	}, nil
}

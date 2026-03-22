package application

import (
	"context"
	"errors"
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

	switch providerType {
	case domain.ProviderTypeClaude:
		provider = NewClaudeProvider(apiKey, config, s.logger)
	case domain.ProviderTypeGPT4o:
		provider = NewGPT4oProvider(apiKey, config, s.logger)
	case domain.ProviderTypeGemini:
		provider = NewGeminiProvider(apiKey, config, s.logger)
	case domain.ProviderTypeMistral:
		provider = NewMistralProvider(apiKey, config, s.logger)
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
}

// NewClaudeProvider creates a new Claude provider
func NewClaudeProvider(apiKey string, config map[string]interface{}, log logger.Logger) *ClaudeProvider {
	return &ClaudeProvider{
		apiKey: apiKey,
		config: config,
		logger: log,
	}
}

// Name returns the provider name
func (p *ClaudeProvider) Name() string { return "claude" }

// IsAvailable checks if provider is available
func (p *ClaudeProvider) IsAvailable() bool { return p.apiKey != "" }

// Complete sends request to Claude API (stub for now)
func (p *ClaudeProvider) Complete(ctx context.Context, req ProviderRequest) (*ProviderResponse, error) {
	// Stub implementation - in production, call Anthropic API
	return &ProviderResponse{
		Text:       "Claude response",
		TokensUsed: 100,
		Model:      "claude-3-opus",
		LatencyMs:  500,
	}, nil
}

// GPT4oProvider implements AIProvider for GPT-4o
type GPT4oProvider struct {
	apiKey string
	config map[string]interface{}
	logger logger.Logger
}

// NewGPT4oProvider creates a new GPT-4o provider
func NewGPT4oProvider(apiKey string, config map[string]interface{}, log logger.Logger) *GPT4oProvider {
	return &GPT4oProvider{
		apiKey: apiKey,
		config: config,
		logger: log,
	}
}

// Name returns the provider name
func (p *GPT4oProvider) Name() string { return "gpt4o" }

// IsAvailable checks if provider is available
func (p *GPT4oProvider) IsAvailable() bool { return p.apiKey != "" }

// Complete sends request to OpenAI API (stub for now)
func (p *GPT4oProvider) Complete(ctx context.Context, req ProviderRequest) (*ProviderResponse, error) {
	return &ProviderResponse{
		Text:       "GPT-4o response",
		TokensUsed: 150,
		Model:      "gpt-4o",
		LatencyMs:  600,
	}, nil
}

// GeminiProvider implements AIProvider for Gemini
type GeminiProvider struct {
	apiKey string
	config map[string]interface{}
	logger logger.Logger
}

// NewGeminiProvider creates a new Gemini provider
func NewGeminiProvider(apiKey string, config map[string]interface{}, log logger.Logger) *GeminiProvider {
	return &GeminiProvider{
		apiKey: apiKey,
		config: config,
		logger: log,
	}
}

// Name returns the provider name
func (p *GeminiProvider) Name() string { return "gemini" }

// IsAvailable checks if provider is available
func (p *GeminiProvider) IsAvailable() bool { return p.apiKey != "" }

// Complete sends request to Google Gemini API (stub for now)
func (p *GeminiProvider) Complete(ctx context.Context, req ProviderRequest) (*ProviderResponse, error) {
	return &ProviderResponse{
		Text:       "Gemini response",
		TokensUsed: 120,
		Model:      "gemini-pro",
		LatencyMs:  400,
	}, nil
}

// MistralProvider implements AIProvider for Mistral
type MistralProvider struct {
	apiKey string
	config map[string]interface{}
	logger logger.Logger
}

// NewMistralProvider creates a new Mistral provider
func NewMistralProvider(apiKey string, config map[string]interface{}, log logger.Logger) *MistralProvider {
	return &MistralProvider{
		apiKey: apiKey,
		config: config,
		logger: log,
	}
}

// Name returns the provider name
func (p *MistralProvider) Name() string { return "mistral" }

// IsAvailable checks if provider is available
func (p *MistralProvider) IsAvailable() bool { return p.apiKey != "" }

// Complete sends request to Mistral API (stub for now)
func (p *MistralProvider) Complete(ctx context.Context, req ProviderRequest) (*ProviderResponse, error) {
	return &ProviderResponse{
		Text:       "Mistral response",
		TokensUsed: 130,
		Model:      "mistral-large",
		LatencyMs:  450,
	}, nil
}

// OllamaProvider implements AIProvider for Ollama
type OllamaProvider struct {
	endpoint string
	config   map[string]interface{}
	logger   logger.Logger
}

// NewOllamaProvider creates a new Ollama provider
func NewOllamaProvider(config map[string]interface{}, log logger.Logger) *OllamaProvider {
	endpoint := "http://localhost:11434"
	if e, ok := config["endpoint"].(string); ok {
		endpoint = e
	}
	return &OllamaProvider{
		endpoint: endpoint,
		config:   config,
		logger:   log,
	}
}

// Name returns the provider name
func (p *OllamaProvider) Name() string { return "ollama" }

// IsAvailable checks if provider is available
func (p *OllamaProvider) IsAvailable() bool { return p.endpoint != "" }

// Complete sends request to Ollama API (stub for now)
func (p *OllamaProvider) Complete(ctx context.Context, req ProviderRequest) (*ProviderResponse, error) {
	return &ProviderResponse{
		Text:       "Ollama response",
		TokensUsed: 110,
		Model:      "mistral",
		LatencyMs:  300,
	}, nil
}

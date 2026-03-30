package application

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

const (
	defaultClaudeModel   = "claude-sonnet-4-20250514"
	defaultClaudeBaseURL = "https://api.anthropic.com"
	defaultMaxTokens     = 1024
	anthropicVersion     = "2023-06-01"
)

// ClaudeProvider calls the Anthropic Claude API via plain net/http.
type ClaudeProvider struct {
	apiKey  string
	model   string
	baseURL string
	client  *http.Client
	logger  zerolog.Logger
}

// AIRequest describes the input for a Claude completion call.
type AIRequest struct {
	SystemPrompt string
	UserMessage  string
	MaxTokens    int
}

// AIResponse holds the result of a Claude completion call.
type AIResponse struct {
	Content    string `json:"content"`
	TokensUsed int    `json:"tokens_used"`
	Model      string `json:"model"`
}

// NewClaudeProvider creates a new ClaudeProvider. If model is empty, it
// defaults to claude-sonnet-4-20250514.
func NewClaudeProvider(apiKey, model string, logger zerolog.Logger) *ClaudeProvider {
	if model == "" {
		model = defaultClaudeModel
	}
	return &ClaudeProvider{
		apiKey:  apiKey,
		model:   model,
		baseURL: defaultClaudeBaseURL,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
		logger: logger.With().Str("component", "claude_provider").Logger(),
	}
}

// IsConfigured returns true when an API key has been set.
func (p *ClaudeProvider) IsConfigured() bool {
	return p.apiKey != ""
}

// ModelName returns the configured model identifier.
func (p *ClaudeProvider) ModelName() string {
	return p.model
}

// Complete sends a message to the Anthropic Messages API and returns the
// assistant response. It uses plain net/http — no external SDK required.
func (p *ClaudeProvider) Complete(ctx context.Context, req AIRequest) (*AIResponse, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("KI nicht konfiguriert")
	}

	maxTokens := req.MaxTokens
	if maxTokens <= 0 {
		maxTokens = defaultMaxTokens
	}

	// Build the Anthropic Messages API request body.
	body := claudeRequestBody{
		Model:    p.model,
		MaxTokens: maxTokens,
		Messages: []claudeMessage{
			{Role: "user", Content: req.UserMessage},
		},
	}
	if req.SystemPrompt != "" {
		body.System = req.SystemPrompt
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	url := p.baseURL + "/v1/messages"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("create http request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", p.apiKey)
	httpReq.Header.Set("anthropic-version", anthropicVersion)

	p.logger.Debug().
		Str("model", p.model).
		Int("max_tokens", maxTokens).
		Msg("calling Claude API")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		p.logger.Error().
			Int("status", resp.StatusCode).
			Str("body", truncate(string(respBody), 500)).
			Msg("Claude API error")
		return nil, fmt.Errorf("Claude API Fehler (HTTP %d): %s", resp.StatusCode, truncate(string(respBody), 200))
	}

	var apiResp claudeResponseBody
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	// Extract the text content from the first content block.
	text := ""
	for _, block := range apiResp.Content {
		if block.Type == "text" {
			text = block.Text
			break
		}
	}

	tokensUsed := apiResp.Usage.InputTokens + apiResp.Usage.OutputTokens

	p.logger.Info().
		Int("input_tokens", apiResp.Usage.InputTokens).
		Int("output_tokens", apiResp.Usage.OutputTokens).
		Str("model", apiResp.Model).
		Msg("Claude API call completed")

	return &AIResponse{
		Content:    text,
		TokensUsed: tokensUsed,
		Model:      apiResp.Model,
	}, nil
}

// ---------------------------------------------------------------------------
// Internal types for Anthropic Messages API JSON serialization
// ---------------------------------------------------------------------------

type claudeRequestBody struct {
	Model     string           `json:"model"`
	MaxTokens int              `json:"max_tokens"`
	System    string           `json:"system,omitempty"`
	Messages  []claudeMessage  `json:"messages"`
}

type claudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type claudeResponseBody struct {
	Content []claudeContentBlock `json:"content"`
	Model   string               `json:"model"`
	Usage   claudeUsage          `json:"usage"`
}

type claudeContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type claudeUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// truncate shortens a string to maxLen characters for safe logging.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

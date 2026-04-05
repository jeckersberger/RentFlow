package application

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/services/ai/internal/domain"
)

const (
	defaultOllamaBaseURL = "http://localhost:11434"
	defaultAIModelName   = "gemma3"
	defaultMaxItems      = 80
	ollamaTimeout        = 12 * time.Second
	promptVersion        = "v1"

	packlistSystemPrompt = "Du bist Disponent fuer Eventtechnik. " +
		"Du erhaeltst ein Projekt mit Titel, Zeitraum und Notizen sowie eine Kandidatenliste " +
		"mit Equipment-Typen und verfuegbarer Menge. " +
		"Erstelle eine Packliste als JSON-Array. " +
		"Nutze ausschliesslich die Kandidatenliste. Erfinde niemals neue IDs. " +
		"Jeder Eintrag hat: equipment_type_id, quantity, reason. " +
		"Antworte ausschliesslich als JSON-Objekt mit dem Key \"suggestions\"."
)

// PacklistService handles AI-driven packlist suggestions via Ollama.
type PacklistService struct {
	ollamaBaseURL string
	modelName     string
	aiEnabled     bool
	client        *http.Client
	logger        zerolog.Logger
}

// PacklistConfig holds the configuration for the PacklistService.
type PacklistConfig struct {
	OllamaBaseURL string
	ModelName     string
	AIEnabled     bool
}

// NewPacklistService creates a new PacklistService.
func NewPacklistService(cfg PacklistConfig, logger zerolog.Logger) *PacklistService {
	baseURL := cfg.OllamaBaseURL
	if baseURL == "" {
		baseURL = defaultOllamaBaseURL
	}
	model := cfg.ModelName
	if model == "" {
		model = defaultAIModelName
	}

	return &PacklistService{
		ollamaBaseURL: strings.TrimRight(baseURL, "/"),
		modelName:     model,
		aiEnabled:     cfg.AIEnabled,
		client: &http.Client{
			Timeout: ollamaTimeout,
		},
		logger: logger.With().Str("service", "packlist").Logger(),
	}
}

// IsEnabled returns whether the AI feature flag is on.
func (s *PacklistService) IsEnabled() bool {
	return s.aiEnabled
}

// ModelName returns the configured Ollama model name.
func (s *PacklistService) ModelName() string {
	return s.modelName
}

// PromptVersion returns the current prompt version identifier.
func (s *PacklistService) PromptVersion() string {
	return promptVersion
}

// CheckOllamaHealth pings Ollama's /api/tags endpoint.
func (s *PacklistService) CheckOllamaHealth(ctx context.Context) error {
	url := s.ollamaBaseURL + "/api/tags"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("create health request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return domain.ErrOllamaUnavailable
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Ollama returned HTTP %d", resp.StatusCode)
	}

	return nil
}

// Suggest generates packlist suggestions using the Ollama API.
func (s *PacklistService) Suggest(ctx context.Context, req domain.PacklistSuggestRequest) (*domain.PacklistSuggestResponse, error) {
	if !s.aiEnabled {
		return nil, domain.ErrAIDisabled
	}

	// Validate input.
	if req.Project.ProjectID == "" {
		return nil, domain.ErrProjectIDRequired
	}
	if len(req.Candidates) == 0 {
		return nil, domain.ErrNoCandidates
	}

	maxItems := req.MaxItems
	if maxItems <= 0 {
		maxItems = defaultMaxItems
	}

	// Build the user prompt.
	userPrompt := s.buildUserPrompt(req, maxItems)

	// Build candidate lookup for validation.
	candidateMap := make(map[string]int, len(req.Candidates))
	for _, c := range req.Candidates {
		candidateMap[c.EquipmentTypeID] = c.AvailableQty
	}

	// Call Ollama.
	rawJSON, err := s.callOllama(ctx, userPrompt)
	if err != nil {
		s.logger.Error().Err(err).Msg("Ollama call failed, returning fallback")
		return s.buildFallback(req.Candidates, "Ollama nicht erreichbar — Fallback-Liste"), nil
	}

	// Parse the JSON response.
	suggestions, parseErr := s.parseResponse(rawJSON)
	if parseErr != nil {
		s.logger.Warn().
			Err(parseErr).
			Str("raw", truncateForLog(rawJSON, 500)).
			Msg("failed to parse Ollama response, returning fallback")
		return s.buildFallback(req.Candidates, "KI-Antwort nicht parsebar — Fallback-Liste"), nil
	}

	// Validate and clamp suggestions.
	validated, warnings := s.validateSuggestions(suggestions, candidateMap, maxItems)

	return &domain.PacklistSuggestResponse{
		Suggestions: validated,
		Warnings:    warnings,
		ModelUsed:   s.modelName,
		Cached:      false,
	}, nil
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

func (s *PacklistService) buildUserPrompt(req domain.PacklistSuggestRequest, maxItems int) string {
	var b strings.Builder

	b.WriteString("Projekt:\n")
	b.WriteString(fmt.Sprintf("  ID: %s\n", req.Project.ProjectID))
	b.WriteString(fmt.Sprintf("  Titel: %s\n", req.Project.Title))
	b.WriteString(fmt.Sprintf("  Von: %s\n", req.Project.DateFrom))
	b.WriteString(fmt.Sprintf("  Bis: %s\n", req.Project.DateTo))
	if req.Project.Notes != "" {
		b.WriteString(fmt.Sprintf("  Notizen: %s\n", req.Project.Notes))
	}
	b.WriteString(fmt.Sprintf("\nMaximale Anzahl Positionen: %d\n", maxItems))
	b.WriteString("\nKandidaten:\n")

	for _, c := range req.Candidates {
		b.WriteString(fmt.Sprintf("  - %s | %s | verfuegbar: %d\n", c.EquipmentTypeID, c.Name, c.AvailableQty))
	}

	return b.String()
}

// ollamaRequest is the JSON body sent to Ollama's /api/generate endpoint.
type ollamaRequest struct {
	Model  string `json:"model"`
	System string `json:"system"`
	Prompt string `json:"prompt"`
	Format string `json:"format"`
	Stream bool   `json:"stream"`
}

// ollamaResponse is the JSON body returned from Ollama's /api/generate.
type ollamaResponse struct {
	Response string `json:"response"`
}

func (s *PacklistService) callOllama(ctx context.Context, userPrompt string) (string, error) {
	body := ollamaRequest{
		Model:  s.modelName,
		System: packlistSystemPrompt,
		Prompt: userPrompt,
		Format: "json",
		Stream: false,
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("marshal ollama request: %w", err)
	}

	url := s.ollamaBaseURL + "/api/generate"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("create ollama request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	s.logger.Debug().
		Str("model", s.modelName).
		Str("url", url).
		Msg("calling Ollama API")

	resp, err := s.client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("ollama request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read ollama response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		s.logger.Error().
			Int("status", resp.StatusCode).
			Str("body", truncateForLog(string(respBody), 500)).
			Msg("Ollama API error")
		return "", fmt.Errorf("Ollama API Fehler (HTTP %d): %s", resp.StatusCode, truncateForLog(string(respBody), 200))
	}

	var ollamaResp ollamaResponse
	if err := json.Unmarshal(respBody, &ollamaResp); err != nil {
		return "", fmt.Errorf("parse ollama response: %w", err)
	}

	return ollamaResp.Response, nil
}

// ollamaJSONResponse is the expected structure inside Ollama's response field.
type ollamaJSONResponse struct {
	Suggestions []domain.PacklistSuggestion `json:"suggestions"`
}

func (s *PacklistService) parseResponse(raw string) ([]domain.PacklistSuggestion, error) {
	raw = stripJSONFences(raw)

	// Try parsing as object with "suggestions" key first.
	var wrapped ollamaJSONResponse
	if err := json.Unmarshal([]byte(raw), &wrapped); err == nil && len(wrapped.Suggestions) > 0 {
		return wrapped.Suggestions, nil
	}

	// Try parsing as a bare array.
	var bare []domain.PacklistSuggestion
	if err := json.Unmarshal([]byte(raw), &bare); err == nil && len(bare) > 0 {
		return bare, nil
	}

	return nil, fmt.Errorf("could not parse response as suggestions JSON")
}

func (s *PacklistService) validateSuggestions(
	suggestions []domain.PacklistSuggestion,
	candidateMap map[string]int,
	maxItems int,
) ([]domain.PacklistSuggestion, []string) {
	var validated []domain.PacklistSuggestion
	var warnings []string

	for _, sug := range suggestions {
		avail, exists := candidateMap[sug.EquipmentTypeID]
		if !exists {
			warnings = append(warnings, fmt.Sprintf("equipment_type_id %q nicht in Kandidatenliste — uebersprungen", sug.EquipmentTypeID))
			continue
		}

		if sug.Quantity < 0 {
			warnings = append(warnings, fmt.Sprintf("equipment_type_id %q: negative Menge (%d) auf 0 korrigiert", sug.EquipmentTypeID, sug.Quantity))
			sug.Quantity = 0
		}

		if sug.Quantity > avail {
			warnings = append(warnings, fmt.Sprintf("equipment_type_id %q: Menge %d > verfuegbar %d — auf %d begrenzt", sug.EquipmentTypeID, sug.Quantity, avail, avail))
			sug.Quantity = avail
		}

		validated = append(validated, sug)

		if len(validated) >= maxItems {
			warnings = append(warnings, fmt.Sprintf("max_items (%d) erreicht — restliche Vorschlaege abgeschnitten", maxItems))
			break
		}
	}

	return validated, warnings
}

func (s *PacklistService) buildFallback(candidates []domain.EquipmentCandidate, reason string) *domain.PacklistSuggestResponse {
	suggestions := make([]domain.PacklistSuggestion, 0, len(candidates))
	for _, c := range candidates {
		qty := 1
		if c.AvailableQty <= 0 {
			continue
		}
		suggestions = append(suggestions, domain.PacklistSuggestion{
			EquipmentTypeID: c.EquipmentTypeID,
			Quantity:        qty,
			Reason:          "Fallback: je 1 Stueck vorgeschlagen",
		})
	}

	return &domain.PacklistSuggestResponse{
		Suggestions: suggestions,
		Warnings:    []string{reason},
		ModelUsed:   s.modelName,
		Cached:      false,
	}
}

// truncateForLog shortens a string for safe logging.
func truncateForLog(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

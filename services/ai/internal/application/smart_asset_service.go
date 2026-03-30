package application

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/rs/zerolog"
)

const smartAssetSystemPrompt = "Du bist ein Experte fuer Veranstaltungstechnik-Equipment. " +
	"Extrahiere aus der Beschreibung: name, category (Mikrofone/Lautsprecher/Mischpulte/Licht/Kabel/Stative/Video/Strom), " +
	"manufacturer, model, description. Antworte als JSON."

// SmartAssetService uses the ClaudeProvider to auto-fill equipment data from
// a free-text description.
type SmartAssetService struct {
	provider *ClaudeProvider
	logger   zerolog.Logger
}

// SmartAssetRequest is the input for the smart-asset endpoint.
type SmartAssetRequest struct {
	Description string `json:"description"`
}

// SmartAssetResponse is the AI-generated equipment metadata.
type SmartAssetResponse struct {
	Name         string `json:"name"`
	Category     string `json:"category"`
	Manufacturer string `json:"manufacturer"`
	Model        string `json:"model"`
	Description  string `json:"description"`
}

// NewSmartAssetService creates a new SmartAssetService backed by the given
// ClaudeProvider.
func NewSmartAssetService(provider *ClaudeProvider, logger zerolog.Logger) *SmartAssetService {
	return &SmartAssetService{
		provider: provider,
		logger:   logger.With().Str("service", "smart_asset").Logger(),
	}
}

// Generate takes a free-text equipment description and returns structured
// metadata extracted by the AI model.
func (s *SmartAssetService) Generate(ctx context.Context, req SmartAssetRequest) (*SmartAssetResponse, error) {
	if strings.TrimSpace(req.Description) == "" {
		return nil, fmt.Errorf("description is required")
	}

	aiReq := AIRequest{
		SystemPrompt: smartAssetSystemPrompt,
		UserMessage:  req.Description,
		MaxTokens:    512,
	}

	aiResp, err := s.provider.Complete(ctx, aiReq)
	if err != nil {
		return nil, fmt.Errorf("KI-Anfrage fehlgeschlagen: %w", err)
	}

	// The model should return a JSON object. Strip any markdown fences that
	// some models add around JSON output.
	raw := stripJSONFences(aiResp.Content)

	var result SmartAssetResponse
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		s.logger.Warn().
			Str("raw_content", aiResp.Content).
			Err(err).
			Msg("failed to parse AI response as JSON")
		return nil, fmt.Errorf("KI-Antwort konnte nicht als JSON verarbeitet werden: %w", err)
	}

	s.logger.Info().
		Str("name", result.Name).
		Str("category", result.Category).
		Str("model", aiResp.Model).
		Msg("smart asset generated")

	return &result, nil
}

// stripJSONFences removes optional ```json ... ``` markdown fencing from the
// model output so that json.Unmarshal works reliably.
func stripJSONFences(s string) string {
	s = strings.TrimSpace(s)

	// Remove leading ```json or ```
	if strings.HasPrefix(s, "```json") {
		s = strings.TrimPrefix(s, "```json")
	} else if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```")
	}

	// Remove trailing ```
	if strings.HasSuffix(s, "```") {
		s = strings.TrimSuffix(s, "```")
	}

	return strings.TrimSpace(s)
}

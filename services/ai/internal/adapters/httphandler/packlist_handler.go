package httphandler

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/pkg/common/response"
	"github.com/jeckersberger/EquipFlow/services/ai/internal/application"
	"github.com/jeckersberger/EquipFlow/services/ai/internal/domain"
)

// PacklistHandler exposes HTTP endpoints for AI-driven packlist suggestions.
type PacklistHandler struct {
	service *application.PacklistService
	logger  zerolog.Logger
}

// NewPacklistHandler creates a new PacklistHandler.
func NewPacklistHandler(service *application.PacklistService, logger zerolog.Logger) *PacklistHandler {
	return &PacklistHandler{
		service: service,
		logger:  logger.With().Str("handler", "packlist").Logger(),
	}
}

// ---------------------------------------------------------------------------
// POST /api/v1/ai/packlist/suggest
// ---------------------------------------------------------------------------

// Suggest generates AI-based packlist suggestions for a project.
func (h *PacklistHandler) Suggest(w http.ResponseWriter, r *http.Request) {
	if !h.service.IsEnabled() {
		response.Error(w, http.StatusServiceUnavailable, "AI_DISABLED", domain.ErrAIDisabled.Error())
		return
	}

	var req domain.PacklistSuggestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "Ungueltiger Request-Body")
		return
	}

	result, err := h.service.Suggest(r.Context(), req)
	if err != nil {
		h.logger.Error().Err(err).Msg("packlist suggestion failed")

		switch err {
		case domain.ErrProjectIDRequired, domain.ErrNoCandidates:
			response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		case domain.ErrAIDisabled:
			response.Error(w, http.StatusServiceUnavailable, "AI_DISABLED", err.Error())
		default:
			response.Error(w, http.StatusInternalServerError, "AI_ERROR", err.Error())
		}
		return
	}

	response.Success(w, result)
}

// ---------------------------------------------------------------------------
// GET /api/v1/ai/readyz
// ---------------------------------------------------------------------------

// Readyz checks whether the Ollama backend is reachable.
func (h *PacklistHandler) Readyz(w http.ResponseWriter, r *http.Request) {
	if !h.service.IsEnabled() {
		response.Error(w, http.StatusServiceUnavailable, "AI_DISABLED", domain.ErrAIDisabled.Error())
		return
	}

	if err := h.service.CheckOllamaHealth(r.Context()); err != nil {
		h.logger.Warn().Err(err).Msg("Ollama health check failed")
		response.Error(w, http.StatusServiceUnavailable, "OLLAMA_UNAVAILABLE", err.Error())
		return
	}

	response.Success(w, map[string]string{"status": "ok"})
}

// ---------------------------------------------------------------------------
// GET /api/v1/ai/meta
// ---------------------------------------------------------------------------

type metaResponse struct {
	ModelName     string `json:"model_name"`
	PromptVersion string `json:"prompt_version"`
	AIEnabled     bool   `json:"ai_enabled"`
}

// Meta returns the current AI configuration metadata.
func (h *PacklistHandler) Meta(w http.ResponseWriter, r *http.Request) {
	resp := metaResponse{
		ModelName:     h.service.ModelName(),
		PromptVersion: h.service.PromptVersion(),
		AIEnabled:     h.service.IsEnabled(),
	}
	response.Success(w, resp)
}

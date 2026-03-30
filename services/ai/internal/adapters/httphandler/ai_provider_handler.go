package httphandler

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/pkg/common/errors"
	"github.com/jeckersberger/EquipFlow/pkg/common/response"
	"github.com/jeckersberger/EquipFlow/services/ai/internal/application"
)

// AIProviderHandler exposes HTTP endpoints for direct AI completion and the
// smart-asset auto-fill feature.
type AIProviderHandler struct {
	provider     *application.ClaudeProvider
	smartAssetSvc *application.SmartAssetService
	logger       zerolog.Logger
}

// NewAIProviderHandler creates a new handler wired to the Claude provider and
// the smart-asset service.
func NewAIProviderHandler(
	provider *application.ClaudeProvider,
	smartAssetSvc *application.SmartAssetService,
	logger zerolog.Logger,
) *AIProviderHandler {
	return &AIProviderHandler{
		provider:     provider,
		smartAssetSvc: smartAssetSvc,
		logger:       logger.With().Str("handler", "ai_provider").Logger(),
	}
}

// ---------------------------------------------------------------------------
// POST /api/v1/ai/complete
// ---------------------------------------------------------------------------

type completeRequest struct {
	SystemPrompt string `json:"system_prompt"`
	UserMessage  string `json:"user_message"`
	MaxTokens    int    `json:"max_tokens"`
}

// Complete handles general AI completion requests.
func (h *AIProviderHandler) Complete(w http.ResponseWriter, r *http.Request) {
	var req completeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiger Request-Body"))
		return
	}

	if req.UserMessage == "" {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "user_message ist erforderlich"))
		return
	}

	aiReq := application.AIRequest{
		SystemPrompt: req.SystemPrompt,
		UserMessage:  req.UserMessage,
		MaxTokens:    req.MaxTokens,
	}

	aiResp, err := h.provider.Complete(r.Context(), aiReq)
	if err != nil {
		h.logger.Error().Err(err).Msg("AI completion failed")
		response.Error(w, http.StatusServiceUnavailable, "AI_ERROR", err.Error())
		return
	}

	response.Success(w, aiResp)
}

// ---------------------------------------------------------------------------
// POST /api/v1/ai/smart-asset
// ---------------------------------------------------------------------------

// SmartAsset handles equipment auto-fill requests.
func (h *AIProviderHandler) SmartAsset(w http.ResponseWriter, r *http.Request) {
	var req application.SmartAssetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiger Request-Body"))
		return
	}

	result, err := h.smartAssetSvc.Generate(r.Context(), req)
	if err != nil {
		h.logger.Error().Err(err).Msg("smart asset generation failed")
		response.Error(w, http.StatusServiceUnavailable, "AI_ERROR", err.Error())
		return
	}

	response.Success(w, result)
}

// ---------------------------------------------------------------------------
// GET /api/v1/ai/status
// ---------------------------------------------------------------------------

type aiStatusResponse struct {
	Configured bool   `json:"configured"`
	Model      string `json:"model,omitempty"`
}

// Status reports whether the AI provider is configured (API key present).
func (h *AIProviderHandler) Status(w http.ResponseWriter, r *http.Request) {
	configured := h.provider.IsConfigured()
	resp := aiStatusResponse{
		Configured: configured,
	}
	if configured {
		resp.Model = h.provider.ModelName()
	}
	response.Success(w, resp)
}

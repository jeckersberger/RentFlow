package httphandler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/pkg/common/errors"
	"github.com/jeckersberger/EquipFlow/pkg/common/response"
	"github.com/jeckersberger/EquipFlow/services/invoice/internal/application"
)

// PortalHandler handles public (unauthenticated) customer portal endpoints.
type PortalHandler struct {
	quoteService *application.QuoteService
	logger       zerolog.Logger
}

// NewPortalHandler creates a new PortalHandler.
func NewPortalHandler(quoteService *application.QuoteService, logger zerolog.Logger) *PortalHandler {
	return &PortalHandler{
		quoteService: quoteService,
		logger:       logger.With().Str("handler", "portal").Logger(),
	}
}

// GetPublicQuote retrieves a quote by its public token (no authentication required).
// GET /api/v1/public/quotes/{token}
func (h *PortalHandler) GetPublicQuote(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Token erforderlich"))
		return
	}

	quote, items, err := h.quoteService.GetByPublicToken(r.Context(), token)
	if err != nil {
		h.logger.Warn().Str("token", token[:min(8, len(token))]).Msg("public quote lookup failed")
		errors.HandleError(w, errors.ErrNotFound)
		return
	}

	response.Success(w, map[string]interface{}{
		"quote": quote,
		"items": items,
	})
}

// RespondToQuote allows a customer to accept or decline a quote via public token.
// POST /api/v1/public/quotes/{token}/respond
func (h *PortalHandler) RespondToQuote(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Token erforderlich"))
		return
	}

	var req struct {
		Response string `json:"response"` // "accepted" or "declined"
		Message  string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiger Request-Body"))
		return
	}

	if req.Response != "accepted" && req.Response != "declined" {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Response muss 'accepted' oder 'declined' sein"))
		return
	}

	err := h.quoteService.RespondViaPortal(r.Context(), token, req.Response, req.Message)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, map[string]string{
		"status":  "ok",
		"message": "Antwort erfolgreich gespeichert",
	})
}

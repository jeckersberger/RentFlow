package httphandler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/pkg/common/errors"
	"github.com/jeckersberger/EquipFlow/pkg/common/middleware"
	"github.com/jeckersberger/EquipFlow/pkg/common/response"
	"github.com/jeckersberger/EquipFlow/services/document/internal/application"
)

// RenderHandler handles template rendering HTTP endpoints.
type RenderHandler struct {
	renderer *application.TemplateRenderer
	logger   zerolog.Logger
}

// NewRenderHandler creates a new RenderHandler.
func NewRenderHandler(
	renderer *application.TemplateRenderer,
	logger zerolog.Logger,
) *RenderHandler {
	return &RenderHandler{
		renderer: renderer,
		logger:   logger.With().Str("handler", "render").Logger(),
	}
}

// ---------------------------------------------------------------------------
// Endpoints
// ---------------------------------------------------------------------------

// Render handles POST /api/v1/documents/render -- render a template to HTML.
func (h *RenderHandler) Render(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	var req application.RenderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiger Request-Body"))
		return
	}

	result, err := h.renderer.Render(r.Context(), claims.TenantID, req)
	if err != nil {
		h.logger.Error().Err(err).
			Str("template_type", req.TemplateType).
			Str("tenant_id", claims.TenantID.String()).
			Msg("template render failed")
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, err.Error()))
		return
	}

	response.Success(w, result)
}

// ListDefaults handles GET /api/v1/documents/templates/defaults -- list default template types.
func (h *RenderHandler) ListDefaults(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	types := application.ListTemplateTypes()
	response.Success(w, types)
}

// PreviewDefault handles GET /api/v1/documents/templates/defaults/{type}/preview.
// It renders the default template with sample data and returns the HTML.
func (h *RenderHandler) PreviewDefault(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	templateType := chi.URLParam(r, "type")
	sampleData := application.GetSampleData(templateType)
	if len(sampleData) == 0 {
		errors.HandleError(w, errors.Wrap(errors.ErrNotFound, "Unbekannter Template-Typ: "+templateType))
		return
	}

	result, err := h.renderer.Render(r.Context(), claims.TenantID, application.RenderRequest{
		TemplateType: templateType,
		Data:         sampleData,
	})
	if err != nil {
		h.logger.Error().Err(err).
			Str("template_type", templateType).
			Msg("preview render failed")
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, err.Error()))
		return
	}

	// Return the rendered HTML directly for preview in a browser.
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(result.HTML))
}

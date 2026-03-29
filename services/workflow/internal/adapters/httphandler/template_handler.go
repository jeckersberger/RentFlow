package httphandler

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/pkg/common/errors"
	"github.com/jeckersberger/EquipFlow/pkg/common/middleware"
	"github.com/jeckersberger/EquipFlow/pkg/common/response"
	"github.com/jeckersberger/EquipFlow/services/workflow/internal/application"
)

// TemplateHandler handles workflow template HTTP endpoints.
type TemplateHandler struct {
	defService *application.DefinitionService
	logger     zerolog.Logger
}

// NewTemplateHandler creates a new TemplateHandler.
func NewTemplateHandler(
	defService *application.DefinitionService,
	logger zerolog.Logger,
) *TemplateHandler {
	return &TemplateHandler{
		defService: defService,
		logger:     logger.With().Str("handler", "template").Logger(),
	}
}

// ListTemplates handles GET /api/v1/workflow-templates — list available templates.
func (h *TemplateHandler) ListTemplates(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	templates := h.defService.ListTemplates()
	response.Success(w, templates)
}

// createFromTemplateRequest is the request body for creating a definition from a template.
type createFromTemplateRequest struct {
	TemplateName string `json:"template_name"`
}

// CreateFromTemplate handles POST /api/v1/workflow-definitions/from-template — create from template.
func (h *TemplateHandler) CreateFromTemplate(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	var req createFromTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiger Request-Body"))
		return
	}

	if req.TemplateName == "" {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "template_name ist erforderlich"))
		return
	}

	def, err := h.defService.CreateFromTemplate(r.Context(), claims.TenantID, claims.UserID, req.TemplateName)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Created(w, def)
}

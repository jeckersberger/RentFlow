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

// DocumentHandler handles document HTTP endpoints.
type DocumentHandler struct {
	documentService *application.DocumentService
	logger          zerolog.Logger
}

// NewDocumentHandler creates a new DocumentHandler.
func NewDocumentHandler(
	documentService *application.DocumentService,
	logger zerolog.Logger,
) *DocumentHandler {
	return &DocumentHandler{
		documentService: documentService,
		logger:          logger.With().Str("handler", "document").Logger(),
	}
}

// ---------------------------------------------------------------------------
// Endpoints
// ---------------------------------------------------------------------------

// Create handles POST /api/v1/documents — Dokument erstellen.
func (h *DocumentHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	var req application.CreateDocumentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiger Request-Body"))
		return
	}

	doc, err := h.documentService.Create(r.Context(), claims.TenantID, &claims.UserID, req)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Created(w, doc)
}

// List handles GET /api/v1/documents — Alle Dokumente.
func (h *DocumentHandler) List(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	docs, err := h.documentService.List(r.Context(), claims.TenantID)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, docs)
}

// GetByID handles GET /api/v1/documents/{id} — Dokument abrufen.
func (h *DocumentHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	doc, err := h.documentService.GetByID(r.Context(), id, claims.TenantID)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, doc)
}

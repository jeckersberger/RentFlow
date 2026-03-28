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

// AttachmentHandler handles attachment HTTP endpoints.
type AttachmentHandler struct {
	attachmentService *application.AttachmentService
	logger            zerolog.Logger
}

// NewAttachmentHandler creates a new AttachmentHandler.
func NewAttachmentHandler(
	attachmentService *application.AttachmentService,
	logger zerolog.Logger,
) *AttachmentHandler {
	return &AttachmentHandler{
		attachmentService: attachmentService,
		logger:            logger.With().Str("handler", "attachment").Logger(),
	}
}

// ---------------------------------------------------------------------------
// Endpoints
// ---------------------------------------------------------------------------

// Create handles POST /api/v1/attachments — Anhang erstellen.
func (h *AttachmentHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	var req application.CreateAttachmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiger Request-Body"))
		return
	}

	att, err := h.attachmentService.Create(r.Context(), claims.TenantID, &claims.UserID, req)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Created(w, att)
}

// List handles GET /api/v1/attachments — Anhaenge auflisten (query: reference_id, reference_type).
func (h *AttachmentHandler) List(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	referenceIDStr := r.URL.Query().Get("reference_id")
	referenceType := r.URL.Query().Get("reference_type")

	if referenceIDStr == "" || referenceType == "" {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "reference_id und reference_type sind erforderlich"))
		return
	}

	referenceID, err := parseUUID(referenceIDStr)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	attachments, err := h.attachmentService.ListByReference(r.Context(), claims.TenantID, referenceID, referenceType)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, attachments)
}

// GetByID handles GET /api/v1/attachments/{id} — Anhang abrufen.
func (h *AttachmentHandler) GetByID(w http.ResponseWriter, r *http.Request) {
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

	att, err := h.attachmentService.GetByID(r.Context(), id, claims.TenantID)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, att)
}

// Delete handles DELETE /api/v1/attachments/{id} — Anhang loeschen.
func (h *AttachmentHandler) Delete(w http.ResponseWriter, r *http.Request) {
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

	if err := h.attachmentService.Delete(r.Context(), id, claims.TenantID); err != nil {
		errors.HandleError(w, err)
		return
	}

	response.NoContent(w)
}

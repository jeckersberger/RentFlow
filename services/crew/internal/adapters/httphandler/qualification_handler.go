package httphandler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/pkg/common/errors"
	"github.com/jeckersberger/EquipFlow/pkg/common/middleware"
	"github.com/jeckersberger/EquipFlow/pkg/common/response"
	"github.com/jeckersberger/EquipFlow/services/crew/internal/application"
)

// QualificationHandler handles crew qualification HTTP endpoints.
type QualificationHandler struct {
	qualificationService *application.QualificationService
	logger               zerolog.Logger
}

// NewQualificationHandler creates a new QualificationHandler.
func NewQualificationHandler(
	qualificationService *application.QualificationService,
	logger zerolog.Logger,
) *QualificationHandler {
	return &QualificationHandler{
		qualificationService: qualificationService,
		logger:               logger.With().Str("handler", "qualification").Logger(),
	}
}

// ---------------------------------------------------------------------------
// Endpoints
// ---------------------------------------------------------------------------

// CreateForMember handles POST /api/v1/crew/{id}/qualifications — Qualifikation erstellen.
func (h *QualificationHandler) CreateForMember(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	crewMemberID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	var req application.CreateQualificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiger Request-Body"))
		return
	}

	qualification, err := h.qualificationService.Create(r.Context(), crewMemberID, claims.TenantID, req)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Created(w, qualification)
}

// ListForMember handles GET /api/v1/crew/{id}/qualifications — Qualifikationen eines Mitglieds.
func (h *QualificationHandler) ListForMember(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	crewMemberID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	qualifications, err := h.qualificationService.ListByMember(r.Context(), crewMemberID, claims.TenantID)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, qualifications)
}

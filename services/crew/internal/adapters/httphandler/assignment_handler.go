package httphandler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/pkg/common/errors"
	"github.com/jeckersberger/EquipFlow/pkg/common/middleware"
	"github.com/jeckersberger/EquipFlow/pkg/common/response"
	"github.com/jeckersberger/EquipFlow/services/crew/internal/application"
)

// AssignmentHandler handles crew assignment HTTP endpoints.
type AssignmentHandler struct {
	assignmentService *application.AssignmentService
	logger            zerolog.Logger
}

// NewAssignmentHandler creates a new AssignmentHandler.
func NewAssignmentHandler(
	assignmentService *application.AssignmentService,
	logger zerolog.Logger,
) *AssignmentHandler {
	return &AssignmentHandler{
		assignmentService: assignmentService,
		logger:            logger.With().Str("handler", "assignment").Logger(),
	}
}

// ---------------------------------------------------------------------------
// Endpoints
// ---------------------------------------------------------------------------

// CreateForMember handles POST /api/v1/crew/{id}/assignments — Assignment erstellen.
func (h *AssignmentHandler) CreateForMember(w http.ResponseWriter, r *http.Request) {
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

	var req application.CreateAssignmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiger Request-Body"))
		return
	}

	assignment, err := h.assignmentService.Create(r.Context(), crewMemberID, claims.TenantID, req)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Created(w, assignment)
}

// ListForMember handles GET /api/v1/crew/{id}/assignments — Assignments eines Mitglieds.
func (h *AssignmentHandler) ListForMember(w http.ResponseWriter, r *http.Request) {
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

	assignments, err := h.assignmentService.ListByMember(r.Context(), crewMemberID, claims.TenantID)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, assignments)
}

// ListAll handles GET /api/v1/crew-assignments — Alle Assignments (mit project_id Filter).
func (h *AssignmentHandler) ListAll(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	var projectID *uuid.UUID
	if v := r.URL.Query().Get("project_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltige project_id"))
			return
		}
		projectID = &id
	}

	assignments, err := h.assignmentService.ListAll(r.Context(), claims.TenantID, projectID)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, assignments)
}

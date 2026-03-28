package httphandler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/pkg/common/errors"
	"github.com/jeckersberger/EquipFlow/pkg/common/middleware"
	"github.com/jeckersberger/EquipFlow/pkg/common/pagination"
	"github.com/jeckersberger/EquipFlow/pkg/common/response"
	"github.com/jeckersberger/EquipFlow/services/project/internal/application"
	"github.com/jeckersberger/EquipFlow/services/project/internal/domain"
)

type ProjectHandler struct {
	projectService *application.ProjectService
	logger         zerolog.Logger
}

func NewProjectHandler(projectService *application.ProjectService, logger zerolog.Logger) *ProjectHandler {
	return &ProjectHandler{
		projectService: projectService,
		logger:         logger.With().Str("handler", "project").Logger(),
	}
}

func (h *ProjectHandler) List(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	p := pagination.Parse(r)
	q := r.URL.Query()

	filter := domain.ProjectFilter{
		Page:    p.Page,
		PerPage: p.PerPage,
	}
	if v := q.Get("status"); v != "" {
		filter.Status = &v
	}
	if v := q.Get("search"); v != "" {
		filter.Search = &v
	}

	items, total, err := h.projectService.List(r.Context(), claims.TenantID, filter)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.SuccessWithMeta(w, items, response.Meta{
		Page:    p.Page,
		PerPage: p.PerPage,
		Total:   total,
	})
}

func (h *ProjectHandler) Get(w http.ResponseWriter, r *http.Request) {
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

	project, err := h.projectService.GetByID(r.Context(), id, claims.TenantID)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, project)
}

func (h *ProjectHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	var req application.CreateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiger Request-Body"))
		return
	}

	created, err := h.projectService.Create(r.Context(), claims.TenantID, req)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Created(w, created)
}

func (h *ProjectHandler) Update(w http.ResponseWriter, r *http.Request) {
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

	var req application.UpdateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiger Request-Body"))
		return
	}

	updated, err := h.projectService.Update(r.Context(), id, claims.TenantID, req)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, updated)
}

func (h *ProjectHandler) Delete(w http.ResponseWriter, r *http.Request) {
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

	if err := h.projectService.Delete(r.Context(), id, claims.TenantID); err != nil {
		errors.HandleError(w, err)
		return
	}

	response.NoContent(w)
}

func (h *ProjectHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
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

	var req application.UpdateProjectStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiger Request-Body"))
		return
	}

	if err := h.projectService.UpdateStatus(r.Context(), id, claims.TenantID, req); err != nil {
		errors.HandleError(w, err)
		return
	}

	response.NoContent(w)
}

func (h *ProjectHandler) Search(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	query := r.URL.Query().Get("q")
	if query == "" {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Suchbegriff (q) ist erforderlich"))
		return
	}

	p := pagination.Parse(r)

	items, total, err := h.projectService.Search(r.Context(), claims.TenantID, query, p.Page, p.PerPage)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.SuccessWithMeta(w, items, response.Meta{
		Page:    p.Page,
		PerPage: p.PerPage,
		Total:   total,
	})
}

// --- Project Equipment endpoints ---

func (h *ProjectHandler) AddEquipment(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	projectID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	var req application.AddProjectEquipmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiger Request-Body"))
		return
	}

	created, err := h.projectService.AddEquipment(r.Context(), claims.TenantID, projectID, req)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Created(w, created)
}

func (h *ProjectHandler) RemoveEquipment(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	eqID, err := parseUUID(chi.URLParam(r, "equipmentId"))
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	if err := h.projectService.RemoveEquipment(r.Context(), claims.TenantID, eqID); err != nil {
		errors.HandleError(w, err)
		return
	}

	response.NoContent(w)
}

func (h *ProjectHandler) ListEquipment(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	projectID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	items, err := h.projectService.ListEquipment(r.Context(), claims.TenantID, projectID)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, items)
}

package http

import (
	"encoding/json"
	"net/http"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/workflow-service/internal/application"
	"github.com/jeckersberger/rentflow/services/workflow-service/internal/domain"
)

type Handler struct {
	workflowSvc *application.WorkflowService
	logger      *logger.Logger
}

func NewHandler(workflowSvc *application.WorkflowService, log *logger.Logger) *Handler {
	return &Handler{
		workflowSvc: workflowSvc,
		logger:      log,
	}
}

func (h *Handler) CreateWorkflow(w http.ResponseWriter, r *http.Request) {
	var req application.CreateWorkflowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	resp, err := h.workflowSvc.CreateWorkflow(r.Context(), tenantID, req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusCreated, resp)
}

func (h *Handler) GetWorkflow(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	resp, err := h.workflowSvc.GetWorkflow(r.Context(), tenantID, id)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, resp)
}

func (h *Handler) ListWorkflows(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	resp, err := h.workflowSvc.ListWorkflows(r.Context(), tenantID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{"data": resp})
}

func (h *Handler) UpdateWorkflow(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req application.CreateWorkflowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	resp, err := h.workflowSvc.UpdateWorkflow(r.Context(), tenantID, id, req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, resp)
}

func (h *Handler) DeleteWorkflow(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	err := h.workflowSvc.DeleteWorkflow(r.Context(), tenantID, id)
	if err != nil {
		h.handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ActivateWorkflow(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req application.ActivateWorkflowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req.Activate = true
	}

	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	resp, err := h.workflowSvc.ActivateWorkflow(r.Context(), tenantID, id, req.Activate)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, resp)
}

func (h *Handler) DeactivateWorkflow(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	resp, err := h.workflowSvc.ActivateWorkflow(r.Context(), tenantID, id, false)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, resp)
}

func (h *Handler) TriggerWorkflow(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req application.TriggerWorkflowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	resp, err := h.workflowSvc.TriggerWorkflow(r.Context(), tenantID, id, req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, resp)
}

func (h *Handler) GetWorkflowRun(w http.ResponseWriter, r *http.Request) {
	runID := r.PathValue("id")
	resp, err := h.workflowSvc.GetWorkflowRun(r.Context(), runID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, resp)
}

func (h *Handler) ListWorkflowRuns(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	resp, err := h.workflowSvc.ListWorkflowRuns(r.Context(), tenantID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{"data": resp})
}

func (h *Handler) CancelWorkflowRun(w http.ResponseWriter, r *http.Request) {
	runID := r.PathValue("id")
	var req application.CancelRunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req.Reason = "cancelled by user"
	}

	resp, err := h.workflowSvc.CancelWorkflowRun(r.Context(), runID, req.Reason)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, resp)
}

func (h *Handler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) respondError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func (h *Handler) handleError(w http.ResponseWriter, err error) {
	if err == domain.ErrWorkflowNotFound || err == domain.ErrWorkflowRunNotFound {
		h.respondError(w, http.StatusNotFound, err.Error())
		return
	}
	if err == domain.ErrInvalidInput || err == domain.ErrTenantIDRequired {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.logger.Error("Unhandled error", err)
	h.respondError(w, http.StatusInternalServerError, "internal server error")
}

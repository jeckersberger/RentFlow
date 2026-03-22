package http

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/workflow-service/internal/application"
	"github.com/jeckersberger/rentflow/services/workflow-service/internal/domain"
)

type Handlers struct {
	workflowService *application.WorkflowService
	triggerService  *application.TriggerService
	actionExecutor  *application.ActionExecutor
	log             logger.Logger
}

func NewHandlers(ws *application.WorkflowService, ts *application.TriggerService, ae *application.ActionExecutor, log logger.Logger) *Handlers {
	return &Handlers{
		workflowService: ws,
		triggerService:  ts,
		actionExecutor:  ae,
		log:             log,
	}
}

// CreateDefinition handles POST /api/v1/workflows/definitions
func (h *Handlers) CreateDefinition(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		http.Error(w, "Missing X-Tenant-ID header", http.StatusBadRequest)
		return
	}

	var req application.CreateWorkflowDefinitionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	tid, _ := uuid.Parse(tenantID)
	resp, err := h.workflowService.CreateDefinition(r.Context(), tid, &req)
	if err != nil {
		h.log.Error("Failed to create definition", err)
		http.Error(w, "Failed to create definition", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// GetDefinition handles GET /api/v1/workflows/definitions/{id}
func (h *Handlers) GetDefinition(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	defID, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "Invalid definition ID", http.StatusBadRequest)
		return
	}

	resp, err := h.workflowService.GetDefinition(r.Context(), defID)
	if err == domain.ErrWorkflowNotFound {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if err != nil {
		h.log.Error("Failed to get definition", err)
		http.Error(w, "Failed to get definition", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// ListDefinitions handles GET /api/v1/workflows/definitions
func (h *Handlers) ListDefinitions(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		http.Error(w, "Missing X-Tenant-ID header", http.StatusBadRequest)
		return
	}

	tid, _ := uuid.Parse(tenantID)
	resps, err := h.workflowService.ListDefinitions(r.Context(), tid)
	if err != nil {
		h.log.Error("Failed to list definitions", err)
		http.Error(w, "Failed to list definitions", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resps)
}

// ListTemplates handles GET /api/v1/workflows/templates
func (h *Handlers) ListTemplates(w http.ResponseWriter, r *http.Request) {
	resps, err := h.workflowService.ListTemplates(r.Context())
	if err != nil {
		h.log.Error("Failed to list templates", err)
		http.Error(w, "Failed to list templates", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resps)
}

// UpdateDefinition handles PUT /api/v1/workflows/definitions/{id}
func (h *Handlers) UpdateDefinition(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	defID, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "Invalid definition ID", http.StatusBadRequest)
		return
	}

	var req application.CreateWorkflowDefinitionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := h.workflowService.UpdateDefinition(r.Context(), defID, &req)
	if err == domain.ErrWorkflowNotFound {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if err != nil {
		h.log.Error("Failed to update definition", err)
		http.Error(w, "Failed to update definition", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// DeleteDefinition handles DELETE /api/v1/workflows/definitions/{id}
func (h *Handlers) DeleteDefinition(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	defID, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "Invalid definition ID", http.StatusBadRequest)
		return
	}

	if err := h.workflowService.DeleteDefinition(r.Context(), defID); err != nil {
		h.log.Error("Failed to delete definition", err)
		http.Error(w, "Failed to delete definition", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// InstantiateWorkflow handles POST /api/v1/workflows/definitions/{id}/instantiate
func (h *Handlers) InstantiateWorkflow(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		http.Error(w, "Missing X-Tenant-ID header", http.StatusBadRequest)
		return
	}

	defID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid definition ID", http.StatusBadRequest)
		return
	}

	var req application.InstantiateWorkflowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	tid, _ := uuid.Parse(tenantID)
	resp, err := h.workflowService.InstantiateWorkflow(r.Context(), tid, defID, &req)
	if err != nil {
		h.log.Error("Failed to instantiate workflow", err)
		http.Error(w, "Failed to instantiate workflow", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// GetInstance handles GET /api/v1/workflows/instances/{id}
func (h *Handlers) GetInstance(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	instID, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "Invalid instance ID", http.StatusBadRequest)
		return
	}

	resp, err := h.workflowService.GetInstance(r.Context(), instID)
	if err == domain.ErrWorkflowInstanceNotFound {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if err != nil {
		h.log.Error("Failed to get instance", err)
		http.Error(w, "Failed to get instance", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// ListInstances handles GET /api/v1/workflows/instances
func (h *Handlers) ListInstances(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		http.Error(w, "Missing X-Tenant-ID header", http.StatusBadRequest)
		return
	}

	tid, _ := uuid.Parse(tenantID)
	resps, err := h.workflowService.ListInstances(r.Context(), tid)
	if err != nil {
		h.log.Error("Failed to list instances", err)
		http.Error(w, "Failed to list instances", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resps)
}

// CancelInstance handles POST /api/v1/workflows/instances/{id}/cancel
func (h *Handlers) CancelInstance(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	instID, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "Invalid instance ID", http.StatusBadRequest)
		return
	}

	if err := h.workflowService.CancelInstance(r.Context(), instID); err != nil {
		h.log.Error("Failed to cancel instance", err)
		http.Error(w, "Failed to cancel instance", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// TriggerWorkflow handles POST /api/v1/workflows/trigger
func (h *Handlers) TriggerWorkflow(w http.ResponseWriter, r *http.Request) {
	var req application.TriggerWorkflowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.triggerService.ProcessEventTrigger(r.Context(), req.EventName, req.Data); err != nil {
		h.log.Error("Failed to trigger workflow", err)
		http.Error(w, "Failed to trigger workflow", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"status": "triggered"})
}

// GetDashboard handles GET /api/v1/workflows/dashboard
func (h *Handlers) GetDashboard(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		http.Error(w, "Missing X-Tenant-ID header", http.StatusBadRequest)
		return
	}

	tid, _ := uuid.Parse(tenantID)
	resp, err := h.workflowService.GetDashboard(r.Context(), tid)
	if err != nil {
		h.log.Error("Failed to get dashboard", err)
		http.Error(w, "Failed to get dashboard", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

package http

import (
	"net/http"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/workflow-service/internal/application"
)

func NewRouter(workflowSvc *application.WorkflowService, log *logger.Logger) *http.ServeMux {
	router := http.NewServeMux()
	handler := NewHandler(workflowSvc, log)

	router.HandleFunc("GET /health", healthHandler)
	router.HandleFunc("GET /ready", readyHandler)

	// Workflows
	router.HandleFunc("GET /api/v1/workflows", handler.ListWorkflows)
	router.HandleFunc("POST /api/v1/workflows", handler.CreateWorkflow)
	router.HandleFunc("GET /api/v1/workflows/{id}", handler.GetWorkflow)
	router.HandleFunc("PUT /api/v1/workflows/{id}", handler.UpdateWorkflow)
	router.HandleFunc("DELETE /api/v1/workflows/{id}", handler.DeleteWorkflow)
	router.HandleFunc("POST /api/v1/workflows/{id}/activate", handler.ActivateWorkflow)
	router.HandleFunc("POST /api/v1/workflows/{id}/deactivate", handler.DeactivateWorkflow)
	router.HandleFunc("POST /api/v1/workflows/{id}/trigger", handler.TriggerWorkflow)

	// Workflow Runs
	router.HandleFunc("GET /api/v1/workflow-runs", handler.ListWorkflowRuns)
	router.HandleFunc("GET /api/v1/workflow-runs/{id}", handler.GetWorkflowRun)
	router.HandleFunc("POST /api/v1/workflow-runs/{id}/cancel", handler.CancelWorkflowRun)

	return router
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy","service":"workflow-service"}`))
}

func readyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ready","service":"workflow-service"}`))
}

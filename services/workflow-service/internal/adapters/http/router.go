package http

import (
	nethttp "net/http"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/workflow-service/internal/application"
)

func SetupRoutes(
	router *nethttp.ServeMux,
	workflowService *application.WorkflowService,
	triggerService *application.TriggerService,
	actionExecutor *application.ActionExecutor,
	log logger.Logger,
) {
	h := NewHandlers(workflowService, triggerService, actionExecutor, log)

	// Workflow Definitions
	router.HandleFunc("POST /api/v1/workflows/definitions", h.CreateDefinition)
	router.HandleFunc("GET /api/v1/workflows/definitions", h.ListDefinitions)
	router.HandleFunc("GET /api/v1/workflows/definitions/{id}", h.GetDefinition)
	router.HandleFunc("PUT /api/v1/workflows/definitions/{id}", h.UpdateDefinition)
	router.HandleFunc("DELETE /api/v1/workflows/definitions/{id}", h.DeleteDefinition)

	// Templates
	router.HandleFunc("GET /api/v1/workflows/templates", h.ListTemplates)

	// Workflow Instances
	router.HandleFunc("POST /api/v1/workflows/definitions/{id}/instantiate", h.InstantiateWorkflow)
	router.HandleFunc("GET /api/v1/workflows/instances", h.ListInstances)
	router.HandleFunc("GET /api/v1/workflows/instances/{id}", h.GetInstance)
	router.HandleFunc("POST /api/v1/workflows/instances/{id}/cancel", h.CancelInstance)

	// Trigger & Dashboard
	router.HandleFunc("POST /api/v1/workflows/trigger", h.TriggerWorkflow)
	router.HandleFunc("GET /api/v1/workflows/dashboard", h.GetDashboard)

	// Alias routes for frontend compatibility (without /definitions sub-path)
	router.HandleFunc("GET /api/v1/workflows", h.ListDefinitions)
	router.HandleFunc("POST /api/v1/workflows", h.CreateDefinition)
	router.HandleFunc("GET /api/v1/workflows/{id}", h.GetDefinition)
	router.HandleFunc("POST /api/v1/workflows/{id}/execute", h.InstantiateWorkflow)
}

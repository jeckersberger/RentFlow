package http

import (
	"net/http"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/document-service/internal/application"
)

func NewRouter(
	docSvc *application.DocumentService,
	tplSvc *application.TemplateService,
	logger logger.Logger,
) *http.ServeMux {
	router := http.NewServeMux()
	handler := NewHandler(docSvc, tplSvc, logger)

	// Health & readiness
	router.HandleFunc("GET /health", healthHandler)
	router.HandleFunc("GET /ready", readyHandler)

	// Document routes
	router.HandleFunc("POST /api/v1/documents", handler.CreateDocument)
	router.HandleFunc("GET /api/v1/documents", handler.ListDocuments)
	router.HandleFunc("GET /api/v1/documents/{id}", handler.GetDocument)
	router.HandleFunc("DELETE /api/v1/documents/{id}", handler.DeleteDocument)
	router.HandleFunc("GET /api/v1/documents/entity/{type}/{id}", handler.ListDocumentsByEntity)

	// Template routes
	router.HandleFunc("POST /api/v1/templates", handler.CreateTemplate)
	router.HandleFunc("GET /api/v1/templates", handler.ListTemplates)
	router.HandleFunc("GET /api/v1/templates/{id}", handler.GetTemplate)
	router.HandleFunc("PUT /api/v1/templates/{id}", handler.UpdateTemplate)
	router.HandleFunc("DELETE /api/v1/templates/{id}", handler.DeleteTemplate)
	router.HandleFunc("POST /api/v1/templates/{id}/preview", handler.GetTemplatePreview)

	return router
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy","service":"document-service"}`))
}

func readyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ready","service":"document-service"}`))
}

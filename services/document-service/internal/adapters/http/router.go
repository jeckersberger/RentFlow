package http

import (
	"net/http"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/document-service/internal/application"
)

func NewRouter(
	docSvc *application.DocumentService,
	sigSvc *application.SignatureService,
	chkSvc *application.ChecksumService,
	log logger.Logger,
) *http.ServeMux {
	router := http.NewServeMux()
	handler := NewHandler(docSvc, sigSvc, chkSvc, log)

	router.HandleFunc("GET /health", healthHandler)
	router.HandleFunc("GET /ready", readyHandler)

	// Document endpoints
	router.HandleFunc("POST /api/v1/documents", handler.CreateDocument)
	router.HandleFunc("GET /api/v1/documents", handler.ListDocuments)
	router.HandleFunc("GET /api/v1/documents/{id}", handler.GetDocument)
	router.HandleFunc("PUT /api/v1/documents/{id}", handler.UpdateDocument)
	router.HandleFunc("POST /api/v1/documents/{id}/generate", handler.GenerateDocument)
	router.HandleFunc("POST /api/v1/documents/{id}/archive", handler.ArchiveDocument)

	// Signature endpoints
	router.HandleFunc("POST /api/v1/documents/{id}/sign", handler.RequestSignature)
	router.HandleFunc("POST /api/v1/documents/{docId}/signatures/{sigId}", handler.SubmitSignature)
	router.HandleFunc("GET /api/v1/documents/{id}/signatures", handler.GetSignatures)

	// Version endpoints
	router.HandleFunc("GET /api/v1/documents/{id}/versions", handler.GetVersions)

	// Delivery note from project
	router.HandleFunc("POST /api/v1/documents/from-project/{projectId}", handler.GenerateDeliveryNote)

	// GoBD verification
	router.HandleFunc("GET /api/v1/documents/verify-chain", handler.VerifyChecksumChain)

	// Scan upload endpoint
	router.HandleFunc("POST /api/v1/documents/upload", handler.UploadScan)

	// Public signature endpoints (no tenant required)
	router.HandleFunc("GET /api/v1/public/sign/{token}", handler.GetPublicSignaturePage)
	router.HandleFunc("POST /api/v1/public/sign/{token}", handler.SubmitPublicSignature)

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

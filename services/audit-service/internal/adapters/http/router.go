package http

import (
	nethttp "net/http"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/audit-service/internal/application"
)

func SetupRoutes(
	router *nethttp.ServeMux,
	auditService *application.AuditService,
	verificationService *application.VerificationService,
	pseudonymizationService *application.PseudonymizationService,
	exportService *application.ExportService,
	log logger.Logger,
) {
	handler := NewAuditHandler(auditService, verificationService, pseudonymizationService, exportService, log)

	router.HandleFunc("POST /api/v1/audit/entries", handler.WriteEntry)
	router.HandleFunc("GET /api/v1/audit/entries", handler.ListEntries)
	router.HandleFunc("GET /api/v1/audit/entries/{id}", handler.GetEntry)
	router.HandleFunc("POST /api/v1/audit/verify", handler.VerifyChain)
	router.HandleFunc("POST /api/v1/audit/pseudonymize/{userID}", handler.PseudonymizeUser)
	router.HandleFunc("POST /api/v1/audit/export", handler.CreateExport)
	router.HandleFunc("GET /api/v1/audit/exports", handler.ListExports)
	router.HandleFunc("GET /api/v1/audit/exports/{id}", handler.GetExport)
	router.HandleFunc("GET /api/v1/audit/dashboard", handler.GetDashboard)
}

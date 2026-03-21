package http

import (
	"net/http"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/insurance-service/internal/application"
)

func NewRouter(
	policySvc *application.PolicyService,
	claimSvc *application.ClaimService,
	riskSvc *application.RiskService,
	log *logger.Logger,
) *http.ServeMux {
	router := http.NewServeMux()
	handler := NewHandler(policySvc, claimSvc, riskSvc, log)

	router.HandleFunc("GET /health", healthHandler)
	router.HandleFunc("GET /ready", readyHandler)

	// Policies
	router.HandleFunc("GET /api/v1/policies", handler.ListPolicies)
	router.HandleFunc("POST /api/v1/policies", handler.CreatePolicy)
	router.HandleFunc("GET /api/v1/policies/{id}", handler.GetPolicy)
	router.HandleFunc("PUT /api/v1/policies/{id}", handler.UpdatePolicy)
	router.HandleFunc("DELETE /api/v1/policies/{id}", handler.DeletePolicy)
	router.HandleFunc("GET /api/v1/policies/expiring", handler.ListExpiringPolicies)

	// Claims
	router.HandleFunc("POST /api/v1/claims", handler.FileClaim)
	router.HandleFunc("GET /api/v1/claims", handler.ListClaims)
	router.HandleFunc("GET /api/v1/claims/{id}", handler.GetClaim)
	router.HandleFunc("PUT /api/v1/claims/{id}", handler.UpdateClaim)
	router.HandleFunc("POST /api/v1/claims/{id}/approve", handler.ApproveClaim)
	router.HandleFunc("POST /api/v1/claims/{id}/reject", handler.RejectClaim)

	// Risk assessments
	router.HandleFunc("GET /api/v1/risk-analysis/equipment/{id}", handler.GetRiskAssessment)
	router.HandleFunc("POST /api/v1/risk-analysis/assess", handler.AssessRisk)

	return router
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy","service":"insurance-service"}`))
}

func readyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ready","service":"insurance-service"}`))
}

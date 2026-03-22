package http

import "net/http"

// RegisterRoutes registers all HTTP routes
func RegisterRoutes(mux *http.ServeMux, handler *Handler) {
	// Policy routes
	mux.HandleFunc("POST /api/v1/insurance/policies", handler.CreatePolicy)
	mux.HandleFunc("GET /api/v1/insurance/policies", handler.GetTenantPolicies)
	mux.HandleFunc("GET /api/v1/insurance/policies/active", handler.GetActivePolicies)
	mux.HandleFunc("GET /api/v1/insurance/policies/{id}", handler.GetPolicy)
	mux.HandleFunc("PUT /api/v1/insurance/policies/{id}", handler.UpdatePolicy)

	// Claim routes
	mux.HandleFunc("POST /api/v1/insurance/claims", handler.CreateClaim)
	mux.HandleFunc("GET /api/v1/insurance/claims", handler.GetTenantClaims)
	mux.HandleFunc("GET /api/v1/insurance/claims/{id}", handler.GetClaim)
	mux.HandleFunc("PUT /api/v1/insurance/claims/{id}", handler.UpdateClaim)
	mux.HandleFunc("POST /api/v1/insurance/claims/{id}/submit", handler.SubmitClaim)
	mux.HandleFunc("POST /api/v1/insurance/claims/{id}/approve", handler.ApproveClaim)
	mux.HandleFunc("POST /api/v1/insurance/claims/{id}/reject", handler.RejectClaim)
	mux.HandleFunc("POST /api/v1/insurance/claims/{id}/settle", handler.SettleClaim)

	// Claim items routes
	mux.HandleFunc("GET /api/v1/insurance/claims/{id}/items", handler.GetClaimItems)
	mux.HandleFunc("POST /api/v1/insurance/claims/{id}/items", handler.CreateClaimItem)

	// Dashboard routes
	mux.HandleFunc("GET /api/v1/insurance/dashboard", handler.GetDashboard)

	// Health routes
	mux.HandleFunc("GET /health", handler.Health)
	mux.HandleFunc("GET /ready", handler.Ready)
}

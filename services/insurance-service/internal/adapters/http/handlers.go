package http

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/jeckersberger/rentflow/services/insurance-service/internal/application"
)

// Handler contains HTTP handlers for the insurance service
type Handler struct {
	policyService *application.PolicyService
	claimService  *application.ClaimService
	logger        zerolog.Logger
}

// NewHandler creates a new HTTP handler
func NewHandler(policyService *application.PolicyService, claimService *application.ClaimService, logger zerolog.Logger) *Handler {
	return &Handler{
		policyService: policyService,
		claimService:  claimService,
		logger:        logger,
	}
}

// Helper function to extract UUID from URL
func extractUUID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}

// Helper function to write JSON response
func writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// Helper function to write error response
func writeError(w http.ResponseWriter, statusCode int, message string) {
	writeJSON(w, statusCode, map[string]string{"error": message})
}

// CreatePolicy creates a new policy
func (h *Handler) CreatePolicy(w http.ResponseWriter, r *http.Request) {
	var req application.CreatePolicyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	policy, err := h.policyService.CreatePolicy(r.Context(), req)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to create policy")
		writeError(w, http.StatusInternalServerError, "failed to create policy")
		return
	}

	writeJSON(w, http.StatusCreated, policy)
}

// GetPolicy retrieves a policy
func (h *Handler) GetPolicy(w http.ResponseWriter, r *http.Request) {
	id, err := extractUUID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid policy ID")
		return
	}

	policy, err := h.policyService.GetPolicy(r.Context(), id)
	if err != nil {
		h.logger.Error().Err(err).Str("id", id.String()).Msg("failed to get policy")
		writeError(w, http.StatusNotFound, "policy not found")
		return
	}

	writeJSON(w, http.StatusOK, policy)
}

// GetTenantPolicies retrieves all policies for a tenant
func (h *Handler) GetTenantPolicies(w http.ResponseWriter, r *http.Request) {
	tenantID, err := extractUUID(r.URL.Query().Get("tenant_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid tenant ID")
		return
	}

	policies, err := h.policyService.GetTenantPolicies(r.Context(), tenantID)
	if err != nil {
		h.logger.Error().Err(err).Str("tenant_id", tenantID.String()).Msg("failed to get tenant policies")
		writeError(w, http.StatusInternalServerError, "failed to get policies")
		return
	}

	writeJSON(w, http.StatusOK, policies)
}

// GetActivePolicies retrieves active policies for a tenant
func (h *Handler) GetActivePolicies(w http.ResponseWriter, r *http.Request) {
	tenantID, err := extractUUID(r.URL.Query().Get("tenant_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid tenant ID")
		return
	}

	policies, err := h.policyService.GetActivePolicies(r.Context(), tenantID)
	if err != nil {
		h.logger.Error().Err(err).Str("tenant_id", tenantID.String()).Msg("failed to get active policies")
		writeError(w, http.StatusInternalServerError, "failed to get active policies")
		return
	}

	writeJSON(w, http.StatusOK, policies)
}

// UpdatePolicy updates a policy
func (h *Handler) UpdatePolicy(w http.ResponseWriter, r *http.Request) {
	id, err := extractUUID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid policy ID")
		return
	}

	var req application.UpdatePolicyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	policy, err := h.policyService.UpdatePolicy(r.Context(), id, req)
	if err != nil {
		h.logger.Error().Err(err).Str("id", id.String()).Msg("failed to update policy")
		writeError(w, http.StatusInternalServerError, "failed to update policy")
		return
	}

	writeJSON(w, http.StatusOK, policy)
}

// CreateClaim creates a new claim
func (h *Handler) CreateClaim(w http.ResponseWriter, r *http.Request) {
	var req application.CreateClaimRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	claim, err := h.claimService.CreateClaim(r.Context(), req)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to create claim")
		writeError(w, http.StatusInternalServerError, "failed to create claim")
		return
	}

	writeJSON(w, http.StatusCreated, claim)
}

// GetClaim retrieves a claim
func (h *Handler) GetClaim(w http.ResponseWriter, r *http.Request) {
	id, err := extractUUID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid claim ID")
		return
	}

	claim, err := h.claimService.GetClaim(r.Context(), id)
	if err != nil {
		h.logger.Error().Err(err).Str("id", id.String()).Msg("failed to get claim")
		writeError(w, http.StatusNotFound, "claim not found")
		return
	}

	writeJSON(w, http.StatusOK, claim)
}

// GetTenantClaims retrieves all claims for a tenant
func (h *Handler) GetTenantClaims(w http.ResponseWriter, r *http.Request) {
	tenantID, err := extractUUID(r.URL.Query().Get("tenant_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid tenant ID")
		return
	}

	claims, err := h.claimService.GetTenantClaims(r.Context(), tenantID)
	if err != nil {
		h.logger.Error().Err(err).Str("tenant_id", tenantID.String()).Msg("failed to get tenant claims")
		writeError(w, http.StatusInternalServerError, "failed to get claims")
		return
	}

	writeJSON(w, http.StatusOK, claims)
}

// UpdateClaim updates a claim
func (h *Handler) UpdateClaim(w http.ResponseWriter, r *http.Request) {
	id, err := extractUUID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid claim ID")
		return
	}

	var req application.UpdateClaimRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	claim, err := h.claimService.UpdateClaim(r.Context(), id, req)
	if err != nil {
		h.logger.Error().Err(err).Str("id", id.String()).Msg("failed to update claim")
		writeError(w, http.StatusInternalServerError, "failed to update claim")
		return
	}

	writeJSON(w, http.StatusOK, claim)
}

// SubmitClaim submits a claim for review
func (h *Handler) SubmitClaim(w http.ResponseWriter, r *http.Request) {
	id, err := extractUUID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid claim ID")
		return
	}

	claim, err := h.claimService.SubmitClaim(r.Context(), id)
	if err != nil {
		h.logger.Error().Err(err).Str("id", id.String()).Msg("failed to submit claim")
		writeError(w, http.StatusInternalServerError, "failed to submit claim")
		return
	}

	writeJSON(w, http.StatusOK, claim)
}

// ApproveClaim approves a claim
func (h *Handler) ApproveClaim(w http.ResponseWriter, r *http.Request) {
	id, err := extractUUID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid claim ID")
		return
	}

	var req application.ApproveClaimRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	claim, err := h.claimService.ApproveClaim(r.Context(), id, req.ApprovedAmount, req.AdjusterNotes)
	if err != nil {
		h.logger.Error().Err(err).Str("id", id.String()).Msg("failed to approve claim")
		writeError(w, http.StatusInternalServerError, "failed to approve claim")
		return
	}

	writeJSON(w, http.StatusOK, claim)
}

// RejectClaim rejects a claim
func (h *Handler) RejectClaim(w http.ResponseWriter, r *http.Request) {
	id, err := extractUUID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid claim ID")
		return
	}

	var req application.RejectClaimRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	claim, err := h.claimService.RejectClaim(r.Context(), id, req.AdjusterNotes)
	if err != nil {
		h.logger.Error().Err(err).Str("id", id.String()).Msg("failed to reject claim")
		writeError(w, http.StatusInternalServerError, "failed to reject claim")
		return
	}

	writeJSON(w, http.StatusOK, claim)
}

// SettleClaim settles a claim
func (h *Handler) SettleClaim(w http.ResponseWriter, r *http.Request) {
	id, err := extractUUID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid claim ID")
		return
	}

	var req application.SettleClaimRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	claim, err := h.claimService.SettleClaim(r.Context(), id, req.SettledAmount, req.AdjusterNotes)
	if err != nil {
		h.logger.Error().Err(err).Str("id", id.String()).Msg("failed to settle claim")
		writeError(w, http.StatusInternalServerError, "failed to settle claim")
		return
	}

	writeJSON(w, http.StatusOK, claim)
}

// CreateClaimItem adds an item to a claim
func (h *Handler) CreateClaimItem(w http.ResponseWriter, r *http.Request) {
	claimID, err := extractUUID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid claim ID")
		return
	}

	var req application.CreateClaimItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	item, err := h.claimService.CreateClaimItem(r.Context(), claimID, req)
	if err != nil {
		h.logger.Error().Err(err).Str("claim_id", claimID.String()).Msg("failed to create claim item")
		writeError(w, http.StatusInternalServerError, "failed to create claim item")
		return
	}

	writeJSON(w, http.StatusCreated, item)
}

// GetClaimItems retrieves all items for a claim
func (h *Handler) GetClaimItems(w http.ResponseWriter, r *http.Request) {
	claimID, err := extractUUID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid claim ID")
		return
	}

	items, err := h.claimService.GetClaimItems(r.Context(), claimID)
	if err != nil {
		h.logger.Error().Err(err).Str("claim_id", claimID.String()).Msg("failed to get claim items")
		writeError(w, http.StatusInternalServerError, "failed to get claim items")
		return
	}

	writeJSON(w, http.StatusOK, items)
}

// GetDashboard returns dashboard statistics
func (h *Handler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	tenantID, err := extractUUID(r.URL.Query().Get("tenant_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid tenant ID")
		return
	}

	claims, err := h.claimService.GetTenantClaims(r.Context(), tenantID)
	if err != nil {
		h.logger.Error().Err(err).Str("tenant_id", tenantID.String()).Msg("failed to get dashboard data")
		writeError(w, http.StatusInternalServerError, "failed to get dashboard data")
		return
	}

	policies, err := h.policyService.GetTenantPolicies(r.Context(), tenantID)
	if err != nil {
		h.logger.Error().Err(err).Str("tenant_id", tenantID.String()).Msg("failed to get dashboard data")
		writeError(w, http.StatusInternalServerError, "failed to get dashboard data")
		return
	}

	response := h.buildDashboardResponse(claims, policies)
	writeJSON(w, http.StatusOK, response)
}

// buildDashboardResponse builds the dashboard statistics response
func (h *Handler) buildDashboardResponse(claims []*application.ClaimResponse, policies []*application.PolicyResponse) *application.DashboardStatsResponse {
	yearlyClaimStats := make(map[int]*application.YearlyClaimStats)
	premiumDev := make(map[int]*application.PremiumDevelopment)

	// Process claims by year
	for _, claim := range claims {
		year := claim.CreatedAt.Year()
		if _, exists := yearlyClaimStats[year]; !exists {
			yearlyClaimStats[year] = &application.YearlyClaimStats{
				Year:       year,
				ClaimCount: 0,
				TotalClaimed: 0,
				TotalApproved: 0,
			}
		}
		stats := yearlyClaimStats[year]
		stats.ClaimCount++
		stats.TotalClaimed += claim.ClaimedAmount
		if claim.ApprovedAmount != nil {
			stats.TotalApproved += *claim.ApprovedAmount
		}
	}

	// Process policies by year
	for _, policy := range policies {
		year := policy.StartDate.Year()
		if _, exists := premiumDev[year]; !exists {
			premiumDev[year] = &application.PremiumDevelopment{
				Year:       year,
				AnnualPremium: 0,
				ActivePolicies: 0,
			}
		}
		dev := premiumDev[year]
		dev.AnnualPremium += policy.PremiumAnnual
		dev.ActivePolicies++
	}

	// Convert maps to slices
	yearlyStats := make([]application.YearlyClaimStats, 0, len(yearlyClaimStats))
	for _, stats := range yearlyClaimStats {
		yearlyStats = append(yearlyStats, *stats)
	}

	premiumStats := make([]application.PremiumDevelopment, 0, len(premiumDev))
	for _, dev := range premiumDev {
		premiumStats = append(premiumStats, *dev)
	}

	return &application.DashboardStatsResponse{
		YearlyStats:       yearlyStats,
		PremiumDevelopment: premiumStats,
	}
}

// Health checks service health
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Ready checks if service is ready
func (h *Handler) Ready(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

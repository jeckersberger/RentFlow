package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/insurance-service/internal/application"
	"github.com/jeckersberger/rentflow/services/insurance-service/internal/domain"
)

type Handler struct {
	policySvc *application.PolicyService
	claimSvc  *application.ClaimService
	riskSvc   *application.RiskService
	logger    logger.Logger
}

func NewHandler(
	policySvc *application.PolicyService,
	claimSvc *application.ClaimService,
	riskSvc *application.RiskService,
	log logger.Logger,
) *Handler {
	return &Handler{
		policySvc: policySvc,
		claimSvc:  claimSvc,
		riskSvc:   riskSvc,
		logger:    log,
	}
}

// Policy endpoints
func (h *Handler) CreatePolicy(w http.ResponseWriter, r *http.Request) {
	var req application.CreatePolicyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	resp, err := h.policySvc.CreatePolicy(r.Context(), tenantID, req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusCreated, resp)
}

func (h *Handler) ListPolicies(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	resp, err := h.policySvc.ListPolicies(r.Context(), tenantID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{"data": resp})
}

func (h *Handler) GetPolicy(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	resp, err := h.policySvc.GetPolicy(r.Context(), tenantID, id)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, resp)
}

func (h *Handler) UpdatePolicy(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req application.UpdatePolicyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	resp, err := h.policySvc.UpdatePolicy(r.Context(), tenantID, id, req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, resp)
}

func (h *Handler) DeletePolicy(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	err := h.policySvc.DeletePolicy(r.Context(), tenantID, id)
	if err != nil {
		h.handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListExpiringPolicies(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	days := 30
	if daysStr := r.URL.Query().Get("days"); daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil {
			days = d
		}
	}

	resp, err := h.policySvc.ListExpiringPolicies(r.Context(), tenantID, days)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{"data": resp})
}

// Claim endpoints
func (h *Handler) FileClaim(w http.ResponseWriter, r *http.Request) {
	var req application.CreateClaimRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	resp, err := h.claimSvc.FileClaim(r.Context(), tenantID, req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusCreated, resp)
}

func (h *Handler) ListClaims(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	resp, err := h.claimSvc.ListClaims(r.Context(), tenantID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{"data": resp})
}

func (h *Handler) GetClaim(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	resp, err := h.claimSvc.GetClaim(r.Context(), tenantID, id)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, resp)
}

func (h *Handler) UpdateClaim(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req application.UpdateClaimRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	resp, err := h.claimSvc.UpdateClaim(r.Context(), tenantID, id, req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, resp)
}

func (h *Handler) ApproveClaim(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req application.ApproveClaimRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	resp, err := h.claimSvc.ApproveClaim(r.Context(), tenantID, id, req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, resp)
}

func (h *Handler) RejectClaim(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req application.RejectClaimRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req.Reason = "No reason provided"
	}

	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	resp, err := h.claimSvc.RejectClaim(r.Context(), tenantID, id, req.Reason)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, resp)
}

// Risk assessment endpoints
func (h *Handler) AssessRisk(w http.ResponseWriter, r *http.Request) {
	var req application.AssessRiskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	resp, err := h.riskSvc.AssessEquipment(r.Context(), tenantID, req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusCreated, resp)
}

func (h *Handler) GetRiskAssessment(w http.ResponseWriter, r *http.Request) {
	equipmentID := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	resp, err := h.riskSvc.GetRiskAssessment(r.Context(), tenantID, equipmentID)
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
	if err == domain.ErrPolicyNotFound || err == domain.ErrClaimNotFound || err == domain.ErrRiskAssessmentNotFound {
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

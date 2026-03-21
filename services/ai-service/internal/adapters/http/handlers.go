package http

import (
	"encoding/json"
	"net/http"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/ai-service/internal/application"
	"github.com/jeckersberger/rentflow/services/ai-service/internal/domain"
)

type Handler struct {
	aiSvc  *application.AIService
	logger logger.Logger
}

func NewHandler(aiSvc *application.AIService, log logger.Logger) *Handler {
	return &Handler{
		aiSvc:  aiSvc,
		logger: log,
	}
}

func (h *Handler) Predict(w http.ResponseWriter, r *http.Request) {
	var req application.PredictRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	resp, err := h.aiSvc.Predict(r.Context(), tenantID, req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, resp)
}

func (h *Handler) Classify(w http.ResponseWriter, r *http.Request) {
	var req application.ClassifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	resp, err := h.aiSvc.Classify(r.Context(), tenantID, req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, resp)
}

func (h *Handler) Anonymize(w http.ResponseWriter, r *http.Request) {
	var req application.AnonymizeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	resp, err := h.aiSvc.Anonymize(r.Context(), tenantID, req.Text)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, resp)
}

func (h *Handler) Deanonymize(w http.ResponseWriter, r *http.Request) {
	var req application.DeanonymizeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	resp, err := h.aiSvc.Deanonymize(r.Context(), tenantID, req.Text, req.MappingID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, resp)
}

func (h *Handler) Suggest(w http.ResponseWriter, r *http.Request) {
	var req application.SuggestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	resp, err := h.aiSvc.Suggest(r.Context(), tenantID, req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, resp)
}

func (h *Handler) GetProviders(w http.ResponseWriter, r *http.Request) {
	providers, err := h.aiSvc.GetProviders(r.Context())
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{"providers": providers})
}

func (h *Handler) GetUsage(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	stats, err := h.aiSvc.GetUsageStats(r.Context(), tenantID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, stats)
}

func (h *Handler) GetAnonymizationRules(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	rules, err := h.aiSvc.GetAnonymizationRules(r.Context(), tenantID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{"rules": rules})
}

func (h *Handler) CreateAnonymizationRule(w http.ResponseWriter, r *http.Request) {
	var req application.AnonymizationRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	resp, err := h.aiSvc.CreateAnonymizationRule(r.Context(), tenantID, req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusCreated, resp)
}

func (h *Handler) DeleteAnonymizationRule(w http.ResponseWriter, r *http.Request) {
	ruleID := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	err := h.aiSvc.DeleteAnonymizationRule(r.Context(), tenantID, ruleID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
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
	if err == domain.ErrAIRequestNotFound || err == domain.ErrAnonymizationRuleNotFound {
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

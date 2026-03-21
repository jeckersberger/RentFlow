package http

import (
	"encoding/json"
	"net/http"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/reporting-service/internal/application"
	"github.com/jeckersberger/rentflow/services/reporting-service/internal/domain"
)

func NewRouter(svc *application.ReportService, log logger.Logger) *http.ServeMux {
	router := http.NewServeMux()
	h := &Handler{svc: svc, logger: log}

	router.HandleFunc("GET /health", healthHandler)
	router.HandleFunc("GET /ready", readyHandler)

	router.HandleFunc("GET /api/v1/reports/kpi", h.GetKPI)
	router.HandleFunc("GET /api/v1/reports/revenue", h.ListReports)
	router.HandleFunc("GET /api/v1/reports/utilization", h.ListReports)
	router.HandleFunc("GET /api/v1/reports/inventory-value", h.ListReports)
	router.HandleFunc("POST /api/v1/reports/generate", h.GenerateReport)
	router.HandleFunc("GET /api/v1/reports/{id}", h.GetReport)
	router.HandleFunc("GET /api/v1/reports/equipment/{id}/history", h.GetEquipmentHistory)

	return router
}

type Handler struct {
	svc    *application.ReportService
	logger logger.Logger
}

func (h *Handler) GetKPI(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	kpis, err := h.svc.GetKPIs(r.Context(), tenantID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{"kpis": kpis})
}

func (h *Handler) GenerateReport(w http.ResponseWriter, r *http.Request) {
	var cmd application.GenerateReportCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cmd.TenantID = r.Header.Get("X-Tenant-ID")
	if cmd.TenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	dto, err := h.svc.GenerateReport(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusCreated, dto)
}

func (h *Handler) GetReport(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	dto, err := h.svc.GetReport(r.Context(), tenantID, id)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, dto)
}

func (h *Handler) ListReports(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	dtos, err := h.svc.ListReports(r.Context(), tenantID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{"data": dtos})
}

func (h *Handler) GetEquipmentHistory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"data":[]}`))
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
	if err == domain.ErrReportNotFound {
		h.respondError(w, http.StatusNotFound, err.Error())
		return
	}
	h.logger.Error("Error", err)
	h.respondError(w, http.StatusInternalServerError, "internal server error")
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy","service":"reporting-service"}`))
}

func readyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ready","service":"reporting-service"}`))
}

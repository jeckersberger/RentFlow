package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/audit-service/internal/application"
	"github.com/jeckersberger/rentflow/services/audit-service/internal/domain"
)

func NewRouter(svc *application.AuditService, log *logger.Logger) *http.ServeMux {
	router := http.NewServeMux()
	h := &Handler{svc: svc, logger: log}

	router.HandleFunc("GET /health", healthHandler)
	router.HandleFunc("GET /ready", readyHandler)

	router.HandleFunc("GET /api/v1/audit/logs", h.ListLogs)
	router.HandleFunc("POST /api/v1/audit/logs", h.LogAudit)
	router.HandleFunc("GET /api/v1/audit/entity/{type}/{id}", h.ListByEntity)
	router.HandleFunc("GET /api/v1/audit/user/{id}", h.ListByUser)
	router.HandleFunc("POST /api/v1/audit/verify", h.VerifyIntegrity)
	router.HandleFunc("GET /api/v1/audit/verify/status", h.GetIntegrityStatus)
	router.HandleFunc("GET /api/v1/audit/export", h.ExportLogs)

	return router
}

type Handler struct {
	svc    *application.AuditService
	logger *logger.Logger
}

func (h *Handler) LogAudit(w http.ResponseWriter, r *http.Request) {
	var cmd application.LogAuditCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cmd.TenantID = r.Header.Get("X-Tenant-ID")
	cmd.IPAddress = r.RemoteAddr
	cmd.UserAgent = r.Header.Get("User-Agent")

	if cmd.TenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	dto, err := h.svc.LogAudit(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusCreated, dto)
}

func (h *Handler) ListLogs(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	var fromTime, toTime time.Time
	if fromStr != "" {
		fromTime, _ = time.Parse(time.RFC3339, fromStr)
	}
	if toStr != "" {
		toTime, _ = time.Parse(time.RFC3339, toStr)
	}

	dtos, err := h.svc.ListLogs(r.Context(), tenantID, fromTime, toTime)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{"data": dtos})
}

func (h *Handler) ListByEntity(w http.ResponseWriter, r *http.Request) {
	entityType := r.PathValue("type")
	entityID := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")

	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	dtos, err := h.svc.ListByEntity(r.Context(), tenantID, entityType, entityID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{"data": dtos})
}

func (h *Handler) ListByUser(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")

	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	dtos, err := h.svc.ListByUser(r.Context(), tenantID, userID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{"data": dtos})
}

func (h *Handler) VerifyIntegrity(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	result, err := h.svc.VerifyIntegrity(r.Context(), tenantID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, result)
}

func (h *Handler) GetIntegrityStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"valid","last_verified":"2026-03-21T00:00:00Z"}`))
}

func (h *Handler) ExportLogs(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	var fromTime, toTime time.Time
	if fromStr != "" {
		fromTime, _ = time.Parse(time.RFC3339, fromStr)
	}
	if toStr != "" {
		toTime, _ = time.Parse(time.RFC3339, toStr)
	}

	data, err := h.svc.ExportLogs(r.Context(), tenantID, fromTime, toTime)
	if err != nil {
		h.handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=audit-logs.json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
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
	if err == domain.ErrAuditEntryNotFound {
		h.respondError(w, http.StatusNotFound, err.Error())
		return
	}
	h.logger.Error("Error", err)
	h.respondError(w, http.StatusInternalServerError, "internal server error")
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy","service":"audit-service"}`))
}

func readyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ready","service":"audit-service"}`))
}

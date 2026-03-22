package http

import (
	"encoding/json"
	nethttp "net/http"

	"github.com/google/uuid"
	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/audit-service/internal/application"
	"github.com/jeckersberger/rentflow/services/audit-service/internal/domain"
)

type AuditHandler struct {
	auditService             *application.AuditService
	verificationService      *application.VerificationService
	pseudonymizationService  *application.PseudonymizationService
	exportService            *application.ExportService
	log                      logger.Logger
}

func NewAuditHandler(
	auditService *application.AuditService,
	verificationService *application.VerificationService,
	pseudonymizationService *application.PseudonymizationService,
	exportService *application.ExportService,
	log logger.Logger,
) *AuditHandler {
	return &AuditHandler{
		auditService:            auditService,
		verificationService:     verificationService,
		pseudonymizationService: pseudonymizationService,
		exportService:           exportService,
		log:                     log,
	}
}

func (h *AuditHandler) WriteEntry(w nethttp.ResponseWriter, r *nethttp.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, nethttp.StatusBadRequest, "X-Tenant-ID header required")
		return
	}

	var req application.WriteAuditEntryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, nethttp.StatusBadRequest, "Invalid request body")
		return
	}

	tenantUUID, _ := uuid.Parse(tenantID)
	cmd := domain.WriteAuditEntryCmd{
		TenantID:    tenantUUID,
		ServiceName: req.ServiceName,
		Operation:   req.Operation,
		EntityType:  req.EntityType,
		EntityID:    req.EntityID,
		UserID:      req.UserID,
		UserName:    req.UserName,
		OldValues:   req.OldValues,
		NewValues:   req.NewValues,
		IPAddress:   req.IPAddress,
		UserAgent:   req.UserAgent,
	}

	entry, err := h.auditService.WriteEntry(r.Context(), cmd)
	if err != nil {
		h.respondError(w, nethttp.StatusInternalServerError, "Failed to write audit entry")
		return
	}

	h.respondJSON(w, nethttp.StatusCreated, entry)
}

func (h *AuditHandler) ListEntries(w nethttp.ResponseWriter, r *nethttp.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, nethttp.StatusBadRequest, "X-Tenant-ID header required")
		return
	}

	tenantUUID, _ := uuid.Parse(tenantID)
	entries, err := h.auditService.ListByTenant(r.Context(), tenantUUID)
	if err != nil {
		h.respondError(w, nethttp.StatusInternalServerError, "Failed to fetch entries")
		return
	}

	h.respondJSON(w, nethttp.StatusOK, entries)
}

func (h *AuditHandler) GetEntry(w nethttp.ResponseWriter, r *nethttp.Request) {
	idStr := r.PathValue("id")
	entryID, err := uuid.Parse(idStr)
	if err != nil {
		h.respondError(w, nethttp.StatusBadRequest, "Invalid entry ID")
		return
	}

	entry, err := h.auditService.GetEntry(r.Context(), entryID)
	if err != nil {
		h.respondError(w, nethttp.StatusNotFound, "Entry not found")
		return
	}

	h.respondJSON(w, nethttp.StatusOK, entry)
}

func (h *AuditHandler) VerifyChain(w nethttp.ResponseWriter, r *nethttp.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, nethttp.StatusBadRequest, "X-Tenant-ID header required")
		return
	}

	tenantUUID, _ := uuid.Parse(tenantID)
	result, err := h.verificationService.VerifyChain(r.Context(), tenantUUID)
	if err != nil {
		h.respondError(w, nethttp.StatusInternalServerError, "Verification failed")
		return
	}

	response := application.VerificationResponse{
		IsValid:      result.IsValid,
		EntriesCount: result.EntriesCount,
	}

	if result.FirstMismatch != nil {
		response.FirstMismatch = &application.MismatchDetailResponse{
			EntryID:          result.FirstMismatch.EntryID,
			SequenceNumber:   result.FirstMismatch.SequenceNumber,
			ComputedChecksum: result.FirstMismatch.ComputedChecksum,
			StoredChecksum:   result.FirstMismatch.StoredChecksum,
		}
	}

	h.respondJSON(w, nethttp.StatusOK, response)
}

func (h *AuditHandler) PseudonymizeUser(w nethttp.ResponseWriter, r *nethttp.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, nethttp.StatusBadRequest, "X-Tenant-ID header required")
		return
	}

	userIDStr := r.PathValue("userID")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		h.respondError(w, nethttp.StatusBadRequest, "Invalid user ID")
		return
	}

	tenantUUID, _ := uuid.Parse(tenantID)
	cmd := domain.PseudonymizeUserCmd{
		TenantID: tenantUUID,
		UserID:   userID,
	}

	err = h.pseudonymizationService.PseudonymizeUser(r.Context(), cmd)
	if err != nil {
		h.respondError(w, nethttp.StatusInternalServerError, "Pseudonymization failed")
		return
	}

	h.respondJSON(w, nethttp.StatusOK, map[string]string{"status": "pseudonymized"})
}

func (h *AuditHandler) CreateExport(w nethttp.ResponseWriter, r *nethttp.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, nethttp.StatusBadRequest, "X-Tenant-ID header required")
		return
	}

	var req application.CreateExportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, nethttp.StatusBadRequest, "Invalid request body")
		return
	}

	tenantUUID, _ := uuid.Parse(tenantID)
	cmd := domain.CreateExportCmd{
		TenantID:   tenantUUID,
		ExportType: req.ExportType,
		DateFrom:   req.DateFrom,
		DateTo:     req.DateTo,
	}

	export, err := h.exportService.CreateExport(r.Context(), cmd)
	if err != nil {
		h.respondError(w, nethttp.StatusInternalServerError, "Failed to create export")
		return
	}

	h.respondJSON(w, nethttp.StatusCreated, export)
}

func (h *AuditHandler) ListExports(w nethttp.ResponseWriter, r *nethttp.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, nethttp.StatusBadRequest, "X-Tenant-ID header required")
		return
	}

	tenantUUID, _ := uuid.Parse(tenantID)
	exports, err := h.exportService.ListExports(r.Context(), tenantUUID)
	if err != nil {
		h.respondError(w, nethttp.StatusInternalServerError, "Failed to fetch exports")
		return
	}

	h.respondJSON(w, nethttp.StatusOK, exports)
}

func (h *AuditHandler) GetExport(w nethttp.ResponseWriter, r *nethttp.Request) {
	idStr := r.PathValue("id")
	exportID, err := uuid.Parse(idStr)
	if err != nil {
		h.respondError(w, nethttp.StatusBadRequest, "Invalid export ID")
		return
	}

	export, err := h.exportService.GetExport(r.Context(), exportID)
	if err != nil {
		h.respondError(w, nethttp.StatusNotFound, "Export not found")
		return
	}

	h.respondJSON(w, nethttp.StatusOK, export)
}

func (h *AuditHandler) GetDashboard(w nethttp.ResponseWriter, r *nethttp.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, nethttp.StatusBadRequest, "X-Tenant-ID header required")
		return
	}

	tenantUUID, _ := uuid.Parse(tenantID)
	stats, err := h.auditService.GetDashboardStats(r.Context(), tenantUUID)
	if err != nil {
		h.respondError(w, nethttp.StatusInternalServerError, "Failed to get dashboard stats")
		return
	}

	h.respondJSON(w, nethttp.StatusOK, stats)
}

func (h *AuditHandler) respondJSON(w nethttp.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func (h *AuditHandler) respondError(w nethttp.ResponseWriter, statusCode int, message string) {
	h.respondJSON(w, statusCode, map[string]string{"error": message})
}

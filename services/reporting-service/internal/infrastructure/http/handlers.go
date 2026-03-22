package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/jeckersberger/rentflow/services/reporting-service/internal/application"
	"github.com/jeckersberger/rentflow/services/reporting-service/internal/domain"
)

type ReportHandler struct {
	reportService *application.ReportService
	kpiService    *application.KPIService
	logger        zerolog.Logger
}

func NewReportHandler(reportService *application.ReportService, kpiService *application.KPIService, logger zerolog.Logger) *ReportHandler {
	return &ReportHandler{
		reportService: reportService,
		kpiService:    kpiService,
		logger:        logger,
	}
}

func extractTenantID(r *http.Request) (uuid.UUID, error) {
	tenantIDStr := r.Header.Get("X-Tenant-ID")
	if tenantIDStr == "" {
		return uuid.Nil, ErrMissingTenantID
	}
	return uuid.Parse(tenantIDStr)
}

func extractUserID(r *http.Request) (uuid.UUID, error) {
	userIDStr := r.Header.Get("X-User-ID")
	if userIDStr == "" {
		return uuid.Nil, ErrMissingUserID
	}
	return uuid.Parse(userIDStr)
}

func (h *ReportHandler) CreateReportDefinition(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID, err := extractTenantID(r)
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	userID, err := extractUserID(r)
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	var req application.CreateReportDefinitionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	paramsJSON, _ := json.Marshal(req.Parameters)
	recipientsJSON, _ := json.Marshal(req.EmailRecipients)

	cmd := &application.CreateReportDefinitionCommand{
		TenantID:        tenantID,
		Name:            req.Name,
		ReportType:      string(req.ReportType),
		Description:     req.Description,
		Parameters:      paramsJSON,
		ScheduleCron:    req.ScheduleCron,
		EmailRecipients: recipientsJSON,
		Format:          string(req.Format),
		CreatedBy:       userID,
	}

	def, err := h.reportService.CreateDefinition(ctx, cmd)
	if err != nil {
		h.errorResponse(w, http.StatusInternalServerError, "Failed to create report definition")
		return
	}

	resp, _ := application.NewReportDefinitionResponse(def)
	h.jsonResponse(w, http.StatusCreated, resp)
}

func (h *ReportHandler) GetReportDefinition(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID, err := extractTenantID(r)
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	defID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, "Invalid definition ID")
		return
	}

	def, err := h.reportService.GetDefinition(ctx, tenantID, defID)
	if err != nil {
		h.errorResponse(w, http.StatusNotFound, "Report definition not found")
		return
	}

	resp, _ := application.NewReportDefinitionResponse(def)
	h.jsonResponse(w, http.StatusOK, resp)
}

func (h *ReportHandler) UpdateReportDefinition(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID, err := extractTenantID(r)
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	defID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, "Invalid definition ID")
		return
	}

	var req application.UpdateReportDefinitionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	paramsJSON, _ := json.Marshal(req.Parameters)
	recipientsJSON, _ := json.Marshal(req.EmailRecipients)

	cmd := &application.UpdateReportDefinitionCommand{
		ReportDefinitionID: defID,
		TenantID:           tenantID,
		Name:               req.Name,
		Description:        req.Description,
		Parameters:         paramsJSON,
		ScheduleCron:       req.ScheduleCron,
		EmailRecipients:    recipientsJSON,
		Format:             string(req.Format),
		IsActive:           req.IsActive,
	}

	def, err := h.reportService.UpdateDefinition(ctx, cmd)
	if err != nil {
		h.errorResponse(w, http.StatusInternalServerError, "Failed to update report definition")
		return
	}

	resp, _ := application.NewReportDefinitionResponse(def)
	h.jsonResponse(w, http.StatusOK, resp)
}

func (h *ReportHandler) ListReportDefinitions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID, err := extractTenantID(r)
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	definitions, err := h.reportService.ListDefinitions(ctx, tenantID)
	if err != nil {
		h.errorResponse(w, http.StatusInternalServerError, "Failed to list report definitions")
		return
	}

	var responses []application.ReportDefinitionResponse
	for _, def := range definitions {
		resp, _ := application.NewReportDefinitionResponse(&def)
		responses = append(responses, *resp)
	}

	h.jsonResponse(w, http.StatusOK, responses)
}

func (h *ReportHandler) GenerateReport(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID, err := extractTenantID(r)
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	defID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, "Invalid definition ID")
		return
	}

	var req application.GenerateReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	paramsJSON, _ := json.Marshal(req.Parameters)

	cmd := &application.GenerateReportCommand{
		ReportDefinitionID: defID,
		TenantID:           tenantID,
		Parameters:         paramsJSON,
	}

	if req.PeriodStart != nil {
		ps := req.PeriodStart.Format("2006-01-02T15:04:05Z07:00")
		cmd.PeriodStart = &ps
	}

	if req.PeriodEnd != nil {
		pe := req.PeriodEnd.Format("2006-01-02T15:04:05Z07:00")
		cmd.PeriodEnd = &pe
	}

	run, err := h.reportService.GenerateReport(ctx, cmd)
	if err != nil {
		h.errorResponse(w, http.StatusInternalServerError, "Failed to generate report")
		return
	}

	resp, _ := application.NewReportRunResponse(run)
	h.jsonResponse(w, http.StatusAccepted, resp)
}

func (h *ReportHandler) GetReportRun(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID, err := extractTenantID(r)
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	runID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, "Invalid run ID")
		return
	}

	run, err := h.reportService.GetRun(ctx, tenantID, runID)
	if err != nil {
		h.errorResponse(w, http.StatusNotFound, "Report run not found")
		return
	}

	resp, _ := application.NewReportRunResponse(run)
	h.jsonResponse(w, http.StatusOK, resp)
}

func (h *ReportHandler) ListReportRuns(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID, err := extractTenantID(r)
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	runs, err := h.reportService.ListAllRuns(ctx, tenantID)
	if err != nil {
		h.errorResponse(w, http.StatusInternalServerError, "Failed to list report runs")
		return
	}

	var responses []application.ReportRunResponse
	for _, run := range runs {
		resp, _ := application.NewReportRunResponse(&run)
		responses = append(responses, *resp)
	}

	h.jsonResponse(w, http.StatusOK, responses)
}

// ExportReportCSV exports a report run to CSV format
func (h *ReportHandler) ExportReportCSV(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID, err := extractTenantID(r)
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	runID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, "Invalid run ID")
		return
	}

	csvData, err := h.reportService.ExportCSV(ctx, tenantID, runID)
	if err != nil {
		h.errorResponse(w, http.StatusInternalServerError, "Failed to export report as CSV")
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=\"report-"+runID.String()+".csv\"")
	w.WriteHeader(http.StatusOK)
	w.Write(csvData)
}

// ExportReportPDF exports a report run to PDF format
func (h *ReportHandler) ExportReportPDF(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID, err := extractTenantID(r)
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	runID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, "Invalid run ID")
		return
	}

	pdfData, err := h.reportService.ExportPDF(ctx, tenantID, runID)
	if err != nil {
		h.errorResponse(w, http.StatusInternalServerError, "Failed to export report as PDF")
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=\"report-"+runID.String()+".pdf\"")
	w.WriteHeader(http.StatusOK)
	w.Write(pdfData)
}

func (h *ReportHandler) GetKPIDashboard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID, err := extractTenantID(r)
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	period := r.URL.Query().Get("period")
	if period == "" {
		period = string(domain.PeriodMonth)
	}

	cmd := &application.GetKPIDashboardCommand{
		TenantID: tenantID,
		Period:   period,
	}

	dashboard, err := h.kpiService.GetDashboard(ctx, cmd)
	if err != nil {
		h.errorResponse(w, http.StatusInternalServerError, "Failed to get KPI dashboard")
		return
	}

	h.jsonResponse(w, http.StatusOK, dashboard)
}

// GetKPIDashboardByRole returns KPIs filtered by user role
func (h *ReportHandler) GetKPIDashboardByRole(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID, err := extractTenantID(r)
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	role := r.URL.Query().Get("role")
	if role == "" {
		h.errorResponse(w, http.StatusBadRequest, "role query parameter is required")
		return
	}

	period := r.URL.Query().Get("period")
	if period == "" {
		period = string(domain.PeriodMonth)
	}

	cmd := &application.GetKPIDashboardCommand{
		TenantID: tenantID,
		Period:   period,
	}

	dashboard, err := h.kpiService.GetDashboardByRole(ctx, cmd, role)
	if err != nil {
		h.errorResponse(w, http.StatusInternalServerError, "Failed to get KPI dashboard")
		return
	}

	h.jsonResponse(w, http.StatusOK, dashboard)
}

func (h *ReportHandler) GetKPITrends(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID, err := extractTenantID(r)
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	kpiType := r.URL.Query().Get("kpi_type")
	if kpiType == "" {
		h.errorResponse(w, http.StatusBadRequest, "kpi_type is required")
		return
	}

	periodsStr := r.URL.Query().Get("periods")
	periods := 12
	if periodsStr != "" {
		if p, err := strconv.Atoi(periodsStr); err == nil && p > 0 {
			periods = p
		}
	}

	cmd := &application.GetKPITrendsCommand{
		TenantID: tenantID,
		KPIType:  kpiType,
		Periods:  periods,
	}

	trends, err := h.kpiService.GetTrends(ctx, cmd)
	if err != nil {
		h.errorResponse(w, http.StatusInternalServerError, "Failed to get KPI trends")
		return
	}

	h.jsonResponse(w, http.StatusOK, trends)
}

func (h *ReportHandler) GetKPISnapshot(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID, err := extractTenantID(r)
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	cmd := &application.SnapshotKPIsCommand{
		TenantID: tenantID,
	}

	if err := h.kpiService.SnapshotKPIs(ctx, cmd); err != nil {
		h.errorResponse(w, http.StatusInternalServerError, "Failed to snapshot KPIs")
		return
	}

	h.jsonResponse(w, http.StatusOK, map[string]string{"status": "snapshots created"})
}

func (h *ReportHandler) Health(w http.ResponseWriter, r *http.Request) {
	h.jsonResponse(w, http.StatusOK, map[string]string{"status": "healthy"})
}

func (h *ReportHandler) Ready(w http.ResponseWriter, r *http.Request) {
	h.jsonResponse(w, http.StatusOK, map[string]string{"status": "ready"})
}

func (h *ReportHandler) jsonResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func (h *ReportHandler) errorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

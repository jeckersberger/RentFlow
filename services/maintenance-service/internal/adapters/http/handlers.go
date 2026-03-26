package http

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/maintenance-service/internal/application"
	"github.com/jeckersberger/rentflow/services/maintenance-service/internal/domain"
)

type Handler struct {
	recordSvc   *application.MaintenanceRecordService
	scheduleSvc *application.MaintenanceScheduleService
	planSvc     *application.MaintenancePlanService
	taskSvc     *application.MaintenanceTaskService
	checklistSvc *application.ChecklistService
	testSvc     *application.ElectricalTestService
	logger      logger.Logger
}

func NewHandler(
	recordSvc *application.MaintenanceRecordService,
	scheduleSvc *application.MaintenanceScheduleService,
	planSvc *application.MaintenancePlanService,
	taskSvc *application.MaintenanceTaskService,
	checklistSvc *application.ChecklistService,
	testSvc *application.ElectricalTestService,
	log logger.Logger,
) *Handler {
	return &Handler{
		recordSvc:   recordSvc,
		scheduleSvc: scheduleSvc,
		planSvc:     planSvc,
		taskSvc:     taskSvc,
		checklistSvc: checklistSvc,
		testSvc:     testSvc,
		logger:      log,
	}
}

func (h *Handler) CreateRecord(w http.ResponseWriter, r *http.Request) {
	var cmd application.CreateMaintenanceRecordCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}
	cmd.TenantID = tenantID

	dto, err := h.recordSvc.CreateRecord(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusCreated, dto)
}

func (h *Handler) GetRecord(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	dto, err := h.recordSvc.GetRecord(r.Context(), tenantID, id)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, dto)
}

func (h *Handler) ListRecords(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	dtos, err := h.recordSvc.ListRecords(r.Context(), tenantID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{"data": dtos})
}

func (h *Handler) UpdateRecord(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	dto, err := h.recordSvc.GetRecord(r.Context(), tenantID, id)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, dto)
}

func (h *Handler) CompleteRecord(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var cmd application.CompleteMaintenanceCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}
	cmd.TenantID = tenantID
	cmd.RecordID = id
	if cmd.CompletedDate.IsZero() {
		cmd.CompletedDate = time.Now()
	}

	dto, err := h.recordSvc.CompleteRecord(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, dto)
}

func (h *Handler) DeleteRecord(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	if err := h.recordSvc.DeleteRecord(r.Context(), tenantID, id); err != nil {
		h.handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) CreateSchedule(w http.ResponseWriter, r *http.Request) {
	var cmd application.CreateScheduleCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}
	cmd.TenantID = tenantID

	dto, err := h.scheduleSvc.CreateSchedule(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusCreated, dto)
}

func (h *Handler) ListSchedules(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	dtos, err := h.scheduleSvc.ListSchedules(r.Context(), tenantID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{"data": dtos})
}

func (h *Handler) GetScheduleByEquipment(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	dto, err := h.scheduleSvc.GetScheduleByEquipment(r.Context(), tenantID, id)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, dto)
}

func (h *Handler) ListOverdue(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	dtos, err := h.recordSvc.ListOverdue(r.Context(), tenantID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{"data": dtos})
}

func (h *Handler) ImportDGUV(w http.ResponseWriter, r *http.Request) {
	h.respondJSON(w, http.StatusOK, map[string]string{"status": "import_queued"})
}

func (h *Handler) GetDGUVStatus(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	dto, err := h.recordSvc.ListByEquipment(r.Context(), tenantID, id)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{"records": dto})
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
	if err == domain.ErrRecordNotFound || err == domain.ErrScheduleNotFound ||
		err == domain.ErrPlanNotFound || err == domain.ErrTaskNotFound ||
		err == domain.ErrChecklistNotFound || err == domain.ErrTestNotFound {
		h.respondError(w, http.StatusNotFound, err.Error())
		return
	}
	if err == domain.ErrInvalidInput || err == domain.ErrTenantIDRequired ||
		err == domain.ErrInvalidTaskStatus {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.logger.Error("Unhandled error", err)
	h.respondError(w, http.StatusInternalServerError, "internal server error")
}

// Maintenance Plans
func (h *Handler) CreatePlan(w http.ResponseWriter, r *http.Request) {
	var cmd application.CreateMaintenancePlanCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}
	cmd.TenantID = tenantID

	dto, err := h.planSvc.CreatePlan(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusCreated, dto)
}

func (h *Handler) GetPlan(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	dto, err := h.planSvc.GetPlan(r.Context(), tenantID, id)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, dto)
}

func (h *Handler) ListPlans(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	dtos, err := h.planSvc.ListPlans(r.Context(), tenantID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{"data": dtos})
}

func (h *Handler) UpdatePlan(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var cmd application.CreateMaintenancePlanCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	dto, err := h.planSvc.UpdatePlan(r.Context(), tenantID, id, cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, dto)
}

func (h *Handler) DeletePlan(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	if err := h.planSvc.DeletePlan(r.Context(), tenantID, id); err != nil {
		h.handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Maintenance Tasks
func (h *Handler) CreateTask(w http.ResponseWriter, r *http.Request) {
	var cmd application.CreateMaintenanceTaskCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}
	cmd.TenantID = tenantID

	dto, err := h.taskSvc.CreateTask(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusCreated, dto)
}

func (h *Handler) GetTask(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	dto, err := h.taskSvc.GetTask(r.Context(), tenantID, id)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, dto)
}

func (h *Handler) ListTasks(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	dtos, err := h.taskSvc.ListTasks(r.Context(), tenantID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{"data": dtos})
}

func (h *Handler) StartTask(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	cmd := application.StartMaintenanceTaskCommand{
		TenantID: tenantID,
		TaskID:   id,
	}

	dto, err := h.taskSvc.StartTask(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, dto)
}

func (h *Handler) CompleteTask(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var cmd application.CompleteMaintenanceTaskCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}
	cmd.TenantID = tenantID
	cmd.TaskID = id

	dto, err := h.taskSvc.CompleteTask(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, dto)
}

func (h *Handler) GetDueTasks(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	dtos, err := h.taskSvc.GetDueTasks(r.Context(), tenantID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{"data": dtos})
}

func (h *Handler) GetOverdueTasks(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	dtos, err := h.taskSvc.GetOverdueTasks(r.Context(), tenantID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{"data": dtos})
}

// Checklists
func (h *Handler) CreateChecklist(w http.ResponseWriter, r *http.Request) {
	var cmd application.CreateChecklistCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}
	cmd.TenantID = tenantID

	dto, err := h.checklistSvc.CreateChecklist(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusCreated, dto)
}

func (h *Handler) GetChecklist(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	dto, err := h.checklistSvc.GetChecklist(r.Context(), tenantID, id)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, dto)
}

func (h *Handler) ListChecklists(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	dtos, err := h.checklistSvc.ListChecklists(r.Context(), tenantID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{"data": dtos})
}

// Electrical Tests
func (h *Handler) RecordTest(w http.ResponseWriter, r *http.Request) {
	var cmd application.RecordElectricalTestCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}
	cmd.TenantID = tenantID

	dto, err := h.testSvc.RecordTest(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusCreated, dto)
}

func (h *Handler) GetTestsByEquipment(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	// Support both path parameter and query parameter for equipment_id
	equipmentID := r.PathValue("id")
	if equipmentID == "" {
		equipmentID = r.URL.Query().Get("equipment_id")
	}

	// If no equipment ID provided, list all tests for the tenant
	if equipmentID == "" {
		dtos, err := h.testSvc.ListTestsByTenant(r.Context(), tenantID)
		if err != nil {
			h.handleError(w, err)
			return
		}
		h.respondJSON(w, http.StatusOK, map[string]interface{}{"data": dtos})
		return
	}

	dtos, err := h.testSvc.ListTestsByEquipment(r.Context(), tenantID, equipmentID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{"data": dtos})
}

func (h *Handler) ImportIzytron(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	dtos, err := h.testSvc.ImportIzytronXML(r.Context(), tenantID, string(body))
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusCreated, map[string]interface{}{"data": dtos})
}

func (h *Handler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	dueTasks, err := h.taskSvc.GetDueTasks(r.Context(), tenantID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	overdueTasks, err := h.taskSvc.GetOverdueTasks(r.Context(), tenantID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	duePlans, err := h.planSvc.GetDuePlans(r.Context(), tenantID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	recentTests, err := h.testSvc.ListTestsByTenant(r.Context(), tenantID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	dashboard := application.MaintenanceDashboardDTO{
		DueTasks:     dueTasks,
		OverdueTasks: overdueTasks,
		DuePlans:     duePlans,
		RecentTests:  recentTests,
	}

	h.respondJSON(w, http.StatusOK, dashboard)
}

// ImportECheck handles CSV file upload from Gossen Metrawatt IZYTRON.IQ
// Expected CSV columns: Prüfling, Prüfdatum, Prüfer, Ergebnis, Norm, Nächste Prüfung
func (h *Handler) ImportECheck(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	// Parse multipart form (max 10 MB)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid multipart form")
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "file field required")
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = ';' // IZYTRON.IQ uses semicolon separator
	reader.LazyQuotes = true

	records, err := reader.ReadAll()
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "failed to parse CSV: "+err.Error())
		return
	}

	if len(records) < 2 {
		h.respondError(w, http.StatusBadRequest, "CSV must contain a header row and at least one data row")
		return
	}

	// Map header columns to indices
	header := records[0]
	colIdx := make(map[string]int)
	for i, col := range header {
		normalized := strings.TrimSpace(strings.ToLower(col))
		// Remove BOM if present
		normalized = strings.TrimPrefix(normalized, "\ufeff")
		colIdx[normalized] = i
	}

	// Find columns by known names (German IZYTRON.IQ export)
	getCol := func(row []string, names ...string) string {
		for _, name := range names {
			if idx, ok := colIdx[name]; ok && idx < len(row) {
				return strings.TrimSpace(row[idx])
			}
		}
		return ""
	}

	var imported []*application.ElectricalTestDTO
	var errors []string

	for i, row := range records[1:] {
		lineNum := i + 2

		equipmentName := getCol(row, "prüfling", "pruefling", "equipment", "barcode", "geräte-id", "geraete-id")
		testDateStr := getCol(row, "prüfdatum", "pruefdatum", "test_date", "datum")
		tester := getCol(row, "prüfer", "pruefer", "tester")
		result := getCol(row, "ergebnis", "result", "bewertung")
		norm := getCol(row, "norm", "prüfnorm", "pruefnorm", "test_type")
		nextDateStr := getCol(row, "nächste prüfung", "naechste pruefung", "next_test", "nächster prüftermin")

		if equipmentName == "" {
			errors = append(errors, fmt.Sprintf("Zeile %d: Prüfling fehlt", lineNum))
			continue
		}

		// Parse test date
		testDate := time.Now()
		for _, layout := range []string{"02.01.2006", "2006-01-02", "02/01/2006", "2.1.2006"} {
			if t, err := time.Parse(layout, testDateStr); err == nil {
				testDate = t
				break
			}
		}

		// Note: nextDateStr from CSV is parsed but RecordTest calculates next_test_date as +1 year
		_ = nextDateStr

		// Map result to domain
		testResult := "passed"
		resultLower := strings.ToLower(result)
		if strings.Contains(resultLower, "nicht bestanden") || strings.Contains(resultLower, "failed") || strings.Contains(resultLower, "n.i.o") {
			testResult = "failed"
		} else if strings.Contains(resultLower, "bedingt") || strings.Contains(resultLower, "conditional") {
			testResult = "conditional"
		}

		// Map norm to test type
		testType := "vde_0701"
		normLower := strings.ToLower(norm)
		if strings.Contains(normLower, "0702") {
			testType = "vde_0702"
		}
		if strings.Contains(normLower, "dguv") {
			testType = "vde_0701" // DGUV V3 maps to VDE 0701/0702
		}

		if tester == "" {
			tester = "IZYTRON Import"
		}

		cmd := application.RecordElectricalTestCommand{
			TenantID:           tenantID,
			EquipmentID:        equipmentName,
			TesterID:           tester,
			TestType:           testType,
			TestDate:           testDate,
			Result:             testResult,
			VisualInspectionOK: testResult != "failed",
			FunctionalTestOK:   testResult != "failed",
			TestDeviceID:       "izytron-csv-import",
			TestDeviceName:     "IZYTRON.IQ CSV Import",
			CertificateNumber:  fmt.Sprintf("ECHECK-CSV-%s-%s-%d", equipmentName, testDate.Format("20060102"), lineNum),
			Notes:              fmt.Sprintf("Importiert aus IZYTRON.IQ CSV (Norm: %s)", norm),
		}

		dto, err := h.testSvc.RecordTest(r.Context(), cmd)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Zeile %d (%s): %s", lineNum, equipmentName, err.Error()))
			continue
		}

		imported = append(imported, dto)
	}

	response := map[string]interface{}{
		"imported_count": len(imported),
		"total_rows":     len(records) - 1,
		"data":           imported,
	}
	if len(errors) > 0 {
		response["errors"] = errors
	}

	h.respondJSON(w, http.StatusCreated, response)
}

package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/invoice-service/internal/application"
	"github.com/jeckersberger/rentflow/services/invoice-service/internal/domain"
)

type Handler struct {
	invoiceSvc *application.InvoiceService
	quoteSvc   *application.QuoteService
	duningSvc  *application.DunningService
	exportSvc  *application.ExportService
	logger     logger.Logger
}

func NewHandler(
	invoiceSvc *application.InvoiceService,
	quoteSvc *application.QuoteService,
	duningSvc *application.DunningService,
	exportSvc *application.ExportService,
	logger logger.Logger,
) *Handler {
	return &Handler{
		invoiceSvc: invoiceSvc,
		quoteSvc:   quoteSvc,
		duningSvc:  duningSvc,
		exportSvc:  exportSvc,
		logger:     logger,
	}
}

// Invoice Handlers

func (h *Handler) CreateInvoice(w http.ResponseWriter, r *http.Request) {
	var cmd application.CreateInvoiceCommand
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

	dto, err := h.invoiceSvc.CreateInvoice(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusCreated, dto)
}

func (h *Handler) GetInvoice(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	dto, err := h.invoiceSvc.GetInvoice(r.Context(), tenantID, id)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, dto)
}

func (h *Handler) ListInvoices(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	limit := 20
	offset := 0
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	status := r.URL.Query().Get("status")
	clientName := r.URL.Query().Get("client_name")

	query := application.ListInvoicesQuery{
		TenantID:   tenantID,
		Status:     getStringPtr(status),
		ClientName: getStringPtr(clientName),
		Limit:      limit,
		Offset:     offset,
	}

	result, err := h.invoiceSvc.ListInvoices(r.Context(), query)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, result)
}

func (h *Handler) SendInvoice(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	var payload struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cmd := application.SendInvoiceCommand{
		ID:       id,
		TenantID: tenantID,
		Email:    payload.Email,
	}

	// SendInvoiceWithPDF generiert PDF, sendet E-Mail und ändert den Status
	dto, err := h.invoiceSvc.SendInvoiceWithPDF(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, dto)
}

func (h *Handler) MarkInvoicePaid(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	var payload struct {
		PaymentMethod string `json:"payment_method"`
		PaymentRef    string `json:"payment_ref"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cmd := application.MarkInvoicePaidCommand{
		ID:            id,
		TenantID:      tenantID,
		PaymentMethod: payload.PaymentMethod,
		PaymentRef:    payload.PaymentRef,
	}

	dto, err := h.invoiceSvc.MarkInvoicePaid(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, dto)
}

func (h *Handler) CancelInvoice(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	var payload struct {
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cmd := application.CancelInvoiceCommand{
		ID:       id,
		TenantID: tenantID,
		Reason:   payload.Reason,
	}

	dto, err := h.invoiceSvc.CancelInvoice(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, dto)
}

func (h *Handler) CreditInvoice(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	var payload struct {
		CreditID string `json:"credit_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cmd := application.CreditInvoiceCommand{
		ID:       id,
		TenantID: tenantID,
		CreditID: payload.CreditID,
	}

	dto, err := h.invoiceSvc.CreditInvoice(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, dto)
}

// Quote Handlers

func (h *Handler) CreateQuote(w http.ResponseWriter, r *http.Request) {
	var cmd application.CreateQuoteCommand
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

	dto, err := h.quoteSvc.CreateQuote(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusCreated, dto)
}

func (h *Handler) GetQuote(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	dto, err := h.quoteSvc.GetQuote(r.Context(), tenantID, id)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, dto)
}

func (h *Handler) ListQuotes(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	limit := 20
	offset := 0
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	status := r.URL.Query().Get("status")
	clientName := r.URL.Query().Get("client_name")

	query := application.ListQuotesQuery{
		TenantID:   tenantID,
		Status:     getStringPtr(status),
		ClientName: getStringPtr(clientName),
		Limit:      limit,
		Offset:     offset,
	}

	result, err := h.quoteSvc.ListQuotes(r.Context(), query)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, result)
}

func (h *Handler) SendQuote(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	var payload struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cmd := application.SendQuoteCommand{
		ID:       id,
		TenantID: tenantID,
		Email:    payload.Email,
	}

	dto, err := h.quoteSvc.SendQuote(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, dto)
}

func (h *Handler) AcceptQuote(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	cmd := application.AcceptQuoteCommand{
		ID:       id,
		TenantID: tenantID,
	}

	dto, err := h.quoteSvc.AcceptQuote(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, dto)
}

func (h *Handler) RejectQuote(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	var payload struct {
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cmd := application.RejectQuoteCommand{
		ID:       id,
		TenantID: tenantID,
		Reason:   payload.Reason,
	}

	dto, err := h.quoteSvc.RejectQuote(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, dto)
}

func (h *Handler) ConvertQuoteToInvoice(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	var payload struct {
		IssueDate     string `json:"issue_date"`
		DueDate       string `json:"due_date"`
		InternalNotes string `json:"internal_notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Parse dates
	issueDate, _ := parseDate(payload.IssueDate)
	dueDate, _ := parseDate(payload.DueDate)

	cmd := application.ConvertQuoteToInvoiceCommand{
		QuoteID:       id,
		TenantID:      tenantID,
		IssueDate:     issueDate,
		DueDate:       dueDate,
		InternalNotes: payload.InternalNotes,
	}

	dto, err := h.quoteSvc.ConvertQuoteToInvoice(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusCreated, dto)
}

// Dunning Handlers

func (h *Handler) GetOverduePotential(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	invoices, err := h.duningSvc.GetOverdueInvoices(r.Context(), tenantID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"overdue_invoices": invoices,
		"count":            len(invoices),
	})
}

func (h *Handler) CreateDunningReminder(w http.ResponseWriter, r *http.Request) {
	invoiceID := r.PathValue("invoiceId")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	var payload struct {
		Level     int      `json:"level"`
		DaysToAdd int      `json:"days_to_add"`
		Fee       *float64 `json:"fee,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cmd := application.CreateDunningCommand{
		TenantID:  tenantID,
		InvoiceID: invoiceID,
		Level:     payload.Level,
		DaysToAdd: payload.DaysToAdd,
		CustomFee: payload.Fee,
	}

	dto, err := h.duningSvc.CreateReminder(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusCreated, dto)
}

func (h *Handler) SendDunningReminder(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	var payload struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cmd := application.SendDunningCommand{
		ID:       id,
		TenantID: tenantID,
		Email:    payload.Email,
	}

	dto, err := h.duningSvc.SendReminder(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, dto)
}

// Export Handlers

func (h *Handler) ExportDATEV(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	fromDate := r.URL.Query().Get("from")
	toDate := r.URL.Query().Get("to")
	if fromDate == "" || toDate == "" {
		h.respondError(w, http.StatusBadRequest, "from and to dates required")
		return
	}

	query := application.ExportDATEVQuery{
		TenantID: tenantID,
		FromDate: fromDate,
		ToDate:   toDate,
	}

	rows, err := h.exportSvc.ExportDATEV(r.Context(), query)
	if err != nil {
		h.handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=datev_export.csv")

	// Write CSV header
	w.Write([]byte("Umsatz,SollHaben,WKZ,Konto,Gegenkonto,Belegdatum,Belegnummer,Buchungstext\n"))

	// Write rows
	for _, row := range rows {
		line := formatDATEVRow(row)
		w.Write([]byte(line + "\n"))
	}
}

func (h *Handler) ExportCSV(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	fromDate := r.URL.Query().Get("from")
	toDate := r.URL.Query().Get("to")
	if fromDate == "" || toDate == "" {
		h.respondError(w, http.StatusBadRequest, "from and to dates required")
		return
	}

	query := application.ExportCSVQuery{
		TenantID: tenantID,
		FromDate: fromDate,
		ToDate:   toDate,
	}

	csv, err := h.exportSvc.ExportCSV(r.Context(), query)
	if err != nil {
		h.handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=invoices_export.csv")
	w.Write([]byte(csv))
}

// Helper methods

func (h *Handler) respondJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) respondError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func (h *Handler) handleError(w http.ResponseWriter, err error) {
	if domainErr, ok := err.(*domain.DomainError); ok {
		switch domainErr.Code {
		case "NOT_FOUND":
			h.respondError(w, http.StatusNotFound, domainErr.Message)
		case "TENANT_REQUIRED", "INVALID_INPUT", "NO_ITEMS", "INVALID_TAX_RATE":
			h.respondError(w, http.StatusBadRequest, domainErr.Message)
		case "INVALID_TRANSITION", "INVALID_STATUS", "CANNOT_MODIFY":
			h.respondError(w, http.StatusConflict, domainErr.Message)
		case "UNAUTHORIZED":
			h.respondError(w, http.StatusUnauthorized, domainErr.Message)
		case "INTEGRITY_CHECK":
			h.respondError(w, http.StatusConflict, domainErr.Message)
		default:
			h.respondError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	h.respondError(w, http.StatusInternalServerError, "internal server error")
}

func getStringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func parseDate(dateStr string) (time.Time, error) {
	if dateStr == "" {
		return time.Now(), nil
	}
	return time.Parse("2006-01-02", dateStr)
}

func formatDATEVRow(row *application.DATEVExportRow) string {
	return fmt.Sprintf(
		"%.2f,%s,%s,%s,%s,%s,%s,%s",
		row.Umsatz, row.SollHaben, row.WKZUmsatz, row.Konto,
		row.Gegenkonto, row.Belegdatum, row.Belegnummer, row.Buchungstext,
	)
}

// New Invoice Service Methods

func (h *Handler) GetOpenInvoices(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	summary, err := h.invoiceSvc.GetOpenInvoices(r.Context(), tenantID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, summary)
}

func (h *Handler) GetInvoicePDF(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	// Prüfen ob HTML-Format explizit angefragt wird (Fallback für Browser-Vorschau)
	if r.URL.Query().Get("format") == "html" {
		html, err := h.invoiceSvc.GenerateInvoiceHTML(r.Context(), tenantID, id)
		if err != nil {
			h.handleError(w, err)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(html))
		return
	}

	// Standard: PDF generieren
	pdfBytes, err := h.invoiceSvc.GenerateInvoicePDF(r.Context(), tenantID, id)
	if err != nil {
		h.handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`inline; filename="invoice_%s.pdf"`, id))
	w.Header().Set("Content-Length", strconv.Itoa(len(pdfBytes)))
	w.WriteHeader(http.StatusOK)
	w.Write(pdfBytes)
}

func (h *Handler) GetQuotePDF(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	// Prüfen ob HTML-Format explizit angefragt wird
	if r.URL.Query().Get("format") == "html" {
		html, err := h.quoteSvc.GenerateQuoteHTML(r.Context(), tenantID, id)
		if err != nil {
			h.handleError(w, err)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(html))
		return
	}

	// Standard: PDF generieren
	pdfBytes, err := h.quoteSvc.GenerateQuotePDF(r.Context(), tenantID, id)
	if err != nil {
		h.handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`inline; filename="quote_%s.pdf"`, id))
	w.Header().Set("Content-Length", strconv.Itoa(len(pdfBytes)))
	w.WriteHeader(http.StatusOK)
	w.Write(pdfBytes)
}

func (h *Handler) CreateInvoiceFromProject(w http.ResponseWriter, r *http.Request) {
	projectID := r.PathValue("projectId")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	var payload struct {
		ClientName string `json:"client_name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	dto, err := h.invoiceSvc.CreateInvoiceFromProject(r.Context(), tenantID, projectID, payload.ClientName)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusCreated, dto)
}

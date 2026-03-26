package http

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/invoice-service/internal/application"
	"github.com/jeckersberger/rentflow/services/invoice-service/internal/domain"
)

type Handler struct {
	invoiceSvc    *application.InvoiceService
	quoteSvc      *application.QuoteService
	creditNoteSvc *application.CreditNoteService
	duningSvc     *application.DunningService
	exportSvc     *application.ExportService
	logger        logger.Logger
}

func NewHandler(
	invoiceSvc *application.InvoiceService,
	quoteSvc *application.QuoteService,
	creditNoteSvc *application.CreditNoteService,
	duningSvc *application.DunningService,
	exportSvc *application.ExportService,
	logger logger.Logger,
) *Handler {
	return &Handler{
		invoiceSvc:    invoiceSvc,
		quoteSvc:      quoteSvc,
		creditNoteSvc: creditNoteSvc,
		duningSvc:     duningSvc,
		exportSvc:     exportSvc,
		logger:        logger,
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

	// Basic input validation
	if strings.TrimSpace(cmd.ClientName) == "" {
		h.respondError(w, http.StatusBadRequest, "client name is required")
		return
	}

	dto, err := h.invoiceSvc.CreateInvoice(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusCreated, dto)
}

func (h *Handler) UpdateInvoice(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
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

	if strings.TrimSpace(cmd.ClientName) == "" {
		h.respondError(w, http.StatusBadRequest, "client name is required")
		return
	}

	// Get existing invoice to preserve invoice number and GoBD hash chain
	existing, err := h.invoiceSvc.GetInvoice(r.Context(), tenantID, id)
	if err != nil {
		h.handleError(w, err)
		return
	}

	// Only allow updates on draft invoices
	if existing.Status != "draft" {
		h.respondError(w, http.StatusConflict, "can only update draft invoices")
		return
	}

	// Update via service
	updateCmd := application.UpdateInvoiceCommand{
		ID:            id,
		TenantID:      tenantID,
		ClientName:    cmd.ClientName,
		ClientAddress: cmd.ClientAddress,
		ClientEmail:   cmd.ClientEmail,
		ClientTaxID:   cmd.ClientTaxID,
		TaxRate:       cmd.TaxRate,
		DueDate:       cmd.DueDate,
		Notes:         cmd.Notes,
		InternalNotes: cmd.InternalNotes,
	}

	dto, err := h.invoiceSvc.UpdateInvoice(r.Context(), updateCmd)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, dto)
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

func (h *Handler) RecordPayment(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	var payload struct {
		Amount float64 `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cmd := application.RecordPaymentCommand{
		ID:       id,
		TenantID: tenantID,
		Amount:   payload.Amount,
	}

	dto, err := h.invoiceSvc.RecordPayment(r.Context(), cmd)
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

func (h *Handler) ConfirmQuote(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	cmd := application.ConfirmQuoteCommand{
		ID:       id,
		TenantID: tenantID,
	}

	dto, err := h.quoteSvc.ConfirmQuote(r.Context(), cmd)
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

// Credit Note Handlers

func (h *Handler) CreateCreditNote(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	var payload struct {
		Reason   string `json:"reason"`
		Items    []struct {
			Description string  `json:"description"`
			Quantity    float64 `json:"quantity"`
			Unit        string  `json:"unit"`
			UnitPrice   float64 `json:"unit_price"`
		} `json:"items"`
		TaxRate  float64 `json:"tax_rate"`
		Currency string  `json:"currency"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	items := make([]application.CreateCreditNoteItemCommand, len(payload.Items))
	for i, item := range payload.Items {
		items[i] = application.CreateCreditNoteItemCommand{
			Description: item.Description,
			Quantity:    item.Quantity,
			Unit:        item.Unit,
			UnitPrice:   item.UnitPrice,
		}
	}

	cmd := application.CreateCreditNoteCommand{
		InvoiceID: id,
		TenantID:  tenantID,
		Reason:    payload.Reason,
		Items:     items,
		TaxRate:   payload.TaxRate,
		Currency:  payload.Currency,
	}

	dto, err := h.creditNoteSvc.CreateCreditNote(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusCreated, dto)
}

func (h *Handler) GetCreditNote(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	dto, err := h.creditNoteSvc.GetCreditNote(r.Context(), tenantID, id)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, dto)
}

func (h *Handler) IssueCreditNote(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	dto, err := h.creditNoteSvc.IssueCreditNote(r.Context(), tenantID, id)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, dto)
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

// Invoice-scoped Dunning Handlers

func (h *Handler) CreateInvoiceDunning(w http.ResponseWriter, r *http.Request) {
	invoiceID := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	var payload struct {
		Level   int      `json:"level"`
		Fee     *float64 `json:"fee,omitempty"`
		Message string   `json:"message,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Map level to grace period days
	daysToAdd := 7
	switch payload.Level {
	case 1:
		daysToAdd = 7  // Zahlungserinnerung: 7 days grace
	case 2:
		daysToAdd = 14 // 1. Mahnung: 14 days
	case 3:
		daysToAdd = 7  // 2. Mahnung: 7 days
	}

	cmd := application.CreateDunningCommand{
		TenantID:  tenantID,
		InvoiceID: invoiceID,
		Level:     payload.Level,
		DaysToAdd: daysToAdd,
		CustomFee: payload.Fee,
	}

	dto, err := h.duningSvc.CreateReminder(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}

	// Store custom message as notes if provided
	if payload.Message != "" {
		dto.Notes = payload.Message
	}

	h.respondJSON(w, http.StatusCreated, dto)
}

func (h *Handler) GetInvoiceDunningHistory(w http.ResponseWriter, r *http.Request) {
	invoiceID := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	dtos, err := h.duningSvc.GetDunningHistory(r.Context(), tenantID, invoiceID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, dtos)
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

// Payment Matching types

type BankTransaction struct {
	Amount    float64 `json:"amount"`
	Reference string  `json:"reference"`
	Date      string  `json:"date"`
	Payer     string  `json:"payer"`
}

type PaymentMatchResult struct {
	Transaction BankTransaction      `json:"transaction"`
	Confidence  string               `json:"confidence"` // "exact", "probable", "no_match"
	Invoice     *application.InvoiceDTO `json:"invoice,omitempty"`
	MatchReason string               `json:"match_reason,omitempty"`
}

type BulkImportResult struct {
	Imported     int                  `json:"imported"`
	Transactions []BankTransaction    `json:"transactions"`
}

// MatchPayment tries to match a bank transaction to an open invoice
func (h *Handler) MatchPayment(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	var tx BankTransaction
	if err := json.NewDecoder(r.Body).Decode(&tx); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Get all open invoices (sent, overdue, partially_paid)
	openSummary, err := h.invoiceSvc.GetOpenInvoices(r.Context(), tenantID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	result := h.matchTransaction(tx, openSummary.Invoices)
	h.respondJSON(w, http.StatusOK, result)
}

// ImportPaymentsCSV imports bank transactions from CSV
func (h *Handler) ImportPaymentsCSV(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	// Parse multipart form (max 10MB)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid form data")
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "file required")
		return
	}
	defer file.Close()

	// Read all file content so we can retry with different delimiters
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "failed to read file")
		return
	}

	// Try semicolon first (German standard), then comma
	transactions := parseCSVTransactions(string(fileBytes), ';')
	if len(transactions) == 0 {
		transactions = parseCSVTransactions(string(fileBytes), ',')
	}

	_ = tenantID

	h.respondJSON(w, http.StatusOK, BulkImportResult{
		Imported:     len(transactions),
		Transactions: transactions,
	})
}

// matchTransaction tries to match a single transaction to open invoices
func (h *Handler) matchTransaction(tx BankTransaction, openInvoices []*application.InvoiceDTO) *PaymentMatchResult {
	result := &PaymentMatchResult{
		Transaction: tx,
		Confidence:  "no_match",
	}

	if len(openInvoices) == 0 {
		return result
	}

	refUpper := strings.ToUpper(tx.Reference)
	payerUpper := strings.ToUpper(tx.Payer)

	// Strategy 1: Exact invoice number in reference text
	for _, inv := range openInvoices {
		invNumUpper := strings.ToUpper(inv.InvoiceNumber)
		if invNumUpper != "" && strings.Contains(refUpper, invNumUpper) {
			result.Confidence = "exact"
			result.Invoice = inv
			result.MatchReason = fmt.Sprintf("Rechnungsnummer %s im Verwendungszweck gefunden", inv.InvoiceNumber)
			return result
		}
	}

	// Strategy 2: Exact amount match with open invoices
	var amountMatches []*application.InvoiceDTO
	for _, inv := range openInvoices {
		matchAmount := inv.RemainingAmount
		if matchAmount == 0 {
			matchAmount = inv.Total
		}
		if math.Abs(tx.Amount-matchAmount) < 0.01 {
			amountMatches = append(amountMatches, inv)
		}
	}

	if len(amountMatches) == 1 {
		result.Confidence = "probable"
		result.Invoice = amountMatches[0]
		result.MatchReason = fmt.Sprintf("Betrag %.2f EUR stimmt mit Rechnung %s ueberein", tx.Amount, amountMatches[0].InvoiceNumber)
		return result
	}

	// Strategy 3: Client name match
	for _, inv := range openInvoices {
		clientUpper := strings.ToUpper(inv.ClientName)
		if clientUpper != "" && (strings.Contains(payerUpper, clientUpper) || strings.Contains(clientUpper, payerUpper)) {
			// If we also have an amount match within the name matches, prefer that
			for _, amtMatch := range amountMatches {
				if amtMatch.ID == inv.ID {
					result.Confidence = "exact"
					result.Invoice = inv
					result.MatchReason = fmt.Sprintf("Kundenname '%s' und Betrag %.2f EUR stimmen ueberein", inv.ClientName, tx.Amount)
					return result
				}
			}
			result.Confidence = "probable"
			result.Invoice = inv
			result.MatchReason = fmt.Sprintf("Kundenname '%s' im Auftraggeber gefunden", inv.ClientName)
			return result
		}
	}

	// If multiple amount matches but no name match, return first as probable
	if len(amountMatches) > 1 {
		result.Confidence = "probable"
		result.Invoice = amountMatches[0]
		result.MatchReason = fmt.Sprintf("Betrag %.2f EUR stimmt mit %d Rechnungen ueberein", tx.Amount, len(amountMatches))
		return result
	}

	return result
}

// parseGermanAmount parses amounts like "1.234,56" or "1234.56"
func parseGermanAmount(s string) (float64, error) {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, " ", "")
	s = strings.TrimSuffix(s, "EUR")
	s = strings.TrimSuffix(s, "€")
	s = strings.TrimSpace(s)

	// German format: 1.234,56 -> convert to 1234.56
	if strings.Contains(s, ",") {
		s = strings.ReplaceAll(s, ".", "")
		s = strings.Replace(s, ",", ".", 1)
	}

	return strconv.ParseFloat(s, 64)
}

// parseCSVTransactions parses bank transactions from CSV content with a given delimiter
func parseCSVTransactions(content string, delimiter rune) []BankTransaction {
	reader := csv.NewReader(strings.NewReader(content))
	reader.Comma = delimiter
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1 // allow variable number of fields

	var transactions []BankTransaction
	lineNum := 0

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}
		lineNum++

		// Skip header row
		if lineNum == 1 {
			lower := strings.ToLower(strings.Join(record, " "))
			if strings.Contains(lower, "datum") || strings.Contains(lower, "date") ||
				strings.Contains(lower, "betrag") || strings.Contains(lower, "amount") {
				continue
			}
		}

		// Expect at least 4 columns: Date, Amount, Reference, Payer
		if len(record) < 4 {
			continue
		}

		amount, err := parseGermanAmount(strings.TrimSpace(record[1]))
		if err != nil {
			continue
		}

		tx := BankTransaction{
			Date:      strings.TrimSpace(record[0]),
			Amount:    amount,
			Reference: strings.TrimSpace(record[2]),
			Payer:     strings.TrimSpace(record[3]),
		}

		transactions = append(transactions, tx)
	}

	return transactions
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

func (h *Handler) GenerateDeliveryNotePDF(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	var data application.DeliveryNoteData
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if data.DeliveryNoteNumber == "" {
		h.respondError(w, http.StatusBadRequest, "delivery note number is required")
		return
	}

	// Check if HTML format is requested
	if r.URL.Query().Get("format") == "html" {
		html, err := h.invoiceSvc.GenerateDeliveryNoteHTML(r.Context(), &data)
		if err != nil {
			h.respondError(w, http.StatusInternalServerError, "failed to generate delivery note: "+err.Error())
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(html))
		return
	}

	// Default: generate PDF
	pdfBytes, err := h.invoiceSvc.GenerateDeliveryNotePDF(r.Context(), &data)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "failed to generate delivery note PDF: "+err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`inline; filename="Lieferschein_%s.pdf"`, data.DeliveryNoteNumber))
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

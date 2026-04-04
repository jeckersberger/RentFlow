package httphandler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/pkg/common/errors"
	"github.com/jeckersberger/EquipFlow/pkg/common/middleware"
	"github.com/jeckersberger/EquipFlow/pkg/common/pagination"
	"github.com/jeckersberger/EquipFlow/pkg/common/response"
	"github.com/jeckersberger/EquipFlow/services/invoice/internal/application"
	"github.com/jeckersberger/EquipFlow/services/invoice/internal/domain"
)

type InvoiceHandler struct {
	invoiceService *application.InvoiceService
	authBaseURL    string
	logger         zerolog.Logger
}

func NewInvoiceHandler(invoiceService *application.InvoiceService, authBaseURL string, logger zerolog.Logger) *InvoiceHandler {
	return &InvoiceHandler{
		invoiceService: invoiceService,
		authBaseURL:    authBaseURL,
		logger:         logger.With().Str("handler", "invoice").Logger(),
	}
}

func (h *InvoiceHandler) List(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	p := pagination.Parse(r)

	status := r.URL.Query().Get("status")
	filter := domain.InvoiceFilter{
		Page:    p.Page,
		PerPage: p.PerPage,
		Status:  status,
	}

	items, total, err := h.invoiceService.List(r.Context(), claims.TenantID, filter)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.SuccessWithMeta(w, items, response.Meta{
		Page:    p.Page,
		PerPage: p.PerPage,
		Total:   total,
	})
}

func (h *InvoiceHandler) Get(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	invoice, err := h.invoiceService.GetByID(r.Context(), id, claims.TenantID)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, invoice)
}

func (h *InvoiceHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	var req application.CreateInvoiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiger Request-Body"))
		return
	}

	created, err := h.invoiceService.Create(r.Context(), claims.TenantID, req)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Created(w, created)
}

func (h *InvoiceHandler) Delete(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	if err := h.invoiceService.Delete(r.Context(), id, claims.TenantID); err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, map[string]string{"message": "Rechnung geloescht"})
}

func (h *InvoiceHandler) Update(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	var req application.UpdateInvoiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiger Request-Body"))
		return
	}

	updated, err := h.invoiceService.Update(r.Context(), id, claims.TenantID, req)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, updated)
}

func (h *InvoiceHandler) Finalize(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	invoice, err := h.invoiceService.Finalize(r.Context(), id, claims.TenantID)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, invoice)
}

func (h *InvoiceHandler) Search(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	query := r.URL.Query().Get("q")
	if query == "" {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Suchbegriff (q) ist erforderlich"))
		return
	}

	p := pagination.Parse(r)

	items, total, err := h.invoiceService.Search(r.Context(), claims.TenantID, query, p.Page, p.PerPage)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.SuccessWithMeta(w, items, response.Meta{
		Page:    p.Page,
		PerPage: p.PerPage,
		Total:   total,
	})
}

func (h *InvoiceHandler) AddItem(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	invoiceID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	var req application.AddItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiger Request-Body"))
		return
	}

	item, err := h.invoiceService.AddItem(r.Context(), invoiceID, claims.TenantID, req)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Created(w, item)
}

func (h *InvoiceHandler) ListItems(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	invoiceID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	items, err := h.invoiceService.ListItems(r.Context(), invoiceID, claims.TenantID)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, items)
}

func (h *InvoiceHandler) RemoveItem(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	_, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	itemID, err := parseUUID(chi.URLParam(r, "itemId"))
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	if err := h.invoiceService.RemoveItem(r.Context(), itemID, claims.TenantID); err != nil {
		errors.HandleError(w, err)
		return
	}

	response.NoContent(w)
}

func (h *InvoiceHandler) AddPayment(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	invoiceID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	var req application.AddPaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiger Request-Body"))
		return
	}

	payment, err := h.invoiceService.AddPayment(r.Context(), invoiceID, claims.TenantID, req)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Created(w, payment)
}

func (h *InvoiceHandler) ListPayments(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	invoiceID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	payments, err := h.invoiceService.ListPayments(r.Context(), invoiceID, claims.TenantID)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, payments)
}

func (h *InvoiceHandler) SendEmail(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	invoice, err := h.invoiceService.GetByID(r.Context(), id, claims.TenantID)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	if invoice.CustomerEmail == "" {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Kunde hat keine E-Mail-Adresse"))
		return
	}

	items, err := h.invoiceService.ListItems(r.Context(), id, claims.TenantID)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	// Generate PDF into buffer
	company := h.fetchCompanyInfo(r.Header.Get("Authorization"))
	data := application.InvoicePDFData{Invoice: invoice, Items: items, Company: company}

	var pdfBuf bytes.Buffer
	if err := application.GenerateInvoicePDF(&pdfBuf, data); err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrInternal, "PDF-Generierung fehlgeschlagen"))
		return
	}

	// Send via notification-service email endpoint
	emailBody, err := application.RenderInvoiceEmailHTML(invoice, company.Name)
	if err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrInternal, "E-Mail-Template fehlgeschlagen"))
		return
	}

	subject := "Rechnung " + invoice.InvoiceNumber + " — " + company.Name
	if invoice.InvoiceType == "credit_note" {
		subject = "Gutschrift " + invoice.InvoiceNumber + " — " + company.Name
	}

	sendErr := h.invoiceService.SendInvoiceEmail(r.Context(), application.InvoiceEmailRequest{
		To:             invoice.CustomerEmail,
		Subject:        subject,
		HTMLBody:       emailBody,
		PDFData:        pdfBuf.Bytes(),
		PDFFilename:    invoice.InvoiceNumber + ".pdf",
		InvoiceID:      id,
		TenantID:       claims.TenantID,
	})
	if sendErr != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrInternal, sendErr.Error()))
		return
	}

	// Update status to "sent" if currently draft/finalized
	if invoice.Status == "draft" || invoice.Status == "finalized" {
		_ = h.invoiceService.UpdateStatus(r.Context(), id, claims.TenantID, "sent")
	}

	response.Success(w, map[string]string{"status": "sent", "to": invoice.CustomerEmail})
}

// CreatePartialInvoice creates a partial invoice from an existing invoice.
func (h *InvoiceHandler) CreatePartialInvoice(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	invoiceID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	var req application.CreatePartialInvoiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiger Request-Body"))
		return
	}

	partial, err := h.invoiceService.CreatePartialInvoice(r.Context(), invoiceID, claims.TenantID, req.Percentage)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Created(w, partial)
}

// fetchCompanyInfo loads tenant data from the auth service and maps it to CompanyInfo.
// Falls back to a minimal default if the auth service is unreachable.
func (h *InvoiceHandler) fetchCompanyInfo(authToken string) application.CompanyInfo {
	if h.authBaseURL == "" {
		h.logger.Warn().Msg("AUTH_BASE_URL not configured, using fallback company info")
		return application.CompanyInfo{Name: "Unbekannt"}
	}

	url := h.authBaseURL + "/api/v1/tenants/current"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to create tenant request")
		return application.CompanyInfo{Name: "Unbekannt"}
	}
	req.Header.Set("Authorization", authToken)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to fetch tenant info from auth service")
		return application.CompanyInfo{Name: "Unbekannt"}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		h.logger.Warn().Int("status", resp.StatusCode).Msg("auth service returned non-200 for tenant info")
		return application.CompanyInfo{Name: "Unbekannt"}
	}

	var envelope struct {
		Data struct {
			Name      string `json:"name"`
			Email     string `json:"email"`
			Phone     string `json:"phone"`
			Street    string `json:"address_street"`
			City      string `json:"address_city"`
			Zip       string `json:"address_zip"`
			TaxNumber string `json:"tax_number"`
			VatID     string `json:"vat_id"`
			IBAN      string `json:"iban"`
			BIC       string `json:"bic"`
			BankName  string `json:"bank_name"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		h.logger.Error().Err(err).Msg("failed to decode tenant info")
		return application.CompanyInfo{Name: "Unbekannt"}
	}

	city := envelope.Data.City
	if envelope.Data.Zip != "" {
		city = envelope.Data.Zip + " " + city
	}

	return application.CompanyInfo{
		Name:      envelope.Data.Name,
		Street:    envelope.Data.Street,
		City:      city,
		Phone:     envelope.Data.Phone,
		Email:     envelope.Data.Email,
		TaxNumber: envelope.Data.TaxNumber,
		IBAN:      envelope.Data.IBAN,
		BIC:       envelope.Data.BIC,
		BankName:  envelope.Data.BankName,
	}
}

func (h *InvoiceHandler) GeneratePDF(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	invoice, err := h.invoiceService.GetByID(r.Context(), id, claims.TenantID)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	items, err := h.invoiceService.ListItems(r.Context(), id, claims.TenantID)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	company := h.fetchCompanyInfo(r.Header.Get("Authorization"))

	data := application.InvoicePDFData{
		Invoice: invoice,
		Items:   items,
		Company: company,
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "inline; filename=\""+invoice.InvoiceNumber+".pdf\"")

	if err := application.GenerateInvoicePDF(w, data); err != nil {
		h.logger.Error().Err(err).Str("invoice_id", id.String()).Msg("PDF generation failed")
		errors.HandleError(w, errors.Wrap(errors.ErrInternal, "PDF-Generierung fehlgeschlagen"))
	}
}

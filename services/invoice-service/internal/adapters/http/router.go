package http

import (
	"net/http"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/invoice-service/internal/application"
)

func NewRouter(
	invoiceSvc *application.InvoiceService,
	quoteSvc *application.QuoteService,
	creditNoteSvc *application.CreditNoteService,
	duningSvc *application.DunningService,
	exportSvc *application.ExportService,
	logger logger.Logger,
) *http.ServeMux {
	router := http.NewServeMux()
	handler := NewHandler(invoiceSvc, quoteSvc, creditNoteSvc, duningSvc, exportSvc, logger)

	// Health & readiness
	router.HandleFunc("GET /health", healthHandler)
	router.HandleFunc("GET /ready", readyHandler)

	// Invoice routes - use actions/ prefix for from-project to avoid conflict with {id}
	router.HandleFunc("POST /api/v1/invoices", handler.CreateInvoice)
	router.HandleFunc("GET /api/v1/invoices", handler.ListInvoices)
	router.HandleFunc("GET /api/v1/invoices/open", handler.GetOpenInvoices)
	router.HandleFunc("POST /api/v1/invoices/actions/from-project/{projectId}", handler.CreateInvoiceFromProject)
	router.HandleFunc("PUT /api/v1/invoices/{id}", handler.UpdateInvoice)
	router.HandleFunc("GET /api/v1/invoices/{id}", handler.GetInvoice)
	router.HandleFunc("GET /api/v1/invoices/{id}/pdf", handler.GetInvoicePDF)
	router.HandleFunc("POST /api/v1/invoices/{id}/send", handler.SendInvoice)
	router.HandleFunc("POST /api/v1/invoices/{id}/mark-paid", handler.MarkInvoicePaid)
	router.HandleFunc("POST /api/v1/invoices/{id}/record-payment", handler.RecordPayment)
	router.HandleFunc("POST /api/v1/invoices/{id}/cancel", handler.CancelInvoice)
	router.HandleFunc("POST /api/v1/invoices/{id}/credit", handler.CreditInvoice)

	// Credit Note routes
	router.HandleFunc("POST /api/v1/invoices/{id}/credit-note", handler.CreateCreditNote)
	router.HandleFunc("GET /api/v1/credit-notes/{id}", handler.GetCreditNote)
	router.HandleFunc("POST /api/v1/credit-notes/{id}/issue", handler.IssueCreditNote)

	// Quote routes
	router.HandleFunc("POST /api/v1/quotes", handler.CreateQuote)
	router.HandleFunc("GET /api/v1/quotes", handler.ListQuotes)
	router.HandleFunc("GET /api/v1/quotes/{id}", handler.GetQuote)
	router.HandleFunc("GET /api/v1/quotes/{id}/pdf", handler.GetQuotePDF)
	router.HandleFunc("POST /api/v1/quotes/{id}/send", handler.SendQuote)
	router.HandleFunc("POST /api/v1/quotes/{id}/accept", handler.AcceptQuote)
	router.HandleFunc("POST /api/v1/quotes/{id}/confirm", handler.ConfirmQuote)
	router.HandleFunc("POST /api/v1/quotes/{id}/reject", handler.RejectQuote)
	router.HandleFunc("POST /api/v1/quotes/{id}/convert-to-invoice", handler.ConvertQuoteToInvoice)

	// Dunning/Payment reminder routes
	router.HandleFunc("GET /api/v1/dunning/overdue", handler.GetOverduePotential)
	router.HandleFunc("POST /api/v1/dunning/{invoiceId}/remind", handler.CreateDunningReminder)
	router.HandleFunc("POST /api/v1/dunning/{id}/send", handler.SendDunningReminder)

	// Invoice-scoped dunning routes (used by frontend)
	router.HandleFunc("POST /api/v1/invoices/{id}/dunning", handler.CreateInvoiceDunning)
	router.HandleFunc("GET /api/v1/invoices/{id}/dunning", handler.GetInvoiceDunningHistory)

	// Export routes
	router.HandleFunc("GET /api/v1/export/datev", handler.ExportDATEV)
	router.HandleFunc("GET /api/v1/export/csv", handler.ExportCSV)

	return router
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy","service":"invoice-service"}`))
}

func readyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ready","service":"invoice-service"}`))
}

package httphandler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jeckersberger/EquipFlow/pkg/common/middleware"
)

func NewRouter(
	invoiceHandler *InvoiceHandler,
	quoteHandler *QuoteHandler,
	dunningHandler *DunningHandler,
	bankingHandler *BankingHandler,
	datevHandler *DatevHandler,
	portalHandler *PortalHandler,
	healthHandler http.HandlerFunc,
	livenessHandler http.HandlerFunc,
	jwtMiddleware func(http.Handler) http.Handler,
	corsMiddleware func(http.Handler) http.Handler,
	recoveryMiddleware func(http.Handler) http.Handler,
	requestIDMiddleware func(http.Handler) http.Handler,
) *chi.Mux {
	r := chi.NewRouter()

	r.Use(recoveryMiddleware)
	r.Use(middleware.MaxBodySize(1 << 20)) // 1 MB body limit
	r.Use(requestIDMiddleware)
	r.Use(corsMiddleware)

	r.Get("/health", healthHandler)
	r.Get("/livez", livenessHandler)

	// Public customer portal (no JWT required, token-based access)
	r.Route("/api/v1/public/quotes", func(r chi.Router) {
		r.Get("/{token}", portalHandler.GetPublicQuote)
		r.Post("/{token}/respond", portalHandler.RespondToQuote)
	})

	r.Group(func(r chi.Router) {
		r.Use(jwtMiddleware)

		r.Route("/api/v1/invoices", func(r chi.Router) {
			// Read: all authenticated users
			r.Get("/", invoiceHandler.List)
			r.Get("/search", invoiceHandler.Search)
			r.Get("/datev-export", datevHandler.ExportCSV)
			r.Get("/{id}", invoiceHandler.Get)
			r.Get("/{id}/items", invoiceHandler.ListItems)
			r.Get("/{id}/payments", invoiceHandler.ListPayments)
			r.Get("/{id}/pdf", invoiceHandler.GeneratePDF)

			// Write: manager or admin only
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole("admin", "manager"))
				r.Post("/", invoiceHandler.Create)
				r.Put("/{id}", invoiceHandler.Update)
				r.Patch("/{id}/finalize", invoiceHandler.Finalize)
				r.Post("/{id}/finalize", invoiceHandler.Finalize)
				r.Post("/{id}/items", invoiceHandler.AddItem)
				r.Delete("/{id}/items/{itemId}", invoiceHandler.RemoveItem)
				r.Post("/{id}/payments", invoiceHandler.AddPayment)
				r.Post("/{id}/send", invoiceHandler.SendEmail)
				r.Post("/{id}/partial", invoiceHandler.CreatePartialInvoice)
			})

			// Delete: admin only
			r.With(middleware.RequireRole("admin")).Delete("/{id}", invoiceHandler.Delete)
		})

		r.Route("/api/v1/quotes", func(r chi.Router) {
			r.Get("/", quoteHandler.List)
			r.Post("/", quoteHandler.Create)
			r.Get("/{id}", quoteHandler.Get)
			r.Put("/{id}", quoteHandler.Update)
			r.Delete("/{id}", quoteHandler.Delete)
			r.Post("/{id}/items", quoteHandler.AddItem)
			r.Get("/{id}/items", quoteHandler.ListItems)
			r.Delete("/{id}/items/{itemId}", quoteHandler.RemoveItem)
			r.Post("/{id}/convert-to-invoice", quoteHandler.ConvertToInvoice)
			r.Patch("/{id}/status", quoteHandler.UpdateStatus)
		})

		// Dunning (Mahnwesen) endpoints.
		r.Route("/api/v1/dunning", func(r chi.Router) {
			r.Get("/overdue", dunningHandler.ListOverdue)
			r.Post("/send-reminder", dunningHandler.SendReminder)
			r.Get("/entries/{invoiceId}", dunningHandler.ListEntries)
			r.Get("/config", dunningHandler.GetConfig)
			r.Put("/config", dunningHandler.UpdateConfig)
			r.Post("/check", dunningHandler.RunCheck)
		})

		// Banking (Bank-Import & Auto-Matching) endpoints.
		r.Route("/api/v1/banking", func(r chi.Router) {
			r.Post("/import", bankingHandler.ImportCSV)
			r.Post("/auto-match", bankingHandler.AutoMatch)
			r.Post("/confirm-match", bankingHandler.ConfirmMatch)
			r.Get("/transactions", bankingHandler.ListTransactions)
		})

		// DATEV Export endpoint.
		r.Route("/api/v1/export", func(r chi.Router) {
			r.Get("/datev", datevHandler.ExportCSV)
		})
	})

	return r
}

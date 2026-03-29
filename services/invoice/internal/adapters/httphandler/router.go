package httphandler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewRouter(
	invoiceHandler *InvoiceHandler,
	quoteHandler *QuoteHandler,
	healthHandler http.HandlerFunc,
	livenessHandler http.HandlerFunc,
	jwtMiddleware func(http.Handler) http.Handler,
	corsMiddleware func(http.Handler) http.Handler,
	recoveryMiddleware func(http.Handler) http.Handler,
	requestIDMiddleware func(http.Handler) http.Handler,
) *chi.Mux {
	r := chi.NewRouter()

	r.Use(recoveryMiddleware)
	r.Use(requestIDMiddleware)
	r.Use(corsMiddleware)

	r.Get("/health", healthHandler)
	r.Get("/livez", livenessHandler)

	r.Group(func(r chi.Router) {
		r.Use(jwtMiddleware)

		r.Route("/api/v1/invoices", func(r chi.Router) {
			r.Get("/", invoiceHandler.List)
			r.Post("/", invoiceHandler.Create)
			r.Get("/search", invoiceHandler.Search)
			r.Get("/{id}", invoiceHandler.Get)
			r.Put("/{id}", invoiceHandler.Update)
			r.Patch("/{id}/finalize", invoiceHandler.Finalize)
			r.Post("/{id}/items", invoiceHandler.AddItem)
			r.Get("/{id}/items", invoiceHandler.ListItems)
			r.Delete("/{id}/items/{itemId}", invoiceHandler.RemoveItem)
			r.Post("/{id}/payments", invoiceHandler.AddPayment)
			r.Get("/{id}/payments", invoiceHandler.ListPayments)
			r.Get("/{id}/pdf", invoiceHandler.GeneratePDF)
			r.Post("/{id}/send", invoiceHandler.SendEmail)
		})

		r.Route("/api/v1/quotes", func(r chi.Router) {
			r.Get("/", quoteHandler.List)
			r.Post("/", quoteHandler.Create)
			r.Get("/{id}", quoteHandler.Get)
			r.Put("/{id}", quoteHandler.Update)
			r.Post("/{id}/items", quoteHandler.AddItem)
			r.Get("/{id}/items", quoteHandler.ListItems)
			r.Delete("/{id}/items/{itemId}", quoteHandler.RemoveItem)
			r.Post("/{id}/convert-to-invoice", quoteHandler.ConvertToInvoice)
			r.Patch("/{id}/status", quoteHandler.UpdateStatus)
		})
	})

	return r
}

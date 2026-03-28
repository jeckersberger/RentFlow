package httphandler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewRouter(
	categoryHandler *CategoryHandler,
	expenseHandler *ExpenseHandler,
	receiptHandler *ReceiptHandler,
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

		r.Route("/api/v1/expense-categories", func(r chi.Router) {
			r.Post("/", categoryHandler.Create)
			r.Get("/", categoryHandler.List)
			r.Get("/{id}", categoryHandler.Get)
			r.Put("/{id}", categoryHandler.Update)
		})

		r.Route("/api/v1/expenses", func(r chi.Router) {
			r.Post("/", expenseHandler.Create)
			r.Get("/", expenseHandler.List)
			r.Get("/{id}", expenseHandler.Get)
			r.Put("/{id}", expenseHandler.Update)
			r.Patch("/{id}/approve", expenseHandler.Approve)
			r.Post("/{id}/receipts", receiptHandler.AddReceipt)
			r.Get("/{id}/receipts", receiptHandler.ListReceipts)
		})
	})

	return r
}

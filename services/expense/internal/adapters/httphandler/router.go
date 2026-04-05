package httphandler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jeckersberger/EquipFlow/pkg/common/middleware"
)

func NewRouter(
	categoryHandler *CategoryHandler,
	expenseHandler *ExpenseHandler,
	receiptHandler *ReceiptHandler,
	recurringHandler *RecurringHandler,
	budgetHandler *BudgetHandler,
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

	r.Group(func(r chi.Router) {
		r.Use(jwtMiddleware)

		r.Route("/api/v1/expense-categories", func(r chi.Router) {
			// Read: all authenticated users
			r.Get("/", categoryHandler.List)
			r.Get("/{id}", categoryHandler.Get)

			// Write: admin, manager, or user
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole("admin", "manager", "user"))
				r.Post("/", categoryHandler.Create)
				r.Put("/{id}", categoryHandler.Update)
			})
		})

		r.Route("/api/v1/expenses", func(r chi.Router) {
			// Read: all authenticated users
			r.Get("/", expenseHandler.List)
			r.Get("/{id}", expenseHandler.Get)
			r.Get("/{id}/receipts", receiptHandler.ListReceipts)

			// Write: admin, manager, or user
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole("admin", "manager", "user"))
				r.Post("/", expenseHandler.Create)
				r.Put("/{id}", expenseHandler.Update)
				r.Patch("/{id}/approve", expenseHandler.Approve)
				r.Post("/{id}/receipts", receiptHandler.AddReceipt)
			})

			r.Route("/recurring", func(r chi.Router) {
				// Read: all authenticated users
				r.Get("/", recurringHandler.List)

				// Write: admin, manager, or user
				r.Group(func(r chi.Router) {
					r.Use(middleware.RequireRole("admin", "manager", "user"))
					r.Post("/", recurringHandler.Create)
					r.Put("/{id}", recurringHandler.Update)
				})

				// Delete: admin or manager only
				r.With(middleware.RequireRole("admin", "manager")).Delete("/{id}", recurringHandler.Delete)
			})

			r.Route("/budgets", func(r chi.Router) {
				// Read: all authenticated users
				r.Get("/", budgetHandler.List)

				// Write: admin, manager, or user
				r.Group(func(r chi.Router) {
					r.Use(middleware.RequireRole("admin", "manager", "user"))
					r.Post("/", budgetHandler.Create)
					r.Put("/{id}", budgetHandler.Update)
				})
			})
		})
	})

	return r
}

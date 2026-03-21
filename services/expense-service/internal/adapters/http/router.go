package http

import (
	"net/http"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/expense-service/internal/application"
)

func NewRouter(
	expSvc *application.ExpenseService,
	categorySvc *application.CategoryService,
	budgetSvc *application.BudgetService,
	logger logger.Logger,
) *http.ServeMux {
	router := http.NewServeMux()
	handler := NewHandler(expSvc, categorySvc, budgetSvc, logger)

	// Health & readiness
	router.HandleFunc("GET /health", healthHandler)
	router.HandleFunc("GET /ready", readyHandler)

	// Expense routes
	router.HandleFunc("POST /api/v1/expenses", handler.CreateExpense)
	router.HandleFunc("GET /api/v1/expenses", handler.ListExpenses)
	router.HandleFunc("GET /api/v1/expenses/{id}", handler.GetExpense)
	router.HandleFunc("PUT /api/v1/expenses/{id}", handler.UpdateExpense)
	router.HandleFunc("POST /api/v1/expenses/{id}/approve", handler.ApproveExpense)
	router.HandleFunc("POST /api/v1/expenses/{id}/reject", handler.RejectExpense)
	router.HandleFunc("DELETE /api/v1/expenses/{id}", handler.DeleteExpense)

	// Category routes
	router.HandleFunc("POST /api/v1/expense-categories", handler.CreateCategory)
	router.HandleFunc("GET /api/v1/expense-categories", handler.ListCategories)
	router.HandleFunc("GET /api/v1/expense-categories/{id}", handler.GetCategory)
	router.HandleFunc("PUT /api/v1/expense-categories/{id}", handler.UpdateCategory)
	router.HandleFunc("DELETE /api/v1/expense-categories/{id}", handler.DeleteCategory)

	// Budget routes
	router.HandleFunc("POST /api/v1/budgets", handler.CreateBudget)
	router.HandleFunc("GET /api/v1/budgets", handler.ListBudgets)
	router.HandleFunc("GET /api/v1/budgets/{id}", handler.GetBudget)
	router.HandleFunc("DELETE /api/v1/budgets/{id}", handler.DeleteBudget)
	router.HandleFunc("GET /api/v1/budgets/status", handler.GetBudgetStatus)

	return router
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy","service":"expense-service"}`))
}

func readyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ready","service":"expense-service"}`))
}

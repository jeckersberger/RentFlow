package httphandler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/pkg/common/errors"
	"github.com/jeckersberger/EquipFlow/pkg/common/middleware"
	"github.com/jeckersberger/EquipFlow/pkg/common/response"
	"github.com/jeckersberger/EquipFlow/services/invoice/internal/application"
)

// BankingHandler handles bank import and auto-matching HTTP endpoints.
type BankingHandler struct {
	bankingService *application.BankingService
	logger         zerolog.Logger
}

// NewBankingHandler creates a new BankingHandler.
func NewBankingHandler(bankingService *application.BankingService, logger zerolog.Logger) *BankingHandler {
	return &BankingHandler{
		bankingService: bankingService,
		logger:         logger.With().Str("handler", "banking").Logger(),
	}
}

// ImportCSV handles CSV upload and imports bank transactions.
func (h *BankingHandler) ImportCSV(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	// Read the CSV data from request body (max 10MB)
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)

	csvData, err := io.ReadAll(r.Body)
	if err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "CSV-Datei konnte nicht gelesen werden"))
		return
	}

	if len(csvData) == 0 {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Leere CSV-Datei"))
		return
	}

	transactions, err := h.bankingService.ImportCSV(r.Context(), claims.TenantID, csvData)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Created(w, map[string]interface{}{
		"imported": len(transactions),
		"transactions": transactions,
	})
}

// AutoMatch runs auto-matching of unmatched bank transactions against open invoices.
func (h *BankingHandler) AutoMatch(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	matched, err := h.bankingService.AutoMatch(r.Context(), claims.TenantID)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, map[string]interface{}{
		"matched": matched,
	})
}

// ConfirmMatch manually confirms a match between a bank transaction and an invoice.
func (h *BankingHandler) ConfirmMatch(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	var req application.ConfirmMatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiger Request-Body"))
		return
	}

	if err := h.bankingService.ConfirmMatch(r.Context(), claims.TenantID, req); err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, map[string]string{"status": "confirmed"})
}

// ListTransactions returns bank transactions with optional filter.
func (h *BankingHandler) ListTransactions(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	var matched *bool
	matchedParam := r.URL.Query().Get("matched")
	if matchedParam == "true" {
		t := true
		matched = &t
	} else if matchedParam == "false" {
		f := false
		matched = &f
	}

	transactions, err := h.bankingService.ListTransactions(r.Context(), claims.TenantID, matched)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, transactions)
}

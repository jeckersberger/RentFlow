package httphandler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/pkg/common/errors"
	"github.com/jeckersberger/EquipFlow/pkg/common/middleware"
	"github.com/jeckersberger/EquipFlow/pkg/common/response"
	"github.com/jeckersberger/EquipFlow/services/expense/internal/application"
)

type ReceiptHandler struct {
	receiptService *application.ReceiptService
	logger         zerolog.Logger
}

func NewReceiptHandler(receiptService *application.ReceiptService, logger zerolog.Logger) *ReceiptHandler {
	return &ReceiptHandler{
		receiptService: receiptService,
		logger:         logger.With().Str("handler", "expense-receipt").Logger(),
	}
}

func (h *ReceiptHandler) AddReceipt(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	expenseID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	var req application.AddReceiptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiger Request-Body"))
		return
	}

	receipt, err := h.receiptService.AddReceipt(r.Context(), expenseID, claims.TenantID, claims.UserID, req)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Created(w, receipt)
}

func (h *ReceiptHandler) ListReceipts(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	expenseID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	receipts, err := h.receiptService.ListReceipts(r.Context(), expenseID, claims.TenantID)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, receipts)
}

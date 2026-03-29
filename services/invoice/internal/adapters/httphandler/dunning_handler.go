package httphandler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/pkg/common/errors"
	"github.com/jeckersberger/EquipFlow/pkg/common/middleware"
	"github.com/jeckersberger/EquipFlow/pkg/common/response"
	"github.com/jeckersberger/EquipFlow/services/invoice/internal/application"
)

// DunningHandler handles dunning (Mahnwesen) HTTP endpoints.
type DunningHandler struct {
	dunningService *application.DunningService
	logger         zerolog.Logger
}

// NewDunningHandler creates a new DunningHandler.
func NewDunningHandler(dunningService *application.DunningService, logger zerolog.Logger) *DunningHandler {
	return &DunningHandler{
		dunningService: dunningService,
		logger:         logger.With().Str("handler", "dunning").Logger(),
	}
}

// ListOverdue returns overdue invoices with suggested dunning levels.
func (h *DunningHandler) ListOverdue(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	overdue, err := h.dunningService.GetOverdueInvoices(r.Context(), claims.TenantID)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, overdue)
}

// SendReminder creates a dunning entry for an invoice.
func (h *DunningHandler) SendReminder(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	var req application.CreateReminderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiger Request-Body"))
		return
	}

	entry, err := h.dunningService.CreateReminder(r.Context(), claims.TenantID, req)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Created(w, entry)
}

// ListEntries returns dunning history for a specific invoice.
func (h *DunningHandler) ListEntries(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	invoiceID, err := parseUUID(chi.URLParam(r, "invoiceId"))
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	entries, err := h.dunningService.ListByInvoice(r.Context(), invoiceID, claims.TenantID)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, entries)
}

// GetConfig returns the dunning configuration for the current tenant.
func (h *DunningHandler) GetConfig(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	cfg, err := h.dunningService.GetConfig(r.Context(), claims.TenantID)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, cfg)
}

// UpdateConfig updates the dunning configuration for the current tenant.
func (h *DunningHandler) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	var req application.UpdateDunningConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiger Request-Body"))
		return
	}

	cfg, err := h.dunningService.UpdateConfig(r.Context(), claims.TenantID, req)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, cfg)
}

// RunCheck triggers an automatic dunning check for the current tenant.
func (h *DunningHandler) RunCheck(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	entries, err := h.dunningService.RunDunningCheck(r.Context(), claims.TenantID)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, entries)
}

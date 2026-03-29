package httphandler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/pkg/common/errors"
	"github.com/jeckersberger/EquipFlow/pkg/common/middleware"
	"github.com/jeckersberger/EquipFlow/pkg/common/pagination"
	"github.com/jeckersberger/EquipFlow/pkg/common/response"
	"github.com/jeckersberger/EquipFlow/services/inventory/internal/application"
)

// PricingHandler handles pricing HTTP endpoints.
type PricingHandler struct {
	pricingService *application.PricingService
	logger         zerolog.Logger
}

// NewPricingHandler creates a new PricingHandler.
func NewPricingHandler(pricingService *application.PricingService, logger zerolog.Logger) *PricingHandler {
	return &PricingHandler{
		pricingService: pricingService,
		logger:         logger.With().Str("handler", "pricing").Logger(),
	}
}

// CalculatePrice calculates the rental price for a specific equipment item.
func (h *PricingHandler) CalculatePrice(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	equipmentID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	var req application.PriceCalculationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiger Request-Body"))
		return
	}
	req.EquipmentID = equipmentID

	// If no category_id supplied in body, try to fill from equipment lookup
	// (done in service via equipmentRepo fallback)

	result, err := h.pricingService.Calculate(r.Context(), claims.TenantID, req)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, result)
}

// ListPriceRules returns all price rules for the current tenant.
func (h *PricingHandler) ListPriceRules(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	p := pagination.Parse(r)

	rules, total, err := h.pricingService.ListPriceRules(r.Context(), claims.TenantID, p.Page, p.PerPage)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.SuccessWithMeta(w, rules, response.Meta{
		Page:    p.Page,
		PerPage: p.PerPage,
		Total:   total,
	})
}

// CreatePriceRule creates a new price rule.
func (h *PricingHandler) CreatePriceRule(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	var req application.CreatePriceRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiger Request-Body"))
		return
	}

	rule, err := h.pricingService.CreatePriceRule(r.Context(), claims.TenantID, req)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Created(w, rule)
}

// UpdatePriceRule updates an existing price rule.
func (h *PricingHandler) UpdatePriceRule(w http.ResponseWriter, r *http.Request) {
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

	var req application.UpdatePriceRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiger Request-Body"))
		return
	}

	rule, err := h.pricingService.UpdatePriceRule(r.Context(), id, claims.TenantID, req)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, rule)
}

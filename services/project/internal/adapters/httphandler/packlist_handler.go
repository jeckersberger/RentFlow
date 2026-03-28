package httphandler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/pkg/common/errors"
	"github.com/jeckersberger/EquipFlow/pkg/common/middleware"
	"github.com/jeckersberger/EquipFlow/pkg/common/response"
	"github.com/jeckersberger/EquipFlow/services/project/internal/application"
)

type PacklistHandler struct {
	packlistService *application.PacklistService
	logger          zerolog.Logger
}

func NewPacklistHandler(packlistService *application.PacklistService, logger zerolog.Logger) *PacklistHandler {
	return &PacklistHandler{
		packlistService: packlistService,
		logger:          logger.With().Str("handler", "packlist").Logger(),
	}
}

func (h *PacklistHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	var req application.CreatePacklistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiger Request-Body"))
		return
	}

	created, err := h.packlistService.Create(r.Context(), claims.TenantID, claims.UserID, req)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Created(w, created)
}

func (h *PacklistHandler) Get(w http.ResponseWriter, r *http.Request) {
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

	pl, err := h.packlistService.GetByID(r.Context(), id, claims.TenantID)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, pl)
}

func (h *PacklistHandler) ListByProject(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	projectID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	items, err := h.packlistService.ListByProject(r.Context(), claims.TenantID, projectID)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, items)
}

func (h *PacklistHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
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

	var req application.UpdatePacklistStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiger Request-Body"))
		return
	}

	if err := h.packlistService.UpdateStatus(r.Context(), id, claims.TenantID, req); err != nil {
		errors.HandleError(w, err)
		return
	}

	response.NoContent(w)
}

func (h *PacklistHandler) AddItem(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	packlistID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	var req application.AddPacklistItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiger Request-Body"))
		return
	}

	item, err := h.packlistService.AddItem(r.Context(), claims.TenantID, packlistID, req)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Created(w, item)
}

func (h *PacklistHandler) GetItems(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	packlistID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	items, err := h.packlistService.GetItems(r.Context(), claims.TenantID, packlistID)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, items)
}

func (h *PacklistHandler) UpdateItemPacked(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	itemID, err := parseUUID(chi.URLParam(r, "itemId"))
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	var req application.UpdatePackedRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiger Request-Body"))
		return
	}

	if err := h.packlistService.UpdateItemPacked(r.Context(), claims.TenantID, itemID, req, claims.UserID); err != nil {
		errors.HandleError(w, err)
		return
	}

	response.NoContent(w)
}

func (h *PacklistHandler) UpdateItemReturned(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	itemID, err := parseUUID(chi.URLParam(r, "itemId"))
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	var req application.UpdateReturnedRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiger Request-Body"))
		return
	}

	if err := h.packlistService.UpdateItemReturned(r.Context(), claims.TenantID, itemID, req); err != nil {
		errors.HandleError(w, err)
		return
	}

	response.NoContent(w)
}

package http

import (
	"encoding/json"
	"net/http"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/federation-service/internal/application"
	"github.com/jeckersberger/rentflow/services/federation-service/internal/domain"
)

type Handler struct {
	peerSvc    *application.PeerService
	sharingSvc *application.SharingService
	logger     logger.Logger
}

func NewHandler(
	peerSvc *application.PeerService,
	sharingSvc *application.SharingService,
	log logger.Logger,
) *Handler {
	return &Handler{
		peerSvc:    peerSvc,
		sharingSvc: sharingSvc,
		logger:     log,
	}
}

// Peers

func (h *Handler) CreatePeer(w http.ResponseWriter, r *http.Request) {
	var cmd application.CreatePeerCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}
	cmd.TenantID = tenantID

	dto, err := h.peerSvc.CreatePeer(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusCreated, dto)
}

func (h *Handler) GetPeer(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	dto, err := h.peerSvc.GetPeer(r.Context(), tenantID, id)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, dto)
}

func (h *Handler) ListPeers(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	dtos, err := h.peerSvc.ListPeers(r.Context(), tenantID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{"data": dtos})
}

func (h *Handler) UpdatePeer(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var cmd application.UpdatePeerCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}
	cmd.TenantID = tenantID
	cmd.PeerID = id

	dto, err := h.peerSvc.UpdatePeer(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, dto)
}

func (h *Handler) DeletePeer(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	if err := h.peerSvc.DeletePeer(r.Context(), tenantID, id); err != nil {
		h.handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Sharing

func (h *Handler) ShareEquipment(w http.ResponseWriter, r *http.Request) {
	var cmd application.ShareEquipmentCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}
	cmd.TenantID = tenantID

	dto, err := h.sharingSvc.ShareEquipment(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusCreated, dto)
}

func (h *Handler) ListCatalog(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	dtos, err := h.sharingSvc.ListCatalog(r.Context(), tenantID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{"data": dtos})
}

// Requests

func (h *Handler) CreateShareRequest(w http.ResponseWriter, r *http.Request) {
	var cmd application.CreateShareRequestCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}
	cmd.TenantID = tenantID

	dto, err := h.sharingSvc.CreateShareRequest(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusCreated, dto)
}

func (h *Handler) ListRequests(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	dtos, err := h.sharingSvc.ListRequests(r.Context(), tenantID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{"data": dtos})
}

func (h *Handler) UpdateRequest(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	var body map[string]string
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	action := body["action"]
	var dto *application.ShareRequestDTO
	var err error

	if action == "approve" {
		dto, err = h.sharingSvc.ApproveRequest(r.Context(), tenantID, id)
	} else if action == "reject" {
		dto, err = h.sharingSvc.RejectRequest(r.Context(), tenantID, id)
	} else {
		h.respondError(w, http.StatusBadRequest, "invalid action")
		return
	}

	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, dto)
}

// Handshake

func (h *Handler) Handshake(w http.ResponseWriter, r *http.Request) {
	var body map[string]string
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	peerID := body["peer_id"]
	if peerID == "" {
		h.respondError(w, http.StatusBadRequest, "peer_id required")
		return
	}

	if err := h.peerSvc.RecordHandshake(r.Context(), tenantID, peerID); err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Helpers

func (h *Handler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) respondError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func (h *Handler) handleError(w http.ResponseWriter, err error) {
	if err == domain.ErrPeerNotFound || err == domain.ErrEquipmentNotFound || err == domain.ErrShareRequestNotFound {
		h.respondError(w, http.StatusNotFound, err.Error())
		return
	}
	if err == domain.ErrTenantIDRequired || err == domain.ErrInvalidInput {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err == domain.ErrPeerBlocked {
		h.respondError(w, http.StatusForbidden, err.Error())
		return
	}

	h.logger.Error("Unhandled error", err)
	h.respondError(w, http.StatusInternalServerError, "internal server error")
}

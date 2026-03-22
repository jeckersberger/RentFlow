package http

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/federation-service/internal/application"
	"github.com/jeckersberger/rentflow/services/federation-service/internal/domain"
)

type Handlers struct {
	partnerService      *application.PartnerService
	sharingService      *application.SharingService
	certificateService  *application.CertificateService
	log                 logger.Logger
}

func NewHandlers(ps *application.PartnerService, ss *application.SharingService, cs *application.CertificateService, log logger.Logger) *Handlers {
	return &Handlers{
		partnerService:     ps,
		sharingService:     ss,
		certificateService: cs,
		log:                log,
	}
}

// Partners

func (h *Handlers) CreatePartner(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		http.Error(w, "Missing X-Tenant-ID header", http.StatusBadRequest)
		return
	}

	var req application.CreatePartnerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	tid, _ := uuid.Parse(tenantID)
	resp, err := h.partnerService.CreatePartner(r.Context(), tid, &req)
	if err != nil {
		h.log.Error("Failed to create partner", err)
		http.Error(w, "Failed to create partner", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (h *Handlers) GetPartner(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	partnerID, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "Invalid partner ID", http.StatusBadRequest)
		return
	}

	resp, err := h.partnerService.GetPartner(r.Context(), partnerID)
	if err == domain.ErrPartnerNotFound {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if err != nil {
		h.log.Error("Failed to get partner", err)
		http.Error(w, "Failed to get partner", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handlers) ListPartners(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		http.Error(w, "Missing X-Tenant-ID header", http.StatusBadRequest)
		return
	}

	tid, _ := uuid.Parse(tenantID)
	resps, err := h.partnerService.ListPartners(r.Context(), tid)
	if err != nil {
		h.log.Error("Failed to list partners", err)
		http.Error(w, "Failed to list partners", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resps)
}

func (h *Handlers) ActivatePartner(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	partnerID, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "Invalid partner ID", http.StatusBadRequest)
		return
	}

	if err := h.partnerService.ActivatePartner(r.Context(), partnerID); err != nil {
		h.log.Error("Failed to activate partner", err)
		http.Error(w, "Failed to activate partner", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handlers) SuspendPartner(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	partnerID, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "Invalid partner ID", http.StatusBadRequest)
		return
	}

	if err := h.partnerService.SuspendPartner(r.Context(), partnerID); err != nil {
		h.log.Error("Failed to suspend partner", err)
		http.Error(w, "Failed to suspend partner", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handlers) GetPartnerEquipment(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	partnerID, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "Invalid partner ID", http.StatusBadRequest)
		return
	}

	resps, err := h.sharingService.GetPartnerEquipment(r.Context(), partnerID)
	if err != nil {
		h.log.Error("Failed to get partner equipment", err)
		http.Error(w, "Failed to get partner equipment", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resps)
}

func (h *Handlers) SyncPartnerEquipment(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	partnerID, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "Invalid partner ID", http.StatusBadRequest)
		return
	}

	if err := h.sharingService.SyncPartnerEquipment(r.Context(), partnerID); err != nil {
		h.log.Error("Failed to sync partner equipment", err)
		http.Error(w, "Failed to sync partner equipment", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"status": "syncing"})
}

// Sub-Rental Requests

func (h *Handlers) CreateRequest(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		http.Error(w, "Missing X-Tenant-ID header", http.StatusBadRequest)
		return
	}

	var req application.CreateSubRentalRequestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	tid, _ := uuid.Parse(tenantID)
	resp, err := h.sharingService.CreateRequest(r.Context(), tid, &req)
	if err != nil {
		h.log.Error("Failed to create request", err)
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (h *Handlers) GetRequest(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	reqID, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "Invalid request ID", http.StatusBadRequest)
		return
	}

	resp, err := h.sharingService.GetRequest(r.Context(), reqID)
	if err == domain.ErrRequestNotFound {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if err != nil {
		h.log.Error("Failed to get request", err)
		http.Error(w, "Failed to get request", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handlers) ListRequests(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		http.Error(w, "Missing X-Tenant-ID header", http.StatusBadRequest)
		return
	}

	tid, _ := uuid.Parse(tenantID)
	resps, err := h.sharingService.ListRequests(r.Context(), tid)
	if err != nil {
		h.log.Error("Failed to list requests", err)
		http.Error(w, "Failed to list requests", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resps)
}

func (h *Handlers) AcceptRequest(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	reqID, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "Invalid request ID", http.StatusBadRequest)
		return
	}

	if err := h.sharingService.AcceptRequest(r.Context(), reqID); err != nil {
		h.log.Error("Failed to accept request", err)
		http.Error(w, "Failed to accept request", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handlers) RejectRequest(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	reqID, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "Invalid request ID", http.StatusBadRequest)
		return
	}

	if err := h.sharingService.RejectRequest(r.Context(), reqID); err != nil {
		h.log.Error("Failed to reject request", err)
		http.Error(w, "Failed to reject request", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handlers) CompleteRequest(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	reqID, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "Invalid request ID", http.StatusBadRequest)
		return
	}

	var body struct {
		HandoverDocumentID uuid.UUID `json:"handover_document_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.sharingService.CompleteRequest(r.Context(), reqID, body.HandoverDocumentID); err != nil {
		h.log.Error("Failed to complete request", err)
		http.Error(w, "Failed to complete request", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Certificates

func (h *Handlers) GenerateCertPair(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		http.Error(w, "Missing X-Tenant-ID header", http.StatusBadRequest)
		return
	}

	tid, _ := uuid.Parse(tenantID)
	resp, err := h.certificateService.GenerateCertPair(r.Context(), tid)
	if err != nil {
		h.log.Error("Failed to generate certificate pair", err)
		http.Error(w, "Failed to generate certificate pair", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (h *Handlers) ListCertificates(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		http.Error(w, "Missing X-Tenant-ID header", http.StatusBadRequest)
		return
	}

	tid, _ := uuid.Parse(tenantID)
	resps, err := h.certificateService.GetCertificates(r.Context(), tid)
	if err != nil {
		h.log.Error("Failed to list certificates", err)
		http.Error(w, "Failed to list certificates", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resps)
}

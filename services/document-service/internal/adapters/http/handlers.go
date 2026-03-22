package http

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/document-service/internal/application"
	"github.com/jeckersberger/rentflow/services/document-service/internal/domain"
)

type Handler struct {
	docSvc  *application.DocumentService
	sigSvc  *application.SignatureService
	chkSvc  *application.ChecksumService
	logger  logger.Logger
}

func NewHandler(
	docSvc *application.DocumentService,
	sigSvc *application.SignatureService,
	chkSvc *application.ChecksumService,
	log logger.Logger,
) *Handler {
	return &Handler{
		docSvc:  docSvc,
		sigSvc:  sigSvc,
		chkSvc:  chkSvc,
		logger:  log,
	}
}

func (h *Handler) CreateDocument(w http.ResponseWriter, r *http.Request) {
	var req application.CreateDocumentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tenantID := r.Header.Get("X-Tenant-ID")
	userID := r.Header.Get("X-User-ID")
	if tenantID == "" || userID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID and user ID required")
		return
	}

	resp, err := h.docSvc.CreateDocument(r.Context(), tenantID, userID, req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusCreated, resp)
}

func (h *Handler) GetDocument(w http.ResponseWriter, r *http.Request) {
	docID := strings.TrimPrefix(r.URL.Path, "/api/v1/documents/")
	if idx := strings.Index(docID, "/"); idx != -1 {
		docID = docID[:idx]
	}

	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	resp, err := h.docSvc.GetDocument(r.Context(), tenantID, docID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, resp)
}

func (h *Handler) ListDocuments(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	resp, err := h.docSvc.ListDocuments(r.Context(), tenantID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, resp)
}

func (h *Handler) UpdateDocument(w http.ResponseWriter, r *http.Request) {
	docID := strings.TrimPrefix(r.URL.Path, "/api/v1/documents/")
	if idx := strings.Index(docID, "/"); idx != -1 {
		docID = docID[:idx]
	}

	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	resp, err := h.docSvc.GetDocument(r.Context(), tenantID, docID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, resp)
}

func (h *Handler) GenerateDocument(w http.ResponseWriter, r *http.Request) {
	docID := strings.TrimPrefix(r.URL.Path, "/api/v1/documents/")
	docID = strings.TrimSuffix(docID, "/generate")

	var req application.GenerateFromTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tenantID := r.Header.Get("X-Tenant-ID")
	userID := r.Header.Get("X-User-ID")
	if tenantID == "" || userID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID and user ID required")
		return
	}

	resp, err := h.docSvc.GenerateFromTemplate(r.Context(), tenantID, userID, docID, req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, resp)
}

func (h *Handler) ArchiveDocument(w http.ResponseWriter, r *http.Request) {
	docID := strings.TrimPrefix(r.URL.Path, "/api/v1/documents/")

	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	err := h.docSvc.ArchiveDocument(r.Context(), tenantID, docID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) RequestSignature(w http.ResponseWriter, r *http.Request) {
	docID := strings.TrimPrefix(r.URL.Path, "/api/v1/documents/")
	docID = strings.TrimSuffix(docID, "/sign")

	var req application.RequestSignatureRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	resp, err := h.sigSvc.RequestSignature(r.Context(), tenantID, docID, req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusCreated, resp)
}

func (h *Handler) SubmitSignature(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/v1/documents/"), "/")
	if len(parts) < 2 {
		h.respondError(w, http.StatusBadRequest, "invalid request")
		return
	}
	docID := parts[0]
	sigID := parts[1]

	var req application.SubmitSignatureRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	resp, err := h.sigSvc.SubmitSignature(r.Context(), tenantID, docID, sigID, req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, resp)
}

func (h *Handler) GetSignatures(w http.ResponseWriter, r *http.Request) {
	docID := strings.TrimPrefix(r.URL.Path, "/api/v1/documents/")
	docID = strings.TrimSuffix(docID, "/signatures")

	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	resp, err := h.sigSvc.GetSignatures(r.Context(), tenantID, docID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, resp)
}

func (h *Handler) GetVersions(w http.ResponseWriter, r *http.Request) {
	docID := strings.TrimPrefix(r.URL.Path, "/api/v1/documents/")
	docID = strings.TrimSuffix(docID, "/versions")

	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	resp, err := h.docSvc.GetVersions(r.Context(), tenantID, docID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, resp)
}

func (h *Handler) GenerateDeliveryNote(w http.ResponseWriter, r *http.Request) {
	projectID := strings.TrimPrefix(r.URL.Path, "/api/v1/documents/from-project/")

	tenantID := r.Header.Get("X-Tenant-ID")
	userID := r.Header.Get("X-User-ID")
	if tenantID == "" || userID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID and user ID required")
		return
	}

	resp, err := h.docSvc.GenerateDeliveryNote(r.Context(), tenantID, userID, projectID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusCreated, resp)
}

func (h *Handler) VerifyChecksumChain(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	resp, err := h.chkSvc.VerifyChecksumChain(r.Context(), tenantID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, resp)
}

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
	switch err {
	case domain.ErrDocumentNotFound, domain.ErrSignatureNotFound, domain.ErrVersionNotFound:
		h.respondError(w, http.StatusNotFound, err.Error())
	case domain.ErrInvalidInput, domain.ErrInvalidDocumentType:
		h.respondError(w, http.StatusBadRequest, err.Error())
	case domain.ErrTenantIDRequired:
		h.respondError(w, http.StatusUnauthorized, err.Error())
	case domain.ErrChecksumMismatch:
		h.respondError(w, http.StatusConflict, err.Error())
	default:
		h.logger.Error("Handler error", err)
		h.respondError(w, http.StatusInternalServerError, "internal server error")
	}
}

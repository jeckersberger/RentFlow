package http

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
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

func (h *Handler) UploadScan(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	userID := r.Header.Get("X-User-ID")
	if tenantID == "" || userID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID and user ID required")
		return
	}

	// Parse multipart form (max 50MB)
	if err := r.ParseMultipartForm(50 * 1024 * 1024); err != nil {
		h.respondError(w, http.StatusBadRequest, "failed to parse form")
		return
	}

	// Get file from form
	file, handler, err := r.FormFile("file")
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "file is required")
		return
	}
	defer file.Close()

	// Get form values
	referenceID := r.FormValue("reference_id")
	scanType := r.FormValue("scan_type")

	if referenceID == "" || scanType == "" {
		h.respondError(w, http.StatusBadRequest, "reference_id and scan_type are required")
		return
	}

	// Create document record for scan
	docID := uuid.New().String()
	docNumber := "SCAN-" + docID[:8]

	// Create request
	createReq := application.CreateDocumentRequest{
		DocumentType:   application.DocumentTypeDTO(scanType),
		ReferenceID:    referenceID,
		DocumentNumber: docNumber,
		Title:          handler.Filename,
		Metadata: map[string]interface{}{
			"original_filename": handler.Filename,
			"file_size":         handler.Size,
			"content_type":      handler.Header.Get("Content-Type"),
		},
	}

	// Create document
	docResp, err := h.docSvc.CreateDocument(r.Context(), tenantID, userID, createReq)
	if err != nil {
		h.handleError(w, err)
		return
	}

	// Save uploaded file to disk
	filePath := filepath.Join("/documents", tenantID, docID, handler.Filename)
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		h.respondError(w, http.StatusInternalServerError, "failed to save file")
		return
	}

	dst, err := os.Create(filePath)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "failed to create file")
		return
	}
	defer dst.Close()

	if _, err := dst.ReadFrom(file); err != nil {
		h.respondError(w, http.StatusInternalServerError, "failed to write file")
		return
	}

	h.logger.Info("Scan uploaded", "id", docID, "tenant", tenantID)
	h.respondJSON(w, http.StatusCreated, docResp)
}

func (h *Handler) GetPublicSignaturePage(w http.ResponseWriter, r *http.Request) {
	sigID := strings.TrimPrefix(r.URL.Path, "/api/v1/public/sign/")

	// Get signature (no tenant check)
	sig, err := h.sigSvc.GetSignatureByID(r.Context(), sigID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	resp := &application.PublicSignatureResponse{
		ID:          sig.ID,
		DocumentID:  sig.DocumentID,
		SignerName:  sig.SignerName,
		SignerEmail: sig.SignerEmail,
		SignedAt:    sig.SignedAt,
		Verified:    sig.Verified,
	}

	h.respondJSON(w, http.StatusOK, resp)
}

func (h *Handler) SubmitPublicSignature(w http.ResponseWriter, r *http.Request) {
	sigID := strings.TrimPrefix(r.URL.Path, "/api/v1/public/sign/")

	var req application.PublicSignatureSubmitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.SignatureData == "" || req.SignerName == "" {
		h.respondError(w, http.StatusBadRequest, "signature_data and signer_name are required")
		return
	}

	// Get signature to find document/tenant info
	sig, err := h.sigSvc.GetSignatureByID(r.Context(), sigID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	// Update signature (public endpoint, no tenant check needed)
	resp, err := h.sigSvc.SubmitPublicSignature(r.Context(), sig.DocumentID, sigID, req)
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

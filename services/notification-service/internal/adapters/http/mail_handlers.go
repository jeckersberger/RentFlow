package http

import (
	"encoding/json"
	nethttp "net/http"

	"github.com/google/uuid"
	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/notification-service/internal/application"
	"github.com/jeckersberger/rentflow/services/notification-service/internal/domain"
)

// MailHandler — HTTP-Handler für das E-Mail-System
type MailHandler struct {
	mailService *application.MailService
	log         logger.Logger
}

func NewMailHandler(mailService *application.MailService, log logger.Logger) *MailHandler {
	return &MailHandler{
		mailService: mailService,
		log:         log,
	}
}

// --- Postfächer ---

// ListMailboxes gibt alle Postfächer eines Mandanten zurück
func (h *MailHandler) ListMailboxes(w nethttp.ResponseWriter, r *nethttp.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, nethttp.StatusBadRequest, "X-Tenant-ID header required")
		return
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		h.respondError(w, nethttp.StatusBadRequest, "Invalid tenant ID")
		return
	}

	mailboxes, err := h.mailService.ListMailboxes(r.Context(), tenantUUID)
	if err != nil {
		h.respondError(w, nethttp.StatusInternalServerError, "Failed to list mailboxes")
		return
	}

	// Passwort aus Antwort entfernen
	var response []application.MailboxResponse
	for _, mb := range mailboxes {
		response = append(response, toMailboxResponse(mb))
	}

	h.respondJSON(w, nethttp.StatusOK, response)
}

// CreateMailbox erstellt ein neues Postfach
func (h *MailHandler) CreateMailbox(w nethttp.ResponseWriter, r *nethttp.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, nethttp.StatusBadRequest, "X-Tenant-ID header required")
		return
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		h.respondError(w, nethttp.StatusBadRequest, "Invalid tenant ID")
		return
	}

	var req application.CreateMailboxRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, nethttp.StatusBadRequest, "Invalid request body")
		return
	}

	mailbox, err := h.mailService.CreateMailbox(r.Context(), tenantUUID, req)
	if err != nil {
		h.respondError(w, nethttp.StatusInternalServerError, "Failed to create mailbox")
		return
	}

	h.respondJSON(w, nethttp.StatusCreated, toMailboxResponse(mailbox))
}

// UpdateMailbox aktualisiert ein bestehendes Postfach
func (h *MailHandler) UpdateMailbox(w nethttp.ResponseWriter, r *nethttp.Request) {
	idStr := r.PathValue("id")
	mailboxID, err := uuid.Parse(idStr)
	if err != nil {
		h.respondError(w, nethttp.StatusBadRequest, "Invalid mailbox ID")
		return
	}

	var req application.UpdateMailboxRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, nethttp.StatusBadRequest, "Invalid request body")
		return
	}

	mailbox, err := h.mailService.UpdateMailbox(r.Context(), mailboxID, req)
	if err != nil {
		h.respondError(w, nethttp.StatusInternalServerError, "Failed to update mailbox")
		return
	}

	h.respondJSON(w, nethttp.StatusOK, toMailboxResponse(mailbox))
}

// DeleteMailbox löscht ein Postfach
func (h *MailHandler) DeleteMailbox(w nethttp.ResponseWriter, r *nethttp.Request) {
	idStr := r.PathValue("id")
	mailboxID, err := uuid.Parse(idStr)
	if err != nil {
		h.respondError(w, nethttp.StatusBadRequest, "Invalid mailbox ID")
		return
	}

	err = h.mailService.DeleteMailbox(r.Context(), mailboxID)
	if err != nil {
		h.respondError(w, nethttp.StatusInternalServerError, "Failed to delete mailbox")
		return
	}

	w.WriteHeader(nethttp.StatusNoContent)
}

// --- E-Mails ---

// ListMails gibt E-Mails mit Filtern zurück
func (h *MailHandler) ListMails(w nethttp.ResponseWriter, r *nethttp.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, nethttp.StatusBadRequest, "X-Tenant-ID header required")
		return
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		h.respondError(w, nethttp.StatusBadRequest, "Invalid tenant ID")
		return
	}

	// Filter aus Query-Parametern
	filter := domain.MailFilter{
		Direction: r.URL.Query().Get("direction"),
		Search:    r.URL.Query().Get("search"),
	}

	if mailboxStr := r.URL.Query().Get("mailbox_id"); mailboxStr != "" {
		if id, err := uuid.Parse(mailboxStr); err == nil {
			filter.MailboxID = &id
		}
	}

	if projectStr := r.URL.Query().Get("project_id"); projectStr != "" {
		if id, err := uuid.Parse(projectStr); err == nil {
			filter.ProjectID = &id
		}
	}

	if readStr := r.URL.Query().Get("is_read"); readStr != "" {
		isRead := readStr == "true"
		filter.IsRead = &isRead
	}

	mails, err := h.mailService.ListMails(r.Context(), tenantUUID, filter)
	if err != nil {
		h.respondError(w, nethttp.StatusInternalServerError, "Failed to list mails")
		return
	}

	var response []application.MailResponse
	for _, m := range mails {
		response = append(response, toMailResponse(m))
	}

	h.respondJSON(w, nethttp.StatusOK, response)
}

// GetMail gibt eine einzelne E-Mail zurück
func (h *MailHandler) GetMail(w nethttp.ResponseWriter, r *nethttp.Request) {
	idStr := r.PathValue("id")
	mailID, err := uuid.Parse(idStr)
	if err != nil {
		h.respondError(w, nethttp.StatusBadRequest, "Invalid mail ID")
		return
	}

	mail, err := h.mailService.GetMail(r.Context(), mailID)
	if err != nil {
		h.respondError(w, nethttp.StatusNotFound, "Mail not found")
		return
	}

	h.respondJSON(w, nethttp.StatusOK, toMailResponse(mail))
}

// MarkMailAsRead markiert eine E-Mail als gelesen
func (h *MailHandler) MarkMailAsRead(w nethttp.ResponseWriter, r *nethttp.Request) {
	idStr := r.PathValue("id")
	mailID, err := uuid.Parse(idStr)
	if err != nil {
		h.respondError(w, nethttp.StatusBadRequest, "Invalid mail ID")
		return
	}

	err = h.mailService.MarkAsRead(r.Context(), mailID)
	if err != nil {
		h.respondError(w, nethttp.StatusInternalServerError, "Failed to mark mail as read")
		return
	}

	h.respondJSON(w, nethttp.StatusOK, map[string]string{"status": "marked as read"})
}

// AssignMailToProject weist eine E-Mail einem Projekt zu
func (h *MailHandler) AssignMailToProject(w nethttp.ResponseWriter, r *nethttp.Request) {
	idStr := r.PathValue("id")
	mailID, err := uuid.Parse(idStr)
	if err != nil {
		h.respondError(w, nethttp.StatusBadRequest, "Invalid mail ID")
		return
	}

	var req application.AssignProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, nethttp.StatusBadRequest, "Invalid request body")
		return
	}

	err = h.mailService.AssignToProject(r.Context(), mailID, req.ProjectID, req.Confidence)
	if err != nil {
		h.respondError(w, nethttp.StatusInternalServerError, "Failed to assign mail to project")
		return
	}

	h.respondJSON(w, nethttp.StatusOK, map[string]string{"status": "assigned to project"})
}

// SendMail sendet eine E-Mail über SMTP
func (h *MailHandler) SendMail(w nethttp.ResponseWriter, r *nethttp.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, nethttp.StatusBadRequest, "X-Tenant-ID header required")
		return
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		h.respondError(w, nethttp.StatusBadRequest, "Invalid tenant ID")
		return
	}

	var req application.SendMailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, nethttp.StatusBadRequest, "Invalid request body")
		return
	}

	cmd := domain.SendMailCmd{
		TenantID:  tenantUUID,
		MailboxID: req.MailboxID,
		To:        req.To,
		Subject:   req.Subject,
		BodyHTML:  req.BodyHTML,
		BodyText:  req.BodyText,
	}

	mail, err := h.mailService.SendMail(r.Context(), cmd)
	if err != nil {
		h.respondError(w, nethttp.StatusInternalServerError, "Failed to send mail")
		return
	}

	h.respondJSON(w, nethttp.StatusCreated, toMailResponse(mail))
}

// --- Hilfsfunktionen ---

func toMailboxResponse(mb *domain.Mailbox) application.MailboxResponse {
	return application.MailboxResponse{
		ID:         mb.ID,
		TenantID:   mb.TenantID,
		Type:       mb.Type,
		UserID:     mb.UserID,
		ImapHost:   mb.ImapHost,
		ImapPort:   mb.ImapPort,
		SmtpHost:   mb.SmtpHost,
		SmtpPort:   mb.SmtpPort,
		Username:   mb.Username,
		UseTLS:     mb.UseTLS,
		IsActive:   mb.IsActive,
		LastSyncAt: mb.LastSyncAt,
		CreatedAt:  mb.CreatedAt,
		UpdatedAt:  mb.UpdatedAt,
	}
}

func toMailResponse(m *domain.Mail) application.MailResponse {
	var attachments []application.MailAttachmentResponse
	for _, a := range m.Attachments {
		attachments = append(attachments, application.MailAttachmentResponse{
			ID:       a.ID,
			Filename: a.Filename,
			MimeType: a.MimeType,
			Size:     a.Size,
		})
	}

	return application.MailResponse{
		ID:          m.ID,
		TenantID:    m.TenantID,
		MailboxID:   m.MailboxID,
		MessageID:   m.MessageID,
		Direction:   m.Direction,
		From:        m.From,
		To:          m.To,
		Subject:     m.Subject,
		BodyHTML:    m.BodyHTML,
		BodyText:    m.BodyText,
		Date:        m.Date,
		IsRead:      m.IsRead,
		ProjectID:   m.ProjectID,
		Confidence:  m.Confidence,
		Attachments: attachments,
		CreatedAt:   m.CreatedAt,
	}
}

func (h *MailHandler) respondJSON(w nethttp.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func (h *MailHandler) respondError(w nethttp.ResponseWriter, statusCode int, message string) {
	h.respondJSON(w, statusCode, map[string]string{"error": message})
}

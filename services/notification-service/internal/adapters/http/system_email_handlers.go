package http

import (
	"encoding/json"
	nethttp "net/http"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/notification-service/internal/application"
)

// SystemEmailHandler handles system/transactional email endpoints
type SystemEmailHandler struct {
	emailService *application.SystemEmailService
	log          logger.Logger
}

// NewSystemEmailHandler creates a new SystemEmailHandler
func NewSystemEmailHandler(emailService *application.SystemEmailService, log logger.Logger) *SystemEmailHandler {
	return &SystemEmailHandler{
		emailService: emailService,
		log:          log,
	}
}

// --- Request DTOs ---

type sendEmailRequest struct {
	To       string `json:"to"`
	Subject  string `json:"subject"`
	BodyHTML string `json:"body_html"`
}

type sendBookingRequestEmail struct {
	To            string `json:"to"`
	RecipientName string `json:"recipient_name"`
	ProjectName   string `json:"project_name"`
	Dates         string `json:"dates"`
	Link          string `json:"link"`
	SenderName    string `json:"sender_name"`
	CompanyName   string `json:"company_name"`
}

type sendInvitationEmail struct {
	To             string `json:"to"`
	RecipientEmail string `json:"recipient_email"`
	InviterName    string `json:"inviter_name"`
	CompanyName    string `json:"company_name"`
	Link           string `json:"link"`
}

type sendPasswordResetEmail struct {
	To            string `json:"to"`
	RecipientName string `json:"recipient_name"`
	Link          string `json:"link"`
	CompanyName   string `json:"company_name"`
	ExpiresIn     string `json:"expires_in"`
}

// SendEmail handles POST /api/v1/notifications/send-email
// Sends a raw email with custom subject and HTML body
func (h *SystemEmailHandler) SendEmail(w nethttp.ResponseWriter, r *nethttp.Request) {
	var req sendEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, nethttp.StatusBadRequest, "Invalid request body")
		return
	}

	if req.To == "" || req.Subject == "" || req.BodyHTML == "" {
		h.respondError(w, nethttp.StatusBadRequest, "Fields 'to', 'subject', and 'body_html' are required")
		return
	}

	if !h.emailService.IsConfigured() {
		h.respondError(w, nethttp.StatusServiceUnavailable, "SMTP is not configured. Set SMTP_HOST, SMTP_PORT, SMTP_USER, SMTP_PASS environment variables.")
		return
	}

	if err := h.emailService.SendEmail(req.To, req.Subject, req.BodyHTML); err != nil {
		h.log.Error("Failed to send system email", err, "to", req.To)
		h.respondError(w, nethttp.StatusInternalServerError, "Failed to send email: "+err.Error())
		return
	}

	h.respondJSON(w, nethttp.StatusOK, map[string]string{"status": "sent", "to": req.To})
}

// SendBookingRequest handles POST /api/v1/notifications/send-email/booking-request
func (h *SystemEmailHandler) SendBookingRequest(w nethttp.ResponseWriter, r *nethttp.Request) {
	var req sendBookingRequestEmail
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, nethttp.StatusBadRequest, "Invalid request body")
		return
	}

	if req.To == "" || req.ProjectName == "" {
		h.respondError(w, nethttp.StatusBadRequest, "Fields 'to' and 'project_name' are required")
		return
	}

	if !h.emailService.IsConfigured() {
		h.respondError(w, nethttp.StatusServiceUnavailable, "SMTP is not configured")
		return
	}

	data := application.BookingRequestData{
		RecipientName: req.RecipientName,
		ProjectName:   req.ProjectName,
		Dates:         req.Dates,
		Link:          req.Link,
		SenderName:    req.SenderName,
		CompanyName:   req.CompanyName,
	}
	if data.CompanyName == "" {
		data.CompanyName = "RentFlow"
	}

	if err := h.emailService.SendBookingRequest(req.To, data); err != nil {
		h.log.Error("Failed to send booking request email", err, "to", req.To)
		h.respondError(w, nethttp.StatusInternalServerError, "Failed to send email: "+err.Error())
		return
	}

	h.respondJSON(w, nethttp.StatusOK, map[string]string{"status": "sent", "to": req.To})
}

// SendInvitation handles POST /api/v1/notifications/send-email/invitation
func (h *SystemEmailHandler) SendInvitation(w nethttp.ResponseWriter, r *nethttp.Request) {
	var req sendInvitationEmail
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, nethttp.StatusBadRequest, "Invalid request body")
		return
	}

	if req.To == "" || req.Link == "" {
		h.respondError(w, nethttp.StatusBadRequest, "Fields 'to' and 'link' are required")
		return
	}

	if !h.emailService.IsConfigured() {
		h.respondError(w, nethttp.StatusServiceUnavailable, "SMTP is not configured")
		return
	}

	data := application.InvitationData{
		RecipientEmail: req.To,
		InviterName:    req.InviterName,
		CompanyName:    req.CompanyName,
		Link:           req.Link,
	}
	if data.CompanyName == "" {
		data.CompanyName = "RentFlow"
	}
	if data.InviterName == "" {
		data.InviterName = "Ein Administrator"
	}

	if err := h.emailService.SendInvitation(req.To, data); err != nil {
		h.log.Error("Failed to send invitation email", err, "to", req.To)
		h.respondError(w, nethttp.StatusInternalServerError, "Failed to send email: "+err.Error())
		return
	}

	h.respondJSON(w, nethttp.StatusOK, map[string]string{"status": "sent", "to": req.To})
}

// SendPasswordReset handles POST /api/v1/notifications/send-email/password-reset
func (h *SystemEmailHandler) SendPasswordReset(w nethttp.ResponseWriter, r *nethttp.Request) {
	var req sendPasswordResetEmail
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, nethttp.StatusBadRequest, "Invalid request body")
		return
	}

	if req.To == "" || req.Link == "" {
		h.respondError(w, nethttp.StatusBadRequest, "Fields 'to' and 'link' are required")
		return
	}

	if !h.emailService.IsConfigured() {
		h.respondError(w, nethttp.StatusServiceUnavailable, "SMTP is not configured")
		return
	}

	data := application.PasswordResetData{
		RecipientName: req.RecipientName,
		Link:          req.Link,
		CompanyName:   req.CompanyName,
		ExpiresIn:     req.ExpiresIn,
	}
	if data.CompanyName == "" {
		data.CompanyName = "RentFlow"
	}
	if data.ExpiresIn == "" {
		data.ExpiresIn = "1 Stunde"
	}
	if data.RecipientName == "" {
		data.RecipientName = req.To
	}

	if err := h.emailService.SendPasswordReset(req.To, data); err != nil {
		h.log.Error("Failed to send password reset email", err, "to", req.To)
		h.respondError(w, nethttp.StatusInternalServerError, "Failed to send email: "+err.Error())
		return
	}

	h.respondJSON(w, nethttp.StatusOK, map[string]string{"status": "sent", "to": req.To})
}

// --- Helpers ---

func (h *SystemEmailHandler) respondJSON(w nethttp.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func (h *SystemEmailHandler) respondError(w nethttp.ResponseWriter, statusCode int, message string) {
	h.respondJSON(w, statusCode, map[string]string{"error": message})
}

package httphandler

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/pkg/common/errors"
	"github.com/jeckersberger/EquipFlow/pkg/common/middleware"
	"github.com/jeckersberger/EquipFlow/pkg/common/response"
	"github.com/jeckersberger/EquipFlow/services/notification/internal/application"
)

type EmailHandler struct {
	emailService *application.EmailService
	logger       zerolog.Logger
}

func NewEmailHandler(emailService *application.EmailService, logger zerolog.Logger) *EmailHandler {
	return &EmailHandler{
		emailService: emailService,
		logger:       logger.With().Str("handler", "email").Logger(),
	}
}

type SendEmailRequest struct {
	To          string `json:"to"`
	Subject     string `json:"subject"`
	HTMLBody    string `json:"html_body"`
	Attachments []struct {
		Filename string `json:"filename"`
		Data     string `json:"data"` // base64 encoded
		MimeType string `json:"mime_type"`
	} `json:"attachments,omitempty"`
}

func (h *EmailHandler) Send(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	var req SendEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiger Request"))
		return
	}

	if req.To == "" || req.Subject == "" {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Empfaenger und Betreff sind erforderlich"))
		return
	}

	sendReq := application.SendRequest{
		To:       req.To,
		Subject:  req.Subject,
		HTMLBody: req.HTMLBody,
	}

	if err := h.emailService.Send(sendReq); err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrInternal, err.Error()))
		return
	}

	response.Success(w, map[string]string{"status": "sent", "to": req.To})
}

// Status returns whether SMTP is configured.
func (h *EmailHandler) Status(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	response.Success(w, map[string]bool{"configured": h.emailService.IsConfigured()})
}

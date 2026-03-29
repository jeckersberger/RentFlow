package application

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"os"
	"strconv"

	"github.com/rs/zerolog"
	"gopkg.in/gomail.v2"
)

// EmailConfig holds SMTP connection settings.
type EmailConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	From     string
	FromName string
}

// LoadEmailConfig reads SMTP settings from environment variables.
func LoadEmailConfig() EmailConfig {
	port, _ := strconv.Atoi(getEnvOrDefault("SMTP_PORT", "587"))
	from := getEnvOrDefault("SMTP_FROM", getEnvOrDefault("SMTP_USER", ""))
	fromName := getEnvOrDefault("SMTP_FROM_NAME", "CrateDesk")

	return EmailConfig{
		Host:     getEnvOrDefault("SMTP_HOST", ""),
		Port:     port,
		User:     getEnvOrDefault("SMTP_USER", ""),
		Password: getEnvOrDefault("SMTP_PASSWORD", ""),
		From:     from,
		FromName: fromName,
	}
}

func getEnvOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// EmailService sends emails via SMTP.
type EmailService struct {
	config EmailConfig
	logger zerolog.Logger
}

func NewEmailService(config EmailConfig, logger zerolog.Logger) *EmailService {
	return &EmailService{
		config: config,
		logger: logger.With().Str("service", "email").Logger(),
	}
}

// IsConfigured returns true if SMTP settings are present.
func (s *EmailService) IsConfigured() bool {
	return s.config.Host != "" && s.config.User != ""
}

// SendRequest holds the data for sending an email.
type SendRequest struct {
	To          string
	Subject     string
	HTMLBody    string
	Attachments []EmailAttachment
}

// EmailAttachment represents a file attachment.
type EmailAttachment struct {
	Filename string
	Data     []byte
	MimeType string
}

// Send sends an email. Returns error if SMTP is not configured.
func (s *EmailService) Send(req SendRequest) error {
	if !s.IsConfigured() {
		return fmt.Errorf("SMTP nicht konfiguriert — bitte SMTP_HOST, SMTP_USER, SMTP_PASSWORD in den Einstellungen setzen")
	}

	m := gomail.NewMessage()
	m.SetAddressHeader("From", s.config.From, s.config.FromName)
	m.SetHeader("To", req.To)
	m.SetHeader("Subject", req.Subject)
	m.SetBody("text/html", req.HTMLBody)

	for _, att := range req.Attachments {
		data := att.Data
		m.Attach(att.Filename, gomail.SetCopyFunc(func(w io.Writer) error {
			_, err := w.Write(data)
			return err
		}))
	}

	d := gomail.NewDialer(s.config.Host, s.config.Port, s.config.User, s.config.Password)

	if err := d.DialAndSend(m); err != nil {
		s.logger.Error().Err(err).Str("to", req.To).Str("subject", req.Subject).Msg("email send failed")
		return fmt.Errorf("E-Mail-Versand fehlgeschlagen: %w", err)
	}

	s.logger.Info().Str("to", req.To).Str("subject", req.Subject).Msg("email sent successfully")
	return nil
}

// InvoiceEmailData holds data for rendering invoice email templates.
type InvoiceEmailData struct {
	CompanyName   string
	CustomerName  string
	InvoiceNumber string
	InvoiceDate   string
	DueDate       string
	TotalGross    string
	InvoiceType   string
}

// RenderInvoiceEmail renders the HTML body for an invoice email.
func RenderInvoiceEmail(data InvoiceEmailData) (string, error) {
	tmpl := `<!DOCTYPE html>
<html>
<head><meta charset="utf-8"></head>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif; max-width: 600px; margin: 0 auto; padding: 20px; color: #333;">
  <div style="border-bottom: 3px solid #005ab6; padding-bottom: 16px; margin-bottom: 24px;">
    <h2 style="margin: 0; color: #005ab6;">{{.CompanyName}}</h2>
  </div>

  <p>Sehr geehrte/r {{.CustomerName}},</p>

  {{if eq .InvoiceType "invoice"}}
  <p>anbei erhalten Sie unsere Rechnung <strong>{{.InvoiceNumber}}</strong> vom {{.InvoiceDate}}.</p>
  {{else if eq .InvoiceType "credit_note"}}
  <p>anbei erhalten Sie unsere Gutschrift <strong>{{.InvoiceNumber}}</strong> vom {{.InvoiceDate}}.</p>
  {{else}}
  <p>anbei erhalten Sie das Dokument <strong>{{.InvoiceNumber}}</strong> vom {{.InvoiceDate}}.</p>
  {{end}}

  <div style="background: #f8f9fa; border-radius: 8px; padding: 16px; margin: 20px 0;">
    <table style="width: 100%; border-collapse: collapse;">
      <tr>
        <td style="padding: 4px 0; color: #666;">Rechnungsnummer:</td>
        <td style="padding: 4px 0; font-weight: bold; text-align: right;">{{.InvoiceNumber}}</td>
      </tr>
      <tr>
        <td style="padding: 4px 0; color: #666;">Rechnungsdatum:</td>
        <td style="padding: 4px 0; text-align: right;">{{.InvoiceDate}}</td>
      </tr>
      <tr>
        <td style="padding: 4px 0; color: #666;">Faellig bis:</td>
        <td style="padding: 4px 0; text-align: right;">{{.DueDate}}</td>
      </tr>
      <tr style="border-top: 1px solid #ddd;">
        <td style="padding: 8px 0 4px; font-weight: bold;">Gesamtbetrag:</td>
        <td style="padding: 8px 0 4px; font-weight: bold; text-align: right; font-size: 1.2em; color: #005ab6;">{{.TotalGross}}</td>
      </tr>
    </table>
  </div>

  <p>Die Rechnung finden Sie als PDF im Anhang.</p>

  <p>Bei Fragen stehen wir Ihnen gerne zur Verfuegung.</p>

  <p>Mit freundlichen Gruessen<br>
  <strong>{{.CompanyName}}</strong></p>

  <div style="border-top: 1px solid #eee; padding-top: 12px; margin-top: 24px; font-size: 12px; color: #999;">
    Diese E-Mail wurde automatisch von CrateDesk generiert.
  </div>
</body>
</html>`

	t, err := template.New("invoice_email").Parse(tmpl)
	if err != nil {
		return "", fmt.Errorf("parse email template: %w", err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute email template: %w", err)
	}

	return buf.String(), nil
}

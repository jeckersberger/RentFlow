package email

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"mime/multipart"
	"net/smtp"
	"net/textproto"
	"strings"
)

// SMTPConfig enthält die SMTP-Konfiguration für den Email-Versand.
type SMTPConfig struct {
	Host     string // SMTP-Server (z.B. "smtp.gmail.com")
	Port     int    // SMTP-Port (587 für TLS, 465 für SSL)
	Username string // SMTP-Benutzername
	Password string // SMTP-Passwort
	FromName string // Absendername (z.B. "EquipFlow")
	FromAddr string // Absender-Email
}

// SMTPSender versendet E-Mails über SMTP mit optionalem PDF-Anhang.
type SMTPSender struct {
	config SMTPConfig
}

// NewSMTPSender erstellt einen neuen SMTP-basierten Email-Sender.
func NewSMTPSender(config SMTPConfig) *SMTPSender {
	return &SMTPSender{config: config}
}

// SendWithAttachment versendet eine E-Mail mit optionalem PDF-Anhang.
// Wird für Rechnungsversand, Angebotsversand und Mahnungen verwendet.
func (s *SMTPSender) SendWithAttachment(to, subject, bodyHTML string, attachment []byte, attachmentName string) error {
	if to == "" {
		return fmt.Errorf("Empfänger-E-Mail-Adresse fehlt")
	}
	if subject == "" {
		return fmt.Errorf("E-Mail-Betreff fehlt")
	}

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// E-Mail-Header setzen
	headers := make(map[string]string)
	headers["From"] = fmt.Sprintf("%s <%s>", s.config.FromName, s.config.FromAddr)
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = fmt.Sprintf("multipart/mixed; boundary=%s", writer.Boundary())

	var headerBuf bytes.Buffer
	for k, v := range headers {
		headerBuf.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	headerBuf.WriteString("\r\n")

	// HTML-Body als Teil der Multipart-Nachricht
	htmlHeader := textproto.MIMEHeader{}
	htmlHeader.Set("Content-Type", "text/html; charset=utf-8")
	htmlHeader.Set("Content-Transfer-Encoding", "quoted-printable")

	htmlPart, err := writer.CreatePart(htmlHeader)
	if err != nil {
		return fmt.Errorf("E-Mail HTML-Teil konnte nicht erstellt werden: %w", err)
	}
	htmlPart.Write([]byte(bodyHTML))

	// PDF-Anhang hinzufügen (falls vorhanden)
	if len(attachment) > 0 && attachmentName != "" {
		attachHeader := textproto.MIMEHeader{}
		attachHeader.Set("Content-Type", "application/pdf")
		attachHeader.Set("Content-Transfer-Encoding", "base64")
		attachHeader.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, attachmentName))

		attachPart, err := writer.CreatePart(attachHeader)
		if err != nil {
			return fmt.Errorf("PDF-Anhang konnte nicht erstellt werden: %w", err)
		}
		encoded := base64.StdEncoding.EncodeToString(attachment)
		// Base64 in 76-Zeichen-Zeilen aufteilen (RFC 2045)
		for i := 0; i < len(encoded); i += 76 {
			end := i + 76
			if end > len(encoded) {
				end = len(encoded)
			}
			attachPart.Write([]byte(encoded[i:end] + "\r\n"))
		}
	}

	writer.Close()

	// Vollständige E-Mail zusammensetzen
	message := append(headerBuf.Bytes(), buf.Bytes()...)

	// SMTP-Authentifizierung und Versand
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	auth := smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)

	recipients := strings.Split(to, ",")
	for i := range recipients {
		recipients[i] = strings.TrimSpace(recipients[i])
	}

	if err := smtp.SendMail(addr, auth, s.config.FromAddr, recipients, message); err != nil {
		return fmt.Errorf("E-Mail-Versand fehlgeschlagen: %w", err)
	}

	return nil
}

// Send versendet eine einfache E-Mail ohne Anhang.
func (s *SMTPSender) Send(to, subject, bodyHTML string) error {
	return s.SendWithAttachment(to, subject, bodyHTML, nil, "")
}

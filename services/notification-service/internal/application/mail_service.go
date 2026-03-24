package application

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/notification-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/notification-service/internal/ports"
)

// MailService — Anwendungslogik für das E-Mail-System
type MailService struct {
	mailRepo ports.MailRepository
	log      logger.Logger
}

func NewMailService(mailRepo ports.MailRepository, log logger.Logger) *MailService {
	return &MailService{
		mailRepo: mailRepo,
		log:      log,
	}
}

// --- Postfächer ---

// CreateMailbox erstellt ein neues E-Mail-Postfach
func (s *MailService) CreateMailbox(ctx context.Context, tenantID uuid.UUID, req CreateMailboxRequest) (*domain.Mailbox, error) {
	mb := &domain.Mailbox{
		ID:       uuid.New(),
		TenantID: tenantID,
		Type:     req.Type,
		UserID:   req.UserID,
		ImapHost: req.ImapHost,
		ImapPort: req.ImapPort,
		SmtpHost: req.SmtpHost,
		SmtpPort: req.SmtpPort,
		Username: req.Username,
		Password: req.Password,
		UseTLS:   req.UseTLS,
		IsActive: req.IsActive,
	}

	if mb.Type == "" {
		mb.Type = "general"
	}
	if mb.ImapPort == 0 {
		mb.ImapPort = 993
	}
	if mb.SmtpPort == 0 {
		mb.SmtpPort = 587
	}

	created, err := s.mailRepo.CreateMailbox(ctx, mb)
	if err != nil {
		s.log.Error("Failed to create mailbox", err)
		return nil, err
	}

	s.log.Info("Mailbox created", "mailboxID", created.ID, "type", created.Type)
	return created, nil
}

// ListMailboxes gibt alle Postfächer eines Mandanten zurück
func (s *MailService) ListMailboxes(ctx context.Context, tenantID uuid.UUID) ([]*domain.Mailbox, error) {
	return s.mailRepo.ListMailboxes(ctx, tenantID)
}

// GetMailbox gibt ein einzelnes Postfach zurück
func (s *MailService) GetMailbox(ctx context.Context, id uuid.UUID) (*domain.Mailbox, error) {
	return s.mailRepo.GetMailboxByID(ctx, id)
}

// UpdateMailbox aktualisiert ein Postfach
func (s *MailService) UpdateMailbox(ctx context.Context, id uuid.UUID, req UpdateMailboxRequest) (*domain.Mailbox, error) {
	existing, err := s.mailRepo.GetMailboxByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Type != "" {
		existing.Type = req.Type
	}
	if req.ImapHost != "" {
		existing.ImapHost = req.ImapHost
	}
	if req.ImapPort != 0 {
		existing.ImapPort = req.ImapPort
	}
	if req.SmtpHost != "" {
		existing.SmtpHost = req.SmtpHost
	}
	if req.SmtpPort != 0 {
		existing.SmtpPort = req.SmtpPort
	}
	if req.Username != "" {
		existing.Username = req.Username
	}
	if req.Password != "" {
		existing.Password = req.Password
	}
	existing.UseTLS = req.UseTLS
	existing.IsActive = req.IsActive
	existing.UserID = req.UserID

	updated, err := s.mailRepo.UpdateMailbox(ctx, existing)
	if err != nil {
		s.log.Error("Failed to update mailbox", err)
		return nil, err
	}

	s.log.Info("Mailbox updated", "mailboxID", updated.ID)
	return updated, nil
}

// DeleteMailbox löscht ein Postfach
func (s *MailService) DeleteMailbox(ctx context.Context, id uuid.UUID) error {
	err := s.mailRepo.DeleteMailbox(ctx, id)
	if err != nil {
		s.log.Error("Failed to delete mailbox", err)
		return err
	}

	s.log.Info("Mailbox deleted", "mailboxID", id)
	return nil
}

// --- E-Mails ---

// ListMails gibt E-Mails mit Filtern zurück
func (s *MailService) ListMails(ctx context.Context, tenantID uuid.UUID, filter domain.MailFilter) ([]*domain.Mail, error) {
	return s.mailRepo.ListMails(ctx, tenantID, filter)
}

// GetMail gibt eine einzelne E-Mail mit Anhängen zurück
func (s *MailService) GetMail(ctx context.Context, id uuid.UUID) (*domain.Mail, error) {
	return s.mailRepo.GetMailByID(ctx, id)
}

// MarkAsRead markiert eine E-Mail als gelesen
func (s *MailService) MarkAsRead(ctx context.Context, id uuid.UUID) error {
	err := s.mailRepo.MarkMailAsRead(ctx, id)
	if err != nil {
		s.log.Error("Failed to mark mail as read", err)
		return err
	}

	s.log.Info("Mail marked as read", "mailID", id)
	return nil
}

// AssignToProject weist eine E-Mail einem Projekt zu
func (s *MailService) AssignToProject(ctx context.Context, mailID, projectID uuid.UUID, confidence float64) error {
	err := s.mailRepo.AssignMailToProject(ctx, mailID, projectID, confidence)
	if err != nil {
		s.log.Error("Failed to assign mail to project", err)
		return err
	}

	s.log.Info("Mail assigned to project", "mailID", mailID, "projectID", projectID, "confidence", confidence)
	return nil
}

// SendMail sendet eine E-Mail über SMTP
func (s *MailService) SendMail(ctx context.Context, cmd domain.SendMailCmd) (*domain.Mail, error) {
	// Postfach laden für SMTP-Konfiguration
	mailbox, err := s.mailRepo.GetMailboxByID(ctx, cmd.MailboxID)
	if err != nil {
		s.log.Error("Failed to get mailbox for sending", err)
		return nil, fmt.Errorf("mailbox not found: %w", err)
	}

	if mailbox.SmtpHost == "" {
		return nil, fmt.Errorf("SMTP not configured for mailbox %s", mailbox.ID)
	}

	// E-Mail-Nachricht erstellen
	body := cmd.BodyText
	if body == "" {
		body = cmd.BodyHTML
	}

	contentType := "text/html"
	if cmd.BodyHTML == "" {
		contentType = "text/plain"
	}

	message := "From: " + mailbox.Username + "\r\n" +
		"To: " + cmd.To + "\r\n" +
		"Subject: " + cmd.Subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: " + contentType + "; charset=utf-8\r\n" +
		"Date: " + time.Now().Format(time.RFC1123Z) + "\r\n" +
		"\r\n" +
		body

	// Über SMTP senden (mit TLS-Unterstützung)
	addr := mailbox.SmtpHost + ":" + strconv.Itoa(mailbox.SmtpPort)
	msgBytes := []byte(message)

	if mailbox.UseTLS && mailbox.SmtpPort == 465 {
		// Implicit TLS (SMTPS)
		err = sendMailViaSMTPS(addr, mailbox.SmtpHost, mailbox.Username, mailbox.Password, cmd.To, msgBytes)
	} else {
		err = sendMailViaSTARTTLS(addr, mailbox.SmtpHost, mailbox.Username, mailbox.Password, mailbox.UseTLS, cmd.To, msgBytes)
	}
	if err != nil {
		s.log.Error("Failed to send mail via SMTP", err, "mailboxID", mailbox.ID)
		return nil, fmt.Errorf("failed to send mail: %w", err)
	}

	// Gesendete E-Mail in DB speichern
	mail := &domain.Mail{
		ID:        uuid.New(),
		TenantID:  cmd.TenantID,
		MailboxID: cmd.MailboxID,
		Direction: "outbound",
		From:      mailbox.Username,
		To:        cmd.To,
		Subject:   cmd.Subject,
		BodyHTML:  cmd.BodyHTML,
		BodyText:  cmd.BodyText,
		Date:      time.Now(),
		IsRead:    true,
	}

	created, err := s.mailRepo.CreateMail(ctx, mail)
	if err != nil {
		s.log.Error("Failed to store sent mail", err)
		// E-Mail wurde gesendet, aber DB-Speicherung fehlgeschlagen
		return mail, nil
	}

	s.log.Info("Mail sent and stored", "mailID", created.ID, "to", cmd.To)
	return created, nil
}

// --- SMTP helper functions ---

func sendMailViaSTARTTLS(addr, host, username, password string, useTLS bool, to string, msg []byte) error {
	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		return fmt.Errorf("dial failed: %w", err)
	}
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		conn.Close()
		return err
	}
	defer client.Close()

	if useTLS {
		if ok, _ := client.Extension("STARTTLS"); ok {
			if err := client.StartTLS(&tls.Config{ServerName: host}); err != nil {
				return fmt.Errorf("STARTTLS failed: %w", err)
			}
		}
	}

	if username != "" && password != "" {
		auth := smtp.PlainAuth("", username, password, host)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("auth failed: %w", err)
		}
	}

	if err := client.Mail(username); err != nil {
		return err
	}
	if err := client.Rcpt(to); err != nil {
		return err
	}
	w, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write(msg); err != nil {
		return err
	}
	w.Close()
	client.Quit()
	return nil
}

func sendMailViaSMTPS(addr, host, username, password, to string, msg []byte) error {
	conn, err := tls.DialWithDialer(&net.Dialer{Timeout: 10 * time.Second}, "tcp", addr, &tls.Config{ServerName: host})
	if err != nil {
		return fmt.Errorf("TLS dial failed: %w", err)
	}
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		conn.Close()
		return err
	}
	defer client.Close()

	if username != "" && password != "" {
		auth := smtp.PlainAuth("", username, password, host)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("auth failed: %w", err)
		}
	}

	if err := client.Mail(username); err != nil {
		return err
	}
	if err := client.Rcpt(to); err != nil {
		return err
	}
	w, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write(msg); err != nil {
		return err
	}
	w.Close()
	client.Quit()
	return nil
}

// FetchMails — Platzhalter für IMAP-Abruf (benötigt go-imap Bibliothek)
// Wird in einem späteren Schritt implementiert mit IMAP IDLE/Polling
func (s *MailService) FetchMails(ctx context.Context, mailboxID uuid.UUID) error {
	s.log.Info("FetchMails called — IMAP polling not yet implemented", "mailboxID", mailboxID)
	return fmt.Errorf("IMAP polling not yet implemented — requires go-imap library")
}

package application

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"os"
	"strconv"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
)

// SMTPConfig holds SMTP connection parameters
type SMTPConfig struct {
	Host       string
	Port       int
	Username   string
	Password   string
	FromName   string
	FromEmail  string
	TLSEnabled bool
}

// SystemEmailService sends transactional emails (invitations, password resets, etc.)
// using SMTP configuration from environment variables.
type SystemEmailService struct {
	config SMTPConfig
	log    logger.Logger
}

// NewSystemEmailService creates a new SystemEmailService from environment variables.
// Env vars: SMTP_HOST, SMTP_PORT, SMTP_USER, SMTP_PASS, SMTP_FROM_NAME, SMTP_FROM_EMAIL, SMTP_TLS
func NewSystemEmailService(log logger.Logger) *SystemEmailService {
	port, _ := strconv.Atoi(os.Getenv("SMTP_PORT"))
	if port == 0 {
		port = 587
	}

	tlsEnabled := true
	if tlsStr := os.Getenv("SMTP_TLS"); tlsStr == "false" || tlsStr == "0" {
		tlsEnabled = false
	}

	fromEmail := os.Getenv("SMTP_FROM_EMAIL")
	if fromEmail == "" {
		fromEmail = os.Getenv("SMTP_FROM")
	}

	// Support both SMTP_PASS and SMTP_PASSWORD env var names
	password := os.Getenv("SMTP_PASS")
	if password == "" {
		password = os.Getenv("SMTP_PASSWORD")
	}

	return &SystemEmailService{
		config: SMTPConfig{
			Host:       os.Getenv("SMTP_HOST"),
			Port:       port,
			Username:   os.Getenv("SMTP_USER"),
			Password:   password,
			FromName:   os.Getenv("SMTP_FROM_NAME"),
			FromEmail:  fromEmail,
			TLSEnabled: tlsEnabled,
		},
		log: log,
	}
}

// IsConfigured returns true if SMTP host is set
func (s *SystemEmailService) IsConfigured() bool {
	return s.config.Host != ""
}

// SendEmail sends an email via SMTP
func (s *SystemEmailService) SendEmail(to, subject, bodyHTML string) error {
	if !s.IsConfigured() {
		return fmt.Errorf("SMTP not configured (SMTP_HOST is empty)")
	}

	cfg := s.config

	// Build the From header
	from := cfg.FromEmail
	if cfg.FromName != "" {
		from = fmt.Sprintf("%s <%s>", cfg.FromName, cfg.FromEmail)
	}

	// Build RFC 2822 message
	msg := "From: " + from + "\r\n" +
		"To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=utf-8\r\n" +
		"Date: " + time.Now().Format(time.RFC1123Z) + "\r\n" +
		"\r\n" +
		bodyHTML

	addr := cfg.Host + ":" + strconv.Itoa(cfg.Port)

	// Port 465 uses implicit TLS (SMTPS)
	if cfg.TLSEnabled && cfg.Port == 465 {
		return s.sendViaSMTPS(addr, cfg, to, []byte(msg))
	}

	// Standard SMTP with optional STARTTLS
	return s.sendViaSTARTTLS(addr, cfg, to, []byte(msg))
}

// sendViaSTARTTLS sends email using STARTTLS (ports 25, 587)
func (s *SystemEmailService) sendViaSTARTTLS(addr string, cfg SMTPConfig, to string, msg []byte) error {
	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server %s: %w", addr, err)
	}

	client, err := smtp.NewClient(conn, cfg.Host)
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Close()

	// STARTTLS
	if cfg.TLSEnabled {
		tlsConfig := &tls.Config{ServerName: cfg.Host}
		if err := client.StartTLS(tlsConfig); err != nil {
			return fmt.Errorf("STARTTLS failed: %w", err)
		}
	}

	// Auth
	if cfg.Username != "" && cfg.Password != "" {
		auth := smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("SMTP auth failed: %w", err)
		}
	}

	// Send
	if err := client.Mail(cfg.FromEmail); err != nil {
		return fmt.Errorf("MAIL FROM failed: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("RCPT TO failed: %w", err)
	}
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("DATA failed: %w", err)
	}
	if _, err := w.Write(msg); err != nil {
		return fmt.Errorf("write message failed: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("close data writer failed: %w", err)
	}

	client.Quit()
	s.log.Info("System email sent", "to", to)
	return nil
}

// sendViaSMTPS sends email over implicit TLS (port 465)
func (s *SystemEmailService) sendViaSMTPS(addr string, cfg SMTPConfig, to string, msg []byte) error {
	tlsConfig := &tls.Config{ServerName: cfg.Host}
	conn, err := tls.DialWithDialer(&net.Dialer{Timeout: 10 * time.Second}, "tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("TLS dial to %s failed: %w", addr, err)
	}

	client, err := smtp.NewClient(conn, cfg.Host)
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Close()

	// Auth
	if cfg.Username != "" && cfg.Password != "" {
		auth := smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("SMTP auth failed: %w", err)
		}
	}

	// Send
	if err := client.Mail(cfg.FromEmail); err != nil {
		return fmt.Errorf("MAIL FROM failed: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("RCPT TO failed: %w", err)
	}
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("DATA failed: %w", err)
	}
	if _, err := w.Write(msg); err != nil {
		return fmt.Errorf("write message failed: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("close data writer failed: %w", err)
	}

	client.Quit()
	s.log.Info("System email sent via SMTPS", "to", to)
	return nil
}

// --- Convenience methods for templated emails ---

// SendBookingRequest sends a booking request email
func (s *SystemEmailService) SendBookingRequest(to string, data BookingRequestData) error {
	subject, body, err := RenderBookingRequestEmail(data)
	if err != nil {
		return fmt.Errorf("failed to render booking request template: %w", err)
	}
	return s.SendEmail(to, subject, body)
}

// SendInvitation sends an invitation email
func (s *SystemEmailService) SendInvitation(to string, data InvitationData) error {
	subject, body, err := RenderInvitationEmail(data)
	if err != nil {
		return fmt.Errorf("failed to render invitation template: %w", err)
	}
	return s.SendEmail(to, subject, body)
}

// SendPasswordReset sends a password reset email
func (s *SystemEmailService) SendPasswordReset(to string, data PasswordResetData) error {
	subject, body, err := RenderPasswordResetEmail(data)
	if err != nil {
		return fmt.Errorf("failed to render password reset template: %w", err)
	}
	return s.SendEmail(to, subject, body)
}

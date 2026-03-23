package application

import (
	"time"

	"github.com/google/uuid"
)

type NotificationResponse struct {
	ID           uuid.UUID              `json:"id"`
	TenantID     uuid.UUID              `json:"tenant_id"`
	UserID       uuid.UUID              `json:"user_id"`
	EventType    string                 `json:"event_type"`
	Title        string                 `json:"title"`
	Body         string                 `json:"body"`
	Data         map[string]interface{} `json:"data"`
	ChannelsSent []string               `json:"channels_sent"`
	Status       string                 `json:"status"`
	IsRead       bool                   `json:"is_read"`
	ReadAt       *time.Time             `json:"read_at"`
	CreatedAt    time.Time              `json:"created_at"`
}

type ChannelResponse struct {
	ID       uuid.UUID              `json:"id"`
	Type     string                 `json:"type"`
	IsActive bool                   `json:"is_active"`
	Config   map[string]interface{} `json:"config"`
}

type PreferenceResponse struct {
	ID              uuid.UUID  `json:"id"`
	EventType       string     `json:"event_type"`
	Channels        []string   `json:"channels"`
	IsEnabled       bool       `json:"is_enabled"`
	QuietHoursStart *time.Time `json:"quiet_hours_start"`
	QuietHoursEnd   *time.Time `json:"quiet_hours_end"`
	DigestMode      string     `json:"digest_mode"`
}

type SendNotificationRequest struct {
	UserID    uuid.UUID              `json:"user_id"`
	EventType string                 `json:"event_type"`
	Title     string                 `json:"title"`
	Body      string                 `json:"body"`
	Data      map[string]interface{} `json:"data"`
}

type BroadcastRequest struct {
	UserIDs   []uuid.UUID            `json:"user_ids"`
	EventType string                 `json:"event_type"`
	Title     string                 `json:"title"`
	Body      string                 `json:"body"`
	Data      map[string]interface{} `json:"data"`
}

type PreferenceRequest struct {
	EventType       string     `json:"event_type"`
	Channels        []string   `json:"channels"`
	IsEnabled       bool       `json:"is_enabled"`
	QuietHoursStart *time.Time `json:"quiet_hours_start"`
	QuietHoursEnd   *time.Time `json:"quiet_hours_end"`
	DigestMode      string     `json:"digest_mode"`
}

type RegisterChannelRequest struct {
	Type   string                 `json:"type"`
	Config map[string]interface{} `json:"config"`
}

// --- Mail DTOs ---

// CreateMailboxRequest — Anfrage zum Erstellen eines Postfachs
type CreateMailboxRequest struct {
	Type     string     `json:"type"`      // "general", "invoices", "personal"
	UserID   *uuid.UUID `json:"user_id"`   // Für persönliche Postfächer
	ImapHost string     `json:"imap_host"` // IMAP-Server
	ImapPort int        `json:"imap_port"` // IMAP-Port
	SmtpHost string     `json:"smtp_host"` // SMTP-Server
	SmtpPort int        `json:"smtp_port"` // SMTP-Port
	Username string     `json:"username"`  // Benutzername
	Password string     `json:"password"`  // Passwort
	UseTLS   bool       `json:"use_tls"`   // TLS verwenden
	IsActive bool       `json:"is_active"` // Aktiv-Status
}

// UpdateMailboxRequest — Anfrage zum Aktualisieren eines Postfachs
type UpdateMailboxRequest struct {
	Type     string     `json:"type"`
	UserID   *uuid.UUID `json:"user_id"`
	ImapHost string     `json:"imap_host"`
	ImapPort int        `json:"imap_port"`
	SmtpHost string     `json:"smtp_host"`
	SmtpPort int        `json:"smtp_port"`
	Username string     `json:"username"`
	Password string     `json:"password"`
	UseTLS   bool       `json:"use_tls"`
	IsActive bool       `json:"is_active"`
}

// MailboxResponse — Antwort für ein Postfach (ohne Passwort)
type MailboxResponse struct {
	ID         uuid.UUID  `json:"id"`
	TenantID   uuid.UUID  `json:"tenant_id"`
	Type       string     `json:"type"`
	UserID     *uuid.UUID `json:"user_id,omitempty"`
	ImapHost   string     `json:"imap_host"`
	ImapPort   int        `json:"imap_port"`
	SmtpHost   string     `json:"smtp_host"`
	SmtpPort   int        `json:"smtp_port"`
	Username   string     `json:"username"`
	UseTLS     bool       `json:"use_tls"`
	IsActive   bool       `json:"is_active"`
	LastSyncAt *time.Time `json:"last_sync_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// MailResponse — Antwort für eine E-Mail
type MailResponse struct {
	ID          uuid.UUID              `json:"id"`
	TenantID    uuid.UUID              `json:"tenant_id"`
	MailboxID   uuid.UUID              `json:"mailbox_id"`
	MessageID   string                 `json:"message_id"`
	Direction   string                 `json:"direction"`
	From        string                 `json:"from"`
	To          string                 `json:"to"`
	Subject     string                 `json:"subject"`
	BodyHTML    string                 `json:"body_html"`
	BodyText    string                 `json:"body_text"`
	Date        time.Time              `json:"date"`
	IsRead      bool                   `json:"is_read"`
	ProjectID   *uuid.UUID             `json:"project_id,omitempty"`
	Confidence  float64                `json:"ai_confidence"`
	Attachments []MailAttachmentResponse `json:"attachments"`
	CreatedAt   time.Time              `json:"created_at"`
}

// MailAttachmentResponse — Antwort für einen E-Mail-Anhang
type MailAttachmentResponse struct {
	ID       uuid.UUID `json:"id"`
	Filename string    `json:"filename"`
	MimeType string    `json:"mime_type"`
	Size     int64     `json:"size_bytes"`
}

// SendMailRequest — Anfrage zum Senden einer E-Mail
type SendMailRequest struct {
	MailboxID uuid.UUID `json:"mailbox_id"`
	To        string    `json:"to"`
	Subject   string    `json:"subject"`
	BodyHTML  string    `json:"body_html"`
	BodyText  string    `json:"body_text"`
}

// AssignProjectRequest — Anfrage zum Zuweisen eines Projekts
type AssignProjectRequest struct {
	ProjectID  uuid.UUID `json:"project_id"`
	Confidence float64   `json:"confidence"`
}

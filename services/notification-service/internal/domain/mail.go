package domain

import (
	"time"

	"github.com/google/uuid"
)

// Mail — E-Mail-Nachricht (eingehend oder ausgehend)
type Mail struct {
	ID          uuid.UUID        // Eindeutige ID
	TenantID    uuid.UUID        // Mandanten-ID
	MailboxID   uuid.UUID        // Zugehöriges Postfach
	MessageID   string           // IMAP Message-ID
	Direction   string           // "inbound" oder "outbound"
	From        string           // Absender-Adresse
	To          string           // Empfänger-Adresse
	Subject     string           // Betreff
	BodyHTML    string           // HTML-Inhalt
	BodyText    string           // Text-Inhalt
	Date        time.Time        // E-Mail-Datum
	IsRead      bool             // Gelesen-Status
	ProjectID   *uuid.UUID       // KI-zugewiesenes Projekt (nil wenn nicht zugewiesen)
	Confidence  float64          // KI-Zuweisungs-Konfidenz (0-1)
	Attachments []MailAttachment // Anhänge
	CreatedAt   time.Time        // Erstellungszeitpunkt
}

// MailAttachment — E-Mail-Anhang
type MailAttachment struct {
	ID          uuid.UUID // Eindeutige ID
	MailID      uuid.UUID // Zugehörige E-Mail
	Filename    string    // Dateiname
	MimeType    string    // MIME-Typ
	Size        int64     // Dateigröße in Bytes
	StoragePath string    // Speicherpfad
	CreatedAt   time.Time // Erstellungszeitpunkt
}

// Mailbox — Konfiguriertes E-Mail-Postfach
type Mailbox struct {
	ID         uuid.UUID  // Eindeutige ID
	TenantID   uuid.UUID  // Mandanten-ID
	Type       string     // "general", "invoices", "personal"
	UserID     *uuid.UUID // Benutzer-ID (für persönliche Postfächer)
	ImapHost   string     // IMAP-Server
	ImapPort   int        // IMAP-Port
	SmtpHost   string     // SMTP-Server
	SmtpPort   int        // SMTP-Port
	Username   string     // Benutzername
	Password   string     // Passwort (verschlüsselt)
	UseTLS     bool       // TLS verwenden
	IsActive   bool       // Aktiv-Status
	LastSyncAt *time.Time // Letzter Synchronisationszeitpunkt
	CreatedAt  time.Time  // Erstellungszeitpunkt
	UpdatedAt  time.Time  // Aktualisierungszeitpunkt
}

// MailFilter — Filter für E-Mail-Abfragen
type MailFilter struct {
	MailboxID *uuid.UUID // Nach Postfach filtern
	ProjectID *uuid.UUID // Nach Projekt filtern
	IsRead    *bool      // Nach Gelesen-Status filtern
	Direction string     // Nach Richtung filtern
	Search    string     // Volltextsuche in Betreff/Absender/Inhalt
}

// AssignMailToProjectCmd — Befehl zum Zuweisen einer E-Mail an ein Projekt
type AssignMailToProjectCmd struct {
	MailID     uuid.UUID
	ProjectID  uuid.UUID
	Confidence float64
}

// SendMailCmd — Befehl zum Senden einer E-Mail
type SendMailCmd struct {
	TenantID  uuid.UUID
	MailboxID uuid.UUID
	To        string
	Subject   string
	BodyHTML  string
	BodyText  string
}

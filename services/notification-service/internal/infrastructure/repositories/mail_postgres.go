package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/notification-service/internal/domain"
)

// PostgresMailRepository — PostgreSQL-Implementierung für E-Mail-Zugriff
type PostgresMailRepository struct {
	db  *sql.DB
	log logger.Logger
}

func NewPostgresMailRepository(db *sql.DB, log logger.Logger) *PostgresMailRepository {
	return &PostgresMailRepository{db: db, log: log}
}

// --- E-Mails ---

func (r *PostgresMailRepository) CreateMail(ctx context.Context, m *domain.Mail) (*domain.Mail, error) {
	query := `INSERT INTO mails (id, tenant_id, mailbox_id, message_id, direction, from_address, to_address, subject, body_html, body_text, mail_date, is_read, project_id, ai_confidence, created_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, NOW())
	RETURNING id, created_at`

	err := r.db.QueryRowContext(ctx, query,
		m.ID, m.TenantID, m.MailboxID, m.MessageID, m.Direction, m.From, m.To,
		m.Subject, m.BodyHTML, m.BodyText, m.Date, m.IsRead, m.ProjectID, m.Confidence,
	).Scan(&m.ID, &m.CreatedAt)

	if err != nil {
		r.log.Error("Failed to create mail", err)
		return nil, domain.ErrDatabaseError
	}

	return m, nil
}

func (r *PostgresMailRepository) GetMailByID(ctx context.Context, id uuid.UUID) (*domain.Mail, error) {
	query := `SELECT id, tenant_id, mailbox_id, message_id, direction, from_address, to_address, subject, body_html, body_text, mail_date, is_read, project_id, ai_confidence, created_at
	FROM mails WHERE id = $1`

	m := &domain.Mail{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&m.ID, &m.TenantID, &m.MailboxID, &m.MessageID, &m.Direction, &m.From, &m.To,
		&m.Subject, &m.BodyHTML, &m.BodyText, &m.Date, &m.IsRead, &m.ProjectID, &m.Confidence, &m.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.ErrMailNotFound
	}
	if err != nil {
		r.log.Error("Failed to get mail by ID", err)
		return nil, domain.ErrDatabaseError
	}

	// Anhänge laden
	attachments, err := r.ListAttachments(ctx, m.ID)
	if err == nil {
		m.Attachments = attachments
	}

	return m, nil
}

func (r *PostgresMailRepository) ListMails(ctx context.Context, tenantID uuid.UUID, filter domain.MailFilter) ([]*domain.Mail, error) {
	query := `SELECT id, tenant_id, mailbox_id, message_id, direction, from_address, to_address, subject, body_html, body_text, mail_date, is_read, project_id, ai_confidence, created_at
	FROM mails WHERE tenant_id = $1`

	args := []interface{}{tenantID}
	argIdx := 2

	if filter.MailboxID != nil {
		query += fmt.Sprintf(" AND mailbox_id = $%d", argIdx)
		args = append(args, *filter.MailboxID)
		argIdx++
	}

	if filter.ProjectID != nil {
		query += fmt.Sprintf(" AND project_id = $%d", argIdx)
		args = append(args, *filter.ProjectID)
		argIdx++
	}

	if filter.IsRead != nil {
		query += fmt.Sprintf(" AND is_read = $%d", argIdx)
		args = append(args, *filter.IsRead)
		argIdx++
	}

	if filter.Direction != "" {
		query += fmt.Sprintf(" AND direction = $%d", argIdx)
		args = append(args, filter.Direction)
		argIdx++
	}

	if filter.Search != "" {
		searchPattern := "%" + strings.ToLower(filter.Search) + "%"
		query += fmt.Sprintf(" AND (LOWER(subject) LIKE $%d OR LOWER(from_address) LIKE $%d OR LOWER(body_text) LIKE $%d)", argIdx, argIdx, argIdx)
		args = append(args, searchPattern)
		argIdx++
	}

	query += " ORDER BY mail_date DESC LIMIT 100"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		r.log.Error("Failed to list mails", err)
		return nil, domain.ErrDatabaseError
	}
	defer rows.Close()

	var mails []*domain.Mail
	for rows.Next() {
		m := &domain.Mail{}
		err := rows.Scan(
			&m.ID, &m.TenantID, &m.MailboxID, &m.MessageID, &m.Direction, &m.From, &m.To,
			&m.Subject, &m.BodyHTML, &m.BodyText, &m.Date, &m.IsRead, &m.ProjectID, &m.Confidence, &m.CreatedAt,
		)
		if err != nil {
			r.log.Error("Failed to scan mail", err)
			continue
		}
		mails = append(mails, m)
	}

	return mails, nil
}

func (r *PostgresMailRepository) MarkMailAsRead(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE mails SET is_read = true WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.log.Error("Failed to mark mail as read", err)
		return domain.ErrDatabaseError
	}
	return nil
}

func (r *PostgresMailRepository) AssignMailToProject(ctx context.Context, mailID, projectID uuid.UUID, confidence float64) error {
	query := `UPDATE mails SET project_id = $1, ai_confidence = $2 WHERE id = $3`
	_, err := r.db.ExecContext(ctx, query, projectID, confidence, mailID)
	if err != nil {
		r.log.Error("Failed to assign mail to project", err)
		return domain.ErrDatabaseError
	}
	return nil
}

// --- Postfächer ---

func (r *PostgresMailRepository) CreateMailbox(ctx context.Context, mb *domain.Mailbox) (*domain.Mailbox, error) {
	query := `INSERT INTO mailboxes (id, tenant_id, type, user_id, imap_host, imap_port, smtp_host, smtp_port, username, password, use_tls, is_active, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NOW(), NOW())
	RETURNING id, created_at, updated_at`

	err := r.db.QueryRowContext(ctx, query,
		mb.ID, mb.TenantID, mb.Type, mb.UserID, mb.ImapHost, mb.ImapPort,
		mb.SmtpHost, mb.SmtpPort, mb.Username, mb.Password, mb.UseTLS, mb.IsActive,
	).Scan(&mb.ID, &mb.CreatedAt, &mb.UpdatedAt)

	if err != nil {
		r.log.Error("Failed to create mailbox", err)
		return nil, domain.ErrDatabaseError
	}

	return mb, nil
}

func (r *PostgresMailRepository) GetMailboxByID(ctx context.Context, id uuid.UUID) (*domain.Mailbox, error) {
	query := `SELECT id, tenant_id, type, user_id, imap_host, imap_port, smtp_host, smtp_port, username, password, use_tls, is_active, last_sync_at, created_at, updated_at
	FROM mailboxes WHERE id = $1`

	mb := &domain.Mailbox{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&mb.ID, &mb.TenantID, &mb.Type, &mb.UserID, &mb.ImapHost, &mb.ImapPort,
		&mb.SmtpHost, &mb.SmtpPort, &mb.Username, &mb.Password, &mb.UseTLS, &mb.IsActive,
		&mb.LastSyncAt, &mb.CreatedAt, &mb.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.ErrMailboxNotFound
	}
	if err != nil {
		r.log.Error("Failed to get mailbox by ID", err)
		return nil, domain.ErrDatabaseError
	}

	return mb, nil
}

func (r *PostgresMailRepository) ListMailboxes(ctx context.Context, tenantID uuid.UUID) ([]*domain.Mailbox, error) {
	query := `SELECT id, tenant_id, type, user_id, imap_host, imap_port, smtp_host, smtp_port, username, password, use_tls, is_active, last_sync_at, created_at, updated_at
	FROM mailboxes WHERE tenant_id = $1 ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		r.log.Error("Failed to list mailboxes", err)
		return nil, domain.ErrDatabaseError
	}
	defer rows.Close()

	var mailboxes []*domain.Mailbox
	for rows.Next() {
		mb := &domain.Mailbox{}
		err := rows.Scan(
			&mb.ID, &mb.TenantID, &mb.Type, &mb.UserID, &mb.ImapHost, &mb.ImapPort,
			&mb.SmtpHost, &mb.SmtpPort, &mb.Username, &mb.Password, &mb.UseTLS, &mb.IsActive,
			&mb.LastSyncAt, &mb.CreatedAt, &mb.UpdatedAt,
		)
		if err != nil {
			r.log.Error("Failed to scan mailbox", err)
			continue
		}
		mailboxes = append(mailboxes, mb)
	}

	return mailboxes, nil
}

func (r *PostgresMailRepository) UpdateMailbox(ctx context.Context, mb *domain.Mailbox) (*domain.Mailbox, error) {
	mb.UpdatedAt = time.Now()
	query := `UPDATE mailboxes SET type = $1, user_id = $2, imap_host = $3, imap_port = $4, smtp_host = $5, smtp_port = $6, username = $7, password = $8, use_tls = $9, is_active = $10, updated_at = NOW()
	WHERE id = $11
	RETURNING updated_at`

	err := r.db.QueryRowContext(ctx, query,
		mb.Type, mb.UserID, mb.ImapHost, mb.ImapPort, mb.SmtpHost, mb.SmtpPort,
		mb.Username, mb.Password, mb.UseTLS, mb.IsActive, mb.ID,
	).Scan(&mb.UpdatedAt)

	if err != nil {
		r.log.Error("Failed to update mailbox", err)
		return nil, domain.ErrDatabaseError
	}

	return mb, nil
}

func (r *PostgresMailRepository) DeleteMailbox(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM mailboxes WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.log.Error("Failed to delete mailbox", err)
		return domain.ErrDatabaseError
	}
	return nil
}

// --- Anhänge ---

func (r *PostgresMailRepository) CreateAttachment(ctx context.Context, a *domain.MailAttachment) (*domain.MailAttachment, error) {
	query := `INSERT INTO mail_attachments (id, mail_id, filename, mime_type, size_bytes, storage_path, created_at)
	VALUES ($1, $2, $3, $4, $5, $6, NOW())
	RETURNING id, created_at`

	err := r.db.QueryRowContext(ctx, query,
		a.ID, a.MailID, a.Filename, a.MimeType, a.Size, a.StoragePath,
	).Scan(&a.ID, &a.CreatedAt)

	if err != nil {
		r.log.Error("Failed to create attachment", err)
		return nil, domain.ErrDatabaseError
	}

	return a, nil
}

func (r *PostgresMailRepository) ListAttachments(ctx context.Context, mailID uuid.UUID) ([]domain.MailAttachment, error) {
	query := `SELECT id, mail_id, filename, mime_type, size_bytes, storage_path, created_at
	FROM mail_attachments WHERE mail_id = $1 ORDER BY filename`

	rows, err := r.db.QueryContext(ctx, query, mailID)
	if err != nil {
		r.log.Error("Failed to list attachments", err)
		return nil, domain.ErrDatabaseError
	}
	defer rows.Close()

	var attachments []domain.MailAttachment
	for rows.Next() {
		a := domain.MailAttachment{}
		err := rows.Scan(&a.ID, &a.MailID, &a.Filename, &a.MimeType, &a.Size, &a.StoragePath, &a.CreatedAt)
		if err != nil {
			r.log.Error("Failed to scan attachment", err)
			continue
		}
		attachments = append(attachments, a)
	}

	return attachments, nil
}

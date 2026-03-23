-- Mail-System: Postfächer, E-Mails und Anhänge

CREATE TABLE IF NOT EXISTS mailboxes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    type VARCHAR(20) NOT NULL DEFAULT 'general',
    user_id UUID,
    imap_host VARCHAR(255) NOT NULL DEFAULT '',
    imap_port INT NOT NULL DEFAULT 993,
    smtp_host VARCHAR(255) NOT NULL DEFAULT '',
    smtp_port INT NOT NULL DEFAULT 587,
    username VARCHAR(255) NOT NULL DEFAULT '',
    password TEXT NOT NULL DEFAULT '',
    use_tls BOOLEAN NOT NULL DEFAULT TRUE,
    is_active BOOLEAN NOT NULL DEFAULT FALSE,
    last_sync_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS mails (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    mailbox_id UUID REFERENCES mailboxes(id),
    message_id VARCHAR(255) NOT NULL DEFAULT '',
    direction VARCHAR(10) NOT NULL DEFAULT 'inbound',
    from_address VARCHAR(255) NOT NULL DEFAULT '',
    to_address VARCHAR(255) NOT NULL DEFAULT '',
    subject TEXT NOT NULL DEFAULT '',
    body_html TEXT NOT NULL DEFAULT '',
    body_text TEXT NOT NULL DEFAULT '',
    mail_date TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    is_read BOOLEAN NOT NULL DEFAULT FALSE,
    project_id UUID,
    ai_confidence DECIMAL(3,2) DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS mail_attachments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    mail_id UUID NOT NULL REFERENCES mails(id),
    filename VARCHAR(255) NOT NULL,
    mime_type VARCHAR(100) NOT NULL DEFAULT '',
    size_bytes BIGINT NOT NULL DEFAULT 0,
    storage_path TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_mails_tenant ON mails(tenant_id);
CREATE INDEX IF NOT EXISTS idx_mails_mailbox ON mails(mailbox_id);
CREATE INDEX IF NOT EXISTS idx_mails_project ON mails(project_id);
CREATE INDEX IF NOT EXISTS idx_mails_date ON mails(mail_date DESC);
CREATE INDEX IF NOT EXISTS idx_mailboxes_tenant ON mailboxes(tenant_id);

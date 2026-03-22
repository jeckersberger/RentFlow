-- Notification channels table
CREATE TABLE notification_channels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    user_id UUID NOT NULL,
    channel_type VARCHAR(20) NOT NULL,
    config JSONB DEFAULT '{}',
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_notification_channels_user ON notification_channels(tenant_id, user_id);
CREATE INDEX idx_notification_channels_type ON notification_channels(channel_type, is_active);

-- User preferences table
CREATE TABLE user_preferences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    user_id UUID NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    channels TEXT[] DEFAULT '{}',
    is_enabled BOOLEAN DEFAULT true,
    quiet_hours_start TIME,
    quiet_hours_end TIME,
    digest_mode VARCHAR(20) DEFAULT 'immediate',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(tenant_id, user_id, event_type)
);

CREATE INDEX idx_user_preferences_user ON user_preferences(tenant_id, user_id);
CREATE INDEX idx_user_preferences_event ON user_preferences(tenant_id, event_type);

-- Notifications table
CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    user_id UUID NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    title VARCHAR(500) NOT NULL,
    body TEXT,
    data JSONB DEFAULT '{}',
    channels_sent TEXT[] DEFAULT '{}',
    status VARCHAR(20) DEFAULT 'pending',
    is_read BOOLEAN DEFAULT false,
    read_at TIMESTAMPTZ,
    sent_at TIMESTAMPTZ,
    scheduled_for TIMESTAMPTZ,
    batch_id UUID,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_notifications_user ON notifications(tenant_id, user_id, is_read);
CREATE INDEX idx_notifications_status ON notifications(status, scheduled_for);
CREATE INDEX idx_notifications_batch ON notifications(batch_id);

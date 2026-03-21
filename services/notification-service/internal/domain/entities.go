package domain

import "time"

type NotificationType string

const (
	NotificationTypeInfo    NotificationType = "info"
	NotificationTypeWarning NotificationType = "warning"
	NotificationTypeSuccess NotificationType = "success"
	NotificationTypeError   NotificationType = "error"
)

type Channel string

const (
	ChannelInApp Channel = "in_app"
	ChannelEmail Channel = "email"
	ChannelPush  Channel = "push"
)

type Notification struct {
	ID        string           `json:"id"`
	TenantID  string           `json:"tenant_id"`
	UserID    string           `json:"user_id"`
	Type      NotificationType `json:"type"`
	Channel   Channel          `json:"channel"`
	Title     string           `json:"title"`
	Message   string           `json:"message"`
	Link      string           `json:"link,omitempty"`
	Read      bool             `json:"read"`
	SentAt    time.Time        `json:"sent_at"`
	CreatedAt time.Time        `json:"created_at"`
}

type NotificationPreference struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id"`
	UserID    string    `json:"user_id"`
	Channel   Channel   `json:"channel"`
	EventType string    `json:"event_type"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

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

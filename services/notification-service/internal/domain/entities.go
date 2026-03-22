package domain

import (
	"time"

	"github.com/google/uuid"
)

type Notification struct {
	ID           uuid.UUID
	TenantID     uuid.UUID
	UserID       uuid.UUID
	EventType    string
	Title        string
	Body         string
	Data         map[string]interface{}
	ChannelsSent []string
	Status       string
	IsRead       bool
	ReadAt       *time.Time
	SentAt       *time.Time
	ScheduledFor *time.Time
	BatchID      *uuid.UUID
	CreatedAt    time.Time
}

type NotificationChannel struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	UserID    uuid.UUID
	Type      string
	Config    map[string]interface{}
	IsActive  bool
	CreatedAt time.Time
}

type UserPreference struct {
	ID              uuid.UUID
	TenantID        uuid.UUID
	UserID          uuid.UUID
	EventType       string
	Channels        []string
	IsEnabled       bool
	QuietHoursStart *time.Time
	QuietHoursEnd   *time.Time
	DigestMode      string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type SendNotificationCmd struct {
	TenantID  uuid.UUID
	UserID    uuid.UUID
	EventType string
	Title     string
	Body      string
	Data      map[string]interface{}
}

type BroadcastNotificationCmd struct {
	TenantID  uuid.UUID
	UserIDs   []uuid.UUID
	EventType string
	Title     string
	Body      string
	Data      map[string]interface{}
}

type UpdatePreferenceCmd struct {
	TenantID        uuid.UUID
	UserID          uuid.UUID
	EventType       string
	Channels        []string
	IsEnabled       bool
	QuietHoursStart *time.Time
	QuietHoursEnd   *time.Time
	DigestMode      string
}

type RegisterChannelCmd struct {
	TenantID  uuid.UUID
	UserID    uuid.UUID
	Type      string
	Config    map[string]interface{}
}

type GetDashboardStatsResponse struct {
	TodayNotificationCount int
	DeliveryRate           float64
	TopEventTypes          []string
	PendingNotifications   int
}

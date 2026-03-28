package domain

import (
	"time"

	"github.com/google/uuid"
)

type Notification struct {
	ID            uuid.UUID  `json:"id"`
	TenantID      uuid.UUID  `json:"tenant_id"`
	UserID        uuid.UUID  `json:"user_id"`
	Type          string     `json:"type"`
	Title         string     `json:"title"`
	Message       string     `json:"message,omitempty"`
	ReferenceID   *uuid.UUID `json:"reference_id,omitempty"`
	ReferenceType string     `json:"reference_type,omitempty"`
	IsRead        bool       `json:"is_read"`
	ReadAt        *time.Time `json:"read_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

type NotificationPreference struct {
	ID        uuid.UUID `json:"id"`
	TenantID  uuid.UUID `json:"tenant_id"`
	UserID    uuid.UUID `json:"user_id"`
	Channel   string    `json:"channel"`
	Type      string    `json:"type"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type NotificationFilter struct {
	Page    int
	PerPage int
	IsRead  *bool
}

type UnreadCount struct {
	Count int64 `json:"count"`
}

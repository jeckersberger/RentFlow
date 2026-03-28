package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type AuditLog struct {
	ID         uuid.UUID        `json:"id"`
	TenantID   uuid.UUID        `json:"tenant_id"`
	UserID     *uuid.UUID       `json:"user_id,omitempty"`
	Action     string           `json:"action"`
	EntityType string           `json:"entity_type"`
	EntityID   *uuid.UUID       `json:"entity_id,omitempty"`
	OldData    json.RawMessage  `json:"old_data,omitempty"`
	NewData    json.RawMessage  `json:"new_data,omitempty"`
	IPAddress  string           `json:"ip_address,omitempty"`
	UserAgent  string           `json:"user_agent,omitempty"`
	CreatedAt  time.Time        `json:"created_at"`
}

type AuditPolicy struct {
	ID            uuid.UUID `json:"id"`
	TenantID      uuid.UUID `json:"tenant_id"`
	EntityType    string    `json:"entity_type"`
	RetentionDays int       `json:"retention_days"`
	LogReads      bool      `json:"log_reads"`
	LogWrites     bool      `json:"log_writes"`
	IsActive      bool      `json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type AuditLogFilter struct {
	Page       int
	PerPage    int
	EntityType string
	EntityID   *uuid.UUID
	UserID     *uuid.UUID
}

type AuditPolicyFilter struct {
	Page    int
	PerPage int
}

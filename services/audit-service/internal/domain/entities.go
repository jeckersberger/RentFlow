package domain

import "time"

type AuditEntry struct {
	ID             string                 `json:"id"`
	TenantID       string                 `json:"tenant_id"`
	Timestamp      time.Time              `json:"timestamp"`
	UserID         string                 `json:"user_id"`
	Action         string                 `json:"action"`
	EntityType     string                 `json:"entity_type"`
	EntityID       string                 `json:"entity_id"`
	PreviousState  map[string]interface{} `json:"previous_state,omitempty"`
	NewState       map[string]interface{} `json:"new_state,omitempty"`
	IPAddress      string                 `json:"ip_address"`
	UserAgent      string                 `json:"user_agent"`
	Hash           string                 `json:"hash"`
	PreviousHash   string                 `json:"previous_hash"`
}

type IntegrityCheck struct {
	LastVerified  time.Time `json:"last_verified"`
	Status        string    `json:"status"` // valid, invalid
	EntriesChecked int       `json:"entries_checked"`
	ErrorsFound   int       `json:"errors_found"`
}

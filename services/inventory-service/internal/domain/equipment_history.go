package domain

import "time"

type EquipmentHistory struct {
	ID          string                 `json:"id"`
	TenantID    string                 `json:"tenant_id"`
	EquipmentID string                 `json:"equipment_id"`
	Action      string                 `json:"action"`
	ChangedBy   string                 `json:"changed_by,omitempty"`
	OldValue    map[string]interface{} `json:"old_value,omitempty"`
	NewValue    map[string]interface{} `json:"new_value,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
}

const (
	ActionCreated       = "created"
	ActionUpdated       = "updated"
	ActionStatusChanged = "status_changed"
	ActionDeleted       = "deleted"
	ActionImageAdded    = "image_added"
)

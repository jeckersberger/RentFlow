package domain

import (
	"time"
)

// Domain Events

type EquipmentCreatedEvent struct {
	EquipmentID     string    `json:"equipment_id"`
	TenantID        string    `json:"tenant_id"`
	Name            string    `json:"name"`
	CategoryID      string    `json:"category_id"`
	SKU             string    `json:"sku"`
	Barcode         string    `json:"barcode"`
	CreatedByUserID string    `json:"created_by_user_id"`
	CreatedAt       time.Time `json:"created_at"`
}

type EquipmentUpdatedEvent struct {
	EquipmentID string    `json:"equipment_id"`
	TenantID    string    `json:"tenant_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type EquipmentStatusChangedEvent struct {
	EquipmentID    string    `json:"equipment_id"`
	TenantID       string    `json:"tenant_id"`
	PreviousStatus string    `json:"previous_status"`
	NewStatus      string    `json:"new_status"`
	Reason         string    `json:"reason"`
	ChangedAt      time.Time `json:"changed_at"`
}

type EquipmentLocationChangedEvent struct {
	EquipmentID string    `json:"equipment_id"`
	TenantID    string    `json:"tenant_id"`
	LocationID  string    `json:"location_id"`
	ChangedAt   time.Time `json:"changed_at"`
}

type EquipmentConditionUpdatedEvent struct {
	EquipmentID string    `json:"equipment_id"`
	TenantID    string    `json:"tenant_id"`
	Condition   string    `json:"condition"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type EquipmentImageAddedEvent struct {
	EquipmentID string    `json:"equipment_id"`
	TenantID    string    `json:"tenant_id"`
	ImageRef    string    `json:"image_ref"`
	AddedAt     time.Time `json:"added_at"`
}

type EquipmentRetiredEvent struct {
	EquipmentID string    `json:"equipment_id"`
	TenantID    string    `json:"tenant_id"`
	Reason      string    `json:"reason"`
	RetiredAt   time.Time `json:"retired_at"`
}

type CategoryCreatedEvent struct {
	CategoryID string    `json:"category_id"`
	TenantID   string    `json:"tenant_id"`
	Name       string    `json:"name"`
	ParentID   *string   `json:"parent_id,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

type CategoryUpdatedEvent struct {
	CategoryID string    `json:"category_id"`
	TenantID   string    `json:"tenant_id"`
	Name       string    `json:"name"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type CategoryDeletedEvent struct {
	CategoryID string    `json:"category_id"`
	TenantID   string    `json:"tenant_id"`
	DeletedAt  time.Time `json:"deleted_at"`
}

type FlightcaseCreatedEvent struct {
	FlightcaseID string    `json:"flightcase_id"`
	TenantID     string    `json:"tenant_id"`
	Name         string    `json:"name"`
	Barcode      string    `json:"barcode"`
	CreatedAt    time.Time `json:"created_at"`
}

type FlightcaseItemAddedEvent struct {
	FlightcaseID string    `json:"flightcase_id"`
	TenantID     string    `json:"tenant_id"`
	EquipmentID  string    `json:"equipment_id"`
	Quantity     int       `json:"quantity"`
	AddedAt      time.Time `json:"added_at"`
}

type FlightcaseItemRemovedEvent struct {
	FlightcaseID string    `json:"flightcase_id"`
	TenantID     string    `json:"tenant_id"`
	EquipmentID  string    `json:"equipment_id"`
	Quantity     int       `json:"quantity"`
	RemovedAt    time.Time `json:"removed_at"`
}

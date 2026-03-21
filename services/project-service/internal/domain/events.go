package domain

import "time"

// Domain Events

type ProjectCreatedEvent struct {
	ProjectID       string    `json:"project_id"`
	TenantID        string    `json:"tenant_id"`
	Name            string    `json:"name"`
	ClientName      string    `json:"client_name"`
	Status          string    `json:"status"`
	CreatedByUserID string    `json:"created_by_user_id"`
	CreatedAt       time.Time `json:"created_at"`
}

type ProjectUpdatedEvent struct {
	ProjectID string    `json:"project_id"`
	TenantID  string    `json:"tenant_id"`
	Name      string    `json:"name"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ProjectStatusChangedEvent struct {
	ProjectID      string    `json:"project_id"`
	TenantID       string    `json:"tenant_id"`
	PreviousStatus string    `json:"previous_status"`
	NewStatus      string    `json:"new_status"`
	ChangedAt      time.Time `json:"changed_at"`
}

type PacklistCreatedEvent struct {
	PacklistID string    `json:"packlist_id"`
	ProjectID  string    `json:"project_id"`
	TenantID   string    `json:"tenant_id"`
	Name       string    `json:"name"`
	CreatedAt  time.Time `json:"created_at"`
}

type PacklistItemAddedEvent struct {
	PacklistID  string    `json:"packlist_id"`
	ProjectID   string    `json:"project_id"`
	TenantID    string    `json:"tenant_id"`
	EquipmentID string    `json:"equipment_id"`
	Quantity    int       `json:"quantity"`
	AddedAt     time.Time `json:"added_at"`
}

type PacklistItemPackedEvent struct {
	PacklistID     string    `json:"packlist_id"`
	ProjectID      string    `json:"project_id"`
	TenantID       string    `json:"tenant_id"`
	ItemID         string    `json:"item_id"`
	EquipmentID    string    `json:"equipment_id"`
	QuantityPacked int       `json:"quantity_packed"`
	PackedAt       time.Time `json:"packed_at"`
}

type PacklistItemReturnedEvent struct {
	PacklistID       string    `json:"packlist_id"`
	ProjectID        string    `json:"project_id"`
	TenantID         string    `json:"tenant_id"`
	ItemID           string    `json:"item_id"`
	EquipmentID      string    `json:"equipment_id"`
	QuantityReturned int       `json:"quantity_returned"`
	ReturnedAt       time.Time `json:"returned_at"`
}

type ReservationCreatedEvent struct {
	ReservationID string    `json:"reservation_id"`
	ProjectID     string    `json:"project_id"`
	TenantID      string    `json:"tenant_id"`
	EquipmentID   string    `json:"equipment_id"`
	StartDate     time.Time `json:"start_date"`
	EndDate       time.Time `json:"end_date"`
	CreatedAt     time.Time `json:"created_at"`
}

type ReservationConfirmedEvent struct {
	ReservationID string    `json:"reservation_id"`
	ProjectID     string    `json:"project_id"`
	TenantID      string    `json:"tenant_id"`
	ConfirmedAt   time.Time `json:"confirmed_at"`
}

type ReservationCancelledEvent struct {
	ReservationID string    `json:"reservation_id"`
	ProjectID     string    `json:"project_id"`
	TenantID      string    `json:"tenant_id"`
	Reason        string    `json:"reason"`
	CancelledAt   time.Time `json:"cancelled_at"`
}

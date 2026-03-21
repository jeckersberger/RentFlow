package domain

import (
	"fmt"
	"time"
)

type MovementType string

const (
	MovementInbound   MovementType = "inbound"
	MovementOutbound  MovementType = "outbound"
	MovementTransfer  MovementType = "transfer"
	MovementReturn    MovementType = "return"
	MovementAdjust    MovementType = "adjustment"
)

type Movement struct {
	ID            string
	TenantID      string
	EquipmentID   string
	FromLocationID *string
	ToLocationID   string
	MovementType  MovementType
	Quantity      int
	Reason        string
	UserID        string
	ProjectID     *string
	Timestamp     time.Time
	CreatedAt     time.Time
}

func NewMovement(
	id string,
	tenantID string,
	equipmentID string,
	toLocationID string,
	movementType MovementType,
	quantity int,
	userID string,
) *Movement {
	return &Movement{
		ID:           id,
		TenantID:     tenantID,
		EquipmentID:  equipmentID,
		ToLocationID: toLocationID,
		MovementType: movementType,
		Quantity:     quantity,
		UserID:       userID,
		Timestamp:    time.Now(),
		CreatedAt:    time.Now(),
	}
}

func (m *Movement) Validate() error {
	if m.TenantID == "" {
		return fmt.Errorf("tenant ID is required")
	}
	if m.EquipmentID == "" {
		return fmt.Errorf("equipment ID is required")
	}
	if m.ToLocationID == "" {
		return fmt.Errorf("to location ID is required")
	}
	if m.Quantity <= 0 {
		return fmt.Errorf("quantity must be greater than 0")
	}
	if m.UserID == "" {
		return fmt.Errorf("user ID is required")
	}
	return nil
}

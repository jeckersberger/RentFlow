package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// DamageType represents the type of damage in a claim
type DamageType string

const (
	DamageTypeTheft     DamageType = "theft"
	DamageTypeBreakage  DamageType = "breakage"
	DamageTypeWater     DamageType = "water"
	DamageTypeFire      DamageType = "fire"
	DamageTypeTransport DamageType = "transport"
	DamageTypeOther     DamageType = "other"
)

// ClaimStatus represents the status of a claim in the workflow
type ClaimStatus string

const (
	ClaimStatusReported   ClaimStatus = "reported"
	ClaimStatusDocumented ClaimStatus = "documented"
	ClaimStatusSubmitted  ClaimStatus = "submitted"
	ClaimStatusInReview   ClaimStatus = "in_review"
	ClaimStatusApproved   ClaimStatus = "approved"
	ClaimStatusRejected   ClaimStatus = "rejected"
	ClaimStatusSettled    ClaimStatus = "settled"
)

// Claim represents an insurance claim
type Claim struct {
	ID              uuid.UUID
	TenantID        uuid.UUID
	PolicyID        uuid.UUID
	ClaimNumber     string
	EquipmentID     *uuid.UUID
	ProjectID       *uuid.UUID
	IncidentDate    time.Time
	ReportedDate    time.Time
	Description     string
	DamageType      DamageType
	Status          ClaimStatus
	ClaimedAmount   float64
	ApprovedAmount  *float64
	SettledAmount   *float64
	AdjusterNotes   *string
	CreatedBy       uuid.UUID
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// CanTransitionTo checks if the claim can transition to the target status
func (c *Claim) CanTransitionTo(targetStatus ClaimStatus) error {
	validTransitions := map[ClaimStatus][]ClaimStatus{
		ClaimStatusReported:   {ClaimStatusDocumented},
		ClaimStatusDocumented: {ClaimStatusSubmitted},
		ClaimStatusSubmitted:  {ClaimStatusInReview},
		ClaimStatusInReview:   {ClaimStatusApproved, ClaimStatusRejected},
		ClaimStatusApproved:   {ClaimStatusSettled},
		ClaimStatusRejected:   {},
		ClaimStatusSettled:    {},
	}

	validStatuses, exists := validTransitions[c.Status]
	if !exists {
		return fmt.Errorf("unknown current status: %s", c.Status)
	}

	for _, valid := range validStatuses {
		if valid == targetStatus {
			return nil
		}
	}

	return fmt.Errorf("cannot transition from %s to %s", c.Status, targetStatus)
}

// TransitionTo transitions the claim to a new status after validation
func (c *Claim) TransitionTo(newStatus ClaimStatus) error {
	if err := c.CanTransitionTo(newStatus); err != nil {
		return err
	}
	c.Status = newStatus
	c.UpdatedAt = time.Now()
	return nil
}

// IsTerminal returns true if the claim is in a terminal state
func (c *Claim) IsTerminal() bool {
	return c.Status == ClaimStatusRejected || c.Status == ClaimStatusSettled
}

// ClaimItem represents an item within a claim
type ClaimItem struct {
	ID               uuid.UUID
	ClaimID          uuid.UUID
	EquipmentID      uuid.UUID
	Description      string
	ReplacementValue float64
	RepairCost       *float64
	PhotoURLs        []string
	CreatedAt        time.Time
}

// GetTotalValue returns the total value (replacement or repair) of the item
func (ci *ClaimItem) GetTotalValue() float64 {
	if ci.RepairCost != nil && *ci.RepairCost > 0 {
		return *ci.RepairCost
	}
	return ci.ReplacementValue
}

package domain

import "errors"

// Sentinel errors for the federation domain.
var (
	ErrPartnerNotFound       = errors.New("federation partner not found")
	ErrListingNotFound       = errors.New("shared listing not found")
	ErrRequestNotFound       = errors.New("federation request not found")
	ErrMissingPartnerName    = errors.New("partner_name is required")
	ErrMissingEquipmentID    = errors.New("equipment_id is required")
	ErrInvalidPartnerStatus  = errors.New("invalid partner status")
	ErrInvalidRequestStatus  = errors.New("invalid request status")
)

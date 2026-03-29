package domain

import "errors"

var (
	ErrPolicyNotFound         = errors.New("insurance policy not found")
	ErrClaimNotFound          = errors.New("insurance claim not found")
	ErrNameRequired           = errors.New("name is required")
	ErrDescriptionRequired    = errors.New("description is required")
	ErrPolicyIDRequired       = errors.New("policy_id is required")
	ErrEquipmentIDRequired    = errors.New("equipment_id is required")
	ErrInvalidStatusTransition = errors.New("invalid claim status transition")
	ErrStatusRequired         = errors.New("status is required")
)

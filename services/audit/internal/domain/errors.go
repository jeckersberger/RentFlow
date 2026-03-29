package domain

import "errors"

var (
	ErrAuditLogNotFound   = errors.New("audit log not found")
	ErrPolicyNotFound     = errors.New("audit policy not found")
	ErrActionRequired     = errors.New("action is required")
	ErrEntityTypeRequired = errors.New("entity_type is required")
	ErrChainBroken        = errors.New("audit hash chain integrity violated")
)

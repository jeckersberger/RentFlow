package domain

import "fmt"

// ErrPolicyNotFound is returned when a policy is not found
type ErrPolicyNotFound struct {
	ID string
}

func (e ErrPolicyNotFound) Error() string {
	return fmt.Sprintf("policy not found: %s", e.ID)
}

// ErrClaimNotFound is returned when a claim is not found
type ErrClaimNotFound struct {
	ID string
}

func (e ErrClaimNotFound) Error() string {
	return fmt.Sprintf("claim not found: %s", e.ID)
}

// ErrClaimItemNotFound is returned when a claim item is not found
type ErrClaimItemNotFound struct {
	ID string
}

func (e ErrClaimItemNotFound) Error() string {
	return fmt.Sprintf("claim item not found: %s", e.ID)
}

// ErrInvalidTransition is returned when a claim state transition is invalid
type ErrInvalidTransition struct {
	From string
	To   string
}

func (e ErrInvalidTransition) Error() string {
	return fmt.Sprintf("invalid transition from %s to %s", e.From, e.To)
}

// ErrPolicyInactive is returned when trying to use an inactive policy
type ErrPolicyInactive struct {
	ID string
}

func (e ErrPolicyInactive) Error() string {
	return fmt.Sprintf("policy is inactive: %s", e.ID)
}

// ErrInsufficientCoverage is returned when the claim exceeds the coverage
type ErrInsufficientCoverage struct {
	ClaimedAmount  float64
	CoverageAmount float64
}

func (e ErrInsufficientCoverage) Error() string {
	return fmt.Sprintf("insufficient coverage: claimed %.2f, available %.2f", e.ClaimedAmount, e.CoverageAmount)
}

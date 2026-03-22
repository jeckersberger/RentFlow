package domain

import (
	"time"

	"github.com/google/uuid"
)

// PolicyType represents the type of insurance policy
type PolicyType string

const (
	PolicyTypeLiability    PolicyType = "liability"
	PolicyTypeComprehensive PolicyType = "comprehensive"
	PolicyTypeTransport    PolicyType = "transport"
	PolicyTypeRenter       PolicyType = "renter"
)

// PolicyStatus represents the status of a policy
type PolicyStatus string

const (
	PolicyStatusActive   PolicyStatus = "active"
	PolicyStatusExpired  PolicyStatus = "expired"
	PolicyStatusCancelled PolicyStatus = "cancelled"
)

// Policy represents an insurance policy
type Policy struct {
	ID             uuid.UUID
	TenantID       uuid.UUID
	PolicyNumber   string
	PolicyType     PolicyType
	Provider       string
	CoverageAmount float64
	Deductible     float64
	PremiumAnnual  float64
	PremiumMonthly float64
	StartDate      time.Time
	EndDate        time.Time
	Status         PolicyStatus
	Notes          *string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// IsExpired checks if the policy has expired
func (p *Policy) IsExpired() bool {
	return time.Now().After(p.EndDate) || p.Status == PolicyStatusExpired
}

// IsActive checks if the policy is currently active
func (p *Policy) IsActive() bool {
	return p.Status == PolicyStatusActive && !p.IsExpired()
}

// CanCover checks if the policy can cover the given amount
func (p *Policy) CanCover(amount float64) bool {
	return amount <= p.CoverageAmount && p.IsActive()
}

// DaysUntilExpiry returns the number of days until the policy expires
func (p *Policy) DaysUntilExpiry() int {
	return int(time.Until(p.EndDate).Hours() / 24)
}

package domain

import (
	"context"

	"github.com/google/uuid"
)

type InsurancePolicyRepository interface {
	Create(ctx context.Context, policy *InsurancePolicy) error
	GetByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*InsurancePolicy, error)
	List(ctx context.Context, tenantID uuid.UUID, filter PolicyFilter) ([]*InsurancePolicy, int64, error)
	Update(ctx context.Context, policy *InsurancePolicy) error
}

type InsuredEquipmentRepository interface {
	Add(ctx context.Context, eq *InsuredEquipment) error
	ListByPolicy(ctx context.Context, policyID uuid.UUID, filter EquipmentFilter) ([]*InsuredEquipment, int64, error)
}

type InsuranceClaimRepository interface {
	Create(ctx context.Context, claim *InsuranceClaim) error
	GetByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*InsuranceClaim, error)
	List(ctx context.Context, tenantID uuid.UUID, filter ClaimFilter) ([]*InsuranceClaim, int64, error)
	Update(ctx context.Context, claim *InsuranceClaim) error
}

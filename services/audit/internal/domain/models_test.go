package domain

import (
	"testing"

	"github.com/google/uuid"
)

func TestAuditLogChecksumChain(t *testing.T) {
	log1 := &AuditLog{
		ID:             uuid.New(),
		Action:         "create",
		EntityType:     "equipment",
		SequenceNumber: 1,
		Checksum:       "sha256:abc",
		PrevChecksum:   "",
	}

	log2 := &AuditLog{
		ID:             uuid.New(),
		Action:         "update",
		EntityType:     "equipment",
		SequenceNumber: 2,
		Checksum:       "sha256:def",
		PrevChecksum:   log1.Checksum,
	}

	if log2.PrevChecksum != log1.Checksum {
		t.Error("chain should link prev to prior checksum")
	}
	if log2.SequenceNumber != log1.SequenceNumber+1 {
		t.Error("sequence numbers should be consecutive")
	}
}

func TestAuditPolicyRetention(t *testing.T) {
	policy := &AuditPolicy{
		EntityType:    "invoice",
		RetentionDays: 3650, // 10 years (GoBD)
		LogReads:      false,
		LogWrites:     true,
		IsActive:      true,
	}

	if policy.RetentionDays < 3650 {
		t.Errorf("invoice retention should be >= 10 years, got %d days", policy.RetentionDays)
	}
	if !policy.LogWrites {
		t.Error("should log writes for invoices")
	}
}

func TestAuditLogEventSourcing(t *testing.T) {
	log := &AuditLog{
		AggregateType: "project",
		AggregateID:   &uuid.UUID{},
		EventType:     "project.status_changed",
		ServiceName:   "project-service",
	}

	if log.AggregateType != "project" {
		t.Errorf("AggregateType = %q, want %q", log.AggregateType, "project")
	}
	if log.EventType != "project.status_changed" {
		t.Errorf("EventType = %q, want %q", log.EventType, "project.status_changed")
	}
}

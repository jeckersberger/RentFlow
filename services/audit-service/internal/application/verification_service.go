package application

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/audit-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/audit-service/internal/ports"
)

type VerificationService struct {
	repo ports.AuditRepository
	log  logger.Logger
}

func NewVerificationService(repo ports.AuditRepository, log logger.Logger) *VerificationService {
	return &VerificationService{
		repo: repo,
		log:  log,
	}
}

func (s *VerificationService) VerifyChain(ctx context.Context, tenantID uuid.UUID) (*domain.VerificationResult, error) {
	entries, err := s.repo.ListByTenant(ctx, tenantID)
	if err != nil {
		s.log.Error("Failed to fetch audit entries", err)
		return nil, err
	}

	result := &domain.VerificationResult{
		IsValid:      true,
		EntriesCount: len(entries),
	}

	var previousChecksum *string

	for _, entry := range entries {
		canonical := fmt.Sprintf("%s|%s|%s|%s|%s|%s|%s",
			entry.TenantID.String(),
			entry.ServiceName,
			entry.Operation,
			entry.EntityType,
			func() string {
				if entry.EntityID != nil {
					return entry.EntityID.String()
				}
				return ""
			}(),
			entry.Timestamp.UTC().Format(time.RFC3339Nano),
			func() string {
				if previousChecksum != nil {
					return *previousChecksum
				}
				return ""
			}())

		hash := sha256.Sum256([]byte(canonical))
		computedChecksum := hex.EncodeToString(hash[:])

		if computedChecksum != entry.Checksum {
			result.IsValid = false
			result.FirstMismatch = &domain.MismatchDetail{
				EntryID:          entry.ID,
				SequenceNumber:   entry.SequenceNumber,
				ComputedChecksum: computedChecksum,
				StoredChecksum:   entry.Checksum,
			}
			s.log.Error("Checksum mismatch detected", nil, "entryID", entry.ID)
			return result, nil
		}

		previousChecksum = &entry.Checksum
	}

	s.log.Info("Verification complete", "isValid", result.IsValid, "entriesChecked", result.EntriesCount)
	return result, nil
}

package application

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/audit-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/audit-service/internal/ports"
)

type PseudonymizationService struct {
	repo ports.AuditRepository
	log  logger.Logger
}

func NewPseudonymizationService(repo ports.AuditRepository, log logger.Logger) *PseudonymizationService {
	return &PseudonymizationService{
		repo: repo,
		log:  log,
	}
}

func (s *PseudonymizationService) PseudonymizeUser(ctx context.Context, cmd domain.PseudonymizeUserCmd) error {
	// Use deterministic pseudonym based on hash of userID
	hash := sha256.Sum256([]byte(cmd.UserID.String()))
	pseudonym := fmt.Sprintf("PSEUDONYM_%s", hex.EncodeToString(hash[:8]))

	s.log.Info("Pseudonymizing user", "userID", cmd.UserID, "tenantID", cmd.TenantID, "pseudonym", pseudonym)

	// Call repository to pseudonymize all entries for this user
	return s.repo.Pseudonymize(ctx, cmd.TenantID, cmd.UserID)
}

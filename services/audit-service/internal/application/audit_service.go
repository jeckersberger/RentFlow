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

type AuditService struct {
	repo ports.AuditRepository
	log  logger.Logger
}

func NewAuditService(repo ports.AuditRepository, log logger.Logger) *AuditService {
	return &AuditService{
		repo: repo,
		log:  log,
	}
}

func (s *AuditService) WriteEntry(ctx context.Context, cmd domain.WriteAuditEntryCmd) (*domain.AuditEntry, error) {
	// Get last checksum for this tenant
	lastChecksum, err := s.repo.GetLastChecksum(ctx, cmd.TenantID)
	if err != nil {
		s.log.Error("Failed to get last checksum", err)
		return nil, err
	}

	// Build canonical string for hashing
	lastChecksumStr := ""
	if lastChecksum != nil {
		lastChecksumStr = *lastChecksum
	}

	canonical := fmt.Sprintf("%s|%s|%s|%s|%s|%s|%s",
		cmd.TenantID.String(),
		cmd.ServiceName,
		cmd.Operation,
		cmd.EntityType,
		func() string {
			if cmd.EntityID != nil {
				return cmd.EntityID.String()
			}
			return ""
		}(),
		time.Now().UTC().Format(time.RFC3339Nano),
		lastChecksumStr)

	// Compute SHA-256
	hash := sha256.Sum256([]byte(canonical))
	checksum := hex.EncodeToString(hash[:])

	// Create entry
	entry := &domain.AuditEntry{
		ID:               uuid.New(),
		TenantID:         cmd.TenantID,
		ServiceName:      cmd.ServiceName,
		Operation:        cmd.Operation,
		EntityType:       cmd.EntityType,
		EntityID:         cmd.EntityID,
		UserID:           cmd.UserID,
		UserName:         cmd.UserName,
		OldValues:        cmd.OldValues,
		NewValues:        cmd.NewValues,
		IPAddress:        cmd.IPAddress,
		UserAgent:        cmd.UserAgent,
		Checksum:         checksum,
		PreviousChecksum: lastChecksum,
		Timestamp:        time.Now().UTC(),
		CreatedAt:        time.Now().UTC(),
	}

	created, err := s.repo.Create(ctx, entry)
	if err != nil {
		s.log.Error("Failed to create audit entry", err)
		return nil, err
	}

	s.log.Info("Audit entry created", "entryID", created.ID, "operation", cmd.Operation)
	return created, nil
}

func (s *AuditService) GetEntry(ctx context.Context, id uuid.UUID) (*domain.AuditEntry, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *AuditService) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*domain.AuditEntry, error) {
	return s.repo.ListByTenant(ctx, tenantID)
}

func (s *AuditService) ListByEntity(ctx context.Context, tenantID uuid.UUID, entityType string, entityID uuid.UUID) ([]*domain.AuditEntry, error) {
	return s.repo.ListByEntity(ctx, tenantID, entityType, entityID)
}

func (s *AuditService) ListByUser(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID) ([]*domain.AuditEntry, error) {
	return s.repo.ListByUser(ctx, tenantID, userID)
}

func (s *AuditService) ListByDateRange(ctx context.Context, tenantID uuid.UUID, from, to interface{}) ([]*domain.AuditEntry, error) {
	return s.repo.ListByDateRange(ctx, tenantID, from, to)
}

func (s *AuditService) GetDashboardStats(ctx context.Context, tenantID uuid.UUID) (*domain.GetDashboardStatsResponse, error) {
	todayCount, err := s.repo.GetTodayCount(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	allEntries, err := s.repo.ListByTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	return &domain.GetDashboardStatsResponse{
		EntriesAdded:       todayCount,
		ChainStatus:        "valid",
		LastVerification:   &now,
		TotalEntries:       len(allEntries),
		PseudonymizedUsers: 0,
	}, nil
}

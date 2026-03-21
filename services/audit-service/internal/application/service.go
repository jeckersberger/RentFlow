package application

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/audit-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/audit-service/internal/ports"
)

type AuditEntryDTO struct {
	ID           string      `json:"id"`
	Timestamp    time.Time   `json:"timestamp"`
	UserID       string      `json:"user_id"`
	Action       string      `json:"action"`
	EntityType   string      `json:"entity_type"`
	EntityID     string      `json:"entity_id"`
	IPAddress    string      `json:"ip_address"`
	UserAgent    string      `json:"user_agent"`
	Hash         string      `json:"hash"`
}

type LogAuditCommand struct {
	TenantID     string                 `json:"tenant_id"`
	UserID       string                 `json:"user_id"`
	Action       string                 `json:"action"`
	EntityType   string                 `json:"entity_type"`
	EntityID     string                 `json:"entity_id"`
	PreviousState map[string]interface{} `json:"previous_state,omitempty"`
	NewState     map[string]interface{} `json:"new_state,omitempty"`
	IPAddress    string                 `json:"ip_address"`
	UserAgent    string                 `json:"user_agent"`
}

type AuditService struct {
	auditRepo     ports.AuditEntryRepository
	integrityRepo ports.IntegrityCheckRepository
	logger        logger.Logger
}

func NewAuditService(
	auditRepo ports.AuditEntryRepository,
	integrityRepo ports.IntegrityCheckRepository,
	log logger.Logger,
) *AuditService {
	return &AuditService{
		auditRepo:     auditRepo,
		integrityRepo: integrityRepo,
		logger:        log,
	}
}

func (s *AuditService) LogAudit(ctx context.Context, cmd LogAuditCommand) (*AuditEntryDTO, error) {
	if cmd.TenantID == "" || cmd.Action == "" || cmd.EntityType == "" {
		return nil, domain.ErrInvalidInput
	}

	// Get previous hash for chain
	lastEntry, _ := s.auditRepo.GetLastEntry(ctx, cmd.TenantID)
	previousHash := ""
	if lastEntry != nil {
		previousHash = lastEntry.Hash
	}

	// Calculate hash
	hashInput := fmt.Sprintf("%s%s%s%s%s%s",
		previousHash,
		time.Now().Format(time.RFC3339Nano),
		cmd.Action,
		cmd.EntityType,
		cmd.EntityID,
		fmt.Sprintf("%v", cmd.NewState),
	)
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(hashInput)))

	entry := &domain.AuditEntry{
		ID:            fmt.Sprintf("audit_%d", time.Now().UnixNano()),
		TenantID:      cmd.TenantID,
		Timestamp:     time.Now(),
		UserID:        cmd.UserID,
		Action:        cmd.Action,
		EntityType:    cmd.EntityType,
		EntityID:      cmd.EntityID,
		PreviousState: cmd.PreviousState,
		NewState:      cmd.NewState,
		IPAddress:     cmd.IPAddress,
		UserAgent:     cmd.UserAgent,
		Hash:          hash,
		PreviousHash:  previousHash,
	}

	if err := s.auditRepo.Create(ctx, entry); err != nil {
		s.logger.Error("Failed to log audit entry", err)
		return nil, err
	}

	return &AuditEntryDTO{
		ID:         entry.ID,
		Timestamp:  entry.Timestamp,
		UserID:     entry.UserID,
		Action:     entry.Action,
		EntityType: entry.EntityType,
		EntityID:   entry.EntityID,
		IPAddress:  entry.IPAddress,
		UserAgent:  entry.UserAgent,
		Hash:       entry.Hash,
	}, nil
}

func (s *AuditService) GetAuditLog(ctx context.Context, tenantID, id string) (*AuditEntryDTO, error) {
	entry, err := s.auditRepo.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	if entry == nil {
		return nil, domain.ErrAuditEntryNotFound
	}

	return &AuditEntryDTO{
		ID:         entry.ID,
		Timestamp:  entry.Timestamp,
		UserID:     entry.UserID,
		Action:     entry.Action,
		EntityType: entry.EntityType,
		EntityID:   entry.EntityID,
		IPAddress:  entry.IPAddress,
		UserAgent:  entry.UserAgent,
		Hash:       entry.Hash,
	}, nil
}

func (s *AuditService) ListLogs(ctx context.Context, tenantID string, fromTime, toTime time.Time) ([]*AuditEntryDTO, error) {
	entries, err := s.auditRepo.ListByTenant(ctx, tenantID, fromTime, toTime)
	if err != nil {
		return nil, err
	}

	dtos := make([]*AuditEntryDTO, len(entries))
	for i, e := range entries {
		dtos[i] = &AuditEntryDTO{
			ID:         e.ID,
			Timestamp:  e.Timestamp,
			UserID:     e.UserID,
			Action:     e.Action,
			EntityType: e.EntityType,
			EntityID:   e.EntityID,
			IPAddress:  e.IPAddress,
			UserAgent:  e.UserAgent,
			Hash:       e.Hash,
		}
	}
	return dtos, nil
}

func (s *AuditService) ListByEntity(ctx context.Context, tenantID, entityType, entityID string) ([]*AuditEntryDTO, error) {
	entries, err := s.auditRepo.ListByEntity(ctx, tenantID, entityType, entityID)
	if err != nil {
		return nil, err
	}

	dtos := make([]*AuditEntryDTO, len(entries))
	for i, e := range entries {
		dtos[i] = &AuditEntryDTO{
			ID:         e.ID,
			Timestamp:  e.Timestamp,
			UserID:     e.UserID,
			Action:     e.Action,
			EntityType: e.EntityType,
			EntityID:   e.EntityID,
			IPAddress:  e.IPAddress,
			UserAgent:  e.UserAgent,
			Hash:       e.Hash,
		}
	}
	return dtos, nil
}

func (s *AuditService) ListByUser(ctx context.Context, tenantID, userID string) ([]*AuditEntryDTO, error) {
	entries, err := s.auditRepo.ListByUser(ctx, tenantID, userID)
	if err != nil {
		return nil, err
	}

	dtos := make([]*AuditEntryDTO, len(entries))
	for i, e := range entries {
		dtos[i] = &AuditEntryDTO{
			ID:         e.ID,
			Timestamp:  e.Timestamp,
			UserID:     e.UserID,
			Action:     e.Action,
			EntityType: e.EntityType,
			EntityID:   e.EntityID,
			IPAddress:  e.IPAddress,
			UserAgent:  e.UserAgent,
			Hash:       e.Hash,
		}
	}
	return dtos, nil
}

func (s *AuditService) VerifyIntegrity(ctx context.Context, tenantID string) (map[string]interface{}, error) {
	entries, err := s.auditRepo.ListByTenant(ctx, tenantID, time.Time{}, time.Now())
	if err != nil {
		return nil, err
	}

	errorsFound := 0
	for i, entry := range entries {
		if i > 0 && entry.PreviousHash != entries[i-1].Hash {
			errorsFound++
		}
	}

	check := &domain.IntegrityCheck{
		LastVerified:  time.Now(),
		Status:        "valid",
		EntriesChecked: len(entries),
		ErrorsFound:   errorsFound,
	}

	if errorsFound > 0 {
		check.Status = "invalid"
		return map[string]interface{}{
			"status":           "FAILED",
			"entries_checked":  len(entries),
			"errors_found":     errorsFound,
			"verified_at":      time.Now(),
		}, nil
	}

	s.integrityRepo.Create(ctx, check)

	return map[string]interface{}{
		"status":           "PASSED",
		"entries_checked":  len(entries),
		"errors_found":     errorsFound,
		"verified_at":      time.Now(),
	}, nil
}

func (s *AuditService) ExportLogs(ctx context.Context, tenantID string, fromTime, toTime time.Time) ([]byte, error) {
	entries, err := s.auditRepo.ListByTenant(ctx, tenantID, fromTime, toTime)
	if err != nil {
		return nil, err
	}

	return json.MarshalIndent(entries, "", "  ")
}

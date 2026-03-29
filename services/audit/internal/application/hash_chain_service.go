package application

import (
	"context"
	"crypto/sha256"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/services/audit/internal/domain"
)

// ---------------------------------------------------------------------------
// Request DTOs
// ---------------------------------------------------------------------------

// AppendWithHashRequest contains the fields needed for a hash-chained audit entry.
type AppendWithHashRequest struct {
	UserID        *uuid.UUID      `json:"user_id"`
	Action        string          `json:"action"`
	EntityType    string          `json:"entity_type"`
	EntityID      *uuid.UUID      `json:"entity_id"`
	AggregateType string          `json:"aggregate_type"`
	AggregateID   *uuid.UUID      `json:"aggregate_id"`
	EventType     string          `json:"event_type"`
	EventPayload  json.RawMessage `json:"event_payload"`
	ServiceName   string          `json:"service_name"`
	IPAddress     string          `json:"ip_address"`
	UserAgent     string          `json:"user_agent"`
}

// VerifyResult contains the result of a chain integrity verification.
type VerifyResult struct {
	Valid       bool   `json:"valid"`
	CheckedFrom int64  `json:"checked_from"`
	CheckedTo   int64  `json:"checked_to"`
	TotalChecked int64 `json:"total_checked"`
	FirstBroken *int64 `json:"first_broken,omitempty"`
}

// ExportFilter defines the date range for exporting audit logs.
type ExportFilter struct {
	FromDate time.Time
	ToDate   time.Time
}

// ---------------------------------------------------------------------------
// Service
// ---------------------------------------------------------------------------

// HashChainService manages hash-chained audit log entries.
type HashChainService struct {
	repo   domain.AuditLogRepository
	logger zerolog.Logger
}

// NewHashChainService creates a new HashChainService.
func NewHashChainService(repo domain.AuditLogRepository, logger zerolog.Logger) *HashChainService {
	return &HashChainService{
		repo:   repo,
		logger: logger.With().Str("service", "hash_chain").Logger(),
	}
}

// computeChecksum calculates SHA-256(id + eventPayload + prevChecksum).
func computeChecksum(id uuid.UUID, eventPayload json.RawMessage, prevChecksum string) string {
	payload := ""
	if eventPayload != nil {
		payload = string(eventPayload)
	}
	data := id.String() + payload + prevChecksum
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash)
}

// AppendWithHash appends an audit log entry with a SHA-256 hash chain.
func (s *HashChainService) AppendWithHash(ctx context.Context, tenantID uuid.UUID, req AppendWithHashRequest) (*domain.AuditLog, error) {
	if req.Action == "" {
		return nil, fmt.Errorf("action is required")
	}
	if req.EntityType == "" {
		return nil, fmt.Errorf("entity_type is required")
	}

	// 1. Get the last checksum for this tenant.
	prevChecksum, err := s.repo.GetLastChecksum(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("get last checksum: %w", err)
	}

	entryID := uuid.New()

	// Ensure event_payload defaults to empty JSON object.
	eventPayload := req.EventPayload
	if eventPayload == nil {
		eventPayload = json.RawMessage("{}")
	}

	// 2. Calculate: checksum = SHA256(entry.ID + entry.EventPayload + prevChecksum).
	checksum := computeChecksum(entryID, eventPayload, prevChecksum)

	entry := &domain.AuditLog{
		ID:            entryID,
		TenantID:      tenantID,
		UserID:        req.UserID,
		Action:        req.Action,
		EntityType:    req.EntityType,
		EntityID:      req.EntityID,
		IPAddress:     req.IPAddress,
		UserAgent:     req.UserAgent,
		Checksum:      checksum,
		PrevChecksum:  prevChecksum,
		AggregateType: req.AggregateType,
		AggregateID:   req.AggregateID,
		EventType:     req.EventType,
		EventPayload:  eventPayload,
		ServiceName:   req.ServiceName,
	}

	// 3. Save.
	if err := s.repo.CreateWithHash(ctx, entry); err != nil {
		return nil, fmt.Errorf("create audit log with hash: %w", err)
	}

	s.logger.Info().
		Str("audit_log_id", entry.ID.String()).
		Str("action", entry.Action).
		Int64("sequence_number", entry.SequenceNumber).
		Msg("hash-chained audit log appended")

	return entry, nil
}

// VerifyChain verifies the integrity of the hash chain for a given tenant
// between the specified sequence numbers (inclusive).
func (s *HashChainService) VerifyChain(ctx context.Context, tenantID uuid.UUID, fromSeq, toSeq int64) (*VerifyResult, error) {
	entries, err := s.repo.ListBySequenceRange(ctx, tenantID, fromSeq, toSeq)
	if err != nil {
		return nil, fmt.Errorf("list by sequence range: %w", err)
	}

	result := &VerifyResult{
		Valid:        true,
		CheckedFrom:  fromSeq,
		CheckedTo:    toSeq,
		TotalChecked: int64(len(entries)),
	}

	for i, entry := range entries {
		expectedPrev := ""
		if i > 0 {
			expectedPrev = entries[i-1].Checksum
		} else if entry.PrevChecksum != "" {
			// First entry in range: we trust the stored prev_checksum.
			expectedPrev = entry.PrevChecksum
		}

		expected := computeChecksum(entry.ID, entry.EventPayload, expectedPrev)
		if entry.Checksum != expected {
			result.Valid = false
			seq := entry.SequenceNumber
			result.FirstBroken = &seq
			break
		}

		// Also verify that prev_checksum matches the previous entry's checksum.
		if i > 0 && entry.PrevChecksum != entries[i-1].Checksum {
			result.Valid = false
			seq := entry.SequenceNumber
			result.FirstBroken = &seq
			break
		}
	}

	return result, nil
}

// ExportCSV writes audit log entries within the given date range as CSV to the writer.
func (s *HashChainService) ExportCSV(ctx context.Context, tenantID uuid.UUID, filter ExportFilter, w io.Writer) error {
	entries, err := s.repo.ListByDateRange(ctx, tenantID, filter.FromDate, filter.ToDate)
	if err != nil {
		return fmt.Errorf("list by date range: %w", err)
	}

	csvWriter := csv.NewWriter(w)
	defer csvWriter.Flush()

	// Write header.
	header := []string{
		"id", "sequence_number", "tenant_id", "user_id", "action",
		"entity_type", "entity_id", "aggregate_type", "aggregate_id",
		"event_type", "event_payload", "service_name",
		"checksum", "prev_checksum",
		"ip_address", "user_agent", "created_at",
	}
	if err := csvWriter.Write(header); err != nil {
		return fmt.Errorf("write csv header: %w", err)
	}

	for _, e := range entries {
		userID := ""
		if e.UserID != nil {
			userID = e.UserID.String()
		}
		entityID := ""
		if e.EntityID != nil {
			entityID = e.EntityID.String()
		}
		aggregateID := ""
		if e.AggregateID != nil {
			aggregateID = e.AggregateID.String()
		}
		payload := ""
		if e.EventPayload != nil {
			payload = string(e.EventPayload)
		}

		row := []string{
			e.ID.String(),
			fmt.Sprintf("%d", e.SequenceNumber),
			e.TenantID.String(),
			userID,
			e.Action,
			e.EntityType,
			entityID,
			e.AggregateType,
			aggregateID,
			e.EventType,
			payload,
			e.ServiceName,
			e.Checksum,
			e.PrevChecksum,
			e.IPAddress,
			e.UserAgent,
			e.CreatedAt.Format(time.RFC3339),
		}
		if err := csvWriter.Write(row); err != nil {
			return fmt.Errorf("write csv row: %w", err)
		}
	}

	return nil
}

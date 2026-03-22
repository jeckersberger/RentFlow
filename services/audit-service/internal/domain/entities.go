package domain

import (
	"time"

	"github.com/google/uuid"
)

type AuditEntry struct {
	ID                uuid.UUID
	TenantID          uuid.UUID
	SequenceNumber    int64
	Timestamp         time.Time
	ServiceName       string
	Operation         string
	EntityType        string
	EntityID          *uuid.UUID
	UserID            *uuid.UUID
	UserName          *string
	OldValues         map[string]interface{}
	NewValues         map[string]interface{}
	IPAddress         *string
	UserAgent         *string
	Checksum          string
	PreviousChecksum  *string
	IsPseudonymized   bool
	CreatedAt         time.Time
}

type AuditExport struct {
	ID             uuid.UUID
	TenantID       uuid.UUID
	ExportType     string
	DateFrom       time.Time
	DateTo         time.Time
	Status         string
	FilePath       *string
	FileSizeBytes  *int64
	Checksum       *string
	RequestedBy    *uuid.UUID
	CompletedAt    *time.Time
	CreatedAt      time.Time
}

type VerificationResult struct {
	IsValid       bool
	EntriesCount  int
	FirstMismatch *MismatchDetail
}

type MismatchDetail struct {
	EntryID          uuid.UUID
	SequenceNumber   int64
	ComputedChecksum string
	StoredChecksum   string
}

type WriteAuditEntryCmd struct {
	TenantID     uuid.UUID
	ServiceName  string
	Operation    string
	EntityType   string
	EntityID     *uuid.UUID
	UserID       *uuid.UUID
	UserName     *string
	OldValues    map[string]interface{}
	NewValues    map[string]interface{}
	IPAddress    *string
	UserAgent    *string
}

type PseudonymizeUserCmd struct {
	TenantID uuid.UUID
	UserID   uuid.UUID
}

type CreateExportCmd struct {
	TenantID   uuid.UUID
	ExportType string
	DateFrom   time.Time
	DateTo     time.Time
	RequestedBy *uuid.UUID
}

type GetDashboardStatsResponse struct {
	EntriesAdded      int
	ChainStatus       string
	LastVerification  *time.Time
	TotalEntries      int
	PseudonymizedUsers int
}

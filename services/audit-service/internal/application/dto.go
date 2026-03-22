package application

import (
	"time"

	"github.com/google/uuid"
)

type AuditEntryResponse struct {
	ID               uuid.UUID                  `json:"id"`
	TenantID         uuid.UUID                  `json:"tenant_id"`
	SequenceNumber   int64                      `json:"sequence_number"`
	Timestamp        time.Time                  `json:"timestamp"`
	ServiceName      string                     `json:"service_name"`
	Operation        string                     `json:"operation"`
	EntityType       string                     `json:"entity_type"`
	EntityID         *uuid.UUID                 `json:"entity_id"`
	UserID           *uuid.UUID                 `json:"user_id"`
	UserName         *string                    `json:"user_name"`
	OldValues        map[string]interface{}     `json:"old_values"`
	NewValues        map[string]interface{}     `json:"new_values"`
	IPAddress        *string                    `json:"ip_address"`
	Checksum         string                     `json:"checksum"`
	IsPseudonymized  bool                       `json:"is_pseudonymized"`
	CreatedAt        time.Time                  `json:"created_at"`
}

type WriteAuditEntryRequest struct {
	ServiceName string                 `json:"service_name"`
	Operation   string                 `json:"operation"`
	EntityType  string                 `json:"entity_type"`
	EntityID    *uuid.UUID             `json:"entity_id"`
	UserID      *uuid.UUID             `json:"user_id"`
	UserName    *string                `json:"user_name"`
	OldValues   map[string]interface{} `json:"old_values"`
	NewValues   map[string]interface{} `json:"new_values"`
	IPAddress   *string                `json:"ip_address"`
	UserAgent   *string                `json:"user_agent"`
}

type CreateExportRequest struct {
	ExportType string    `json:"export_type"`
	DateFrom   time.Time `json:"date_from"`
	DateTo     time.Time `json:"date_to"`
}

type VerificationResponse struct {
	IsValid       bool                   `json:"is_valid"`
	EntriesCount  int                    `json:"entries_count"`
	FirstMismatch *MismatchDetailResponse `json:"first_mismatch"`
}

type MismatchDetailResponse struct {
	EntryID          uuid.UUID `json:"entry_id"`
	SequenceNumber   int64     `json:"sequence_number"`
	ComputedChecksum string    `json:"computed_checksum"`
	StoredChecksum   string    `json:"stored_checksum"`
}

type ExportResponse struct {
	ID            uuid.UUID  `json:"id"`
	ExportType    string     `json:"export_type"`
	DateFrom      time.Time  `json:"date_from"`
	DateTo        time.Time  `json:"date_to"`
	Status        string     `json:"status"`
	FileSizeBytes *int64     `json:"file_size_bytes"`
	Checksum      *string    `json:"checksum"`
	CompletedAt   *time.Time `json:"completed_at"`
	CreatedAt     time.Time  `json:"created_at"`
}

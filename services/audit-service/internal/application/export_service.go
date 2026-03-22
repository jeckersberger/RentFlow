package application

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/audit-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/audit-service/internal/ports"
)

type ExportService struct {
	auditRepo  ports.AuditRepository
	exportRepo ports.ExportRepository
	log        logger.Logger
}

func NewExportService(auditRepo ports.AuditRepository, exportRepo ports.ExportRepository, log logger.Logger) *ExportService {
	return &ExportService{
		auditRepo:  auditRepo,
		exportRepo: exportRepo,
		log:        log,
	}
}

func (s *ExportService) CreateExport(ctx context.Context, cmd domain.CreateExportCmd) (*domain.AuditExport, error) {
	exportID := uuid.New()

	export := &domain.AuditExport{
		ID:         exportID,
		TenantID:   cmd.TenantID,
		ExportType: cmd.ExportType,
		DateFrom:   cmd.DateFrom,
		DateTo:     cmd.DateTo,
		Status:     "processing",
		RequestedBy: cmd.RequestedBy,
		CreatedAt:  time.Now().UTC(),
	}

	created, err := s.exportRepo.Create(ctx, export)
	if err != nil {
		s.log.Error("Failed to create export record", err)
		return nil, err
	}

	go s.processExport(context.Background(), created)

	return created, nil
}

func (s *ExportService) processExport(ctx context.Context, export *domain.AuditExport) {
	entries, err := s.auditRepo.ListByTenant(ctx, export.TenantID)
	if err != nil {
		s.log.Error("Failed to fetch audit entries for export", err)
		export.Status = "failed"
		s.exportRepo.Update(ctx, export)
		return
	}

	// Create ZIP file in memory
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)
	defer zipWriter.Close()

	// Write audit log as JSON
	auditData, _ := json.MarshalIndent(entries, "", "  ")
	auditFile, _ := zipWriter.Create("audit_log.json")
	auditFile.Write(auditData)

	// Write metadata
	metadata := map[string]interface{}{
		"export_id":   export.ID,
		"tenant_id":   export.TenantID,
		"export_type": export.ExportType,
		"date_from":   export.DateFrom,
		"date_to":     export.DateTo,
		"entry_count": len(entries),
		"created_at":  time.Now().UTC(),
	}
	metadataData, _ := json.MarshalIndent(metadata, "", "  ")
	metadataFile, _ := zipWriter.Create("metadata.json")
	metadataFile.Write(metadataData)

	zipWriter.Close()

	// Compute checksum of the ZIP
	hash := sha256.Sum256(buf.Bytes())
	checksum := hex.EncodeToString(hash[:])

	// Update export record
	export.Status = "completed"
	export.FileSizeBytes = new(int64)
	*export.FileSizeBytes = int64(buf.Len())
	export.Checksum = &checksum
	now := time.Now().UTC()
	export.CompletedAt = &now

	s.exportRepo.Update(ctx, export)
	s.log.Info("Export completed", "exportID", export.ID, "size", *export.FileSizeBytes)
}

func (s *ExportService) GetExport(ctx context.Context, exportID uuid.UUID) (*domain.AuditExport, error) {
	return s.exportRepo.GetByID(ctx, exportID)
}

func (s *ExportService) ListExports(ctx context.Context, tenantID uuid.UUID) ([]*domain.AuditExport, error) {
	return s.exportRepo.ListByTenant(ctx, tenantID)
}

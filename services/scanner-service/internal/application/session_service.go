package application

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/scanner-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/scanner-service/internal/ports"
)

type SessionService struct {
	sessionRepo   ports.ScanSessionRepository
	scanRepo      ports.ScanEventRepository
	queueRepo     ports.OfflineQueueRepository
	deviceRepo    ports.DeviceRepository
	inventorySvc  ports.InventoryServiceClient
	logger        logger.Logger
}

func NewSessionService(
	sessionRepo ports.ScanSessionRepository,
	scanRepo ports.ScanEventRepository,
	queueRepo ports.OfflineQueueRepository,
	deviceRepo ports.DeviceRepository,
	inventorySvc ports.InventoryServiceClient,
	logger logger.Logger,
) *SessionService {
	return &SessionService{
		sessionRepo:   sessionRepo,
		scanRepo:      scanRepo,
		queueRepo:     queueRepo,
		deviceRepo:    deviceRepo,
		inventorySvc:  inventorySvc,
		logger:        logger,
	}
}

func (s *SessionService) StartSession(ctx context.Context, cmd StartScanSessionCommand) (*ScanSessionDTO, error) {
	if cmd.TenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}
	if cmd.UserID == "" {
		return nil, domain.NewDomainError("USER_REQUIRED", "user ID is required", nil)
	}
	if cmd.Context == "" {
		return nil, domain.NewDomainError("CONTEXT_REQUIRED", "scan context is required", nil)
	}

	scanContext := domain.ScanContext(cmd.Context)

	sessionID := fmt.Sprintf("sess_%s", uuid.New().String()[:12])

	session := domain.NewScanSession(
		sessionID,
		cmd.TenantID,
		cmd.UserID,
		scanContext,
		cmd.DeviceType,
		cmd.DeviceID,
	)
	session.ProjectID = cmd.ProjectID

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, domain.NewDomainError("CREATE_ERROR", "failed to create scan session", err)
	}

	s.logger.Info("Scan session started", "id", session.ID, "context", scanContext)
	return ScanSessionToDTO(session), nil
}

func (s *SessionService) EndSession(ctx context.Context, cmd EndScanSessionCommand) (*ScanSessionDTO, error) {
	if cmd.TenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}
	if cmd.SessionID == "" {
		return nil, domain.NewDomainError("SESSION_ID_REQUIRED", "session ID is required", nil)
	}

	session, err := s.sessionRepo.GetByID(ctx, cmd.TenantID, cmd.SessionID)
	if err != nil {
		return nil, domain.NewDomainError("NOT_FOUND", "scan session not found", err)
	}

	session.End()

	if err := s.sessionRepo.Update(ctx, session); err != nil {
		return nil, domain.NewDomainError("UPDATE_ERROR", "failed to end scan session", err)
	}

	s.logger.Info("Scan session ended", "id", session.ID, "total_scans", session.TotalScans)
	return ScanSessionToDTO(session), nil
}

func (s *SessionService) ProcessSessionScan(ctx context.Context, cmd ProcessSessionScanCommand) (*ScanResultDTO, error) {
	if cmd.TenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}
	if cmd.SessionID == "" {
		return nil, domain.NewDomainError("SESSION_ID_REQUIRED", "session ID is required", nil)
	}
	if cmd.Barcode == "" {
		return nil, domain.NewDomainError("BARCODE_REQUIRED", "barcode is required", nil)
	}

	session, err := s.sessionRepo.GetByID(ctx, cmd.TenantID, cmd.SessionID)
	if err != nil {
		return nil, domain.NewDomainError("NOT_FOUND", "scan session not found", err)
	}

	if session.EndedAt != nil {
		return nil, domain.NewDomainError("SESSION_ENDED", "scan session has already ended", nil)
	}

	session.IncrementTotal()

	scanID := fmt.Sprintf("scan_%s", uuid.New().String()[:12])
	scanType := domain.ScanCheckOut
	if session.Context == domain.ContextCheckIn {
		scanType = domain.ScanCheckIn
	} else if session.Context == domain.ContextInventory {
		scanType = domain.ScanInventory
	} else if session.Context == domain.ContextWarehouseStore {
		scanType = domain.ScanMovement
	}

	event := domain.NewScanEvent(
		scanID,
		cmd.TenantID,
		cmd.Barcode,
		scanType,
		session.UserID,
		cmd.DeviceID,
		cmd.DeviceType,
	)
	event.SessionID = cmd.SessionID
	event.ProjectID = session.ProjectID
	event.Latitude = cmd.Latitude
	event.Longitude = cmd.Longitude
	event.Notes = cmd.Notes

	if err := event.Validate(); err != nil {
		session.IncrementFailed()
		s.sessionRepo.Update(ctx, session)
		return nil, domain.NewDomainError("VALIDATION_ERROR", err.Error(), nil)
	}

	var equipmentID string
	result := "success"
	message := "Scan processed successfully"

	if s.inventorySvc != nil {
		eqID, err := s.inventorySvc.ResolveBarcode(ctx, cmd.TenantID, cmd.Barcode)
		if err != nil {
			result = "warning"
			message = "Barcode resolved but equipment lookup pending"
			s.logger.Warn("Failed to resolve barcode", "barcode", cmd.Barcode, "error", err)
		} else {
			equipmentID = eqID
			event.MarkProcessed(equipmentID)
		}
	}

	if err := s.scanRepo.Create(ctx, event); err != nil {
		session.IncrementFailed()
		s.sessionRepo.Update(ctx, session)
		return nil, domain.NewDomainError("CREATE_ERROR", "failed to create scan event", err)
	}

	session.IncrementSuccessful()
	if err := s.sessionRepo.Update(ctx, session); err != nil {
		s.logger.Error("Failed to update session counters", err)
	}

	s.logger.Info("Session scan processed", "id", event.ID, "barcode", cmd.Barcode)

	return &ScanResultDTO{
		ID:          event.ID,
		SessionID:   cmd.SessionID,
		Barcode:     cmd.Barcode,
		EquipmentID: equipmentID,
		Result:      result,
		Message:     message,
		Timestamp:   event.Timestamp.Format("2006-01-02T15:04:05Z"),
	}, nil
}

func (s *SessionService) GetSessionProtocol(ctx context.Context, tenantID, sessionID string) ([]ScanEventDTO, error) {
	if tenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}
	if sessionID == "" {
		return nil, domain.NewDomainError("SESSION_ID_REQUIRED", "session ID is required", nil)
	}

	// Verify session exists
	_, err := s.sessionRepo.GetByID(ctx, tenantID, sessionID)
	if err != nil {
		return nil, domain.NewDomainError("NOT_FOUND", "scan session not found", err)
	}

	query := &ports.ScanListQuery{
		SessionID: &sessionID,
		Limit:     1000,
		Offset:    0,
	}
	result, err := s.scanRepo.List(ctx, tenantID, query)
	if err != nil {
		return nil, domain.NewDomainError("QUERY_ERROR", "failed to get session protocol", err)
	}

	dtos := make([]ScanEventDTO, len(result.Items))
	for i, event := range result.Items {
		dtos[i] = *ScanEventToDTO(event)
	}

	return dtos, nil
}

func (s *SessionService) UploadSignature(ctx context.Context, tenantID, sessionID, signatureData string) error {
	if tenantID == "" {
		return domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}
	if sessionID == "" {
		return domain.NewDomainError("SESSION_ID_REQUIRED", "session ID is required", nil)
	}

	session, err := s.sessionRepo.GetByID(ctx, tenantID, sessionID)
	if err != nil {
		return domain.NewDomainError("NOT_FOUND", "scan session not found", err)
	}

	session.SignatureData = signatureData
	session.UpdatedAt = time.Now()
	if err := s.sessionRepo.Update(ctx, session); err != nil {
		return domain.NewDomainError("UPDATE_ERROR", "failed to save signature", err)
	}

	s.logger.Info("Signature uploaded for session", "id", sessionID)
	return nil
}

func (s *SessionService) QueueOfflineScan(ctx context.Context, tenantID, deviceID string, scanCmd ProcessScanCommand) (*OfflineQueueDTO, error) {
	if tenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}
	if deviceID == "" {
		return nil, domain.NewDomainError("DEVICE_ID_REQUIRED", "device ID is required", nil)
	}

	// Check queue size
	count, err := s.queueRepo.GetCount(ctx, tenantID, "pending")
	if err != nil {
		s.logger.Error("Failed to check queue size", err)
	} else if count >= 500 {
		return nil, domain.NewDomainError("QUEUE_FULL", "offline queue is full, please sync now", nil)
	}

	payload, err := json.Marshal(scanCmd)
	if err != nil {
		return nil, domain.NewDomainError("MARSHAL_ERROR", "failed to marshal scan command", err)
	}

	queueItemID := fmt.Sprintf("queue_%s", uuid.New().String()[:12])
	item := domain.NewOfflineQueueItem(
		queueItemID,
		tenantID,
		deviceID,
		string(payload),
	)

	if err := s.queueRepo.Create(ctx, item); err != nil {
		return nil, domain.NewDomainError("CREATE_ERROR", "failed to queue offline scan", err)
	}

	s.logger.Info("Offline scan queued", "id", item.ID, "barcode", scanCmd.Barcode)
	return OfflineQueueToDTO(item), nil
}

func (s *SessionService) SyncOfflineQueue(ctx context.Context, tenantID string, limit int) (*SyncResultDTO, error) {
	if tenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	if limit <= 0 || limit > 500 {
		limit = 100
	}

	items, err := s.queueRepo.GetPending(ctx, tenantID, limit)
	if err != nil {
		return nil, domain.NewDomainError("QUERY_ERROR", "failed to get pending offline scans", err)
	}

	result := &SyncResultDTO{
		TotalItems: len(items),
		Conflicts:  []string{},
	}

	for _, item := range items {
		var scanCmd ProcessScanCommand
		if err := json.Unmarshal([]byte(item.Payload), &scanCmd); err != nil {
			item.MarkFailed()
			s.queueRepo.Update(ctx, item)
			result.FailedItems++
			result.Conflicts = append(result.Conflicts, fmt.Sprintf("%s: unmarshal error", item.ID))
			continue
		}

		scanCmd.TenantID = tenantID
		_, err := s.processScanDirect(ctx, scanCmd)
		if err != nil {
			item.MarkFailed()
			s.queueRepo.Update(ctx, item)
			result.FailedItems++
			result.Conflicts = append(result.Conflicts, fmt.Sprintf("%s: %v", item.ID, err))
		} else {
			item.MarkSynced()
			s.queueRepo.Update(ctx, item)
			result.SyncedItems++
		}
	}

	if result.SyncedItems > 0 {
		result.Message = fmt.Sprintf("Successfully synced %d of %d offline scans", result.SyncedItems, result.TotalItems)
	} else {
		result.Message = "No offline scans to sync"
	}

	s.logger.Info("Offline queue synced", "total", result.TotalItems, "synced", result.SyncedItems, "failed", result.FailedItems)
	return result, nil
}

func (s *SessionService) processScanDirect(ctx context.Context, cmd ProcessScanCommand) (*ScanEventDTO, error) {
	if cmd.TenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}
	if cmd.Barcode == "" {
		return nil, domain.NewDomainError("BARCODE_REQUIRED", "barcode is required", nil)
	}

	scanID := fmt.Sprintf("scan_%s", uuid.New().String()[:12])

	event := domain.NewScanEvent(
		scanID,
		cmd.TenantID,
		cmd.Barcode,
		cmd.ScanType,
		cmd.UserID,
		cmd.DeviceID,
		cmd.DeviceType,
	)

	event.ProjectID = cmd.ProjectID
	event.LocationID = cmd.LocationID
	event.Latitude = cmd.Latitude
	event.Longitude = cmd.Longitude
	event.Notes = cmd.Notes

	if err := event.Validate(); err != nil {
		return nil, domain.NewDomainError("VALIDATION_ERROR", err.Error(), nil)
	}

	if s.inventorySvc != nil {
		equipmentID, err := s.inventorySvc.ResolveBarcode(ctx, cmd.TenantID, cmd.Barcode)
		if err == nil {
			event.MarkProcessed(equipmentID)
		} else {
			s.logger.Warn("Failed to resolve barcode", "barcode", cmd.Barcode, "error", err)
		}
	}

	if err := s.scanRepo.Create(ctx, event); err != nil {
		return nil, domain.NewDomainError("CREATE_ERROR", "failed to create scan event", err)
	}

	return ScanEventToDTO(event), nil
}

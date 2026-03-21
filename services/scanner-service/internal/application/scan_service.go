package application

import (
	"context"
	"fmt"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/scanner-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/scanner-service/internal/ports"
)

type ScanService struct {
	scanRepo     ports.ScanEventRepository
	deviceRepo   ports.DeviceRepository
	inventorySvc ports.InventoryServiceClient
	logger       logger.Logger
}

func NewScanService(
	scanRepo ports.ScanEventRepository,
	deviceRepo ports.DeviceRepository,
	inventorySvc ports.InventoryServiceClient,
	logger logger.Logger,
) *ScanService {
	return &ScanService{
		scanRepo:     scanRepo,
		deviceRepo:   deviceRepo,
		inventorySvc: inventorySvc,
		logger:       logger,
	}
}

func (s *ScanService) ProcessScan(ctx context.Context, cmd ProcessScanCommand) (*ScanEventDTO, error) {
	if cmd.TenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}
	if cmd.Barcode == "" {
		return nil, domain.NewDomainError("BARCODE_REQUIRED", "barcode is required", nil)
	}

	// Generate scan ID
	scanID := fmt.Sprintf("scan_%d", hashString(cmd.TenantID+cmd.Barcode+time.Now().String()))

	// Create scan event
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

	// Validate
	if err := event.Validate(); err != nil {
		return nil, domain.NewDomainError("VALIDATION_ERROR", err.Error(), nil)
	}

	// Try to resolve barcode via inventory service
	if s.inventorySvc != nil {
		equipmentID, err := s.inventorySvc.ResolveBarcode(ctx, cmd.TenantID, cmd.Barcode)
		if err == nil {
			event.MarkProcessed(equipmentID)
		} else {
			s.logger.Warn("Failed to resolve barcode", "barcode", cmd.Barcode, "error", err)
		}
	}

	// Persist
	if err := s.scanRepo.Create(ctx, event); err != nil {
		return nil, domain.NewDomainError("CREATE_ERROR", "failed to create scan event", err)
	}

	s.logger.Info("Scan processed", "id", event.ID, "barcode", cmd.Barcode)
	return ScanEventToDTO(event), nil
}

func (s *ScanService) ProcessBatch(ctx context.Context, cmd BatchScanCommand) ([]ScanEventDTO, error) {
	if cmd.TenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}
	if len(cmd.Scans) == 0 {
		return nil, domain.NewDomainError("INVALID_INPUT", "scans array is empty", nil)
	}

	results := make([]ScanEventDTO, 0, len(cmd.Scans))

	for _, scanCmd := range cmd.Scans {
		scanCmd.TenantID = cmd.TenantID
		dto, err := s.ProcessScan(ctx, scanCmd)
		if err != nil {
			s.logger.Error("Failed to process scan in batch", err)
			continue
		}
		results = append(results, *dto)
	}

	return results, nil
}

func (s *ScanService) GetHistory(ctx context.Context, query ScanHistoryQuery) (*PaginatedResult, error) {
	if query.TenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	repoQuery := &ports.ScanListQuery{
		EquipmentID: query.EquipmentID,
		ProjectID:   query.ProjectID,
		UserID:      query.UserID,
		DeviceID:    query.DeviceID,
		Status:      query.Status,
		Limit:       query.Limit,
		Offset:      query.Offset,
	}

	result, err := s.scanRepo.List(ctx, query.TenantID, repoQuery)
	if err != nil {
		return nil, domain.NewDomainError("QUERY_ERROR", "failed to get scan history", err)
	}

	dtos := make([]*ScanEventDTO, len(result.Items))
	for i, event := range result.Items {
		dtos[i] = ScanEventToDTO(event)
	}

	return &PaginatedResult{
		Data:   dtos,
		Total:  result.Total,
		Limit:  result.Limit,
		Offset: result.Offset,
	}, nil
}

func (s *ScanService) ResolveBarcode(ctx context.Context, tenantID, barcode string) (*ScanEventDTO, error) {
	if tenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}
	if barcode == "" {
		return nil, domain.NewDomainError("BARCODE_REQUIRED", "barcode is required", nil)
	}

	if s.inventorySvc == nil {
		return nil, domain.NewDomainError("SERVICE_UNAVAILABLE", "inventory service not available", nil)
	}

	equipmentID, err := s.inventorySvc.ResolveBarcode(ctx, tenantID, barcode)
	if err != nil {
		return nil, domain.NewDomainError("NOT_FOUND", "barcode not found", err)
	}

	// Return a synthetic scan event with resolved equipment
	event := &domain.ScanEvent{
		ID:          equipmentID,
		TenantID:    tenantID,
		Barcode:     barcode,
		EquipmentID: equipmentID,
		Status:      domain.ScanProcessed,
	}

	return ScanEventToDTO(event), nil
}

func (s *ScanService) SyncOfflineScans(ctx context.Context, cmd SyncOfflineCommand) ([]ScanEventDTO, error) {
	if cmd.TenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	results := make([]ScanEventDTO, 0, len(cmd.Scans))

	for _, scanCmd := range cmd.Scans {
		scanCmd.TenantID = cmd.TenantID
		dto, err := s.ProcessScan(ctx, scanCmd)
		if err != nil {
			s.logger.Error("Failed to sync offline scan", err)
			continue
		}
		results = append(results, *dto)
	}

	return results, nil
}

func (s *ScanService) RegisterDevice(ctx context.Context, cmd RegisterDeviceCommand) (*DeviceDTO, error) {
	if cmd.TenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}
	if cmd.Name == "" {
		return nil, domain.NewDomainError("NAME_REQUIRED", "device name is required", nil)
	}

	// Check if serial already exists
	existing, _ := s.deviceRepo.GetBySerial(ctx, cmd.TenantID, cmd.Serial)
	if existing != nil {
		return nil, domain.NewDomainError("DEVICE_EXISTS", "device with this serial already exists", nil)
	}

	deviceID := fmt.Sprintf("dev_%d", hashString(cmd.TenantID+cmd.Serial))
	device := domain.NewDevice(deviceID, cmd.TenantID, cmd.Name, cmd.Type, cmd.Serial)
	device.Location = cmd.Location

	if err := s.deviceRepo.Create(ctx, device); err != nil {
		return nil, domain.NewDomainError("CREATE_ERROR", "failed to register device", err)
	}

	s.logger.Info("Device registered", "id", device.ID, "name", cmd.Name)
	return DeviceToDTO(device), nil
}

func (s *ScanService) ListDevices(ctx context.Context, tenantID string, limit, offset int) (*PaginatedResult, error) {
	if tenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	devices, total, err := s.deviceRepo.List(ctx, tenantID, limit, offset)
	if err != nil {
		return nil, domain.NewDomainError("QUERY_ERROR", "failed to list devices", err)
	}

	dtos := make([]*DeviceDTO, len(devices))
	for i, device := range devices {
		dtos[i] = DeviceToDTO(device)
	}

	return &PaginatedResult{
		Data:   dtos,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}, nil
}

func hashString(s string) int64 {
	h := int64(5381)
	for _, c := range s {
		h = ((h << 5) + h) + int64(c)
	}
	return h & 0x7FFFFFFFFFFFFFFF
}

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
	scanRepo          ports.ScanEventRepository
	deviceRepo        ports.DeviceRepository
	scannerDeviceRepo ports.ScannerDeviceRepository
	inventorySvc      ports.InventoryServiceClient
	logger            logger.Logger
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

// SetScannerDeviceRepo sets the scanner device repository (optional dependency)
func (s *ScanService) SetScannerDeviceRepo(repo ports.ScannerDeviceRepository) {
	s.scannerDeviceRepo = repo
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

// =====================================================
// Scanner App API contract service methods
// =====================================================

// ResolveEquipment resolves a barcode/RFID code to full equipment detail.
// Implements POST /api/v1/scanner/scan business logic.
func (s *ScanService) ResolveEquipment(ctx context.Context, tenantID, code, scanType string) (*ports.EquipmentDetail, error) {
	if tenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}
	if code == "" {
		return nil, domain.NewDomainError("BARCODE_REQUIRED", "code is required", nil)
	}
	if s.inventorySvc == nil {
		return nil, domain.NewDomainError("SERVICE_UNAVAILABLE", "inventory service not available", nil)
	}

	equipment, err := s.inventorySvc.ResolveEquipment(ctx, tenantID, code, scanType)
	if err != nil {
		return nil, domain.NewDomainError("NOT_FOUND", "equipment not found for given code", err)
	}

	return equipment, nil
}

// CheckOutEquipment checks out equipment to a project.
// Implements POST /api/v1/scanner/checkout business logic.
func (s *ScanService) CheckOutEquipment(ctx context.Context, tenantID string, equipmentIDs []string, projectID, notes string) (*ports.CheckoutResult, error) {
	if tenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}
	if len(equipmentIDs) == 0 {
		return nil, domain.NewDomainError("INVALID_INPUT", "equipment_ids is required", nil)
	}
	if projectID == "" {
		return nil, domain.NewDomainError("INVALID_INPUT", "project_id is required", nil)
	}
	if s.inventorySvc == nil {
		return nil, domain.NewDomainError("SERVICE_UNAVAILABLE", "inventory service not available", nil)
	}

	result, err := s.inventorySvc.CheckOutEquipment(ctx, tenantID, equipmentIDs, projectID, notes)
	if err != nil {
		return nil, domain.NewDomainError("CHECKOUT_ERROR", "failed to check out equipment", err)
	}

	return result, nil
}

// CheckInEquipment checks in equipment with condition ratings.
// Implements POST /api/v1/scanner/checkin business logic.
func (s *ScanService) CheckInEquipment(ctx context.Context, tenantID string, equipmentIDs []string, conditionRatings map[string]ConditionRatingPayload) (*ports.CheckinResult, error) {
	if tenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}
	if len(equipmentIDs) == 0 {
		return nil, domain.NewDomainError("INVALID_INPUT", "equipment_ids is required", nil)
	}
	if s.inventorySvc == nil {
		return nil, domain.NewDomainError("SERVICE_UNAVAILABLE", "inventory service not available", nil)
	}

	// Convert payload ratings to port ratings
	portRatings := make(map[string]ports.ConditionRating, len(conditionRatings))
	for k, v := range conditionRatings {
		portRatings[k] = ports.ConditionRating{
			Rating:         v.Rating,
			Notes:          v.Notes,
			DamageReported: v.DamageReported,
		}
	}

	result, err := s.inventorySvc.CheckInEquipment(ctx, tenantID, equipmentIDs, portRatings)
	if err != nil {
		return nil, domain.NewDomainError("CHECKIN_ERROR", "failed to check in equipment", err)
	}

	return result, nil
}

// ProcessBulkActions processes a batch of offline checkout/checkin actions idempotently.
// Implements POST /api/v1/scanner/bulk business logic.
func (s *ScanService) ProcessBulkActions(ctx context.Context, tenantID string, actions []BulkAction) (*BulkResult, error) {
	if tenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	result := &BulkResult{
		Results: make([]BulkActionResult, 0, len(actions)),
	}

	for _, action := range actions {
		ar := BulkActionResult{Type: action.Type}

		switch action.Type {
		case "checkout":
			coResult, err := s.CheckOutEquipment(ctx, tenantID, action.EquipmentIDs, action.ProjectID, action.Notes)
			if err != nil {
				ar.Success = false
				ar.Error = err.Error()
				result.Failed++
			} else {
				ar.Success = true
				ar.Data = map[string]interface{}{
					"checked_out": coResult.CheckedOut,
					"project":     coResult.Project,
				}
				result.Processed++
			}

		case "checkin":
			ciResult, err := s.CheckInEquipment(ctx, tenantID, action.EquipmentIDs, action.ConditionRatings)
			if err != nil {
				ar.Success = false
				ar.Error = err.Error()
				result.Failed++
			} else {
				ar.Success = true
				ar.Data = map[string]interface{}{
					"checked_in":     ciResult.CheckedIn,
					"damage_reports": ciResult.DamageReports,
				}
				result.Processed++
			}

		default:
			ar.Success = false
			ar.Error = "unknown action type: " + action.Type
			result.Failed++
		}

		result.Results = append(result.Results, ar)
	}

	return result, nil
}

// AdhocBooking creates a reservation on-the-fly from the scanner.
// Implements POST /api/v1/scanner/adhoc-booking business logic.
func (s *ScanService) AdhocBooking(ctx context.Context, cmd AdhocBookingCommand) (*AdhocBookingResult, error) {
	if cmd.TenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}
	if cmd.EquipmentID == "" {
		return nil, domain.NewDomainError("INVALID_INPUT", "equipment_id is required", nil)
	}
	if cmd.ProjectID == "" {
		return nil, domain.NewDomainError("INVALID_INPUT", "project_id is required", nil)
	}
	if s.inventorySvc == nil {
		return nil, domain.NewDomainError("SERVICE_UNAVAILABLE", "inventory service not available", nil)
	}

	// Use checkout to create the ad-hoc assignment
	coResult, err := s.inventorySvc.CheckOutEquipment(ctx, cmd.TenantID, []string{cmd.EquipmentID}, cmd.ProjectID, cmd.Notes)
	if err != nil {
		return nil, domain.NewDomainError("CHECKOUT_ERROR", "failed to create ad-hoc booking", err)
	}

	s.logger.Info("Ad-hoc booking created", "equipment_id", cmd.EquipmentID, "project_id", cmd.ProjectID)

	return &AdhocBookingResult{
		Success:     true,
		EquipmentID: cmd.EquipmentID,
		ProjectID:   cmd.ProjectID,
		CheckedOut:  coResult.CheckedOut,
		Notes:       cmd.Notes,
	}, nil
}

// =====================================================
// Scanner Device Management ("Find My Scanner")
// =====================================================

// RegisterScannerDevice registers or updates a scanner device
func (s *ScanService) RegisterScannerDevice(ctx context.Context, tenantID, deviceID, deviceName, fcmToken string) (*ports.ScannerDevice, error) {
	if tenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}
	if deviceID == "" {
		return nil, domain.NewDomainError("INVALID_INPUT", "device_id is required", nil)
	}
	if deviceName == "" {
		return nil, domain.NewDomainError("INVALID_INPUT", "device_name is required", nil)
	}
	if s.scannerDeviceRepo == nil {
		return nil, domain.NewDomainError("SERVICE_UNAVAILABLE", "scanner device management not available", nil)
	}

	var fcmPtr *string
	if fcmToken != "" {
		fcmPtr = &fcmToken
	}

	device := &ports.ScannerDevice{
		TenantID:   tenantID,
		DeviceID:   deviceID,
		DeviceName: deviceName,
		DeviceType: "handheld",
		FCMToken:   fcmPtr,
	}

	if err := s.scannerDeviceRepo.Upsert(ctx, device); err != nil {
		return nil, domain.NewDomainError("CREATE_ERROR", "failed to register scanner device", err)
	}

	s.logger.Info("Scanner device registered", "device_id", deviceID, "name", deviceName)
	return device, nil
}

// ListScannerDevices returns all registered scanner devices for a tenant
func (s *ScanService) ListScannerDevices(ctx context.Context, tenantID string) ([]*ports.ScannerDevice, error) {
	if tenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}
	if s.scannerDeviceRepo == nil {
		return nil, domain.NewDomainError("SERVICE_UNAVAILABLE", "scanner device management not available", nil)
	}

	devices, err := s.scannerDeviceRepo.List(ctx, tenantID)
	if err != nil {
		return nil, domain.NewDomainError("QUERY_ERROR", "failed to list scanner devices", err)
	}

	if devices == nil {
		devices = []*ports.ScannerDevice{}
	}

	return devices, nil
}

// RingScannerDevice sets ring_requested=true for a device (webapp triggers)
func (s *ScanService) RingScannerDevice(ctx context.Context, id string) error {
	if id == "" {
		return domain.NewDomainError("INVALID_INPUT", "device id is required", nil)
	}
	if s.scannerDeviceRepo == nil {
		return domain.NewDomainError("SERVICE_UNAVAILABLE", "scanner device management not available", nil)
	}

	if err := s.scannerDeviceRepo.SetRingRequested(ctx, id, true); err != nil {
		return domain.NewDomainError("UPDATE_ERROR", "failed to set ring request", err)
	}

	s.logger.Info("Ring requested for scanner device", "id", id)
	return nil
}

// CheckRingRequest checks if ring is requested for a device (app polls this)
func (s *ScanService) CheckRingRequest(ctx context.Context, tenantID, deviceID string) (bool, error) {
	if s.scannerDeviceRepo == nil {
		return false, domain.NewDomainError("SERVICE_UNAVAILABLE", "scanner device management not available", nil)
	}

	// Also update last_seen
	_ = s.scannerDeviceRepo.UpdateLastSeen(ctx, tenantID, deviceID)

	device, err := s.scannerDeviceRepo.GetByDeviceID(ctx, tenantID, deviceID)
	if err != nil {
		return false, domain.NewDomainError("NOT_FOUND", "scanner device not found", err)
	}

	return device.RingRequested, nil
}

// AckRing acknowledges the ring request (app sends after playing sound)
func (s *ScanService) AckRing(ctx context.Context, tenantID, deviceID string) error {
	if s.scannerDeviceRepo == nil {
		return domain.NewDomainError("SERVICE_UNAVAILABLE", "scanner device management not available", nil)
	}

	device, err := s.scannerDeviceRepo.GetByDeviceID(ctx, tenantID, deviceID)
	if err != nil {
		return domain.NewDomainError("NOT_FOUND", "scanner device not found", err)
	}

	if err := s.scannerDeviceRepo.SetRingRequested(ctx, device.ID, false); err != nil {
		return domain.NewDomainError("UPDATE_ERROR", "failed to acknowledge ring", err)
	}

	s.logger.Info("Ring acknowledged for scanner device", "device_id", deviceID)
	return nil
}

func hashString(s string) int64 {
	h := int64(5381)
	for _, c := range s {
		h = ((h << 5) + h) + int64(c)
	}
	return h & 0x7FFFFFFFFFFFFFFF
}

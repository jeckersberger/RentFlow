package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/services/scanner/internal/domain"
)

// ---------------------------------------------------------------------------
// Request DTOs
// ---------------------------------------------------------------------------

// RegisterDeviceRequest holds the data needed to register a scanner device.
type RegisterDeviceRequest struct {
	DeviceID   string `json:"device_id"`
	DeviceName string `json:"device_name,omitempty"`
	DeviceType string `json:"device_type,omitempty"`
	FCMToken   string `json:"fcm_token,omitempty"`
}

// RingAckRequest holds the data for acknowledging a ring.
type RingAckRequest struct {
	DeviceID string `json:"device_id"`
}

// ---------------------------------------------------------------------------
// Service
// ---------------------------------------------------------------------------

// DeviceService implements the application-level use cases for scanner devices.
type DeviceService struct {
	deviceRepo domain.ScannerDeviceRepository
	logger     zerolog.Logger
}

// NewDeviceService constructs a new DeviceService.
func NewDeviceService(
	deviceRepo domain.ScannerDeviceRepository,
	logger zerolog.Logger,
) *DeviceService {
	return &DeviceService{
		deviceRepo: deviceRepo,
		logger:     logger.With().Str("service", "device").Logger(),
	}
}

// Register registers or updates a scanner device for a tenant.
func (s *DeviceService) Register(
	ctx context.Context,
	tenantID uuid.UUID,
	req RegisterDeviceRequest,
) (*domain.ScannerDevice, error) {
	if req.DeviceID == "" {
		return nil, domain.ErrMissingDeviceID
	}

	deviceType := req.DeviceType
	if deviceType == "" {
		deviceType = "cf-h906"
	}

	now := time.Now()
	device := &domain.ScannerDevice{
		ID:         uuid.New(),
		TenantID:   tenantID,
		DeviceID:   req.DeviceID,
		DeviceName: req.DeviceName,
		DeviceType: deviceType,
		FCMToken:   req.FCMToken,
		LastSeen:   &now,
	}

	if err := s.deviceRepo.Upsert(ctx, device); err != nil {
		s.logger.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Str("device_id", req.DeviceID).
			Msg("failed to register device")
		return nil, fmt.Errorf("register device: %w", err)
	}

	s.logger.Info().
		Str("device_id", device.DeviceID).
		Str("tenant_id", tenantID.String()).
		Msg("device registered")

	return device, nil
}

// List returns all scanner devices for a tenant.
func (s *DeviceService) List(
	ctx context.Context,
	tenantID uuid.UUID,
) ([]*domain.ScannerDevice, error) {
	devices, err := s.deviceRepo.List(ctx, tenantID)
	if err != nil {
		s.logger.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Msg("failed to list devices")
		return nil, fmt.Errorf("list devices: %w", err)
	}
	return devices, nil
}

// CheckRing checks whether a ring has been requested for a device (polling endpoint).
func (s *DeviceService) CheckRing(
	ctx context.Context,
	tenantID uuid.UUID,
	deviceID string,
) (bool, error) {
	if deviceID == "" {
		return false, domain.ErrMissingDeviceID
	}

	device, err := s.deviceRepo.GetByDeviceID(ctx, tenantID, deviceID)
	if err != nil {
		s.logger.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Str("device_id", deviceID).
			Msg("failed to check ring status")
		return false, fmt.Errorf("check ring: %w", err)
	}

	// Update last_seen on every poll.
	_ = s.deviceRepo.UpdateLastSeen(ctx, tenantID, deviceID)

	return device.RingRequested, nil
}

// AckRing acknowledges and clears the ring request for a device.
func (s *DeviceService) AckRing(
	ctx context.Context,
	tenantID uuid.UUID,
	req RingAckRequest,
) error {
	if req.DeviceID == "" {
		return domain.ErrMissingDeviceID
	}

	device, err := s.deviceRepo.GetByDeviceID(ctx, tenantID, req.DeviceID)
	if err != nil {
		s.logger.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Str("device_id", req.DeviceID).
			Msg("failed to find device for ring ack")
		return fmt.Errorf("ack ring: %w", err)
	}

	if err := s.deviceRepo.SetRingRequested(ctx, device.ID, tenantID, false); err != nil {
		s.logger.Error().Err(err).
			Str("device_id", req.DeviceID).
			Msg("failed to ack ring")
		return fmt.Errorf("ack ring: %w", err)
	}

	s.logger.Info().
		Str("device_id", req.DeviceID).
		Str("tenant_id", tenantID.String()).
		Msg("ring acknowledged")

	return nil
}

// TriggerRing sets the ring_requested flag for a device (triggered from webapp).
func (s *DeviceService) TriggerRing(
	ctx context.Context,
	id uuid.UUID,
	tenantID uuid.UUID,
) error {
	if err := s.deviceRepo.SetRingRequested(ctx, id, tenantID, true); err != nil {
		s.logger.Error().Err(err).
			Str("device_uuid", id.String()).
			Str("tenant_id", tenantID.String()).
			Msg("failed to trigger ring")
		return fmt.Errorf("trigger ring: %w", err)
	}

	s.logger.Info().
		Str("device_uuid", id.String()).
		Str("tenant_id", tenantID.String()).
		Msg("ring triggered")

	return nil
}

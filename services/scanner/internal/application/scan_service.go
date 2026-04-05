package application

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/services/scanner/internal/domain"
)

// ---------------------------------------------------------------------------
// Request / Response DTOs
// ---------------------------------------------------------------------------

// ScanRequest holds the data for a single scan lookup.
type ScanRequest struct {
	DeviceID string `json:"device_id"`
	Barcode  string `json:"barcode,omitempty"`
	RFIDTag  string `json:"rfid_tag,omitempty"`
}

// CheckoutRequest holds the data for checking out equipment to a project (bulk).
type CheckoutRequest struct {
	DeviceID  string         `json:"device_id"`
	ProjectID uuid.UUID      `json:"project_id"`
	Items     []CheckoutItem `json:"items"`
}

// CheckoutItem represents a single item in a checkout request.
type CheckoutItem struct {
	EquipmentID uuid.UUID `json:"equipment_id"`
	Barcode     string    `json:"barcode,omitempty"`
	RFIDTag     string    `json:"rfid_tag,omitempty"`
}

// CheckinRequest holds the data for checking in equipment (bulk, with optional rating).
type CheckinRequest struct {
	DeviceID  string        `json:"device_id"`
	ProjectID *uuid.UUID    `json:"project_id,omitempty"`
	Items     []CheckinItem `json:"items"`
}

// CheckinItem represents a single item in a checkin request.
type CheckinItem struct {
	EquipmentID     uuid.UUID `json:"equipment_id"`
	Barcode         string    `json:"barcode,omitempty"`
	RFIDTag         string    `json:"rfid_tag,omitempty"`
	LocationID      *uuid.UUID `json:"location_id,omitempty"`
	ConditionRating *int      `json:"condition_rating,omitempty"`
	ConditionNotes  string    `json:"condition_notes,omitempty"`
}

// BulkSyncRequest holds an array of offline scan events to sync.
type BulkSyncRequest struct {
	Events []BulkScanEvent `json:"events"`
}

// BulkScanEvent represents a single event from the offline queue.
type BulkScanEvent struct {
	EventID         string     `json:"event_id,omitempty"` // Client-generated UUID for idempotency
	DeviceID        string     `json:"device_id"`
	Barcode         string     `json:"barcode,omitempty"`
	RFIDTag         string     `json:"rfid_tag,omitempty"`
	EquipmentID     *uuid.UUID `json:"equipment_id,omitempty"`
	Action          string     `json:"action"`
	ProjectID       *uuid.UUID `json:"project_id,omitempty"`
	LocationID      *uuid.UUID `json:"location_id,omitempty"`
	ConditionRating *int       `json:"condition_rating,omitempty"`
	ConditionNotes  string     `json:"condition_notes,omitempty"`
	GPSLat          *float64   `json:"gps_lat,omitempty"`
	GPSLng          *float64   `json:"gps_lng,omitempty"`
	Timestamp       time.Time  `json:"timestamp"`
}

// BulkSyncResponse reports how many events were inserted vs skipped.
type BulkSyncResponse struct {
	Inserted int `json:"inserted"`
	Skipped  int `json:"skipped"`
	Total    int `json:"total"`
}

// AdhocBookingRequest holds the data for a spontaneous booking.
type AdhocBookingRequest struct {
	DeviceID        string     `json:"device_id"`
	Barcode         string     `json:"barcode,omitempty"`
	RFIDTag         string     `json:"rfid_tag,omitempty"`
	EquipmentID     *uuid.UUID `json:"equipment_id,omitempty"`
	ProjectID       *uuid.UUID `json:"project_id,omitempty"`
	LocationID      *uuid.UUID `json:"location_id,omitempty"`
	ConditionRating *int       `json:"condition_rating,omitempty"`
	ConditionNotes  string     `json:"condition_notes,omitempty"`
	GPSLat          *float64   `json:"gps_lat,omitempty"`
	GPSLng          *float64   `json:"gps_lng,omitempty"`
}

// ---------------------------------------------------------------------------
// Service
// ---------------------------------------------------------------------------

// ScanService implements the application-level use cases for scanning.
type ScanService struct {
	eventRepo        domain.ScanEventRepository
	deviceRepo       domain.ScannerDeviceRepository
	inventoryBaseURL string // Base URL for inventory service (e.g., "http://inventory:8004")
	logger           zerolog.Logger
}

// NewScanService constructs a new ScanService.
func NewScanService(
	eventRepo domain.ScanEventRepository,
	deviceRepo domain.ScannerDeviceRepository,
	inventoryBaseURL string,
	logger zerolog.Logger,
) *ScanService {
	return &ScanService{
		eventRepo:        eventRepo,
		deviceRepo:       deviceRepo,
		inventoryBaseURL: inventoryBaseURL,
		logger:           logger.With().Str("service", "scan").Logger(),
	}
}

// Scan records a simple scan event (barcode or RFID lookup).
func (s *ScanService) Scan(
	ctx context.Context,
	tenantID uuid.UUID,
	userID uuid.UUID,
	req ScanRequest,
) (*domain.ScanEvent, error) {
	if req.Barcode == "" && req.RFIDTag == "" {
		return nil, domain.ErrMissingIdentifier
	}

	now := time.Now()
	event := &domain.ScanEvent{
		ID:        uuid.New(),
		TenantID:  tenantID,
		UserID:    userID,
		DeviceID:  req.DeviceID,
		Barcode:   req.Barcode,
		RFIDTag:   req.RFIDTag,
		Action:    domain.ActionScan,
		Timestamp: now,
		SyncedAt:  now,
	}

	if err := s.eventRepo.Create(ctx, event); err != nil {
		s.logger.Error().Err(err).Str("tenant_id", tenantID.String()).Msg("failed to create scan event")
		return nil, fmt.Errorf("scan: %w", err)
	}

	// Update device last_seen (best-effort).
	if req.DeviceID != "" {
		_ = s.deviceRepo.UpdateLastSeen(ctx, tenantID, req.DeviceID)
	}

	s.logger.Info().
		Str("event_id", event.ID.String()).
		Str("tenant_id", tenantID.String()).
		Str("action", domain.ActionScan).
		Msg("scan event recorded")

	return event, nil
}

// Checkout records checkout events for multiple equipment items to a project.
func (s *ScanService) Checkout(
	ctx context.Context,
	tenantID uuid.UUID,
	userID uuid.UUID,
	req CheckoutRequest,
) ([]*domain.ScanEvent, error) {
	if req.ProjectID == uuid.Nil {
		return nil, domain.ErrMissingProjectID
	}

	now := time.Now()
	var events []*domain.ScanEvent

	for _, item := range req.Items {
		event := &domain.ScanEvent{
			ID:          uuid.New(),
			TenantID:    tenantID,
			UserID:      userID,
			DeviceID:    req.DeviceID,
			Barcode:     item.Barcode,
			RFIDTag:     item.RFIDTag,
			EquipmentID: &item.EquipmentID,
			Action:      domain.ActionCheckout,
			ProjectID:   &req.ProjectID,
			Timestamp:   now,
			SyncedAt:    now,
		}

		if err := s.eventRepo.Create(ctx, event); err != nil {
			s.logger.Error().Err(err).
				Str("tenant_id", tenantID.String()).
				Str("equipment_id", item.EquipmentID.String()).
				Msg("failed to create checkout event")
			return nil, fmt.Errorf("checkout: %w", err)
		}

		events = append(events, event)
	}

	// Update device last_seen (best-effort).
	if req.DeviceID != "" {
		_ = s.deviceRepo.UpdateLastSeen(ctx, tenantID, req.DeviceID)
	}

	s.logger.Info().
		Str("tenant_id", tenantID.String()).
		Int("count", len(events)).
		Msg("checkout events recorded")

	return events, nil
}

// Checkin records checkin events for multiple equipment items (with optional condition rating).
func (s *ScanService) Checkin(
	ctx context.Context,
	tenantID uuid.UUID,
	userID uuid.UUID,
	req CheckinRequest,
) ([]*domain.ScanEvent, error) {
	now := time.Now()
	var events []*domain.ScanEvent

	for _, item := range req.Items {
		event := &domain.ScanEvent{
			ID:              uuid.New(),
			TenantID:        tenantID,
			UserID:          userID,
			DeviceID:        req.DeviceID,
			Barcode:         item.Barcode,
			RFIDTag:         item.RFIDTag,
			EquipmentID:     &item.EquipmentID,
			Action:          domain.ActionCheckin,
			ProjectID:       req.ProjectID,
			LocationID:      item.LocationID,
			ConditionRating: item.ConditionRating,
			ConditionNotes:  item.ConditionNotes,
			Timestamp:       now,
			SyncedAt:        now,
		}

		if err := s.eventRepo.Create(ctx, event); err != nil {
			s.logger.Error().Err(err).
				Str("tenant_id", tenantID.String()).
				Str("equipment_id", item.EquipmentID.String()).
				Msg("failed to create checkin event")
			return nil, fmt.Errorf("checkin: %w", err)
		}

		// Update equipment location in inventory service (best-effort, non-blocking)
		if item.LocationID != nil && s.inventoryBaseURL != "" {
			go s.updateEquipmentLocation(tenantID, item.EquipmentID, *item.LocationID)
		}

		events = append(events, event)
	}

	// Update device last_seen (best-effort).
	if req.DeviceID != "" {
		_ = s.deviceRepo.UpdateLastSeen(ctx, tenantID, req.DeviceID)
	}

	s.logger.Info().
		Str("tenant_id", tenantID.String()).
		Int("count", len(events)).
		Msg("checkin events recorded")

	return events, nil
}

// BulkSync processes an array of offline scan events with timestamp-based deduplication.
func (s *ScanService) BulkSync(
	ctx context.Context,
	tenantID uuid.UUID,
	userID uuid.UUID,
	req BulkSyncRequest,
) (*BulkSyncResponse, error) {
	inserted := 0
	skipped := 0

	for _, ev := range req.Events {
		// Validate action.
		if !domain.ValidAction(ev.Action) {
			s.logger.Warn().
				Str("action", ev.Action).
				Str("device_id", ev.DeviceID).
				Msg("skipping event with invalid action during bulk sync")
			skipped++
			continue
		}

		// Idempotency: if event_id is provided, check processed_events table first.
		// This ensures replayed offline events are deduplicated correctly.
		if ev.EventID != "" {
			eventUUID, parseErr := uuid.Parse(ev.EventID)
			if parseErr == nil {
				exists, err := s.eventRepo.IsEventProcessed(ctx, eventUUID, tenantID)
				if err != nil {
					s.logger.Error().Err(err).Msg("failed processed event check during bulk sync")
					return nil, fmt.Errorf("bulk sync processed check: %w", err)
				}
				if exists {
					skipped++
					continue
				}
			}
		}

		// Fallback: timestamp-based deduplication for older clients without event_id.
		exists, err := s.eventRepo.ExistsByDedup(ctx, tenantID, ev.DeviceID, ev.Timestamp)
		if err != nil {
			s.logger.Error().Err(err).Msg("failed dedup check during bulk sync")
			return nil, fmt.Errorf("bulk sync dedup check: %w", err)
		}
		if exists {
			skipped++
			continue
		}

		now := time.Now()
		event := &domain.ScanEvent{
			ID:              uuid.New(),
			TenantID:        tenantID,
			UserID:          userID,
			DeviceID:        ev.DeviceID,
			Barcode:         ev.Barcode,
			RFIDTag:         ev.RFIDTag,
			EquipmentID:     ev.EquipmentID,
			Action:          ev.Action,
			ProjectID:       ev.ProjectID,
			LocationID:      ev.LocationID,
			ConditionRating: ev.ConditionRating,
			ConditionNotes:  ev.ConditionNotes,
			GPSLat:          ev.GPSLat,
			GPSLng:          ev.GPSLng,
			Timestamp:       ev.Timestamp,
			SyncedAt:        now,
		}

		if err := s.eventRepo.Create(ctx, event); err != nil {
			s.logger.Error().Err(err).
				Str("device_id", ev.DeviceID).
				Msg("failed to insert event during bulk sync")
			return nil, fmt.Errorf("bulk sync insert: %w", err)
		}

		// Mark event as processed for idempotency (best-effort, non-fatal)
		if ev.EventID != "" {
			eventUUID, parseErr := uuid.Parse(ev.EventID)
			if parseErr == nil {
				_ = s.eventRepo.MarkEventProcessed(ctx, eventUUID, tenantID, ev.Action, ev.Barcode)
			}
		}

		inserted++

		// Update device last_seen (best-effort).
		if ev.DeviceID != "" {
			_ = s.deviceRepo.UpdateLastSeen(ctx, tenantID, ev.DeviceID)
		}
	}

	s.logger.Info().
		Str("tenant_id", tenantID.String()).
		Int("inserted", inserted).
		Int("skipped", skipped).
		Int("total", len(req.Events)).
		Msg("bulk sync completed")

	return &BulkSyncResponse{
		Inserted: inserted,
		Skipped:  skipped,
		Total:    len(req.Events),
	}, nil
}

// AdhocBooking records a spontaneous booking event.
func (s *ScanService) AdhocBooking(
	ctx context.Context,
	tenantID uuid.UUID,
	userID uuid.UUID,
	req AdhocBookingRequest,
) (*domain.ScanEvent, error) {
	if req.Barcode == "" && req.RFIDTag == "" && req.EquipmentID == nil {
		return nil, domain.ErrMissingIdentifier
	}

	now := time.Now()
	event := &domain.ScanEvent{
		ID:              uuid.New(),
		TenantID:        tenantID,
		UserID:          userID,
		DeviceID:        req.DeviceID,
		Barcode:         req.Barcode,
		RFIDTag:         req.RFIDTag,
		EquipmentID:     req.EquipmentID,
		Action:          domain.ActionAdhocBooking,
		ProjectID:       req.ProjectID,
		LocationID:      req.LocationID,
		ConditionRating: req.ConditionRating,
		ConditionNotes:  req.ConditionNotes,
		GPSLat:          req.GPSLat,
		GPSLng:          req.GPSLng,
		Timestamp:       now,
		SyncedAt:        now,
	}

	if err := s.eventRepo.Create(ctx, event); err != nil {
		s.logger.Error().Err(err).Str("tenant_id", tenantID.String()).Msg("failed to create adhoc booking event")
		return nil, fmt.Errorf("adhoc booking: %w", err)
	}

	// Update device last_seen (best-effort).
	if req.DeviceID != "" {
		_ = s.deviceRepo.UpdateLastSeen(ctx, tenantID, req.DeviceID)
	}

	s.logger.Info().
		Str("event_id", event.ID.String()).
		Str("tenant_id", tenantID.String()).
		Msg("adhoc booking event recorded")

	return event, nil
}

// ListEvents returns a filtered, paginated list of scan events for a tenant.
func (s *ScanService) ListEvents(
	ctx context.Context,
	tenantID uuid.UUID,
	filter domain.ScanEventFilter,
) ([]*domain.ScanEvent, int64, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PerPage < 1 {
		filter.PerPage = 20
	}

	items, total, err := s.eventRepo.List(ctx, tenantID, filter)
	if err != nil {
		s.logger.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Msg("failed to list scan events")
		return nil, 0, fmt.Errorf("list scan events: %w", err)
	}
	return items, total, nil
}

// updateEquipmentLocation notifies the inventory service about a location change (best-effort).
func (s *ScanService) updateEquipmentLocation(tenantID, equipmentID, locationID uuid.UUID) {
	url := fmt.Sprintf("%s/api/v1/equipment/%s/location", s.inventoryBaseURL, equipmentID.String())
	body := fmt.Sprintf(`{"location_id":"%s"}`, locationID.String())

	req, err := http.NewRequest("PATCH", url, strings.NewReader(body))
	if err != nil {
		s.logger.Warn().Err(err).Msg("failed to build location update request")
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", tenantID.String())
	req.Header.Set("X-Internal-Service", "scanner")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		s.logger.Warn().Err(err).Str("equipment_id", equipmentID.String()).Msg("failed to update equipment location")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		s.logger.Warn().Int("status", resp.StatusCode).Str("equipment_id", equipmentID.String()).Msg("inventory location update failed")
	}
}

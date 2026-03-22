package application

import "github.com/jeckersberger/rentflow/services/scanner-service/internal/domain"

func ScanEventToDTO(event *domain.ScanEvent) *ScanEventDTO {
	return &ScanEventDTO{
		ID:          event.ID,
		TenantID:    event.TenantID,
		Barcode:     event.Barcode,
		ScanType:    string(event.ScanType),
		EquipmentID: event.EquipmentID,
		ProjectID:   event.ProjectID,
		LocationID:  event.LocationID,
		UserID:      event.UserID,
		DeviceID:    event.DeviceID,
		DeviceType:  string(event.DeviceType),
		Timestamp:   event.Timestamp.Format("2006-01-02T15:04:05Z"),
		Latitude:    event.Latitude,
		Longitude:   event.Longitude,
		Notes:       event.Notes,
		Status:      string(event.Status),
		CreatedAt:   event.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func DeviceToDTO(device *domain.Device) *DeviceDTO {
	return &DeviceDTO{
		ID:        device.ID,
		TenantID:  device.TenantID,
		Name:      device.Name,
		Type:      string(device.Type),
		Serial:    device.Serial,
		Active:    device.Active,
		Location:  device.Location,
		CreatedAt: device.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: device.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

type ScanSessionDTO struct {
	ID              string  `json:"id"`
	TenantID        string  `json:"tenant_id"`
	UserID          string  `json:"user_id"`
	Context         string  `json:"context"`
	ProjectID       *string `json:"project_id"`
	StartedAt       string  `json:"started_at"`
	EndedAt         *string `json:"ended_at"`
	DeviceType      string  `json:"device_type"`
	DeviceID        string  `json:"device_id"`
	TotalScans      int     `json:"total_scans"`
	SuccessfulScans int     `json:"successful_scans"`
	FailedScans     int     `json:"failed_scans"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
}

type ScanResultDTO struct {
	ID          string `json:"id"`
	SessionID   string `json:"session_id"`
	Barcode     string `json:"barcode"`
	EquipmentID string `json:"equipment_id"`
	Result      string `json:"result"` // success, warning, error
	Message     string `json:"message"`
	Timestamp   string `json:"timestamp"`
}

type OfflineQueueDTO struct {
	ID         string  `json:"id"`
	TenantID   string  `json:"tenant_id"`
	DeviceID   string  `json:"device_id"`
	SyncStatus string  `json:"sync_status"`
	CreatedAt  string  `json:"created_at"`
	SyncedAt   *string `json:"synced_at"`
}

type SyncResultDTO struct {
	TotalItems  int      `json:"total_items"`
	SyncedItems int      `json:"synced_items"`
	FailedItems int      `json:"failed_items"`
	Message     string   `json:"message"`
	Conflicts   []string `json:"conflicts"`
}

func ScanSessionToDTO(session *domain.ScanSession) *ScanSessionDTO {
	dto := &ScanSessionDTO{
		ID:              session.ID,
		TenantID:        session.TenantID,
		UserID:          session.UserID,
		Context:         string(session.Context),
		ProjectID:       session.ProjectID,
		StartedAt:       session.StartedAt.Format("2006-01-02T15:04:05Z"),
		DeviceType:      string(session.DeviceType),
		DeviceID:        session.DeviceID,
		TotalScans:      session.TotalScans,
		SuccessfulScans: session.SuccessfulScans,
		FailedScans:     session.FailedScans,
		CreatedAt:       session.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:       session.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if session.EndedAt != nil {
		endStr := session.EndedAt.Format("2006-01-02T15:04:05Z")
		dto.EndedAt = &endStr
	}
	return dto
}

func OfflineQueueToDTO(item *domain.OfflineQueueItem) *OfflineQueueDTO {
	dto := &OfflineQueueDTO{
		ID:         item.ID,
		TenantID:   item.TenantID,
		DeviceID:   item.DeviceID,
		SyncStatus: item.SyncStatus,
		CreatedAt:  item.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if item.SyncedAt != nil {
		syncStr := item.SyncedAt.Format("2006-01-02T15:04:05Z")
		dto.SyncedAt = &syncStr
	}
	return dto
}

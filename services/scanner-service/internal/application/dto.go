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

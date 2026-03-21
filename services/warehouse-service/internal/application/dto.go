package application

import "github.com/jeckersberger/rentflow/services/warehouse-service/internal/domain"

func LocationToDTO(loc *domain.Location) *LocationDTO {
	return &LocationDTO{
		ID:           loc.ID,
		TenantID:     loc.TenantID,
		Name:         loc.Name,
		Type:         string(loc.Type),
		ParentID:     loc.ParentID,
		Path:         loc.Path,
		Capacity:     loc.Capacity,
		CurrentCount: loc.CurrentCount,
		SortOrder:    loc.SortOrder,
		Barcode:      loc.Barcode,
		CreatedAt:    loc.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:    loc.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func MovementToDTO(mov *domain.Movement) *MovementDTO {
	return &MovementDTO{
		ID:             mov.ID,
		TenantID:       mov.TenantID,
		EquipmentID:    mov.EquipmentID,
		FromLocationID: mov.FromLocationID,
		ToLocationID:   mov.ToLocationID,
		MovementType:   string(mov.MovementType),
		Quantity:       mov.Quantity,
		Reason:         mov.Reason,
		UserID:         mov.UserID,
		ProjectID:      mov.ProjectID,
		Timestamp:      mov.Timestamp.Format("2006-01-02T15:04:05Z"),
		CreatedAt:      mov.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func InventoryCheckToDTO(check *domain.InventoryCheck) *InventoryCheckDTO {
	dto := &InventoryCheckDTO{
		ID:         check.ID,
		TenantID:   check.TenantID,
		Name:       check.Name,
		LocationID: check.LocationID,
		Status:     string(check.Status),
		Items:      make([]InventoryCheckItemDTO, len(check.Items)),
		CreatedAt:  check.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:  check.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	if check.StartedAt != nil {
		startedAtStr := check.StartedAt.Format("2006-01-02T15:04:05Z")
		dto.StartedAt = &startedAtStr
	}

	if check.CompletedAt != nil {
		completedAtStr := check.CompletedAt.Format("2006-01-02T15:04:05Z")
		dto.CompletedAt = &completedAtStr
	}

	for i, item := range check.Items {
		dto.Items[i] = InventoryCheckItemDTO{
			EquipmentID:   item.EquipmentID,
			ExpectedCount: item.ExpectedCount,
			ActualCount:   item.ActualCount,
			Status:        item.Status,
			Notes:         item.Notes,
		}
		if item.ScannedAt != nil {
			scannedAtStr := item.ScannedAt.Format("2006-01-02T15:04:05Z")
			dto.Items[i].ScannedAt = &scannedAtStr
		}
	}

	return dto
}

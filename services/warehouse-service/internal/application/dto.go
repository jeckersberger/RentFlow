package application

import "github.com/jeckersberger/rentflow/services/warehouse-service/internal/domain"

// DTO type definitions for existing domain models
type LocationDTO struct {
	ID           string `json:"id"`
	TenantID     string `json:"tenant_id"`
	Name         string `json:"name"`
	Type         string `json:"type"`
	ParentID     *string `json:"parent_id,omitempty"`
	Path         string `json:"path"`
	Capacity     int `json:"capacity"`
	CurrentCount int `json:"current_count"`
	SortOrder    int `json:"sort_order"`
	Barcode      string `json:"barcode"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

type MovementDTO struct {
	ID             string `json:"id"`
	TenantID       string `json:"tenant_id"`
	EquipmentID    string `json:"equipment_id"`
	FromLocationID *string `json:"from_location_id,omitempty"`
	ToLocationID   string `json:"to_location_id"`
	MovementType   string `json:"movement_type"`
	Quantity       int `json:"quantity"`
	Reason         string `json:"reason"`
	UserID         string `json:"user_id"`
	ProjectID      *string `json:"project_id,omitempty"`
	Timestamp      string `json:"timestamp"`
	CreatedAt      string `json:"created_at"`
}

type InventoryCheckDTO struct {
	ID          string `json:"id"`
	TenantID    string `json:"tenant_id"`
	Name        string `json:"name"`
	LocationID  *string `json:"location_id,omitempty"`
	Status      string `json:"status"`
	Items       []InventoryCheckItemDTO `json:"items"`
	StartedAt   *string `json:"started_at,omitempty"`
	CompletedAt *string `json:"completed_at,omitempty"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type InventoryCheckItemDTO struct {
	EquipmentID   string `json:"equipment_id"`
	ExpectedCount int `json:"expected_count"`
	ActualCount   int `json:"actual_count"`
	Status        string `json:"status"`
	Notes         string `json:"notes"`
	ScannedAt     *string `json:"scanned_at,omitempty"`
}

type LocationTreeNode struct {
	*LocationDTO
	Children []*LocationTreeNode `json:"children"`
}

type PaginatedResult struct {
	Data   interface{} `json:"data"`
	Total  int `json:"total"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

// DTO type definitions for warehouse domain models
type WarehouseDTO struct {
	ID               string `json:"id"`
	TenantID         string `json:"tenant_id"`
	Name             string `json:"name"`
	Code             string `json:"code"`
	Address          string `json:"address,omitempty"`
	City             string `json:"city,omitempty"`
	PostalCode       string `json:"postal_code,omitempty"`
	Country          string `json:"country,omitempty"`
	Latitude         *float64 `json:"latitude,omitempty"`
	Longitude        *float64 `json:"longitude,omitempty"`
	TotalCapacity    int `json:"total_capacity"`
	CurrentOccupancy int `json:"current_occupancy"`
	Status           string `json:"status"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
}

type ZoneDTO struct {
	ID             string `json:"id"`
	TenantID       string `json:"tenant_id"`
	WarehouseID    string `json:"warehouse_id"`
	Name           string `json:"name"`
	Code           string `json:"code"`
	ZoneType       string `json:"zone_type,omitempty"`
	Description    string `json:"description,omitempty"`
	TemperatureMin *float64 `json:"temperature_min,omitempty"`
	TemperatureMax *float64 `json:"temperature_max,omitempty"`
	HumidityMin    *float64 `json:"humidity_min,omitempty"`
	HumidityMax    *float64 `json:"humidity_max,omitempty"`
	Capacity       int `json:"capacity"`
	Occupancy      int `json:"occupancy"`
	SortOrder      int `json:"sort_order"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}

type RackDTO struct {
	ID             string `json:"id"`
	TenantID       string `json:"tenant_id"`
	ZoneID         string `json:"zone_id"`
	WarehouseID    string `json:"warehouse_id"`
	Name           string `json:"name"`
	Code           string `json:"code"`
	RackType       string `json:"rack_type"`
	Aisle          string `json:"aisle,omitempty"`
	RowNumber      int `json:"row_number"`
	ColumnNumber   int `json:"column_number"`
	Capacity       int `json:"capacity"`
	Occupancy      int `json:"occupancy"`
	Height         *float64 `json:"height,omitempty"`
	Width          *float64 `json:"width,omitempty"`
	Depth          *float64 `json:"depth,omitempty"`
	WeightCapacity *float64 `json:"weight_capacity,omitempty"`
	SortOrder      int `json:"sort_order"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}

type BayDTO struct {
	ID             string `json:"id"`
	TenantID       string `json:"tenant_id"`
	RackID         string `json:"rack_id"`
	ZoneID         string `json:"zone_id"`
	WarehouseID    string `json:"warehouse_id"`
	Name           string `json:"name"`
	Code           string `json:"code"`
	BayNumber      int `json:"bay_number"`
	BayLevel       int `json:"bay_level"`
	Capacity       int `json:"capacity"`
	Occupancy      int `json:"occupancy"`
	WeightCapacity *float64 `json:"weight_capacity,omitempty"`
	SortOrder      int `json:"sort_order"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}

type StockLocationDTO struct {
	ID             string `json:"id"`
	TenantID       string `json:"tenant_id"`
	WarehouseID    string `json:"warehouse_id"`
	ZoneID         string `json:"zone_id"`
	RackID         string `json:"rack_id"`
	BayID          string `json:"bay_id"`
	LocationCode   string `json:"location_code"`
	Barcode        string `json:"barcode,omitempty"`
	QRCodeLabel    string `json:"qr_code_label,omitempty"`
	LocationType   string `json:"location_type"`
	Capacity       int `json:"capacity"`
	Occupancy      int `json:"occupancy"`
	WeightCapacity *float64 `json:"weight_capacity,omitempty"`
	CurrentWeight  *float64 `json:"current_weight,omitempty"`
	EquipmentID    *string `json:"equipment_id,omitempty"`
	IsAvailable    bool `json:"is_available"`
	AccessLevel    string `json:"access_level"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}

type InventoryItemDTO struct {
	ID            string `json:"id"`
	TenantID      string `json:"tenant_id"`
	CheckID       string `json:"check_id"`
	EquipmentID   string `json:"equipment_id"`
	LocationID    *string `json:"location_id,omitempty"`
	ExpectedCount int `json:"expected_count"`
	ActualCount   int `json:"actual_count"`
	Variance      int `json:"variance"`
	Status        string `json:"status"`
	ScannedAt     *string `json:"scanned_at,omitempty"`
	Notes         string `json:"notes,omitempty"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

type InventoryCheckExtendedDTO struct {
	ID          string `json:"id"`
	TenantID    string `json:"tenant_id"`
	WarehouseID *string `json:"warehouse_id,omitempty"`
	ZoneID      *string `json:"zone_id,omitempty"`
	CheckType   string `json:"check_type"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Status      string `json:"status"`
	Items       []InventoryItemDTO `json:"items"`
	StartedAt   *string `json:"started_at,omitempty"`
	CompletedAt *string `json:"completed_at,omitempty"`
	CompletedBy *string `json:"completed_by,omitempty"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// Conversion functions
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
		LocationID: check.ZoneID,
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
			Status:        string(item.Status),
			Notes:         item.Notes,
		}
		if item.ScannedAt != nil {
			scannedAtStr := item.ScannedAt.Format("2006-01-02T15:04:05Z")
			dto.Items[i].ScannedAt = &scannedAtStr
		}
	}

	return dto
}

// Warehouse domain to DTO conversions
func WarehouseToDTO(w *domain.Warehouse) *WarehouseDTO {
	return &WarehouseDTO{
		ID:               w.ID,
		TenantID:         w.TenantID,
		Name:             w.Name,
		Code:             w.Code,
		Address:          w.Address,
		City:             w.City,
		PostalCode:       w.PostalCode,
		Country:          w.Country,
		Latitude:         w.Latitude,
		Longitude:        w.Longitude,
		TotalCapacity:    w.TotalCapacity,
		CurrentOccupancy: w.CurrentOccupancy,
		Status:           string(w.Status),
		CreatedAt:        w.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:        w.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func ZoneToDTO(z *domain.Zone) *ZoneDTO {
	return &ZoneDTO{
		ID:             z.ID,
		TenantID:       z.TenantID,
		WarehouseID:    z.WarehouseID,
		Name:           z.Name,
		Code:           z.Code,
		ZoneType:       z.ZoneType,
		Description:    z.Description,
		TemperatureMin: z.TemperatureMin,
		TemperatureMax: z.TemperatureMax,
		HumidityMin:    z.HumidityMin,
		HumidityMax:    z.HumidityMax,
		Capacity:       z.Capacity,
		Occupancy:      z.Occupancy,
		SortOrder:      z.SortOrder,
		CreatedAt:      z.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:      z.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func RackToDTO(r *domain.Rack) *RackDTO {
	return &RackDTO{
		ID:             r.ID,
		TenantID:       r.TenantID,
		ZoneID:         r.ZoneID,
		WarehouseID:    r.WarehouseID,
		Name:           r.Name,
		Code:           r.Code,
		RackType:       r.RackType,
		Aisle:          r.Aisle,
		RowNumber:      r.RowNumber,
		ColumnNumber:   r.ColumnNumber,
		Capacity:       r.Capacity,
		Occupancy:      r.Occupancy,
		Height:         r.Height,
		Width:          r.Width,
		Depth:          r.Depth,
		WeightCapacity: r.WeightCapacity,
		SortOrder:      r.SortOrder,
		CreatedAt:      r.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:      r.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func BayToDTO(b *domain.Bay) *BayDTO {
	return &BayDTO{
		ID:             b.ID,
		TenantID:       b.TenantID,
		RackID:         b.RackID,
		ZoneID:         b.ZoneID,
		WarehouseID:    b.WarehouseID,
		Name:           b.Name,
		Code:           b.Code,
		BayNumber:      b.BayNumber,
		BayLevel:       b.BayLevel,
		Capacity:       b.Capacity,
		Occupancy:      b.Occupancy,
		WeightCapacity: b.WeightCapacity,
		SortOrder:      b.SortOrder,
		CreatedAt:      b.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:      b.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func StockLocationToDTO(sl *domain.StockLocation) *StockLocationDTO {
	return &StockLocationDTO{
		ID:             sl.ID,
		TenantID:       sl.TenantID,
		WarehouseID:    sl.WarehouseID,
		ZoneID:         sl.ZoneID,
		RackID:         sl.RackID,
		BayID:          sl.BayID,
		LocationCode:   sl.LocationCode,
		Barcode:        sl.Barcode,
		QRCodeLabel:    sl.QRCodeLabel,
		LocationType:   sl.LocationType,
		Capacity:       sl.Capacity,
		Occupancy:      sl.Occupancy,
		WeightCapacity: sl.WeightCapacity,
		CurrentWeight:  sl.CurrentWeight,
		EquipmentID:    sl.EquipmentID,
		IsAvailable:    sl.IsAvailable,
		AccessLevel:    sl.AccessLevel,
		CreatedAt:      sl.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:      sl.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func InventoryCheckExtendedToDTO(ic *domain.InventoryCheck) *InventoryCheckExtendedDTO {
	dto := &InventoryCheckExtendedDTO{
		ID:          ic.ID,
		TenantID:    ic.TenantID,
		WarehouseID: ic.WarehouseID,
		ZoneID:      ic.ZoneID,
		CheckType:   string(ic.CheckType),
		Name:        ic.Name,
		Description: ic.Description,
		Status:      string(ic.Status),
		Items:       make([]InventoryItemDTO, len(ic.Items)),
		CreatedAt:   ic.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   ic.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	if ic.StartedAt != nil {
		startedAtStr := ic.StartedAt.Format("2006-01-02T15:04:05Z")
		dto.StartedAt = &startedAtStr
	}

	if ic.CompletedAt != nil {
		completedAtStr := ic.CompletedAt.Format("2006-01-02T15:04:05Z")
		dto.CompletedAt = &completedAtStr
	}

	if ic.CompletedBy != nil {
		dto.CompletedBy = ic.CompletedBy
	}

	for i, item := range ic.Items {
		itemDTO := InventoryItemDTO{
			ID:            item.ID,
			TenantID:      item.TenantID,
			CheckID:       item.CheckID,
			EquipmentID:   item.EquipmentID,
			LocationID:    item.LocationID,
			ExpectedCount: item.ExpectedCount,
			ActualCount:   item.ActualCount,
			Variance:      item.Variance,
			Status:        string(item.Status),
			Notes:         item.Notes,
			CreatedAt:     item.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:     item.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}
		if item.ScannedAt != nil {
			scannedAtStr := item.ScannedAt.Format("2006-01-02T15:04:05Z")
			itemDTO.ScannedAt = &scannedAtStr
		}
		dto.Items[i] = itemDTO
	}

	return dto
}

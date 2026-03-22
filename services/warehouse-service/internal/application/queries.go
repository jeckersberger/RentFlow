package application

import "time"

type HistoryQuery struct {
	EquipmentID  *string
	FromLocation *string
	ToLocation   *string
	MovementType *string
	StartDate    *time.Time
	EndDate      *time.Time
	Limit        int
	Offset       int
}

type LocationContentsDTO struct {
	LocationID string      `json:"location_id"`
	Items      interface{} `json:"items"`
	Total      int         `json:"total"`
}

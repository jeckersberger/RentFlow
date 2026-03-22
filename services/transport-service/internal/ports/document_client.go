package ports

import "context"

// DeliveryNoteItem represents an item in a delivery note
type DeliveryNoteItem struct {
	EquipmentID string
	WeightKg    float64
	VolumeM3    float64
}

// DeliveryNoteResult contains the result of generating a delivery note
type DeliveryNoteResult struct {
	DeliveryNoteNumber string
	DocumentURL        string
}

// DocumentClient defines the interface for communicating with document-service
type DocumentClient interface {
	// GenerateDeliveryNote generates a delivery note document for a tour
	GenerateDeliveryNote(ctx context.Context, tenantID string, tourID string, equipmentItems []DeliveryNoteItem) (*DeliveryNoteResult, error)
}

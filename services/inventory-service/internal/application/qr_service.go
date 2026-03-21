package application

import (
	"context"
	"fmt"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/inventory-service/internal/ports"
	"github.com/skip2/go-qrcode"
)

type QRService struct {
	equipRepo ports.EquipmentRepository
	logger    logger.Logger
}

func NewQRService(
	equipRepo ports.EquipmentRepository,
	logger logger.Logger,
) *QRService {
	return &QRService{
		equipRepo: equipRepo,
		logger:    logger,
	}
}

// GenerateQRCode generates a QR code PNG for the given equipment ID.
// The QR code contains: rentflow://equipment/{equipment_id}
func (s *QRService) GenerateQRCode(
	ctx context.Context,
	tenantID string,
	equipmentID string,
	size int,
) ([]byte, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant ID is required")
	}
	if equipmentID == "" {
		return nil, fmt.Errorf("equipment ID is required")
	}

	// Verify equipment exists
	_, err := s.equipRepo.GetByID(ctx, tenantID, equipmentID)
	if err != nil {
		return nil, fmt.Errorf("equipment not found: %w", err)
	}

	// Default size
	if size == 0 {
		size = 256
	}
	if size < 50 || size > 2048 {
		return nil, fmt.Errorf("size must be between 50 and 2048 pixels")
	}

	// Generate QR code content
	content := fmt.Sprintf("rentflow://equipment/%s", equipmentID)

	// Generate PNG
	qr, err := qrcode.New(content, qrcode.Medium)
	if err != nil {
		return nil, fmt.Errorf("failed to generate QR code: %w", err)
	}

	png, err := qr.PNG(size)
	if err != nil {
		return nil, fmt.Errorf("failed to encode QR code: %w", err)
	}

	s.logger.Info("QR code generated", "equipment_id", equipmentID, "size", size)
	return png, nil
}

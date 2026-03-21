package application

import (
	"context"
	"fmt"
	"strings"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/inventory-service/internal/ports"
)

type LabelService struct {
	equipRepo ports.EquipmentRepository
	logger    logger.Logger
}

func NewLabelService(
	equipRepo ports.EquipmentRepository,
	logger logger.Logger,
) *LabelService {
	return &LabelService{
		equipRepo: equipRepo,
		logger:    logger,
	}
}

// GenerateZPLLabel generates ZPL II code for Zebra thermal printers.
// Standard 50x25mm label format.
func (s *LabelService) GenerateZPLLabel(
	ctx context.Context,
	tenantID string,
	equipmentID string,
) (string, error) {
	if tenantID == "" {
		return "", fmt.Errorf("tenant ID is required")
	}
	if equipmentID == "" {
		return "", fmt.Errorf("equipment ID is required")
	}

	eq, err := s.equipRepo.GetByID(ctx, tenantID, equipmentID)
	if err != nil {
		return "", fmt.Errorf("equipment not found: %w", err)
	}

	// Sanitize text for ZPL (remove special characters that could interfere)
	name := sanitizeZPLText(eq.Name)
	sku := sanitizeZPLText(eq.SKU)
	barcode := eq.Barcode

	// Generate ZPL II code
	// Standard 50x25mm = ~142x70 dots at 203dpi
	zpl := generateZPLCode(name, sku, barcode, equipmentID)

	s.logger.Info("ZPL label generated", "equipment_id", equipmentID, "barcode", barcode)
	return zpl, nil
}

func generateZPLCode(name, sku, barcode, equipmentID string) string {
	// ZPL II format for 50x25mm label (203 dpi)
	// Label size: 50mm width = ~567 dots, 25mm height = ~283 dots
	var zpl strings.Builder

	// Start label format
	zpl.WriteString("^XA\n")
	zpl.WriteString("^LL283\n") // Label length in dots (25mm)
	zpl.WriteString("^PW567\n") // Label width in dots (50mm)

	// Equipment Name (top, large font)
	zpl.WriteString("^FO30,15\n")
	zpl.WriteString("^A0N,20,20\n")
	zpl.WriteString(fmt.Sprintf("^FD%s^FS\n", name))

	// SKU (middle-top)
	zpl.WriteString("^FO30,40\n")
	zpl.WriteString("^A0N,14,14\n")
	zpl.WriteString(fmt.Sprintf("^FDSKU: %s^FS\n", sku))

	// Code128 Barcode
	zpl.WriteString("^FO30,60\n")
	zpl.WriteString("^BY2,3,60\n")
	zpl.WriteString(fmt.Sprintf("^BC,,Y\n"))
	zpl.WriteString(fmt.Sprintf("^FD%s^FS\n", barcode))

	// QR Code (right side, smaller)
	zpl.WriteString("^FO410,30\n")
	zpl.WriteString("^BQN,2,5\n")
	zpl.WriteString(fmt.Sprintf("^FDLA,rentflow://equipment/%s^FS\n", equipmentID))

	// End label
	zpl.WriteString("^XZ\n")

	return zpl.String()
}

func sanitizeZPLText(text string) string {
	// Remove characters that have special meaning in ZPL
	replacer := strings.NewReplacer(
		"^", "",
		"~", "",
		"&", "and",
		"'", "",
		"\"", "",
	)
	return replacer.Replace(text)
}

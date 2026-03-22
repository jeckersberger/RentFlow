package application

import (
	"fmt"
	"strings"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/warehouse-service/internal/domain"
)

type ZPLService struct {
	logger logger.Logger
}

func NewZPLService(logger logger.Logger) *ZPLService {
	return &ZPLService{
		logger: logger,
	}
}

// GenerateLocationLabel generates ZPL format label for a warehouse location
func (s *ZPLService) GenerateLocationLabel(location *domain.StockLocation) string {
	// ZPL (Zebra Programming Language) format for 4x6 label
	// ^XA = Start of label
	// ^XZ = End of label
	// ^FO = Field Origin (x,y position)
	// ^A = Font selection
	// ^FD = Field Data
	// ^BC = Barcode

	zpl := strings.Builder{}

	// Start ZPL label
	zpl.WriteString("^XA\n")
	zpl.WriteString("^MMT\n")
	zpl.WriteString("^PW812\n")
	zpl.WriteString("^LL1218\n")

	// Title field - Location Code (large)
	zpl.WriteString("^FO50,50\n")
	zpl.WriteString("^AF,36,20\n")
	zpl.WriteString(fmt.Sprintf("^FD%s^FS\n", location.LocationCode))

	// Warehouse/Zone/Rack/Bay info
	zpl.WriteString("^FO50,120\n")
	zpl.WriteString("^AF,24,12\n")
	zpl.WriteString(fmt.Sprintf("^FDWarehouse: %s^FS\n", location.WarehouseID[:8]))

	zpl.WriteString("^FO50,160\n")
	zpl.WriteString("^AF,24,12\n")
	zpl.WriteString(fmt.Sprintf("^FDZone: %s^FS\n", location.ZoneID[:8]))

	zpl.WriteString("^FO50,200\n")
	zpl.WriteString("^AF,24,12\n")
	zpl.WriteString(fmt.Sprintf("^FDRack: %s^FS\n", location.RackID[:8]))

	zpl.WriteString("^FO50,240\n")
	zpl.WriteString("^AF,24,12\n")
	zpl.WriteString(fmt.Sprintf("^FDBay: %s^FS\n", location.BayID[:8]))

	// Barcode - generate from location code
	zpl.WriteString("^FO100,300\n")
	zpl.WriteString("^BCR,100,Y,N,N\n")
	zpl.WriteString(fmt.Sprintf("^FD%s^FS\n", location.LocationCode))

	// QR Code - encode location ID for fast lookup
	zpl.WriteString("^FO450,300\n")
	zpl.WriteString("^BQN,2,8\n")
	zpl.WriteString(fmt.Sprintf("^FDLC,%s^FS\n", location.ID))

	// Capacity info
	zpl.WriteString("^FO50,550\n")
	zpl.WriteString("^AF,24,12\n")
	zpl.WriteString(fmt.Sprintf("^FDCapacity: %d^FS\n", location.Capacity))

	// Status
	status := "Available"
	if !location.IsAvailable {
		status = "Unavailable"
	}
	zpl.WriteString("^FO50,590\n")
	zpl.WriteString("^AF,24,12\n")
	zpl.WriteString(fmt.Sprintf("^FDStatus: %s^FS\n", status))

	// End ZPL label
	zpl.WriteString("^XZ\n")

	return zpl.String()
}

// GenerateWarehouseLabel generates a ZPL label for warehouse identification
func (s *ZPLService) GenerateWarehouseLabel(warehouse *domain.Warehouse) string {
	zpl := strings.Builder{}

	zpl.WriteString("^XA\n")
	zpl.WriteString("^MMT\n")
	zpl.WriteString("^PW812\n")
	zpl.WriteString("^LL1218\n")

	// Warehouse name and code
	zpl.WriteString("^FO50,50\n")
	zpl.WriteString("^AF,36,20\n")
	zpl.WriteString(fmt.Sprintf("^FD%s^FS\n", warehouse.Code))

	zpl.WriteString("^FO50,120\n")
	zpl.WriteString("^AF,24,12\n")
	zpl.WriteString(fmt.Sprintf("^FD%s^FS\n", warehouse.Name))

	// Location
	if warehouse.City != "" || warehouse.Country != "" {
		zpl.WriteString("^FO50,170\n")
		zpl.WriteString("^AF,20,10\n")
		zpl.WriteString(fmt.Sprintf("^FD%s, %s^FS\n", warehouse.City, warehouse.Country))
	}

	// Barcode
	zpl.WriteString("^FO100,250\n")
	zpl.WriteString("^BCR,100,Y,N,N\n")
	zpl.WriteString(fmt.Sprintf("^FD%s^FS\n", warehouse.Code))

	// QR Code
	zpl.WriteString("^FO450,250\n")
	zpl.WriteString("^BQN,2,8\n")
	zpl.WriteString(fmt.Sprintf("^FDWH,%s^FS\n", warehouse.ID))

	// Capacity
	zpl.WriteString("^FO50,500\n")
	zpl.WriteString("^AF,24,12\n")
	zpl.WriteString(fmt.Sprintf("^FDCapacity: %d^FS\n", warehouse.TotalCapacity))

	// Status
	zpl.WriteString("^FO50,540\n")
	zpl.WriteString("^AF,24,12\n")
	zpl.WriteString(fmt.Sprintf("^FDStatus: %s^FS\n", warehouse.Status))

	zpl.WriteString("^XZ\n")

	return zpl.String()
}

// GenerateZoneLabel generates a ZPL label for zone identification
func (s *ZPLService) GenerateZoneLabel(zone *domain.Zone) string {
	zpl := strings.Builder{}

	zpl.WriteString("^XA\n")
	zpl.WriteString("^MMT\n")
	zpl.WriteString("^PW812\n")
	zpl.WriteString("^LL1218\n")

	// Zone name and code
	zpl.WriteString("^FO50,50\n")
	zpl.WriteString("^AF,36,20\n")
	zpl.WriteString(fmt.Sprintf("^FD%s^FS\n", zone.Code))

	zpl.WriteString("^FO50,120\n")
	zpl.WriteString("^AF,24,12\n")
	zpl.WriteString(fmt.Sprintf("^FD%s^FS\n", zone.Name))

	// Zone type
	if zone.ZoneType != "" {
		zpl.WriteString("^FO50,170\n")
		zpl.WriteString("^AF,20,10\n")
		zpl.WriteString(fmt.Sprintf("^FDType: %s^FS\n", zone.ZoneType))
	}

	// Temperature if applicable
	if zone.TemperatureMin != nil && zone.TemperatureMax != nil {
		zpl.WriteString("^FO50,210\n")
		zpl.WriteString("^AF,20,10\n")
		zpl.WriteString(fmt.Sprintf("^FDTemp: %.1f-%.1fC^FS\n", *zone.TemperatureMin, *zone.TemperatureMax))
	}

	// Barcode
	zpl.WriteString("^FO100,280\n")
	zpl.WriteString("^BCR,80,Y,N,N\n")
	zpl.WriteString(fmt.Sprintf("^FD%s^FS\n", zone.Code))

	// QR Code
	zpl.WriteString("^FO450,280\n")
	zpl.WriteString("^BQN,2,8\n")
	zpl.WriteString(fmt.Sprintf("^FDZN,%s^FS\n", zone.ID))

	// Capacity
	zpl.WriteString("^FO50,500\n")
	zpl.WriteString("^AF,24,12\n")
	zpl.WriteString(fmt.Sprintf("^FDCapacity: %d^FS\n", zone.Capacity))

	zpl.WriteString("^XZ\n")

	return zpl.String()
}

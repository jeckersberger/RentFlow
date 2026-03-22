package http

import (
	"net/http"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/warehouse-service/internal/application"
)

func NewRouter(
	locationSvc *application.LocationService,
	movementSvc *application.MovementService,
	inventoryCheckSvc *application.InventoryCheckService,
	warehouseSvc *application.WarehouseService,
	zplSvc *application.ZPLService,
	logger logger.Logger,
) *http.ServeMux {
	router := http.NewServeMux()
	handler := NewHandler(locationSvc, movementSvc, inventoryCheckSvc, warehouseSvc, zplSvc, logger)

	// Health & readiness
	router.HandleFunc("GET /health", healthHandler)
	router.HandleFunc("GET /ready", readyHandler)

	// Location routes
	router.HandleFunc("POST /api/v1/locations", handler.CreateLocation)
	router.HandleFunc("GET /api/v1/locations", handler.ListLocations)
	router.HandleFunc("GET /api/v1/locations/tree", handler.GetLocationTree)
	router.HandleFunc("GET /api/v1/locations/{id}", handler.GetLocation)
	router.HandleFunc("PUT /api/v1/locations/{id}", handler.UpdateLocation)
	router.HandleFunc("DELETE /api/v1/locations/{id}", handler.DeleteLocation)
	router.HandleFunc("GET /api/v1/locations/{id}/occupancy", handler.GetLocationOccupancy)

	// Movement routes
	router.HandleFunc("POST /api/v1/movements", handler.RecordMovement)
	router.HandleFunc("GET /api/v1/movements", handler.GetMovementHistory)
	router.HandleFunc("GET /api/v1/movements/equipment/{id}", handler.GetEquipmentHistory)

	// Inventory Check routes
	router.HandleFunc("POST /api/v1/inventory-checks", handler.StartInventoryCheck)
	router.HandleFunc("GET /api/v1/inventory-checks", handler.ListInventoryChecks)
	router.HandleFunc("GET /api/v1/inventory-checks/{id}", handler.GetInventoryCheck)
	router.HandleFunc("POST /api/v1/inventory-checks/{id}/scan", handler.ScanInventoryItem)
	router.HandleFunc("POST /api/v1/inventory-checks/{id}/complete", handler.CompleteInventoryCheck)
	router.HandleFunc("GET /api/v1/inventory-checks/{id}/discrepancies", handler.GetDiscrepancies)

	// Warehouse hierarchy routes (M2.1 Phase 2)
	router.HandleFunc("POST /api/v1/warehouses", handler.CreateWarehouse)
	router.HandleFunc("GET /api/v1/warehouses", handler.ListWarehouses)
	router.HandleFunc("GET /api/v1/warehouses/{id}", handler.GetWarehouse)

	// Zone routes
	router.HandleFunc("POST /api/v1/zones", handler.CreateZone)

	// Rack routes
	router.HandleFunc("POST /api/v1/racks", handler.CreateRack)

	// Bay routes
	router.HandleFunc("POST /api/v1/bays", handler.CreateBay)

	return router
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy","service":"warehouse-service"}`))
}

func readyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ready","service":"warehouse-service"}`))
}

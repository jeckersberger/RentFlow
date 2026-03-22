package http

import (
	"net/http"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/inventory-service/internal/application"
)

func NewRouter(
	equipmentSvc *application.EquipmentService,
	categorySvc *application.CategoryService,
	flightcaseSvc *application.FlightcaseService,
	logger logger.Logger,
) *http.ServeMux {
	router := http.NewServeMux()
	handler := NewHandler(equipmentSvc, categorySvc, flightcaseSvc, logger)

	// Health & readiness
	router.HandleFunc("GET /health", healthHandler)
	router.HandleFunc("GET /ready", readyHandler)

	// Equipment routes - use lookup/ prefix for barcode to avoid conflict with {id} wildcard
	router.HandleFunc("POST /api/v1/equipment", handler.CreateEquipment)
	router.HandleFunc("GET /api/v1/equipment", handler.ListEquipment)
	router.HandleFunc("GET /api/v1/equipment/search", handler.SearchEquipment)
	router.HandleFunc("GET /api/v1/equipment/lookup/barcode/{barcode}", handler.GetEquipmentByBarcode)
	router.HandleFunc("GET /api/v1/equipment/lookup/rfid/{tag}", handler.GetEquipmentByRfid)
	router.HandleFunc("POST /api/v1/equipment/availability-check", handler.BatchCheckAvailability)
	router.HandleFunc("POST /api/v1/equipment/import", handler.ImportEquipmentFromCSV)
	router.HandleFunc("GET /api/v1/equipment/{id}", handler.GetEquipment)
	router.HandleFunc("PUT /api/v1/equipment/{id}", handler.UpdateEquipment)
	router.HandleFunc("PATCH /api/v1/equipment/{id}/status", handler.ChangeEquipmentStatus)
	router.HandleFunc("PATCH /api/v1/equipment/{id}/condition", handler.UpdateEquipmentCondition)
	router.HandleFunc("POST /api/v1/equipment/{id}/images", handler.AddEquipmentImage)
	router.HandleFunc("POST /api/v1/equipment/{id}/check-out", handler.CheckOutEquipment)
	router.HandleFunc("POST /api/v1/equipment/{id}/check-in", handler.CheckInEquipment)
	router.HandleFunc("PATCH /api/v1/equipment/{id}/rfid", handler.AssignRfidTag)
	router.HandleFunc("DELETE /api/v1/equipment/{id}", handler.DeleteEquipment)
	router.HandleFunc("GET /api/v1/equipment/{id}/price", handler.GetEquipmentPrice)
	router.HandleFunc("GET /api/v1/equipment/{id}/availability", handler.CheckEquipmentAvailability)
	router.HandleFunc("GET /api/v1/equipment/{id}/qr-code", handler.GetEquipmentQRCode)
	router.HandleFunc("GET /api/v1/equipment/{id}/label", handler.GetEquipmentLabel)
	router.HandleFunc("GET /api/v1/equipment/{id}/history", handler.GetEquipmentHistory)

	// Category routes
	router.HandleFunc("POST /api/v1/categories", handler.CreateCategory)
	router.HandleFunc("GET /api/v1/categories", handler.ListCategories)
	router.HandleFunc("GET /api/v1/categories/{id}", handler.GetCategory)
	router.HandleFunc("PUT /api/v1/categories/{id}", handler.UpdateCategory)
	router.HandleFunc("DELETE /api/v1/categories/{id}", handler.DeleteCategory)

	// Flightcase routes
	router.HandleFunc("POST /api/v1/flightcases", handler.CreateFlightcase)
	router.HandleFunc("GET /api/v1/flightcases", handler.ListFlightcases)
	router.HandleFunc("GET /api/v1/flightcases/{id}", handler.GetFlightcase)
	router.HandleFunc("PUT /api/v1/flightcases/{id}", handler.UpdateFlightcase)
	router.HandleFunc("POST /api/v1/flightcases/{id}/items", handler.AddFlightcaseItem)
	router.HandleFunc("DELETE /api/v1/flightcases/{id}/items", handler.RemoveFlightcaseItem)
	router.HandleFunc("DELETE /api/v1/flightcases/{id}", handler.DeleteFlightcase)

	return router
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy","service":"inventory-service"}`))
}

func readyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ready","service":"inventory-service"}`))
}

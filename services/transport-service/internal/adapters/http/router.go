package http

import (
	"net/http"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/transport-service/internal/application"
)

func NewRouter(vehicleSvc *application.VehicleService, tourSvc *application.TourService, log logger.Logger) *http.ServeMux {
	router := http.NewServeMux()
	handler := NewHandler(vehicleSvc, tourSvc, log)

	// Health checks
	router.HandleFunc("GET /health", healthHandler)
	router.HandleFunc("GET /ready", readyHandler)

	// Vehicles
	router.HandleFunc("POST /api/v1/transport/vehicles", handler.CreateVehicle)
	router.HandleFunc("GET /api/v1/transport/vehicles", handler.ListVehicles)
	router.HandleFunc("GET /api/v1/transport/vehicles/{id}", handler.GetVehicle)
	router.HandleFunc("PUT /api/v1/transport/vehicles/{id}", handler.UpdateVehicle)

	// Tours
	router.HandleFunc("POST /api/v1/transport/tours", handler.CreateTour)
	router.HandleFunc("GET /api/v1/transport/tours", handler.ListTours)
	router.HandleFunc("GET /api/v1/transport/tours/{id}", handler.GetTour)
	router.HandleFunc("PATCH /api/v1/transport/tours/{id}/status", handler.UpdateTourStatus)
	router.HandleFunc("POST /api/v1/transport/tours/{id}/start", handler.StartTour)
	router.HandleFunc("POST /api/v1/transport/tours/{id}/complete", handler.CompleteTour)

	// Tour Equipment
	router.HandleFunc("POST /api/v1/transport/tours/{id}/equipment", handler.AddEquipmentToTour)
	router.HandleFunc("DELETE /api/v1/transport/tours/{id}/equipment/{equipmentId}", handler.RemoveEquipmentFromTour)

	// Tour Capacity
	router.HandleFunc("GET /api/v1/transport/tours/{id}/capacity", handler.GetTourCapacity)

	// Driver Logs
	router.HandleFunc("POST /api/v1/transport/tours/{id}/driver-log", handler.LogDriverActivity)
	router.HandleFunc("GET /api/v1/transport/tours/{id}/driver-logs", handler.GetDriverLogs)

	return router
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy","service":"transport-service"}`))
}

func readyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ready","service":"transport-service"}`))
}

package http

import (
	"net/http"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/transport-service/internal/application"
)

func NewRouter(vehicleSvc *application.VehicleService, tourSvc *application.TourService, log logger.Logger) *http.ServeMux {
	router := http.NewServeMux()
	handler := NewHandler(vehicleSvc, tourSvc, log)

	router.HandleFunc("GET /health", healthHandler)
	router.HandleFunc("GET /ready", readyHandler)

	router.HandleFunc("POST /api/v1/vehicles", handler.CreateVehicle)
	router.HandleFunc("GET /api/v1/vehicles", handler.ListVehicles)
	router.HandleFunc("GET /api/v1/vehicles/{id}", handler.GetVehicle)
	router.HandleFunc("PUT /api/v1/vehicles/{id}", handler.UpdateVehicle)

	router.HandleFunc("POST /api/v1/tours", handler.CreateTour)
	router.HandleFunc("GET /api/v1/tours", handler.ListTours)
	router.HandleFunc("GET /api/v1/tours/{id}", handler.GetTour)
	router.HandleFunc("PATCH /api/v1/tours/{id}/status", handler.UpdateTourStatus)
	router.HandleFunc("GET /api/v1/tours/date/{date}", handler.ListToursByDate)

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

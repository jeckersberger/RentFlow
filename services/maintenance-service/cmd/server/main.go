package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/config"
	"github.com/jeckersberger/rentflow/pkg/common/database"
	"github.com/jeckersberger/rentflow/pkg/common/logger"
	httpAdapter "github.com/jeckersberger/rentflow/services/maintenance-service/internal/adapters/http"
	"github.com/jeckersberger/rentflow/services/maintenance-service/internal/application"
	"github.com/jeckersberger/rentflow/services/maintenance-service/internal/infrastructure/clients"
	"github.com/jeckersberger/rentflow/services/maintenance-service/internal/infrastructure/repositories"
)

const (
	serviceName = "maintenance-service"
	servicePort = 8010
)

func main() {
	cfg := config.Load(serviceName)
	if cfg.ServicePort == 8080 {
		cfg.ServicePort = servicePort
	}
	log := logger.New(cfg.LogLevel, serviceName)

	log.Info("Starting service", "name", serviceName, "port", cfg.ServicePort, "env", cfg.Environment)

	dbPool, err := database.NewPostgresPool(cfg.ConnectionString())
	if err != nil {
		log.Fatal("Failed to connect to database", err)
	}
	defer dbPool.Close()

	log.Info("Connected to database")

	recordRepo := repositories.NewMaintenanceRecordPostgres(dbPool)
	scheduleRepo := repositories.NewMaintenanceSchedulePostgres(dbPool)
	planRepo := repositories.NewMaintenancePlanPostgres(dbPool)
	taskRepo := repositories.NewMaintenanceTaskPostgres(dbPool)
	checklistRepo := repositories.NewChecklistPostgres(dbPool)
	testRepo := repositories.NewElectricalTestPostgres(dbPool)

	log.Info("Repositories initialized")

	// Initialize equipment locker client
	equipmentServiceURL := os.Getenv("EQUIPMENT_SERVICE_URL")
	if equipmentServiceURL == "" {
		equipmentServiceURL = "http://equipment-service:8002"
	}
	equipmentLocker := clients.NewEquipmentHTTPClient(equipmentServiceURL, nil, log)

	recordSvc := application.NewMaintenanceRecordService(recordRepo, log)
	scheduleSvc := application.NewMaintenanceScheduleService(scheduleRepo, recordRepo, log)
	planSvc := application.NewMaintenancePlanService(planRepo, taskRepo, log)
	taskSvc := application.NewMaintenanceTaskService(taskRepo, planRepo, equipmentLocker, log)
	checklistSvc := application.NewChecklistService(checklistRepo, log)
	testSvc := application.NewElectricalTestService(testRepo, log)

	log.Info("Services initialized")

	router := httpAdapter.NewRouter(recordSvc, scheduleSvc, planSvc, taskSvc, checklistSvc, testSvc, log)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.ServicePort),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Info("Listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("Server shutdown error", err)
	}

	log.Info("Server stopped")
}

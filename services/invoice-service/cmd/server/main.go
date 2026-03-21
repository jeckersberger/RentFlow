package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/config"
	"github.com/jeckersberger/rentflow/pkg/common/database"
	"github.com/jeckersberger/rentflow/pkg/common/logger"
	httpAdapter "github.com/jeckersberger/rentflow/services/invoice-service/internal/adapters/http"
	"github.com/jeckersberger/rentflow/services/invoice-service/internal/application"
	emailInfra "github.com/jeckersberger/rentflow/services/invoice-service/internal/infrastructure/email"
	pdfInfra "github.com/jeckersberger/rentflow/services/invoice-service/internal/infrastructure/pdf"
	"github.com/jeckersberger/rentflow/services/invoice-service/internal/infrastructure/repositories"
)

const (
	serviceName = "invoice-service"
	servicePort = 8006
)

func main() {
	cfg := config.Load(serviceName)
	if cfg.ServicePort == 8080 {
		cfg.ServicePort = servicePort
	}
	log := logger.New(cfg.LogLevel, serviceName)

	log.Info("Starting service", "name", serviceName, "port", cfg.ServicePort, "env", cfg.Environment)

	// Initialize database
	dbPool, err := database.NewPostgresPool(cfg.ConnectionString())
	if err != nil {
		log.Fatal("Failed to connect to database", err)
	}
	defer dbPool.Close()

	log.Info("Connected to database")

	// Initialize repositories
	invoiceRepo := repositories.NewInvoicePostgres(dbPool)
	quoteRepo := repositories.NewQuotePostgres(dbPool)
	dunningRepo := repositories.NewDunningPostgres(dbPool)
	seqRepo := repositories.NewNumberSequencePostgres(dbPool)

	log.Info("Repositories initialized")

	// Initialize services
	invoiceSvc := application.NewInvoiceService(invoiceRepo, seqRepo, log)
	quoteSvc := application.NewQuoteService(quoteRepo, invoiceRepo, seqRepo, log)
	duningSvc := application.NewDunningService(dunningRepo, invoiceRepo, log)
	exportSvc := application.NewExportService(invoiceRepo, log)

	// PDF-Generator initialisieren (go-pdf/fpdf – reine Go-Bibliothek, kein Chrome nötig)
	pdfGen := pdfInfra.NewFPDFGenerator()
	invoiceSvc.SetPDFGenerator(pdfGen)
	quoteSvc.SetPDFGenerator(pdfGen)
	log.Info("PDF generator initialized (fpdf)")

	// Email-Sender initialisieren (nur wenn SMTP konfiguriert ist)
	smtpHost := os.Getenv("SMTP_HOST")
	if smtpHost != "" {
		smtpPort, _ := strconv.Atoi(os.Getenv("SMTP_PORT"))
		if smtpPort == 0 {
			smtpPort = 587
		}
		emailSender := emailInfra.NewSMTPSender(emailInfra.SMTPConfig{
			Host:     smtpHost,
			Port:     smtpPort,
			Username: os.Getenv("SMTP_USERNAME"),
			Password: os.Getenv("SMTP_PASSWORD"),
			FromName: getEnvOrDefault("SMTP_FROM_NAME", "RentFlow"),
			FromAddr: os.Getenv("SMTP_FROM_ADDRESS"),
		})
		invoiceSvc.SetEmailSender(emailSender)
		log.Info("Email sender initialized", "host", smtpHost, "port", smtpPort)
	} else {
		log.Warn("SMTP not configured - email sending disabled. Set SMTP_HOST to enable.")
	}

	log.Info("Services initialized")

	// Setup router
	router := httpAdapter.NewRouter(invoiceSvc, quoteSvc, duningSvc, exportSvc, log)

	// Create HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.ServicePort),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Info("Listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server error", err)
		}
	}()

	// Wait for interrupt signal
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

func getEnvOrDefault(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

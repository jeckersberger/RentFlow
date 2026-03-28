package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jeckersberger/EquipFlow/pkg/common/config"
	"github.com/jeckersberger/EquipFlow/pkg/common/database"
	"github.com/jeckersberger/EquipFlow/pkg/common/health"
	"github.com/jeckersberger/EquipFlow/pkg/common/logger"
	"github.com/jeckersberger/EquipFlow/pkg/common/middleware"
	"github.com/jeckersberger/EquipFlow/services/invoice/internal/adapters/httphandler"
	"github.com/jeckersberger/EquipFlow/services/invoice/internal/application"
	"github.com/jeckersberger/EquipFlow/services/invoice/internal/infrastructure/postgres"
)

func main() {
	setDefaults()
	cfg := config.Load()

	log := logger.New(cfg.LogLevel)
	log = logger.WithService(log, "invoice")

	pubKeyPath := cfg.JWTPublicKeyPath
	if _, err := os.Stat(pubKeyPath); os.IsNotExist(err) {
		log.Warn().Str("path", pubKeyPath).Msg("JWT public key not found - generating dev key pair")
		devPubPath, genErr := generateDevPublicKey()
		if genErr != nil {
			log.Fatal().Err(genErr).Msg("failed to generate dev key pair")
		}
		pubKeyPath = devPubPath
		os.Setenv("JWT_PUBLIC_KEY_PATH", pubKeyPath)
		cfg = config.Load()
	}

	ctx := context.Background()
	pool, err := database.NewPool(ctx, cfg.DSN())
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}
	defer pool.Close()

	if err := database.RunMigrations(ctx, pool, "migrations"); err != nil {
		log.Fatal().Err(err).Msg("failed to run migrations")
	}
	log.Info().Msg("database migrations applied")

	invoiceRepo := postgres.NewInvoiceRepo(pool)
	itemRepo := postgres.NewItemRepo(pool)
	seqRepo := postgres.NewSequenceRepo(pool)
	paymentRepo := postgres.NewPaymentRepo(pool)

	invoiceSvc := application.NewInvoiceService(invoiceRepo, itemRepo, seqRepo, paymentRepo, log)

	invoiceH := httphandler.NewInvoiceHandler(invoiceSvc, log)
	healthH := health.Handler(pool, nil)
	livenessH := health.LivenessHandler()

	jwtMW := middleware.JWTAuth(cfg.JWTPublicKeyPath)
	corsMW := middleware.CORS(middleware.CORSConfig{
		AllowedOrigins: cfg.CORSOrigins,
		AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type", "X-Request-ID", "X-Tenant-Slug"},
		MaxAge:         86400,
	})
	recoveryMW := middleware.Recovery(log)
	requestIDMW := middleware.RequestID

	router := httphandler.NewRouter(
		invoiceH,
		healthH, livenessH,
		jwtMW, corsMW, recoveryMW, requestIDMW,
	)

	addr := ":" + cfg.Port
	log.Info().Str("addr", addr).Msg("invoice-service starting")

	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server failed")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down invoice-service...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("server shutdown error")
	}
	log.Info().Msg("invoice-service stopped")
}

func setDefaults() {
	if os.Getenv("DB_NAME") == "" {
		os.Setenv("DB_NAME", "invoice_service")
	}
	if os.Getenv("PORT") == "" {
		os.Setenv("PORT", "8006")
	}
}

func generateDevPublicKey() (string, error) {
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", fmt.Errorf("generate RSA key: %w", err)
	}

	pubDER, err := x509.MarshalPKIXPublicKey(&privKey.PublicKey)
	if err != nil {
		return "", fmt.Errorf("marshal public key: %w", err)
	}

	pubPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubDER,
	})

	pubPath := "/tmp/cratedesk_invoice_public.pem"
	if err := os.WriteFile(pubPath, pubPEM, 0644); err != nil {
		return "", fmt.Errorf("write public key: %w", err)
	}

	os.Stderr.WriteString("[WARN] JWT public key not found - generated dev key at " + pubPath + "\n")
	return pubPath, nil
}

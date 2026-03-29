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
	"github.com/jeckersberger/EquipFlow/services/notification/internal/adapters/httphandler"
	"github.com/jeckersberger/EquipFlow/services/notification/internal/application"
	"github.com/jeckersberger/EquipFlow/services/notification/internal/infrastructure/postgres"
)

func main() {
	// 1. Load configuration with notification-service defaults.
	setDefaults()
	cfg := config.Load()

	// 2. Create logger.
	log := logger.New(cfg.LogLevel)
	log = logger.WithService(log, "notification")

	// 3. Load JWT public key (for verifying tokens issued by auth service).
	pubKeyPath := cfg.JWTPublicKeyPath
	if _, err := os.Stat(pubKeyPath); os.IsNotExist(err) {
		log.Warn().Str("path", pubKeyPath).Msg("JWT public key not found — generating dev key pair")
		devPubPath, genErr := generateDevPublicKey()
		if genErr != nil {
			log.Fatal().Err(genErr).Msg("failed to generate dev key pair")
		}
		pubKeyPath = devPubPath
		os.Setenv("JWT_PUBLIC_KEY_PATH", pubKeyPath)
		cfg = config.Load()
	}

	// 4. Connect to PostgreSQL.
	ctx := context.Background()
	pool, err := database.NewPool(ctx, cfg.DSN())
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}
	defer pool.Close()

	// 5. Run database migrations.
	if err := database.RunMigrations(ctx, pool, "migrations"); err != nil {
		log.Fatal().Err(err).Msg("failed to run migrations")
	}
	log.Info().Msg("database migrations applied")

	// 6. Create repositories.
	notifRepo := postgres.NewNotificationRepo(pool)
	prefRepo := postgres.NewPreferenceRepo(pool)

	// 7. Create application services.
	notifSvc := application.NewNotificationService(notifRepo, log)
	prefSvc := application.NewPreferenceService(prefRepo, log)
	emailCfg := application.LoadEmailConfig()
	emailSvc := application.NewEmailService(emailCfg, log)

	if emailSvc.IsConfigured() {
		log.Info().Str("host", emailCfg.Host).Msg("SMTP configured")
	} else {
		log.Warn().Msg("SMTP not configured — email sending disabled")
	}

	// 8. Create WebSocket hub.
	wsHub := httphandler.NewWSHub(log)
	go wsHub.Run()
	log.Info().Msg("WebSocket hub started")

	// 9. Create HTTP handlers.
	notifH := httphandler.NewNotificationHandler(notifSvc, log)
	prefH := httphandler.NewPreferenceHandler(prefSvc, log)
	emailH := httphandler.NewEmailHandler(emailSvc, log)
	wsH := httphandler.NewWSHandler(wsHub, log)
	healthH := health.Handler(pool, nil)
	livenessH := health.LivenessHandler()

	// 10. Create middleware.
	jwtMW := middleware.JWTAuth(cfg.JWTPublicKeyPath)
	corsMW := middleware.CORS(middleware.CORSConfig{
		AllowedOrigins: cfg.CORSOrigins,
		AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type", "X-Request-ID", "X-Tenant-Slug"},
		MaxAge:         86400,
	})
	recoveryMW := middleware.Recovery(log)
	requestIDMW := middleware.RequestID

	// 11. Create router.
	router := httphandler.NewRouter(
		notifH, prefH, emailH, wsH,
		healthH, livenessH,
		jwtMW, corsMW, recoveryMW, requestIDMW,
	)

	// 12. Start HTTP server.
	addr := ":" + cfg.Port
	log.Info().Str("addr", addr).Msg("notification-service starting")

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

	// 13. Graceful shutdown.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down notification-service...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("server shutdown error")
	}
	log.Info().Msg("notification-service stopped")
}

// setDefaults overrides environment variable defaults for the notification-service.
func setDefaults() {
	if os.Getenv("DB_NAME") == "" {
		os.Setenv("DB_NAME", "notification_service")
	}
	if os.Getenv("PORT") == "" {
		os.Setenv("PORT", "8015")
	}
}

// generateDevPublicKey creates a throwaway 2048-bit RSA key pair and saves the
// public key to a temporary file. Returns the path to the public key PEM file.
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

	pubPath := "/tmp/cratedesk_notification_public.pem"
	if err := os.WriteFile(pubPath, pubPEM, 0644); err != nil {
		return "", fmt.Errorf("write public key: %w", err)
	}

	os.Stderr.WriteString("[WARN] JWT public key not found — generated dev key at " + pubPath + "\n")
	return pubPath, nil
}

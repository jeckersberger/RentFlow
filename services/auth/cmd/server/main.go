package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/jeckersberger/EquipFlow/pkg/common/config"
	"github.com/jeckersberger/EquipFlow/pkg/common/database"
	"github.com/jeckersberger/EquipFlow/pkg/common/health"
	"github.com/jeckersberger/EquipFlow/pkg/common/logger"
	"github.com/jeckersberger/EquipFlow/pkg/common/middleware"
	"github.com/jeckersberger/EquipFlow/services/auth/internal/adapters/httphandler"
	"github.com/jeckersberger/EquipFlow/services/auth/internal/application"
	"github.com/jeckersberger/EquipFlow/services/auth/internal/infrastructure/postgres"
)

func main() {
	// 1. Load configuration with auth-service defaults.
	setDefaults()
	cfg := config.Load()

	// 2. Create logger.
	log := logger.New(cfg.LogLevel)
	log = logger.WithService(log, "auth")

	// 3. Load or generate RSA key pair (must happen before middleware init).
	privateKey, publicKey, err := loadRSAKeys(cfg.JWTPrivateKeyPath, cfg.JWTPublicKeyPath)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load RSA keys")
	}
	// Reload config in case dev keys were generated and env was updated.
	cfg = config.Load()

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

	// 6. Connect to Redis.
	redisOpts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		log.Fatal().Err(err).Str("url", cfg.RedisURL).Msg("failed to parse Redis URL")
	}
	redisClient := redis.NewClient(redisOpts)
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Warn().Err(err).Msg("Redis not reachable — rate limiting and sessions may be degraded")
	} else {
		log.Info().Msg("Redis connected")
	}
	defer redisClient.Close()

	// 7. Create repositories.
	tenantRepo := postgres.NewTenantRepo(pool)
	userRepo := postgres.NewUserRepo(pool)
	sessionRepo := postgres.NewSessionRepo(pool)
	qrLoginRepo := postgres.NewQRLoginRepo(pool)
	resetRepo := postgres.NewPasswordResetRepo(pool)
	configRepo := postgres.NewConfigRepo(pool)
	setupRepo := postgres.NewSetupRepo(pool)

	// 8. Create application services.
	authSvc := application.NewAuthService(
		userRepo, sessionRepo, qrLoginRepo, resetRepo,
		privateKey, publicKey,
		cfg.JWTAccessExpiry, cfg.JWTRefreshExpiry,
		log,
	)
	userSvc := application.NewUserService(userRepo, log)
	tenantSvc := application.NewTenantService(tenantRepo, log)
	setupSvc := application.NewSetupService(setupRepo, tenantRepo, userRepo, log)
	configSvc := application.NewConfigService(configRepo, log)

	// 9. Create HTTP handlers.
	authH := httphandler.NewAuthHandler(authSvc, userSvc, tenantRepo, log)
	userH := httphandler.NewUserHandler(userSvc, log)
	tenantH := httphandler.NewTenantHandler(tenantSvc, log)
	setupH := httphandler.NewSetupHandler(setupSvc, authSvc, log)
	configH := httphandler.NewConfigHandler(configSvc, log)
	healthH := health.Handler(pool, redisClient)
	livenessH := health.LivenessHandler()

	// 10. Create middleware.
	jwtMW := middleware.JWTAuth(cfg.JWTPublicKeyPath)
	corsMW := middleware.CORS(middleware.CORSConfig{
		AllowedOrigins: cfg.CORSOrigins,
		AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type", "X-Request-ID", "X-Tenant-Slug"},
		MaxAge:         86400,
	})
	rateLimitMW := middleware.RateLimit(middleware.RateLimitConfig{
		MaxRequests: 10,
		Window:      time.Minute,
		Client:      redisClient,
	})
	recoveryMW := middleware.Recovery(log)
	requestIDMW := middleware.RequestID

	// 11. Create router.
	router := httphandler.NewRouter(
		authH, userH, tenantH, setupH, configH,
		healthH, livenessH,
		jwtMW, corsMW, rateLimitMW, recoveryMW, requestIDMW,
	)

	// 12. Start HTTP server.
	addr := ":" + cfg.Port
	log.Info().Str("addr", addr).Msg("auth-service starting")

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

	log.Info().Msg("shutting down auth-service...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("server shutdown error")
	}
	log.Info().Msg("auth-service stopped")
}

// setDefaults overrides environment variable defaults for the auth-service.
func setDefaults() {
	if os.Getenv("DB_NAME") == "" {
		os.Setenv("DB_NAME", "auth_service")
	}
	if os.Getenv("PORT") == "" {
		os.Setenv("PORT", "8001")
	}
}

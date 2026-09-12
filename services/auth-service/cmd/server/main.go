package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	nethttp "net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"

	"github.com/jeckersberger/rentflow/pkg/common/cache"
	"github.com/jeckersberger/rentflow/pkg/common/config"
	"github.com/jeckersberger/rentflow/pkg/common/logger"
	commonmw "github.com/jeckersberger/rentflow/pkg/common/middleware"
	authhttp "github.com/jeckersberger/rentflow/services/auth-service/internal/adapters/http"
	"github.com/jeckersberger/rentflow/services/auth-service/internal/application"
	"github.com/jeckersberger/rentflow/services/auth-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/auth-service/internal/infrastructure/repositories"
)

const (
	serviceName = "auth-service"
	servicePort = 8001
)

func main() {
	// Load configuration
	cfg := config.Load(serviceName)
	if cfg.ServicePort == 8080 {
		cfg.ServicePort = servicePort
	}
	log := logger.New(cfg.LogLevel, serviceName)

	log.Info("Starting service", "name", serviceName, "port", cfg.ServicePort, "env", cfg.Environment)

	// Connect to PostgreSQL
	db, err := connectPostgres(cfg.ConnectionString(), log)
	if err != nil {
		log.Fatal("Failed to connect to database", err)
	}
	defer db.Close()

	log.Info("Connected to database")

	// Load or create the persistent RSA key pair. Concurrent service instances
	// converge on the same database row instead of generating different keys.
	signingKeyRepo := repositories.NewPostgresSigningKeyRepository(db)
	signingKeyPair, err := application.LoadOrCreateSigningKeyPair(context.Background(), signingKeyRepo)
	if err != nil {
		log.Fatal("failed to load persistent RSA signing key", err)
	}

	// Create token manager with persisted RSA keys
	tokenMgr, err := application.NewTokenManager(signingKeyPair.PrivateKeyPEM, signingKeyPair.PublicKeyPEM)
	if err != nil {
		log.Fatal("failed to create token manager", err)
	}

	// Create repositories
	userRepo := repositories.NewPostgresUserRepository(db, log)
	tenantRepo := repositories.NewPostgresTenantRepository(db, log)
	configRepo := repositories.NewPostgresConfigRepository(db, log)

	// Connect to Redis (fuer Sessions und Brute-Force-Schutz)
	redisAddr := os.Getenv("REDIS_URL")
	if redisAddr == "" {
		redisHost := os.Getenv("REDIS_HOST")
		redisPort := os.Getenv("REDIS_PORT")
		if redisHost == "" {
			redisHost = "localhost"
		}
		if redisPort == "" {
			redisPort = "6379"
		}
		redisAddr = redisHost + ":" + redisPort
	}
	redisPassword := os.Getenv("REDIS_PASSWORD")
	redisCache, err := cache.NewRedisCache(redisAddr, redisPassword, 0)
	if err != nil {
		log.Fatal("Failed to connect to Redis", err)
	}
	defer redisCache.Close()
	log.Info("Redis connected", "addr", redisAddr)

	// Create session manager
	sessionMgr := application.NewSessionManager(redisCache, log)

	// Create invitation repository
	invitationRepo := repositories.NewPostgresInvitationRepository(db, log)

	// Create services
	userService := application.NewUserService(userRepo, tenantRepo, tokenMgr, log)
	userService.SetInvitationRepo(invitationRepo)
	userService.SetDB(db)

	// Configure notification-service URL for email sending
	notificationURL := os.Getenv("NOTIFICATION_SERVICE_URL")
	if notificationURL == "" {
		notificationURL = "http://localhost:8015"
	}
	userService.SetNotificationURL(notificationURL)
	log.Info("Notification service configured", "url", notificationURL)
	tenantService := application.NewTenantService(tenantRepo, log)
	configService := application.NewConfigService(configRepo, log)
	setupService := application.NewSetupService(db, userService, tenantService, log)

	// Initialize setup state. The bootstrap secret is written to a local
	// owner-readable file rather than emitted into application logs.
	setupToken, err := setupService.InitializeSetupState(context.Background())
	if err != nil {
		log.Error("failed to initialize setup state", err)
	} else {
		setupTokenPath, fileErr := persistSetupToken(setupToken)
		if fileErr != nil {
			log.Fatal("failed to persist setup token securely", fileErr)
		}
		if setupToken != "" {
			log.Info("Initial setup required; setup token written to local file", "path", setupTokenPath)
		}
	}

	// Seed superadmin if configured
	seedSuperadmin(db, userRepo, tenantRepo, userService, log)

	// Setup router
	router := nethttp.NewServeMux()

	// Health & readiness
	router.HandleFunc("GET /health", healthHandler(serviceName))
	router.HandleFunc("GET /ready", readyHandler(serviceName, db, log))

	// JWKS endpoint
	router.HandleFunc("GET /api/v1/auth/.well-known/jwks", jwksHandler(tokenMgr, log))

	// Setup API routes
	authhttp.SetupRoutes(router, userService, tenantService, setupService, configService, tokenMgr, sessionMgr, log)

	// Wrap router with setup guard middleware
	setupGuardMiddleware := authhttp.SetupGuardMiddleware(setupService, log)
	handlerWithSetupGuard := setupGuardMiddleware(router)

	// Create HTTP server
	srv := &nethttp.Server{
		Addr:         fmt.Sprintf(":%d", cfg.ServicePort),
		Handler:      commonmw.SimpleCORSMiddleware()(handlerWithSetupGuard),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Info("Listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != nethttp.ErrServerClosed {
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
	srv.Shutdown(ctx)
	log.Info("Server stopped")
}

func healthHandler(serviceName string) nethttp.HandlerFunc {
	return func(w nethttp.ResponseWriter, r *nethttp.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(nethttp.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "healthy",
			"service":   serviceName,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	}
}

func readyHandler(serviceName string, db *sql.DB, log logger.Logger) nethttp.HandlerFunc {
	return func(w nethttp.ResponseWriter, r *nethttp.Request) {
		// Check database connection
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := db.PingContext(ctx); err != nil {
			log.Error("Database not ready", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(nethttp.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":  "not ready",
				"service": serviceName,
				"reason":  "database connection failed",
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(nethttp.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "ready",
			"service": serviceName,
		})
	}
}

func connectPostgres(connectionString string, log logger.Logger) (*sql.DB, error) {
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection with a context timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	return db, nil
}

func jwksHandler(tokenMgr *application.TokenManager, log logger.Logger) nethttp.HandlerFunc {
	return func(w nethttp.ResponseWriter, r *nethttp.Request) {
		jwk, err := tokenMgr.JWK()
		if err != nil {
			log.Error("failed to build JWKS response", err)
			nethttp.Error(w, "JWKS unavailable", nethttp.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "public, max-age=300")
		w.WriteHeader(nethttp.StatusOK)
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"keys": []application.PublicJWK{jwk},
		}); err != nil {
			log.Error("failed to encode JWKS response", err)
		}
	}
}

func seedSuperadmin(
	db *sql.DB,
	userRepo *repositories.PostgresUserRepository,
	tenantRepo *repositories.PostgresTenantRepository,
	userService *application.UserService,
	log logger.Logger,
) {
	// Read environment variables
	superadminEmail := os.Getenv("SUPERADMIN_EMAIL")
	superadminPassword := os.Getenv("SUPERADMIN_PASSWORD")
	defaultTenantName := os.Getenv("DEFAULT_TENANT_NAME")
	defaultTenantSlug := os.Getenv("DEFAULT_TENANT_SLUG")

	// Only seed if all required env vars are set
	if superadminEmail == "" || superadminPassword == "" || defaultTenantName == "" || defaultTenantSlug == "" {
		log.Debug("Superadmin seeding not configured (missing env vars)")
		return
	}

	ctx := context.Background()

	// Check if any users exist
	var count int
	err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM auth.users").Scan(&count)
	if err != nil && err != sql.ErrNoRows {
		log.Error("failed to check user count", err)
		return
	}

	if count > 0 {
		log.Debug("Users already exist, skipping superadmin seed")
		return
	}

	// Create default tenant
	tenantID := generateUUID()
	tenant := domain.NewTenant(tenantID, defaultTenantName, defaultTenantSlug)
	if err := tenantRepo.Save(ctx, tenant); err != nil {
		log.Error("failed to create default tenant", err)
		return
	}
	log.Info("default tenant created", "id", tenantID, "name", defaultTenantName)

	// Create superadmin user
	userID := generateUUID()
	user := domain.NewUser(userID, superadminEmail, "", "Super", "Admin", tenantID)
	user.AssignRole("superadmin")

	// Hash password
	passwordMgr := application.NewPasswordManager()
	passwordHash, err := passwordMgr.HashPassword(superadminPassword)
	if err != nil {
		log.Error("failed to hash superadmin password", err)
		return
	}
	user.PasswordHash = passwordHash

	if err := userRepo.Save(ctx, user); err != nil {
		log.Error("failed to create superadmin user", err)
		return
	}

	log.Info("superadmin user created", "id", userID, "email", superadminEmail)
}

func generateUUID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}

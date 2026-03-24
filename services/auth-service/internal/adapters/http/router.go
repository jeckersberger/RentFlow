package http

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/pkg/common/middleware"
	"github.com/jeckersberger/rentflow/services/auth-service/internal/application"
)

// SetupRoutes sets up all HTTP routes
func SetupRoutes(
	mux *http.ServeMux,
	userService *application.UserService,
	tenantService *application.TenantService,
	setupService *application.SetupService,
	configService *application.ConfigService,
	tokenMgr *application.TokenManager,
	sessionMgr *application.SessionManager,
	log logger.Logger,
) {
	handlers := NewHandlers(userService, tenantService, setupService, log)
	handlers.sessionMgr = sessionMgr
	configHandlers := NewConfigHandlers(configService, log)

	// Rate-Limiter: 10 Anfragen pro Minute pro IP fuer Login
	loginRateLimiter := NewRateLimiter(10, 1*time.Minute, log)

	// Setup routes (no authentication required, always accessible)
	mux.HandleFunc("GET /api/v1/setup/status", handlers.GetSetupStatus)
	mux.HandleFunc("POST /api/v1/setup/complete", handlers.CompleteSetup)

	// Auth routes (no authentication required)
	// Login mit Rate-Limiting und Brute-Force-Schutz
	loginHandler := http.HandlerFunc(handlers.Login)
	protectedLogin := BruteForceMiddleware(sessionMgr, log)(RateLimitMiddleware(loginRateLimiter)(loginHandler))
	mux.Handle("POST /api/v1/auth/login", protectedLogin)

	mux.HandleFunc("POST /api/v1/auth/register", handlers.Register)
	mux.HandleFunc("POST /api/v1/auth/refresh", handlers.Refresh)
	mux.HandleFunc("POST /api/v1/auth/qr-login", handlers.QRLogin)
	mux.HandleFunc("POST /api/v1/auth/forgot-password", handlers.ForgotPassword)
	mux.HandleFunc("POST /api/v1/auth/reset-password", handlers.ResetPassword)

	// Authenticated routes - use a wrapper that supports RS256
	authMiddleware := createRS256Middleware(tokenMgr.PublicKeyPEM(), log)

	// Logout (authenticated)
	mux.HandleFunc("POST /api/v1/auth/logout", authMiddleware(http.HandlerFunc(handlers.Logout)).ServeHTTP)

	// User profile (authenticated)
	mux.HandleFunc("GET /api/v1/auth/me", authMiddleware(http.HandlerFunc(handlers.GetMe)).ServeHTTP)
	mux.HandleFunc("PUT /api/v1/auth/password", authMiddleware(http.HandlerFunc(handlers.ChangePassword)).ServeHTTP)
	mux.HandleFunc("PUT /api/v1/auth/profile", authMiddleware(http.HandlerFunc(handlers.UpdateProfile)).ServeHTTP)

	// User management (admin only)
	mux.HandleFunc("GET /api/v1/users", authMiddleware(http.HandlerFunc(handlers.ListUsers)).ServeHTTP)
	mux.HandleFunc("GET /api/v1/users/{id}", authMiddleware(http.HandlerFunc(handlers.GetUser)).ServeHTTP)
	mux.HandleFunc("PUT /api/v1/users/{id}/roles", authMiddleware(http.HandlerFunc(handlers.AssignRole)).ServeHTTP)
	mux.HandleFunc("DELETE /api/v1/users/{id}", authMiddleware(http.HandlerFunc(handlers.DeleteUser)).ServeHTTP)

	// Invitation management (admin only, authenticated)
	mux.HandleFunc("POST /api/v1/users/invite", authMiddleware(http.HandlerFunc(handlers.InviteUser)).ServeHTTP)
	mux.HandleFunc("GET /api/v1/invitations", authMiddleware(http.HandlerFunc(handlers.ListInvitations)).ServeHTTP)

	// Accept invitation (no auth required - public endpoint)
	mux.HandleFunc("POST /api/v1/invitations/{token}/accept", handlers.AcceptInvitation)

	// Tenant management
	mux.HandleFunc("POST /api/v1/tenants", handlers.CreateTenant)
	mux.HandleFunc("GET /api/v1/tenants/{id}", handlers.GetTenant)
	mux.HandleFunc("PUT /api/v1/tenants/{id}", handlers.UpdateTenant)

	// System endpoints (authenticated)
	mux.HandleFunc("GET /api/v1/system/version", authMiddleware(http.HandlerFunc(handlers.GetSystemVersion)).ServeHTTP)

	// Tenant config (authenticated)
	mux.HandleFunc("GET /api/v1/config", authMiddleware(http.HandlerFunc(configHandlers.GetAllConfigs)).ServeHTTP)
	mux.HandleFunc("GET /api/v1/config/{key...}", authMiddleware(http.HandlerFunc(configHandlers.GetConfig)).ServeHTTP)
	mux.HandleFunc("PUT /api/v1/config/{key...}", authMiddleware(http.HandlerFunc(configHandlers.SetConfig)).ServeHTTP)
	mux.HandleFunc("POST /api/v1/config/smtp-test", authMiddleware(http.HandlerFunc(configHandlers.TestSMTP)).ServeHTTP)
}

// createRS256Middleware creates an RS256 JWT validation middleware
func createRS256Middleware(publicKeyPEM string, log logger.Logger) func(http.Handler) http.Handler {
	// Parse public key
	var publicKey *rsa.PublicKey
	if publicKeyPEM != "" {
		block, _ := pem.Decode([]byte(publicKeyPEM))
		if block != nil {
			pubKeyInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
			if err == nil {
				publicKey, _ = pubKeyInterface.(*rsa.PublicKey)
			}
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract JWT from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "missing authorization header", http.StatusUnauthorized)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "invalid authorization header", http.StatusUnauthorized)
				return
			}

			token := parts[1]

			// For now, use the basic JWT verification if public key is available
			// Otherwise, fall back to using the legacy middleware
			if publicKey != nil {
				claims, err := verifyRS256Token(token, publicKey)
				if err != nil {
					log.Warn("failed to verify token", "error", err.Error())
					http.Error(w, "invalid token", http.StatusUnauthorized)
					return
				}

				// Convert to common middleware Claims and store in context
				commonClaims := &middleware.Claims{
					UserID:   claims.Subject,
					TenantID: claims.TenantID,
					Email:    claims.Email,
					Roles:    claims.Roles,
				}
				ctx := middleware.WithClaims(r.Context(), commonClaims)

				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// Fallback to legacy validation if public key is not available
			next.ServeHTTP(w, r)
		})
	}
}

func verifyRS256Token(token string, publicKey *rsa.PublicKey) (*application.AccessTokenClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidTokenFormat
	}

	headerB64, payloadB64, signatureB64 := parts[0], parts[1], parts[2]

	// Verify signature
	message := headerB64 + "." + payloadB64
	hash := sha256.Sum256([]byte(message))

	signature, err := base64.RawURLEncoding.DecodeString(signatureB64)
	if err != nil {
		return nil, err
	}

	err = rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hash[:], signature)
	if err != nil {
		return nil, ErrInvalidSignature
	}

	// Decode payload
	payloadJSON, err := base64.RawURLEncoding.DecodeString(payloadB64)
	if err != nil {
		return nil, err
	}

	var claims application.AccessTokenClaims
	if err := json.Unmarshal(payloadJSON, &claims); err != nil {
		return nil, err
	}

	return &claims, nil
}

var (
	ErrInvalidTokenFormat = errors.New("invalid token format")
	ErrInvalidSignature   = errors.New("invalid signature")
)

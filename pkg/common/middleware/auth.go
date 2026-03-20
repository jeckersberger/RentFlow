package middleware

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

// Claims represents the claims in a JWT token
// This is a simplified JWT implementation without external dependencies
type Claims struct {
	UserID   string   `json:"user_id"`
	TenantID string   `json:"tenant_id"`
	Email    string   `json:"email"`
	Roles    []string `json:"roles"`
	IssuedAt int64    `json:"iat"`
	ExpiresAt int64   `json:"exp"`
	Issuer   string   `json:"iss"`
}

// SimpleJWT is a minimal JWT implementation without external dependencies
// In production, use github.com/golang-jwt/jwt/v5
type SimpleJWT struct {
	secret string
}

// NewSimpleJWT creates a new JWT handler
func NewSimpleJWT(secret string) *SimpleJWT {
	return &SimpleJWT{secret: secret}
}

// CreateToken creates a new JWT token
func (j *SimpleJWT) CreateToken(claims *Claims) (string, error) {
	claims.IssuedAt = time.Now().Unix()
	claims.ExpiresAt = time.Now().Add(time.Hour).Unix()

	// Create header
	header := map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	}

	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", err
	}
	headerB64 := base64.RawURLEncoding.EncodeToString(headerJSON)

	// Create payload
	payloadJSON, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	payloadB64 := base64.RawURLEncoding.EncodeToString(payloadJSON)

	// Create signature
	message := headerB64 + "." + payloadB64
	h := hmac.New(sha256.New, []byte(j.secret))
	h.Write([]byte(message))
	signature := base64.RawURLEncoding.EncodeToString(h.Sum(nil))

	return message + "." + signature, nil
}

// VerifyToken verifies a JWT token and returns the claims
func (j *SimpleJWT) VerifyToken(token string) (*Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidToken
	}

	headerB64, payloadB64, signatureB64 := parts[0], parts[1], parts[2]

	// Verify signature
	message := headerB64 + "." + payloadB64
	h := hmac.New(sha256.New, []byte(j.secret))
	h.Write([]byte(message))
	expectedSignature := base64.RawURLEncoding.EncodeToString(h.Sum(nil))

	if signatureB64 != expectedSignature {
		return nil, ErrInvalidSignature
	}

	// Decode payload
	payloadJSON, err := base64.RawURLEncoding.DecodeString(payloadB64)
	if err != nil {
		return nil, err
	}

	var claims Claims
	if err := json.Unmarshal(payloadJSON, &claims); err != nil {
		return nil, err
	}

	// Check expiration
	if time.Now().Unix() > claims.ExpiresAt {
		return nil, ErrTokenExpired
	}

	return &claims, nil
}

// JWT-related errors
var (
	ErrInvalidToken    = &AuthError{Code: "INVALID_TOKEN", Message: "Invalid token"}
	ErrInvalidSignature = &AuthError{Code: "INVALID_SIGNATURE", Message: "Invalid token signature"}
	ErrTokenExpired    = &AuthError{Code: "TOKEN_EXPIRED", Message: "Token has expired"}
	ErrMissingToken    = &AuthError{Code: "MISSING_TOKEN", Message: "Missing authorization token"}
	ErrInvalidScheme   = &AuthError{Code: "INVALID_SCHEME", Message: "Invalid authorization scheme"}
)

// AuthError represents an authentication error
type AuthError struct {
	Code    string
	Message string
}

// Error implements the error interface
func (e *AuthError) Error() string {
	return e.Message
}

// JWTAuthMiddleware creates a middleware that validates JWT tokens
func JWTAuthMiddleware(secret string) func(http.Handler) http.Handler {
	jwt := NewSimpleJWT(secret)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				writeAuthError(w, http.StatusUnauthorized, ErrMissingToken)
				return
			}

			// Check for Bearer scheme
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				writeAuthError(w, http.StatusUnauthorized, ErrInvalidScheme)
				return
			}

			token := parts[1]

			// Verify token
			claims, err := jwt.VerifyToken(token)
			if err != nil {
				statusCode := http.StatusUnauthorized
				writeAuthError(w, statusCode, err)
				return
			}

			// Add claims to context
			ctx := WithClaims(r.Context(), claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole creates a middleware that checks for required roles
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := ExtractClaims(r.Context())
			if claims == nil {
				writeAuthError(w, http.StatusUnauthorized, ErrMissingToken)
				return
			}

			// Check if user has any of the required roles
			hasRole := false
			for _, role := range roles {
				for _, userRole := range claims.Roles {
					if userRole == role {
						hasRole = true
						break
					}
				}
				if hasRole {
					break
				}
			}

			if !hasRole {
				writeAuthError(w, http.StatusForbidden, &AuthError{
					Code:    "INSUFFICIENT_PERMISSIONS",
					Message: "User does not have required permissions",
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Context keys for claims
type contextKey string

const (
	claimsKey contextKey = "claims"
)

// WithClaims adds claims to the context
func WithClaims(ctx context.Context, claims *Claims) context.Context {
	return context.WithValue(ctx, claimsKey, claims)
}

// ExtractClaims extracts claims from the context
func ExtractClaims(ctx context.Context) *Claims {
	claims, ok := ctx.Value(claimsKey).(*Claims)
	if !ok {
		return nil
	}
	return claims
}

// GetUserID extracts the user ID from the context
func GetUserID(ctx context.Context) string {
	claims := ExtractClaims(ctx)
	if claims == nil {
		return ""
	}
	return claims.UserID
}

// GetTenantIDFromClaims extracts the tenant ID from the context
func GetTenantIDFromClaims(ctx context.Context) string {
	claims := ExtractClaims(ctx)
	if claims == nil {
		return ""
	}
	return claims.TenantID
}

// GetEmail extracts the email from the context
func GetEmail(ctx context.Context) string {
	claims := ExtractClaims(ctx)
	if claims == nil {
		return ""
	}
	return claims.Email
}

// GetRoles extracts the roles from the context
func GetRoles(ctx context.Context) []string {
	claims := ExtractClaims(ctx)
	if claims == nil {
		return []string{}
	}
	return claims.Roles
}

// HasRole checks if the user has a specific role
func HasRole(ctx context.Context, role string) bool {
	roles := GetRoles(ctx)
	for _, r := range roles {
		if r == role {
			return true
		}
	}
	return false
}

// writeAuthError writes an authentication error response
func writeAuthError(w http.ResponseWriter, statusCode int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	authErr, ok := err.(*AuthError)
	if !ok {
		authErr = &AuthError{
			Code:    "UNKNOWN_ERROR",
			Message: err.Error(),
		}
	}

	response := map[string]interface{}{
		"code":    authErr.Code,
		"message": authErr.Message,
	}

	json.NewEncoder(w).Encode(response)
}

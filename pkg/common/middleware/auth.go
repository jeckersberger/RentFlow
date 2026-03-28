package middleware

import (
	"context"
	"crypto/rsa"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/jeckersberger/EquipFlow/pkg/common/response"
)

type contextKey string

const claimsKey contextKey = "claims"

// Claims represents the JWT token payload used across all services.
type Claims struct {
	UserID   uuid.UUID `json:"user_id"`
	TenantID uuid.UUID `json:"tenant_id"`
	Email    string    `json:"email"`
	Role     string    `json:"role"`
	jwt.RegisteredClaims
}

// JWTAuth returns a chi-compatible middleware that validates RS256 JWT tokens.
// It reads the public key from the given PEM file path.
func JWTAuth(publicKeyPath string) func(http.Handler) http.Handler {
	keyData, err := os.ReadFile(publicKeyPath)
	if err != nil {
		panic("middleware: failed to read JWT public key: " + err.Error())
	}

	pubKey, err := jwt.ParseRSAPublicKeyFromPEM(keyData)
	if err != nil {
		panic("middleware: failed to parse JWT public key: " + err.Error())
	}

	return jwtMiddleware(pubKey)
}

func jwtMiddleware(pubKey *rsa.PublicKey) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				response.Error(w, http.StatusUnauthorized, "MISSING_TOKEN", "Authorization header required")
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
				response.Error(w, http.StatusUnauthorized, "INVALID_TOKEN", "Authorization header must be Bearer {token}")
				return
			}

			claims := &Claims{}
			token, err := jwt.ParseWithClaims(parts[1], claims, func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return pubKey, nil
			})

			if err != nil || !token.Valid {
				response.Error(w, http.StatusUnauthorized, "INVALID_TOKEN", "Token is invalid or expired")
				return
			}

			ctx := context.WithValue(r.Context(), claimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetClaims extracts the JWT claims from the request context.
// Returns nil if no claims are present.
func GetClaims(ctx context.Context) *Claims {
	claims, _ := ctx.Value(claimsKey).(*Claims)
	return claims
}

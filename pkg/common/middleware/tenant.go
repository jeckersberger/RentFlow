package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"

	"github.com/jeckersberger/EquipFlow/pkg/common/response"
)

const tenantKey contextKey = "tenant_id"

// TenantFromJWT extracts the tenant_id from JWT claims (set by JWTAuth)
// and places it in the request context for downstream handlers.
// This middleware must be placed after JWTAuth in the chain.
func TenantFromJWT(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := GetClaims(r.Context())
		if claims == nil {
			response.Error(w, http.StatusUnauthorized, "MISSING_CLAIMS", "No JWT claims found in context")
			return
		}

		if claims.TenantID == uuid.Nil {
			response.Error(w, http.StatusForbidden, "MISSING_TENANT", "Token does not contain a tenant_id")
			return
		}

		ctx := context.WithValue(r.Context(), tenantKey, claims.TenantID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetTenantID extracts the tenant UUID from the request context.
// Returns uuid.Nil if no tenant ID is present.
func GetTenantID(ctx context.Context) uuid.UUID {
	tid, _ := ctx.Value(tenantKey).(uuid.UUID)
	return tid
}

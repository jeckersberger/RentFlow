package middleware

import (
	"context"
	"net/http"
	"strings"
)

// TenantContext keys
type tenantContextKey string

const (
	tenantIDKey tenantContextKey = "tenant_id"
)

// TenantMiddleware extracts the tenant ID from the request
// Tenant ID can come from:
// 1. JWT claims (preferred)
// 2. X-Tenant-ID header
// 3. Subdomain (subdomain.example.com)
func TenantMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tenantID := ""

			// First, try to get from JWT claims
			claims := ExtractClaims(r.Context())
			if claims != nil && claims.TenantID != "" {
				tenantID = claims.TenantID
			}

			// Fall back to header — only accept X-Tenant-ID from internal service-to-service calls.
			// Internal services must set the X-Internal-Service header to authenticate the override.
			if tenantID == "" && r.Header.Get("X-Internal-Service") == "rentflow" {
				tenantID = r.Header.Get("X-Tenant-ID")
			}

			// Fall back to subdomain extraction
			if tenantID == "" {
				tenantID = extractTenantFromHost(r.Host)
			}

			// Create new context with tenant ID
			ctx := WithTenantID(r.Context(), tenantID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetTenantID extracts the tenant ID from the context
func GetTenantID(ctx context.Context) string {
	if tenantID, ok := ctx.Value(tenantIDKey).(string); ok {
		return tenantID
	}
	return ""
}

// WithTenantID adds a tenant ID to the context
func WithTenantID(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, tenantIDKey, tenantID)
}

// RequireTenant is a middleware that requires a valid tenant ID
func RequireTenant() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tenantID := GetTenantID(r.Context())
			if tenantID == "" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`{"code":"MISSING_TENANT","message":"Tenant ID is required"}`))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// EnforceTenant creates a middleware that enforces a specific tenant
// This is useful for routes that should only be accessible by users of a specific tenant
func EnforceTenant(requiredTenant string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tenantID := GetTenantID(r.Context())
			if tenantID != requiredTenant {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte(`{"code":"TENANT_MISMATCH","message":"Access denied for this tenant"}`))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// extractTenantFromHost extracts the tenant from the host header
// Format: tenant.example.com
func extractTenantFromHost(host string) string {
	// Remove port if present
	if idx := strings.Index(host, ":"); idx != -1 {
		host = host[:idx]
	}

	// Split by dots
	parts := strings.Split(host, ".")
	if len(parts) > 1 {
		// Check if it's not localhost
		if parts[len(parts)-1] != "localhost" {
			return parts[0]
		}
	}

	return ""
}

// IsSingleTenant checks if the application is in single-tenant mode
// In single-tenant mode, all requests belong to the same tenant
func IsSingleTenant(ctx context.Context) bool {
	// This could be configured per application
	// For now, we return false (multi-tenant by default)
	return false
}

// GetSingleTenantID returns the single tenant ID if in single-tenant mode
func GetSingleTenantID() string {
	// This would be configured in the application config
	return "default"
}

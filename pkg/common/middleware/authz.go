package middleware

import (
	"net/http"

	"github.com/jeckersberger/EquipFlow/pkg/common/response"
)

// RequireRole returns a middleware that checks the JWT role claim against the allowed roles.
// Requests from users without a matching role receive 403 Forbidden.
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]bool, len(roles))
	for _, r := range roles {
		allowed[r] = true
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := GetClaims(r.Context())
			if claims == nil || !allowed[claims.Role] {
				response.Error(w, http.StatusForbidden, "FORBIDDEN", "Keine Berechtigung fuer diese Aktion")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

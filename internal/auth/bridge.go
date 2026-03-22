package auth

import (
	"net/http"

	"github.com/pebblr/pebblr/internal/domain"
	"github.com/pebblr/pebblr/internal/rbac"
)

// ClaimsBridge returns middleware that converts UserClaims (set by auth.Middleware)
// into a domain.User stored via rbac.WithUser. This bridge is used by both the
// static authenticator and the future OIDC authenticator.
func ClaimsBridge(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := ClaimsFromContext(r.Context())
		if claims == nil {
			http.Error(w, `{"error":{"code":"UNAUTHORIZED","message":"no claims in context"}}`, http.StatusUnauthorized)
			return
		}

		role := domain.RoleRep
		if len(claims.Roles) > 0 {
			role = claims.Roles[0]
		}

		user := &domain.User{
			ID:      claims.Sub,
			Email:   claims.Email,
			Name:    claims.Name,
			Role:    role,
			TeamIDs: claims.TeamIDs,
		}

		ctx := rbac.WithUser(r.Context(), user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

package auth

import (
	"net/http"

	"github.com/pebblr/pebblr/internal/domain"
	"github.com/pebblr/pebblr/internal/rbac"
)

// rolePrecedence maps roles to their privilege level for comparison.
// Higher value = higher privilege.
var rolePrecedence = map[domain.Role]int{
	domain.RoleRep:     1,
	domain.RoleManager: 2,
	domain.RoleAdmin:   3,
}

// highestRole returns the most privileged role from the list.
// Falls back to RoleRep if the list is empty.
func highestRole(roles []domain.Role) domain.Role {
	best := domain.RoleRep
	bestLevel := 0
	for _, r := range roles {
		if level, ok := rolePrecedence[r]; ok && level > bestLevel {
			best = r
			bestLevel = level
		}
	}
	return best
}

// ClaimsBridge returns middleware that converts UserClaims (set by auth.Middleware)
// into a domain.User stored via rbac.WithUser. This bridge is used by both the
// static authenticator and the future OIDC authenticator.
func ClaimsBridge(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := ClaimsFromContext(r.Context())
		if claims == nil {
			writeJSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "no claims in context")
			return
		}

		role := highestRole(claims.Roles)

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

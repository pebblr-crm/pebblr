package auth

import (
	"net/http"

	"github.com/pebblr/pebblr/internal/domain"
	"github.com/pebblr/pebblr/internal/rbac"
)

// rolePriority maps roles to their privilege level. Higher is more privileged.
var rolePriority = map[domain.Role]int{
	domain.RoleRep:     0,
	domain.RoleManager: 1,
	domain.RoleAdmin:   2,
}

// highestRole returns the most privileged role from the slice, defaulting
// to RoleRep if the slice is empty or contains no valid roles.
func highestRole(roles []domain.Role) domain.Role {
	best := domain.RoleRep
	bestPri := -1
	for _, r := range roles {
		if p, ok := rolePriority[r]; ok && p > bestPri {
			best = r
			bestPri = p
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

		// Pick the highest-privilege role deterministically.
		// Azure AD does not guarantee role ordering, so relying on
		// claims.Roles[0] could silently change effective permissions.
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

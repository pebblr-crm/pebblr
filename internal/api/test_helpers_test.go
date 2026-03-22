package api_test

import (
	"net/http"

	"github.com/pebblr/pebblr/internal/domain"
	"github.com/pebblr/pebblr/internal/rbac"
)

// injectUser wraps an HTTP handler to inject the given user into the request context.
func injectUser(user *domain.User, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := rbac.WithUser(r.Context(), user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func testAdminUser() *domain.User {
	return &domain.User{
		ID:      "admin-1",
		Name:    "Admin User",
		Role:    domain.RoleAdmin,
		TeamIDs: []string{"team-1"},
	}
}

func testRepUser() *domain.User {
	return &domain.User{
		ID:      "rep-1",
		Name:    "Rep User",
		Role:    domain.RoleRep,
		TeamIDs: []string{"team-1"},
	}
}
